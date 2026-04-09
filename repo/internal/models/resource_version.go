package models

import (
	"time"

	"github.com/google/uuid"
)

type ResourceVersion struct {
	ID            uuid.UUID `db:"id" json:"id"`
	ResourceID    uuid.UUID `db:"resource_id" json:"resource_id"`
	VersionNumber int       `db:"version_number" json:"version_number"`
	FileName      string    `db:"file_name" json:"file_name"`
	MimeType      string    `db:"mime_type" json:"mime_type"`
	SizeBytes     int64     `db:"size_bytes" json:"size_bytes"`
	SHA256Hash    string    `db:"sha256_hash" json:"sha256_hash"`
	StoragePath   string    `db:"storage_path" json:"storage_path"`
	ExtractedText *string   `db:"extracted_text" json:"-"`
	UploadedBy    uuid.UUID `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
