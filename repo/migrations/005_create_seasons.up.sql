CREATE TYPE season_status AS ENUM (
    'Planning',
    'Active',
    'Completed',
    'Archived'
);

CREATE TABLE seasons (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    status      season_status NOT NULL DEFAULT 'Planning',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_season_dates CHECK (end_date > start_date)
);
