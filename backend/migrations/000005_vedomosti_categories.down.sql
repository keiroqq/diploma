DELETE FROM categories
WHERE slug IN (
  'business',
  'economics',
  'finance',
  'opinion',
  'politics',
  'technology',
  'realty',
  'auto',
  'lifestyle'
);
