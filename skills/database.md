---
name: dinar-dz-database
description: Current PostgreSQL 16 schema and persistence rules for the Dinar DZ catalogue, guest analytics, and merchant identity/verification flows. Use this skill whenever changing migrations, SQL, repositories, authentication sessions, authorization policies, private document metadata, notifications, or audit records.
---

# Database Schema

This document describes the schema that is implemented in
`apps/api/db/migrations/` as of migrations `000001` through `000010`. Future
designs are listed separately and must not be treated as deployed tables.

## Current Persistence Stack

| Concern | Current implementation |
|---|---|
| Database | PostgreSQL 16 |
| Go access | `database/sql` with the pgx v5 stdlib driver |
| Migrations | Ordered `.up.sql` and `.down.sql` files in `apps/api/db/migrations/` |
| IDs | PostgreSQL-generated UUIDs for catalogue/analytics rows; application-generated UUIDs for EPIC-02 rows |
| Timestamps | `TIMESTAMPTZ`, stored in UTC |
| Authorization | Casbin v3 rules loaded from `auth_policies` |
| Private files | MinIO/S3-compatible `dinar-private` bucket; PostgreSQL stores metadata and object keys only |
| Email delivery | Durable PostgreSQL outbox claimed by the worker |
| Redis | Runtime/cache infrastructure only; not the source of truth for current database entities |

There is no SQLBoiler-generated model layer or integrated migration runner in
the repository today. Do not claim either is active until it is implemented.
The API has an in-memory fallback used by isolated tests and local flows when no
database is configured; it is not durable and is not a production data store.

## Source of Truth

- Migration files define the physical schema.
- Parameterized SQL in `apps/api/internal/auth/`,
  `apps/api/internal/guest/`, and `apps/api/internal/notifications/` defines the
  current repository behavior.
- Raw refresh tokens and uploaded document bytes must never be stored in
  PostgreSQL.
- PostgreSQL is authoritative for refresh-token state, merchant verification,
  authorization policies, notification delivery state, analytics deduplication,
  and audit records.
- Never edit an applied migration. Add the next numbered `up` and `down` pair.

## Implemented Entity Map

```text
catalog_categories ──< catalog_subcategories

brands ──< subcategory_brands >── category/subcategory slugs
  │
  └──────────────────────────────< demo_products ──< demo_product_media

guest_offer_events              guest_offer_reports
                                         │
                                         └── guest_offer_report_rollup (view)

users ──< refresh_tokens
  │
  ├── merchant_profiles
  ├── merchant_shops
  ├── merchant_verifications
  ├──< merchant_documents ── object bytes in private MinIO
  └──< audit_log

auth_policies                   notification_outbox
```

The guest catalogue currently uses an embedded JSON manifest for deterministic
public reads and mirrors catalogue/demo data into PostgreSQL for local seeding,
merchant/admin workflows, and reporting. There is not yet a production
`offers` table.

## Implemented Tables

### `demo_products`

Local/demo catalogue products.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK, `gen_random_uuid()` |
| `external_id` | `TEXT` | NOT NULL, UNIQUE |
| `title_ar`, `title_fr` | `TEXT` | NOT NULL |
| `description_ar`, `description_fr` | `TEXT` | NOT NULL |
| `category_slug`, `subcategory_slug` | `TEXT` | NOT NULL |
| `brand_slug` | `TEXT` | nullable |
| `wilaya_code` | `TEXT` | NOT NULL |
| `city_name`, `city_name_ar` | `TEXT` | NOT NULL |
| `source_url` | `TEXT` | NOT NULL |
| `status` | `TEXT` | `draft`, `approved`, or `archived`; default `approved` |
| `audience_gender` | `TEXT` | `men`, `women`, or `unisex`; default `unisex` |
| `age_min`, `age_max` | `INTEGER` | inclusive range `0..120`; `age_max >= age_min` |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |

Indexes cover category/subcategory, location, category/subcategory/brand, and
audience range filtering.

### `demo_product_media`

