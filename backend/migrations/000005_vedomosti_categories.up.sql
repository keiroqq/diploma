INSERT INTO categories (name, slug, description) VALUES
  ('Бизнес', 'business', 'Компании, рынки, сделки и деловая среда.'),
  ('Экономика', 'economics', 'Макроэкономика, государственная экономическая политика и рынки.'),
  ('Финансы', 'finance', 'Финансовые рынки, банки, инвестиции и деньги.'),
  ('Мнения', 'opinion', 'Колонки, мнения, аналитика и авторские материалы.'),
  ('Политика', 'politics', 'Политика, государство, регулирование и общественные решения.'),
  ('Технологии', 'technology', 'Технологии, цифровые сервисы, телеком и инновации.'),
  ('Недвижимость', 'realty', 'Жилая и коммерческая недвижимость, девелопмент и рынок жилья.'),
  ('Авто', 'auto', 'Автомобильный рынок, производители, продажи и транспорт.'),
  ('Стиль жизни', 'lifestyle', 'Стиль жизни, культура, общество и потребительские практики.')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tag_aliases (category_id, provider, raw_tag, raw_tag_slug)
SELECT c.id, 'any', alias.raw_tag, alias.raw_tag_slug
FROM categories c
JOIN (
  VALUES
    ('business', 'business', 'business'),
    ('business', 'бизнес', 'biznes'),
    ('economics', 'economics', 'economics'),
    ('economics', 'экономика', 'ekonomika'),
    ('finance', 'finance', 'finance'),
    ('finance', 'финансы', 'finansy'),
    ('opinion', 'opinion', 'opinion'),
    ('opinion', 'мнения', 'mneniia'),
    ('opinion', 'мнения', 'mneniya'),
    ('opinion', 'колонки', 'kolonki'),
    ('politics', 'politics', 'politics'),
    ('politics', 'политика', 'politika'),
    ('technology', 'technology', 'technology'),
    ('technology', 'технологии', 'tekhnologii'),
    ('realty', 'realty', 'realty'),
    ('realty', 'недвижимость', 'nedvizhimost'),
    ('auto', 'auto', 'auto'),
    ('auto', 'авто', 'avto'),
    ('auto', 'автомобили', 'avtomobili'),
    ('lifestyle', 'lifestyle', 'lifestyle'),
    ('lifestyle', 'стиль жизни', 'stil-zhizni')
) AS alias(category_slug, raw_tag, raw_tag_slug) ON alias.category_slug = c.slug
ON CONFLICT (provider, raw_tag_slug) DO NOTHING;

INSERT INTO feed_item_categories (item_id, category_id)
SELECT DISTINCT fit.item_id, ta.category_id
FROM feed_item_tags fit
JOIN tags t ON t.id = fit.tag_id
JOIN tag_aliases ta ON ta.raw_tag_slug = t.slug
WHERE ta.provider IN ('any', 'vedomosti')
ON CONFLICT DO NOTHING;

WITH source_categories(feed_url, category_slug) AS (
  VALUES
    ('https://www.vedomosti.ru/rss/rubric/business.xml', 'business'),
    ('https://www.vedomosti.ru/rss/rubric/economics.xml', 'economics'),
    ('https://www.vedomosti.ru/rss/rubric/finance.xml', 'finance'),
    ('https://www.vedomosti.ru/rss/rubric/opinion.xml', 'opinion'),
    ('https://www.vedomosti.ru/rss/rubric/politics.xml', 'politics'),
    ('https://www.vedomosti.ru/rss/rubric/technology.xml', 'technology'),
    ('https://www.vedomosti.ru/rss/rubric/realty.xml', 'realty'),
    ('https://www.vedomosti.ru/rss/rubric/auto.xml', 'auto'),
    ('https://www.vedomosti.ru/rss/rubric/management.xml', 'management'),
    ('https://www.vedomosti.ru/rss/rubric/lifestyle.xml', 'lifestyle')
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
