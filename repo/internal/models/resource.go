package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	ResourceTypeDocument ResourceType = "Document"
	ResourceTypeVideo    ResourceType = "Video"
	ResourceTypeLink     ResourceType = "Link"
)

var ValidResourceTypes = map[ResourceType]bool{
	ResourceTypeDocument: true,
	ResourceTypeVideo:    true,
	ResourceTypeLink:     true,
}

func (t *ResourceType) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("resource type cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("resource type must be a string")
		}
		sv = string(bv)
	}
	*t = ResourceType(sv)
	return nil
}

func (t ResourceType) Value() (driver.Value, error) {
	return string(t), nil
}

type ResourceVisibility string

const (
	VisibilityStaff    ResourceVisibility = "Staff"
	VisibilityEnrolled ResourceVisibility = "Enrolled"
)

var ValidVisibilities = map[ResourceVisibility]bool{
	VisibilityStaff:    true,
	VisibilityEnrolled: true,
}

func (v *ResourceVisibility) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("visibility cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("visibility must be a string")
		}
		sv = string(bv)
	}
	*v = ResourceVisibility(sv)
	return nil
}

func (v ResourceVisibility) Value() (driver.Value, error) {
	return string(v), nil
}

type Resource struct {
	ID              uuid.UUID          `db:"id" json:"id"`
	CourseID        uuid.UUID          `db:"course_id" json:"course_id"`
	NodeID          *uuid.UUID         `db:"node_id" json:"node_id,omitempty"`
	Title           string             `db:"title" json:"title"`
	Description     *string            `db:"description" json:"description,omitempty"`
	ResourceType    ResourceType       `db:"resource_type" json:"resource_type"`
	Visibility      ResourceVisibility `db:"visibility" json:"visibility"`
	LinkURL         *string            `db:"link_url" json:"link_url,omitempty"`
	LatestVersionID *uuid.UUID         `db:"latest_version_id" json:"latest_version_id,omitempty"`
	CreatedBy       uuid.UUID          `db:"created_by" json:"created_by"`
	CreatedAt       time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `db:"updated_at" json:"updated_at"`
}

// AllowedMimeTypes is the allowlist for file uploads
var AllowedMimeTypes = map[string]bool{
	// Documents
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-powerpoint": true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":      true,
	"text/csv":        true,
	"application/rtf": true,
	// Videos
	"video/mp4":       true,
	"video/mpeg":      true,
	"video/webm":      true,
	"video/x-msvideo": true,
	"video/quicktime": true,
	// Images (for embedded content)
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// TextExtractableMimeTypes are types that support full-text extraction
var TextExtractableMimeTypes = map[string]bool{
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}
