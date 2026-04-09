package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReconciliationReport struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	ReportDate       time.Time       `db:"report_date" json:"report_date"`
	GeneratedBy      uuid.UUID       `db:"generated_by" json:"generated_by"`
	TotalObligations decimal.Decimal `db:"total_obligations" json:"total_obligations"`
	TotalSettled     decimal.Decimal `db:"total_settled" json:"total_settled"`
	TotalFailed      decimal.Decimal `db:"total_failed" json:"total_failed"`
	EntryCount       int             `db:"entry_count" json:"entry_count"`
	CSVPath          *string         `db:"csv_path" json:"-"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
}

// DailyLedgerSummary is used for aggregation queries
type DailyLedgerSummary struct {
	Day              time.Time       `db:"day" json:"day"`
	TotalObligations decimal.Decimal `db:"total_obligations" json:"total_obligations"`
	TotalSettled     decimal.Decimal `db:"total_settled" json:"total_settled"`
	TotalFailed      decimal.Decimal `db:"total_failed" json:"total_failed"`
	EntryCount       int             `db:"entry_count" json:"entry_count"`
}
