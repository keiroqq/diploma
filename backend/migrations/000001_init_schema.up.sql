CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email varchar(255) NOT NULL UNIQUE,
  password_hash varchar(255) NOT NULL,
  username varchar(120) NOT NULL,
  role varchar(32) NOT NULL DEFAULT 'user',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'))
);

CREATE TABLE feeds (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name varchar(120) NOT NULL,
  description text NOT NULL DEFAULT '',
  icon varchar(64) NOT NULL DEFAULT 'newspaper',
  theme_color varchar(32) NOT NULL DEFAULT '#2563eb',
  layout_type varchar(32) NOT NULL DEFAULT 'cards',
  is_default boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_feeds_user_id ON feeds(user_id);

CREATE TABLE sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  name varchar(160) NOT NULL,
  type varchar(32) NOT NULL DEFAULT 'rss',
  url text NOT NULL DEFAULT '',
  feed_url text NOT NULL,
  description text NOT NULL DEFAULT '',
  language varchar(16) NOT NULL DEFAULT 'ru',
  is_public boolean NOT NULL DEFAULT false,
  status varchar(32) NOT NULL DEFAULT 'active',
  last_fetched_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT sources_type_check CHECK (type IN ('rss', 'api', 'weather', 'currency', 'youtube', 'vk')),
  CONSTRAINT sources_status_check CHECK (status IN ('active', 'pending', 'disabled', 'error'))
);

CREATE INDEX idx_sources_created_by ON sources(created_by);
CREATE INDEX idx_sources_is_public ON sources(is_public);

CREATE TABLE feed_sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  feed_id uuid NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  source_id uuid NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  is_enabled boolean NOT NULL DEFAULT true,
  priority int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(feed_id, source_id)
);

CREATE INDEX idx_feed_sources_feed_id ON feed_sources(feed_id);
CREATE INDEX idx_feed_sources_source_id ON feed_sources(source_id);

CREATE TABLE feed_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id uuid NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  external_id varchar(512) NOT NULL DEFAULT '',
  guid text,
  title text NOT NULL,
  url text NOT NULL,
  canonical_url text,
  excerpt text NOT NULL DEFAULT '',
  content_html text NOT NULL DEFAULT '',
  image_url text NOT NULL DEFAULT '',
  author varchar(160) NOT NULL DEFAULT '',
  published_at timestamptz NOT NULL,
  fetched_at timestamptz NOT NULL DEFAULT now(),
  content_hash varchar(128) NOT NULL DEFAULT '',
  raw_data jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_feed_items_source_guid_unique
  ON feed_items(source_id, guid)
  WHERE guid IS NOT NULL AND guid <> '';

CREATE UNIQUE INDEX idx_feed_items_source_canonical_url_unique
  ON feed_items(source_id, canonical_url)
  WHERE canonical_url IS NOT NULL AND canonical_url <> '';

CREATE INDEX idx_feed_items_source_id ON feed_items(source_id);
CREATE INDEX idx_feed_items_published_at ON feed_items(published_at DESC);
CREATE INDEX idx_feed_items_source_published_at ON feed_items(source_id, published_at DESC);

CREATE TABLE tags (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL UNIQUE,
  slug varchar(140) NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feed_item_tags (
  item_id uuid NOT NULL REFERENCES feed_items(id) ON DELETE CASCADE,
  tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (item_id, tag_id)
);

CREATE INDEX idx_feed_item_tags_tag_id ON feed_item_tags(tag_id);

CREATE TABLE filter_rules (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  feed_id uuid NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  rule_type varchar(32) NOT NULL,
  target_type varchar(32) NOT NULL,
  value varchar(255) NOT NULL,
  weight int NOT NULL DEFAULT 0,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT filter_rules_rule_type_check CHECK (rule_type IN ('include', 'exclude', 'boost', 'downrank')),
  CONSTRAINT filter_rules_target_type_check CHECK (target_type IN ('keyword', 'tag', 'source', 'author'))
);

CREATE INDEX idx_filter_rules_feed_id ON filter_rules(feed_id);

CREATE TABLE saved_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES feed_items(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(user_id, item_id)
);

CREATE INDEX idx_saved_items_user_id ON saved_items(user_id);
CREATE INDEX idx_saved_items_item_id ON saved_items(item_id);
