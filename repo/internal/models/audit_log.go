package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LogTier string

const (
	TierAccess    LogTier = "access"
	TierOperation LogTier = "operation"
	TierAudit     LogTier = "audit"
)

// RetentionDays per tier
var TierRetentionDays = map[LogTier]int{
	TierAccess:    30,
	TierOperation: 180,
	TierAudit:     2555, // ~7 years
}

type AuditLog struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	EntityType     string          `db:"entity_type" json:"entity_type"`
	EntityID       uuid.UUID       `db:"entity_id" json:"entity_id"`
	Action         string          `db:"action" json:"action"`
	ActorID        uuid.UUID       `db:"actor_id" json:"actor_id"`
	Details        json.RawMessage `db:"details" json:"details,omitempty"`
	Tier           LogTier         `db:"tier" json:"tier"`
	Reason         *string         `db:"reason" json:"reason,omitempty"`
	Source         *string         `db:"source" json:"source,omitempty"`
	Workstation    *string         `db:"workstation" json:"workstation,omitempty"`
	BeforeSnapshot json.RawMessage `db:"before_snapshot" json:"before_snapshot,omitempty"`
	AfterSnapshot  json.RawMessage `db:"after_snapshot" json:"after_snapshot,omitempty"`
	ContentHash    *string         `db:"content_hash" json:"content_hash,omitempty"`
	ExpiresAt      *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type AuditHashChain struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ChainDate    time.Time `db:"chain_date" json:"chain_date"`
	EntryCount   int       `db:"entry_count" json:"entry_count"`
	BatchHash    string    `db:"batch_hash" json:"batch_hash"`
	PreviousHash *string   `db:"previous_hash" json:"previous_hash,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
