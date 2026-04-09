package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

const maxTagsPerResource = 20
const maxTagLength = 32

var (
	ErrResourceNotFound     = errors.New("resource not found")
	ErrTooManyTags          = errors.New("maximum 20 tags per resource")
	ErrTagTooLong           = errors.New("tag must be at most 32 characters")
	ErrMimeTypeNotAllowed   = errors.New("mime type is not in the allowlist")
	ErrVersionNotFound          = errors.New("version not found")
	ErrResourceAccessDenied     = errors.New("access denied to this resource")
	ErrExtractedTextNotAllowed  = errors.New("extracted_text is only allowed for PDF and DOCX files")
)

type ResourceService struct {
	resourceRepo *repository.ResourceRepository
	versionRepo  *repository.ResourceVersionRepository
	tagRepo      *repository.ResourceTagRepository
	courseRepo   *repository.CourseRepository
	memberRepo   *repository.CourseMembershipRepository
	audit        *AuditService
	cfg          *config.Config
}

func NewResourceService(
	resourceRepo *repository.ResourceRepository,
	versionRepo *repository.ResourceVersionRepository,
	tagRepo *repository.ResourceTagRepository,
	courseRepo *repository.CourseRepository,
	memberRepo *repository.CourseMembershipRepository,
	audit *AuditService,
	cfg *config.Config,
) *ResourceService {
	return &ResourceService{
		resourceRepo: resourceRepo,
		versionRepo:  versionRepo,
		tagRepo:      tagRepo,
		courseRepo:    courseRepo,
		memberRepo:   memberRepo,
		audit:        audit,
		cfg:          cfg,
	}
}

func (s *ResourceService) CreateResource(ctx context.Context, req *dto.CreateResourceRequest, actorID uuid.UUID, callerRole models.Role) (*models.Resource, []string, error) {
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid course_id")
	}
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		return nil, nil, fmt.Errorf("course not found")
	}

	// Require course staff or admin to create resources
	if err := s.requireCourseStaff(ctx, courseID, actorID, callerRole); err != nil {
		return nil, nil, err
	}

	if req.Title == "" {
		return nil, nil, fmt.Errorf("title is required")
	}

	resType := models.ResourceType(req.ResourceType)
	if !models.ValidResourceTypes[resType] {
		return nil, nil, fmt.Errorf("invalid resource_type: %s", req.ResourceType)
	}

	visibility := models.ResourceVisibility(req.Visibility)
	if !models.ValidVisibilities[visibility] {
		return nil, nil, fmt.Errorf("invalid visibility: %s", req.Visibility)
	}

	if resType == models.ResourceTypeLink && (req.LinkURL == nil || *req.LinkURL == "") {
		return nil, nil, fmt.Errorf("link_url is required for Link resources")
	}

	// Validate tags
	if err := validateTags(req.Tags); err != nil {
		return nil, nil, err
	}

	var nodeID *uuid.UUID
	if req.NodeID != nil {
		nid, err := uuid.Parse(*req.NodeID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid node_id")
		}
		nodeID = &nid
	}

	now := time.Now()
	resource := &models.Resource{
		ID:           uuid.New(),
		CourseID:     courseID,
		NodeID:       nodeID,
		Title:        req.Title,
		Description:  req.Description,
		ResourceType: resType,
		Visibility:   visibility,
		LinkURL:      req.LinkURL,
		CreatedBy:    actorID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		return nil, nil, err
	}

	// Create tags
	tags := s.saveTags(ctx, resource.ID, req.Tags)

	s.audit.Log(ctx, "resource", resource.ID, actorID, "created", map[string]interface{}{
		"course_id":     courseID,
		"title":         req.Title,
		"resource_type": string(resType),
		"visibility":    string(visibility),
	})

	return resource, tags, nil
}

func (s *ResourceService) GetResource(ctx context.Context, id, callerID uuid.UUID, callerRole models.Role) (*models.Resource, []string, error) {
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, ErrResourceNotFound
	}

	if err := s.checkResourceAccess(ctx, resource, callerID, callerRole); err != nil {
		return nil, nil, err
	}

	tags, _ := s.getTagStrings(ctx, id)
	return resource, tags, nil
}

func (s *ResourceService) ListResources(ctx context.Context, courseID, callerID uuid.UUID, callerRole models.Role, offset, limit int) ([]models.Resource, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Determine visibility based on membership
	if callerRole == models.RoleAdministrator {
		return s.resourceRepo.ListByCourse(ctx, courseID, offset, limit)
	}

	m, err := s.memberRepo.GetByAccountAndCourse(ctx, callerID, courseID)
	if err != nil {
		return nil, ErrNotCourseMember
	}

	if m.Role == models.MembershipRoleStaff {
		return s.resourceRepo.ListByCourse(ctx, courseID, offset, limit)
	}

	// Enrolled users see only Enrolled-visibility resources
	return s.resourceRepo.ListByCourseAndVisibility(ctx, courseID, models.VisibilityEnrolled, offset, limit)
}

