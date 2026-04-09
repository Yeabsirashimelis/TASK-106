package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MatchRepository struct {
	db *sqlx.DB
}

func NewMatchRepository(db *sqlx.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `INSERT INTO matches (id, season_id, round, home_team_id, away_team_id, venue_id, scheduled_at, status, override_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, query,
		match.ID, match.SeasonID, match.Round, match.HomeTeamID, match.AwayTeamID,
		match.VenueID, match.ScheduledAt, match.Status, match.OverrideReason,
		match.CreatedAt, match.UpdatedAt)
	if err != nil {
		return fmt.Errorf("match_repo.Create: %w", err)
	}
	return nil
}

func (r *MatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	var match models.Match
	query := `SELECT id, season_id, round, home_team_id, away_team_id, venue_id, scheduled_at, status, override_reason, created_at, updated_at
		FROM matches WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&match); err != nil {
		return nil, fmt.Errorf("match_repo.GetByID: %w", err)
	}
	return &match, nil
}

func (r *MatchRepository) ListBySeason(ctx context.Context, seasonID uuid.UUID, offset, limit int) ([]models.Match, error) {
	var matches []models.Match
	query := `SELECT id, season_id, round, home_team_id, away_team_id, venue_id, scheduled_at, status, override_reason, created_at, updated_at
		FROM matches WHERE season_id = $1 ORDER BY round, scheduled_at LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &matches, query, seasonID, limit, offset); err != nil {
		return nil, fmt.Errorf("match_repo.ListBySeason: %w", err)
	}
	return matches, nil
}

func (r *MatchRepository) ListByRound(ctx context.Context, seasonID uuid.UUID, round int) ([]models.Match, error) {
	var matches []models.Match
	query := `SELECT id, season_id, round, home_team_id, away_team_id, venue_id, scheduled_at, status, override_reason, created_at, updated_at
		FROM matches WHERE season_id = $1 AND round = $2 ORDER BY scheduled_at`
	if err := r.db.SelectContext(ctx, &matches, query, seasonID, round); err != nil {
		return nil, fmt.Errorf("match_repo.ListByRound: %w", err)
	}
	return matches, nil
}

func (r *MatchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MatchStatus, overrideReason *string) error {
	query := `UPDATE matches SET status = $1, override_reason = COALESCE($2, override_reason), updated_at = NOW() WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, overrideReason, id)
	if err != nil {
		return fmt.Errorf("match_repo.UpdateStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("match_repo.UpdateStatus: match not found")
	}
	return nil
}

func (r *MatchRepository) Update(ctx context.Context, match *models.Match) error {
	query := `UPDATE matches SET round = $1, home_team_id = $2, away_team_id = $3, venue_id = $4,
		scheduled_at = $5, override_reason = $6, updated_at = NOW()
		WHERE id = $7 AND status = 'Draft'`
	result, err := r.db.ExecContext(ctx, query,
		match.Round, match.HomeTeamID, match.AwayTeamID, match.VenueID,
		match.ScheduledAt, match.OverrideReason, match.ID)
	if err != nil {
		return fmt.Errorf("match_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("match_repo.Update: match not found or not in Draft status")
	}
	return nil
}

// CheckVenueConflict checks if a venue has any overlapping matches within 90 minutes
func (r *MatchRepository) CheckVenueConflict(ctx context.Context, venueID uuid.UUID, scheduledAt time.Time, excludeMatchID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM matches
		WHERE venue_id = $1
		AND status NOT IN ('Canceled')
		AND ABS(EXTRACT(EPOCH FROM (scheduled_at - $2::timestamptz))) < 5400
		AND ($3::uuid IS NULL OR id != $3)`
	var count int
	if err := r.db.QueryRowxContext(ctx, query, venueID, scheduledAt, excludeMatchID).Scan(&count); err != nil {
		return false, fmt.Errorf("match_repo.CheckVenueConflict: %w", err)
	}
	return count > 0, nil
}

// CheckDuplicatePairing checks if a pairing already exists in the given round
func (r *MatchRepository) CheckDuplicatePairing(ctx context.Context, seasonID uuid.UUID, round int, homeTeamID, awayTeamID uuid.UUID, excludeMatchID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM matches
		WHERE season_id = $1 AND round = $2
		AND status != 'Canceled'
		AND ((home_team_id = $3 AND away_team_id = $4) OR (home_team_id = $4 AND away_team_id = $3))
		AND ($5::uuid IS NULL OR id != $5)`
	var count int
	if err := r.db.QueryRowxContext(ctx, query, seasonID, round, homeTeamID, awayTeamID, excludeMatchID).Scan(&count); err != nil {
		return false, fmt.Errorf("match_repo.CheckDuplicatePairing: %w", err)
	}
	return count > 0, nil
}

// GetConsecutiveHomeAway returns the count of consecutive home or away games for a team
// ending at the proposed match, ordered by scheduled_at.
func (r *MatchRepository) CountConsecutiveHomeGames(ctx context.Context, seasonID, teamID uuid.UUID, scheduledAt time.Time, excludeMatchID *uuid.UUID) (int, error) {
	// Get the last 3 home games before and including this scheduled time
	query := `SELECT COUNT(*) FROM (
		SELECT home_team_id FROM matches
		WHERE season_id = $1
		AND home_team_id = $2
		AND scheduled_at <= $3
		AND status != 'Canceled'
		AND ($4::uuid IS NULL OR id != $4)
		ORDER BY scheduled_at DESC
		LIMIT 3
	) sub`
	var count int
	if err := r.db.QueryRowxContext(ctx, query, seasonID, teamID, scheduledAt, excludeMatchID).Scan(&count); err != nil {
		return 0, fmt.Errorf("match_repo.CountConsecutiveHomeGames: %w", err)
	}
	return count, nil
}

func (r *MatchRepository) CountConsecutiveAwayGames(ctx context.Context, seasonID, teamID uuid.UUID, scheduledAt time.Time, excludeMatchID *uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM (
		SELECT away_team_id FROM matches
		WHERE season_id = $1
		AND away_team_id = $2
		AND scheduled_at <= $3
		AND status != 'Canceled'
		AND ($4::uuid IS NULL OR id != $4)
		ORDER BY scheduled_at DESC
		LIMIT 3
	) sub`
	var count int
	if err := r.db.QueryRowxContext(ctx, query, seasonID, teamID, scheduledAt, excludeMatchID).Scan(&count); err != nil {
		return 0, fmt.Errorf("match_repo.CountConsecutiveAwayGames: %w", err)
	}
	return count, nil
}

// GetTeamRecentMatches gets a team's recent matches (both home and away) ordered by scheduled_at DESC
func (r *MatchRepository) GetTeamRecentMatches(ctx context.Context, seasonID, teamID uuid.UUID, beforeTime time.Time, limit int) ([]models.Match, error) {
	var matches []models.Match
	query := `SELECT id, season_id, round, home_team_id, away_team_id, venue_id, scheduled_at, status, override_reason, created_at, updated_at
		FROM matches
		WHERE season_id = $1
		AND (home_team_id = $2 OR away_team_id = $2)
		AND scheduled_at < $3
		AND status != 'Canceled'
		ORDER BY scheduled_at DESC
		LIMIT $4`
	if err := r.db.SelectContext(ctx, &matches, query, seasonID, teamID, beforeTime, limit); err != nil {
		return nil, fmt.Errorf("match_repo.GetTeamRecentMatches: %w", err)
	}
	return matches, nil
}
