CREATE TABLE reconciliation_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_date     DATE NOT NULL,
    generated_by    UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    total_obligations NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_settled     NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_failed      NUMERIC(12,2) NOT NULL DEFAULT 0,
    entry_count       INTEGER NOT NULL DEFAULT 0,
    csv_path          TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reconciliation_date ON reconciliation_reports (report_date);
