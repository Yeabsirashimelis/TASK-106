DROP TABLE IF EXISTS idempotency_keys;
CREATE UNIQUE INDEX IF NOT EXISTS uq_idempotency_24h
    ON payments_ledger (account_id, idempotency_key)
    WHERE idempotency_expires_at IS NOT NULL;
