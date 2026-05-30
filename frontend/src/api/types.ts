export type User = {
  id: string;
  email: string;
  username: string;
  role: string;
  created_at: string;
};

export type AuthResponse = {
  token: string;
  user: User;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type RegisterRequest = LoginRequest & {
  username: string;
};

export type Feed = {
  id: string;
  name: string;
  description: string;
  icon: string;
  theme_color: string;
  layout_type: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};

export type Source = {
  id: string;
  name: string;
  type: string;
  url: string;
  feed_url: string;
  description: string;
  language: string;
  is_public: boolean;
  storage_mode: "server" | "local";
  status: "active" | "pending" | "disabled" | "error";
  last_fetched_at?: string;
  created_at?: string;
  updated_at?: string;
};

export type CreateSourceRequest = {
  name: string;
  type: "rss";
  url: string;
  feed_url: string;
  description: string;
  language: string;
  is_public: boolean;
  storage_mode: "server" | "local";
};

export type UpdateSourceRequest = CreateSourceRequest & {
  status: "active" | "pending" | "disabled" | "error";
};

export type FeedSource = {
  id: string;
  feed_id: string;
  source_id: string;
  is_enabled: boolean;
  priority: number;
  created_at: string;
  source?: Source;
};

export type CreateFeedRequest = {
  name: string;
  description: string;
  icon: string;
  theme_color: string;
  layout_type: string;
  is_default?: boolean;
};

export type UpdateFeedRequest = CreateFeedRequest;

export type CatalogSource = {
  id: string;
  provider: string;
  title: string;
  description: string;
  page_url: string;
  feed_url?: string;
  tags: string[];
};

export type Topic = {
  id: string;
  title: string;
  description: string;
  sources: CatalogSource[];
};

export type ConnectedCatalogSource = {
  catalog_source_id: string;
  source_id: string;
  feed_source_id: string;
  title: string;
  page_url: string;
  feed_url: string;
  created_at: string;
};

export type ConnectCatalogSourcesResponse = {
  connected: ConnectedCatalogSource[];
};

export type Category = {
  id: string;
  name: string;
  slug: string;
  description: string;
  created_at: string;
};

export type Item = {
  id: string;
  source_id: string;
  source_name: string;
  title: string;
  url: string;
  excerpt: string;
  image_url: string;
  author: string;
  published_at: string;
  tags: string[];
  categories: string[];
  score: number;
  is_saved: boolean;
  storage_mode?: "server" | "local";
  feed_id?: string;
  category_slugs?: string[];
  search_text?: string;
  cached_at?: string;
};

export type ReaderItem = Item & {
  content_html: string;
  reader_html: string;
  has_full_content: boolean;
};

export type PreviewItem = Item & {
  category_slugs: string[];
  search_text: string;
};

export type PreviewItemsResponse = {
  source_id: string;
  feed_url: string;
  items: PreviewItem[];
};

export type FeedItemsResponse = {
  items: Item[];
  mode: "today" | "archive" | "all";
  next_cursor?: string;
};

export type SavedItemsResponse = {
  items: Item[];
};

export type SearchItemsResponse = {
  items: Item[];
  query: string;
};

export type RefreshResult = {
  sources: {
    source_id: string;
    feed_url: string;
    items_found: number;
    items_created: number;
    items_skipped: number;
    skipped: boolean;
    reason?: string;
    error?: string;
    last_fetched_at?: string;
  }[];
  items_found: number;
  items_created: number;
  items_skipped: number;
  errors?: string[];
};
