package unit_tests

import (
	"errors"
	"testing"

	"github.com/eaglepoint/authapi/internal/service"
)

// Fix 5: Account create error sentinels for proper HTTP mapping

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
}
