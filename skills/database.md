---
name: dinar-dz-database
description: PostgreSQL 16 schema, migrations, indexes, visibility rules, SQLBoiler generation, Casbin policies, Redis boundaries, search, payments, and transactional rules for Dinar DZ. Use this skill whenever creating or changing tables, migrations, SQL, SQLBoiler repositories, offer visibility, merchant verification, trial or subscription logic, Chargily payments, ads, notifications, audit logs, or Arabic/French search.
---

# Database Schema

**Engine:** PostgreSQL 16  
**ORM:** SQLBoiler v4 generated models and query modifiers  
**Migrations:** `golang-migrate` files in `apps/api/db/migrations/`  
**ID strategy:** UUID v4 (`gen_random_uuid()`) for business entities; identity `bigint` for append-only queues/logs  
**Soft deletes:** status/archived state for business entities; never hard-delete audit or payment history  
**Timestamps:** `TIMESTAMPTZ DEFAULT NOW()` in UTC  
**Cache boundary:** Redis stores cache, counters, short locks, and deduplication keys only; PostgreSQL remains the source of truth  

---

## Entity Relationship Diagram

```text
┌──────────┐       ┌────────────┐       ┌───────────────┐
│  users   │──────▶│ merchants  │──────▶│ merchant_docs │
│ role     │       │ status     │       │ private keys  │
└────┬─────┘       └─────┬──────┘       └───────────────┘
     │                    │
     │              ┌─────▼─────┐       ┌──────────────┐
     │              │  offers   │──────▶│ offer_media  │
     │              │ approval  │       │ public keys  │
     │              └─────┬─────┘       └──────────────┘
     │                    │
     │      ┌─────────────┼──────────────┐
     │      ▼             ▼              ▼
┌────▼─────┐       ┌────────────┐  ┌───────────────┐
│ refresh_tokens │  │ payments   │  │ offer_stats   │
└──────────┘       │ Chargily   │  │ daily        │
                    └─────┬──────┘  └───────────────┘
                          ▼
                    ┌──────────────┐
                    │ subscriptions│
                    └──────────────┘

┌──────────┐  ┌──────────┐  ┌─────────────┐  ┌─────────────┐
│ reports  │  │ ads      │  │ notifications│ │ jobs        │
└────┬─────┘  └────┬─────┘  └─────────────┘  └──────┬──────┘
     │              │                                 │
     └──────────────┴──────────────┬──────────────────┘
                                   ▼
                             ┌────────────┐
                             │ audit_logs │
                             └────────────┘

             ┌──────────────┐       ┌──────────────┐
             │ casbin_rule │       │ categories   │
             │ roles/policy│       │ + wilayas    │
             └──────────────┘       └──────────────┘
```

---

## Tables

### 1. `wilayas`

Algerian wilaya reference data.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `SMALLINT` | PK | | Official wilaya code |
| `name_ar` | `TEXT` | NOT NULL | | Arabic name |
| `name_fr` | `TEXT` | NOT NULL | | French name |
| `slug` | `TEXT` | UNIQUE, NOT NULL | | URL-safe identifier |

### 2. `categories`

Hierarchical offer categories.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `SMALLINT` | PK | identity | Category identifier |
| `parent_id` | `SMALLINT` | FK → `categories(id)` | NULL | Parent category |
| `slug` | `TEXT` | UNIQUE, NOT NULL | | URL-safe identifier |
| `name_ar` | `TEXT` | NOT NULL | | Arabic name |
| `name_fr` | `TEXT` | NOT NULL | | French name |
| `icon` | `TEXT` | | NULL | Local icon name |
| `sort_order` | `SMALLINT` | NOT NULL | `0` | Display order |
| `is_active` | `BOOLEAN` | NOT NULL | `TRUE` | Available to merchants |

**Indexes:** `categories_parent_idx(parent_id)`, unique `slug`.

### 3. `settings`

Administrative key/value settings. Values are JSONB so numeric and structured settings remain typed.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `key` | `TEXT` | PK | | Setting key |
| `value` | `JSONB` | NOT NULL | | Setting value |
| `updated_by` | `UUID` | FK → `users(id)` | NULL | Admin who changed it |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Last change |

**Seeded keys:** `trial_days`, `subscription_price_dzd`, `subscription_days`, `max_images_per_offer`, `max_active_offers`, `manual_payment_info`.

### 4. `users`

