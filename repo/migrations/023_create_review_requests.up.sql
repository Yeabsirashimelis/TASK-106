CREATE TYPE review_request_status AS ENUM (
    'In Review',
    'Approved',
    'Rejected',
    'Returned'
);

CREATE TABLE review_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_type     VARCHAR(100) NOT NULL,
    entity_type     VARCHAR(100) NOT NULL,
    entity_id       UUID NOT NULL,
    required_levels INTEGER NOT NULL DEFAULT 1,
    current_level   INTEGER NOT NULL DEFAULT 1,
    status          review_request_status NOT NULL DEFAULT 'In Review',
    submitted_by    UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    final_decision  VARCHAR(50),
    parent_id       UUID REFERENCES review_requests(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_requests_type ON review_requests (review_type);
CREATE INDEX idx_review_requests_entity ON review_requests (entity_type, entity_id);
CREATE INDEX idx_review_requests_status ON review_requests (status);
CREATE INDEX idx_review_requests_submitter ON review_requests (submitted_by);
CREATE INDEX idx_review_requests_parent ON review_requests (parent_id);
