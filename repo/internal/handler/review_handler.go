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

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// Config endpoints

func (h *ReviewHandler) CreateConfig(c echo.Context) error {
	var req dto.CreateReviewConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	cfg, err := h.reviewService.CreateConfig(c.Request().Context(), &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLevels) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToReviewConfigResponse(cfg))
}

func (h *ReviewHandler) GetConfig(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	}

	cfg, err := h.reviewService.GetConfig(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewConfigResponse(cfg))
}

func (h *ReviewHandler) ListConfigs(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	configs, err := h.reviewService.ListConfigs(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewConfigResponseList(configs))
}

func (h *ReviewHandler) UpdateConfig(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	}

	var req dto.UpdateReviewConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	cfg, err := h.reviewService.UpdateConfig(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReviewConfigNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewConfigResponse(cfg))
}

func (h *ReviewHandler) DeleteConfig(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	if err := h.reviewService.DeleteConfig(c.Request().Context(), id, actorID); err != nil {
		if errors.Is(err, service.ErrReviewConfigNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Request endpoints

func (h *ReviewHandler) SubmitRequest(c echo.Context) error {
	var req dto.SubmitReviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	submitterID, _ := c.Get("account_id").(uuid.UUID)
	request, err := h.reviewService.SubmitRequest(c.Request().Context(), &req, submitterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	levels, _ := h.reviewService.ListLevels(c.Request().Context(), request.ID)
	return c.JSON(http.StatusCreated, dto.ToReviewRequestResponse(request, levels))
}

func (h *ReviewHandler) GetRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	request, levels, err := h.reviewService.GetRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewRequestResponse(request, levels))
}

func (h *ReviewHandler) ListRequests(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	var statusFilter *string
	if s := c.QueryParam("status"); s != "" {
		statusFilter = &s
	}

	requests, err := h.reviewService.ListRequests(c.Request().Context(), statusFilter, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewRequestResponseList(requests))
}

func (h *ReviewHandler) ListMyAssignments(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	assigneeID, _ := c.Get("account_id").(uuid.UUID)
	requests, err := h.reviewService.ListMyAssignments(c.Request().Context(), assigneeID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewRequestResponseList(requests))
}

func (h *ReviewHandler) ListByEntity(c echo.Context) error {
	entityType := c.QueryParam("entity_type")
	entityIDStr := c.QueryParam("entity_id")
	if entityType == "" || entityIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "entity_type and entity_id are required")
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid entity_id")
	}

	requests, err := h.reviewService.ListByEntity(c.Request().Context(), entityType, entityID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewRequestResponseList(requests))
}

func (h *ReviewHandler) ListFollowUpRequests(c echo.Context) error {
	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	requests, err := h.reviewService.ListFollowUpRequests(c.Request().Context(), parentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewRequestResponseList(requests))
}

func (h *ReviewHandler) ResubmitRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	request, err := h.reviewService.ResubmitAfterReturn(c.Request().Context(), id, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReviewRequestNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	levels, _ := h.reviewService.ListLevels(c.Request().Context(), request.ID)
	return c.JSON(http.StatusOK, dto.ToReviewRequestResponse(request, levels))
}

// Level endpoints

func (h *ReviewHandler) AssignLevel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid level id")
	}

	var req dto.AssignLevelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	level, err := h.reviewService.AssignLevel(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReviewLevelNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewLevelResponse(level))
}

func (h *ReviewHandler) DecideLevel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid level id")
	}

	var req dto.DecideLevelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	deciderID, _ := c.Get("account_id").(uuid.UUID)
	request, _, err := h.reviewService.DecideLevel(c.Request().Context(), id, &req, deciderID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReviewLevelNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrLevelNotPending):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrRequestNotInReview):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrNotCurrentLevel):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrNotAssignee):
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrDecisionRequired):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	levels, _ := h.reviewService.ListLevels(c.Request().Context(), request.ID)
	return c.JSON(http.StatusOK, dto.ToReviewRequestResponse(request, levels))
}

func (h *ReviewHandler) ListLevels(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("request_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	levels, err := h.reviewService.ListLevels(c.Request().Context(), requestID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewLevelResponseList(levels))
}

// Follow-up endpoints

func (h *ReviewHandler) AddFollowUp(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	var req dto.CreateFollowUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	authorID, _ := c.Get("account_id").(uuid.UUID)
	fu, err := h.reviewService.AddFollowUp(c.Request().Context(), requestID, &req, authorID)
	if err != nil {
		if errors.Is(err, service.ErrReviewRequestNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToFollowUpResponse(fu))
}

func (h *ReviewHandler) ListFollowUps(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	followUps, err := h.reviewService.ListFollowUps(c.Request().Context(), requestID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToFollowUpResponseList(followUps))
}
