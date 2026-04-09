ALTER TABLE resources DROP CONSTRAINT IF EXISTS fk_resources_latest_version;
DROP TABLE IF EXISTS resource_versions;
