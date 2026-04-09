package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrReviewConfigNotFound  = errors.New("review config not found")
	ErrReviewRequestNotFound = errors.New("review request not found")
	ErrReviewLevelNotFound   = errors.New("review level not found")
	ErrInvalidLevels         = errors.New("required_levels must be between 1 and 3")
	ErrRequestNotInReview    = errors.New("review request is not in review")
	ErrLevelNotPending       = errors.New("level is not in Pending status")
	ErrNotCurrentLevel       = errors.New("can only decide on the current level")
	ErrNotAssignee           = errors.New("you are not assigned to this level")
	ErrDecisionRequired      = errors.New("decision is required (Approved, Rejected, or Returned)")
	ErrRequestAlreadyFinal   = errors.New("review request already has a final decision")
)

// DispositionCallback is called when a review reaches final approval or rejection.
// Implementations write-back to the originating record.
type DispositionCallback func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error

type ReviewService struct {
	repo                *repository.ReviewRepository
	audit               *AuditService
	dispositionCallbacks map[string]DispositionCallback
}

func NewReviewService(
	repo *repository.ReviewRepository,
	audit *AuditService,
) *ReviewService {
	return &ReviewService{
		repo:                repo,
		audit:               audit,
		dispositionCallbacks: make(map[string]DispositionCallback),
	}
}

// RegisterDisposition registers a callback for write-back to originating records
func (s *ReviewService) RegisterDisposition(entityType string, cb DispositionCallback) {
	s.dispositionCallbacks[entityType] = cb
}

// Config management

func (s *ReviewService) CreateConfig(ctx context.Context, req *dto.CreateReviewConfigRequest, actorID uuid.UUID) (*models.ReviewConfig, error) {
	if req.ReviewType == "" {
		return nil, fmt.Errorf("review_type is required")
	}
	if req.RequiredLevels < 1 || req.RequiredLevels > 3 {
		return nil, ErrInvalidLevels
	}

	now := time.Now()
	cfg := &models.ReviewConfig{
		ID:             uuid.New(),
		ReviewType:     req.ReviewType,
		Description:    req.Description,
		RequiredLevels: req.RequiredLevels,
		CreatedBy:      actorID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateConfig(ctx, cfg); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "review_config", cfg.ID, actorID, "created", map[string]interface{}{
		"review_type":     cfg.ReviewType,
		"required_levels": cfg.RequiredLevels,
	})

	return cfg, nil
}

func (s *ReviewService) GetConfig(ctx context.Context, id uuid.UUID) (*models.ReviewConfig, error) {
	cfg, err := s.repo.GetConfigByID(ctx, id)
	if err != nil {
		return nil, ErrReviewConfigNotFound
	}
	return cfg, nil
}

func (s *ReviewService) ListConfigs(ctx context.Context, offset, limit int) ([]models.ReviewConfig, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListConfigs(ctx, offset, limit)
}

func (s *ReviewService) UpdateConfig(ctx context.Context, id uuid.UUID, req *dto.UpdateReviewConfigRequest, actorID uuid.UUID) (*models.ReviewConfig, error) {
	cfg, err := s.repo.GetConfigByID(ctx, id)
	if err != nil {
		return nil, ErrReviewConfigNotFound
	}

	if req.Description != nil {
		cfg.Description = req.Description
	}
	if req.RequiredLevels != nil {
		if *req.RequiredLevels < 1 || *req.RequiredLevels > 3 {
			return nil, ErrInvalidLevels
		}
		cfg.RequiredLevels = *req.RequiredLevels
	}

	if err := s.repo.UpdateConfig(ctx, cfg); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "review_config", id, actorID, "updated", map[string]interface{}{
		"required_levels": cfg.RequiredLevels,
	})

	return cfg, nil
}

func (s *ReviewService) DeleteConfig(ctx context.Context, id, actorID uuid.UUID) error {
	if _, err := s.repo.GetConfigByID(ctx, id); err != nil {
		return ErrReviewConfigNotFound
	}
	if err := s.repo.DeleteConfig(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, "review_config", id, actorID, "deleted", nil)
	return nil
}

// Submit a new review request

