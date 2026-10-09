-- Keep the catalogue explicit about whether a brand is global or local.
ALTER TABLE brands ADD COLUMN IF NOT EXISTS origin TEXT NOT NULL DEFAULT 'global';
ALTER TABLE brands DROP CONSTRAINT IF EXISTS brands_origin_check;
ALTER TABLE brands ADD CONSTRAINT brands_origin_check CHECK (origin IN ('global', 'local'));
CREATE INDEX IF NOT EXISTS idx_brands_origin ON brands (origin);
