package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ModerationReviewRepository struct {
	db *sqlx.DB
}

func NewModerationReviewRepository(db *sqlx.DB) *ModerationReviewRepository {
	return &ModerationReviewRepository{db: db}
}

func (r *ModerationReviewRepository) Create(ctx context.Context, review *models.ModerationReview) error {
	query := `INSERT INTO moderation_reviews (id, content_type, content_id, content_snippet, status, moderator_id, reason, decided_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.ContentType, review.ContentID, review.ContentSnippet,
		review.Status, review.ModeratorID, review.Reason, review.DecidedAt,
		review.CreatedAt, review.UpdatedAt)
	if err != nil {
		return fmt.Errorf("moderation_review_repo.Create: %w", err)
	}
	return nil
}

func (r *ModerationReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ModerationReview, error) {
	var review models.ModerationReview
	query := `SELECT id, content_type, content_id, content_snippet, status, moderator_id, reason, decided_at, created_at, updated_at
		FROM moderation_reviews WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&review); err != nil {
		return nil, fmt.Errorf("moderation_review_repo.GetByID: %w", err)
	}
	return &review, nil
}

func (r *ModerationReviewRepository) List(ctx context.Context, status *models.ReviewStatus, offset, limit int) ([]models.ModerationReview, error) {
	var reviews []models.ModerationReview
	if status != nil {
		query := `SELECT id, content_type, content_id, content_snippet, status, moderator_id, reason, decided_at, created_at, updated_at
			FROM moderation_reviews WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		if err := r.db.SelectContext(ctx, &reviews, query, *status, limit, offset); err != nil {
			return nil, fmt.Errorf("moderation_review_repo.List: %w", err)
		}
	} else {
		query := `SELECT id, content_type, content_id, content_snippet, status, moderator_id, reason, decided_at, created_at, updated_at
			FROM moderation_reviews ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		if err := r.db.SelectContext(ctx, &reviews, query, limit, offset); err != nil {
			return nil, fmt.Errorf("moderation_review_repo.List: %w", err)
		}
	}
	return reviews, nil
}

func (r *ModerationReviewRepository) Update(ctx context.Context, review *models.ModerationReview) error {
	query := `UPDATE moderation_reviews SET status = $1, moderator_id = $2, reason = $3, decided_at = $4, updated_at = NOW()
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, review.Status, review.ModeratorID, review.Reason, review.DecidedAt, review.ID)
	if err != nil {
		return fmt.Errorf("moderation_review_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("moderation_review_repo.Update: review not found")
	}
	return nil
}
