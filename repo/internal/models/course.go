package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "Draft"
	CourseStatusPublished CourseStatus = "Published"
	CourseStatusArchived  CourseStatus = "Archived"
)

var ValidCourseStatuses = map[CourseStatus]bool{
	CourseStatusDraft:     true,
	CourseStatusPublished: true,
	CourseStatusArchived:  true,
}

func (s *CourseStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("course status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("course status must be a string")
		}
		sv = string(bv)
	}
	*s = CourseStatus(sv)
	return nil
}

func (s CourseStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type Course struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Title       string       `db:"title" json:"title"`
	Description *string      `db:"description" json:"description,omitempty"`
	Status      CourseStatus `db:"status" json:"status"`
	CreatedBy   uuid.UUID    `db:"created_by" json:"created_by"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}