func (s *ReviewService) SubmitRequest(ctx context.Context, req *dto.SubmitReviewRequest, submitterID uuid.UUID) (*models.ReviewRequest, error) {
	if req.ReviewType == "" || req.EntityType == "" || req.EntityID == "" {
		return nil, fmt.Errorf("review_type, entity_type, and entity_id are required")
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity_id")
	}

	// Look up config to determine required levels
	cfg, err := s.repo.GetConfigByType(ctx, req.ReviewType)
	if err != nil {
		return nil, fmt.Errorf("no review config found for type: %s", req.ReviewType)
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id")
		}
		// Verify parent exists
		if _, err := s.repo.GetRequestByID(ctx, pid); err != nil {
			return nil, fmt.Errorf("parent review request not found")
		}
		parentID = &pid
	}

	now := time.Now()
	request := &models.ReviewRequest{
		ID:             uuid.New(),
		ReviewType:     req.ReviewType,
		EntityType:     req.EntityType,
		EntityID:       entityID,
		RequiredLevels: cfg.RequiredLevels,
		CurrentLevel:   1,
		Status:         models.ReviewRequestInReview,
		SubmittedBy:    submitterID,
		ParentID:       parentID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateRequest(ctx, request); err != nil {
		return nil, err
	}

	// Create all level records upfront
	for i := 1; i <= cfg.RequiredLevels; i++ {
		level := &models.ReviewLevel{
			ID:        uuid.New(),
			RequestID: request.ID,
			Level:     i,
			Status:    models.LevelPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.repo.CreateLevel(ctx, level); err != nil {
			return nil, err
		}
	}

	s.audit.Log(ctx, "review_request", request.ID, submitterID, "submitted", map[string]interface{}{
		"review_type":     req.ReviewType,
		"entity_type":     req.EntityType,
		"entity_id":       req.EntityID,
		"required_levels": cfg.RequiredLevels,
		"parent_id":       parentID,
	})

	return request, nil
}

// Get request with levels

func (s *ReviewService) GetRequest(ctx context.Context, id uuid.UUID) (*models.ReviewRequest, []models.ReviewLevel, error) {
	request, err := s.repo.GetRequestByID(ctx, id)
	if err != nil {
		return nil, nil, ErrReviewRequestNotFound
	}
	levels, err := s.repo.ListLevelsByRequest(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return request, levels, nil
}

func (s *ReviewService) ListRequests(ctx context.Context, statusFilter *string, offset, limit int) ([]models.ReviewRequest, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var status *models.ReviewRequestStatus
	if statusFilter != nil {
		st := models.ReviewRequestStatus(*statusFilter)
		if models.ValidReviewRequestStatuses[st] {
			status = &st
		}
	}
	return s.repo.ListRequests(ctx, status, offset, limit)
}

func (s *ReviewService) ListMyAssignments(ctx context.Context, assigneeID uuid.UUID, offset, limit int) ([]models.ReviewRequest, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByAssignee(ctx, assigneeID, offset, limit)
}

func (s *ReviewService) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.ReviewRequest, error) {
	return s.repo.ListByEntity(ctx, entityType, entityID)
}

func (s *ReviewService) ListFollowUpRequests(ctx context.Context, parentID uuid.UUID) ([]models.ReviewRequest, error) {
	return s.repo.ListFollowUps(ctx, parentID)
}

func (s *ReviewService) ListLevels(ctx context.Context, requestID uuid.UUID) ([]models.ReviewLevel, error) {
	return s.repo.ListLevelsByRequest(ctx, requestID)
}

// Assign a reviewer to a specific level

func (s *ReviewService) AssignLevel(ctx context.Context, levelID uuid.UUID, req *dto.AssignLevelRequest, actorID uuid.UUID) (*models.ReviewLevel, error) {
	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		return nil, fmt.Errorf("invalid assignee_id")
	}

	level, err := s.repo.GetLevelByID(ctx, levelID)
	if err != nil {
		return nil, ErrReviewLevelNotFound
	}

	level.AssigneeID = &assigneeID
	if err := s.repo.UpdateLevel(ctx, level); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "review_level", levelID, actorID, "assigned", map[string]interface{}{
		"request_id":  level.RequestID,
		"level":       level.Level,
		"assignee_id": assigneeID,
	})

	return level, nil
}

// Decide on a level: Approved, Rejected, or Returned

