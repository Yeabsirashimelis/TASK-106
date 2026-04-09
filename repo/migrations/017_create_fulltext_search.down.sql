DROP TRIGGER IF EXISTS trg_resource_versions_search_update ON resource_versions;
DROP FUNCTION IF EXISTS resource_versions_search_update;
ALTER TABLE resource_versions DROP COLUMN IF EXISTS text_search_vector;

DROP TRIGGER IF EXISTS trg_resources_search_update ON resources;
DROP FUNCTION IF EXISTS resources_search_update;
ALTER TABLE resources DROP COLUMN IF EXISTS search_vector;
