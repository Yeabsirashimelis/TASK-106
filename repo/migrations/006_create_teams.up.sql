CREATE TABLE teams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    season_id   UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name, season_id)
);

CREATE INDEX idx_teams_season ON teams (season_id);
