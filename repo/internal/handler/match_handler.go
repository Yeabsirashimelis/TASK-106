package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MatchHandler struct {
	matchService *service.MatchService
}

func NewMatchHandler(matchService *service.MatchService) *MatchHandler {
	return &MatchHandler{matchService: matchService}
}

func (h *MatchHandler) CreateMatch(c echo.Context) error {
	var req dto.CreateMatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	match, err := h.matchService.CreateMatch(c.Request().Context(), &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrOverrideRequired) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToMatchResponse(match))
}

func (h *MatchHandler) UpdateMatch(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid match id")
	}

	var req dto.UpdateMatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	match, err := h.matchService.UpdateMatch(c.Request().Context(), id, &req, actorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMatchNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrMatchNotDraft):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrOverrideRequired):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		default:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return c.JSON(http.StatusOK, dto.ToMatchResponse(match))
}

func (h *MatchHandler) GetMatch(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid match id")
	}

	match, err := h.matchService.GetMatch(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "match not found")
	}

	return c.JSON(http.StatusOK, dto.ToMatchResponse(match))
}

func (h *MatchHandler) ListMatches(c echo.Context) error {
	seasonID, err := uuid.Parse(c.QueryParam("season_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "season_id query parameter is required")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Filter by round if specified
	if roundStr := c.QueryParam("round"); roundStr != "" {
		round, err := strconv.Atoi(roundStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid round parameter")
		}
		matches, err := h.matchService.ListMatchesByRound(c.Request().Context(), seasonID, round)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
		return c.JSON(http.StatusOK, dto.ToMatchResponseList(matches))
	}

	matches, err := h.matchService.ListMatches(c.Request().Context(), seasonID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToMatchResponseList(matches))
}

func (h *MatchHandler) TransitionStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid match id")
	}

	var req dto.TransitionMatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Status == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	match, err := h.matchService.TransitionStatus(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrMatchNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToMatchResponse(match))
}

func (h *MatchHandler) ImportSchedule(c echo.Context) error {
	var req dto.ImportScheduleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.Matches) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "matches array is required")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	resp, err := h.matchService.ImportSchedule(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *MatchHandler) GenerateSchedule(c echo.Context) error {
	var req dto.GenerateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	resp, err := h.matchService.GenerateSchedule(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

// Assignment handlers

func (h *MatchHandler) CreateAssignment(c echo.Context) error {
	var req dto.CreateAssignmentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	assignedBy, _ := c.Get("account_id").(uuid.UUID)
	assignment, err := h.matchService.CreateAssignment(c.Request().Context(), &req, assignedBy)
	if err != nil {
		if errors.Is(err, service.ErrAssignmentLocked) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if errors.Is(err, service.ErrMatchNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToAssignmentResponse(assignment))
}

func (h *MatchHandler) ReassignAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid assignment id")
	}

	var req dto.ReassignRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	reassignedBy, _ := c.Get("account_id").(uuid.UUID)
	assignment, err := h.matchService.ReassignAssignment(c.Request().Context(), id, &req, reassignedBy)
	if err != nil {
		if errors.Is(err, service.ErrAssignmentLocked) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if errors.Is(err, service.ErrReassignmentReason) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToAssignmentResponse(assignment))
}

func (h *MatchHandler) ListAssignments(c echo.Context) error {
	matchID, err := uuid.Parse(c.Param("match_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid match_id")
	}

	assignments, err := h.matchService.ListAssignments(c.Request().Context(), matchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToAssignmentResponseList(assignments))
}

func (h *MatchHandler) DeleteAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid assignment id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	if err := h.matchService.DeleteAssignment(c.Request().Context(), id, actorID); err != nil {
		if errors.Is(err, service.ErrAssignmentLocked) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
