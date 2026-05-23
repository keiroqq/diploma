CREATE TABLE categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL UNIQUE,
  slug varchar(140) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tag_aliases (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  provider varchar(64) NOT NULL DEFAULT 'any',
  raw_tag varchar(160) NOT NULL,
  raw_tag_slug varchar(180) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(provider, raw_tag_slug)
);

CREATE INDEX idx_tag_aliases_category_id ON tag_aliases(category_id);
CREATE INDEX idx_tag_aliases_raw_tag_slug ON tag_aliases(raw_tag_slug);

CREATE TABLE feed_item_categories (
  item_id uuid NOT NULL REFERENCES feed_items(id) ON DELETE CASCADE,
  category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (item_id, category_id)
);

CREATE INDEX idx_feed_item_categories_category_id ON feed_item_categories(category_id);

INSERT INTO categories (name, slug, description) VALUES
  ('Backend', 'backend', 'Серверная разработка, API, языки backend-разработки.'),
  ('Frontend', 'frontend', 'Клиентская веб-разработка, JavaScript, TypeScript и UI.'),
  ('Mobile', 'mobile', 'Мобильная разработка под Android, iOS и кроссплатформенные стеки.'),
  ('GameDev', 'gamedev', 'Разработка игр, игровые движки и индустрия игр.'),
  ('DevOps', 'devops', 'DevOps, CI/CD, контейнеризация и эксплуатация сервисов.'),
  ('Databases', 'databases', 'Базы данных, SQL, хранилища и обработка данных.'),
  ('Security', 'security', 'Информационная безопасность, защита, пентест и криптография.'),
  ('AI', 'ai', 'Искусственный интеллект, машинное обучение, нейросети и LLM.'),
  ('Design', 'design', 'UI/UX, дизайн интерфейсов и продуктовый дизайн.'),
  ('Management', 'management', 'Менеджмент, управление командами, продуктами и проектами.'),
  ('Marketing', 'marketing', 'Маркетинг, продвижение и рост цифровых продуктов.'),
  ('Science', 'science', 'Научпоп, космос, физика и исследования.'),
  ('Hardware', 'hardware', 'Железо, электроника, embedded и DIY.')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tag_aliases (category_id, provider, raw_tag, raw_tag_slug)
SELECT c.id, 'any', alias.raw_tag, alias.raw_tag_slug
FROM categories c
JOIN (
  VALUES
    ('backend', 'backend', 'backend'),
    ('backend', 'бэкенд', 'bekend'),
    ('backend', 'go', 'go'),
    ('backend', 'golang', 'golang'),
    ('backend', 'java', 'java'),
    ('backend', 'python', 'python'),
    ('backend', 'api', 'api'),
    ('backend', 'rest api', 'rest-api'),
    ('backend', 'микросервисы', 'mikroservisy'),
    ('frontend', 'frontend', 'frontend'),
    ('frontend', 'фронтенд', 'frontend'),
    ('frontend', 'javascript', 'javascript'),
    ('frontend', 'typescript', 'typescript'),
    ('frontend', 'react', 'react'),
    ('frontend', 'vue', 'vue'),
    ('frontend', 'ui', 'ui'),
    ('mobile', 'mobile', 'mobile'),
    ('mobile', 'android', 'android'),
    ('mobile', 'ios', 'ios'),
    ('mobile', 'flutter', 'flutter'),
    ('mobile', 'react native', 'react-native'),
    ('gamedev', 'gamedev', 'gamedev'),
    ('gamedev', 'game development', 'game-development'),
    ('gamedev', 'игры', 'igry'),
    ('gamedev', 'unity', 'unity'),
    ('gamedev', 'unreal engine', 'unreal-engine'),
    ('devops', 'devops', 'devops'),
    ('devops', 'docker', 'docker'),
    ('devops', 'kubernetes', 'kubernetes'),
    ('devops', 'ci/cd', 'ci-cd'),
    ('devops', 'linux', 'linux'),
    ('devops', 'администрирование', 'administrirovanie'),
    ('databases', 'databases', 'databases'),
    ('databases', 'базы данных', 'bazy-dannyh'),
    ('databases', 'postgresql', 'postgresql'),
    ('databases', 'postgres', 'postgres'),
    ('databases', 'sql', 'sql'),
    ('databases', 'mysql', 'mysql'),
    ('databases', 'redis', 'redis'),
    ('security', 'security', 'security'),
    ('security', 'infosec', 'infosec'),
    ('security', 'information security', 'information-security'),
    ('security', 'информационная безопасность', 'informacionnaya-bezopasnost'),
    ('security', 'иб', 'ib'),
    ('security', 'pentest', 'pentest'),
    ('security', 'пентест', 'pentest'),
    ('security', 'криптография', 'kriptografiya'),
    ('ai', 'ai', 'ai'),
    ('ai', 'artificial intelligence', 'artificial-intelligence'),
    ('ai', 'ии', 'ii'),
    ('ai', 'искусственный интеллект', 'iskusstvennyi-intellekt'),
    ('ai', 'ml', 'ml'),
    ('ai', 'machine learning', 'machine-learning'),
    ('ai', 'машинное обучение', 'mashinnoe-obuchenie'),
    ('ai', 'llm', 'llm'),
    ('ai', 'нейросети', 'neiroseti'),
    ('ai', 'rag', 'rag'),
    ('design', 'design', 'design'),
    ('design', 'дизайн', 'dizain'),
    ('design', 'ux', 'ux'),
    ('design', 'ui/ux', 'ui-ux'),
    ('management', 'management', 'management'),
    ('management', 'менеджмент', 'menedzhment'),
    ('management', 'product management', 'product-management'),
    ('management', 'продакт-менеджмент', 'prodakt-menedzhment'),
    ('marketing', 'marketing', 'marketing'),
    ('marketing', 'маркетинг', 'marketing'),
    ('marketing', 'growth', 'growth'),
    ('science', 'science', 'science'),
    ('science', 'научпоп', 'nauchpop'),
    ('science', 'космос', 'kosmos'),
    ('science', 'physics', 'physics'),
    ('hardware', 'hardware', 'hardware'),
    ('hardware', 'железо', 'zhelezo'),
    ('hardware', 'электроника', 'elektronika'),
    ('hardware', 'embedded', 'embedded'),
    ('hardware', 'diy', 'diy')
) AS alias(category_slug, raw_tag, raw_tag_slug) ON alias.category_slug = c.slug
ON CONFLICT (provider, raw_tag_slug) DO NOTHING;

INSERT INTO feed_item_categories (item_id, category_id)
SELECT DISTINCT fit.item_id, ta.category_id
FROM feed_item_tags fit
JOIN tags t ON t.id = fit.tag_id
JOIN tag_aliases ta ON ta.raw_tag_slug = t.slug
ON CONFLICT DO NOTHING;