Google-authenticated accounts. Fine-grained administrator permissions are managed by Casbin; `role` remains the coarse account role.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | User identifier |
| `google_sub` | `TEXT` | UNIQUE, NOT NULL | | Google subject identifier |
| `email` | `CITEXT` | UNIQUE, NOT NULL | | Verified email |
| `full_name` | `TEXT` | NOT NULL | | Display name |
| `avatar_url` | `TEXT` | | NULL | Avatar URL |
| `role` | `user_role` | NOT NULL | `'merchant'` | Coarse account role |
| `is_active` | `BOOLEAN` | NOT NULL | `TRUE` | Account status |
| `last_login_at` | `TIMESTAMPTZ` | | NULL | Last successful login |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Indexes:** unique `google_sub`, unique `email`.

### 5. `refresh_tokens`

Rotated JWT refresh tokens. Only a hash is stored; raw refresh JWTs never enter
PostgreSQL. Access JWTs are short-lived and are not stored in this table.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Refresh-token record |
| `user_id` | `UUID` | FK → `users(id)` ON DELETE CASCADE | | Owner |
| `token_hash` | `BYTEA` | UNIQUE, NOT NULL | | SHA-256/HMAC hash of refresh JWT |
| `jti` | `UUID` | UNIQUE, NOT NULL | | JWT ID |
| `family_id` | `UUID` | NOT NULL | | Rotation/reuse family |
| `replaced_by_hash` | `BYTEA` | | NULL | Hash of replacement token |
| `user_agent` | `TEXT` | | NULL | Client user agent |
| `ip` | `INET` | | NULL | Request IP; redact in logs |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL | | Expiry |
| `revoked_at` | `TIMESTAMPTZ` | | NULL | Revocation time |
| `revoke_reason` | `TEXT` | | NULL | logout, reuse, incident, admin |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Indexes:** `refresh_tokens_user_idx(user_id)`, `refresh_tokens_family_idx(family_id)`,
`refresh_tokens_expires_idx(expires_at)`, unique `jti` and `token_hash`.

### 6. `merchants`

Merchant onboarding, shop identity, verification, trial, and access state.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Merchant identifier |
| `user_id` | `UUID` | UNIQUE, FK → `users(id)` | | Account owner |
| `full_name` | `TEXT` | | NULL | Legal/contact name |
| `phone_e164` | `TEXT` | CHECK Algeria format | NULL | Personal phone |
| `personal_wilaya_id` | `SMALLINT` | FK → `wilayas(id)` | NULL | Personal wilaya |
| `personal_address` | `TEXT` | | NULL | Personal address |
| `onboarding_step` | `SMALLINT` | NOT NULL | `1` | Current onboarding step |
| `shop_name` | `TEXT` | | NULL | Public shop name |
| `slug` | `TEXT` | UNIQUE | NULL | Public shop slug |
| `description` | `TEXT` | | NULL | Shop description |
| `logo_key` | `TEXT` | | NULL | Public storage key |
| `category_id` | `SMALLINT` | FK → `categories(id)` | NULL | Main category |
| `wilaya_id` | `SMALLINT` | FK → `wilayas(id)` | NULL | Shop wilaya |
| `commune` | `TEXT` | | NULL | Commune |
| `address` | `TEXT` | | NULL | Shop address |
| `whatsapp_e164` | `TEXT` | CHECK Algeria format | NULL | WhatsApp contact |
| `shop_phone_e164` | `TEXT` | | NULL | Shop phone |
| `facebook_url` | `TEXT` | | NULL | Facebook URL |
| `instagram_url` | `TEXT` | | NULL | Instagram URL |
| `website_url` | `TEXT` | | NULL | Website URL |
| `register_number` | `TEXT` | UNIQUE when present | NULL | Commercial register number |
| `register_expires_on` | `DATE` | | NULL | Register expiry |
| `verification_status` | `verification_status` | NOT NULL | `'incomplete'` | Review state |
| `rejection_reason` | `TEXT` | | NULL | Review reason |
| `consent_accepted_at` | `TIMESTAMPTZ` | | NULL | Privacy consent |
| `submitted_at` | `TIMESTAMPTZ` | | NULL | Submitted for review |
| `verified_at` | `TIMESTAMPTZ` | | NULL | Approval time |
| `verified_by` | `UUID` | FK → `users(id)` | NULL | Approving admin |
| `trial_started_at` | `TIMESTAMPTZ` | | NULL | One-time trial start |
| `trial_ends_at` | `TIMESTAMPTZ` | | NULL | Trial end |
| `access_until` | `TIMESTAMPTZ` | | NULL | Max trial/subscription end |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Last change |

**Constraints:** `trial_started_at` and `trial_ends_at` are both null or both non-null.

**Indexes:** register and phone uniqueness, `merchants_status_idx(verification_status, submitted_at)`, `merchants_access_idx(access_until)`.

### 7. `merchant_documents`

