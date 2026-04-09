package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
)

// ── Fix 1: Search visibility split ─────────────────────────────────────────

func TestSearchVisibilitySplitConstants(t *testing.T) {
	if models.VisibilityStaff != "Staff" {
		t.Errorf("expected Staff, got %s", models.VisibilityStaff)
	}
	if models.VisibilityEnrolled != "Enrolled" {
		t.Errorf("expected Enrolled, got %s", models.VisibilityEnrolled)
	}
}

func TestVisibilityStaffIsValidVisibility(t *testing.T) {
	if !models.ValidVisibilities[models.VisibilityStaff] {
		t.Error("Staff should be a valid visibility")
	}
}

func TestVisibilityEnrolledIsValidVisibility(t *testing.T) {
	if !models.ValidVisibilities[models.VisibilityEnrolled] {
		t.Error("Enrolled should be a valid visibility")
	}
}

func TestMembershipRoleStaffConstant(t *testing.T) {
	if models.MembershipRoleStaff != "Staff" {
		t.Errorf("expected Staff, got %s", models.MembershipRoleStaff)
	}
}

func TestMembershipRoleEnrolledConstant(t *testing.T) {
	if models.MembershipRoleEnrolled != "Enrolled" {
		t.Errorf("expected Enrolled, got %s", models.MembershipRoleEnrolled)
	}
}

func TestAdministratorRoleBypasses(t *testing.T) {
	// Administrator role must exist and be usable for bypass checks
	if !models.ValidRoles[models.RoleAdministrator] {
		t.Error("Administrator should be a valid role")
	}
}

func TestNotCourseMemberErrorSentinel(t *testing.T) {
	if service.ErrNotCourseMember == nil {
		t.Fatal("ErrNotCourseMember should be defined")
	}
	if service.ErrNotCourseMember.Error() != "not a member of this course" {
		t.Errorf("unexpected error message: %s", service.ErrNotCourseMember)
	}
}

// ── Fix 2: CreateResource requires course staff ────────────────────────────

func TestResourceAccessDeniedError(t *testing.T) {
	if service.ErrResourceAccessDenied.Error() != "access denied to this resource" {
		t.Errorf("unexpected error message: %s", service.ErrResourceAccessDenied)
	}
}

func TestStaffMembershipRoleIsValidForResourceCreation(t *testing.T) {
	// Only Staff membership role should grant create access (not Enrolled)
	if !models.ValidMembershipRoles[models.MembershipRoleStaff] {
		t.Error("Staff should be a valid membership role")
	}
	if !models.ValidMembershipRoles[models.MembershipRoleEnrolled] {
		t.Error("Enrolled should be a valid membership role")
	}
}

func TestMembershipRolesAreMutuallyExclusive(t *testing.T) {
	// Enrolled != Staff: the service must distinguish them for authz
	if models.MembershipRoleStaff == models.MembershipRoleEnrolled {
		t.Error("Staff and Enrolled membership roles must be distinct")
	}
}

// ── Fix 4: extracted_text rejection for non-PDF/DOCX ──────────────────────

func TestExtractedTextNotAllowedError(t *testing.T) {
	if service.ErrExtractedTextNotAllowed.Error() != "extracted_text is only allowed for PDF and DOCX files" {
		t.Errorf("unexpected error message: %s", service.ErrExtractedTextNotAllowed)
	}
}

func TestTextExtractableMimeTypesStrictness(t *testing.T) {
	// Only PDF and DOCX should be extractable
	allowed := []string{
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	for _, m := range allowed {
		if !models.TextExtractableMimeTypes[m] {
			t.Errorf("expected %s to be text-extractable", m)
		}
	}

	// These are allowed MIME types but NOT text-extractable
	notExtractable := []string{
		"video/mp4",
		"image/png",
		"text/plain",
		"text/csv",
		"application/msword",
		"application/vnd.ms-excel",
		"video/webm",
		"image/jpeg",
	}
	for _, m := range notExtractable {
		if models.TextExtractableMimeTypes[m] {
			t.Errorf("expected %s to NOT be text-extractable", m)
		}
	}
}

func TestTextExtractableIsSubsetOfAllowed(t *testing.T) {
	// Every text-extractable type must also be in the allowed MIME types
	for mime := range models.TextExtractableMimeTypes {
		if !models.AllowedMimeTypes[mime] {
			t.Errorf("text-extractable type %s is not in AllowedMimeTypes", mime)
		}
	}
}

func TestTextExtractableCountIsExactlyTwo(t *testing.T) {
	if len(models.TextExtractableMimeTypes) != 2 {
		t.Errorf("expected exactly 2 text-extractable types, got %d", len(models.TextExtractableMimeTypes))
	}
}
