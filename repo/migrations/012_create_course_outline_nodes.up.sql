CREATE TYPE outline_node_type AS ENUM (
    'Chapter',
    'Unit'
);

CREATE TABLE course_outline_nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES course_outline_nodes(id) ON DELETE CASCADE,
    node_type   outline_node_type NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outline_nodes_course ON course_outline_nodes (course_id);
CREATE INDEX idx_outline_nodes_parent ON course_outline_nodes (parent_id);
CREATE INDEX idx_outline_nodes_order ON course_outline_nodes (course_id, parent_id, order_index);