Public demo media metadata. `product_id` references `demo_products(id)` with
`ON DELETE CASCADE`.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK, `gen_random_uuid()` |
| `product_id` | `UUID` | FK, NOT NULL |
| `object_key` | `TEXT` | NOT NULL, UNIQUE |
| `bucket_name` | `TEXT` | default `dinar-public` |
| `source_url`, `public_url` | `TEXT` | NOT NULL |
| `content_type` | `TEXT` | default `image/webp` |
| `byte_size` | `BIGINT` | default `0` |
| `position` | `INTEGER` | greater than `0`, default `1` |
| `created_at` | `TIMESTAMPTZ` | default `now()` |

### `brands`

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK, `gen_random_uuid()` |
| `slug` | `TEXT` | NOT NULL, UNIQUE |
| `name_ar`, `name_fr` | `TEXT` | NOT NULL |
| `logo_url` | `TEXT` | nullable |
| `is_active` | `BOOLEAN` | default `TRUE` |
| `origin` | `TEXT` | `global` or `local`; default `global` |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |

### `subcategory_brands`

Many-to-many brand availability per category/subcategory. Its composite primary
key is `(brand_id, category_slug, subcategory_slug)`. `brand_id` references
`brands(id)` with `ON DELETE CASCADE`; `sort_order` defaults to `0`.

### `catalog_categories`

| Column | Type | Constraints/default |
|---|---|---|
| `slug` | `TEXT` | PK |
| `name_ar`, `name_fr` | `TEXT` | NOT NULL |
| `icon` | `TEXT` | NOT NULL |
| `sort_order` | `INTEGER` | default `0` |
| `is_active` | `BOOLEAN` | default `TRUE` |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |

### `catalog_subcategories`

| Column | Type | Constraints/default |
|---|---|---|
| `slug` | `TEXT` | PK |
| `category_slug` | `TEXT` | FK to `catalog_categories(slug)`, `ON DELETE CASCADE` |
| `name_ar`, `name_fr` | `TEXT` | NOT NULL |
| `sort_order` | `INTEGER` | default `0` |
| `is_active` | `BOOLEAN` | default `TRUE` |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |

The `(category_slug, sort_order)` index supports ordered navigation.

### `guest_offer_events`

Privacy-preserving, deduplicated guest engagement events.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK, `gen_random_uuid()` |
| `offer_id` | `TEXT` | NOT NULL |
| `event_type` | `TEXT` | `view` or `whatsapp` |
| `visitor_hash` | `TEXT` | NOT NULL |
| `bucket_start` | `TIMESTAMPTZ` | NOT NULL |
| `created_at` | `TIMESTAMPTZ` | default `now()` |

The unique key `(offer_id, event_type, visitor_hash, bucket_start)` provides
time-bucket deduplication. Raw IP addresses or visitor identifiers are not
stored.

### `guest_offer_reports`

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK, `gen_random_uuid()` |
| `offer_id` | `TEXT` | NOT NULL |
| `visitor_hash` | `TEXT` | NOT NULL |
| `reason` | `TEXT` | `misleading`, `expired`, `wrong_price`, `inappropriate`, or `other` |
| `message` | `TEXT` | default empty; maximum 1,000 characters |
| `created_at` | `TIMESTAMPTZ` | default `now()` |

Only one report per `(offer_id, visitor_hash)` is stored. Indexes support
visitor throttling and review-queue lookup.

### `users`

Google-authenticated identities.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK; generated by the application |
| `email` | `TEXT` | NOT NULL, UNIQUE |
| `name`, `first_name`, `last_name` | `TEXT` | default empty |
| `picture_url` | `TEXT` | default empty |
| `email_verified` | `BOOLEAN` | default `FALSE` |
| `role` | `TEXT` | `user`, `merchant`, or `admin`; default `merchant` |
| `google_subject` | `TEXT` | nullable; unique when present |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |

On Google login, the application looks up by provider subject or normalized
email. A legacy email-only row may be linked when `google_subject` is null; a
row already bound to a different Google subject is rejected.

### `refresh_tokens`

Durable rotated refresh sessions. Only SHA-256 hashes are stored; access tokens
are not persisted.

