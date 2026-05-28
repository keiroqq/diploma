DROP INDEX IF EXISTS idx_sources_public_url_unique;
DROP INDEX IF EXISTS idx_sources_storage_mode;

ALTER TABLE sources
  DROP CONSTRAINT IF EXISTS sources_storage_mode_check;

ALTER TABLE sources
  DROP COLUMN IF EXISTS storage_mode;
