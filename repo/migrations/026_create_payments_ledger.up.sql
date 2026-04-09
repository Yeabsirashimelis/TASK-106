CREATE TYPE payment_status AS ENUM (
    'Obligation',
    'Settled',
    'Failed'
);

CREATE TYPE payment_channel AS ENUM (
    'Cash',
    'Check',
    'Wire Transfer',
    'Internal Transfer',
    'Journal Entry'
);

CREATE TABLE payments_ledger (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    idempotency_key  VARCHAR(255) NOT NULL,
    amount_usd       NUMERIC(12,2) NOT NULL,
    description      TEXT,
    channel          payment_channel NOT NULL,
    status           payment_status NOT NULL DEFAULT 'Obligation',
    finance_clerk_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    retry_count      INTEGER NOT NULL DEFAULT 0,
    reference_type   VARCHAR(100),
    reference_id     UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at       TIMESTAMPTZ,
    CONSTRAINT uq_idempotency UNIQUE (account_id, idempotency_key)
);

CREATE INDEX idx_payments_account ON payments_ledger (account_id);
CREATE INDEX idx_payments_status ON payments_ledger (status);
CREATE INDEX idx_payments_clerk ON payments_ledger (finance_clerk_id);
CREATE INDEX idx_payments_created ON payments_ledger (created_at);
CREATE INDEX idx_payments_idempotency ON payments_ledger (account_id, idempotency_key);
CREATE INDEX idx_payments_reference ON payments_ledger (reference_type, reference_id);
