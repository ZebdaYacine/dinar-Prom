DELETE FROM auth_policies
WHERE (role, resource, action) IN (
  ('merchant', 'merchant_offers', 'write'),
  ('admin', 'merchant_offers', 'write'),
  ('admin', 'offer_review', 'write')
);

DELETE FROM notification_outbox WHERE kind = 'offer_status';
ALTER TABLE notification_outbox DROP CONSTRAINT IF EXISTS notification_outbox_kind_check;
ALTER TABLE notification_outbox ADD CONSTRAINT notification_outbox_kind_check
  CHECK (kind = 'merchant_verification_status');

DROP VIEW IF EXISTS merchant_offer_stats_daily;
DROP VIEW IF EXISTS public_offers;
DROP TABLE IF EXISTS offer_status_history;
DROP TABLE IF EXISTS offer_media;
DROP TABLE IF EXISTS offers;
DROP TABLE IF EXISTS app_settings;
