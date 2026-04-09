package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Config CRUD

func (r *ReviewRepository) CreateConfig(ctx context.Context, cfg *models.ReviewConfig) error {
	query := `INSERT INTO review_configs (id, review_type, description, required_levels, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		cfg.ID, cfg.ReviewType, cfg.Description, cfg.RequiredLevels,
		cfg.CreatedBy, cfg.CreatedAt, cfg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("review_repo.CreateConfig: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*models.ReviewConfig, error) {
	var cfg models.ReviewConfig
	query := `SELECT id, review_type, description, required_levels, created_by, created_at, updated_at
		FROM review_configs WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&cfg); err != nil {
		return nil, fmt.Errorf("review_repo.GetConfigByID: %w", err)
	}
	return &cfg, nil
}

func (r *ReviewRepository) GetConfigByType(ctx context.Context, reviewType string) (*models.ReviewConfig, error) {
	var cfg models.ReviewConfig
	query := `SELECT id, review_type, description, required_levels, created_by, created_at, updated_at
		FROM review_configs WHERE review_type = $1`
	if err := r.db.QueryRowxContext(ctx, query, reviewType).StructScan(&cfg); err != nil {
		return nil, fmt.Errorf("review_repo.GetConfigByType: %w", err)
	}
	return &cfg, nil
}

func (r *ReviewRepository) ListConfigs(ctx context.Context, offset, limit int) ([]models.ReviewConfig, error) {
	var configs []models.ReviewConfig
	query := `SELECT id, review_type, description, required_levels, created_by, created_at, updated_at
		FROM review_configs ORDER BY review_type LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &configs, query, limit, offset); err != nil {
		return nil, fmt.Errorf("review_repo.ListConfigs: %w", err)
	}
	return configs, nil
}

func (r *ReviewRepository) UpdateConfig(ctx context.Context, cfg *models.ReviewConfig) error {
	query := `UPDATE review_configs SET description = $1, required_levels = $2, updated_at = NOW()
		WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, cfg.Description, cfg.RequiredLevels, cfg.ID)
	if err != nil {
		return fmt.Errorf("review_repo.UpdateConfig: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("review_repo.UpdateConfig: config not found")
	}
	return nil
}

func (r *ReviewRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM review_configs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("review_repo.DeleteConfig: %w", err)
	}
	return nil
}

// ReviewRequest CRUD

func (r *ReviewRepository) CreateRequest(ctx context.Context, req *models.ReviewRequest) error {
	query := `INSERT INTO review_requests (id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		req.ID, req.ReviewType, req.EntityType, req.EntityID,
		req.RequiredLevels, req.CurrentLevel, req.Status, req.SubmittedBy,
		req.FinalDecision, req.ParentID, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("review_repo.CreateRequest: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*models.ReviewRequest, error) {
	var req models.ReviewRequest
	query := `SELECT id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at
		FROM review_requests WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&req); err != nil {
		return nil, fmt.Errorf("review_repo.GetRequestByID: %w", err)
	}
	return &req, nil
}

func (r *ReviewRepository) ListRequests(ctx context.Context, status *models.ReviewRequestStatus, offset, limit int) ([]models.ReviewRequest, error) {
	var requests []models.ReviewRequest
	if status != nil {
		query := `SELECT id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at
			FROM review_requests WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		if err := r.db.SelectContext(ctx, &requests, query, *status, limit, offset); err != nil {
			return nil, fmt.Errorf("review_repo.ListRequests: %w", err)
		}
	} else {
		query := `SELECT id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at
			FROM review_requests ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		if err := r.db.SelectContext(ctx, &requests, query, limit, offset); err != nil {
			return nil, fmt.Errorf("review_repo.ListRequests: %w", err)
		}
	}
	return requests, nil
}

func (r *ReviewRepository) ListByAssignee(ctx context.Context, assigneeID uuid.UUID, offset, limit int) ([]models.ReviewRequest, error) {
	var requests []models.ReviewRequest
	query := `SELECT DISTINCT rr.id, rr.review_type, rr.entity_type, rr.entity_id, rr.required_levels, rr.current_level, rr.status, rr.submitted_by, rr.final_decision, rr.parent_id, rr.created_at, rr.updated_at
		FROM review_requests rr
		JOIN review_levels rl ON rl.request_id = rr.id
		WHERE rl.assignee_id = $1 AND rl.status = 'Pending'
		ORDER BY rr.created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &requests, query, assigneeID, limit, offset); err != nil {
		return nil, fmt.Errorf("review_repo.ListByAssignee: %w", err)
	}
	return requests, nil
}

func (r *ReviewRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.ReviewRequest, error) {
	var requests []models.ReviewRequest
	query := `SELECT id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at
		FROM review_requests WHERE entity_type = $1 AND entity_id = $2 ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &requests, query, entityType, entityID); err != nil {
		return nil, fmt.Errorf("review_repo.ListByEntity: %w", err)
	}
	return requests, nil
}

