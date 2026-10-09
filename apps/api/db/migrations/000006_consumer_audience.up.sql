-- Audience targeting for catalogue products. The range is inclusive and
-- describes the intended consumer, not a restriction on who may view an offer.
ALTER TABLE demo_products
    ADD COLUMN IF NOT EXISTS audience_gender TEXT NOT NULL DEFAULT 'unisex',
    ADD COLUMN IF NOT EXISTS age_min INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS age_max INTEGER NOT NULL DEFAULT 120;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'demo_products_audience_gender_check') THEN
        ALTER TABLE demo_products ADD CONSTRAINT demo_products_audience_gender_check CHECK (audience_gender IN ('men', 'women', 'unisex'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'demo_products_age_min_check') THEN
        ALTER TABLE demo_products ADD CONSTRAINT demo_products_age_min_check CHECK (age_min BETWEEN 0 AND 120);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'demo_products_age_max_check') THEN
        ALTER TABLE demo_products ADD CONSTRAINT demo_products_age_max_check CHECK (age_max BETWEEN 0 AND 120 AND age_max >= age_min);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_demo_products_audience ON demo_products (audience_gender, age_min, age_max);
