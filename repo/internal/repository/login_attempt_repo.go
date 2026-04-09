package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type LoginAttemptRepository struct {
	db *sqlx.DB
}

func NewLoginAttemptRepository(db *sqlx.DB) *LoginAttemptRepository {
	return &LoginAttemptRepository{db: db}
}

func (r *LoginAttemptRepository) Record(ctx context.Context, attempt *models.LoginAttempt) error {
	query := `INSERT INTO login_attempts (id, account_id, success, ip_address, attempted_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.AccountID, attempt.Success,
		attempt.IPAddress, attempt.AttemptedAt)
	if err != nil {
		return fmt.Errorf("login_attempt_repo.Record: %w", err)
	}
	return nil
}

func (r *LoginAttemptRepository) CountRecentFailures(ctx context.Context, accountID uuid.UUID, window time.Duration) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM login_attempts
		WHERE account_id = $1 AND success = FALSE AND attempted_at > NOW() - $2::interval`
	if err := r.db.QueryRowxContext(ctx, query, accountID, fmt.Sprintf("%d seconds", int(window.Seconds()))).Scan(&count); err != nil {
		return 0, fmt.Errorf("login_attempt_repo.CountRecentFailures: %w", err)
	}
	return count, nil
}
