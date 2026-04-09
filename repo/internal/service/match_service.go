package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrVenueConflict      = errors.New("venue has a conflicting match within 90 minutes")
	ErrDuplicatePairing   = errors.New("duplicate team pairing in this round")
	ErrConsecutiveHome    = errors.New("team would exceed 3 consecutive home games")
	ErrConsecutiveAway    = errors.New("team would exceed 3 consecutive away games")
	ErrOverrideRequired   = errors.New("override_reason is required to bypass scheduling conflict")
	ErrInvalidTransition  = errors.New("invalid match status transition")
	ErrMatchNotFound      = errors.New("match not found")
	ErrMatchNotDraft      = errors.New("match can only be edited in Draft status")
	ErrAssignmentLocked   = errors.New("assignments are locked once the match is In-Progress or beyond")
	ErrReassignmentReason = errors.New("reassignment reason is required")
)

const maxConsecutiveHomeAway = 3

type MatchService struct {
	matchRepo      *repository.MatchRepository
	seasonRepo     *repository.SeasonRepository
	teamRepo       *repository.TeamRepository
	venueRepo      *repository.VenueRepository
	assignmentRepo *repository.MatchAssignmentRepository
	audit          *AuditService
}

func NewMatchService(
	matchRepo *repository.MatchRepository,
	seasonRepo *repository.SeasonRepository,
	teamRepo *repository.TeamRepository,
	venueRepo *repository.VenueRepository,
	assignmentRepo *repository.MatchAssignmentRepository,
	audit *AuditService,
) *MatchService {
	return &MatchService{
		matchRepo:      matchRepo,
		seasonRepo:     seasonRepo,
		teamRepo:       teamRepo,
		venueRepo:      venueRepo,
		assignmentRepo: assignmentRepo,
		audit:          audit,
	}
}

type validationResult struct {
	violations []string
}

func (v *validationResult) hasViolations() bool {
	return len(v.violations) > 0
}

func (v *validationResult) summary() string {
	return strings.Join(v.violations, "; ")
}

