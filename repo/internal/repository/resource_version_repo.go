package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ResourceVersionRepository struct {
	db *sqlx.DB
}

func NewResourceVersionRepository(db *sqlx.DB) *ResourceVersionRepository {
	return &ResourceVersionRepository{db: db}
}

func (r *ResourceVersionRepository) Create(ctx context.Context, v *models.ResourceVersion) error {
	query := `INSERT INTO resource_versions (id, resource_id, version_number, file_name, mime_type, size_bytes, sha256_hash, storage_path, extracted_text, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.ResourceID, v.VersionNumber, v.FileName, v.MimeType,
		v.SizeBytes, v.SHA256Hash, v.StoragePath, v.ExtractedText,
		v.UploadedBy, v.CreatedAt)
	if err != nil {
		return fmt.Errorf("resource_version_repo.Create: %w", err)
	}
	return nil
}

func (r *ResourceVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ResourceVersion, error) {
	var v models.ResourceVersion
	query := `SELECT id, resource_id, version_number, file_name, mime_type, size_bytes, sha256_hash, storage_path, extracted_text, uploaded_by, created_at
		FROM resource_versions WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&v); err != nil {
		return nil, fmt.Errorf("resource_version_repo.GetByID: %w", err)
	}
	return &v, nil
}

func (r *ResourceVersionRepository) ListByResource(ctx context.Context, resourceID uuid.UUID) ([]models.ResourceVersion, error) {
	var versions []models.ResourceVersion
	query := `SELECT id, resource_id, version_number, file_name, mime_type, size_bytes, sha256_hash, storage_path, extracted_text, uploaded_by, created_at
		FROM resource_versions WHERE resource_id = $1 ORDER BY version_number DESC`
	if err := r.db.SelectContext(ctx, &versions, query, resourceID); err != nil {
		return nil, fmt.Errorf("resource_version_repo.ListByResource: %w", err)
	}
	return versions, nil
}

func (r *ResourceVersionRepository) GetLatestVersionNumber(ctx context.Context, resourceID uuid.UUID) (int, error) {
	var num *int
	query := `SELECT MAX(version_number) FROM resource_versions WHERE resource_id = $1`
	if err := r.db.QueryRowxContext(ctx, query, resourceID).Scan(&num); err != nil {
		return 0, fmt.Errorf("resource_version_repo.GetLatestVersionNumber: %w", err)
	}
	if num == nil {
		return 0, nil
	}
	return *num, nil
}

func (r *ResourceVersionRepository) GetBySHA256(ctx context.Context, sha256Hash string) (*models.ResourceVersion, error) {
	var v models.ResourceVersion
	query := `SELECT id, resource_id, version_number, file_name, mime_type, size_bytes, sha256_hash, storage_path, extracted_text, uploaded_by, created_at
		FROM resource_versions WHERE sha256_hash = $1 LIMIT 1`
	if err := r.db.QueryRowxContext(ctx, query, sha256Hash).StructScan(&v); err != nil {
		return nil, fmt.Errorf("resource_version_repo.GetBySHA256: %w", err)
	}
	return &v, nil
}
