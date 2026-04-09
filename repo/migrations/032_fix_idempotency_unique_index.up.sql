-- Fix idempotency unique index to allow key reuse after 24h window expires.
-- The previous permanent unique on (account_id, idempotency_key) blocked reuse
-- even after the window expired. Including window_start makes the constraint
-- deterministic and allows a new row with the same key after expiry.

DROP INDEX IF EXISTS uq_idempotency_active;
CREATE UNIQUE INDEX uq_idempotency_active
    ON idempotency_keys (account_id, idempotency_key, window_start);
