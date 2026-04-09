-- Fix #4: Replace permanent unique constraint with partial index for 24h window
-- Drop the permanent unique constraint
ALTER TABLE payments_ledger DROP CONSTRAINT IF EXISTS uq_idempotency;

-- Create a partial unique index that only covers recent entries (non-expired)
-- The service layer enforces the 24h window; the DB index prevents races within that window.
-- We add an idempotency_expires_at column to make the constraint enforceable at DB level.
ALTER TABLE payments_ledger ADD COLUMN IF NOT EXISTS idempotency_expires_at TIMESTAMPTZ;

-- Backfill: set expiry for existing rows
UPDATE payments_ledger SET idempotency_expires_at = created_at + INTERVAL '24 hours' WHERE idempotency_expires_at IS NULL;

-- Create unique index only for non-expired idempotency keys
CREATE UNIQUE INDEX IF NOT EXISTS uq_idempotency_24h
    ON payments_ledger (account_id, idempotency_key)
    WHERE idempotency_expires_at > NOW();
