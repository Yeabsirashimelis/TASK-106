package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

func TestValidRoles(t *testing.T) {
	expected := []models.Role{
		models.RoleAdministrator,
		models.RoleScheduler,
		models.RoleInstructor,
		models.RoleReviewer,
		models.RoleFinanceClerk,
		models.RoleAuditor,
	}
	for _, r := range expected {
		if !models.ValidRoles[r] {
			t.Errorf("expected %s to be a valid role", r)
		}
	}
	if models.ValidRoles[models.Role("SuperAdmin")] {
		t.Error("SuperAdmin should not be a valid role")
	}
}

func TestValidAccountStatuses(t *testing.T) {
	expected := []models.Status{
		models.StatusActive,
		models.StatusFrozen,
		models.StatusDeactivated,
	}
	for _, s := range expected {
		if !models.ValidStatuses[s] {
			t.Errorf("expected %s to be a valid status", s)
		}
	}
	if models.ValidStatuses[models.Status("Banned")] {
		t.Error("Banned should not be a valid status")
	}
}
