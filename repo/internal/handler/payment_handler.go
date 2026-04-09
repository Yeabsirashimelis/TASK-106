package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	var req dto.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	entry, isDuplicate, err := h.paymentService.CreatePayment(c.Request().Context(), &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp := dto.ToPaymentResponse(entry)
	if isDuplicate {
		// Return original response with 200 instead of 201
		return c.JSON(http.StatusOK, resp)
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *PaymentHandler) GetPayment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payment id")
	}

	entry, err := h.paymentService.GetPayment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponse(entry))
}

func (h *PaymentHandler) ListPayments(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	var statusFilter *string
	if s := c.QueryParam("status"); s != "" {
		statusFilter = &s
	}

	entries, err := h.paymentService.ListPayments(c.Request().Context(), statusFilter, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponseList(entries))
}

func (h *PaymentHandler) ListByAccount(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("account_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account_id")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	entries, err := h.paymentService.ListByAccount(c.Request().Context(), accountID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponseList(entries))
}

func (h *PaymentHandler) SignPosting(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payment id")
	}

	clerkID, _ := c.Get("account_id").(uuid.UUID)
	entry, err := h.paymentService.SignPosting(c.Request().Context(), id, clerkID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotObligation) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponse(entry))
}

func (h *PaymentHandler) FailSettlement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payment id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	entry, err := h.paymentService.FailSettlement(c.Request().Context(), id, actorID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotObligation) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponse(entry))
}

func (h *PaymentHandler) RetrySettlement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payment id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	entry, err := h.paymentService.RetrySettlement(c.Request().Context(), id, actorID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotFailed) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if errors.Is(err, service.ErrMaxRetriesExceeded) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponse(entry))
}

func (h *PaymentHandler) ListFailedRetriable(c echo.Context) error {
	entries, err := h.paymentService.ListFailedRetriable(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToPaymentResponseList(entries))
}

// Reconciliation endpoints

func (h *PaymentHandler) GetDailySummary(c echo.Context) error {
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "date query parameter is required (YYYY-MM-DD)")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
	}

	summary, err := h.paymentService.GetDailySummary(c.Request().Context(), date)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToDailySummaryResponse(summary))
}

func (h *PaymentHandler) GetSummaryRange(c echo.Context) error {
	startStr := c.QueryParam("start_date")
	endStr := c.QueryParam("end_date")
	if startStr == "" || endStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "start_date and end_date query parameters are required")
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date format")
	}
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date format")
	}

	summaries, err := h.paymentService.GetSummaryRange(c.Request().Context(), startDate, endDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToDailySummaryResponseList(summaries))
}

func (h *PaymentHandler) GenerateReconciliation(c echo.Context) error {
	var req dto.GenerateReconciliationRequest
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

	actorID, _ := c.Get("account_id").(uuid.UUID)
	report, err := h.paymentService.GenerateReconciliationReport(c.Request().Context(), date, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToReconciliationReportResponse(report))
}

func (h *PaymentHandler) ListReconciliationReports(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	reports, err := h.paymentService.ListReconciliationReports(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToReconciliationReportResponseList(reports))
}

func (h *PaymentHandler) GetReconciliationReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	report, err := h.paymentService.GetReconciliationReport(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToReconciliationReportResponse(report))
}

func (h *PaymentHandler) DownloadReconciliationCSV(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid report id")
	}

	csvPath, err := h.paymentService.GetCSVPath(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrReconNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	f, err := os.Open(csvPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "CSV file not found on disk")
	}
	defer f.Close()

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="reconciliation_%s.csv"`, id.String()))
	return c.Stream(http.StatusOK, "text/csv", f)
}
