package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, account_id, token_hash, device_id, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.AccountID, token.TokenHash,
		token.DeviceID, token.ExpiresAt, token.Revoked, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("refresh_token_repo.Create: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	query := `SELECT id, account_id, token_hash, device_id, expires_at, revoked, created_at
		FROM refresh_tokens WHERE token_hash = $1`
	if err := r.db.QueryRowxContext(ctx, query, tokenHash).StructScan(&token); err != nil {
		return nil, fmt.Errorf("refresh_token_repo.GetByHash: %w", err)
	}
	return &token, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("refresh_token_repo.Revoke: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForAccount(ctx context.Context, accountID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE account_id = $1 AND revoked = FALSE`
	_, err := r.db.ExecContext(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("refresh_token_repo.RevokeAllForAccount: %w", err)
	}
	return nil
}
