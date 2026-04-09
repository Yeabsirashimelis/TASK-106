package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

type SeasonService struct {
	seasonRepo *repository.SeasonRepository
	teamRepo   *repository.TeamRepository
	venueRepo  *repository.VenueRepository
	audit      *AuditService
}

func NewSeasonService(
	seasonRepo *repository.SeasonRepository,
	teamRepo *repository.TeamRepository,
	venueRepo *repository.VenueRepository,
	audit *AuditService,
) *SeasonService {
	return &SeasonService{
		seasonRepo: seasonRepo,
		teamRepo:   teamRepo,
		venueRepo:  venueRepo,
		audit:      audit,
	}
}

func (s *SeasonService) CreateSeason(ctx context.Context, req *dto.CreateSeasonRequest, actorID uuid.UUID) (*models.Season, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, use YYYY-MM-DD")
	}
	if !endDate.After(startDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	now := time.Now()
	season := &models.Season{
		ID:        uuid.New(),
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    models.SeasonPlanning,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.seasonRepo.Create(ctx, season); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "season", season.ID, actorID, "created", map[string]interface{}{
		"name":       season.Name,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	})

	return season, nil
}

func (s *SeasonService) GetSeason(ctx context.Context, id uuid.UUID) (*models.Season, error) {
	return s.seasonRepo.GetByID(ctx, id)
}

func (s *SeasonService) ListSeasons(ctx context.Context, offset, limit int) ([]models.Season, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.seasonRepo.List(ctx, offset, limit)
}

func (s *SeasonService) CreateTeam(ctx context.Context, req *dto.CreateTeamRequest, actorID uuid.UUID) (*models.Team, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	seasonID, err := uuid.Parse(req.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("invalid season_id")
	}

	// Verify season exists
	if _, err := s.seasonRepo.GetByID(ctx, seasonID); err != nil {
		return nil, fmt.Errorf("season not found")
	}

	team := &models.Team{
		ID:        uuid.New(),
		Name:      req.Name,
		SeasonID:  seasonID,
		CreatedAt: time.Now(),
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "team", team.ID, actorID, "created", map[string]interface{}{
		"name":      team.Name,
		"season_id": team.SeasonID,
	})

	return team, nil
}

func (s *SeasonService) ListTeams(ctx context.Context, seasonID uuid.UUID) ([]models.Team, error) {
	return s.teamRepo.ListBySeason(ctx, seasonID)
}

func (s *SeasonService) CreateVenue(ctx context.Context, req *dto.CreateVenueRequest, actorID uuid.UUID) (*models.Venue, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	venue := &models.Venue{
		ID:        uuid.New(),
		Name:      req.Name,
		Location:  req.Location,
		Capacity:  req.Capacity,
		CreatedAt: time.Now(),
	}

	if err := s.venueRepo.Create(ctx, venue); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "venue", venue.ID, actorID, "created", map[string]interface{}{
		"name": venue.Name,
	})

	return venue, nil
}

func (s *SeasonService) ListVenues(ctx context.Context, offset, limit int) ([]models.Venue, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.venueRepo.List(ctx, offset, limit)
}
