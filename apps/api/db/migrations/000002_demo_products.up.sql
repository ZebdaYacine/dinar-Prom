CREATE TABLE IF NOT EXISTS demo_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT NOT NULL UNIQUE,
    title_ar TEXT NOT NULL,
    title_fr TEXT NOT NULL,
    description_ar TEXT NOT NULL,
    description_fr TEXT NOT NULL,
    category_slug TEXT NOT NULL,
    subcategory_slug TEXT NOT NULL,
    wilaya_code TEXT NOT NULL,
    city_name TEXT NOT NULL,
    city_name_ar TEXT NOT NULL,
    source_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'approved' CHECK (status IN ('draft', 'approved', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS demo_product_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES demo_products(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL UNIQUE,
    bucket_name TEXT NOT NULL DEFAULT 'dinar-public',
    source_url TEXT NOT NULL,
    public_url TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'image/webp',
    byte_size BIGINT NOT NULL DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 1 CHECK (position > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_demo_products_category ON demo_products (category_slug, subcategory_slug);
CREATE INDEX IF NOT EXISTS idx_demo_products_location ON demo_products (wilaya_code, city_name);
CREATE INDEX IF NOT EXISTS idx_demo_product_media_product ON demo_product_media (product_id, position);
