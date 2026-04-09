CREATE TABLE review_follow_ups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      UUID NOT NULL REFERENCES review_requests(id) ON DELETE CASCADE,
    author_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    content         TEXT NOT NULL,
    level           INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_follow_ups_request ON review_follow_ups (request_id);