func (s *ReviewService) DecideLevel(ctx context.Context, levelID uuid.UUID, req *dto.DecideLevelRequest, deciderID uuid.UUID) (*models.ReviewRequest, *models.ReviewLevel, error) {
	if req.Decision == "" {
		return nil, nil, ErrDecisionRequired
	}

	decision := models.ReviewLevelStatus(req.Decision)
	if decision != models.LevelApproved && decision != models.LevelRejected && decision != models.LevelReturned {
		return nil, nil, ErrDecisionRequired
	}

	level, err := s.repo.GetLevelByID(ctx, levelID)
	if err != nil {
		return nil, nil, ErrReviewLevelNotFound
	}

	if level.Status != models.LevelPending {
		return nil, nil, ErrLevelNotPending
	}

	request, err := s.repo.GetRequestByID(ctx, level.RequestID)
	if err != nil {
		return nil, nil, ErrReviewRequestNotFound
	}

	if request.Status != models.ReviewRequestInReview {
		return nil, nil, ErrRequestNotInReview
	}

	if level.Level != request.CurrentLevel {
		return nil, nil, ErrNotCurrentLevel
	}

	// If assignee is set, verify the decider is the assignee
	if level.AssigneeID != nil && *level.AssigneeID != deciderID {
		return nil, nil, ErrNotAssignee
	}

	// Update level
	now := time.Now()
	level.Status = decision
	level.Decision = &req.Decision
	level.Annotation = req.Annotation
	level.DecidedBy = &deciderID
	level.DecidedAt = &now

	if err := s.repo.UpdateLevel(ctx, level); err != nil {
		return nil, nil, err
	}

	s.audit.Log(ctx, "review_level", levelID, deciderID, "decided", map[string]interface{}{
		"request_id": request.ID,
		"level":      level.Level,
		"decision":   req.Decision,
		"annotation": req.Annotation,
	})

	// Update request based on decision
	switch decision {
	case models.LevelApproved:
		if level.Level >= request.RequiredLevels {
			// Final level approved — mark request as approved
			approvedStr := "Approved"
			request.Status = models.ReviewRequestApproved
			request.FinalDecision = &approvedStr
			s.audit.Log(ctx, "review_request", request.ID, deciderID, "final_approved", map[string]interface{}{
				"entity_type": request.EntityType,
				"entity_id":   request.EntityID,
			})
			// Disposition write-back
			s.executeDisposition(ctx, request.EntityType, request.EntityID, "Approved")
		} else {
			// Advance to next level
			request.CurrentLevel = level.Level + 1
		}

	case models.LevelRejected:
		// Any rejection terminates the request
		rejectedStr := "Rejected"
		request.Status = models.ReviewRequestRejected
		request.FinalDecision = &rejectedStr
		s.audit.Log(ctx, "review_request", request.ID, deciderID, "final_rejected", map[string]interface{}{
			"entity_type": request.EntityType,
			"entity_id":   request.EntityID,
			"at_level":    level.Level,
		})
		// Disposition write-back
		s.executeDisposition(ctx, request.EntityType, request.EntityID, "Rejected")

	case models.LevelReturned:
		// Return for supplement — request goes to Returned status
		request.Status = models.ReviewRequestReturned
		s.audit.Log(ctx, "review_request", request.ID, deciderID, "returned_for_supplement", map[string]interface{}{
			"entity_type": request.EntityType,
			"entity_id":   request.EntityID,
			"at_level":    level.Level,
			"annotation":  req.Annotation,
		})
	}

	if err := s.repo.UpdateRequest(ctx, request); err != nil {
		return nil, nil, err
	}

	return request, level, nil
}

// ResubmitAfterReturn re-opens a returned request at the level it was returned from
func (s *ReviewService) ResubmitAfterReturn(ctx context.Context, requestID, actorID uuid.UUID) (*models.ReviewRequest, error) {
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, ErrReviewRequestNotFound
	}

	if request.Status != models.ReviewRequestReturned {
		return nil, fmt.Errorf("only returned requests can be resubmitted")
	}

	// Reset the returned level back to Pending
	level, err := s.repo.GetLevelByRequestAndLevel(ctx, requestID, request.CurrentLevel)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	level.Status = models.LevelPending
	level.Decision = nil
	level.Annotation = nil
	level.DecidedBy = nil
	level.DecidedAt = nil
	level.UpdatedAt = now

	if err := s.repo.UpdateLevel(ctx, level); err != nil {
		return nil, err
	}

	request.Status = models.ReviewRequestInReview
	if err := s.repo.UpdateRequest(ctx, request); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "review_request", requestID, actorID, "resubmitted", map[string]interface{}{
		"level": request.CurrentLevel,
	})

	return request, nil
}

// Follow-up records

func (s *ReviewService) AddFollowUp(ctx context.Context, requestID uuid.UUID, req *dto.CreateFollowUpRequest, authorID uuid.UUID) (*models.ReviewFollowUp, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, ErrReviewRequestNotFound
	}

	fu := &models.ReviewFollowUp{
		ID:        uuid.New(),
		RequestID: requestID,
		AuthorID:  authorID,
		Content:   req.Content,
		Level:     request.CurrentLevel,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateFollowUp(ctx, fu); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "review_follow_up", fu.ID, authorID, "added", map[string]interface{}{
		"request_id": requestID,
		"level":      fu.Level,
	})

	return fu, nil
}

func (s *ReviewService) ListFollowUps(ctx context.Context, requestID uuid.UUID) ([]models.ReviewFollowUp, error) {
	return s.repo.ListFollowUpsByRequest(ctx, requestID)
}

// Disposition write-back to originating record
func (s *ReviewService) executeDisposition(ctx context.Context, entityType string, entityID uuid.UUID, decision string) {
	if cb, ok := s.dispositionCallbacks[entityType]; ok {
		cb(ctx, entityType, entityID, decision)
	}
}

// ExecuteDispositionPublic exposes executeDisposition for testing
func (s *ReviewService) ExecuteDispositionPublic(ctx context.Context, entityType string, entityID uuid.UUID, decision string) {
	s.executeDisposition(ctx, entityType, entityID, decision)
}
