CREATE TABLE resource_tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    tag         VARCHAR(32) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (resource_id, tag)
);

CREATE INDEX idx_resource_tags_resource ON resource_tags (resource_id);
CREATE INDEX idx_resource_tags_tag ON resource_tags (tag);