func (s *ResourceService) UpdateResource(ctx context.Context, id uuid.UUID, req *dto.UpdateResourceRequest, actorID uuid.UUID, callerRole models.Role) (*models.Resource, []string, error) {
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, ErrResourceNotFound
	}

	// Require course staff or admin
	if err := s.requireCourseStaff(ctx, resource.CourseID, actorID, callerRole); err != nil {
		return nil, nil, err
	}

	if req.Title != nil {
		resource.Title = *req.Title
	}
	if req.Description != nil {
		resource.Description = req.Description
	}
	if req.Visibility != nil {
		vis := models.ResourceVisibility(*req.Visibility)
		if !models.ValidVisibilities[vis] {
			return nil, nil, fmt.Errorf("invalid visibility: %s", *req.Visibility)
		}
		resource.Visibility = vis
	}
	if req.NodeID != nil {
		if *req.NodeID == "" {
			resource.NodeID = nil
		} else {
			nid, err := uuid.Parse(*req.NodeID)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid node_id")
			}
			resource.NodeID = &nid
		}
	}
	if req.LinkURL != nil {
		resource.LinkURL = req.LinkURL
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		return nil, nil, err
	}

	// Update tags if provided
	var tags []string
	if req.Tags != nil {
		if err := validateTags(req.Tags); err != nil {
			return nil, nil, err
		}
		tags = s.replaceTags(ctx, resource.ID, req.Tags)
	} else {
		tags, _ = s.getTagStrings(ctx, id)
	}

	s.audit.Log(ctx, "resource", resource.ID, actorID, "updated", map[string]interface{}{
		"title":      resource.Title,
		"visibility": string(resource.Visibility),
	})

	return resource, tags, nil
}

func (s *ResourceService) SearchResources(ctx context.Context, courseID uuid.UUID, query string, callerID uuid.UUID, callerRole models.Role, offset, limit int) ([]models.Resource, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Admin and course staff: search all visibility levels
	if callerRole == models.RoleAdministrator {
		return s.resourceRepo.Search(ctx, courseID, query, offset, limit)
	}

	m, err := s.memberRepo.GetByAccountAndCourse(ctx, callerID, courseID)
	if err != nil {
		return nil, ErrNotCourseMember
	}

	if m.Role == models.MembershipRoleStaff {
		return s.resourceRepo.Search(ctx, courseID, query, offset, limit)
	}

	// Enrolled users: search only Enrolled-visibility resources
	return s.resourceRepo.SearchWithVisibility(ctx, courseID, query, models.VisibilityEnrolled, offset, limit)
}

// Version management

func (s *ResourceService) UploadVersion(ctx context.Context, resourceID uuid.UUID, fileName, mimeType string, sizeBytes int64, fileContent io.Reader, extractedText *string, actorID uuid.UUID, callerRole models.Role) (*models.ResourceVersion, error) {
	resource, err := s.resourceRepo.GetByID(ctx, resourceID)
	if err != nil {
		return nil, ErrResourceNotFound
	}

	// Require course staff or admin
	if err := s.requireCourseStaff(ctx, resource.CourseID, actorID, callerRole); err != nil {
		return nil, err
	}

	// Validate mime type
	if !models.AllowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("%w: %s", ErrMimeTypeNotAllowed, mimeType)
	}

	// Reject extracted_text for non-PDF/DOCX files
	if extractedText != nil && *extractedText != "" && !models.TextExtractableMimeTypes[mimeType] {
		return nil, ErrExtractedTextNotAllowed
	}

	// Compute SHA-256 while writing to disk
	storageDir := filepath.Join(s.cfg.StoragePath, resource.CourseID.String(), resourceID.String())
	if err := os.MkdirAll(storageDir, 0750); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}

	versionID := uuid.New()
	storagePath := filepath.Join(storageDir, versionID.String())

	f, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	hasher := sha256.New()
	written, err := io.Copy(f, io.TeeReader(fileContent, hasher))
	f.Close()
	if err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("writing file: %w", err)
	}

	sha256Hash := hex.EncodeToString(hasher.Sum(nil))

	// Check for deduplication — if same hash exists, reuse storage path
	existing, err := s.versionRepo.GetBySHA256(ctx, sha256Hash)
	if err == nil && existing != nil {
		// Same file already exists — remove the new copy and reference existing path
		os.Remove(storagePath)
		storagePath = existing.StoragePath
	}

	// Get next version number
	latestNum, err := s.versionRepo.GetLatestVersionNumber(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("getting latest version: %w", err)
	}

	version := &models.ResourceVersion{
		ID:            versionID,
		ResourceID:    resourceID,
		VersionNumber: latestNum + 1,
		FileName:      fileName,
		MimeType:      mimeType,
		SizeBytes:     written,
		SHA256Hash:    sha256Hash,
		StoragePath:   storagePath,
		ExtractedText: extractedText,
		UploadedBy:    actorID,
		CreatedAt:     time.Now(),
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	// Update latest version pointer
	if err := s.resourceRepo.UpdateLatestVersion(ctx, resourceID, versionID); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "resource_version", versionID, actorID, "uploaded", map[string]interface{}{
		"resource_id":    resourceID,
		"version_number": version.VersionNumber,
		"file_name":      fileName,
		"mime_type":      mimeType,
		"size_bytes":     written,
		"sha256_hash":    sha256Hash,
	})

	return version, nil
}

