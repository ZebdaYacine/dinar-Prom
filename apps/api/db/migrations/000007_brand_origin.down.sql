ALTER TABLE brands DROP CONSTRAINT IF EXISTS brands_origin_check;
DROP INDEX IF EXISTS idx_brands_origin;
ALTER TABLE brands DROP COLUMN IF EXISTS origin;
