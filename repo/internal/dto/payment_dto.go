package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePaymentRequest struct {
	AccountID      string  `json:"account_id"`
	IdempotencyKey string  `json:"idempotency_key"`
	AmountUSD      string  `json:"amount_usd"`
	Description    *string `json:"description,omitempty"`
	Channel        string  `json:"channel"`
	ReferenceType  *string `json:"reference_type,omitempty"`
	ReferenceID    *string `json:"reference_id,omitempty"`
}

type SignPostingRequest struct {
	// Finance Clerk confirms/signs the posting
}

type RetrySettlementRequest struct {
	// Retry a failed settlement
}

type PaymentResponse struct {
	ID             uuid.UUID       `json:"id"`
	AccountID      uuid.UUID       `json:"account_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	AmountUSD      decimal.Decimal `json:"amount_usd"`
	Description    *string         `json:"description,omitempty"`
	Channel        string          `json:"channel"`
	Status         string          `json:"status"`
	FinanceClerkID *uuid.UUID      `json:"finance_clerk_id,omitempty"`
	RetryCount     int             `json:"retry_count"`
	ReferenceType  *string         `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID      `json:"reference_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	SettledAt      *time.Time      `json:"settled_at,omitempty"`
}

func ToPaymentResponse(e *models.PaymentLedgerEntry) PaymentResponse {
	return PaymentResponse{
		ID:             e.ID,
		AccountID:      e.AccountID,
		IdempotencyKey: e.IdempotencyKey,
		AmountUSD:      e.AmountUSD,
		Description:    e.Description,
		Channel:        string(e.Channel),
		Status:         string(e.Status),
		FinanceClerkID: e.FinanceClerkID,
		RetryCount:     e.RetryCount,
		ReferenceType:  e.ReferenceType,
		ReferenceID:    e.ReferenceID,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		SettledAt:      e.SettledAt,
	}
}

func ToPaymentResponseList(entries []models.PaymentLedgerEntry) []PaymentResponse {
	result := make([]PaymentResponse, len(entries))
	for i, e := range entries {
		result[i] = ToPaymentResponse(&e)
	}
	return result
}

type DailySummaryResponse struct {
	Day              string          `json:"day"`
	TotalObligations decimal.Decimal `json:"total_obligations"`
	TotalSettled     decimal.Decimal `json:"total_settled"`
	TotalFailed      decimal.Decimal `json:"total_failed"`
	EntryCount       int             `json:"entry_count"`
}

func ToDailySummaryResponse(s *models.DailyLedgerSummary) DailySummaryResponse {
	return DailySummaryResponse{
		Day:              s.Day.Format("2006-01-02"),
		TotalObligations: s.TotalObligations,
		TotalSettled:     s.TotalSettled,
		TotalFailed:      s.TotalFailed,
		EntryCount:       s.EntryCount,
	}
}

func ToDailySummaryResponseList(summaries []models.DailyLedgerSummary) []DailySummaryResponse {
	result := make([]DailySummaryResponse, len(summaries))
	for i, s := range summaries {
		result[i] = ToDailySummaryResponse(&s)
	}
	return result
}

type ReconciliationReportResponse struct {
	ID               uuid.UUID       `json:"id"`
	ReportDate       string          `json:"report_date"`
	GeneratedBy      uuid.UUID       `json:"generated_by"`
	TotalObligations decimal.Decimal `json:"total_obligations"`
	TotalSettled     decimal.Decimal `json:"total_settled"`
	TotalFailed      decimal.Decimal `json:"total_failed"`
	EntryCount       int             `json:"entry_count"`
	CreatedAt        time.Time       `json:"created_at"`
}

func ToReconciliationReportResponse(r *models.ReconciliationReport) ReconciliationReportResponse {
	return ReconciliationReportResponse{
		ID:               r.ID,
		ReportDate:       r.ReportDate.Format("2006-01-02"),
		GeneratedBy:      r.GeneratedBy,
		TotalObligations: r.TotalObligations,
		TotalSettled:     r.TotalSettled,
		TotalFailed:      r.TotalFailed,
		EntryCount:       r.EntryCount,
		CreatedAt:        r.CreatedAt,
	}
}

func ToReconciliationReportResponseList(reports []models.ReconciliationReport) []ReconciliationReportResponse {
	result := make([]ReconciliationReportResponse, len(reports))
	for i, r := range reports {
		result[i] = ToReconciliationReportResponse(&r)
	}
	return result
}

type GenerateReconciliationRequest struct {
	Date string `json:"date"`
}

type ReconciliationRangeRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
