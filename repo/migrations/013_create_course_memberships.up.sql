CREATE TYPE membership_role AS ENUM (
    'Staff',
    'Enrolled'
);

CREATE TABLE course_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role        membership_role NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, account_id)
);

CREATE INDEX idx_course_memberships_course ON course_memberships (course_id);
CREATE INDEX idx_course_memberships_account ON course_memberships (account_id);
