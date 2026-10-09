-- Durable, privacy-preserving guest analytics and abuse reports.
-- visitor_hash is a one-way application hash; raw IPs and visitor identifiers
-- are never persisted. Redis may be used as a short-lived accelerator later,
-- but PostgreSQL remains the durable source for these public actions.
CREATE TABLE IF NOT EXISTS guest_offer_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('view', 'whatsapp')),
    visitor_hash TEXT NOT NULL,
    bucket_start TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (offer_id, event_type, visitor_hash, bucket_start)
);

CREATE INDEX IF NOT EXISTS idx_guest_offer_events_offer_time
    ON guest_offer_events (offer_id, event_type, created_at DESC);

CREATE TABLE IF NOT EXISTS guest_offer_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id TEXT NOT NULL,
    visitor_hash TEXT NOT NULL,
    reason TEXT NOT NULL CHECK (reason IN ('misleading', 'expired', 'wrong_price', 'inappropriate', 'other')),
    message TEXT NOT NULL DEFAULT '' CHECK (char_length(message) <= 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (offer_id, visitor_hash)
);

CREATE INDEX IF NOT EXISTS idx_guest_offer_reports_visitor_time
    ON guest_offer_reports (visitor_hash, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_guest_offer_reports_review_signal
    ON guest_offer_reports (offer_id, created_at DESC);

CREATE OR REPLACE VIEW guest_offer_report_rollup AS
SELECT
    offer_id,
    count(*)::INTEGER AS report_count,
    count(DISTINCT visitor_hash)::INTEGER AS distinct_visitor_count,
    max(created_at) AS last_reported_at,
    (count(DISTINCT visitor_hash) >= 3) AS needs_review
FROM guest_offer_reports
GROUP BY offer_id;