Private onboarding documents. `storage_key` must never be returned to a normal client.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Document identifier |
| `merchant_id` | `UUID` | FK → `merchants(id)` ON DELETE CASCADE | | Merchant |
| `type` | `document_type` | NOT NULL | | Identity/register type |
| `storage_key` | `TEXT` | NOT NULL | | Private bucket key |
| `mime_type` | `TEXT` | NOT NULL | | Validated MIME type |
| `size_bytes` | `INTEGER` | CHECK > 0 | | File size |
| `sha256` | `BYTEA` | NOT NULL | | File digest |
| `status` | `document_status` | NOT NULL | `'pending'` | Review state |
| `rejection_reason` | `TEXT` | | NULL | Review reason |
| `uploaded_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Upload time |
| `reviewed_at` | `TIMESTAMPTZ` | | NULL | Review time |
| `reviewed_by` | `UUID` | FK → `users(id)` | NULL | Reviewing admin |

**Constraint:** one current non-rejected document per `(merchant_id, type)`.

### 8. `offers`

Merchant discount offers. Public visibility is defined only by the `public_offers` view below.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Offer identifier |
| `merchant_id` | `UUID` | FK → `merchants(id)` ON DELETE RESTRICT | | Owner |
| `category_id` | `SMALLINT` | FK → `categories(id)` | | Category |
| `wilaya_id` | `SMALLINT` | FK → `wilayas(id)` | | Wilaya |
| `address` | `TEXT` | | NULL | Offer address |
| `whatsapp_e164` | `TEXT` | NOT NULL, CHECK | | Contact number |
| `title` | `TEXT` | CHECK 5–120 chars | | Offer title |
| `description` | `TEXT` | CHECK ≤ 2000 chars | `''` | Offer description |
| `price_before_dzd` | `INTEGER` | CHECK > 0 | NULL | Original price |
| `price_after_dzd` | `INTEGER` | CHECK > 0 | NULL | Discounted price |
| `discount_percent` | `SMALLINT` | CHECK 1–99 | | Discount percentage |
| `starts_at` | `TIMESTAMPTZ` | NOT NULL | | Start time |
| `ends_at` | `TIMESTAMPTZ` | NOT NULL | | End time |
| `status` | `offer_status` | NOT NULL | `'draft'` | Moderation state |
| `rejection_reason` | `TEXT` | | NULL | Review reason |
| `submitted_at` | `TIMESTAMPTZ` | | NULL | Submission time |
| `reviewed_at` | `TIMESTAMPTZ` | | NULL | Review time |
| `reviewed_by` | `UUID` | FK → `users(id)` | NULL | Reviewing admin |
| `published_at` | `TIMESTAMPTZ` | | NULL | First approval |
| `report_count` | `INTEGER` | NOT NULL | `0` | Number of reports |
| `needs_review` | `BOOLEAN` | NOT NULL | `FALSE` | Report threshold reached |
| `is_featured` | `BOOLEAN` | NOT NULL | `FALSE` | Featured flag |
| `featured_until` | `TIMESTAMPTZ` | | NULL | Feature expiry |
| `search_text` | `TEXT` | NOT NULL | `''` | Normalized search text |
| `search_vector` | `TSVECTOR` | GENERATED | | PostgreSQL search vector |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Last change |

**Constraints:** `ends_at > starts_at`; duration ≤ 180 days; both prices are null or both are present with `price_after_dzd < price_before_dzd`.

**Indexes:** merchant/status, visible dates, wilaya/category/date filters, pending review, GIN search vector, and trigram search text.

### 9. `offer_media`

Public processed offer images.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Media identifier |
| `offer_id` | `UUID` | FK → `offers(id)` ON DELETE CASCADE | | Offer |
| `kind` | `media_kind` | NOT NULL | `'image'` | Media type |
| `storage_key` | `TEXT` | NOT NULL | | Root key; variants `_400`, `_800`, `_1600` |
| `width` | `INTEGER` | | NULL | Original width |
| `height` | `INTEGER` | | NULL | Original height |
| `sort_order` | `SMALLINT` | NOT NULL | `0` | Display order |
| `alt_text` | `TEXT` | | NULL | Accessible alt text |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Rule:** maximum image count is enforced by a trigger using `settings.max_images_per_offer`.

### 10. `subscriptions`

Merchant access periods. The current access date is copied to `merchants.access_until` transactionally.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Subscription identifier |
| `merchant_id` | `UUID` | FK → `merchants(id)` ON DELETE RESTRICT | | Merchant |
| `status` | `subscription_status` | NOT NULL | `'active'` | Subscription state |
| `starts_at` | `TIMESTAMPTZ` | NOT NULL | | Access start |
| `ends_at` | `TIMESTAMPTZ` | NOT NULL | | Access end |
| `amount_dzd` | `INTEGER` | CHECK ≥ 0 | | Amount in DZD |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Constraint:** `ends_at > starts_at`.

### 11. `payments`

Chargily or manual payment records. Card data is never stored here.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Payment identifier |
| `merchant_id` | `UUID` | FK → `merchants(id)` ON DELETE RESTRICT | | Merchant |
| `subscription_id` | `UUID` | FK → `subscriptions(id)` | NULL | Created after approval |
| `method` | `payment_method` | NOT NULL | | `chargily` or `manual` |
| `status` | `payment_status` | NOT NULL | `'pending'` | Payment state |
| `amount_dzd` | `INTEGER` | CHECK > 0 | | Expected amount |
| `reference` | `TEXT` | UNIQUE, NOT NULL | | Internal reference |
| `provider_checkout_id` | `TEXT` | UNIQUE | NULL | Chargily checkout reference |
| `provider_payload` | `JSONB` | | NULL | Minimal non-sensitive provider metadata |
| `proof_key` | `TEXT` | | NULL | Private manual proof key |
| `proof_reference` | `TEXT` | | NULL | Manual transfer reference |
| `rejection_reason` | `TEXT` | | NULL | Review reason |
| `reviewed_by` | `UUID` | FK → `users(id)` | NULL | Reviewing admin |
| `reviewed_at` | `TIMESTAMPTZ` | | NULL | Review time |
| `paid_at` | `TIMESTAMPTZ` | | NULL | Payment confirmation time |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Rule:** `proof_key` is required for a manual payment before approval. Never store PAN, CVV, expiry, or cardholder secrets.

### 12. `webhook_events`

Idempotency ledger for provider webhooks.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `provider` | `TEXT` | PK part | | `chargily` |
| `event_id` | `TEXT` | PK part | | Provider event identifier |
| `event_type` | `TEXT` | NOT NULL | | Provider event type |
| `payload_hash` | `BYTEA` | NOT NULL | | Hash for audit/replay detection |
| `payload` | `JSONB` | | NULL | Minimum safe payload only |
| `received_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Receipt time |
| `processed_at` | `TIMESTAMPTZ` | | NULL | Successful handling time |

