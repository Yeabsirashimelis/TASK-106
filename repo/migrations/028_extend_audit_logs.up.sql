-- Add retention tier, snapshots, source/device info, and hash chain fields to audit_logs

CREATE TYPE log_tier AS ENUM (
    'access',
    'operation',
    'audit'
);

ALTER TABLE audit_logs
    ADD COLUMN tier              log_tier NOT NULL DEFAULT 'operation',
    ADD COLUMN reason            TEXT,
    ADD COLUMN source            VARCHAR(255),
    ADD COLUMN workstation       VARCHAR(255),
    ADD COLUMN before_snapshot   JSONB,
    ADD COLUMN after_snapshot    JSONB,
    ADD COLUMN content_hash      VARCHAR(64),
    ADD COLUMN expires_at        TIMESTAMPTZ;

CREATE INDEX idx_audit_logs_tier ON audit_logs (tier);
CREATE INDEX idx_audit_logs_expires ON audit_logs (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_audit_logs_action ON audit_logs (action);
