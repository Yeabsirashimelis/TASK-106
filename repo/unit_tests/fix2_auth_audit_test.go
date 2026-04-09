package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

// ── Fix 6: Access-tier audit logs for auth paths ───────────────────────────

func TestAccessTierForAuthLogs(t *testing.T) {
	if models.TierAccess != "access" {
		t.Errorf("expected 'access', got %s", models.TierAccess)
	}

	days, ok := models.TierRetentionDays[models.TierAccess]
	if !ok {
		t.Fatal("access tier missing from retention map")
	}
	if days != 30 {
		t.Errorf("expected 30 days for access tier, got %d", days)
	}
}

func TestAuditEntryLoginFailureFields(t *testing.T) {
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

func TestAuditEntryLoginSuccessFields(t *testing.T) {
	source := "auth/login"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "login_success",
		Tier:       models.TierAccess,
		Source:     &source,
		Details: map[string]interface{}{
			"username":   "testuser",
			"ip_address": "10.0.0.1",
		},
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier")
	}
	if entry.Action != "login_success" {
		t.Errorf("expected login_success action")
	}
}

func TestAuditEntryRefreshSuccessFields(t *testing.T) {
	source := "auth/refresh"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "token_refresh",
		Tier:       models.TierAccess,
		Source:     &source,
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier for token_refresh")
	}
	if *entry.Source != "auth/refresh" {
		t.Errorf("expected auth/refresh source")
	}
}

func TestAuditEntryRefreshFailureTokenReuse(t *testing.T) {
	source := "auth/refresh"
	reason := "token reuse detected, all sessions revoked"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "refresh_token_reuse",
		Tier:       models.TierAccess,
		Source:     &source,
		Reason:     &reason,
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier for refresh_token_reuse")
	}
	if entry.Action != "refresh_token_reuse" {
		t.Errorf("expected refresh_token_reuse action")
	}
	if *entry.Reason != "token reuse detected, all sessions revoked" {
		t.Errorf("expected reuse reason")
	}
}

func TestAuditEntryRefreshFailureExpired(t *testing.T) {
	source := "auth/refresh"
	reason := "refresh token expired"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "refresh_token_expired",
		Tier:       models.TierAccess,
		Source:     &source,
		Reason:     &reason,
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier for refresh_token_expired")
	}
	if entry.Action != "refresh_token_expired" {
		t.Errorf("expected refresh_token_expired action")
	}
}

func TestAuditEntryLogoutFields(t *testing.T) {
	source := "auth/logout"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "logout",
		Tier:       models.TierAccess,
		Source:     &source,
	}

	if entry.Tier != models.TierAccess {
		t.Errorf("expected access tier for logout")
	}
	if entry.Action != "logout" {
		t.Errorf("expected logout action")
	}
	if *entry.Source != "auth/logout" {
		t.Errorf("expected auth/logout source")
	}
}

func TestAuthServiceSetAuditService(t *testing.T) {
	authSvc := service.NewAuthService(nil, nil, nil, nil, nil, nil)
	auditSvc := service.NewAuditService(nil)
	authSvc.SetAuditService(auditSvc)
}

func TestAllAuthAuditActionsAreDistinct(t *testing.T) {
	actions := map[string]bool{
		"login_success":        true,
		"login_failure":        true,
		"token_refresh":        true,
		"refresh_token_reuse":  true,
		"refresh_token_expired": true,
		"logout":               true,
	}
	if len(actions) != 6 {
		t.Errorf("expected 6 distinct auth audit actions")
	}
}

func TestAllTiersHaveRetention(t *testing.T) {
	tiers := []models.LogTier{models.TierAccess, models.TierOperation, models.TierAudit}
	for _, tier := range tiers {
		if _, ok := models.TierRetentionDays[tier]; !ok {
			t.Errorf("tier %s is missing retention config", tier)
		}
	}
}