func (s *MatchService) CreateMatch(ctx context.Context, req *dto.CreateMatchRequest, actorID uuid.UUID) (*models.Match, error) {
	seasonID, err := uuid.Parse(req.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("invalid season_id")
	}
	homeTeamID, err := uuid.Parse(req.HomeTeamID)
	if err != nil {
		return nil, fmt.Errorf("invalid home_team_id")
	}
	awayTeamID, err := uuid.Parse(req.AwayTeamID)
	if err != nil {
		return nil, fmt.Errorf("invalid away_team_id")
	}
	if homeTeamID == awayTeamID {
		return nil, fmt.Errorf("home and away teams must be different")
	}
	venueID, err := uuid.Parse(req.VenueID)
	if err != nil {
		return nil, fmt.Errorf("invalid venue_id")
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at, use RFC3339 format")
	}
	if req.Round < 1 {
		return nil, fmt.Errorf("round must be a positive integer")
	}

	// Verify referenced entities exist
	if _, err := s.seasonRepo.GetByID(ctx, seasonID); err != nil {
		return nil, fmt.Errorf("season not found")
	}
	if _, err := s.teamRepo.GetByID(ctx, homeTeamID); err != nil {
		return nil, fmt.Errorf("home team not found")
	}
	if _, err := s.teamRepo.GetByID(ctx, awayTeamID); err != nil {
		return nil, fmt.Errorf("away team not found")
	}
	if _, err := s.venueRepo.GetByID(ctx, venueID); err != nil {
		return nil, fmt.Errorf("venue not found")
	}

	// Run scheduling validations
	vr, err := s.validateScheduling(ctx, seasonID, req.Round, homeTeamID, awayTeamID, venueID, scheduledAt, nil)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if vr.hasViolations() {
		if req.OverrideReason == nil || *req.OverrideReason == "" {
			return nil, fmt.Errorf("%w: %s", ErrOverrideRequired, vr.summary())
		}
	}

	now := time.Now()
	match := &models.Match{
		ID:             uuid.New(),
		SeasonID:       seasonID,
		Round:          req.Round,
		HomeTeamID:     homeTeamID,
		AwayTeamID:     awayTeamID,
		VenueID:        venueID,
		ScheduledAt:    scheduledAt,
		Status:         models.MatchDraft,
		OverrideReason: req.OverrideReason,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.matchRepo.Create(ctx, match); err != nil {
		return nil, err
	}

	auditDetails := map[string]interface{}{
		"round":        match.Round,
		"home_team_id": match.HomeTeamID,
		"away_team_id": match.AwayTeamID,
		"venue_id":     match.VenueID,
		"scheduled_at": match.ScheduledAt,
	}
	if vr.hasViolations() {
		auditDetails["violations_overridden"] = vr.violations
		auditDetails["override_reason"] = *req.OverrideReason
	}
	s.audit.Log(ctx, "match", match.ID, actorID, "created", auditDetails)

	return match, nil
}

func (s *MatchService) UpdateMatch(ctx context.Context, matchID uuid.UUID, req *dto.UpdateMatchRequest, actorID uuid.UUID) (*models.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, ErrMatchNotFound
	}
	if match.Status != models.MatchDraft {
		return nil, ErrMatchNotDraft
	}

	// Apply updates
	if req.Round != nil {
		match.Round = *req.Round
	}
	if req.HomeTeamID != nil {
		id, err := uuid.Parse(*req.HomeTeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid home_team_id")
		}
		match.HomeTeamID = id
	}
	if req.AwayTeamID != nil {
		id, err := uuid.Parse(*req.AwayTeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid away_team_id")
		}
		match.AwayTeamID = id
	}
	if match.HomeTeamID == match.AwayTeamID {
		return nil, fmt.Errorf("home and away teams must be different")
	}
	if req.VenueID != nil {
		id, err := uuid.Parse(*req.VenueID)
		if err != nil {
			return nil, fmt.Errorf("invalid venue_id")
		}
		match.VenueID = id
	}
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled_at, use RFC3339 format")
		}
		match.ScheduledAt = t
	}

	// Run scheduling validations
	vr, err := s.validateScheduling(ctx, match.SeasonID, match.Round, match.HomeTeamID, match.AwayTeamID, match.VenueID, match.ScheduledAt, &matchID)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	overrideReason := req.OverrideReason
	if vr.hasViolations() {
		if overrideReason == nil || *overrideReason == "" {
			return nil, fmt.Errorf("%w: %s", ErrOverrideRequired, vr.summary())
		}
	}
	match.OverrideReason = overrideReason

	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}

	auditDetails := map[string]interface{}{
		"round":        match.Round,
		"home_team_id": match.HomeTeamID,
		"away_team_id": match.AwayTeamID,
		"venue_id":     match.VenueID,
		"scheduled_at": match.ScheduledAt,
	}
	if vr.hasViolations() {
		auditDetails["violations_overridden"] = vr.violations
		auditDetails["override_reason"] = *overrideReason
	}
	s.audit.Log(ctx, "match", match.ID, actorID, "updated", auditDetails)

	return match, nil
}

func (s *MatchService) GetMatch(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrMatchNotFound
	}
	return match, nil
}

func (s *MatchService) ListMatches(ctx context.Context, seasonID uuid.UUID, offset, limit int) ([]models.Match, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.matchRepo.ListBySeason(ctx, seasonID, offset, limit)
}

func (s *MatchService) ListMatchesByRound(ctx context.Context, seasonID uuid.UUID, round int) ([]models.Match, error) {
	return s.matchRepo.ListByRound(ctx, seasonID, round)
}

func (s *MatchService) TransitionStatus(ctx context.Context, matchID uuid.UUID, req *dto.TransitionMatchRequest, actorID uuid.UUID) (*models.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, ErrMatchNotFound
	}

	newStatus := models.MatchStatus(req.Status)
	if !models.ValidMatchStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	if !models.CanTransition(match.Status, newStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, match.Status, newStatus)
	}

	if err := s.matchRepo.UpdateStatus(ctx, matchID, newStatus, req.OverrideReason); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "match", matchID, actorID, "status_transition", map[string]interface{}{
		"from":            string(match.Status),
		"to":              req.Status,
		"override_reason": req.OverrideReason,
	})

	match.Status = newStatus
	if req.OverrideReason != nil {
		match.OverrideReason = req.OverrideReason
	}
	return match, nil
}

