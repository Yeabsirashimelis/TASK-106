package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

func TestCourseStatusValidation(t *testing.T) {
	valid := []models.CourseStatus{
		models.CourseStatusDraft,
		models.CourseStatusPublished,
		models.CourseStatusArchived,
	}
	for _, s := range valid {
		if !models.ValidCourseStatuses[s] {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if models.ValidCourseStatuses[models.CourseStatus("Deleted")] {
		t.Error("Deleted should not be valid")
	}
}

func TestOutlineNodeTypeValidation(t *testing.T) {
	if !models.ValidNodeTypes[models.NodeTypeChapter] {
		t.Error("Chapter should be valid")
	}
	if !models.ValidNodeTypes[models.NodeTypeUnit] {
		t.Error("Unit should be valid")
	}
	if models.ValidNodeTypes[models.OutlineNodeType("Section")] {
		t.Error("Section should not be valid")
	}
}

func TestMembershipRoleValidation(t *testing.T) {
	if !models.ValidMembershipRoles[models.MembershipRoleStaff] {
		t.Error("Staff should be valid")
	}
	if !models.ValidMembershipRoles[models.MembershipRoleEnrolled] {
		t.Error("Enrolled should be valid")
	}
	if models.ValidMembershipRoles[models.MembershipRole("Guest")] {
		t.Error("Guest should not be valid")
	}
}

func TestResourceTypeValidation(t *testing.T) {
	valid := []models.ResourceType{
		models.ResourceTypeDocument,
		models.ResourceTypeVideo,
		models.ResourceTypeLink,
	}
	for _, rt := range valid {
		if !models.ValidResourceTypes[rt] {
			t.Errorf("expected %s to be valid", rt)
		}
	}
}

func TestVisibilityValidation(t *testing.T) {
	if !models.ValidVisibilities[models.VisibilityStaff] {
		t.Error("Staff visibility should be valid")
	}
	if !models.ValidVisibilities[models.VisibilityEnrolled] {
		t.Error("Enrolled visibility should be valid")
	}
}

func TestMimeTypeAllowlist(t *testing.T) {
	allowed := []string{
		"application/pdf",
		"video/mp4",
		"image/png",
		"text/plain",
	}
	for _, m := range allowed {
		if !models.AllowedMimeTypes[m] {
			t.Errorf("expected %s to be allowed", m)
		}
	}

	blocked := []string{
		"application/x-executable",
		"application/javascript",
		"text/html",
	}
	for _, m := range blocked {
		if models.AllowedMimeTypes[m] {
			t.Errorf("expected %s to be blocked", m)
		}
	}
}

func TestTextExtractableMimeTypes(t *testing.T) {
	if !models.TextExtractableMimeTypes["application/pdf"] {
		t.Error("PDF should be text-extractable")
	}
	if !models.TextExtractableMimeTypes["application/vnd.openxmlformats-officedocument.wordprocessingml.document"] {
		t.Error("DOCX should be text-extractable")
	}
	if models.TextExtractableMimeTypes["video/mp4"] {
		t.Error("MP4 should not be text-extractable")
	}
}
