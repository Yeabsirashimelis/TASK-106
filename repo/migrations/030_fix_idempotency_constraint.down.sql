DROP INDEX IF EXISTS uq_idempotency_24h;
ALTER TABLE payments_ledger DROP COLUMN IF EXISTS idempotency_expires_at;
ALTER TABLE payments_ledger ADD CONSTRAINT uq_idempotency UNIQUE (account_id, idempotency_key);
