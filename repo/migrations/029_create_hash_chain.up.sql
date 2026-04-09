CREATE TABLE audit_hash_chain (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_date      DATE NOT NULL UNIQUE,
    entry_count     INTEGER NOT NULL DEFAULT 0,
    batch_hash      VARCHAR(64) NOT NULL,
    previous_hash   VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_hash_chain_date ON audit_hash_chain (chain_date);
