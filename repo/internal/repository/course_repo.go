package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CourseRepository struct {
	db *sqlx.DB
}

func NewCourseRepository(db *sqlx.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) Create(ctx context.Context, course *models.Course) error {
	query := `INSERT INTO courses (id, title, description, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		course.ID, course.Title, course.Description, course.Status,
		course.CreatedBy, course.CreatedAt, course.UpdatedAt)
	if err != nil {
		return fmt.Errorf("course_repo.Create: %w", err)
	}
	return nil
}

func (r *CourseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	var course models.Course
	query := `SELECT id, title, description, status, created_by, created_at, updated_at FROM courses WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&course); err != nil {
		return nil, fmt.Errorf("course_repo.GetByID: %w", err)
	}
	return &course, nil
}

func (r *CourseRepository) List(ctx context.Context, offset, limit int) ([]models.Course, error) {
	var courses []models.Course
	query := `SELECT id, title, description, status, created_by, created_at, updated_at
		FROM courses ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &courses, query, limit, offset); err != nil {
		return nil, fmt.Errorf("course_repo.List: %w", err)
	}
	return courses, nil
}

func (r *CourseRepository) ListPublished(ctx context.Context, offset, limit int) ([]models.Course, error) {
	var courses []models.Course
	query := `SELECT id, title, description, status, created_by, created_at, updated_at
		FROM courses WHERE status = 'Published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &courses, query, limit, offset); err != nil {
		return nil, fmt.Errorf("course_repo.ListPublished: %w", err)
	}
	return courses, nil
}

func (r *CourseRepository) Update(ctx context.Context, course *models.Course) error {
	query := `UPDATE courses SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, course.Title, course.Description, course.Status, course.ID)
	if err != nil {
		return fmt.Errorf("course_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("course_repo.Update: course not found")
	}
	return nil
}

func (r *CourseRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CourseStatus) error {
	query := `UPDATE courses SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("course_repo.UpdateStatus: %w", err)
	}
	return nil
}
