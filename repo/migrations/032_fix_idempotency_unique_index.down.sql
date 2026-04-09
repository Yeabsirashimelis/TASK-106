DROP INDEX IF EXISTS uq_idempotency_active;
CREATE UNIQUE INDEX uq_idempotency_active
    ON idempotency_keys (account_id, idempotency_key);
