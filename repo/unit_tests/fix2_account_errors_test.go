package unit_tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
)

// ── Fix 5: Account create error mapping ────────────────────────────────────

func TestInvalidRoleErrorSentinel(t *testing.T) {
	if service.ErrInvalidRole == nil {
		t.Fatal("ErrInvalidRole should be defined")
	}
	if service.ErrInvalidRole.Error() != "invalid role" {
		t.Errorf("unexpected message: %s", service.ErrInvalidRole)
	}
}

func TestDuplicateUsernameErrorSentinel(t *testing.T) {
	if service.ErrDuplicateUsername == nil {
		t.Fatal("ErrDuplicateUsername should be defined")
	}
	if service.ErrDuplicateUsername.Error() != "username already exists" {
		t.Errorf("unexpected message: %s", service.ErrDuplicateUsername)
	}
}

func TestPasswordPolicyErrorSentinel(t *testing.T) {
	if service.ErrPasswordPolicy == nil {
		t.Fatal("ErrPasswordPolicy should be defined")
	}
}

func TestErrorSentinelsAreDistinct(t *testing.T) {
	if errors.Is(service.ErrInvalidRole, service.ErrDuplicateUsername) {
		t.Error("ErrInvalidRole and ErrDuplicateUsername should be distinct")
	}
	if errors.Is(service.ErrInvalidRole, service.ErrPasswordPolicy) {
		t.Error("ErrInvalidRole and ErrPasswordPolicy should be distinct")
	}
	if errors.Is(service.ErrDuplicateUsername, service.ErrPasswordPolicy) {
		t.Error("ErrDuplicateUsername and ErrPasswordPolicy should be distinct")
	}
}

func TestWrappedInvalidRoleErrorIsDetectable(t *testing.T) {
	// The service wraps ErrInvalidRole with fmt.Errorf("%w: ..."), verify errors.Is works
	wrapped := fmt.Errorf("%w: FakeRole", service.ErrInvalidRole)
	if !errors.Is(wrapped, service.ErrInvalidRole) {
		t.Error("wrapped ErrInvalidRole should still be detectable via errors.Is")
	}
}

func TestWrappedDuplicateUsernameErrorIsDetectable(t *testing.T) {
	wrapped := fmt.Errorf("%w: testuser", service.ErrDuplicateUsername)
	if !errors.Is(wrapped, service.ErrDuplicateUsername) {
		t.Error("wrapped ErrDuplicateUsername should still be detectable via errors.Is")
	}
}

func TestPasswordValidationRejectsWeak(t *testing.T) {
	weakPasswords := []string{
		"short",           // too short
		"alllowercase12",  // no uppercase
		"ALLUPPERCASE12",  // no lowercase
		"NoDigitsHere!!",  // no digit
	}
	for _, pw := range weakPasswords {
		err := service.ValidatePassword(pw)
		if err == nil {
			t.Errorf("expected password %q to be rejected", pw)
		}
		if !errors.Is(err, service.ErrPasswordPolicy) {
			t.Errorf("expected ErrPasswordPolicy for %q, got: %v", pw, err)
		}
	}
}

func TestPasswordValidationAcceptsStrong(t *testing.T) {
	strong := "StrongP4ssw0rd!"
	if err := service.ValidatePassword(strong); err != nil {
		t.Errorf("expected strong password to pass validation, got: %v", err)
	}
}

func TestAllValidRolesAreRecognized(t *testing.T) {
	expected := []models.Role{
		models.RoleAdministrator,
		models.RoleScheduler,
		models.RoleInstructor,
		models.RoleReviewer,
		"Finance Clerk",
		"Auditor",
	}
	for _, r := range expected {
		if !models.ValidRoles[r] {
			t.Errorf("role %s should be valid", r)
		}
	}
}

func TestInvalidRoleIsRejected(t *testing.T) {
	if models.ValidRoles[models.Role("Hacker")] {
		t.Error("arbitrary role should not be valid")
	}
}
