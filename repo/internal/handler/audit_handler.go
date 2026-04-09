package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuditHandler struct {
	auditService        *service.AuditService
	observabilityService *service.ObservabilityService
}

func NewAuditHandler(auditService *service.AuditService, observabilityService *service.ObservabilityService) *AuditHandler {
	return &AuditHandler{
		auditService:        auditService,
		observabilityService: observabilityService,
	}
}

func (h *AuditHandler) GetAuditLog(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid audit log id")
	}

	log, err := h.auditService.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "audit log not found")
	}

	return c.JSON(http.StatusOK, dto.ToAuditLogResponse(log))
}

func (h *AuditHandler) ListByEntity(c echo.Context) error {
	entityType := c.QueryParam("entity_type")
	entityIDStr := c.QueryParam("entity_id")
	if entityType == "" || entityIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "entity_type and entity_id are required")
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid entity_id")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	logs, err := h.auditService.ListByEntity(c.Request().Context(), entityType, entityID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToAuditLogResponseList(logs))
}

func (h *AuditHandler) ListByActor(c echo.Context) error {
	actorID, err := uuid.Parse(c.Param("actor_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid actor_id")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	logs, err := h.auditService.ListByActor(c.Request().Context(), actorID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToAuditLogResponseList(logs))
}

func (h *AuditHandler) QueryLogs(c echo.Context) error {
	params := repository.AuditQueryParams{}

	if v := c.QueryParam("actor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid actor_id")
		}
		params.ActorID = &id
	}
	if v := c.QueryParam("entity_type"); v != "" {
		params.EntityType = &v
	}
	if v := c.QueryParam("action"); v != "" {
		params.Action = &v
	}
	if v := c.QueryParam("tier"); v != "" {
		tier := models.LogTier(v)
		params.Tier = &tier
	}
	if v := c.QueryParam("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start_time, use RFC3339")
		}
		params.StartTime = &t
	}
	if v := c.QueryParam("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_time, use RFC3339")
		}
		params.EndTime = &t
	}

	params.Offset, _ = strconv.Atoi(c.QueryParam("offset"))
	params.Limit, _ = strconv.Atoi(c.QueryParam("limit"))

	logs, err := h.auditService.Query(c.Request().Context(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToAuditLogResponseList(logs))
}

func (h *AuditHandler) ExportCSV(c echo.Context) error {
	params := repository.AuditQueryParams{}

	if v := c.QueryParam("actor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid actor_id")
		}
		params.ActorID = &id
	}
	if v := c.QueryParam("entity_type"); v != "" {
		params.EntityType = &v
	}
	if v := c.QueryParam("action"); v != "" {
		params.Action = &v
	}
	if v := c.QueryParam("tier"); v != "" {
		tier := models.LogTier(v)
		params.Tier = &tier
	}
	if v := c.QueryParam("start_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start_time")
		}
		params.StartTime = &t
	}
	if v := c.QueryParam("end_time"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_time")
		}
		params.EndTime = &t
	}

	csvPath, err := h.auditService.ExportCSV(c.Request().Context(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	f, err := os.Open(csvPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "CSV file not found")
	}
	defer f.Close()

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="audit_export.csv"`))
	return c.Stream(http.StatusOK, "text/csv", f)
}

func (h *AuditHandler) CountByTier(c echo.Context) error {
	counts, err := h.auditService.CountByTier(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	resp := dto.TierCountResponse{
		Access:    counts[models.TierAccess],
		Operation: counts[models.TierOperation],
		Audit:     counts[models.TierAudit],
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *AuditHandler) PurgeExpired(c echo.Context) error {
	count, err := h.auditService.PurgeExpired(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.PurgeResponse{DeletedCount: count})
}

// Hash chain endpoints

func (h *AuditHandler) BuildHashChain(c echo.Context) error {
	var req dto.BuildChainRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Date == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "date is required (YYYY-MM-DD)")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date format")
	}

	chain, err := h.auditService.BuildDailyHashChain(c.Request().Context(), date)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToHashChainResponse(chain))
}

func (h *AuditHandler) VerifyHashChain(c echo.Context) error {
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "date query parameter is required")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date format")
	}

	valid, message, err := h.auditService.VerifyHashChain(c.Request().Context(), date)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.VerifyChainResponse{Valid: valid, Message: message})
}

func (h *AuditHandler) ListHashChain(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	chains, err := h.auditService.ListHashChain(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToHashChainResponseList(chains))
}

// Health and metrics (these are on public endpoints, handled in router)

func (h *AuditHandler) DetailedHealth(c echo.Context) error {
	status := h.observabilityService.HealthCheck(c.Request().Context())
	code := http.StatusOK
	if status.Status != "ok" {
		code = http.StatusServiceUnavailable
	}
	return c.JSON(code, status)
}

func (h *AuditHandler) GetMetrics(c echo.Context) error {
	return c.JSON(http.StatusOK, h.observabilityService.GetMetrics())
}