func (s *MatchService) ImportSchedule(ctx context.Context, req *dto.ImportScheduleRequest, actorID uuid.UUID) (*dto.ImportScheduleResponse, error) {
	seasonID, err := uuid.Parse(req.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("invalid season_id")
	}
	if _, err := s.seasonRepo.GetByID(ctx, seasonID); err != nil {
		return nil, fmt.Errorf("season not found")
	}

	resp := &dto.ImportScheduleResponse{}

	for i, entry := range req.Matches {
		createReq := &dto.CreateMatchRequest{
			SeasonID:       req.SeasonID,
			Round:          entry.Round,
			HomeTeamID:     entry.HomeTeamID,
			AwayTeamID:     entry.AwayTeamID,
			VenueID:        entry.VenueID,
			ScheduledAt:    entry.ScheduledAt,
			OverrideReason: entry.OverrideReason,
		}

		match, err := s.CreateMatch(ctx, createReq, actorID)
		if err != nil {
			resp.Errors = append(resp.Errors, dto.ImportError{
				Index:   i,
				Message: err.Error(),
			})
			continue
		}
		resp.Created++
		resp.Matches = append(resp.Matches, dto.ToMatchResponse(match))
	}

	return resp, nil
}

// GenerateSchedule creates a round-robin schedule for all teams in a season.
// It assigns venues round-robin and spaces matches by IntervalDays.
// All matches go through the same validation as CreateMatch.
func (s *MatchService) GenerateSchedule(ctx context.Context, req *dto.GenerateScheduleRequest, actorID uuid.UUID) (*dto.GenerateScheduleResponse, error) {
	seasonID, err := uuid.Parse(req.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("invalid season_id")
	}
	if _, err := s.seasonRepo.GetByID(ctx, seasonID); err != nil {
		return nil, fmt.Errorf("season not found")
	}

	if len(req.VenueIDs) == 0 {
		return nil, fmt.Errorf("at least one venue_id is required")
	}
	venueIDs := make([]uuid.UUID, len(req.VenueIDs))
	for i, v := range req.VenueIDs {
		vid, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid venue_id at index %d", i)
		}
		if _, err := s.venueRepo.GetByID(ctx, vid); err != nil {
			return nil, fmt.Errorf("venue not found: %s", v)
		}
		venueIDs[i] = vid
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date, use YYYY-MM-DD")
	}

	startTime := "14:00"
	if req.StartTime != "" {
		startTime = req.StartTime
	}
	parsedStartTime, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time, use HH:MM")
	}
	intervalDays := req.IntervalDays
	if intervalDays < 1 {
		intervalDays = 7
	}

	// Get all teams in the season
	teams, err := s.teamRepo.ListBySeason(ctx, seasonID)
	if err != nil {
		return nil, fmt.Errorf("listing teams: %w", err)
	}
	if len(teams) < 2 {
		return nil, fmt.Errorf("need at least 2 teams to generate a schedule")
	}

	// Generate round-robin pairings
	n := len(teams)
	// If odd number of teams, add a "bye" (nil)
	teamIDs := make([]uuid.UUID, n)
	for i, t := range teams {
		teamIDs[i] = t.ID
	}

	hasBye := false
	if n%2 != 0 {
		hasBye = true
		teamIDs = append(teamIDs, uuid.Nil) // bye marker
		n++
	}

	numRounds := n - 1
	resp := &dto.GenerateScheduleResponse{Rounds: numRounds}

	matchDate := startDate
	venueIdx := 0

	for round := 1; round <= numRounds; round++ {
		for i := 0; i < n/2; i++ {
			home := teamIDs[i]
			away := teamIDs[n-1-i]

			// Skip bye pairings
			if hasBye && (home == uuid.Nil || away == uuid.Nil) {
				continue
			}

			venueID := venueIDs[venueIdx%len(venueIDs)]
			venueIdx++

			// Stagger matches by slot to avoid venue overlap conflicts when using
			// a single venue (conflict rule blocks overlaps within 90 minutes).
			slotTime := time.Date(
				matchDate.Year(),
				matchDate.Month(),
				matchDate.Day(),
				parsedStartTime.Hour(),
				parsedStartTime.Minute(),
				0,
				0,
				time.UTC,
			).Add(time.Duration(i) * 90 * time.Minute)

			scheduledAt := slotTime.Format(time.RFC3339)

			createReq := &dto.CreateMatchRequest{
				SeasonID:    req.SeasonID,
				Round:       round,
				HomeTeamID:  home.String(),
				AwayTeamID:  away.String(),
				VenueID:     venueID.String(),
				ScheduledAt: scheduledAt,
			}

			match, err := s.CreateMatch(ctx, createReq, actorID)
			if err != nil {
				resp.Errors = append(resp.Errors, dto.ImportError{
					Index:   resp.Created + len(resp.Errors),
					Message: fmt.Sprintf("round %d: %s vs %s: %s", round, home, away, err.Error()),
				})
				continue
			}
			resp.Created++
			resp.Matches = append(resp.Matches, dto.ToMatchResponse(match))
		}

		// Rotate teams for next round (keep first team fixed, rotate the rest)
		last := teamIDs[n-1]
		copy(teamIDs[2:], teamIDs[1:n-1])
		teamIDs[1] = last

		matchDate = matchDate.Add(time.Duration(intervalDays) * 24 * time.Hour)
	}

	s.audit.Log(ctx, "season", seasonID, actorID, "schedule_generated", map[string]interface{}{
		"rounds":  numRounds,
		"created": resp.Created,
		"errors":  len(resp.Errors),
	})

	return resp, nil
}

