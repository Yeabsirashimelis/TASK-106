package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "Pending"
	ReviewApproved ReviewStatus = "Approved"
	ReviewRejected ReviewStatus = "Rejected"
)

var ValidReviewStatuses = map[ReviewStatus]bool{
	ReviewPending:  true,
	ReviewApproved: true,
	ReviewRejected: true,
}

var ValidReviewTransitions = map[ReviewStatus][]ReviewStatus{
	ReviewPending:  {ReviewApproved, ReviewRejected},
	ReviewApproved: {},
	ReviewRejected: {},
}

func CanTransitionReview(from, to ReviewStatus) bool {
	allowed, ok := ValidReviewTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (s *ReviewStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("review status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("review status must be a string")
		}
		sv = string(bv)
	}
	*s = ReviewStatus(sv)
	return nil
}

func (s ReviewStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type ModerationReview struct {
	ID             uuid.UUID    `db:"id" json:"id"`
	ContentType    string       `db:"content_type" json:"content_type"`
	ContentID      uuid.UUID    `db:"content_id" json:"content_id"`
	ContentSnippet *string      `db:"content_snippet" json:"content_snippet,omitempty"`
	Status         ReviewStatus `db:"status" json:"status"`
	ModeratorID    *uuid.UUID   `db:"moderator_id" json:"moderator_id,omitempty"`
	Reason         *string      `db:"reason" json:"reason,omitempty"`
	DecidedAt      *time.Time   `db:"decided_at" json:"decided_at,omitempty"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at" json:"updated_at"`
}
