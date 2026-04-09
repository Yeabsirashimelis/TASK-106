CREATE TYPE resource_type AS ENUM (
    'Document',
    'Video',
    'Link'
);

CREATE TYPE resource_visibility AS ENUM (
    'Staff',
    'Enrolled'
);

CREATE TABLE resources (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id         UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    node_id           UUID REFERENCES course_outline_nodes(id) ON DELETE SET NULL,
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    resource_type     resource_type NOT NULL,
    visibility        resource_visibility NOT NULL DEFAULT 'Staff',
    link_url          TEXT,
    latest_version_id UUID,
    created_by        UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resources_course ON resources (course_id);
CREATE INDEX idx_resources_node ON resources (node_id);
CREATE INDEX idx_resources_type ON resources (resource_type);
CREATE INDEX idx_resources_visibility ON resources (visibility);
