CREATE TYPE review_status AS ENUM (
    'Pending',
    'Approved',
    'Rejected'
);

CREATE TABLE moderation_reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type    VARCHAR(100) NOT NULL,
    content_id      UUID NOT NULL,
    content_snippet TEXT,
    status          review_status NOT NULL DEFAULT 'Pending',
    moderator_id    UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    reason          TEXT,
    decided_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moderation_reviews_status ON moderation_reviews (status);
CREATE INDEX idx_moderation_reviews_content ON moderation_reviews (content_type, content_id);
CREATE INDEX idx_moderation_reviews_moderator ON moderation_reviews (moderator_id);
