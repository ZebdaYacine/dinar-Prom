-- Brand taxonomy is scoped to a category/subcategory. Product rows keep the
-- stable brand slug so brand filtering remains cheap and indexable.
CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    name_ar TEXT NOT NULL,
    name_fr TEXT NOT NULL,
    logo_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subcategory_brands (
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    category_slug TEXT NOT NULL,
    subcategory_slug TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (brand_id, category_slug, subcategory_slug)
);

ALTER TABLE demo_products ADD COLUMN IF NOT EXISTS brand_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_demo_products_brand
    ON demo_products (category_slug, subcategory_slug, brand_slug);
CREATE INDEX IF NOT EXISTS idx_subcategory_brands_lookup
    ON subcategory_brands (category_slug, subcategory_slug, sort_order);
