package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type CreateResourceRequest struct {
	CourseID     string   `json:"course_id"`
	NodeID       *string  `json:"node_id,omitempty"`
	Title        string   `json:"title"`
	Description  *string  `json:"description,omitempty"`
	ResourceType string   `json:"resource_type"`
	Visibility   string   `json:"visibility"`
	LinkURL      *string  `json:"link_url,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type UpdateResourceRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Visibility  *string  `json:"visibility,omitempty"`
	NodeID      *string  `json:"node_id,omitempty"`
	LinkURL     *string  `json:"link_url,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type ResourceResponse struct {
	ID              uuid.UUID  `json:"id"`
	CourseID        uuid.UUID  `json:"course_id"`
	NodeID          *uuid.UUID `json:"node_id,omitempty"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	ResourceType    string     `json:"resource_type"`
	Visibility      string     `json:"visibility"`
	LinkURL         *string    `json:"link_url,omitempty"`
	LatestVersionID *uuid.UUID `json:"latest_version_id,omitempty"`
	Tags            []string   `json:"tags"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func ToResourceResponse(r *models.Resource, tags []string) ResourceResponse {
	if tags == nil {
		tags = []string{}
	}
	return ResourceResponse{
		ID:              r.ID,
		CourseID:        r.CourseID,
		NodeID:          r.NodeID,
		Title:           r.Title,
		Description:     r.Description,
		ResourceType:    string(r.ResourceType),
		Visibility:      string(r.Visibility),
		LinkURL:         r.LinkURL,
		LatestVersionID: r.LatestVersionID,
		Tags:            tags,
		CreatedBy:       r.CreatedBy,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

type VersionResponse struct {
	ID            uuid.UUID `json:"id"`
	ResourceID    uuid.UUID `json:"resource_id"`
	VersionNumber int       `json:"version_number"`
	FileName      string    `json:"file_name"`
	MimeType      string    `json:"mime_type"`
	SizeBytes     int64     `json:"size_bytes"`
	SHA256Hash    string    `json:"sha256_hash"`
	UploadedBy    uuid.UUID `json:"uploaded_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func ToVersionResponse(v *models.ResourceVersion) VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		ResourceID:    v.ResourceID,
		VersionNumber: v.VersionNumber,
		FileName:      v.FileName,
		MimeType:      v.MimeType,
		SizeBytes:     v.SizeBytes,
		SHA256Hash:    v.SHA256Hash,
		UploadedBy:    v.UploadedBy,
		CreatedAt:     v.CreatedAt,
	}
}

func ToVersionResponseList(versions []models.ResourceVersion) []VersionResponse {
	result := make([]VersionResponse, len(versions))
	for i, v := range versions {
		result[i] = ToVersionResponse(&v)
	}
	return result
}
