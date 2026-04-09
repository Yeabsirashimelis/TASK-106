CREATE TYPE match_status AS ENUM (
    'Draft',
    'Scheduled',
    'In-Progress',
    'Final',
    'Canceled'
);

CREATE TABLE matches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id       UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    round           INTEGER NOT NULL,
    home_team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    away_team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    venue_id        UUID NOT NULL REFERENCES venues(id) ON DELETE RESTRICT,
    scheduled_at    TIMESTAMPTZ NOT NULL,
    status          match_status NOT NULL DEFAULT 'Draft',
    override_reason TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_different_teams CHECK (home_team_id != away_team_id)
);

CREATE INDEX idx_matches_season ON matches (season_id);
CREATE INDEX idx_matches_round ON matches (season_id, round);
CREATE INDEX idx_matches_venue_time ON matches (venue_id, scheduled_at);
CREATE INDEX idx_matches_home_team ON matches (home_team_id);
CREATE INDEX idx_matches_away_team ON matches (away_team_id);
CREATE INDEX idx_matches_status ON matches (status);
