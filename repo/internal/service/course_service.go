package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrCourseNotFound    = errors.New("course not found")
	ErrNodeNotFound      = errors.New("outline node not found")
	ErrNotCourseMember   = errors.New("not a member of this course")
	ErrNotCourseStaff    = errors.New("must be course staff to perform this action")
	ErrInvalidNodeType   = errors.New("invalid node type")
	ErrInvalidVisibility = errors.New("access denied based on resource visibility")
)

type CourseService struct {
	courseRepo     *repository.CourseRepository
	outlineRepo   *repository.CourseOutlineRepository
	membershipRepo *repository.CourseMembershipRepository
	audit          *AuditService
}

func NewCourseService(
	courseRepo *repository.CourseRepository,
	outlineRepo *repository.CourseOutlineRepository,
	membershipRepo *repository.CourseMembershipRepository,
	audit *AuditService,
) *CourseService {
	return &CourseService{
		courseRepo:     courseRepo,
		outlineRepo:   outlineRepo,
		membershipRepo: membershipRepo,
		audit:          audit,
	}
}

func (s *CourseService) CreateCourse(ctx context.Context, req *dto.CreateCourseRequest, actorID uuid.UUID) (*models.Course, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	now := time.Now()
	course := &models.Course{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      models.CourseStatusDraft,
		CreatedBy:   actorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.courseRepo.Create(ctx, course); err != nil {
		return nil, err
	}

	// Auto-add creator as Staff
	membership := &models.CourseMembership{
		ID:        uuid.New(),
		CourseID:  course.ID,
		AccountID: actorID,
		Role:      models.MembershipRoleStaff,
		CreatedAt: now,
	}
	s.membershipRepo.Create(ctx, membership)

	s.audit.Log(ctx, "course", course.ID, actorID, "created", map[string]interface{}{
		"title": course.Title,
	})

	return course, nil
}

func (s *CourseService) GetCourse(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole models.Role) (*models.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourseNotFound
	}
	// Published courses are visible to all authenticated users
	if course.Status == models.CourseStatusPublished {
		return course, nil
	}
	// Non-published: require membership or admin
	if _, err := s.CheckAccess(ctx, id, callerID, callerRole); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *CourseService) ListCourses(ctx context.Context, offset, limit int, callerRole models.Role) ([]models.Course, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	// Administrators and Instructors see all; others see only published
	if callerRole == models.RoleAdministrator || callerRole == models.RoleInstructor {
		return s.courseRepo.List(ctx, offset, limit)
	}
	return s.courseRepo.ListPublished(ctx, offset, limit)
}

func (s *CourseService) UpdateCourse(ctx context.Context, id uuid.UUID, req *dto.UpdateCourseRequest, actorID uuid.UUID) (*models.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourseNotFound
	}

	if err := s.requireStaff(ctx, id, actorID); err != nil {
		return nil, err
	}

	// Capture state before mutation for audit snapshot
	origTitle := course.Title
	origStatus := string(course.Status)

	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.Description != nil {
		course.Description = req.Description
	}
	if req.Status != nil {
		status := models.CourseStatus(*req.Status)
		if !models.ValidCourseStatuses[status] {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		course.Status = status
	}

	beforeSnapshot := map[string]interface{}{
		"title":  origTitle,
		"status": origStatus,
	}

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	s.audit.LogExtended(ctx, &AuditEntry{
		EntityType:     "course",
		EntityID:       course.ID,
		ActorID:        actorID,
		Action:         "updated",
		Tier:           models.TierAudit,
		BeforeSnapshot: beforeSnapshot,
		AfterSnapshot: map[string]interface{}{
			"title":  course.Title,
			"status": string(course.Status),
		},
		Details: map[string]interface{}{
			"title":  course.Title,
			"status": string(course.Status),
		},
	})

	return course, nil
}

// Outline operations

func (s *CourseService) CreateOutlineNode(ctx context.Context, req *dto.CreateOutlineNodeRequest, actorID uuid.UUID) (*models.CourseOutlineNode, error) {
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return nil, fmt.Errorf("invalid course_id")
	}
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		return nil, ErrCourseNotFound
	}
	if err := s.requireStaff(ctx, courseID, actorID); err != nil {
		return nil, err
	}

	nodeType := models.OutlineNodeType(req.NodeType)
	if !models.ValidNodeTypes[nodeType] {
		return nil, ErrInvalidNodeType
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id")
		}
		parentID = &pid
		// Verify parent exists and belongs to same course
		parent, err := s.outlineRepo.GetByID(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("parent node not found")
		}
		if parent.CourseID != courseID {
			return nil, fmt.Errorf("parent node belongs to a different course")
		}
	}

	now := time.Now()
	node := &models.CourseOutlineNode{
		ID:          uuid.New(),
		CourseID:    courseID,
		ParentID:    parentID,
		NodeType:    nodeType,
		Title:       req.Title,
		Description: req.Description,
		OrderIndex:  req.OrderIndex,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.outlineRepo.Create(ctx, node); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "course_outline_node", node.ID, actorID, "created", map[string]interface{}{
		"course_id": courseID,
		"node_type": string(nodeType),
		"title":     req.Title,
	})

	return node, nil
}

