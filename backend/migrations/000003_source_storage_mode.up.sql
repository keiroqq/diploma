ALTER TABLE sources
  ADD COLUMN storage_mode varchar(32) NOT NULL DEFAULT 'server';

ALTER TABLE sources
  ADD CONSTRAINT sources_storage_mode_check CHECK (storage_mode IN ('server', 'local'));

CREATE INDEX idx_sources_storage_mode ON sources(storage_mode);

CREATE UNIQUE INDEX idx_sources_public_url_unique
  ON sources(url)
  WHERE is_public = true;
