package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, entry *models.PaymentLedgerEntry) error {
	query := `INSERT INTO payments_ledger (id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at, idempotency_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.AccountID, entry.IdempotencyKey, entry.AmountUSD,
		entry.Description, entry.Channel, entry.Status, entry.FinanceClerkID,
		entry.RetryCount, entry.ReferenceType, entry.ReferenceID,
		entry.CreatedAt, entry.UpdatedAt, entry.SettledAt, entry.IdempotencyExpiresAt)
	if err != nil {
		return fmt.Errorf("payment_repo.Create: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PaymentLedgerEntry, error) {
	var entry models.PaymentLedgerEntry
	query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
		FROM payments_ledger WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&entry); err != nil {
		return nil, fmt.Errorf("payment_repo.GetByID: %w", err)
	}
	return &entry, nil
}

func (r *PaymentRepository) GetByIdempotencyKey(ctx context.Context, accountID uuid.UUID, key string) (*models.PaymentLedgerEntry, error) {
	var entry models.PaymentLedgerEntry
	query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
		FROM payments_ledger WHERE account_id = $1 AND idempotency_key = $2`
	if err := r.db.QueryRowxContext(ctx, query, accountID, key).StructScan(&entry); err != nil {
		return nil, fmt.Errorf("payment_repo.GetByIdempotencyKey: %w", err)
	}
	return &entry, nil
}

func (r *PaymentRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]models.PaymentLedgerEntry, error) {
	var entries []models.PaymentLedgerEntry
	query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
		FROM payments_ledger WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &entries, query, accountID, limit, offset); err != nil {
		return nil, fmt.Errorf("payment_repo.ListByAccount: %w", err)
	}
	return entries, nil
}

func (r *PaymentRepository) List(ctx context.Context, status *models.PaymentStatus, offset, limit int) ([]models.PaymentLedgerEntry, error) {
	var entries []models.PaymentLedgerEntry
	if status != nil {
		query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
			FROM payments_ledger WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		if err := r.db.SelectContext(ctx, &entries, query, *status, limit, offset); err != nil {
			return nil, fmt.Errorf("payment_repo.List: %w", err)
		}
	} else {
		query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
			FROM payments_ledger ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		if err := r.db.SelectContext(ctx, &entries, query, limit, offset); err != nil {
			return nil, fmt.Errorf("payment_repo.List: %w", err)
		}
	}
	return entries, nil
}

func (r *PaymentRepository) ListFailedRetriable(ctx context.Context) ([]models.PaymentLedgerEntry, error) {
	var entries []models.PaymentLedgerEntry
	query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
		FROM payments_ledger WHERE status = 'Failed' AND retry_count < $1 ORDER BY created_at`
	if err := r.db.SelectContext(ctx, &entries, query, models.MaxSettlementRetries); err != nil {
		return nil, fmt.Errorf("payment_repo.ListFailedRetriable: %w", err)
	}
	return entries, nil
}

func (r *PaymentRepository) Update(ctx context.Context, entry *models.PaymentLedgerEntry) error {
	query := `UPDATE payments_ledger SET status = $1, finance_clerk_id = $2, retry_count = $3, settled_at = $4, updated_at = NOW()
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, entry.Status, entry.FinanceClerkID, entry.RetryCount, entry.SettledAt, entry.ID)
	if err != nil {
		return fmt.Errorf("payment_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payment_repo.Update: entry not found")
	}
	return nil
}

// Reconciliation queries

func (r *PaymentRepository) GetDailySummary(ctx context.Context, date time.Time) (*models.DailyLedgerSummary, error) {
	var summary models.DailyLedgerSummary
	query := `SELECT
		$1::date AS day,
		COALESCE(SUM(CASE WHEN status = 'Obligation' THEN amount_usd ELSE 0 END), 0) AS total_obligations,
		COALESCE(SUM(CASE WHEN status = 'Settled' THEN amount_usd ELSE 0 END), 0) AS total_settled,
		COALESCE(SUM(CASE WHEN status = 'Failed' THEN amount_usd ELSE 0 END), 0) AS total_failed,
		COUNT(*) AS entry_count
		FROM payments_ledger
		WHERE created_at >= $1::date AND created_at < ($1::date + INTERVAL '1 day')`
	if err := r.db.QueryRowxContext(ctx, query, date).StructScan(&summary); err != nil {
		return nil, fmt.Errorf("payment_repo.GetDailySummary: %w", err)
	}
	return &summary, nil
}

func (r *PaymentRepository) GetDailySummaryRange(ctx context.Context, startDate, endDate time.Time) ([]models.DailyLedgerSummary, error) {
	var summaries []models.DailyLedgerSummary
	query := `SELECT
		DATE(created_at) AS day,
		COALESCE(SUM(CASE WHEN status = 'Obligation' THEN amount_usd ELSE 0 END), 0) AS total_obligations,
		COALESCE(SUM(CASE WHEN status = 'Settled' THEN amount_usd ELSE 0 END), 0) AS total_settled,
		COALESCE(SUM(CASE WHEN status = 'Failed' THEN amount_usd ELSE 0 END), 0) AS total_failed,
		COUNT(*) AS entry_count
		FROM payments_ledger
		WHERE created_at >= $1::date AND created_at < ($2::date + INTERVAL '1 day')
		GROUP BY DATE(created_at)
		ORDER BY day`
	if err := r.db.SelectContext(ctx, &summaries, query, startDate, endDate); err != nil {
		return nil, fmt.Errorf("payment_repo.GetDailySummaryRange: %w", err)
	}
	return summaries, nil
}