// Assignments

func (s *MatchService) CreateAssignment(ctx context.Context, req *dto.CreateAssignmentRequest, assignedBy uuid.UUID) (*models.MatchAssignment, error) {
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id")
	}
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account_id")
	}

	role := models.AssignmentRole(req.Role)
	if !models.ValidAssignmentRoles[role] {
		return nil, fmt.Errorf("invalid assignment role: %s", req.Role)
	}

	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, ErrMatchNotFound
	}

	// Assignments locked once In-Progress or beyond
	if match.Status == models.MatchInProgress || match.Status == models.MatchFinal || match.Status == models.MatchCanceled {
		return nil, ErrAssignmentLocked
	}

	now := time.Now()
	assignment := &models.MatchAssignment{
		ID:         uuid.New(),
		MatchID:    matchID,
		AccountID:  accountID,
		Role:       role,
		AssignedBy: assignedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "match_assignment", assignment.ID, assignedBy, "created", map[string]interface{}{
		"match_id":   matchID,
		"account_id": accountID,
		"role":       string(role),
	})

	return assignment, nil
}

func (s *MatchService) ReassignAssignment(ctx context.Context, assignmentID uuid.UUID, req *dto.ReassignRequest, reassignedBy uuid.UUID) (*models.MatchAssignment, error) {
	if req.Reason == "" {
		return nil, ErrReassignmentReason
	}

	newAccountID, err := uuid.Parse(req.NewAccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid new_account_id")
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found")
	}

	match, err := s.matchRepo.GetByID(ctx, assignment.MatchID)
	if err != nil {
		return nil, ErrMatchNotFound
	}

	// Assignments locked once In-Progress or beyond
	if match.Status == models.MatchInProgress || match.Status == models.MatchFinal || match.Status == models.MatchCanceled {
		return nil, ErrAssignmentLocked
	}

	oldAccountID := assignment.AccountID
	assignment.AccountID = newAccountID
	assignment.AssignedBy = reassignedBy
	assignment.ReassignmentReason = &req.Reason

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "match_assignment", assignmentID, reassignedBy, "reassigned", map[string]interface{}{
		"match_id":       assignment.MatchID,
		"old_account_id": oldAccountID,
		"new_account_id": newAccountID,
		"reason":         req.Reason,
	})

	return assignment, nil
}

