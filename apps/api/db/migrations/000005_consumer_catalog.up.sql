-- Consumer catalogue mirror used by merchant/admin workflows and reporting.
-- The public guest API embeds the same JSON so the local demo remains
-- deterministic even when the database is unavailable during startup.
CREATE TABLE IF NOT EXISTS catalog_categories (
    slug TEXT PRIMARY KEY,
    name_ar TEXT NOT NULL,
    name_fr TEXT NOT NULL,
    icon TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS catalog_subcategories (
    slug TEXT PRIMARY KEY,
    category_slug TEXT NOT NULL REFERENCES catalog_categories(slug) ON DELETE CASCADE,
    name_ar TEXT NOT NULL,
    name_fr TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_catalog_subcategories_category
    ON catalog_subcategories (category_slug, sort_order);
