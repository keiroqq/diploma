INSERT INTO categories (name, slug, description) VALUES
  ('Спорт', 'sports', 'Спортивные новости, соревнования и индустрия спорта.'),
  ('Футбол', 'football', 'Футбол, клубы, сборные, турниры и трансферы.'),
  ('Хоккей', 'hockey', 'Хоккей, лиги, клубы и турниры.'),
  ('Баскетбол', 'basketball', 'Баскетбол, клубы, лиги и турниры.'),
  ('Формула-1', 'formula-1', 'Формула-1, автоспорт, команды и гонки.'),
  ('Теннис', 'tennis', 'Теннис, турниры, игроки и рейтинги.'),
  ('Единоборства', 'combat-sports', 'Бокс, MMA, UFC и другие единоборства.'),
  ('Волейбол', 'volleyball', 'Волейбол, клубы, сборные и турниры.'),
  ('Легкая атлетика', 'athletics', 'Легкая атлетика, соревнования и спортсмены.'),
  ('Велоспорт', 'cycling', 'Велоспорт, гонки, команды и спортсмены.'),
  ('Водные виды', 'water-sports', 'Плавание и другие водные виды спорта.'),
  ('Шахматы', 'chess', 'Шахматы, турниры, партии и игроки.'),
  ('Футзал', 'futsal', 'Футзал, клубы, сборные и турниры.'),
  ('Гандбол', 'handball', 'Гандбол, клубы, сборные и турниры.'),
  ('Гимнастика', 'gymnastics', 'Гимнастика, соревнования и спортсмены.'),
  ('Фигурное катание', 'figure-skating', 'Фигурное катание, турниры и спортсмены.'),
  ('Биатлон', 'biathlon', 'Биатлон, гонки, сборные и спортсмены.')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tag_aliases (category_id, provider, raw_tag, raw_tag_slug)
SELECT c.id, 'any', alias.raw_tag, alias.raw_tag_slug
FROM categories c
JOIN (
  VALUES
    ('sports', 'sports', 'sports'),
    ('sports', 'спорт', 'sport'),
    ('football', 'football', 'football'),
    ('football', 'футбол', 'futbol'),
    ('hockey', 'hockey', 'hockey'),
    ('hockey', 'хоккей', 'khokkei'),
    ('hockey', 'хоккей', 'hokkei'),
    ('basketball', 'basketball', 'basketball'),
    ('basketball', 'баскетбол', 'basketbol'),
    ('formula-1', 'formula-1', 'formula-1'),
    ('formula-1', 'f1', 'f1'),
    ('formula-1', 'формула-1', 'formula-1'),
    ('formula-1', 'автоспорт', 'avtosport'),
    ('tennis', 'tennis', 'tennis'),
    ('tennis', 'теннис', 'tennis'),
    ('combat-sports', 'combat sports', 'combat-sports'),
    ('combat-sports', 'boxing', 'boxing'),
    ('combat-sports', 'mma', 'mma'),
    ('combat-sports', 'ufc', 'ufc'),
    ('combat-sports', 'бокс', 'boks'),
    ('combat-sports', 'единоборства', 'edinoborstva'),
    ('volleyball', 'volleyball', 'volleyball'),
    ('volleyball', 'волейбол', 'voleibol'),
    ('athletics', 'athletics', 'athletics'),
    ('athletics', 'легкая атлетика', 'legkaia-atletika'),
    ('athletics', 'легкая атлетика', 'legkaya-atletika'),
    ('cycling', 'cycling', 'cycling'),
    ('cycling', 'велоспорт', 'velosport'),
    ('water-sports', 'water sports', 'water-sports'),
    ('water-sports', 'водные виды', 'vodnye-vidy'),
    ('water-sports', 'плавание', 'plavanie'),
    ('chess', 'chess', 'chess'),
    ('chess', 'шахматы', 'shakhmaty'),
    ('futsal', 'futsal', 'futsal'),
    ('futsal', 'футзал', 'futzal'),
    ('handball', 'handball', 'handball'),
    ('handball', 'гандбол', 'gandbol'),
    ('gymnastics', 'gymnastics', 'gymnastics'),
    ('gymnastics', 'гимнастика', 'gimnastika'),
    ('figure-skating', 'figure skating', 'figure-skating'),
    ('figure-skating', 'фигурное катание', 'figurnoe-katanie'),
    ('biathlon', 'biathlon', 'biathlon'),
    ('biathlon', 'биатлон', 'biatlon')
) AS alias(category_slug, raw_tag, raw_tag_slug) ON alias.category_slug = c.slug
ON CONFLICT (provider, raw_tag_slug) DO NOTHING;

INSERT INTO feed_item_categories (item_id, category_id)
SELECT DISTINCT fit.item_id, ta.category_id
FROM feed_item_tags fit
JOIN tags t ON t.id = fit.tag_id
JOIN tag_aliases ta ON ta.raw_tag_slug = t.slug
WHERE ta.provider IN ('any', 'sports')
ON CONFLICT DO NOTHING;

WITH source_categories(feed_url, category_slug) AS (
  VALUES
    ('https://www.sports.ru/rss/rubric/208.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/208.xml', 'football'),
    ('https://www.sports.ru/rss/rubric/209.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/209.xml', 'hockey'),
    ('https://www.sports.ru/rss/rubric/210.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/210.xml', 'basketball'),
    ('https://www.sports.ru/rss/rubric/211.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/211.xml', 'formula-1'),
    ('https://www.sports.ru/rss/rubric/212.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/212.xml', 'tennis'),
    ('https://www.sports.ru/rss/rubric/213.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/213.xml', 'combat-sports'),
    ('https://www.sports.ru/rss/rubric/214.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/214.xml', 'volleyball'),
    ('https://www.sports.ru/rss/rubric/215.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/215.xml', 'athletics'),
    ('https://www.sports.ru/rss/rubric/216.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/216.xml', 'cycling'),
    ('https://www.sports.ru/rss/rubric/217.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/217.xml', 'water-sports'),
    ('https://www.sports.ru/rss/rubric/218.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/218.xml', 'chess'),
    ('https://www.sports.ru/rss/rubric/219.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/219.xml', 'futsal'),
    ('https://www.sports.ru/rss/rubric/220.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/220.xml', 'handball'),
    ('https://www.sports.ru/rss/rubric/221.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/221.xml', 'gymnastics'),
    ('https://www.sports.ru/rss/rubric/223.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/223.xml', 'figure-skating'),
    ('https://www.sports.ru/rss/rubric/225.xml', 'sports'),
    ('https://www.sports.ru/rss/rubric/225.xml', 'biathlon')
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
