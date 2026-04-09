package service

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

type AuditService struct {
	repo *repository.AuditLogRepository
	cfg  *config.Config
}

func NewAuditService(repo *repository.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// SetConfig sets the config (called after construction so existing callers are unaffected)
func (s *AuditService) SetConfig(cfg *config.Config) {
	s.cfg = cfg
}

// Log records an audit entry with the default operation tier (backward-compatible)
func (s *AuditService) Log(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action string, details interface{}) {
	s.LogExtended(ctx, &AuditEntry{
		EntityType: entityType,
		EntityID:   entityID,
		ActorID:    actorID,
		Action:     action,
		Details:    details,
		Tier:       models.TierOperation,
	})
}

// AuditEntry is the extended input for LogExtended
type AuditEntry struct {
	EntityType     string
	EntityID       uuid.UUID
	ActorID        uuid.UUID
	Action         string
	Details        interface{}
	Tier           models.LogTier
	Reason         *string
	Source         *string
	Workstation    *string
	BeforeSnapshot interface{}
	AfterSnapshot  interface{}
}

// LogExtended records an audit entry with all extended fields
func (s *AuditService) LogExtended(ctx context.Context, entry *AuditEntry) {
	var detailsJSON json.RawMessage
	if entry.Details != nil {
		if b, err := json.Marshal(entry.Details); err == nil {
			detailsJSON = b
		}
	}
	var beforeJSON json.RawMessage
	if entry.BeforeSnapshot != nil {
		if b, err := json.Marshal(entry.BeforeSnapshot); err == nil {
			beforeJSON = b
		}
	}
	var afterJSON json.RawMessage
	if entry.AfterSnapshot != nil {
		if b, err := json.Marshal(entry.AfterSnapshot); err == nil {
			afterJSON = b
		}
	}

	tier := entry.Tier
	if tier == "" {
		tier = models.TierOperation
	}

	now := time.Now()

	// Compute content hash for integrity
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		entry.EntityType, entry.EntityID, entry.ActorID,
		entry.Action, string(detailsJSON), now.Format(time.RFC3339Nano))
	h := sha256.Sum256([]byte(hashInput))
	contentHash := hex.EncodeToString(h[:])

	// Compute expiration based on tier
	var expiresAt *time.Time
	if days, ok := models.TierRetentionDays[tier]; ok {
		exp := now.Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &exp
	}

	log := &models.AuditLog{
		ID:             uuid.New(),
		EntityType:     entry.EntityType,
		EntityID:       entry.EntityID,
		Action:         entry.Action,
		ActorID:        entry.ActorID,
		Details:        detailsJSON,
		Tier:           tier,
		Reason:         entry.Reason,
		Source:         entry.Source,
		Workstation:    entry.Workstation,
		BeforeSnapshot: beforeJSON,
		AfterSnapshot:  afterJSON,
		ContentHash:    &contentHash,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
	}
	// Best-effort logging — don't fail the main operation
	s.repo.Create(ctx, log)
}

// Query methods

func (s *AuditService) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuditService) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByEntity(ctx, entityType, entityID, offset, limit)
}

func (s *AuditService) ListByActor(ctx context.Context, actorID uuid.UUID, offset, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByActor(ctx, actorID, offset, limit)
}

func (s *AuditService) Query(ctx context.Context, params repository.AuditQueryParams) ([]models.AuditLog, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.Query(ctx, params)
}

func (s *AuditService) CountByTier(ctx context.Context) (map[models.LogTier]int, error) {
	return s.repo.CountByTier(ctx)
}

// Retention cleanup
func (s *AuditService) PurgeExpired(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpired(ctx)
}

// Hash chain operations

