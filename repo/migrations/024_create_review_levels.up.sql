CREATE TYPE review_level_status AS ENUM (
    'Pending',
    'Approved',
    'Rejected',
    'Returned'
);

CREATE TABLE review_levels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      UUID NOT NULL REFERENCES review_requests(id) ON DELETE CASCADE,
    level           INTEGER NOT NULL,
    assignee_id     UUID REFERENCES accounts(id) ON DELETE SET NULL,
    status          review_level_status NOT NULL DEFAULT 'Pending',
    decision        VARCHAR(50),
    annotation      TEXT,
    decided_by      UUID REFERENCES accounts(id) ON DELETE SET NULL,
    decided_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (request_id, level)
);

CREATE INDEX idx_review_levels_request ON review_levels (request_id);
CREATE INDEX idx_review_levels_assignee ON review_levels (assignee_id);
CREATE INDEX idx_review_levels_status ON review_levels (status);
