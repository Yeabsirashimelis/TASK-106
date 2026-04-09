package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const idempotencyWindow = 24 * time.Hour

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrDuplicatePayment   = errors.New("duplicate idempotency key")
	ErrNotObligation      = errors.New("payment is not in Obligation status")
	ErrNotFailed          = errors.New("payment is not in Failed status")
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrReconNotFound      = errors.New("reconciliation report not found")
)

type PaymentService struct {
	repo  *repository.PaymentRepository
	audit *AuditService
	cfg   *config.Config
}

func NewPaymentService(
	repo *repository.PaymentRepository,
	audit *AuditService,
	cfg *config.Config,
) *PaymentService {
	return &PaymentService{
		repo:  repo,
		audit: audit,
		cfg:   cfg,
	}
}

// CreatePayment creates a new obligation with idempotency. If a duplicate key is found
// within 24 hours for the same account, the original entry is returned.
func (s *PaymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest, actorID uuid.UUID) (*models.PaymentLedgerEntry, bool, error) {
	if req.IdempotencyKey == "" {
		return nil, false, fmt.Errorf("idempotency_key is required")
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, false, fmt.Errorf("invalid account_id")
	}

	amount, err := decimal.NewFromString(req.AmountUSD)
	if err != nil {
		return nil, false, fmt.Errorf("invalid amount_usd")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, false, ErrInvalidAmount
	}

	channel := models.PaymentChannel(req.Channel)
	if !models.ValidPaymentChannels[channel] {
		return nil, false, fmt.Errorf("invalid channel: %s", req.Channel)
	}

	// Check idempotency: look for active key in dedicated table within 24h window
	now := time.Now()
	activeKey, findErr := s.repo.FindActiveIdempotencyKey(ctx, accountID, req.IdempotencyKey, now)
	if findErr == nil && activeKey != nil {
		// Active key found — return original payment
		orig, getErr := s.repo.GetByID(ctx, activeKey.PaymentID)
		if getErr == nil {
			return orig, true, nil
		}
	}

	// Clean up expired idempotency keys (best-effort)
	s.repo.DeleteExpiredIdempotencyKeys(ctx, now)

	var refType *string
	var refID *uuid.UUID
	if req.ReferenceType != nil {
		refType = req.ReferenceType
	}
	if req.ReferenceID != nil {
		rid, err := uuid.Parse(*req.ReferenceID)
		if err != nil {
			return nil, false, fmt.Errorf("invalid reference_id")
		}
		refID = &rid
	}

	idempotencyExpiry := now.Add(idempotencyWindow)
	entry := &models.PaymentLedgerEntry{
		ID:                   uuid.New(),
		AccountID:            accountID,
		IdempotencyKey:       req.IdempotencyKey,
		AmountUSD:            amount,
		Description:          req.Description,
		Channel:              channel,
		Status:               models.PaymentObligation,
		RetryCount:           0,
		ReferenceType:        refType,
		ReferenceID:          refID,
		CreatedAt:            now,
		UpdatedAt:            now,
		IdempotencyExpiresAt: &idempotencyExpiry,
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, false, err
	}

	// Record idempotency key in dedicated table
	ik := &models.IdempotencyKey{
		ID:             uuid.New(),
		AccountID:      accountID,
		IdempotencyKey: req.IdempotencyKey,
		PaymentID:      entry.ID,
		WindowStart:    now,
		WindowEnd:      idempotencyExpiry,
		CreatedAt:      now,
	}
	s.repo.CreateIdempotencyKey(ctx, ik)

	s.audit.Log(ctx, "payment", entry.ID, actorID, "obligation_created", map[string]interface{}{
		"account_id":      accountID,
		"amount_usd":      amount.String(),
		"channel":         string(channel),
		"idempotency_key": req.IdempotencyKey,
	})

	return entry, false, nil
}

