CREATE TYPE report_status AS ENUM (
    'Open',
    'Under Review',
    'Resolved',
    'Dismissed'
);

CREATE TYPE report_category AS ENUM (
    'Spam',
    'Harassment',
    'Inappropriate Content',
    'Policy Violation',
    'Other'
);

CREATE TABLE reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id     UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    target_type     VARCHAR(100) NOT NULL,
    target_id       UUID NOT NULL,
    category        report_category NOT NULL,
    description     TEXT NOT NULL,
    status          report_status NOT NULL DEFAULT 'Open',
    assigned_to     UUID REFERENCES accounts(id) ON DELETE SET NULL,
    resolution      TEXT,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_reporter ON reports (reporter_id);
CREATE INDEX idx_reports_target ON reports (target_type, target_id);
CREATE INDEX idx_reports_status ON reports (status);
CREATE INDEX idx_reports_assigned ON reports (assigned_to);
CREATE INDEX idx_reports_created ON reports (reporter_id, created_at);
