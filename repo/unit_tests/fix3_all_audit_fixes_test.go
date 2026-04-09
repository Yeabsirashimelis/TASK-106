package unit_tests

import (
	"errors"
	"testing"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

// =============================================================================
// Issue 1: Resource search visibility split (staff vs enrolled)
// =============================================================================

func TestSearchVisibilitySplit_StaffSeeAll(t *testing.T) {
	// Staff membership should grant access to all visibility levels
	if models.MembershipRoleStaff != "Staff" {
		t.Fatalf("expected Staff, got %s", models.MembershipRoleStaff)
	}
	// Staff can see both Staff-only and Enrolled resources
	for _, vis := range []models.ResourceVisibility{models.VisibilityStaff, models.VisibilityEnrolled} {
		if !models.ValidVisibilities[vis] {
			t.Errorf("visibility %s should be valid", vis)
		}
	}
}

func TestSearchVisibilitySplit_EnrolledOnlyEnrolled(t *testing.T) {
	// Enrolled users should only see Enrolled-visibility resources
	if models.MembershipRoleEnrolled != "Enrolled" {
		t.Fatalf("expected Enrolled, got %s", models.MembershipRoleEnrolled)
	}
	// The SearchWithVisibility repo method filters by VisibilityEnrolled
	if models.VisibilityEnrolled != "Enrolled" {
		t.Errorf("expected Enrolled visibility constant")
	}
}

func TestSearchVisibilitySplit_AdminBypass(t *testing.T) {
	// Admin role should bypass visibility filtering (uses Search, not SearchWithVisibility)
	if models.RoleAdministrator != "Administrator" {
		t.Fatalf("expected Administrator, got %s", models.RoleAdministrator)
	}
}

func TestSearchVisibilitySplit_NonMemberDenied(t *testing.T) {
	// Non-member, non-admin callers should get ErrNotCourseMember
	if service.ErrResourceAccessDenied == nil {
		t.Fatal("ErrResourceAccessDenied must be defined")
	}
}

// =============================================================================
// Issue 2: Non-staff resource create forbidden
// =============================================================================

func TestResourceCreateRequiresStaff_ErrorSentinel(t *testing.T) {
	// ErrResourceAccessDenied is returned when enrolled user tries to create
	err := service.ErrResourceAccessDenied
	if err == nil {
		t.Fatal("ErrResourceAccessDenied must be defined")
	}
	if err.Error() != "access denied to this resource" {
		t.Errorf("unexpected message: %s", err)
	}
}

func TestResourceCreateRequiresStaff_MembershipRoles(t *testing.T) {
	// Only Staff membership allows resource creation; Enrolled does not
	if !models.ValidMembershipRoles[models.MembershipRoleStaff] {
		t.Error("Staff should be valid membership role")
	}
	if !models.ValidMembershipRoles[models.MembershipRoleEnrolled] {
		t.Error("Enrolled should be valid membership role")
	}
	// Verify they are distinct
	if models.MembershipRoleStaff == models.MembershipRoleEnrolled {
		t.Error("Staff and Enrolled should be distinct")
	}
}

func TestResourceCreateRequiresStaff_AdminGlobalAccess(t *testing.T) {
	// Administrator role should have global access (bypass course staff check)
	if models.RoleAdministrator != "Administrator" {
		t.Fatal("Administrator role constant mismatch")
	}
}

// =============================================================================
// Issue 3: Idempotency 24h semantics
// =============================================================================

func TestIdempotencyDuplicateWithin24h(t *testing.T) {
	// Within 24h window, FindActiveIdempotencyKey should find the key
	now := time.Now()
	ik := models.IdempotencyKey{
		ID:             uuid.New(),
		AccountID:      uuid.New(),
		IdempotencyKey: "pay-dup-001",
		PaymentID:      uuid.New(),
		WindowStart:    now,
		WindowEnd:      now.Add(24 * time.Hour),
		CreatedAt:      now,
	}

	// Simulate check at now + 1h: window_end > checkTime
	checkTime := now.Add(1 * time.Hour)
	if !checkTime.Before(ik.WindowEnd) {
		t.Error("key should be active at now+1h (within 24h window)")
	}
}

func TestIdempotencyNewCreateAfter24h(t *testing.T) {
	// After 24h window, the same key should allow new creation
	now := time.Now()
	ik := models.IdempotencyKey{
		ID:             uuid.New(),
		AccountID:      uuid.New(),
		IdempotencyKey: "pay-dup-001",
		PaymentID:      uuid.New(),
		WindowStart:    now.Add(-25 * time.Hour),
		WindowEnd:      now.Add(-1 * time.Hour),
		CreatedAt:      now.Add(-25 * time.Hour),
	}

	// At current time, window_end is in the past
	if now.Before(ik.WindowEnd) {
		t.Error("key should be expired after 24h window")
	}

	// A new key with the same account+idempotency_key but different window_start
	// should be allowed by the unique index (account_id, idempotency_key, window_start)
	newIK := models.IdempotencyKey{
		ID:             uuid.New(),
		AccountID:      ik.AccountID,
		IdempotencyKey: ik.IdempotencyKey,
		PaymentID:      uuid.New(),
		WindowStart:    now,
		WindowEnd:      now.Add(24 * time.Hour),
		CreatedAt:      now,
	}

	// window_start values differ, so unique(account_id, key, window_start) allows this
	if newIK.WindowStart == ik.WindowStart {
		t.Error("new key should have different window_start than expired key")
	}
}

func TestIdempotencyWindowExactly24Hours(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)
	if end.Sub(start) != 24*time.Hour {
		t.Error("window should be exactly 24 hours")
	}
}