// SignPosting is the Finance Clerk's explicit confirmation that settles an obligation.
func (s *PaymentService) SignPosting(ctx context.Context, paymentID, clerkID uuid.UUID) (*models.PaymentLedgerEntry, error) {
	entry, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	if entry.Status != models.PaymentObligation {
		return nil, ErrNotObligation
	}

	beforeStatus := string(entry.Status)

	now := time.Now()
	entry.Status = models.PaymentSettled
	entry.FinanceClerkID = &clerkID
	entry.SettledAt = &now

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, err
	}

	s.audit.LogExtended(ctx, &AuditEntry{
		EntityType: "payment",
		EntityID:   paymentID,
		ActorID:    clerkID,
		Action:     "posting_signed",
		Tier:       models.TierAudit,
		BeforeSnapshot: map[string]interface{}{
			"status": beforeStatus,
		},
		AfterSnapshot: map[string]interface{}{
			"status":           string(entry.Status),
			"finance_clerk_id": clerkID,
			"settled_at":       now,
		},
		Details: map[string]interface{}{
			"account_id": entry.AccountID,
			"amount_usd": entry.AmountUSD.String(),
			"channel":    string(entry.Channel),
		},
	})

	return entry, nil
}

// FailSettlement marks an obligation as failed (e.g., check bounced, wire rejected).
func (s *PaymentService) FailSettlement(ctx context.Context, paymentID, actorID uuid.UUID) (*models.PaymentLedgerEntry, error) {
	entry, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	if entry.Status != models.PaymentObligation {
		return nil, ErrNotObligation
	}

	entry.Status = models.PaymentFailed

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "payment", paymentID, actorID, "settlement_failed", map[string]interface{}{
		"account_id": entry.AccountID,
		"amount_usd": entry.AmountUSD.String(),
	})

	return entry, nil
}

// RetrySettlement retries a failed settlement (up to MaxSettlementRetries).
// On retry, it moves back to Obligation for the clerk to re-sign.
func (s *PaymentService) RetrySettlement(ctx context.Context, paymentID, actorID uuid.UUID) (*models.PaymentLedgerEntry, error) {
	entry, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	if entry.Status != models.PaymentFailed {
		return nil, ErrNotFailed
	}

	if entry.RetryCount >= models.MaxSettlementRetries {
		return nil, ErrMaxRetriesExceeded
	}

	entry.RetryCount++
	entry.Status = models.PaymentObligation
	entry.FinanceClerkID = nil
	entry.SettledAt = nil

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "payment", paymentID, actorID, "settlement_retried", map[string]interface{}{
		"account_id":  entry.AccountID,
		"retry_count": entry.RetryCount,
		"amount_usd":  entry.AmountUSD.String(),
	})

	return entry, nil
}

// Query methods

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*models.PaymentLedgerEntry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return entry, nil
}

func (s *PaymentService) ListPayments(ctx context.Context, statusFilter *string, offset, limit int) ([]models.PaymentLedgerEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var status *models.PaymentStatus
	if statusFilter != nil {
		st := models.PaymentStatus(*statusFilter)
		if models.ValidPaymentStatuses[st] {
			status = &st
		}
	}
	return s.repo.List(ctx, status, offset, limit)
}

func (s *PaymentService) ListByAccount(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]models.PaymentLedgerEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByAccount(ctx, accountID, offset, limit)
}

func (s *PaymentService) ListFailedRetriable(ctx context.Context) ([]models.PaymentLedgerEntry, error) {
	return s.repo.ListFailedRetriable(ctx)
}

// Reconciliation

func (s *PaymentService) GetDailySummary(ctx context.Context, date time.Time) (*models.DailyLedgerSummary, error) {
	return s.repo.GetDailySummary(ctx, date)
}

func (s *PaymentService) GetSummaryRange(ctx context.Context, startDate, endDate time.Time) ([]models.DailyLedgerSummary, error) {
	return s.repo.GetDailySummaryRange(ctx, startDate, endDate)
}

