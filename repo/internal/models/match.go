package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MatchStatus string

const (
	MatchDraft      MatchStatus = "Draft"
	MatchScheduled  MatchStatus = "Scheduled"
	MatchInProgress MatchStatus = "In-Progress"
	MatchFinal      MatchStatus = "Final"
	MatchCanceled   MatchStatus = "Canceled"
)

var ValidMatchStatuses = map[MatchStatus]bool{
	MatchDraft:      true,
	MatchScheduled:  true,
	MatchInProgress: true,
	MatchFinal:      true,
	MatchCanceled:   true,
}

// ValidMatchTransitions defines allowed status transitions
var ValidMatchTransitions = map[MatchStatus][]MatchStatus{
	MatchDraft:      {MatchScheduled, MatchCanceled},
	MatchScheduled:  {MatchInProgress, MatchCanceled},
	MatchInProgress: {MatchFinal, MatchCanceled},
	MatchFinal:      {},
	MatchCanceled:   {},
}

func CanTransition(from, to MatchStatus) bool {
	allowed, ok := ValidMatchTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (s *MatchStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("match status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("match status must be a string")
		}
		sv = string(bv)
	}
	*s = MatchStatus(sv)
	return nil
}

func (s MatchStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type Match struct {
	ID             uuid.UUID   `db:"id" json:"id"`
	SeasonID       uuid.UUID   `db:"season_id" json:"season_id"`
	Round          int         `db:"round" json:"round"`
	HomeTeamID     uuid.UUID   `db:"home_team_id" json:"home_team_id"`
	AwayTeamID     uuid.UUID   `db:"away_team_id" json:"away_team_id"`
	VenueID        uuid.UUID   `db:"venue_id" json:"venue_id"`
	ScheduledAt    time.Time   `db:"scheduled_at" json:"scheduled_at"`
	Status         MatchStatus `db:"status" json:"status"`
	OverrideReason *string     `db:"override_reason" json:"override_reason,omitempty"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}
