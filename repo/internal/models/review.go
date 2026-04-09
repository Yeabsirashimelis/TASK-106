package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReviewConfig defines how many approval levels a review type requires

type ReviewConfig struct {
	ID             uuid.UUID `db:"id" json:"id"`
	ReviewType     string    `db:"review_type" json:"review_type"`
	Description    *string   `db:"description" json:"description,omitempty"`
	RequiredLevels int       `db:"required_levels" json:"required_levels"`
	CreatedBy      uuid.UUID `db:"created_by" json:"created_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// ReviewRequest is the top-level review record linked to an entity

type ReviewRequestStatus string

const (
	ReviewRequestInReview ReviewRequestStatus = "In Review"
	ReviewRequestApproved ReviewRequestStatus = "Approved"
	ReviewRequestRejected ReviewRequestStatus = "Rejected"
	ReviewRequestReturned ReviewRequestStatus = "Returned"
)

var ValidReviewRequestStatuses = map[ReviewRequestStatus]bool{
	ReviewRequestInReview: true,
	ReviewRequestApproved: true,
	ReviewRequestRejected: true,
	ReviewRequestReturned: true,
}

func (s *ReviewRequestStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("review request status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("review request status must be a string")
		}
		sv = string(bv)
	}
	*s = ReviewRequestStatus(sv)
	return nil
}

func (s ReviewRequestStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type ReviewRequest struct {
	ID             uuid.UUID           `db:"id" json:"id"`
	ReviewType     string              `db:"review_type" json:"review_type"`
	EntityType     string              `db:"entity_type" json:"entity_type"`
	EntityID       uuid.UUID           `db:"entity_id" json:"entity_id"`
	RequiredLevels int                 `db:"required_levels" json:"required_levels"`
	CurrentLevel   int                 `db:"current_level" json:"current_level"`
	Status         ReviewRequestStatus `db:"status" json:"status"`
	SubmittedBy    uuid.UUID           `db:"submitted_by" json:"submitted_by"`
	FinalDecision  *string             `db:"final_decision" json:"final_decision,omitempty"`
	ParentID       *uuid.UUID          `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt      time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `db:"updated_at" json:"updated_at"`
}

// ReviewLevel tracks the decision at each approval level

type ReviewLevelStatus string

const (
	LevelPending  ReviewLevelStatus = "Pending"
	LevelApproved ReviewLevelStatus = "Approved"
	LevelRejected ReviewLevelStatus = "Rejected"
	LevelReturned ReviewLevelStatus = "Returned"
)

var ValidLevelStatuses = map[ReviewLevelStatus]bool{
	LevelPending:  true,
	LevelApproved: true,
	LevelRejected: true,
	LevelReturned: true,
}

func (s *ReviewLevelStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("review level status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("review level status must be a string")
		}
		sv = string(bv)
	}
	*s = ReviewLevelStatus(sv)
	return nil
}

func (s ReviewLevelStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type ReviewLevel struct {
	ID         uuid.UUID         `db:"id" json:"id"`
	RequestID  uuid.UUID         `db:"request_id" json:"request_id"`
	Level      int               `db:"level" json:"level"`
	AssigneeID *uuid.UUID        `db:"assignee_id" json:"assignee_id,omitempty"`
	Status     ReviewLevelStatus `db:"status" json:"status"`
	Decision   *string           `db:"decision" json:"decision,omitempty"`
	Annotation *string           `db:"annotation" json:"annotation,omitempty"`
	DecidedBy  *uuid.UUID        `db:"decided_by" json:"decided_by,omitempty"`
	DecidedAt  *time.Time        `db:"decided_at" json:"decided_at,omitempty"`
	CreatedAt  time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time         `db:"updated_at" json:"updated_at"`
}

// ReviewFollowUp is supplementary material linked to a review

type ReviewFollowUp struct {
	ID        uuid.UUID `db:"id" json:"id"`
	RequestID uuid.UUID `db:"request_id" json:"request_id"`
	AuthorID  uuid.UUID `db:"author_id" json:"author_id"`
	Content   string    `db:"content" json:"content"`
	Level     int       `db:"level" json:"level"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
