package dto

import (
	"encoding/json"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type AuditLogResponse struct {
	ID             uuid.UUID       `json:"id"`
	EntityType     string          `json:"entity_type"`
	EntityID       uuid.UUID       `json:"entity_id"`
	Action         string          `json:"action"`
	ActorID        uuid.UUID       `json:"actor_id"`
	Details        json.RawMessage `json:"details,omitempty"`
	Tier           string          `json:"tier"`
	Reason         *string         `json:"reason,omitempty"`
	Source         *string         `json:"source,omitempty"`
	Workstation    *string         `json:"workstation,omitempty"`
	BeforeSnapshot json.RawMessage `json:"before_snapshot,omitempty"`
	AfterSnapshot  json.RawMessage `json:"after_snapshot,omitempty"`
	ContentHash    *string         `json:"content_hash,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func ToAuditLogResponse(l *models.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:             l.ID,
		EntityType:     l.EntityType,
		EntityID:       l.EntityID,
		Action:         l.Action,
		ActorID:        l.ActorID,
		Details:        l.Details,
		Tier:           string(l.Tier),
		Reason:         l.Reason,
		Source:         l.Source,
		Workstation:    l.Workstation,
		BeforeSnapshot: l.BeforeSnapshot,
		AfterSnapshot:  l.AfterSnapshot,
		ContentHash:    l.ContentHash,
		ExpiresAt:      l.ExpiresAt,
		CreatedAt:      l.CreatedAt,
	}
}

func ToAuditLogResponseList(logs []models.AuditLog) []AuditLogResponse {
	result := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		result[i] = ToAuditLogResponse(&l)
	}
	return result
}

type HashChainResponse struct {
	ID           uuid.UUID `json:"id"`
	ChainDate    string    `json:"chain_date"`
	EntryCount   int       `json:"entry_count"`
	BatchHash    string    `json:"batch_hash"`
	PreviousHash *string   `json:"previous_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func ToHashChainResponse(c *models.AuditHashChain) HashChainResponse {
	return HashChainResponse{
		ID:           c.ID,
		ChainDate:    c.ChainDate.Format("2006-01-02"),
		EntryCount:   c.EntryCount,
		BatchHash:    c.BatchHash,
		PreviousHash: c.PreviousHash,
		CreatedAt:    c.CreatedAt,
	}
}

func ToHashChainResponseList(chains []models.AuditHashChain) []HashChainResponse {
	result := make([]HashChainResponse, len(chains))
	for i, c := range chains {
		result[i] = ToHashChainResponse(&c)
	}
	return result
}

type VerifyChainResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type TierCountResponse struct {
	Access    int `json:"access"`
	Operation int `json:"operation"`
	Audit     int `json:"audit"`
}

type PurgeResponse struct {
	DeletedCount int64 `json:"deleted_count"`
}

type BuildChainRequest struct {
	Date string `json:"date"`
}