func (s *MatchService) ListAssignments(ctx context.Context, matchID uuid.UUID) ([]models.MatchAssignment, error) {
	return s.assignmentRepo.ListByMatch(ctx, matchID)
}

func (s *MatchService) DeleteAssignment(ctx context.Context, assignmentID, actorID uuid.UUID) error {
	assignment, err := s.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("assignment not found")
	}

	match, err := s.matchRepo.GetByID(ctx, assignment.MatchID)
	if err != nil {
		return ErrMatchNotFound
	}

	if match.Status == models.MatchInProgress || match.Status == models.MatchFinal || match.Status == models.MatchCanceled {
		return ErrAssignmentLocked
	}

	if err := s.assignmentRepo.Delete(ctx, assignmentID); err != nil {
		return err
	}

	s.audit.Log(ctx, "match_assignment", assignmentID, actorID, "deleted", map[string]interface{}{
		"match_id":   assignment.MatchID,
		"account_id": assignment.AccountID,
		"role":       string(assignment.Role),
	})

	return nil
}

// validateScheduling checks home/away balance, duplicate pairings, and venue conflicts
func (s *MatchService) validateScheduling(
	ctx context.Context,
	seasonID uuid.UUID,
	round int,
	homeTeamID, awayTeamID, venueID uuid.UUID,
	scheduledAt time.Time,
	excludeMatchID *uuid.UUID,
) (*validationResult, error) {
	vr := &validationResult{}

	// 1. Check venue time conflict (90-minute window)
	hasConflict, err := s.matchRepo.CheckVenueConflict(ctx, venueID, scheduledAt, excludeMatchID)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		vr.violations = append(vr.violations, ErrVenueConflict.Error())
	}

	// 2. Check duplicate pairings within the round
	hasDuplicate, err := s.matchRepo.CheckDuplicatePairing(ctx, seasonID, round, homeTeamID, awayTeamID, excludeMatchID)
	if err != nil {
		return nil, err
	}
	if hasDuplicate {
		vr.violations = append(vr.violations, ErrDuplicatePairing.Error())
	}

	// 3. Check consecutive home/away balance (max 3)
	if err := s.checkConsecutiveBalance(ctx, seasonID, homeTeamID, awayTeamID, scheduledAt, excludeMatchID, vr); err != nil {
		return nil, err
	}

	return vr, nil
}

func (s *MatchService) checkConsecutiveBalance(
	ctx context.Context,
	seasonID, homeTeamID, awayTeamID uuid.UUID,
	scheduledAt time.Time,
	excludeMatchID *uuid.UUID,
	vr *validationResult,
) error {
	// Check home team's consecutive home games
	recentHomeMatches, err := s.matchRepo.GetTeamRecentMatches(ctx, seasonID, homeTeamID, scheduledAt, maxConsecutiveHomeAway)
	if err != nil {
		return err
	}
	consecutiveHome := 0
	for _, m := range recentHomeMatches {
		if excludeMatchID != nil && m.ID == *excludeMatchID {
			continue
		}
		if m.HomeTeamID == homeTeamID {
			consecutiveHome++
		} else {
			break
		}
	}
	if consecutiveHome >= maxConsecutiveHomeAway {
		vr.violations = append(vr.violations, fmt.Sprintf("home team: %s", ErrConsecutiveHome.Error()))
	}

	// Check away team's consecutive away games
	recentAwayMatches, err := s.matchRepo.GetTeamRecentMatches(ctx, seasonID, awayTeamID, scheduledAt, maxConsecutiveHomeAway)
	if err != nil {
		return err
	}
	consecutiveAway := 0
	for _, m := range recentAwayMatches {
		if excludeMatchID != nil && m.ID == *excludeMatchID {
			continue
		}
		if m.AwayTeamID == awayTeamID {
			consecutiveAway++
		} else {
			break
		}
	}
	if consecutiveAway >= maxConsecutiveHomeAway {
		vr.violations = append(vr.violations, fmt.Sprintf("away team: %s", ErrConsecutiveAway.Error()))
	}

	return nil
}
