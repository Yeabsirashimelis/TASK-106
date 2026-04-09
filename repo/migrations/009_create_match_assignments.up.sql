CREATE TYPE assignment_role AS ENUM (
    'Referee',
    'Assistant Referee',
    'Staff'
);

CREATE TABLE match_assignments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id             UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    account_id           UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    role                 assignment_role NOT NULL,
    assigned_by          UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    reassignment_reason  TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (match_id, account_id, role)
);

CREATE INDEX idx_match_assignments_match ON match_assignments (match_id);
CREATE INDEX idx_match_assignments_account ON match_assignments (account_id);
