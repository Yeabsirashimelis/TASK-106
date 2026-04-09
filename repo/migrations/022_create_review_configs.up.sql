CREATE TABLE review_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_type     VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    required_levels INTEGER NOT NULL DEFAULT 1,
    created_by      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_levels CHECK (required_levels >= 1 AND required_levels <= 3)
);

CREATE INDEX idx_review_configs_type ON review_configs (review_type);
