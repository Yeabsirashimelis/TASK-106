package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) CreateReport(c echo.Context) error {
	var req dto.CreateReportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	reporterID, _ := c.Get("account_id").(uuid.UUID)
	report, err := h.reportService.CreateReport(c.Request().Context(), &req, reporterID)
	if err != nil {
		if errors.Is(err, service.ErrReportLimitExc) {
			return echo.NewHTTPError(http.StatusTooManyRequests, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToReportResponse(report))
}

func (h *ReportHandler) GetReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	report, err := h.reportService.GetReport(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReportResponse(report))
}

func (h *ReportHandler) ListReports(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	var statusFilter *string
	if s := c.QueryParam("status"); s != "" {
		statusFilter = &s
	}

	reports, err := h.reportService.ListReports(c.Request().Context(), statusFilter, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReportResponseList(reports))
}

func (h *ReportHandler) UpdateReportStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	var req dto.UpdateReportStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	report, err := h.reportService.UpdateReportStatus(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrInvalidReportTrans) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReportResponse(report))
}

func (h *ReportHandler) AssignReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	var req dto.AssignReportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	report, err := h.reportService.AssignReport(c.Request().Context(), id, &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReportResponse(report))
}

// Evidence

func (h *ReportHandler) UploadEvidence(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read uploaded file")
	}
	defer src.Close()

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	evidence, err := h.reportService.UploadEvidence(c.Request().Context(), reportID, file.Filename, mimeType, file.Size, src, actorID)
	if err != nil {
		if errors.Is(err, service.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToEvidenceResponse(evidence))
}

func (h *ReportHandler) ListEvidence(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	evidence, err := h.reportService.ListEvidence(c.Request().Context(), reportID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToEvidenceResponseList(evidence))
}

func (h *ReportHandler) DownloadEvidence(c echo.Context) error {
	id, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid evidence id")
	}

	evidence, err := h.reportService.GetEvidenceFile(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	f, err := os.Open(evidence.StoragePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "file not found on disk")
	}
	defer f.Close()

	c.Response().Header().Set("Content-Type", evidence.MimeType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, evidence.FileName))
	c.Response().Header().Set("X-Content-SHA256", evidence.SHA256Hash)
	return c.Stream(http.StatusOK, evidence.MimeType, f)
}

// Notes

func (h *ReportHandler) AddNote(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	var req dto.AddNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	authorID, _ := c.Get("account_id").(uuid.UUID)
	note, err := h.reportService.AddNote(c.Request().Context(), reportID, &req, authorID)
	if err != nil {
		if errors.Is(err, service.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToNoteResponse(note))
}

func (h *ReportHandler) ListNotes(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	notes, err := h.reportService.ListNotes(c.Request().Context(), reportID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToNoteResponseList(notes))
}
