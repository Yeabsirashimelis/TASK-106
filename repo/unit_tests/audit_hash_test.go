package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
)

func TestAuditEntryIntegrity(t *testing.T) {
	auditSvc := service.NewAuditService(nil) // nil repo - only testing hash verification

	// The VerifyEntryIntegrity function checks content_hash without needing a DB
	// Entries without a hash are assumed valid (pre-extension)
	entry := &models.AuditLog{
		ContentHash: nil,
	}
	if !auditSvc.VerifyEntryIntegrity(entry) {
		t.Error("entry without hash should be considered valid")
	}
}

func TestRetentionTierDays(t *testing.T) {
	tests := []struct {
		tier     models.LogTier
		expected int
	}{
		{models.TierAccess, 30},
		{models.TierOperation, 180},
		{models.TierAudit, 2555},
	}
	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			days, ok := models.TierRetentionDays[tt.tier]
			if !ok {
				t.Fatalf("tier %s not found in TierRetentionDays", tt.tier)
			}
			if days != tt.expected {
				t.Errorf("expected %d days for tier %s, got %d", tt.expected, tt.tier, days)
			}
		})
	}
}

func TestHashRefreshTokenDeterministic(t *testing.T) {
	input := "test-refresh-token-value"
	hash1 := service.HashRefreshToken(input)
	hash2 := service.HashRefreshToken(input)
	if hash1 != hash2 {
		t.Error("hashing same input should produce same result")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash1))
	}
}

func TestHashRefreshTokenDifferentInputs(t *testing.T) {
	hash1 := service.HashRefreshToken("input1")
	hash2 := service.HashRefreshToken("input2")
	if hash1 == hash2 {
		t.Error("different inputs should produce different hashes")
	}
}
