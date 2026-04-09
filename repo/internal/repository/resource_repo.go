package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ResourceRepository struct {
	db *sqlx.DB
}

func NewResourceRepository(db *sqlx.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(ctx context.Context, res *models.Resource) error {
	query := `INSERT INTO resources (id, course_id, node_id, title, description, resource_type, visibility, link_url, latest_version_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		res.ID, res.CourseID, res.NodeID, res.Title, res.Description,
		res.ResourceType, res.Visibility, res.LinkURL, res.LatestVersionID,
		res.CreatedBy, res.CreatedAt, res.UpdatedAt)
	if err != nil {
		return fmt.Errorf("resource_repo.Create: %w", err)
	}
	return nil
}

func (r *ResourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Resource, error) {
	var res models.Resource
	query := `SELECT id, course_id, node_id, title, description, resource_type, visibility, link_url, latest_version_id, created_by, created_at, updated_at
		FROM resources WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&res); err != nil {
		return nil, fmt.Errorf("resource_repo.GetByID: %w", err)
	}
	return &res, nil
}

func (r *ResourceRepository) ListByCourse(ctx context.Context, courseID uuid.UUID, offset, limit int) ([]models.Resource, error) {
	var resources []models.Resource
	query := `SELECT id, course_id, node_id, title, description, resource_type, visibility, link_url, latest_version_id, created_by, created_at, updated_at
		FROM resources WHERE course_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &resources, query, courseID, limit, offset); err != nil {
		return nil, fmt.Errorf("resource_repo.ListByCourse: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepository) ListByCourseAndVisibility(ctx context.Context, courseID uuid.UUID, visibility models.ResourceVisibility, offset, limit int) ([]models.Resource, error) {
	var resources []models.Resource
	query := `SELECT id, course_id, node_id, title, description, resource_type, visibility, link_url, latest_version_id, created_by, created_at, updated_at
		FROM resources WHERE course_id = $1 AND visibility = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	if err := r.db.SelectContext(ctx, &resources, query, courseID, visibility, limit, offset); err != nil {
		return nil, fmt.Errorf("resource_repo.ListByCourseAndVisibility: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepository) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]models.Resource, error) {
	var resources []models.Resource
	query := `SELECT id, course_id, node_id, title, description, resource_type, visibility, link_url, latest_version_id, created_by, created_at, updated_at
		FROM resources WHERE node_id = $1 ORDER BY title`
	if err := r.db.SelectContext(ctx, &resources, query, nodeID); err != nil {
		return nil, fmt.Errorf("resource_repo.ListByNode: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepository) Update(ctx context.Context, res *models.Resource) error {
	query := `UPDATE resources SET title = $1, description = $2, visibility = $3, node_id = $4, link_url = $5, updated_at = NOW()
		WHERE id = $6`
	result, err := r.db.ExecContext(ctx, query, res.Title, res.Description, res.Visibility, res.NodeID, res.LinkURL, res.ID)
	if err != nil {
		return fmt.Errorf("resource_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("resource_repo.Update: resource not found")
	}
	return nil
}

func (r *ResourceRepository) UpdateVisibility(ctx context.Context, id uuid.UUID, visibility models.ResourceVisibility) error {
	query := `UPDATE resources SET visibility = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, visibility, id)
	if err != nil {
		return fmt.Errorf("resource_repo.UpdateVisibility: %w", err)
	}
	return nil
}

func (r *ResourceRepository) UpdateLatestVersion(ctx context.Context, resourceID, versionID uuid.UUID) error {
	query := `UPDATE resources SET latest_version_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, versionID, resourceID)
	if err != nil {
		return fmt.Errorf("resource_repo.UpdateLatestVersion: %w", err)
	}
	return nil
}

func (r *ResourceRepository) Search(ctx context.Context, courseID uuid.UUID, query string, offset, limit int) ([]models.Resource, error) {
	var resources []models.Resource
	searchQuery := `SELECT DISTINCT r.id, r.course_id, r.node_id, r.title, r.description, r.resource_type, r.visibility, r.link_url, r.latest_version_id, r.created_by, r.created_at, r.updated_at
		FROM resources r
		LEFT JOIN resource_versions rv ON rv.resource_id = r.id
		WHERE r.course_id = $1
		AND (
			r.search_vector @@ plainto_tsquery('english', $2)
			OR rv.text_search_vector @@ plainto_tsquery('english', $2)
		)
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4`
	if err := r.db.SelectContext(ctx, &resources, searchQuery, courseID, query, limit, offset); err != nil {
		return nil, fmt.Errorf("resource_repo.Search: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepository) SearchWithVisibility(ctx context.Context, courseID uuid.UUID, query string, visibility models.ResourceVisibility, offset, limit int) ([]models.Resource, error) {
	var resources []models.Resource
	searchQuery := `SELECT DISTINCT r.id, r.course_id, r.node_id, r.title, r.description, r.resource_type, r.visibility, r.link_url, r.latest_version_id, r.created_by, r.created_at, r.updated_at
		FROM resources r
		LEFT JOIN resource_versions rv ON rv.resource_id = r.id
		WHERE r.course_id = $1
		AND r.visibility = $5
		AND (
			r.search_vector @@ plainto_tsquery('english', $2)
			OR rv.text_search_vector @@ plainto_tsquery('english', $2)
		)
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4`
	if err := r.db.SelectContext(ctx, &resources, searchQuery, courseID, query, limit, offset, visibility); err != nil {
		return nil, fmt.Errorf("resource_repo.SearchWithVisibility: %w", err)
	}
	return resources, nil
}

func (r *ResourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM resources WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("resource_repo.Delete: %w", err)
	}
	return nil
}
