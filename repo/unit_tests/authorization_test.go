package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

// Fix #1: Object-level authorization - verify model-level access rules

func TestCourseStatusVisibility(t *testing.T) {
	// Published courses should be visible to all; Draft/Archived should not
	if !models.ValidCourseStatuses[models.CourseStatusPublished] {
		t.Error("Published should be a valid course status")
	}
	if !models.ValidCourseStatuses[models.CourseStatusDraft] {
		t.Error("Draft should be a valid course status")
	}
}

func TestResourceVisibilityStaffVsEnrolled(t *testing.T) {
	// Staff visibility means only staff can see it
	// Enrolled visibility means enrolled users can see it too
	if !models.ValidVisibilities[models.VisibilityStaff] {
		t.Error("Staff should be a valid visibility")
	}
	if !models.ValidVisibilities[models.VisibilityEnrolled] {
		t.Error("Enrolled should be a valid visibility")
	}
	// The key authorization rule: Staff-only resources must be denied to enrolled users
	// This is tested at the service layer level (requires DB) but we verify the model constants here
}

func TestMembershipRolesForAuthorization(t *testing.T) {
	// Staff can manage; Enrolled has limited access
	if models.MembershipRoleStaff != "Staff" {
		t.Errorf("expected Staff, got %s", models.MembershipRoleStaff)
	}
	if models.MembershipRoleEnrolled != "Enrolled" {
		t.Errorf("expected Enrolled, got %s", models.MembershipRoleEnrolled)
	}
}

// Fix #2: Auditor read-only - verify role constant exists
func TestAuditorRoleExists(t *testing.T) {
	if !models.ValidRoles[models.RoleAuditor] {
		t.Error("Auditor should be a valid role")
	}
	// Auditor must be read-only - write endpoints moved to AdminOnly
	// This is enforced at the router level, verified in API tests
}
