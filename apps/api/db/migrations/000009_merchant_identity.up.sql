CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY, email text NOT NULL UNIQUE, name text NOT NULL DEFAULT '', first_name text NOT NULL DEFAULT '', last_name text NOT NULL DEFAULT '', picture_url text NOT NULL DEFAULT '',
  email_verified boolean NOT NULL DEFAULT false, role text NOT NULL DEFAULT 'merchant' CHECK (role IN ('user','merchant','admin')),
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS refresh_tokens (
  token_hash text PRIMARY KEY, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, family_id text NOT NULL,
  jti text NOT NULL UNIQUE, expires_at timestamptz NOT NULL, revoked_at timestamptz, replaced_by_hash text,
  user_agent text NOT NULL DEFAULT '', ip_hash text NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS refresh_tokens_family_idx ON refresh_tokens(family_id);
CREATE TABLE IF NOT EXISTS merchant_profiles (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, phone_e164 text NOT NULL, wilaya_id text NOT NULL,
  address text NOT NULL, updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS merchant_shops (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, name text NOT NULL, slug text NOT NULL UNIQUE,
  category_id text NOT NULL, description text NOT NULL DEFAULT '', wilaya_id text NOT NULL, city text NOT NULL DEFAULT '',
  address text NOT NULL DEFAULT '', phone text NOT NULL DEFAULT '', whatsapp text NOT NULL DEFAULT '', register_number text NOT NULL,
  register_expires_on date NOT NULL, slug_edits smallint NOT NULL DEFAULT 0 CHECK (slug_edits BETWEEN 0 AND 1), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS merchant_documents (
  id uuid PRIMARY KEY, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  document_type text NOT NULL CHECK (document_type IN ('id_front','id_back','commercial_register')),
  content_type text NOT NULL, byte_size bigint NOT NULL CHECK (byte_size > 0 AND byte_size <= 8388608), sha256 text NOT NULL,
  storage_key text NOT NULL UNIQUE, created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(user_id, document_type)
);
CREATE TABLE IF NOT EXISTS merchant_verifications (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'incomplete' CHECK (status IN ('incomplete','pending','approved','rejected','suspended')),
  consent_accepted_at timestamptz, rejection_reason text NOT NULL DEFAULT '', rejection_document text NOT NULL DEFAULT '', updated_at timestamptz NOT NULL DEFAULT now()
);
