package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CourseOutlineRepository struct {
	db *sqlx.DB
}

func NewCourseOutlineRepository(db *sqlx.DB) *CourseOutlineRepository {
	return &CourseOutlineRepository{db: db}
}

func (r *CourseOutlineRepository) Create(ctx context.Context, node *models.CourseOutlineNode) error {
	query := `INSERT INTO course_outline_nodes (id, course_id, parent_id, node_type, title, description, order_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		node.ID, node.CourseID, node.ParentID, node.NodeType,
		node.Title, node.Description, node.OrderIndex,
		node.CreatedAt, node.UpdatedAt)
	if err != nil {
		return fmt.Errorf("course_outline_repo.Create: %w", err)
	}
	return nil
}

func (r *CourseOutlineRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CourseOutlineNode, error) {
	var node models.CourseOutlineNode
	query := `SELECT id, course_id, parent_id, node_type, title, description, order_index, created_at, updated_at
		FROM course_outline_nodes WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&node); err != nil {
		return nil, fmt.Errorf("course_outline_repo.GetByID: %w", err)
	}
	return &node, nil
}

func (r *CourseOutlineRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]models.CourseOutlineNode, error) {
	var nodes []models.CourseOutlineNode
	query := `SELECT id, course_id, parent_id, node_type, title, description, order_index, created_at, updated_at
		FROM course_outline_nodes WHERE course_id = $1
		ORDER BY COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'), order_index`
	if err := r.db.SelectContext(ctx, &nodes, query, courseID); err != nil {
		return nil, fmt.Errorf("course_outline_repo.ListByCourse: %w", err)
	}
	return nodes, nil
}

func (r *CourseOutlineRepository) ListChildren(ctx context.Context, parentID uuid.UUID) ([]models.CourseOutlineNode, error) {
	var nodes []models.CourseOutlineNode
	query := `SELECT id, course_id, parent_id, node_type, title, description, order_index, created_at, updated_at
		FROM course_outline_nodes WHERE parent_id = $1 ORDER BY order_index`
	if err := r.db.SelectContext(ctx, &nodes, query, parentID); err != nil {
		return nil, fmt.Errorf("course_outline_repo.ListChildren: %w", err)
	}
	return nodes, nil
}

func (r *CourseOutlineRepository) Update(ctx context.Context, node *models.CourseOutlineNode) error {
	query := `UPDATE course_outline_nodes SET title = $1, description = $2, order_index = $3, parent_id = $4, updated_at = NOW()
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, node.Title, node.Description, node.OrderIndex, node.ParentID, node.ID)
	if err != nil {
		return fmt.Errorf("course_outline_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("course_outline_repo.Update: node not found")
	}
	return nil
}

func (r *CourseOutlineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM course_outline_nodes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("course_outline_repo.Delete: %w", err)
	}
	return nil
}
