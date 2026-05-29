DELETE FROM tag_aliases
WHERE provider = 'any'
  AND raw_tag_slug IN ('hi-tech', 'hitech', 'stil');

DELETE FROM categories
WHERE slug IN (
  'world',
  'society',
  'accidents',
  'culture',
  'market',
  'telecom',
  'regions'
);
