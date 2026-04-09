package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuditLogRepository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action,
		log.ActorID, log.Details, log.Tier, log.Reason,
		log.Source, log.Workstation, log.BeforeSnapshot, log.AfterSnapshot,
		log.ContentHash, log.ExpiresAt, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("audit_log_repo.Create: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	var log models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&log); err != nil {
		return nil, fmt.Errorf("audit_log_repo.GetByID: %w", err)
	}
	return &log, nil
}

func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	if err := r.db.SelectContext(ctx, &logs, query, entityType, entityID, limit, offset); err != nil {
		return nil, fmt.Errorf("audit_log_repo.ListByEntity: %w", err)
	}
	return logs, nil
}

func (r *AuditLogRepository) ListByActor(ctx context.Context, actorID uuid.UUID, offset, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs WHERE actor_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &logs, query, actorID, limit, offset); err != nil {
		return nil, fmt.Errorf("audit_log_repo.ListByActor: %w", err)
	}
	return logs, nil
}

func (r *AuditLogRepository) ListByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs WHERE created_at >= $1 AND created_at < $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	if err := r.db.SelectContext(ctx, &logs, query, start, end, limit, offset); err != nil {
		return nil, fmt.Errorf("audit_log_repo.ListByTimeRange: %w", err)
	}
	return logs, nil
}

type AuditQueryParams struct {
	ActorID    *uuid.UUID
	EntityType *string
	Action     *string
	Tier       *models.LogTier
	StartTime  *time.Time
	EndTime    *time.Time
	Offset     int
	Limit      int
}

func (r *AuditLogRepository) Query(ctx context.Context, params AuditQueryParams) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if params.ActorID != nil {
		query += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, *params.ActorID)
		argIdx++
	}
	if params.EntityType != nil {
		query += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, *params.EntityType)
		argIdx++
	}
	if params.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *params.Action)
		argIdx++
	}
	if params.Tier != nil {
		query += fmt.Sprintf(" AND tier = $%d", argIdx)
		args = append(args, string(*params.Tier))
		argIdx++
	}
	if params.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *params.StartTime)
		argIdx++
	}
	if params.EndTime != nil {
		query += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *params.EndTime)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return nil, fmt.Errorf("audit_log_repo.Query: %w", err)
	}
	return logs, nil
}

// GetDayEntries returns all audit entries for a specific day (for hash chain computation)
func (r *AuditLogRepository) GetDayEntries(ctx context.Context, date time.Time) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `SELECT id, entity_type, entity_id, action, actor_id, details, tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at, created_at
		FROM audit_logs
		WHERE created_at >= $1::date AND created_at < ($1::date + INTERVAL '1 day')
		ORDER BY created_at, id`
	if err := r.db.SelectContext(ctx, &logs, query, date); err != nil {
		return nil, fmt.Errorf("audit_log_repo.GetDayEntries: %w", err)
	}
	return logs, nil
}

// DeleteExpired removes logs past their expiration date
func (r *AuditLogRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM audit_logs WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("audit_log_repo.DeleteExpired: %w", err)
	}
	return result.RowsAffected()
}

// CountByTier returns entry counts per tier
func (r *AuditLogRepository) CountByTier(ctx context.Context) (map[models.LogTier]int, error) {
	type row struct {
		Tier  string `db:"tier"`
		Count int    `db:"count"`
	}
	var rows []row
	query := `SELECT tier, COUNT(*) AS count FROM audit_logs GROUP BY tier`
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("audit_log_repo.CountByTier: %w", err)
	}
	result := make(map[models.LogTier]int)
	for _, r := range rows {
		result[models.LogTier(r.Tier)] = r.Count
	}
	return result, nil
}

// Hash chain operations

func (r *AuditLogRepository) CreateHashChain(ctx context.Context, chain *models.AuditHashChain) error {
	query := `INSERT INTO audit_hash_chain (id, chain_date, entry_count, batch_hash, previous_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		chain.ID, chain.ChainDate, chain.EntryCount, chain.BatchHash,
		chain.PreviousHash, chain.CreatedAt)
	if err != nil {
		return fmt.Errorf("audit_log_repo.CreateHashChain: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) GetHashChainByDate(ctx context.Context, date time.Time) (*models.AuditHashChain, error) {
	var chain models.AuditHashChain
	query := `SELECT id, chain_date, entry_count, batch_hash, previous_hash, created_at
		FROM audit_hash_chain WHERE chain_date = $1`
	if err := r.db.QueryRowxContext(ctx, query, date).StructScan(&chain); err != nil {
		return nil, fmt.Errorf("audit_log_repo.GetHashChainByDate: %w", err)
	}
	return &chain, nil
}

func (r *AuditLogRepository) GetLatestHashChain(ctx context.Context) (*models.AuditHashChain, error) {
	var chain models.AuditHashChain
	query := `SELECT id, chain_date, entry_count, batch_hash, previous_hash, created_at
		FROM audit_hash_chain ORDER BY chain_date DESC LIMIT 1`
	if err := r.db.QueryRowxContext(ctx, query).StructScan(&chain); err != nil {
		return nil, fmt.Errorf("audit_log_repo.GetLatestHashChain: %w", err)
	}
	return &chain, nil
}

func (r *AuditLogRepository) ListHashChain(ctx context.Context, offset, limit int) ([]models.AuditHashChain, error) {
	var chains []models.AuditHashChain
	query := `SELECT id, chain_date, entry_count, batch_hash, previous_hash, created_at
		FROM audit_hash_chain ORDER BY chain_date DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &chains, query, limit, offset); err != nil {
		return nil, fmt.Errorf("audit_log_repo.ListHashChain: %w", err)
	}
	return chains, nil
}

// DB returns the underlying database connection for health checks
func (r *AuditLogRepository) DB() *sqlx.DB {
	return r.db
}