func (r *PaymentRepository) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.PaymentLedgerEntry, error) {
	var entries []models.PaymentLedgerEntry
	query := `SELECT id, account_id, idempotency_key, amount_usd, description, channel, status, finance_clerk_id, retry_count, reference_type, reference_id, created_at, updated_at, settled_at
		FROM payments_ledger
		WHERE created_at >= $1::date AND created_at < ($2::date + INTERVAL '1 day')
		ORDER BY created_at`
	if err := r.db.SelectContext(ctx, &entries, query, startDate, endDate); err != nil {
		return nil, fmt.Errorf("payment_repo.ListByDateRange: %w", err)
	}
	return entries, nil
}

// Idempotency key management

func (r *PaymentRepository) FindActiveIdempotencyKey(ctx context.Context, accountID uuid.UUID, key string, now time.Time) (*models.IdempotencyKey, error) {
	var ik models.IdempotencyKey
	query := `SELECT id, account_id, idempotency_key, payment_id, window_start, window_end, created_at
		FROM idempotency_keys WHERE account_id = $1 AND idempotency_key = $2 AND window_end > $3`
	if err := r.db.QueryRowxContext(ctx, query, accountID, key, now).StructScan(&ik); err != nil {
		return nil, fmt.Errorf("payment_repo.FindActiveIdempotencyKey: %w", err)
	}
	return &ik, nil
}

func (r *PaymentRepository) CreateIdempotencyKey(ctx context.Context, ik *models.IdempotencyKey) error {
	query := `INSERT INTO idempotency_keys (id, account_id, idempotency_key, payment_id, window_start, window_end, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, ik.ID, ik.AccountID, ik.IdempotencyKey, ik.PaymentID, ik.WindowStart, ik.WindowEnd, ik.CreatedAt)
	if err != nil {
		return fmt.Errorf("payment_repo.CreateIdempotencyKey: %w", err)
	}
	return nil
}

func (r *PaymentRepository) DeleteExpiredIdempotencyKeys(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE window_end <= $1`, now)
	if err != nil {
		return 0, fmt.Errorf("payment_repo.DeleteExpiredIdempotencyKeys: %w", err)
	}
	return result.RowsAffected()
}

func (r *PaymentRepository) DeleteExpiredIdempotencyKeyForAccount(ctx context.Context, accountID uuid.UUID, key string, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM idempotency_keys WHERE account_id = $1 AND idempotency_key = $2 AND window_end <= $3`,
		accountID, key, now)
	if err != nil {
		return fmt.Errorf("payment_repo.DeleteExpiredIdempotencyKeyForAccount: %w", err)
	}
	return nil
}

// Reconciliation report CRUD

func (r *PaymentRepository) CreateReconciliationReport(ctx context.Context, report *models.ReconciliationReport) error {
	query := `INSERT INTO reconciliation_reports (id, report_date, generated_by, total_obligations, total_settled, total_failed, entry_count, csv_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.ReportDate, report.GeneratedBy,
		report.TotalObligations, report.TotalSettled, report.TotalFailed,
		report.EntryCount, report.CSVPath, report.CreatedAt)
	if err != nil {
		return fmt.Errorf("payment_repo.CreateReconciliationReport: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetReconciliationReport(ctx context.Context, id uuid.UUID) (*models.ReconciliationReport, error) {
	var report models.ReconciliationReport
	query := `SELECT id, report_date, generated_by, total_obligations, total_settled, total_failed, entry_count, csv_path, created_at
		FROM reconciliation_reports WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&report); err != nil {
		return nil, fmt.Errorf("payment_repo.GetReconciliationReport: %w", err)
	}
	return &report, nil
}

func (r *PaymentRepository) ListReconciliationReports(ctx context.Context, offset, limit int) ([]models.ReconciliationReport, error) {
	var reports []models.ReconciliationReport
	query := `SELECT id, report_date, generated_by, total_obligations, total_settled, total_failed, entry_count, csv_path, created_at
		FROM reconciliation_reports ORDER BY report_date DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &reports, query, limit, offset); err != nil {
		return nil, fmt.Errorf("payment_repo.ListReconciliationReports: %w", err)
	}
	return reports, nil
}

// ExpireIdempotencyKeys is a cleanup helper (not called automatically)
func (r *PaymentRepository) CountByIdempotencyKeyRecent(ctx context.Context, accountID uuid.UUID, key string, within time.Duration) (int, error) {
	var count int
	cutoff := time.Now().Add(-within)
	query := `SELECT COUNT(*) FROM payments_ledger WHERE account_id = $1 AND idempotency_key = $2 AND created_at > $3`
	if err := r.db.QueryRowxContext(ctx, query, accountID, key, cutoff).Scan(&count); err != nil {
		return 0, fmt.Errorf("payment_repo.CountByIdempotencyKeyRecent: %w", err)
	}
	return count, nil
}

// Unused but useful for reporting: sum by arbitrary grouping
func (r *PaymentRepository) SumByStatus(ctx context.Context) (map[models.PaymentStatus]decimal.Decimal, error) {
	type row struct {
		Status PaymentStatus   `db:"status"`
		Total  decimal.Decimal `db:"total"`
	}
	var rows []row
	query := `SELECT status, COALESCE(SUM(amount_usd), 0) AS total FROM payments_ledger GROUP BY status`
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("payment_repo.SumByStatus: %w", err)
	}
	result := make(map[models.PaymentStatus]decimal.Decimal)
	for _, r := range rows {
		result[r.Status] = r.Total
	}
	return result, nil
}

type PaymentStatus = models.PaymentStatus
