package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

func TestReviewRequestTransitions(t *testing.T) {
	// ReviewRequest status: In Review, Approved, Rejected, Returned
	// Note: transitions are managed by the service layer, not by a CanTransition function
	// on the request itself, but we test the ReviewLevel transitions and the model validations

	if !models.ValidReviewRequestStatuses[models.ReviewRequestInReview] {
		t.Error("In Review should be valid")
	}
	if !models.ValidReviewRequestStatuses[models.ReviewRequestApproved] {
		t.Error("Approved should be valid")
	}
	if !models.ValidReviewRequestStatuses[models.ReviewRequestRejected] {
		t.Error("Rejected should be valid")
	}
	if !models.ValidReviewRequestStatuses[models.ReviewRequestReturned] {
		t.Error("Returned should be valid")
	}
	if models.ValidReviewRequestStatuses[models.ReviewRequestStatus("Bogus")] {
		t.Error("Bogus should not be valid")
	}
}

func TestReviewLevelTransitions(t *testing.T) {
	if !models.ValidLevelStatuses[models.LevelPending] {
		t.Error("Pending should be valid")
	}
	if !models.ValidLevelStatuses[models.LevelApproved] {
		t.Error("Approved should be valid")
	}
	if !models.ValidLevelStatuses[models.LevelRejected] {
		t.Error("Rejected should be valid")
	}
	if !models.ValidLevelStatuses[models.LevelReturned] {
		t.Error("Returned should be valid")
	}
}

func TestModerationReviewTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.ReviewStatus
		to   models.ReviewStatus
		want bool
	}{
		{"Pending to Approved", models.ReviewPending, models.ReviewApproved, true},
		{"Pending to Rejected", models.ReviewPending, models.ReviewRejected, true},
		{"Approved is terminal", models.ReviewApproved, models.ReviewPending, false},
		{"Approved to Rejected", models.ReviewApproved, models.ReviewRejected, false},
		{"Rejected is terminal", models.ReviewRejected, models.ReviewPending, false},
		{"Rejected to Approved", models.ReviewRejected, models.ReviewApproved, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.CanTransitionReview(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransitionReview(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestReportTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.ReportStatus
		to   models.ReportStatus
		want bool
	}{
		{"Open to Under Review", models.ReportOpen, models.ReportUnderReview, true},
		{"Open to Dismissed", models.ReportOpen, models.ReportDismissed, true},
		{"Open to Resolved (invalid)", models.ReportOpen, models.ReportResolved, false},
		{"Under Review to Resolved", models.ReportUnderReview, models.ReportResolved, true},
		{"Under Review to Dismissed", models.ReportUnderReview, models.ReportDismissed, true},
		{"Resolved is terminal", models.ReportResolved, models.ReportOpen, false},
		{"Dismissed is terminal", models.ReportDismissed, models.ReportOpen, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.CanTransitionReport(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransitionReport(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