func (s *PaymentService) GenerateReconciliationReport(ctx context.Context, date time.Time, actorID uuid.UUID) (*models.ReconciliationReport, error) {
	// Get daily summary
	summary, err := s.repo.GetDailySummary(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("getting daily summary: %w", err)
	}

	// Get all entries for the day to export as CSV
	entries, err := s.repo.ListByDateRange(ctx, date, date)
	if err != nil {
		return nil, fmt.Errorf("listing entries: %w", err)
	}

	// Generate CSV
	reportID := uuid.New()
	csvDir := filepath.Join(s.cfg.StoragePath, "reconciliation")
	os.MkdirAll(csvDir, 0750)
	csvPath := filepath.Join(csvDir, fmt.Sprintf("%s_%s.csv", date.Format("2006-01-02"), reportID.String()))

	if err := s.writeCSV(csvPath, entries); err != nil {
		return nil, fmt.Errorf("writing CSV: %w", err)
	}

	report := &models.ReconciliationReport{
		ID:               reportID,
		ReportDate:       date,
		GeneratedBy:      actorID,
		TotalObligations: summary.TotalObligations,
		TotalSettled:     summary.TotalSettled,
		TotalFailed:      summary.TotalFailed,
		EntryCount:       summary.EntryCount,
		CSVPath:          &csvPath,
		CreatedAt:        time.Now(),
	}

	if err := s.repo.CreateReconciliationReport(ctx, report); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "reconciliation_report", reportID, actorID, "generated", map[string]interface{}{
		"report_date":       date.Format("2006-01-02"),
		"total_obligations": summary.TotalObligations.String(),
		"total_settled":     summary.TotalSettled.String(),
		"total_failed":      summary.TotalFailed.String(),
		"entry_count":       summary.EntryCount,
	})

	return report, nil
}

func (s *PaymentService) GetReconciliationReport(ctx context.Context, id uuid.UUID) (*models.ReconciliationReport, error) {
	report, err := s.repo.GetReconciliationReport(ctx, id)
	if err != nil {
		return nil, ErrReconNotFound
	}
	return report, nil
}

func (s *PaymentService) ListReconciliationReports(ctx context.Context, offset, limit int) ([]models.ReconciliationReport, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListReconciliationReports(ctx, offset, limit)
}

func (s *PaymentService) GetCSVPath(ctx context.Context, reportID uuid.UUID) (string, error) {
	report, err := s.repo.GetReconciliationReport(ctx, reportID)
	if err != nil {
		return "", ErrReconNotFound
	}
	if report.CSVPath == nil {
		return "", fmt.Errorf("no CSV file for this report")
	}
	return *report.CSVPath, nil
}

func (s *PaymentService) writeCSV(path string, entries []models.PaymentLedgerEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	w.Write([]string{
		"id", "account_id", "idempotency_key", "amount_usd", "description",
		"channel", "status", "finance_clerk_id", "retry_count",
		"reference_type", "reference_id", "created_at", "settled_at",
	})

	for _, e := range entries {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		clerkID := ""
		if e.FinanceClerkID != nil {
			clerkID = e.FinanceClerkID.String()
		}
		refType := ""
		if e.ReferenceType != nil {
			refType = *e.ReferenceType
		}
		refID := ""
		if e.ReferenceID != nil {
			refID = e.ReferenceID.String()
		}
		settledAt := ""
		if e.SettledAt != nil {
			settledAt = e.SettledAt.Format(time.RFC3339)
		}

		w.Write([]string{
			e.ID.String(),
			e.AccountID.String(),
			e.IdempotencyKey,
			e.AmountUSD.String(),
			desc,
			string(e.Channel),
			string(e.Status),
			clerkID,
			fmt.Sprintf("%d", e.RetryCount),
			refType,
			refID,
			e.CreatedAt.Format(time.RFC3339),
			settledAt,
		})
	}

	return nil
}
