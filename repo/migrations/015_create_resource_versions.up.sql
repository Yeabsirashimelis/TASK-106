CREATE TABLE resource_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id     UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    version_number  INTEGER NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(255) NOT NULL,
    size_bytes      BIGINT NOT NULL,
    sha256_hash     VARCHAR(64) NOT NULL,
    storage_path    TEXT NOT NULL,
    extracted_text  TEXT,
    uploaded_by     UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (resource_id, version_number)
);

CREATE INDEX idx_resource_versions_resource ON resource_versions (resource_id);
CREATE INDEX idx_resource_versions_sha256 ON resource_versions (sha256_hash);

-- Add foreign key from resources.latest_version_id to resource_versions
ALTER TABLE resources
    ADD CONSTRAINT fk_resources_latest_version
    FOREIGN KEY (latest_version_id) REFERENCES resource_versions(id) ON DELETE SET NULL;
