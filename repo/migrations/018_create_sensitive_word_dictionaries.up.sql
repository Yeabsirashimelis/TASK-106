CREATE TABLE sensitive_word_dictionaries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_by  UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sensitive_words (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dictionary_id   UUID NOT NULL REFERENCES sensitive_word_dictionaries(id) ON DELETE CASCADE,
    word            VARCHAR(255) NOT NULL,
    severity        VARCHAR(50) NOT NULL DEFAULT 'medium',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (dictionary_id, word)
);

CREATE INDEX idx_sensitive_words_dictionary ON sensitive_words (dictionary_id);
CREATE INDEX idx_sensitive_words_word ON sensitive_words (LOWER(word));