func (r *ReviewRepository) ListFollowUps(ctx context.Context, parentID uuid.UUID) ([]models.ReviewRequest, error) {
	var requests []models.ReviewRequest
	query := `SELECT id, review_type, entity_type, entity_id, required_levels, current_level, status, submitted_by, final_decision, parent_id, created_at, updated_at
		FROM review_requests WHERE parent_id = $1 ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &requests, query, parentID); err != nil {
		return nil, fmt.Errorf("review_repo.ListFollowUps: %w", err)
	}
	return requests, nil
}

func (r *ReviewRepository) UpdateRequest(ctx context.Context, req *models.ReviewRequest) error {
	query := `UPDATE review_requests SET current_level = $1, status = $2, final_decision = $3, updated_at = NOW()
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, req.CurrentLevel, req.Status, req.FinalDecision, req.ID)
	if err != nil {
		return fmt.Errorf("review_repo.UpdateRequest: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("review_repo.UpdateRequest: request not found")
	}
	return nil
}

// ReviewLevel CRUD

func (r *ReviewRepository) CreateLevel(ctx context.Context, level *models.ReviewLevel) error {
	query := `INSERT INTO review_levels (id, request_id, level, assignee_id, status, decision, annotation, decided_by, decided_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, query,
		level.ID, level.RequestID, level.Level, level.AssigneeID,
		level.Status, level.Decision, level.Annotation,
		level.DecidedBy, level.DecidedAt, level.CreatedAt, level.UpdatedAt)
	if err != nil {
		return fmt.Errorf("review_repo.CreateLevel: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetLevelByID(ctx context.Context, id uuid.UUID) (*models.ReviewLevel, error) {
	var level models.ReviewLevel
	query := `SELECT id, request_id, level, assignee_id, status, decision, annotation, decided_by, decided_at, created_at, updated_at
		FROM review_levels WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&level); err != nil {
		return nil, fmt.Errorf("review_repo.GetLevelByID: %w", err)
	}
	return &level, nil
}

func (r *ReviewRepository) GetLevelByRequestAndLevel(ctx context.Context, requestID uuid.UUID, level int) (*models.ReviewLevel, error) {
	var rl models.ReviewLevel
	query := `SELECT id, request_id, level, assignee_id, status, decision, annotation, decided_by, decided_at, created_at, updated_at
		FROM review_levels WHERE request_id = $1 AND level = $2`
	if err := r.db.QueryRowxContext(ctx, query, requestID, level).StructScan(&rl); err != nil {
		return nil, fmt.Errorf("review_repo.GetLevelByRequestAndLevel: %w", err)
	}
	return &rl, nil
}

func (r *ReviewRepository) ListLevelsByRequest(ctx context.Context, requestID uuid.UUID) ([]models.ReviewLevel, error) {
	var levels []models.ReviewLevel
	query := `SELECT id, request_id, level, assignee_id, status, decision, annotation, decided_by, decided_at, created_at, updated_at
		FROM review_levels WHERE request_id = $1 ORDER BY level`
	if err := r.db.SelectContext(ctx, &levels, query, requestID); err != nil {
		return nil, fmt.Errorf("review_repo.ListLevelsByRequest: %w", err)
	}
	return levels, nil
}

func (r *ReviewRepository) UpdateLevel(ctx context.Context, level *models.ReviewLevel) error {
	query := `UPDATE review_levels SET assignee_id = $1, status = $2, decision = $3, annotation = $4, decided_by = $5, decided_at = $6, updated_at = NOW()
		WHERE id = $7`
	result, err := r.db.ExecContext(ctx, query,
		level.AssigneeID, level.Status, level.Decision, level.Annotation,
		level.DecidedBy, level.DecidedAt, level.ID)
	if err != nil {
		return fmt.Errorf("review_repo.UpdateLevel: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("review_repo.UpdateLevel: level not found")
	}
	return nil
}

// FollowUp CRUD

func (r *ReviewRepository) CreateFollowUp(ctx context.Context, fu *models.ReviewFollowUp) error {
	query := `INSERT INTO review_follow_ups (id, request_id, author_id, content, level, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, fu.ID, fu.RequestID, fu.AuthorID, fu.Content, fu.Level, fu.CreatedAt)
	if err != nil {
		return fmt.Errorf("review_repo.CreateFollowUp: %w", err)
	}
	return nil
}

func (r *ReviewRepository) ListFollowUpsByRequest(ctx context.Context, requestID uuid.UUID) ([]models.ReviewFollowUp, error) {
	var followUps []models.ReviewFollowUp
	query := `SELECT id, request_id, author_id, content, level, created_at
		FROM review_follow_ups WHERE request_id = $1 ORDER BY created_at`
	if err := r.db.SelectContext(ctx, &followUps, query, requestID); err != nil {
		return nil, fmt.Errorf("review_repo.ListFollowUpsByRequest: %w", err)
	}
	return followUps, nil
}
