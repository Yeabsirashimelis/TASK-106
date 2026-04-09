package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type CreateMatchRequest struct {
	SeasonID       string  `json:"season_id"`
	Round          int     `json:"round"`
	HomeTeamID     string  `json:"home_team_id"`
	AwayTeamID     string  `json:"away_team_id"`
	VenueID        string  `json:"venue_id"`
	ScheduledAt    string  `json:"scheduled_at"`
	OverrideReason *string `json:"override_reason,omitempty"`
}

type UpdateMatchRequest struct {
	Round          *int    `json:"round,omitempty"`
	HomeTeamID     *string `json:"home_team_id,omitempty"`
	AwayTeamID     *string `json:"away_team_id,omitempty"`
	VenueID        *string `json:"venue_id,omitempty"`
	ScheduledAt    *string `json:"scheduled_at,omitempty"`
	OverrideReason *string `json:"override_reason,omitempty"`
}

type TransitionMatchRequest struct {
	Status         string  `json:"status"`
	OverrideReason *string `json:"override_reason,omitempty"`
}

type ImportMatchEntry struct {
	Round          int     `json:"round"`
	HomeTeamID     string  `json:"home_team_id"`
	AwayTeamID     string  `json:"away_team_id"`
	VenueID        string  `json:"venue_id"`
	ScheduledAt    string  `json:"scheduled_at"`
	OverrideReason *string `json:"override_reason,omitempty"`
}

type GenerateScheduleRequest struct {
	SeasonID       string   `json:"season_id"`
	VenueIDs       []string `json:"venue_ids"`
	StartDate      string   `json:"start_date"`
	IntervalDays   int      `json:"interval_days"`
	StartTime      string   `json:"start_time"`
}

type GenerateScheduleResponse struct {
	Created    int              `json:"created"`
	Rounds     int              `json:"rounds"`
	Errors     []ImportError    `json:"errors,omitempty"`
	Matches    []MatchResponse  `json:"matches"`
}

type ImportScheduleRequest struct {
	SeasonID string             `json:"season_id"`
	Matches  []ImportMatchEntry `json:"matches"`
}

type ImportScheduleResponse struct {
	Created    int              `json:"created"`
	Errors     []ImportError    `json:"errors,omitempty"`
	Matches    []MatchResponse  `json:"matches"`
}

type ImportError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type MatchResponse struct {
	ID             uuid.UUID `json:"id"`
	SeasonID       uuid.UUID `json:"season_id"`
	Round          int       `json:"round"`
	HomeTeamID     uuid.UUID `json:"home_team_id"`
	AwayTeamID     uuid.UUID `json:"away_team_id"`
	VenueID        uuid.UUID `json:"venue_id"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	Status         string    `json:"status"`
	OverrideReason *string   `json:"override_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func ToMatchResponse(m *models.Match) MatchResponse {
	return MatchResponse{
		ID:             m.ID,
		SeasonID:       m.SeasonID,
		Round:          m.Round,
		HomeTeamID:     m.HomeTeamID,
		AwayTeamID:     m.AwayTeamID,
		VenueID:        m.VenueID,
		ScheduledAt:    m.ScheduledAt,
		Status:         string(m.Status),
		OverrideReason: m.OverrideReason,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ToMatchResponseList(matches []models.Match) []MatchResponse {
	result := make([]MatchResponse, len(matches))
	for i, m := range matches {
		result[i] = ToMatchResponse(&m)
	}
	return result
}

type CreateAssignmentRequest struct {
	MatchID   string `json:"match_id"`
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
}

type ReassignRequest struct {
	NewAccountID string `json:"new_account_id"`
	Reason       string `json:"reason"`
}

type AssignmentResponse struct {
	ID                 uuid.UUID `json:"id"`
	MatchID            uuid.UUID `json:"match_id"`
	AccountID          uuid.UUID `json:"account_id"`
	Role               string    `json:"role"`
	AssignedBy         uuid.UUID `json:"assigned_by"`
	ReassignmentReason *string   `json:"reassignment_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func ToAssignmentResponse(a *models.MatchAssignment) AssignmentResponse {
	return AssignmentResponse{
		ID:                 a.ID,
		MatchID:            a.MatchID,
		AccountID:          a.AccountID,
		Role:               string(a.Role),
		AssignedBy:         a.AssignedBy,
		ReassignmentReason: a.ReassignmentReason,
		CreatedAt:          a.CreatedAt,
		UpdatedAt:          a.UpdatedAt,
	}
}

func ToAssignmentResponseList(assignments []models.MatchAssignment) []AssignmentResponse {
	result := make([]AssignmentResponse, len(assignments))
	for i, a := range assignments {
		result[i] = ToAssignmentResponse(&a)
	}
	return result
}
