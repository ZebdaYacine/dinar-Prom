ALTER TABLE users ADD COLUMN IF NOT EXISTS google_subject text;
CREATE UNIQUE INDEX IF NOT EXISTS users_google_subject_idx ON users (google_subject) WHERE google_subject IS NOT NULL;

CREATE TABLE IF NOT EXISTS auth_policies (
  role text NOT NULL,
  resource text NOT NULL,
  action text NOT NULL,
  PRIMARY KEY (role, resource, action)
);

INSERT INTO auth_policies (role, resource, action) VALUES
  ('merchant', 'merchant_identity', 'write'),
  ('admin', 'merchant_identity', 'write'),
  ('admin', 'merchant_documents', 'read')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS notification_outbox (
  id uuid PRIMARY KEY,
  kind text NOT NULL CHECK (kind = 'merchant_verification_status'),
  recipient text NOT NULL,
  payload jsonb NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'sent', 'failed')),
  attempts integer NOT NULL DEFAULT 0,
  available_at timestamptz NOT NULL DEFAULT now(),
  last_error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  sent_at timestamptz
);
CREATE INDEX IF NOT EXISTS notification_outbox_pending_idx
  ON notification_outbox (available_at, created_at)
  WHERE status IN ('pending', 'failed');

CREATE TABLE IF NOT EXISTS audit_log (
  id uuid PRIMARY KEY,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id text NOT NULL,
  request_id text NOT NULL DEFAULT '',
  ip_address inet,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS audit_log_resource_idx ON audit_log (resource_type, resource_id, created_at DESC);
