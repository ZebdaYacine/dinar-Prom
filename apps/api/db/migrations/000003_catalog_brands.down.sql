DROP INDEX IF EXISTS idx_subcategory_brands_lookup;
DROP INDEX IF EXISTS idx_demo_products_brand;
ALTER TABLE demo_products DROP COLUMN IF EXISTS brand_slug;
DROP TABLE IF EXISTS subcategory_brands;
DROP TABLE IF EXISTS brands;