**Rule:** the `(provider, event_id)` primary key is authoritative. Redis replay locks are only an optimization.

### 13. `reports`

Anonymous or user-submitted offer reports.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Report identifier |
| `offer_id` | `UUID` | FK → `offers(id)` ON DELETE CASCADE | | Reported offer |
| `reason` | `report_reason` | NOT NULL | | Report category |
| `message` | `TEXT` | CHECK ≤ 500 chars | NULL | Additional details |
| `reporter_hash` | `BYTEA` | NOT NULL | | Salted hash; never raw IP |
| `status` | `report_status` | NOT NULL | `'open'` | Review state |
| `handled_by` | `UUID` | FK → `users(id)` | NULL | Admin handler |
| `handled_at` | `TIMESTAMPTZ` | | NULL | Resolution time |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Constraint:** one open report per `(offer_id, reporter_hash)`.

### 14. `ad_slots`

Known advertisement placements.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `code` | `TEXT` | PK | | `home_top`, `list_inline`, `detail_sidebar`, `footer_banner` |
| `name_ar` | `TEXT` | NOT NULL | | Arabic label |
| `name_fr` | `TEXT` | NOT NULL | | French label |
| `width` | `SMALLINT` | NOT NULL | | Reserved width |
| `height` | `SMALLINT` | NOT NULL | | Reserved height |

### 15. `ads`

Targeted advertisements.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Ad identifier |
| `slot_code` | `TEXT` | FK → `ad_slots(code)` | | Placement |
| `advertiser` | `TEXT` | NOT NULL | | Advertiser name |
| `title` | `TEXT` | NOT NULL | | Ad title |
| `image_key` | `TEXT` | NOT NULL | | Public media key |
| `target_url` | `TEXT` | CHECK `https?://` | | Destination |
| `wilaya_id` | `SMALLINT` | FK → `wilayas(id)` | NULL | NULL means all wilayas |
| `category_id` | `SMALLINT` | FK → `categories(id)` | NULL | Optional category |
| `priority` | `SMALLINT` | NOT NULL | `0` | Selection priority |
| `status` | `ad_status` | NOT NULL | `'draft'` | Ad state |
| `starts_at` | `TIMESTAMPTZ` | NOT NULL | | Start time |
| `ends_at` | `TIMESTAMPTZ` | NOT NULL | | End time |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Constraint:** `ends_at > starts_at`. **Index:** active ads by slot and priority.

### 16. `ad_stats_daily`

Daily ad counters flushed from Redis.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `ad_id` | `UUID` | PK part, FK → `ads(id)` ON DELETE CASCADE | | Ad |
| `day` | `DATE` | PK part | | Reporting day |
| `impressions` | `INTEGER` | NOT NULL | `0` | Impressions |
| `clicks` | `INTEGER` | NOT NULL | `0` | Clicks |

