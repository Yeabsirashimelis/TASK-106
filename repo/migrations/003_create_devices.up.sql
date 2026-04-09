CREATE TABLE devices (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    fingerprint_hash VARCHAR(255) NOT NULL,
    device_name      VARCHAR(255),
    last_seen_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, fingerprint_hash)
);

CREATE INDEX idx_devices_account ON devices (account_id);