func (s *ResourceService) ListVersions(ctx context.Context, resourceID uuid.UUID, callerID uuid.UUID, callerRole models.Role) ([]models.ResourceVersion, error) {
	resource, err := s.resourceRepo.GetByID(ctx, resourceID)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	if err := s.checkResourceAccess(ctx, resource, callerID, callerRole); err != nil {
		return nil, err
	}
	return s.versionRepo.ListByResource(ctx, resourceID)
}

func (s *ResourceService) GetVersion(ctx context.Context, versionID uuid.UUID) (*models.ResourceVersion, error) {
	v, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}
	return v, nil
}

func (s *ResourceService) GetDownloadPath(ctx context.Context, versionID, callerID uuid.UUID, callerRole models.Role) (*models.ResourceVersion, error) {
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	resource, err := s.resourceRepo.GetByID(ctx, version.ResourceID)
	if err != nil {
		return nil, ErrResourceNotFound
	}

	if err := s.checkResourceAccess(ctx, resource, callerID, callerRole); err != nil {
		return nil, err
	}

	return version, nil
}

// Access control

func (s *ResourceService) checkResourceAccess(ctx context.Context, resource *models.Resource, callerID uuid.UUID, callerRole models.Role) error {
	if callerRole == models.RoleAdministrator {
		return nil
	}

	m, err := s.memberRepo.GetByAccountAndCourse(ctx, callerID, resource.CourseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrResourceAccessDenied
		}
		return err
	}

	// Staff can see everything
	if m.Role == models.MembershipRoleStaff {
		return nil
	}

	// Enrolled users can only see Enrolled-visibility resources
	if resource.Visibility == models.VisibilityStaff {
		return ErrResourceAccessDenied
	}

	return nil
}

// Course-level authorization helpers

func (s *ResourceService) requireCourseStaff(ctx context.Context, courseID, actorID uuid.UUID, callerRole models.Role) error {
	if callerRole == models.RoleAdministrator {
		return nil
	}
	m, err := s.memberRepo.GetByAccountAndCourse(ctx, actorID, courseID)
	if err != nil {
		return ErrResourceAccessDenied
	}
	if m.Role != models.MembershipRoleStaff {
		return ErrResourceAccessDenied
	}
	return nil
}

func (s *ResourceService) checkCourseMembership(ctx context.Context, courseID, callerID uuid.UUID, callerRole models.Role) error {
	if callerRole == models.RoleAdministrator {
		return nil
	}
	_, err := s.memberRepo.GetByAccountAndCourse(ctx, callerID, courseID)
	if err != nil {
		return ErrNotCourseMember
	}
	return nil
}

// Tag helpers

func validateTags(tags []string) error {
	if len(tags) > maxTagsPerResource {
		return ErrTooManyTags
	}
	for _, t := range tags {
		if len(t) > maxTagLength {
			return fmt.Errorf("%w: '%s'", ErrTagTooLong, t)
		}
		if t == "" {
			return fmt.Errorf("tags cannot be empty strings")
		}
	}
	return nil
}

func (s *ResourceService) saveTags(ctx context.Context, resourceID uuid.UUID, tagStrings []string) []string {
	now := time.Now()
	for _, t := range tagStrings {
		tag := &models.ResourceTag{
			ID:         uuid.New(),
			ResourceID: resourceID,
			Tag:        t,
			CreatedAt:  now,
		}
		s.tagRepo.Create(ctx, tag)
	}
	return tagStrings
}

func (s *ResourceService) replaceTags(ctx context.Context, resourceID uuid.UUID, tagStrings []string) []string {
	now := time.Now()
	tags := make([]models.ResourceTag, len(tagStrings))
	for i, t := range tagStrings {
		tags[i] = models.ResourceTag{
			ID:         uuid.New(),
			ResourceID: resourceID,
			Tag:        t,
			CreatedAt:  now,
		}
	}
	s.tagRepo.ReplaceAll(ctx, resourceID, tags)
	return tagStrings
}

func (s *ResourceService) getTagStrings(ctx context.Context, resourceID uuid.UUID) ([]string, error) {
	tags, err := s.tagRepo.ListByResource(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(tags))
	for i, t := range tags {
		result[i] = t.Tag
	}
	return result, nil
}