| Column | Type | Constraints/default |
|---|---|---|
| `token_hash` | `TEXT` | PK |
| `user_id` | `UUID` | FK to `users(id)`, `ON DELETE CASCADE` |
| `family_id` | `TEXT` | NOT NULL, indexed |
| `jti` | `TEXT` | NOT NULL, UNIQUE |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL |
| `revoked_at` | `TIMESTAMPTZ` | nullable |
| `replaced_by_hash` | `TEXT` | nullable |
| `user_agent` | `TEXT` | default empty |
| `ip_hash` | `TEXT` | default empty |

`ip_hash` is a salted application hash. Never persist or log the raw refresh
token.

### `merchant_profiles`

One row per user, keyed by `user_id` with `ON DELETE CASCADE`.

| Column | Type | Constraints/default |
|---|---|---|
| `phone_e164` | `TEXT` | NOT NULL |
| `wilaya_id` | `TEXT` | NOT NULL |
| `address` | `TEXT` | NOT NULL |
| `updated_at` | `TIMESTAMPTZ` | default `now()` |

### `merchant_shops`

One row per user, keyed by `user_id` with `ON DELETE CASCADE`.

| Column | Type | Constraints/default |
|---|---|---|
| `name` | `TEXT` | NOT NULL |
| `slug` | `TEXT` | NOT NULL, UNIQUE |
| `category_id`, `wilaya_id` | `TEXT` | NOT NULL |
| `description`, `city`, `address`, `phone`, `whatsapp` | `TEXT` | default empty where declared |
| `register_number` | `TEXT` | NOT NULL |
| `register_expires_on` | `DATE` | NOT NULL |
| `slug_edits` | `SMALLINT` | `0..1`, default `0` |
| `updated_at` | `TIMESTAMPTZ` | default `now()` |

The current schema does not yet enforce foreign keys from shop category/wilaya
text values to the catalogue/reference tables.

### `merchant_documents`

Private document metadata; file bytes live in MinIO.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK; generated by the application |
| `user_id` | `UUID` | FK to `users(id)`, `ON DELETE CASCADE` |
| `document_type` | `TEXT` | `id_front`, `id_back`, or `commercial_register` |
| `content_type` | `TEXT` | NOT NULL |
| `byte_size` | `BIGINT` | `1..8,388,608` bytes |
| `sha256` | `TEXT` | NOT NULL |
| `storage_key` | `TEXT` | NOT NULL, UNIQUE |
| `created_at` | `TIMESTAMPTZ` | default `now()` |

`(user_id, document_type)` is unique, so each upload replaces the metadata for
that document type. The normal merchant API response must not expose
`storage_key` or file bytes. Admin download access uses a five-minute signed URL
and writes an audit record.

### `merchant_verifications`

One verification state row per user.

| Column | Type | Constraints/default |
|---|---|---|
| `user_id` | `UUID` | PK, FK to `users(id)`, `ON DELETE CASCADE` |
| `status` | `TEXT` | `incomplete`, `pending`, `approved`, `rejected`, or `suspended` |
| `consent_accepted_at` | `TIMESTAMPTZ` | nullable |
| `rejection_reason` | `TEXT` | default empty |
| `rejection_document` | `TEXT` | default empty |
| `updated_at` | `TIMESTAMPTZ` | default `now()` |

Profile, shop, and document edits are locked while status is `pending`,
`approved`, or `suspended`. A rejected merchant may correct and resubmit.

### `auth_policies`

Casbin policy storage with primary key `(role, resource, action)`.

Seeded rules:

| Role | Resource | Action |
|---|---|---|
| `merchant` | `merchant_identity` | `write` |
| `admin` | `merchant_identity` | `write` |
| `admin` | `merchant_documents` | `read` |

Authorization is default-deny. The application loads the user's role and all
policy rows from PostgreSQL, then evaluates them with Casbin. Client-supplied
roles are never authoritative.

### `notification_outbox`

Durable email-delivery queue.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK; generated by the application |
| `kind` | `TEXT` | currently only `merchant_verification_status` |
| `recipient` | `TEXT` | NOT NULL |
| `payload` | `JSONB` | NOT NULL |
| `status` | `TEXT` | `pending`, `processing`, `sent`, or `failed` |
| `attempts` | `INTEGER` | default `0` |
| `available_at` | `TIMESTAMPTZ` | default `now()` |
| `last_error` | `TEXT` | default empty |
| `created_at`, `updated_at` | `TIMESTAMPTZ` | default `now()` |
| `sent_at` | `TIMESTAMPTZ` | nullable |

