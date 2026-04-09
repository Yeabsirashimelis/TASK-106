-- Full-text search columns
ALTER TABLE resources ADD COLUMN search_vector tsvector;

-- Populate search_vector from title and description
CREATE OR REPLACE FUNCTION resources_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_resources_search_update
    BEFORE INSERT OR UPDATE OF title, description ON resources
    FOR EACH ROW EXECUTE FUNCTION resources_search_update();

CREATE INDEX idx_resources_search ON resources USING GIN (search_vector);

-- Full-text search on extracted text in resource_versions
ALTER TABLE resource_versions ADD COLUMN text_search_vector tsvector;

CREATE OR REPLACE FUNCTION resource_versions_search_update() RETURNS trigger AS $$
BEGIN
    IF NEW.extracted_text IS NOT NULL THEN
        NEW.text_search_vector := to_tsvector('english', NEW.extracted_text);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_resource_versions_search_update
    BEFORE INSERT OR UPDATE OF extracted_text ON resource_versions
    FOR EACH ROW EXECUTE FUNCTION resource_versions_search_update();

CREATE INDEX idx_resource_versions_text_search ON resource_versions USING GIN (text_search_vector);