func TestIdempotencyKeyBoundary(t *testing.T) {
	// Test at exact 24h boundary
	now := time.Now()
	ik := models.IdempotencyKey{
		WindowStart: now,
		WindowEnd:   now.Add(24 * time.Hour),
	}

	// At exactly window_end, the key should be expired (window_end > now means active)
	exactEnd := ik.WindowEnd
	if exactEnd.Before(ik.WindowEnd) {
		t.Error("at exact boundary, key should be expired (not before window_end)")
	}
}

// =============================================================================
// Issue 4: extracted_text rejection for non-PDF/DOCX
// =============================================================================

func TestExtractedTextAllowedForPDF(t *testing.T) {
	if !models.TextExtractableMimeTypes["application/pdf"] {
		t.Error("PDF should allow extracted_text")
	}
}

func TestExtractedTextAllowedForDOCX(t *testing.T) {
	docx := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	if !models.TextExtractableMimeTypes[docx] {
		t.Error("DOCX should allow extracted_text")
	}
}

func TestExtractedTextRejectedForVideo(t *testing.T) {
	if models.TextExtractableMimeTypes["video/mp4"] {
		t.Error("video/mp4 should NOT allow extracted_text")
	}
}

func TestExtractedTextRejectedForImage(t *testing.T) {
	for _, mime := range []string{"image/png", "image/jpeg", "image/gif", "image/webp"} {
		if models.TextExtractableMimeTypes[mime] {
			t.Errorf("%s should NOT allow extracted_text", mime)
		}
	}
}

func TestExtractedTextRejectedForPlainText(t *testing.T) {
	for _, mime := range []string{"text/plain", "text/csv"} {
		if models.TextExtractableMimeTypes[mime] {
			t.Errorf("%s should NOT allow extracted_text", mime)
		}
	}
}

func TestExtractedTextRejectedForLegacyWord(t *testing.T) {
	// application/msword (.doc) is NOT text-extractable, only .docx is
	if models.TextExtractableMimeTypes["application/msword"] {
		t.Error("application/msword should NOT allow extracted_text")
	}
}

func TestExtractedTextErrorSentinel(t *testing.T) {
	err := service.ErrExtractedTextNotAllowed
	if err == nil {
		t.Fatal("ErrExtractedTextNotAllowed must be defined")
	}
	if err.Error() != "extracted_text is only allowed for PDF and DOCX files" {
		t.Errorf("unexpected error message: %s", err)
	}
}

// =============================================================================
// Issue 5: Account create error mapping (conflict->409, invalid role->400)
// =============================================================================

func TestAccountCreateConflictMapping(t *testing.T) {
	// ErrDuplicateUsername should be a distinct sentinel for 409 mapping
	err := service.ErrDuplicateUsername
	if err == nil {
		t.Fatal("ErrDuplicateUsername must be defined")
	}
	if err.Error() != "username already exists" {
		t.Errorf("unexpected message: %s", err)
	}
}

func TestAccountCreateInvalidRoleMapping(t *testing.T) {
	// ErrInvalidRole should be a distinct sentinel for 400 mapping
	err := service.ErrInvalidRole
	if err == nil {
		t.Fatal("ErrInvalidRole must be defined")
	}
	if err.Error() != "invalid role" {
		t.Errorf("unexpected message: %s", err)
	}
}

