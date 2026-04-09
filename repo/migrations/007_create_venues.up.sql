CREATE TABLE venues (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    location    VARCHAR(500),
    capacity    INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