A partial index on `(available_at, created_at)` covers `pending` and `failed`
rows.

### `audit_log`

Security/business audit records. The table name is singular.

| Column | Type | Constraints/default |
|---|---|---|
| `id` | `UUID` | PK; generated by the application |
| `actor_user_id` | `UUID` | nullable FK to `users(id)`, `ON DELETE SET NULL` |
| `action` | `TEXT` | NOT NULL |
| `resource_type`, `resource_id` | `TEXT` | NOT NULL |
| `request_id` | `TEXT` | default empty |
| `ip_address` | `INET` | nullable |
| `metadata` | `JSONB` | default `{}` |
| `created_at` | `TIMESTAMPTZ` | default `now()` |

The implemented write is an admin document-view event. Audit rows are business
history and should be treated as append-only by application code.

## Implemented View

### `guest_offer_report_rollup`

Groups reports by `offer_id` and exposes:

- `report_count`
- `distinct_visitor_count`
- `last_reported_at`
- `needs_review`, true at three or more distinct reporters

There is no `public_offers` view yet because the production offer model has not
been migrated.

## Extensions and Database Types

The only installed extension is:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

The current migrations use text columns with `CHECK` constraints rather than
PostgreSQL enum types. `citext` and `pg_trgm` are not installed. There are no
database triggers in the implemented schema.

## Critical Transaction Boundaries

### Google identity upsert

1. Begin a transaction.
2. Find the user by Google subject or case-insensitive email.
3. Reject an existing row bound to a different subject.
4. Insert or update the user and create the initial verification row.
5. Commit both changes together.

### Refresh-token rotation and reuse response

1. Hash the cookie value and lock its row with `SELECT ... FOR UPDATE`.
2. If the token is expired or already revoked, revoke the entire token family.
3. Otherwise revoke the current token, store its replacement hash, and insert
   the new token in the same transaction.
4. Set the new raw token only in the secure HTTP-only cookie after commit.

Logout revokes the whole family. Raw tokens never enter SQL or logs.

### Private document replacement

1. Validate document type, detected MIME type, and the 8 MiB limit.
2. Upload the new object to private storage.
3. Begin a transaction and lock the existing document row for that user/type.
4. Upsert the new metadata and commit.
5. Delete the old object after commit. If the SQL operation fails, delete the
   newly uploaded object as compensation.

This is a coordinated storage/database operation, not a distributed
transaction. Cleanup failures must be observable so orphaned objects can be
reconciled.

### Merchant submission

1. Lock the verification row.
2. Reject edits/submission from locked states.
3. Verify that profile, shop, all three documents, and consent exist.
4. Set status to `pending` and clear old rejection fields.
5. Insert the status email into `notification_outbox` in the same transaction.
6. Commit status and message atomically.

Admin verification decisions follow the same outbox rule: the status update and
email enqueue both commit or neither does.

### Admin document access

1. Require the Casbin `merchant_documents/read` permission.
2. Read the private object key from PostgreSQL.
3. Generate a signed URL with a maximum five-minute TTL.
4. Insert the `audit_log` row before returning the URL.

### Outbox worker

1. Select one ready row using `FOR UPDATE SKIP LOCKED`.
2. Reclaim a `processing` row only when it has been stale for ten minutes.
3. Mark it `processing`, increment attempts, and commit the claim.
4. Send through SMTP.
5. Mark it `sent`, or `failed` with a capped exponential-minute retry delay.

The worker is at-least-once. Email providers may still receive a duplicate if a
process dies after delivery but before the final database update.

## Migration Order

