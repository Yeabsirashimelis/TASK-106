package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ResourceTagRepository struct {
	db *sqlx.DB
}

func NewResourceTagRepository(db *sqlx.DB) *ResourceTagRepository {
	return &ResourceTagRepository{db: db}
}

func (r *ResourceTagRepository) Create(ctx context.Context, tag *models.ResourceTag) error {
	query := `INSERT INTO resource_tags (id, resource_id, tag, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, tag.ID, tag.ResourceID, tag.Tag, tag.CreatedAt)
	if err != nil {
		return fmt.Errorf("resource_tag_repo.Create: %w", err)
	}
	return nil
}

func (r *ResourceTagRepository) ListByResource(ctx context.Context, resourceID uuid.UUID) ([]models.ResourceTag, error) {
	var tags []models.ResourceTag
	query := `SELECT id, resource_id, tag, created_at FROM resource_tags WHERE resource_id = $1 ORDER BY tag`
	if err := r.db.SelectContext(ctx, &tags, query, resourceID); err != nil {
		return nil, fmt.Errorf("resource_tag_repo.ListByResource: %w", err)
	}
	return tags, nil
}

func (r *ResourceTagRepository) CountByResource(ctx context.Context, resourceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM resource_tags WHERE resource_id = $1`
	if err := r.db.QueryRowxContext(ctx, query, resourceID).Scan(&count); err != nil {
		return 0, fmt.Errorf("resource_tag_repo.CountByResource: %w", err)
	}
	return count, nil
}

func (r *ResourceTagRepository) DeleteByResourceAndTag(ctx context.Context, resourceID uuid.UUID, tag string) error {
	query := `DELETE FROM resource_tags WHERE resource_id = $1 AND tag = $2`
	_, err := r.db.ExecContext(ctx, query, resourceID, tag)
	if err != nil {
		return fmt.Errorf("resource_tag_repo.DeleteByResourceAndTag: %w", err)
	}
	return nil
}

func (r *ResourceTagRepository) ReplaceAll(ctx context.Context, resourceID uuid.UUID, tags []models.ResourceTag) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("resource_tag_repo.ReplaceAll: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM resource_tags WHERE resource_id = $1`, resourceID)
	if err != nil {
		return fmt.Errorf("resource_tag_repo.ReplaceAll: %w", err)
	}

	for _, tag := range tags {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO resource_tags (id, resource_id, tag, created_at) VALUES ($1, $2, $3, $4)`,
			tag.ID, tag.ResourceID, tag.Tag, tag.CreatedAt)
		if err != nil {
			return fmt.Errorf("resource_tag_repo.ReplaceAll: %w", err)
		}
	}

	return tx.Commit()
}
