package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

// Config DTOs

type CreateReviewConfigRequest struct {
	ReviewType     string  `json:"review_type"`
	Description    *string `json:"description,omitempty"`
	RequiredLevels int     `json:"required_levels"`
}

type UpdateReviewConfigRequest struct {
	Description    *string `json:"description,omitempty"`
	RequiredLevels *int    `json:"required_levels,omitempty"`
}

type ReviewConfigResponse struct {
	ID             uuid.UUID `json:"id"`
	ReviewType     string    `json:"review_type"`
	Description    *string   `json:"description,omitempty"`
	RequiredLevels int       `json:"required_levels"`
	CreatedBy      uuid.UUID `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func ToReviewConfigResponse(c *models.ReviewConfig) ReviewConfigResponse {
	return ReviewConfigResponse{
		ID:             c.ID,
		ReviewType:     c.ReviewType,
		Description:    c.Description,
		RequiredLevels: c.RequiredLevels,
		CreatedBy:      c.CreatedBy,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func ToReviewConfigResponseList(configs []models.ReviewConfig) []ReviewConfigResponse {
	result := make([]ReviewConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = ToReviewConfigResponse(&c)
	}
	return result
}

// Request DTOs

type SubmitReviewRequest struct {
	ReviewType string  `json:"review_type"`
	EntityType string  `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	ParentID   *string `json:"parent_id,omitempty"`
}

type ReviewRequestResponse struct {
	ID             uuid.UUID            `json:"id"`
	ReviewType     string               `json:"review_type"`
	EntityType     string               `json:"entity_type"`
	EntityID       uuid.UUID            `json:"entity_id"`
	RequiredLevels int                  `json:"required_levels"`
	CurrentLevel   int                  `json:"current_level"`
	Status         string               `json:"status"`
	SubmittedBy    uuid.UUID            `json:"submitted_by"`
	FinalDecision  *string              `json:"final_decision,omitempty"`
	ParentID       *uuid.UUID           `json:"parent_id,omitempty"`
	Levels         []ReviewLevelResponse `json:"levels,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func ToReviewRequestResponse(r *models.ReviewRequest, levels []models.ReviewLevel) ReviewRequestResponse {
	resp := ReviewRequestResponse{
		ID:             r.ID,
		ReviewType:     r.ReviewType,
		EntityType:     r.EntityType,
		EntityID:       r.EntityID,
		RequiredLevels: r.RequiredLevels,
		CurrentLevel:   r.CurrentLevel,
		Status:         string(r.Status),
		SubmittedBy:    r.SubmittedBy,
		FinalDecision:  r.FinalDecision,
		ParentID:       r.ParentID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if levels != nil {
		resp.Levels = ToReviewLevelResponseList(levels)
	}
	return resp
}

func ToReviewRequestResponseList(requests []models.ReviewRequest) []ReviewRequestResponse {
	result := make([]ReviewRequestResponse, len(requests))
	for i, r := range requests {
		result[i] = ToReviewRequestResponse(&r, nil)
	}
	return result
}

// Level DTOs

type AssignLevelRequest struct {
	AssigneeID string `json:"assignee_id"`
}

type DecideLevelRequest struct {
	Decision   string  `json:"decision"`
	Annotation *string `json:"annotation,omitempty"`
}

type ReviewLevelResponse struct {
	ID         uuid.UUID  `json:"id"`
	RequestID  uuid.UUID  `json:"request_id"`
	Level      int        `json:"level"`
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`
	Status     string     `json:"status"`
	Decision   *string    `json:"decision,omitempty"`
	Annotation *string    `json:"annotation,omitempty"`
	DecidedBy  *uuid.UUID `json:"decided_by,omitempty"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func ToReviewLevelResponse(l *models.ReviewLevel) ReviewLevelResponse {
	return ReviewLevelResponse{
		ID:         l.ID,
		RequestID:  l.RequestID,
		Level:      l.Level,
		AssigneeID: l.AssigneeID,
		Status:     string(l.Status),
		Decision:   l.Decision,
		Annotation: l.Annotation,
		DecidedBy:  l.DecidedBy,
		DecidedAt:  l.DecidedAt,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

func ToReviewLevelResponseList(levels []models.ReviewLevel) []ReviewLevelResponse {
	result := make([]ReviewLevelResponse, len(levels))
	for i, l := range levels {
		result[i] = ToReviewLevelResponse(&l)
	}
	return result
}

// Follow-up DTOs

type CreateFollowUpRequest struct {
	Content string `json:"content"`
}

type FollowUpResponse struct {
	ID        uuid.UUID `json:"id"`
	RequestID uuid.UUID `json:"request_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	Level     int       `json:"level"`
	CreatedAt time.Time `json:"created_at"`
}

func ToFollowUpResponse(f *models.ReviewFollowUp) FollowUpResponse {
	return FollowUpResponse{
		ID:        f.ID,
		RequestID: f.RequestID,
		AuthorID:  f.AuthorID,
		Content:   f.Content,
		Level:     f.Level,
		CreatedAt: f.CreatedAt,
	}
}

func ToFollowUpResponseList(followUps []models.ReviewFollowUp) []FollowUpResponse {
	result := make([]FollowUpResponse, len(followUps))
	for i, f := range followUps {
		result[i] = ToFollowUpResponse(&f)
	}
	return result
}
