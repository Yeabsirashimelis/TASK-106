package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

func TestMatchTransitions_ValidPaths(t *testing.T) {
	tests := []struct {
		name string
		from models.MatchStatus
		to   models.MatchStatus
		want bool
	}{
		// Draft transitions
		{"Draft to Scheduled", models.MatchDraft, models.MatchScheduled, true},
		{"Draft to Canceled", models.MatchDraft, models.MatchCanceled, true},
		{"Draft to InProgress", models.MatchDraft, models.MatchInProgress, false},
		{"Draft to Final", models.MatchDraft, models.MatchFinal, false},

		// Scheduled transitions
		{"Scheduled to InProgress", models.MatchScheduled, models.MatchInProgress, true},
		{"Scheduled to Canceled", models.MatchScheduled, models.MatchCanceled, true},
		{"Scheduled to Draft", models.MatchScheduled, models.MatchDraft, false},
		{"Scheduled to Final", models.MatchScheduled, models.MatchFinal, false},

		// InProgress transitions
		{"InProgress to Final", models.MatchInProgress, models.MatchFinal, true},
		{"InProgress to Canceled", models.MatchInProgress, models.MatchCanceled, true},
		{"InProgress to Draft", models.MatchInProgress, models.MatchDraft, false},
		{"InProgress to Scheduled", models.MatchInProgress, models.MatchScheduled, false},

		// Terminal states
		{"Final to anything", models.MatchFinal, models.MatchDraft, false},
		{"Final to Canceled", models.MatchFinal, models.MatchCanceled, false},
		{"Canceled to anything", models.MatchCanceled, models.MatchDraft, false},
		{"Canceled to Scheduled", models.MatchCanceled, models.MatchScheduled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.CanTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestMatchTransitions_FullLifecycle(t *testing.T) {
	// Normal lifecycle: Draft -> Scheduled -> InProgress -> Final
	lifecycle := []models.MatchStatus{
		models.MatchDraft,
		models.MatchScheduled,
		models.MatchInProgress,
		models.MatchFinal,
	}
	for i := 0; i < len(lifecycle)-1; i++ {
		if !models.CanTransition(lifecycle[i], lifecycle[i+1]) {
			t.Errorf("expected transition from %s to %s to be valid", lifecycle[i], lifecycle[i+1])
		}
	}
}

func TestMatchTransitions_InvalidStatus(t *testing.T) {
	got := models.CanTransition(models.MatchStatus("NonExistent"), models.MatchDraft)
	if got {
		t.Error("expected false for invalid status")
	}
}
