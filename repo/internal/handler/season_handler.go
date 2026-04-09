package handler

import (
	"net/http"
	"strconv"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SeasonHandler struct {
	seasonService *service.SeasonService
}

func NewSeasonHandler(seasonService *service.SeasonService) *SeasonHandler {
	return &SeasonHandler{seasonService: seasonService}
}

func (h *SeasonHandler) CreateSeason(c echo.Context) error {
	var req dto.CreateSeasonRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	season, err := h.seasonService.CreateSeason(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToSeasonResponse(season))
}

func (h *SeasonHandler) GetSeason(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid season id")
	}

	season, err := h.seasonService.GetSeason(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "season not found")
	}

	return c.JSON(http.StatusOK, dto.ToSeasonResponse(season))
}

func (h *SeasonHandler) ListSeasons(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	seasons, err := h.seasonService.ListSeasons(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToSeasonResponseList(seasons))
}

func (h *SeasonHandler) CreateTeam(c echo.Context) error {
	var req dto.CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	team, err := h.seasonService.CreateTeam(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToTeamResponse(team))
}

func (h *SeasonHandler) ListTeams(c echo.Context) error {
	seasonID, err := uuid.Parse(c.Param("season_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid season_id")
	}

	teams, err := h.seasonService.ListTeams(c.Request().Context(), seasonID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToTeamResponseList(teams))
}

func (h *SeasonHandler) CreateVenue(c echo.Context) error {
	var req dto.CreateVenueRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	venue, err := h.seasonService.CreateVenue(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToVenueResponse(venue))
}

func (h *SeasonHandler) ListVenues(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	venues, err := h.seasonService.ListVenues(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToVenueResponseList(venues))
}
