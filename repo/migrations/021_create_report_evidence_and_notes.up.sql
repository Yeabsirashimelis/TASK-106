CREATE TABLE report_evidence (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id       UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    file_name       VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(255) NOT NULL,
    size_bytes      BIGINT NOT NULL,
    sha256_hash     VARCHAR(64) NOT NULL,
    storage_path    TEXT NOT NULL,
    uploaded_by     UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_evidence_report ON report_evidence (report_id);

CREATE TABLE report_notes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id   UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_notes_report ON report_notes (report_id);