### 17. `offer_stats_daily`

Daily offer counters flushed from Redis.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `offer_id` | `UUID` | PK part, FK → `offers(id)` ON DELETE CASCADE | | Offer |
| `day` | `DATE` | PK part | | Reporting day |
| `views` | `INTEGER` | NOT NULL | `0` | Deduplicated views |
| `wa_clicks` | `INTEGER` | NOT NULL | `0` | WhatsApp clicks |

### 18. `notifications`

Header notification list and system notifications.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `UUID` | PK | `gen_random_uuid()` | Notification identifier |
| `user_id` | `UUID` | FK → `users(id)` ON DELETE CASCADE | | Recipient |
| `type` | `TEXT` | NOT NULL | | Notification type |
| `dedupe_key` | `TEXT` | NOT NULL | | Idempotency key |
| `title` | `TEXT` | NOT NULL | | Localized/rendered title |
| `body` | `TEXT` | NOT NULL | | Localized/rendered body |
| `data` | `JSONB` | NOT NULL | `'{}'` | Safe navigation metadata |
| `read_at` | `TIMESTAMPTZ` | | NULL | Read time |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Constraint:** unique `(user_id, dedupe_key)`.

### 19. `jobs`

Durable PostgreSQL job queue claimed with `FOR UPDATE SKIP LOCKED`.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `BIGINT` | PK | identity | Job identifier |
| `kind` | `TEXT` | NOT NULL | | Job type |
| `payload` | `JSONB` | NOT NULL | | Job data |
| `run_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Earliest run |
| `attempts` | `SMALLINT` | NOT NULL | `0` | Retry count |
| `last_error` | `TEXT` | | NULL | Sanitized failure |
| `locked_at` | `TIMESTAMPTZ` | | NULL | Claim timestamp |
| `done_at` | `TIMESTAMPTZ` | | NULL | Completion timestamp |

**Index:** `jobs_ready_idx(run_at) WHERE done_at IS NULL`.

### 20. `audit_logs`

Append-only security and business audit trail.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `BIGINT` | PK | identity | Audit identifier |
| `actor_id` | `UUID` | FK → `users(id)` | NULL | Actor, if authenticated |
| `action` | `TEXT` | NOT NULL | | Action name |
| `entity_type` | `TEXT` | NOT NULL | | Entity kind |
| `entity_id` | `TEXT` | NOT NULL | | Entity identifier |
| `meta` | `JSONB` | NOT NULL | `'{}'` | Safe audit metadata |
| `ip` | `INET` | | NULL | Optional request IP |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Event time |

**Rule:** application role has no `UPDATE` or `DELETE` permission on this table.

### 21. `casbin_rule`

Persistent Casbin RBAC and route/action policy rows. The table is loaded through a SQLBoiler-backed Casbin adapter.

| Column | Type | Constraints | Default | Description |
|---|---|---|---|---|
| `id` | `BIGINT` | PK | identity | Policy row |
| `ptype` | `TEXT` | NOT NULL | | `p` or `g` |
| `v0` | `TEXT` | NOT NULL | | Subject/role |
| `v1` | `TEXT` | NOT NULL | | Object/path |
| `v2` | `TEXT` | NOT NULL | | Action |
| `v3` | `TEXT` | NOT NULL | `''` | Effect or extra field |
| `v4` | `TEXT` | NOT NULL | `''` | Optional field |
| `v5` | `TEXT` | NOT NULL | `''` | Optional field |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Creation time |

**Constraint:** unique normalized policy tuple `(ptype, v0, v1, v2, v3, v4, v5)`.

---

## Views

### `public_offers`

The single source of truth for public offer visibility:

```sql
CREATE VIEW public_offers AS
SELECT o.*
FROM offers o
JOIN merchants m ON m.id = o.merchant_id
WHERE o.status = 'approved'
  AND o.starts_at <= now()
  AND o.ends_at > now()
  AND m.verification_status = 'approved'
  AND m.access_until IS NOT NULL
  AND m.access_until > now();
```

All public lists, details, similar offers, shop pages, sitemap data, and public counts must use this view. Do not duplicate these conditions in handlers.

---

## Foreign Key Map

```text
users
 ├── refresh_tokens.user_id
 ├── merchants.user_id
 ├── merchant_documents.reviewed_by
 ├── offers.reviewed_by
 ├── payments.reviewed_by
 ├── notifications.user_id
 ├── audit_logs.actor_id
 └── settings.updated_by

wilayas
 ├── merchants.personal_wilaya_id / wilaya_id
 ├── offers.wilaya_id
 └── ads.wilaya_id

categories
 ├── categories.parent_id
 ├── merchants.category_id
 ├── offers.category_id
 └── ads.category_id

