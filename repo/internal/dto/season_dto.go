package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type CreateSeasonRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SeasonResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToSeasonResponse(s *models.Season) SeasonResponse {
	return SeasonResponse{
		ID:        s.ID,
		Name:      s.Name,
		StartDate: s.StartDate.Format("2006-01-02"),
		EndDate:   s.EndDate.Format("2006-01-02"),
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToSeasonResponseList(seasons []models.Season) []SeasonResponse {
	result := make([]SeasonResponse, len(seasons))
	for i, s := range seasons {
		result[i] = ToSeasonResponse(&s)
	}
	return result
}

type CreateTeamRequest struct {
	Name     string `json:"name"`
	SeasonID string `json:"season_id"`
}

type TeamResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SeasonID  uuid.UUID `json:"season_id"`
	CreatedAt time.Time `json:"created_at"`
}

func ToTeamResponse(t *models.Team) TeamResponse {
	return TeamResponse{
		ID:        t.ID,
		Name:      t.Name,
		SeasonID:  t.SeasonID,
		CreatedAt: t.CreatedAt,
	}
}

func ToTeamResponseList(teams []models.Team) []TeamResponse {
	result := make([]TeamResponse, len(teams))
	for i, t := range teams {
		result[i] = ToTeamResponse(&t)
	}
	return result
}

type CreateVenueRequest struct {
	Name     string  `json:"name"`
	Location *string `json:"location,omitempty"`
	Capacity *int    `json:"capacity,omitempty"`
}

type VenueResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Location  *string   `json:"location,omitempty"`
	Capacity  *int      `json:"capacity,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func ToVenueResponse(v *models.Venue) VenueResponse {
	return VenueResponse{
		ID:        v.ID,
		Name:      v.Name,
		Location:  v.Location,
		Capacity:  v.Capacity,
		CreatedAt: v.CreatedAt,
	}
}

func ToVenueResponseList(venues []models.Venue) []VenueResponse {
	result := make([]VenueResponse, len(venues))
	for i, v := range venues {
		result[i] = ToVenueResponse(&v)
	}
	return result
}
