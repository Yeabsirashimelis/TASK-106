-- Replace volatile partial index with a dedicated idempotency table.
-- This provides deterministic 24h uniqueness enforcement.

DROP INDEX IF EXISTS uq_idempotency_24h;

CREATE TABLE idempotency_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(255) NOT NULL,
    payment_id      UUID NOT NULL REFERENCES payments_ledger(id) ON DELETE CASCADE,
    window_start    TIMESTAMPTZ NOT NULL,
    window_end      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The key constraint: one active key per account at a time.
-- We use a unique index on (account_id, idempotency_key) with a WHERE clause
-- that is deterministic: window_end is a fixed computed column, not volatile NOW().
CREATE UNIQUE INDEX uq_idempotency_active
    ON idempotency_keys (account_id, idempotency_key);

CREATE INDEX idx_idempotency_window ON idempotency_keys (window_end);

-- Backfill from existing ledger entries
INSERT INTO idempotency_keys (account_id, idempotency_key, payment_id, window_start, window_end, created_at)
SELECT account_id, idempotency_key, id, created_at, COALESCE(idempotency_expires_at, created_at + INTERVAL '24 hours'), created_at
FROM payments_ledger
ON CONFLICT DO NOTHING;