merchants
 ├── merchant_documents.merchant_id
 ├── offers.merchant_id
 ├── subscriptions.merchant_id
 └── payments.merchant_id

offers
 ├── offer_media.offer_id
 ├── reports.offer_id
 └── offer_stats_daily.offer_id

ads
 └── ad_stats_daily.ad_id
```

---

## Cascade Rules

| Parent | Child | ON DELETE |
|---|---|---|
| `users` | `refresh_tokens` | CASCADE |
| `users` | `notifications` | CASCADE |
| `merchants` | `merchant_documents` | CASCADE |
| `offers` | `offer_media` | CASCADE |
| `offers` | `reports` | CASCADE |
| `offers` | `offer_stats_daily` | CASCADE |
| `ads` | `ad_stats_daily` | CASCADE |
| payment/audit/business parents | history | RESTRICT/NO ACTION |

Hard deletion of merchants, offers, payments, subscriptions, webhook events, or audit logs is forbidden in normal application flows.

---

## PostgreSQL Types

```sql
CREATE TYPE user_role           AS ENUM ('merchant', 'admin');
CREATE TYPE verification_status AS ENUM ('incomplete', 'pending', 'approved', 'rejected', 'suspended');
CREATE TYPE document_type       AS ENUM ('id_front', 'id_back', 'commercial_register');
CREATE TYPE document_status     AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE offer_status        AS ENUM ('draft', 'pending', 'approved', 'rejected', 'archived', 'removed');
CREATE TYPE payment_method      AS ENUM ('chargily', 'manual');
CREATE TYPE payment_status      AS ENUM ('pending', 'under_review', 'paid', 'failed', 'canceled', 'expired', 'rejected');
CREATE TYPE subscription_status AS ENUM ('active', 'canceled');
CREATE TYPE report_reason       AS ENUM ('expired_offer', 'wrong_price', 'fake', 'inappropriate', 'other');
CREATE TYPE report_status       AS ENUM ('open', 'resolved', 'dismissed');
CREATE TYPE ad_status            AS ENUM ('draft', 'active', 'paused', 'ended');
CREATE TYPE media_kind           AS ENUM ('image', 'video');
```

### Role model

`users.role` is the coarse account role. Fine-grained administration is Casbin RBAC:

| Casbin subject | Scope |
|---|---|
| `merchant` | Merchant onboarding, own offers, own billing |
| `admin` | Full admin route/action access |
| `admin_reviewer` | Merchant/offer/payment review permissions explicitly listed in policy |
| `admin_support` | Read-only support permissions explicitly listed in policy |

The default Casbin effect is deny. Roles are never accepted from request bodies, local storage, or client claims.

### Status values

| Table | Column | Values |
|---|---|---|
| `merchants` | `verification_status` | `incomplete`, `pending`, `approved`, `rejected`, `suspended` |
| `merchant_documents` | `status` | `pending`, `approved`, `rejected` |
| `offers` | `status` | `draft`, `pending`, `approved`, `rejected`, `archived`, `removed` |
| `payments` | `status` | `pending`, `under_review`, `paid`, `failed`, `canceled`, `expired`, `rejected` |
| `subscriptions` | `status` | `active`, `canceled` |
| `reports` | `status` | `open`, `resolved`, `dismissed` |
| `ads` | `status` | `draft`, `active`, `paused`, `ended` |

---

## Extensions and Indexes

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS citext;
```

Important indexes:

| Index | Table | Predicate/use |
|---|---|---|
| `merchants_status_idx` | `merchants` | review queue by status and submission |
| `merchants_access_idx` | `merchants` | public access expiry |
| `offers_visible_idx` | `offers` | `status = 'approved'` |
| `offers_filter_idx` | `offers` | approved wilaya/category/date |
| `offers_review_idx` | `offers` | `status = 'pending'` |
| `offers_search_idx` | `offers` | GIN `search_vector` |
| `offers_trgm_idx` | `offers` | GIN trigram `search_text` |
| `reports_open_idx` | `reports` | `status = 'open'` |
| `ads_active_idx` | `ads` | `status = 'active'` |
| `jobs_ready_idx` | `jobs` | `done_at IS NULL` |

---

## Triggers and Constraints

- `set_updated_at()` updates `merchants.updated_at` and `offers.updated_at` on every update.
- `enforce_max_media()` reads `settings.max_images_per_offer` and rejects images above the configured limit.
- Phone numbers must be normalized to Algerian E.164 before persistence.
- `offers` enforce title, description, price, date, and duration constraints in PostgreSQL.
- `merchant_documents` permits one current non-rejected document per type.
- `payments` require a manual proof key before manual approval.
- `webhook_events` uses `(provider, event_id)` as the idempotency constraint.
- `audit_logs` is append-only and must not be updated or deleted by the application role.

