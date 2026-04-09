package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MatchAssignmentRepository struct {
	db *sqlx.DB
}

func NewMatchAssignmentRepository(db *sqlx.DB) *MatchAssignmentRepository {
	return &MatchAssignmentRepository{db: db}
}

func (r *MatchAssignmentRepository) Create(ctx context.Context, assignment *models.MatchAssignment) error {
	query := `INSERT INTO match_assignments (id, match_id, account_id, role, assigned_by, reassignment_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.MatchID, assignment.AccountID, assignment.Role,
		assignment.AssignedBy, assignment.ReassignmentReason,
		assignment.CreatedAt, assignment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("match_assignment_repo.Create: %w", err)
	}
	return nil
}

func (r *MatchAssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MatchAssignment, error) {
	var assignment models.MatchAssignment
	query := `SELECT id, match_id, account_id, role, assigned_by, reassignment_reason, created_at, updated_at
		FROM match_assignments WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&assignment); err != nil {
		return nil, fmt.Errorf("match_assignment_repo.GetByID: %w", err)
	}
	return &assignment, nil
}

func (r *MatchAssignmentRepository) ListByMatch(ctx context.Context, matchID uuid.UUID) ([]models.MatchAssignment, error) {
	var assignments []models.MatchAssignment
	query := `SELECT id, match_id, account_id, role, assigned_by, reassignment_reason, created_at, updated_at
		FROM match_assignments WHERE match_id = $1 ORDER BY role, created_at`
	if err := r.db.SelectContext(ctx, &assignments, query, matchID); err != nil {
		return nil, fmt.Errorf("match_assignment_repo.ListByMatch: %w", err)
	}
	return assignments, nil
}

func (r *MatchAssignmentRepository) Update(ctx context.Context, assignment *models.MatchAssignment) error {
	query := `UPDATE match_assignments SET account_id = $1, assigned_by = $2, reassignment_reason = $3, updated_at = NOW()
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query,
		assignment.AccountID, assignment.AssignedBy, assignment.ReassignmentReason, assignment.ID)
	if err != nil {
		return fmt.Errorf("match_assignment_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("match_assignment_repo.Update: assignment not found")
	}
	return nil
}

func (r *MatchAssignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM match_assignments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("match_assignment_repo.Delete: %w", err)
	}
	return nil
}
