package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CourseMembershipRepository struct {
	db *sqlx.DB
}

func NewCourseMembershipRepository(db *sqlx.DB) *CourseMembershipRepository {
	return &CourseMembershipRepository{db: db}
}

func (r *CourseMembershipRepository) Create(ctx context.Context, m *models.CourseMembership) error {
	query := `INSERT INTO course_memberships (id, course_id, account_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.CourseID, m.AccountID, m.Role, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("course_membership_repo.Create: %w", err)
	}
	return nil
}

func (r *CourseMembershipRepository) GetByAccountAndCourse(ctx context.Context, accountID, courseID uuid.UUID) (*models.CourseMembership, error) {
	var m models.CourseMembership
	query := `SELECT id, course_id, account_id, role, created_at
		FROM course_memberships WHERE account_id = $1 AND course_id = $2`
	if err := r.db.QueryRowxContext(ctx, query, accountID, courseID).StructScan(&m); err != nil {
		return nil, fmt.Errorf("course_membership_repo.GetByAccountAndCourse: %w", err)
	}
	return &m, nil
}

func (r *CourseMembershipRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]models.CourseMembership, error) {
	var memberships []models.CourseMembership
	query := `SELECT id, course_id, account_id, role, created_at
		FROM course_memberships WHERE course_id = $1 ORDER BY role, created_at`
	if err := r.db.SelectContext(ctx, &memberships, query, courseID); err != nil {
		return nil, fmt.Errorf("course_membership_repo.ListByCourse: %w", err)
	}
	return memberships, nil
}

func (r *CourseMembershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM course_memberships WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("course_membership_repo.Delete: %w", err)
	}
	return nil
}
