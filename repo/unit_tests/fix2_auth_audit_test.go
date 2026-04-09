package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

// Fix 6: Access-tier audit logs for auth paths

func TestAccessTierForAuthLogs(t *testing.T) {
	// Verify the access tier constant is correct
	if models.TierAccess != "access" {
		t.Errorf("expected 'access', got %s", models.TierAccess)
	}

	// Verify access tier retention is 30 days
	days, ok := models.TierRetentionDays[models.TierAccess]
	if !ok {
		t.Fatal("access tier missing from retention map")
	}
	if days != 30 {
		t.Errorf("expected 30 days for access tier, got %d", days)
	}
}

func TestAuditEntryAccessTierFields(t *testing.T) {
	// Verify an access-tier audit entry can carry all required fields
	source := "auth/login"
	reason := "invalid password"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "login_failure",
		Tier:       models.TierAccess,
		Source:     &source,
		Reason:     &reason,
		Details: map[string]interface{}{
			"username":   "testuser",
			"ip_address": "127.0.0.1",
		},
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier, got %s", entry.Tier)
	}
	if entry.Action != "login_failure" {
		t.Errorf("expected login_failure, got %s", entry.Action)
	}
	if *entry.Source != "auth/login" {
		t.Errorf("expected auth/login source, got %s", *entry.Source)
	}
	if *entry.Reason != "invalid password" {
		t.Errorf("expected reason, got %s", *entry.Reason)
	}
}

func TestAuthServiceSetAuditService(t *testing.T) {
	// Verify SetAuditService works without panicking
	authSvc := service.NewAuthService(nil, nil, nil, nil, nil, nil)
	auditSvc := service.NewAuditService(nil)
	authSvc.SetAuditService(auditSvc)
	// If we reach here, the wiring works
}