func (s *CourseService) GetOutlineTree(ctx context.Context, courseID, callerID uuid.UUID, callerRole models.Role) ([]dto.OutlineTreeNode, error) {
	if _, err := s.CheckAccess(ctx, courseID, callerID, callerRole); err != nil {
		return nil, err
	}
	nodes, err := s.outlineRepo.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	return dto.BuildOutlineTree(nodes), nil
}

func (s *CourseService) UpdateOutlineNode(ctx context.Context, nodeID uuid.UUID, req *dto.UpdateOutlineNodeRequest, actorID uuid.UUID) (*models.CourseOutlineNode, error) {
	node, err := s.outlineRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, ErrNodeNotFound
	}
	if err := s.requireStaff(ctx, node.CourseID, actorID); err != nil {
		return nil, err
	}

	if req.Title != nil {
		node.Title = *req.Title
	}
	if req.Description != nil {
		node.Description = req.Description
	}
	if req.OrderIndex != nil {
		node.OrderIndex = *req.OrderIndex
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			node.ParentID = nil
		} else {
			pid, err := uuid.Parse(*req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("invalid parent_id")
			}
			node.ParentID = &pid
		}
	}

	if err := s.outlineRepo.Update(ctx, node); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "course_outline_node", nodeID, actorID, "updated", map[string]interface{}{
		"title":       node.Title,
		"order_index": node.OrderIndex,
	})

	return node, nil
}

func (s *CourseService) DeleteOutlineNode(ctx context.Context, nodeID, actorID uuid.UUID) error {
	node, err := s.outlineRepo.GetByID(ctx, nodeID)
	if err != nil {
		return ErrNodeNotFound
	}
	if err := s.requireStaff(ctx, node.CourseID, actorID); err != nil {
		return err
	}

	if err := s.outlineRepo.Delete(ctx, nodeID); err != nil {
		return err
	}

	s.audit.Log(ctx, "course_outline_node", nodeID, actorID, "deleted", map[string]interface{}{
		"course_id": node.CourseID,
		"title":     node.Title,
	})

	return nil
}

// Membership operations

func (s *CourseService) AddMember(ctx context.Context, courseID uuid.UUID, req *dto.AddMemberRequest, actorID uuid.UUID) (*models.CourseMembership, error) {
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		return nil, ErrCourseNotFound
	}
	if err := s.requireStaff(ctx, courseID, actorID); err != nil {
		return nil, err
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account_id")
	}
	role := models.MembershipRole(req.Role)
	if !models.ValidMembershipRoles[role] {
		return nil, fmt.Errorf("invalid membership role: %s", req.Role)
	}

	membership := &models.CourseMembership{
		ID:        uuid.New(),
		CourseID:  courseID,
		AccountID: accountID,
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "course_membership", membership.ID, actorID, "added", map[string]interface{}{
		"course_id":  courseID,
		"account_id": accountID,
		"role":       string(role),
	})

	return membership, nil
}

func (s *CourseService) ListMembers(ctx context.Context, courseID, actorID uuid.UUID, callerRole models.Role) ([]models.CourseMembership, error) {
	if err := s.requireStaffOrAdmin(ctx, courseID, actorID, callerRole); err != nil {
		return nil, err
	}
	return s.membershipRepo.ListByCourse(ctx, courseID)
}

func (s *CourseService) RemoveMember(ctx context.Context, membershipID, actorID uuid.UUID, callerRole models.Role, courseID uuid.UUID) error {
	if err := s.requireStaffOrAdmin(ctx, courseID, actorID, callerRole); err != nil {
		return err
	}
	return s.membershipRepo.Delete(ctx, membershipID)
}

// CheckAccess returns the membership if the user has access, or nil if the user is an admin
func (s *CourseService) CheckAccess(ctx context.Context, courseID, accountID uuid.UUID, callerRole models.Role) (*models.CourseMembership, error) {
	// Administrators always have access
	if callerRole == models.RoleAdministrator {
		return nil, nil
	}
	m, err := s.membershipRepo.GetByAccountAndCourse(ctx, accountID, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotCourseMember
		}
		return nil, err
	}
	return m, nil
}

func (s *CourseService) requireStaff(ctx context.Context, courseID, actorID uuid.UUID) error {
	m, err := s.membershipRepo.GetByAccountAndCourse(ctx, actorID, courseID)
	if err != nil {
		return ErrNotCourseStaff
	}
	if m.Role != models.MembershipRoleStaff {
		return ErrNotCourseStaff
	}
	return nil
}

func (s *CourseService) requireStaffOrAdmin(ctx context.Context, courseID, actorID uuid.UUID, callerRole models.Role) error {
	if callerRole == models.RoleAdministrator {
		return nil
	}
	return s.requireStaff(ctx, courseID, actorID)
}
