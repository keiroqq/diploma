INSERT INTO categories (name, slug, description) VALUES
  ('Мир', 'world', 'Международная повестка, дипломатия и события за рубежом.'),
  ('Общество', 'society', 'Общество, социальные процессы, городская и гражданская повестка.'),
  ('Происшествия', 'accidents', 'Происшествия, инциденты, ЧП и оперативная хроника.'),
  ('Культура', 'culture', 'Культура, искусство, медиа и культурная индустрия.'),
  ('Потребительский рынок', 'market', 'Розница, потребительский спрос, товары и услуги.'),
  ('Телекоммуникации', 'telecom', 'Связь, телекоммуникации, операторы и цифровая инфраструктура.'),
  ('Регионы', 'regions', 'Региональная повестка городов и субъектов.')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tag_aliases (category_id, provider, raw_tag, raw_tag_slug)
SELECT c.id, 'any', alias.raw_tag, alias.raw_tag_slug
FROM categories c
JOIN (
  VALUES
    ('world', 'world', 'world'),
    ('world', 'мир', 'mir'),
    ('world', 'в мире', 'v-mire'),
    ('society', 'society', 'society'),
    ('society', 'общество', 'obshchestvo'),
    ('accidents', 'accidents', 'accidents'),
    ('accidents', 'происшествия', 'proisshestviia'),
    ('accidents', 'происшествия', 'proisshestviya'),
    ('accidents', 'чп', 'chp'),
    ('culture', 'culture', 'culture'),
    ('culture', 'культура', 'kultura'),
    ('market', 'market', 'market'),
    ('market', 'потребительский рынок', 'potrebitelskii-rynok'),
    ('market', 'потребительский рынок', 'potrebitelskij-rynok'),
    ('telecom', 'telecom', 'telecom'),
    ('telecom', 'телекоммуникации', 'telekommunikatsii'),
    ('technology', 'hi-tech', 'hi-tech'),
    ('technology', 'hitech', 'hitech'),
    ('lifestyle', 'стиль', 'stil'),
    ('regions', 'regions', 'regions'),
    ('regions', 'регионы', 'regiony'),
    ('regions', 'санкт-петербург', 'sankt-peterburg'),
    ('regions', 'екатеринбург', 'ekaterinburg'),
    ('regions', 'новосибирск', 'novosibirsk'),
    ('regions', 'самара', 'samara'),
    ('regions', 'казань', 'kazan'),
    ('regions', 'краснодар', 'krasnodar')
) AS alias(category_slug, raw_tag, raw_tag_slug) ON alias.category_slug = c.slug
ON CONFLICT (provider, raw_tag_slug) DO NOTHING;

INSERT INTO feed_item_categories (item_id, category_id)
SELECT DISTINCT fit.item_id, ta.category_id
FROM feed_item_tags fit
JOIN tags t ON t.id = fit.tag_id
JOIN tag_aliases ta ON ta.raw_tag_slug = t.slug
WHERE ta.provider IN ('any', 'kommersant')
ON CONFLICT DO NOTHING;

WITH source_categories(feed_url, category_slug) AS (
  VALUES
    ('https://www.kommersant.ru/RSS/section-politics.xml', 'politics'),
    ('https://www.kommersant.ru/RSS/section-economics.xml', 'economics'),
    ('https://www.kommersant.ru/RSS/section-business.xml', 'business'),
    ('https://www.kommersant.ru/RSS/section-telecom.xml', 'telecom'),
    ('https://www.kommersant.ru/RSS/section-telecom.xml', 'technology'),
    ('https://www.kommersant.ru/RSS/section-market.xml', 'market'),
    ('https://www.kommersant.ru/RSS/section-market.xml', 'business'),
    ('https://www.kommersant.ru/RSS/section-world.xml', 'world'),
    ('https://www.kommersant.ru/RSS/section-world.xml', 'politics'),
    ('https://www.kommersant.ru/RSS/section-accidents.xml', 'accidents'),
    ('https://www.kommersant.ru/RSS/section-society.xml', 'society'),
    ('https://www.kommersant.ru/RSS/section-culture.xml', 'culture'),
    ('https://www.kommersant.ru/RSS/section-sport.xml', 'sports'),
    ('https://www.kommersant.ru/RSS/section-auto.xml', 'auto'),
    ('https://www.kommersant.ru/RSS/section-hitech.xml', 'technology'),
    ('https://www.kommersant.ru/RSS/section-style.xml', 'lifestyle'),
    ('https://www.kommersant.ru/rss/regions/piter_all.xml', 'regions'),
    ('https://www.kommersant.ru/rss/regions/ekaterinburg_all.xml', 'regions'),
    ('https://www.kommersant.ru/rss/regions/novosibirsk_all.xml', 'regions'),
    ('https://www.kommersant.ru/rss/regions/samara_all.xml', 'regions'),
    ('https://www.kommersant.ru/rss/regions/kazan_all.xml', 'regions'),
    ('https://www.kommersant.ru/rss/regions/krasnodar_all.xml', 'regions')
)
INSERT INTO feed_item_categories (item_id, category_id)
SELECT DISTINCT fi.id, c.id
FROM feed_items fi
JOIN sources s ON s.id = fi.source_id
JOIN source_categories sc
  ON trim(trailing '/' from s.feed_url) = sc.feed_url
  OR trim(trailing '/' from s.url) = sc.feed_url
JOIN categories c ON c.slug = sc.category_slug
ON CONFLICT DO NOTHING;
