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

type ModerationHandler struct {
	moderationService *service.ModerationService
}

func NewModerationHandler(moderationService *service.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

// Dictionary endpoints

func (h *ModerationHandler) CreateDictionary(c echo.Context) error {
	var req dto.CreateDictionaryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	d, err := h.moderationService.CreateDictionary(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToDictionaryResponse(d))
}

func (h *ModerationHandler) GetDictionary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	d, err := h.moderationService.GetDictionary(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToDictionaryResponse(d))
}

func (h *ModerationHandler) ListDictionaries(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	dicts, err := h.moderationService.ListDictionaries(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToDictionaryResponseList(dicts))
}

func (h *ModerationHandler) UpdateDictionary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	var req dto.UpdateDictionaryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	d, err := h.moderationService.UpdateDictionary(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrDictionaryNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToDictionaryResponse(d))
}

func (h *ModerationHandler) DeleteDictionary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	if err := h.moderationService.DeleteDictionary(c.Request().Context(), id, actorID); err != nil {
		if errors.Is(err, service.ErrDictionaryNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Word endpoints

func (h *ModerationHandler) AddWord(c echo.Context) error {
	dictID, err := uuid.Parse(c.Param("dict_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	var req dto.AddWordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	w, err := h.moderationService.AddWord(c.Request().Context(), dictID, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrDictionaryNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToWordResponse(w))
}

func (h *ModerationHandler) AddWords(c echo.Context) error {
	dictID, err := uuid.Parse(c.Param("dict_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	var req dto.AddWordsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	words, err := h.moderationService.AddWords(c.Request().Context(), dictID, &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToWordResponseList(words))
}

func (h *ModerationHandler) ListWords(c echo.Context) error {
	dictID, err := uuid.Parse(c.Param("dict_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dictionary id")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	words, err := h.moderationService.ListWords(c.Request().Context(), dictID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToWordResponseList(words))
}

func (h *ModerationHandler) DeleteWord(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid word id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	if err := h.moderationService.DeleteWord(c.Request().Context(), id, actorID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Content check endpoint

func (h *ModerationHandler) CheckContent(c echo.Context) error {
	var req dto.CheckContentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "text is required")
	}

	result, err := h.moderationService.CheckContent(c.Request().Context(), req.Text)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, result)
}

// Review endpoints

func (h *ModerationHandler) CreateReview(c echo.Context) error {
	var req dto.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	review, err := h.moderationService.CreateReview(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToReviewResponse(review))
}

func (h *ModerationHandler) GetReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid review id")
	}

	review, err := h.moderationService.GetReview(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewResponse(review))
}

func (h *ModerationHandler) ListReviews(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	var statusFilter *string
	if s := c.QueryParam("status"); s != "" {
		statusFilter = &s
	}

	reviews, err := h.moderationService.ListReviews(c.Request().Context(), statusFilter, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReviewResponseList(reviews))
}

func (h *ModerationHandler) DecideReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid review id")
	}

	var req dto.DecideReviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	review, err := h.moderationService.DecideReview(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrReviewReasonReq) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrInvalidReviewTrans) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReviewResponse(review))
}