---

## Arabic/French Search

Normalize once in Go before saving and before searching through `search.Normalize`:

- remove Arabic diacritics and tatweel;
- normalize Arabic letter variants to their canonical forms;
- convert Arabic-Indic digits to Latin digits;
- lowercase Latin text;
- build `search_text` from title, description, shop name, category name, and wilaya name.

```sql
SELECT o.id, o.title, o.description,
       ts_rank_cd(o.search_vector, to_tsquery('simple', $2)) AS rank
FROM public_offers o
WHERE o.search_vector @@ to_tsquery('simple', $2)
   OR o.search_text % $1
ORDER BY rank DESC, o.ends_at ASC
LIMIT $3 OFFSET $4;
```

Repositories must whitelist sortable fields and parameterize all values. Never concatenate user input into SQL.

---

## Critical Transactions

### Approve offer and start trial

1. Start a PostgreSQL transaction.
2. Lock the merchant row with `SELECT ... FOR UPDATE`.
3. Approve only a `pending` offer.
4. Start the one-time trial only if `trial_started_at IS NULL`.
5. Write the admin audit entry in the same transaction.
6. Commit, then invalidate public-offer cache keys through Redis.

### Activate a subscription after payment

1. Verify webhook signature before database work.
2. Insert `(provider, event_id)` into `webhook_events`; duplicate events return HTTP 200 without side effects.
3. Lock the merchant and payment rows.
4. Compare provider amount/currency with the pending payment.
5. Insert the subscription from `GREATEST(now(), access_until)`.
6. Mark the payment paid and update `merchants.access_until` in the same transaction.
7. Write audit data and commit.

`success_url` never activates access. Redis replay locks are advisory only; PostgreSQL uniqueness and row locks are authoritative.

### Casbin policy mutation

1. Require the current actor's Casbin permission.
2. Insert/update/delete the normalized `casbin_rule` row in PostgreSQL.
3. Write old and new policy values to `audit_logs` in the same transaction.
4. Publish `dinar:casbin:policy:changed` through Redis after commit.

---

## Redis Data Boundary

Redis is not a second database for business entities.

| Key family | Purpose | TTL/flush |
|---|---|---|
| `dinar:cache:categories:*` | Reference cache | 10 minutes; invalidate on admin change |
| `dinar:cache:offers:*` | Public list/detail cache | 60 seconds; invalidate on moderation |
| `dinar:ratelimit:*` | Atomic rate windows | Window TTL |
| `dinar:stats:views:*` | View counters | Flush every 30 seconds to PostgreSQL |
| `dinar:dedupe:view:*` | View deduplication | 1 hour |
| `dinar:casbin:policy:changed` | Policy invalidation event | Pub/Sub |
| `dinar:lock:*` | Short distributed locks | Explicit short TTL |

Redis outage must never authorize a request, duplicate a payment, or lose durable audit data. Public cache may degrade to PostgreSQL with an alert; security-sensitive rate limits fail closed.

---

## SQLBoiler Rules

- Generated models live under `apps/api/db/models/`; never edit them manually.
- Repositories use SQLBoiler models and query modifiers for normal CRUD.
- Parameterized repository SQL is allowed for `FOR UPDATE SKIP LOCKED`, full-text ranking, or PostgreSQL features SQLBoiler cannot express cleanly.
- Handlers never import `database/sql`, SQLBoiler models, or Redis clients directly.
- Regenerate models after every schema migration:

```bash
sqlboiler psql --config db/sqlboiler.toml
```

Do not place database passwords in commands or CI logs.

---

## Migration Order

| # | File | Creates/changes |
|---|---|---|
| 1 | `0001_extensions_enums.up.sql` | Extensions and PostgreSQL enums |
| 2 | `0002_reference_tables.up.sql` | `wilayas`, `categories` |
| 3 | `0003_users_refresh_tokens.up.sql` | `users`, `refresh_tokens` |
| 4 | `0004_settings_merchants_documents.up.sql` | `settings`, `merchants`, `merchant_documents` |
| 5 | `0005_offers_media.up.sql` | `offers`, `offer_media` and indexes |
| 6 | `0006_billing.up.sql` | `subscriptions`, `payments`, `webhook_events` |
| 7 | `0007_reports_ads_stats.up.sql` | Reports, ads, daily statistics |
| 8 | `0008_notifications_jobs_audit.up.sql` | Notifications, jobs, audit logs |
| 9 | `0009_casbin_rules.up.sql` | Persistent Casbin policy rows and indexes |
| 10 | `0010_public_offers_view.up.sql` | `public_offers` view |
| 11 | `0011_search_and_constraints.up.sql` | Search indexes, triggers, and constraints |
| 12 | `0012_seed_reference_data.up.sql` | Wilayas, categories, settings, ad slots |

