CREATE TYPE course_status AS ENUM (
    'Draft',
    'Published',
    'Archived'
);

CREATE TABLE courses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    status      course_status NOT NULL DEFAULT 'Draft',
    created_by  UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_status ON courses (status);
CREATE INDEX idx_courses_created_by ON courses (created_by);
