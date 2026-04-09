package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
)

// Fix 1: Search visibility split - enrolled users only see Enrolled-visibility
func TestSearchVisibilitySplitConstants(t *testing.T) {
	// Verify the visibility constants used for search filtering
	if models.VisibilityStaff != "Staff" {
		t.Errorf("expected Staff, got %s", models.VisibilityStaff)
	}
	if models.VisibilityEnrolled != "Enrolled" {
		t.Errorf("expected Enrolled, got %s", models.VisibilityEnrolled)
	}
}

// Fix 2: CreateResource requires course staff
func TestResourceAccessDeniedError(t *testing.T) {
	// The error sentinel used when non-staff tries to create/update resources
	if service.ErrResourceAccessDenied.Error() != "access denied to this resource" {
		t.Errorf("unexpected error message: %s", service.ErrResourceAccessDenied)
	}
}

// Fix 4: extracted_text rejection for non-PDF/DOCX
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
	}
	for _, m := range notExtractable {
		if models.TextExtractableMimeTypes[m] {
			t.Errorf("expected %s to NOT be text-extractable", m)
		}
	}
}
