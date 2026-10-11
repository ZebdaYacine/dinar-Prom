CREATE TABLE IF NOT EXISTS app_settings (
  key text PRIMARY KEY,
  value jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO app_settings (key, value) VALUES
  ('max_active_offers', '20'::jsonb),
  ('max_images_per_offer', '10'::jsonb)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS offers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  merchant_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  title_ar text NOT NULL DEFAULT '',
  title_fr text NOT NULL DEFAULT '',
  description_ar text NOT NULL DEFAULT '',
  description_fr text NOT NULL DEFAULT '',
  category_slug text NOT NULL DEFAULT '',
  subcategory_slug text NOT NULL DEFAULT '',
  brand_slug text,
  price_before_dzd integer,
  price_after_dzd integer,
  discount_percent smallint NOT NULL DEFAULT 0 CHECK (discount_percent BETWEEN 0 AND 99),
  starts_at timestamptz,
  ends_at timestamptz,
  wilaya_code text NOT NULL DEFAULT '',
  city text NOT NULL DEFAULT '',
  address text NOT NULL DEFAULT '',
  whatsapp text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','pending','approved','rejected','paused','archived','removed')),
  rejection_reason text NOT NULL DEFAULT '',
  submitted_at timestamptz,
  approved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (price_before_dzd IS NULL OR price_before_dzd > 0),
  CHECK (price_after_dzd IS NULL OR price_after_dzd > 0),
  CHECK (starts_at IS NULL OR ends_at IS NULL OR (ends_at > starts_at AND ends_at <= starts_at + interval '180 days'))
);

CREATE INDEX IF NOT EXISTS offers_merchant_updated_idx ON offers (merchant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS offers_review_queue_idx ON offers (submitted_at, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS offers_public_filter_idx ON offers (category_slug, wilaya_code, ends_at) WHERE status = 'approved';

CREATE TABLE IF NOT EXISTS offer_media (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  offer_id uuid NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  object_key_400 text NOT NULL UNIQUE,
  object_key_800 text NOT NULL UNIQUE,
  object_key_1600 text NOT NULL UNIQUE,
  source_width integer NOT NULL CHECK (source_width >= 400),
  source_height integer NOT NULL CHECK (source_height >= 400),
  position smallint NOT NULL CHECK (position BETWEEN 1 AND 10),
  alt_text text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT offer_media_position_unique UNIQUE (offer_id, position) DEFERRABLE INITIALLY IMMEDIATE
);
CREATE INDEX IF NOT EXISTS offer_media_offer_idx ON offer_media (offer_id, position);

CREATE TABLE IF NOT EXISTS offer_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  offer_id uuid NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  from_status text,
  to_status text NOT NULL,
  reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS offer_status_history_offer_idx ON offer_status_history (offer_id, created_at DESC);

CREATE OR REPLACE VIEW public_offers AS
SELECT o.*
FROM offers o
JOIN merchant_verifications mv ON mv.user_id = o.merchant_id
WHERE o.status = 'approved'
  AND mv.status = 'approved'
  AND o.starts_at <= now()
  AND o.ends_at > now();

CREATE OR REPLACE VIEW merchant_offer_stats_daily AS
SELECT o.merchant_id,
       e.offer_id::uuid AS offer_id,
       (e.created_at AT TIME ZONE 'UTC')::date AS day,
       count(*) FILTER (WHERE e.event_type = 'view')::integer AS views,
       count(*) FILTER (WHERE e.event_type = 'whatsapp')::integer AS whatsapp_clicks
FROM guest_offer_events e
JOIN offers o ON o.id::text = e.offer_id
GROUP BY o.merchant_id, e.offer_id, (e.created_at AT TIME ZONE 'UTC')::date;

ALTER TABLE notification_outbox DROP CONSTRAINT IF EXISTS notification_outbox_kind_check;
ALTER TABLE notification_outbox ADD CONSTRAINT notification_outbox_kind_check
  CHECK (kind IN ('merchant_verification_status', 'offer_status'));

INSERT INTO auth_policies (role, resource, action) VALUES
  ('merchant', 'merchant_offers', 'write'),
  ('admin', 'merchant_offers', 'write'),
  ('admin', 'offer_review', 'write')
ON CONFLICT DO NOTHING;