Never modify an applied migration. Every schema change is a new migration with `up` and `down` files. CI must run `up → down → up` on a disposable PostgreSQL database.

---

## Running Migrations

```bash
cd apps/api

# Apply pending migrations
make migrate-up

# Roll back the latest migration
make migrate-down

# Regenerate SQLBoiler models after the schema is ready
sqlboiler psql --config db/sqlboiler.toml
```

Run migrations as a separate deployment step before starting the new API/worker version.

---

## Seeds

- `wilayas`: official Algerian wilaya list with Arabic/French names and slugs.
- `categories`: clothing, electronics, home, appliances, beauty, food, supermarket, automotive, sports, children, services, and other.
- `settings`: trial length, subscription price/duration, image limit, active-offer limit, and manual payment instructions.
- `ad_slots`: `home_top`, `list_inline`, `detail_sidebar`, and `footer_banner`.
- Casbin policies: minimum admin bootstrap policy only; further roles are granted through an audited admin flow.

### Local demo catalogue

Migration `000002_demo_products.up.sql` creates the local-only
`demo_products` and `demo_product_media` tables. Migration
`000005_consumer_catalog.up.sql` adds the database mirror for the eight
consumer categories and their subcategories. The active embedded demo manifest
contains one representative listing per requested category plus makeup and
fragrance examples in beauty. Migration 000006_consumer_audience adds the
audience_gender, age_min, and age_max fields to demo_products with PostgreSQL
checks for men, women, unisex, and ages from 0 to 120. Each
`demo_product_media.object_key` points to an object in the MinIO
`dinar-public` bucket, while `source_url` preserves the origin for development
traceability. These tables are not a replacement for the production
`offers`/`offer_media` model; they allow the guest catalogue to show realistic
media before merchant persistence is implemented.

Run the idempotent local seed from the repository root:

```bash
ENV_FILE=.env.example ./scripts/seed-demo-products.sh
```

### Category, subcategory, and brand design

The catalogue taxonomy is intentionally normalized into three levels:

```text
categories 1 ── * subcategories 1 ── * subcategory_brands * ── 1 brands
                                      │
                                      └── offers.brand_id / demo_products.brand_slug
```

`catalog_categories` and `catalog_subcategories` own the seeded navigation
hierarchy. `brands` stores localized brand names and an optional logo reference.
`subcategory_brands`
controls which brands are valid and their display order for a selected
subcategory. Product rows store the stable brand reference, allowing
`GET /api/v1/brands?category=...&subcategory=...` and `GET /api/v1/offers?...&brand=...`
to use the same taxonomy without duplicating localized labels in every offer.
Migrations `000003_catalog_brands.up.sql` and `000005_consumer_catalog.up.sql`
add the brand tables, consumer catalogue tables, lookup indexes, and the demo
product brand column.

### Consumer audience filtering

Audience data describes the intended consumer for discovery; it is not an
access-control rule. An unisex product matches both men and women filters.
Age filtering uses inclusive interval overlap: a product matches age_from when
age_max is greater than or equal to age_from, and matches age_to when age_min
is less than or equal to age_to. The API accepts gender, age_from, and age_to
on GET /api/v1/offers, while the web app loads the catalogue once and applies
the same filters locally without a page reload.

---

## Maintenance and Backups

- Run encrypted `pg_dump -Fc` daily and WAL archiving where available.
- Test restoration monthly in an isolated environment.
- Remove expired/revoked refresh tokens and completed jobs according to retention policy.
- Aggregate or archive `offer_stats_daily` after 24 months.
- Keep autovacuum enabled and monitor `offers`, `jobs`, `audit_logs`, and policy-table growth.
- Never use a Redis backup as the only backup for payments, permissions, subscriptions, or audit logs.

---

## Pre-merge Checklist

- [ ] Migration has a working `down` file and is safe to run in CI.
- [ ] `CHECK`, `UNIQUE`, and `FOREIGN KEY` rules are in PostgreSQL, not only in Go.
- [ ] Every repeated `WHERE`/`JOIN` has an appropriate index.
- [ ] All public offer queries use `public_offers`.
- [ ] SQLBoiler models were regenerated and generated files were not manually edited.
- [ ] Casbin policy changes are persisted in PostgreSQL, audited, and published through Redis after commit.
- [ ] Default authorization is deny; no role is accepted from the client.
- [ ] Payment/webhook idempotency remains enforced by PostgreSQL constraints.
- [ ] No card data, raw IPs, private storage keys, or secrets are written to logs or unsafe JSON.
- [ ] Search, public visibility, cache invalidation, and Redis outage tests pass.
