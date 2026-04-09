package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

// Fix #5: Compliance traceability - verify audit extended fields

func TestAuditEntryStructHasExtendedFields(t *testing.T) {
	entry := &service.AuditEntry{
		EntityType:     "course",
		EntityID:       uuid.New(),
		ActorID:        uuid.New(),
		Action:         "updated",
		Tier:           models.TierAudit,
		Reason:         strPtr("compliance requirement"),
		Source:         strPtr("web_ui"),
		Workstation:    strPtr("admin-pc-01"),
		BeforeSnapshot: map[string]string{"title": "Old Title"},
		AfterSnapshot:  map[string]string{"title": "New Title"},
		Details:        map[string]string{"change": "title"},
	}

	if entry.Tier != models.TierAudit {
		t.Errorf("expected audit tier, got %s", entry.Tier)
	}
	if entry.Reason == nil || *entry.Reason == "" {
		t.Error("reason should be set")
	}
	if entry.Source == nil || *entry.Source == "" {
		t.Error("source should be set")
	}
	if entry.BeforeSnapshot == nil {
		t.Error("before_snapshot should be set")
	}
	if entry.AfterSnapshot == nil {
		t.Error("after_snapshot should be set")
	}
}

func TestAuditTierRetentionDays(t *testing.T) {
	// Verify all three tiers have correct retention
	tests := []struct {
		tier     models.LogTier
		minDays  int
	}{
		{models.TierAccess, 30},
		{models.TierOperation, 180},
		{models.TierAudit, 2555}, // ~7 years
	}
	for _, tt := range tests {
		days, ok := models.TierRetentionDays[tt.tier]
		if !ok {
			t.Fatalf("tier %s missing from retention map", tt.tier)
		}
		if days < tt.minDays {
			t.Errorf("tier %s: expected at least %d days, got %d", tt.tier, tt.minDays, days)
		}
	}
}

func TestAuditLogModelHasSnapshotFields(t *testing.T) {
	// Verify the model struct has the extended fields
	log := models.AuditLog{
		Tier:        models.TierAudit,
		Reason:      strPtr("test reason"),
		Source:      strPtr("api"),
		Workstation: strPtr("test-ws"),
	}

	if log.Tier != models.TierAudit {
		t.Error("Tier field not working")
	}
	if log.Reason == nil || *log.Reason != "test reason" {
		t.Error("Reason field not working")
	}
}

func strPtr(s string) *string {
	return &s
}