// BuildDailyHashChain computes the hash for a specific day and chains it to the previous day
func (s *AuditService) BuildDailyHashChain(ctx context.Context, date time.Time) (*models.AuditHashChain, error) {
	// Check if chain already exists for this date
	existing, err := s.repo.GetHashChainByDate(ctx, date)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Get all entries for the day
	entries, err := s.repo.GetDayEntries(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("getting day entries: %w", err)
	}

	// Compute batch hash over all entries
	batchHash := s.computeBatchHash(entries)

	// Get previous day's hash
	var previousHash *string
	prev, err := s.repo.GetLatestHashChain(ctx)
	if err == nil && prev != nil {
		previousHash = &prev.BatchHash
	}

	// Chain: hash(previousHash + batchHash)
	chainedInput := ""
	if previousHash != nil {
		chainedInput = *previousHash + "|"
	}
	chainedInput += batchHash
	ch := sha256.Sum256([]byte(chainedInput))
	finalHash := hex.EncodeToString(ch[:])

	chain := &models.AuditHashChain{
		ID:           uuid.New(),
		ChainDate:    date,
		EntryCount:   len(entries),
		BatchHash:    finalHash,
		PreviousHash: previousHash,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateHashChain(ctx, chain); err != nil {
		return nil, err
	}

	return chain, nil
}

// VerifyHashChain verifies the integrity of the hash chain for a given date
func (s *AuditService) VerifyHashChain(ctx context.Context, date time.Time) (bool, string, error) {
	chain, err := s.repo.GetHashChainByDate(ctx, date)
	if err != nil {
		return false, "no hash chain found for this date", nil
	}

	// Re-compute batch hash from entries
	entries, err := s.repo.GetDayEntries(ctx, date)
	if err != nil {
		return false, "", fmt.Errorf("getting day entries: %w", err)
	}

	batchHash := s.computeBatchHash(entries)

	// Re-compute chained hash
	chainedInput := ""
	if chain.PreviousHash != nil {
		chainedInput = *chain.PreviousHash + "|"
	}
	chainedInput += batchHash
	ch := sha256.Sum256([]byte(chainedInput))
	expectedHash := hex.EncodeToString(ch[:])

	if chain.BatchHash != expectedHash {
		return false, fmt.Sprintf("hash mismatch: expected %s, got %s", expectedHash, chain.BatchHash), nil
	}

	if chain.EntryCount != len(entries) {
		return false, fmt.Sprintf("entry count mismatch: expected %d, got %d", chain.EntryCount, len(entries)), nil
	}

	return true, "integrity verified", nil
}

func (s *AuditService) ListHashChain(ctx context.Context, offset, limit int) ([]models.AuditHashChain, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListHashChain(ctx, offset, limit)
}

func (s *AuditService) computeBatchHash(entries []models.AuditLog) string {
	hasher := sha256.New()
	for _, e := range entries {
		// Hash each entry's content hash (or reconstruct from fields)
		content := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			e.ID, e.EntityType, e.EntityID, e.Action, e.ActorID, e.CreatedAt.Format(time.RFC3339Nano))
		if e.ContentHash != nil {
			content += "|" + *e.ContentHash
		}
		hasher.Write([]byte(content))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// CSV export
func (s *AuditService) ExportCSV(ctx context.Context, params repository.AuditQueryParams) (string, error) {
	// Remove limit for export
	params.Limit = 10000
	params.Offset = 0

	logs, err := s.repo.Query(ctx, params)
	if err != nil {
		return "", err
	}

	storagePath := "./storage"
	if s.cfg != nil {
		storagePath = s.cfg.StoragePath
	}
	csvDir := filepath.Join(storagePath, "audit-exports")
	os.MkdirAll(csvDir, 0750)
	csvPath := filepath.Join(csvDir, fmt.Sprintf("audit_%s.csv", uuid.New().String()))

	f, err := os.Create(csvPath)
	if err != nil {
		return "", fmt.Errorf("creating CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{
		"id", "entity_type", "entity_id", "action", "actor_id", "tier",
		"reason", "source", "workstation", "details",
		"before_snapshot", "after_snapshot", "content_hash", "created_at",
	})

	for _, l := range logs {
		reason := ""
		if l.Reason != nil {
			reason = *l.Reason
		}
		source := ""
		if l.Source != nil {
			source = *l.Source
		}
		ws := ""
		if l.Workstation != nil {
			ws = *l.Workstation
		}
		hash := ""
		if l.ContentHash != nil {
			hash = *l.ContentHash
		}

		w.Write([]string{
			l.ID.String(),
			l.EntityType,
			l.EntityID.String(),
			l.Action,
			l.ActorID.String(),
			string(l.Tier),
			reason,
			source,
			ws,
			string(l.Details),
			string(l.BeforeSnapshot),
			string(l.AfterSnapshot),
			hash,
			l.CreatedAt.Format(time.RFC3339),
		})
	}

	return csvPath, nil
}

// VerifyEntryIntegrity checks a single entry's content hash
func (s *AuditService) VerifyEntryIntegrity(entry *models.AuditLog) bool {
	if entry.ContentHash == nil {
		return true // Pre-extension entries without hash are assumed valid
	}
	hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		entry.EntityType, entry.EntityID, entry.ActorID,
		entry.Action, string(entry.Details), entry.CreatedAt.Format(time.RFC3339Nano))
	h := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(h[:]) == *entry.ContentHash
}

// DB exposes the underlying DB for health checks
func (s *AuditService) DB() *repository.AuditLogRepository {
	return s.repo
}

