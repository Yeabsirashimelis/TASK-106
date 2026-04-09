package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ReportStatus string

const (
	ReportOpen        ReportStatus = "Open"
	ReportUnderReview ReportStatus = "Under Review"
	ReportResolved    ReportStatus = "Resolved"
	ReportDismissed   ReportStatus = "Dismissed"
)

var ValidReportStatuses = map[ReportStatus]bool{
	ReportOpen:        true,
	ReportUnderReview: true,
	ReportResolved:    true,
	ReportDismissed:   true,
}

var ValidReportTransitions = map[ReportStatus][]ReportStatus{
	ReportOpen:        {ReportUnderReview, ReportDismissed},
	ReportUnderReview: {ReportResolved, ReportDismissed},
	ReportResolved:    {},
	ReportDismissed:   {},
}

func CanTransitionReport(from, to ReportStatus) bool {
	allowed, ok := ValidReportTransitions[from]
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

func (s *ReportStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("report status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("report status must be a string")
		}
		sv = string(bv)
	}
	*s = ReportStatus(sv)
	return nil
}

func (s ReportStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type ReportCategory string

const (
	CategorySpam          ReportCategory = "Spam"
	CategoryHarassment    ReportCategory = "Harassment"
	CategoryInappropriate ReportCategory = "Inappropriate Content"
	CategoryPolicyViol    ReportCategory = "Policy Violation"
	CategoryOther         ReportCategory = "Other"
)

var ValidReportCategories = map[ReportCategory]bool{
	CategorySpam:          true,
	CategoryHarassment:    true,
	CategoryInappropriate: true,
	CategoryPolicyViol:    true,
	CategoryOther:         true,
}

func (c *ReportCategory) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("report category cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("report category must be a string")
		}
		sv = string(bv)
	}
	*c = ReportCategory(sv)
	return nil
}

func (c ReportCategory) Value() (driver.Value, error) {
	return string(c), nil
}

type Report struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	ReporterID  uuid.UUID      `db:"reporter_id" json:"reporter_id"`
	TargetType  string         `db:"target_type" json:"target_type"`
	TargetID    uuid.UUID      `db:"target_id" json:"target_id"`
	Category    ReportCategory `db:"category" json:"category"`
	Description string         `db:"description" json:"description"`
	Status      ReportStatus   `db:"status" json:"status"`
	AssignedTo  *uuid.UUID     `db:"assigned_to" json:"assigned_to,omitempty"`
	Resolution  *string        `db:"resolution" json:"resolution,omitempty"`
	ResolvedAt  *time.Time     `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

type ReportEvidence struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ReportID    uuid.UUID `db:"report_id" json:"report_id"`
	FileName    string    `db:"file_name" json:"file_name"`
	MimeType    string    `db:"mime_type" json:"mime_type"`
	SizeBytes   int64     `db:"size_bytes" json:"size_bytes"`
	SHA256Hash  string    `db:"sha256_hash" json:"sha256_hash"`
	StoragePath string    `db:"storage_path" json:"-"`
	UploadedBy  uuid.UUID `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type ReportNote struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ReportID  uuid.UUID `db:"report_id" json:"report_id"`
	AuthorID  uuid.UUID `db:"author_id" json:"author_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
