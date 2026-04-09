-- Fix #4: remove permanent unique constraint and add expiry column
-- NOTE: avoid volatile NOW() in index predicates to keep migration portable.
-- The active 24h behavior is enforced by service/repository logic and
-- the dedicated idempotency_keys table introduced in migration 031.

-- Drop the permanent unique constraint
ALTER TABLE payments_ledger DROP CONSTRAINT IF EXISTS uq_idempotency;

-- Add expiry column retained for compatibility and backfill semantics.
ALTER TABLE payments_ledger ADD COLUMN IF NOT EXISTS idempotency_expires_at TIMESTAMPTZ;

-- Backfill: set expiry for existing rows
UPDATE payments_ledger SET idempotency_expires_at = created_at + INTERVAL '24 hours' WHERE idempotency_expires_at IS NULL;

-- Non-unique support index for historical lookups.
CREATE INDEX IF NOT EXISTS idx_payments_idempotency_lookup
    ON payments_ledger (account_id, idempotency_key, idempotency_expires_at);
