package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `INSERT INTO accounts (id, username, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		account.ID, account.Username, account.PasswordHash,
		account.Role, account.Status, account.CreatedAt, account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("account_repo.Create: %w", err)
	}
	return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	var account models.Account
	query := `SELECT id, username, password_hash, role, status, created_at, updated_at FROM accounts WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&account); err != nil {
		return nil, fmt.Errorf("account_repo.GetByID: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	var account models.Account
	query := `SELECT id, username, password_hash, role, status, created_at, updated_at FROM accounts WHERE username = $1`
	if err := r.db.QueryRowxContext(ctx, query, username).StructScan(&account); err != nil {
		return nil, fmt.Errorf("account_repo.GetByUsername: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.Status) error {
	query := `UPDATE accounts SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("account_repo.UpdateStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("account_repo.UpdateStatus: account not found")
	}
	return nil
}

func (r *AccountRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE accounts SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("account_repo.UpdatePassword: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("account_repo.UpdatePassword: account not found")
	}
	return nil
}

func (r *AccountRepository) List(ctx context.Context, offset, limit int) ([]models.Account, error) {
	var accounts []models.Account
	query := `SELECT id, username, password_hash, role, status, created_at, updated_at
		FROM accounts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &accounts, query, limit, offset); err != nil {
		return nil, fmt.Errorf("account_repo.List: %w", err)
	}
	return accounts, nil
}