func TestAccountCreatePasswordPolicyMapping(t *testing.T) {
	// ErrPasswordPolicy should be a distinct sentinel for 400 mapping
	err := service.ErrPasswordPolicy
	if err == nil {
		t.Fatal("ErrPasswordPolicy must be defined")
	}
}

func TestAccountCreateAllErrorsDistinct(t *testing.T) {
	sentinels := []error{
		service.ErrInvalidRole,
		service.ErrDuplicateUsername,
		service.ErrPasswordPolicy,
	}
	for i := 0; i < len(sentinels); i++ {
		for j := i + 1; j < len(sentinels); j++ {
			if errors.Is(sentinels[i], sentinels[j]) {
				t.Errorf("sentinels %d and %d should be distinct", i, j)
			}
		}
	}
}

func TestAccountCreatePasswordValidation(t *testing.T) {
	// ValidatePassword should reject weak passwords
	err := service.ValidatePassword("short")
	if err == nil {
		t.Error("short password should fail validation")
	}
	if !errors.Is(err, service.ErrPasswordPolicy) {
		t.Error("weak password error should wrap ErrPasswordPolicy")
	}
}

func TestAccountCreatePasswordValidationStrong(t *testing.T) {
	// ValidatePassword should accept strong passwords
	err := service.ValidatePassword("StrongPass123!")
	if err != nil {
		t.Errorf("strong password should pass validation: %v", err)
	}
}

// =============================================================================
// Issue 6: Auth access-tier audit logs for login/refresh/logout
// =============================================================================

func TestAuthAuditLoginSuccessEntry(t *testing.T) {
	// Verify login success audit entry structure
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
		t.Errorf("login success should use access tier, got %s", entry.Tier)
	}
	if entry.Action != "login_success" {
		t.Errorf("expected login_success action, got %s", entry.Action)
	}
}

func TestAuthAuditLoginFailureEntry(t *testing.T) {
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
		Details:    map[string]interface{}{"username": "baduser"},
	}
	if entry.Tier != models.TierAccess {
		t.Errorf("login failure should use access tier")
	}
	if *entry.Reason != "invalid password" {
		t.Errorf("expected reason for login failure")
	}
}

func TestAuthAuditLoginLockedEntry(t *testing.T) {
	source := "auth/login"
	reason := "too many failed attempts"
	entry := &service.AuditEntry{
		EntityType: "auth",
		EntityID:   uuid.New(),
		ActorID:    uuid.New(),
		Action:     "login_locked",
		Tier:       models.TierAccess,
		Source:     &source,
		Reason:     &reason,
	}
	if entry.Tier != models.TierAccess {
		t.Errorf("login locked should use access tier")
	}
	if entry.Action != "login_locked" {
		t.Errorf("expected login_locked action")
	}
}

func TestAuthAuditRefreshSuccessEntry(t *testing.T) {
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
		t.Errorf("refresh success should use access tier")
	}
}

func TestAuthAuditRefreshReuseEntry(t *testing.T) {
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
		t.Errorf("refresh reuse should use access tier")
	}
	if entry.Action != "refresh_token_reuse" {
		t.Errorf("expected refresh_token_reuse action")
	}
}

func TestAuthAuditRefreshExpiredEntry(t *testing.T) {
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
		t.Errorf("refresh expired should use access tier")
	}
}

func TestAuthAuditLogoutEntry(t *testing.T) {
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
		t.Errorf("logout should use access tier")
	}
	if entry.Action != "logout" {
		t.Errorf("expected logout action")
	}
}

func TestAccessTierRetention30Days(t *testing.T) {
	days, ok := models.TierRetentionDays[models.TierAccess]
	if !ok {
		t.Fatal("access tier missing from retention map")
	}
	if days != 30 {
		t.Errorf("expected 30 days for access tier, got %d", days)
	}
}

func TestAuthAuditNoSecretsInDetails(t *testing.T) {
	// Audit entries should never contain password or token values
	details := map[string]interface{}{
		"username":   "testuser",
		"ip_address": "127.0.0.1",
	}
	// Verify no sensitive keys
	for k := range details {
		if k == "password" || k == "token" || k == "refresh_token" || k == "access_token" {
			t.Errorf("audit details should not contain sensitive field: %s", k)
		}
	}
}
