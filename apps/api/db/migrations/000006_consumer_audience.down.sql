DROP INDEX IF EXISTS idx_demo_products_audience;
ALTER TABLE demo_products
    DROP CONSTRAINT IF EXISTS demo_products_age_max_check,
    DROP CONSTRAINT IF EXISTS demo_products_age_min_check,
    DROP CONSTRAINT IF EXISTS demo_products_audience_gender_check,
    DROP COLUMN IF EXISTS age_max,
    DROP COLUMN IF EXISTS age_min,
    DROP COLUMN IF EXISTS audience_gender;
