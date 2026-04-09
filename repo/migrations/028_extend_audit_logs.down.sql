DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_expires;
DROP INDEX IF EXISTS idx_audit_logs_tier;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS reason,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS workstation,
    DROP COLUMN IF EXISTS before_snapshot,
    DROP COLUMN IF EXISTS after_snapshot,
    DROP COLUMN IF EXISTS content_hash,
    DROP COLUMN IF EXISTS expires_at;

DROP TYPE IF EXISTS log_tier;