| # | File | Implemented change |
|---|---|---|
| 1 | `000001_create_extensions` | Enables `pgcrypto` |
| 2 | `000002_demo_products` | Demo products and public media metadata |
| 3 | `000003_catalog_brands` | Brands, subcategory mappings, product brand |
| 4 | `000004_seed_catalog_brands` | Initial brand and mapping seed |
| 5 | `000005_consumer_catalog` | Category/subcategory mirror |
| 6 | `000006_consumer_audience` | Gender and age targeting |
| 7 | `000007_brand_origin` | Global/local brand classification |
| 8 | `000008_guest_analytics` | Guest events, reports, and report rollup |
| 9 | `000009_merchant_identity` | Users, sessions, onboarding, documents, verification |
| 10 | `000010_epic02_completion` | Google subject, Casbin policies, outbox, audit log |

Every numbered migration has an `.up.sql` and `.down.sql` file. A future
migration must use the next sequence number and must not rewrite these files.

## Running Migrations Locally

The repository does not yet track applied versions or expose `make migrate-up`.
On a brand-new local database, apply all migrations in lexical order from the
repository root:

```bash
for migration in apps/api/db/migrations/*.up.sql; do
  docker compose --env-file .env -f docker-compose.dev.yml exec -T postgres \
    sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
    < "$migration"
done
```

On an existing database, apply only the new migration after confirming the last
applied version. Example for EPIC-02 completion:

```bash
docker compose --env-file .env -f docker-compose.dev.yml exec -T postgres \
  sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < apps/api/db/migrations/000010_epic02_completion.up.sql
```

The `IF NOT EXISTS` clauses make several DDL statements repeatable, but they are
not a substitute for a migration ledger. Add a real migration runner before
automated multi-instance production deployments.

## Local Demo Catalogue Seed

After migration `000001` has enabled `pgcrypto`, the seed script starts
PostgreSQL and MinIO, applies catalogue migrations `000002` through `000007`,
uploads public demo images, and upserts the JSON catalogue data:

```bash
ENV_FILE=.env.example ./scripts/seed-demo-products.sh
```

Its destructive operations are scoped to demo catalogue rows and objects. It is
for local/demo data, not production merchant offers.

## Security and Data-Handling Rules

- Parameterize every SQL value. Never concatenate request input into SQL.
- Never store raw refresh tokens, OAuth authorization codes, or private file
  bytes in PostgreSQL.
- Never return private `storage_key` values to merchant clients.
- Treat `visitor_hash` and `ip_hash` as pseudonymous data; keep salts outside
  the database.
- Authorize signed-document access from database-backed policy and write the
  audit record before returning the URL.
- Keep authorization default-deny when PostgreSQL or policy loading fails.
- Keep merchant verification changes and their notification outbox insert in
  one transaction.
- Preserve uniqueness and `CHECK` constraints when repository validation is
  added or changed; Go validation is not a replacement for database integrity.
- Avoid hard-deleting audit history. Define and document retention before
  purging sessions, outbox rows, analytics, or uploaded documents.

## Planned, Not Yet Migrated

The following were part of the broader product design but do not exist in the
current database:

- canonical `offers` and `offer_media` tables plus a `public_offers` view;
- normalized wilaya/reference settings tables;
- merchant trial and subscription access periods;
- Chargily/manual payments and webhook idempotency records;
- advertisements and daily statistics tables;
- user-facing notifications and a general-purpose jobs table;
- a generalized `casbin_rule` adapter table;
- Arabic/French full-text and trigram search indexes;
- SQLBoiler-generated models;
- a migration ledger/runner and deployment migration command;
- production backup, archival, and retention automation.

Add these only through reviewed migrations and update this document when the
schema becomes real.

## Pre-Merge Checklist

- [ ] The new migration has matching `up` and `down` files.
- [ ] Existing applied migrations were not modified.
- [ ] Table and column names match repository SQL exactly.
- [ ] `CHECK`, `UNIQUE`, and foreign-key rules are enforced in PostgreSQL where appropriate.
- [ ] New repeated filters, joins, and queue claims have suitable indexes.
- [ ] Multi-row business changes use an explicit transaction and row locks where needed.
- [ ] Authorization remains default-deny and roles are loaded from trusted storage.
- [ ] Refresh tokens remain hashed and rotation/reuse tests pass.
- [ ] Private document bytes remain outside PostgreSQL and signed URL access is audited.
- [ ] Verification status and notification enqueue remain atomic.
- [ ] Migration, API, worker, and database smoke tests pass on a disposable database.
