---
name: dinar-dz-backend
description: Architecture, conventions, endpoints, and code patterns for the Dinar DZ Go Fiber API using PostgreSQL, Redis, SQLBoiler, and Fiber. Use this skill whenever writing or changing Go backend code for Dinar DZ, including Google OAuth JWT authentication, merchant onboarding, offers, image upload and processing, Chargily payments and webhooks, admin review, ads, background jobs, middleware, error handling, or API endpoints, even if the user only says "add an endpoint" or "fix the API".
---

# Dinar DZ — Backend (Go + Fiber)

Reference: [`../userstory.md`](../userstory.md) for product rules and `database.md` for the schema. The backend is the **single source of truth for business logic**.

## 1. Technology stack

| Purpose | Library |
|---|---|
| HTTP | `github.com/gofiber/fiber/v2` |
| Database | PostgreSQL 16 through `database/sql` + `github.com/jackc/pgx/v5/stdlib` |
| ORM and model generation | `github.com/volatiletech/sqlboiler/v4` + `boil`, `queries`, `qm` |
| Cache and counters | Redis 7 through `github.com/redis/go-redis/v9` |
| Live notifications | Fiber WebSocket + Redis Pub/Sub, with PostgreSQL as durable storage |
| Authorization and roles | Casbin v2 through `github.com/casbin/casbin/v2` (RBAC + route/action policies) |
| Migrations | `github.com/golang-migrate/migrate/v4` |
| OAuth | `golang.org/x/oauth2` + `google` |
| Validation | `github.com/go-playground/validator/v10` |
| File storage | MinIO S3 through `github.com/minio/minio-go/v7` |
| Images | `github.com/davidbyttow/govips/v2` (libvips) |
| Scheduling | `github.com/robfig/cron/v3` |
| Logging | `log/slog` (JSON) |
| Configuration | Environment variables through `caarlos0/env/v11` |
| Testing | `testing` + `testify` + `testcontainers-go` (real PostgreSQL, Redis, and MinIO) |
| Documentation | `swaggo/swag` to generate OpenAPI (used by the frontend for types) |
| Email | SMTP through `wneessen/go-mail` (or the Resend API) |

## 2. Project structure

```
apps/api/
  cmd/
    api/main.go            # runs the HTTP server
    worker/main.go         # runs cron and the job handler (or: api --mode=worker)
  internal/
    config/                # reads and validates environment at startup
    platform/
      db/                  # database/sql + SQLBoiler + transaction/health helpers
      cache/               # Redis client, key builders, TTLs, invalidation
      authz/               # Casbin enforcer, model, policy adapter, and policy watcher
      storage/             # MinIO S3 client: public and private buckets
      mailer/
      chargily/            # Chargily client and signature verification
      imaging/             # image processing (WebP, resize, EXIF)
      phone/               # Algerian phone normalization
      search/              # normalization and tsquery building
    http/
      server.go            # creates the Fiber app and registers routes
      middleware/          # auth, authz/Casbin, csrf, ratelimit, requestid, recover
      response/            # success and error envelopes
    auth/                  # handler + service (Google OAuth, JWT, refresh rotation)
    merchants/             # onboarding, documents, profile
    offers/                # CRUD, media, public queries
    billing/               # checkout, manual proof, subscriptions
    admin/                 # review queues, settings, audit
    ads/
    reports/
    stats/                 # view and click aggregation
    jobs/                  # tasks: expiry reminders, cleanup, email
  db/
    migrations/
    models/                # generated SQLBoiler models; do not edit manually
    queries/               # repositories and complex queries
    sqlboiler.toml         # PostgreSQL model-generation settings
  authz/
    model.conf             # Casbin RBAC model
    policy.csv              # development only; production loads policy from PostgreSQL
  Dockerfile
```

**Layer rule inside every module:**
`handler.go` (HTTP only) → `service.go` (business rules and transactions) → `repo` (SQLBoiler models and queries).
- The handler reads and validates input, calls the service, and writes the response. **No SQL or business rules.**
- The service does not know Fiber; it accepts `context.Context` and ordinary types and is testable without HTTP.
- The repository owns `database/sql` and SQLBoiler and maps generated models to domain types when needed.
- PostgreSQL is durable truth; Redis is a temporary cache, counter, or short lock and is never the only store for payments, permissions, or audit records.
- Avoid circular imports; communicate through small interfaces defined by the consumer.

Generate SQLBoiler models after every schema change in development or CI, then review the diff before merging:

```bash
go install github.com/volatiletech/sqlboiler/v4/cmd/sqlboiler@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
sqlboiler psql --config db/sqlboiler.toml
```

Never edit generated files manually or place a PostgreSQL password in a command or CI log.

## 3. Configuration (environment variables)

```
APP_ENV=production|development
APP_URL=https://dinar.dz
API_ADDR=:8080
DATABASE_URL=postgres://...
REDIS_URL=redis://redis:6379/0
REDIS_PASSWORD=
REDIS_PREFIX=dinar:
REDIS_MAX_RETRIES=3
CACHE_DEFAULT_TTL_SECONDS=600
REDIS_CHANNEL_NOTIFICATIONS=dinar:notifications
CASBIN_MODEL_PATH=/app/authz/model.conf
CASBIN_POLICY_SOURCE=postgres
GOOGLE_CLIENT_ID= GOOGLE_CLIENT_SECRET= GOOGLE_REDIRECT_URL=https://dinar.dz/api/v1/auth/google/callback
JWT_ISSUER=dinar-api  JWT_AUDIENCE=dinar-web  JWT_ALGORITHM=RS256
JWT_ACCESS_TTL_SECONDS=600  JWT_REFRESH_TTL_SECONDS=2592000  JWT_CLOCK_SKEW_SECONDS=30
JWT_PRIVATE_KEY_FILE=/run/secrets/jwt_private_key  JWT_PUBLIC_KEY_FILE=/run/secrets/jwt_public_key
REFRESH_TOKEN_COOKIE_NAME=__Host-dinar-refresh  COOKIE_SECURE=true
CSRF_COOKIE_NAME=__Host-dinar-csrf  CSRF_HEADER_NAME=X-CSRF-Token
WS_TICKET_TTL_SECONDS=30  ALLOWED_ORIGINS=https://dinar.dz  TRUSTED_HOSTS=dinar.dz
ADMIN_EMAILS=you@gmail.com,other@gmail.com
MINIO_ENDPOINT=minio:9000  MINIO_REGION=us-east-1
# Use true when MinIO has TLS; use false only on an isolated internal Docker network.
MINIO_USE_SSL=false
MINIO_ACCESS_KEY=  MINIO_SECRET_KEY=
MINIO_BUCKET_PUBLIC=dinar-public  MINIO_BUCKET_PRIVATE=dinar-private
PUBLIC_MEDIA_BASE_URL=https://media.dinar.dz
CHARGILY_API_KEY= CHARGILY_SECRET= CHARGILY_BASE_URL=https://pay.chargily.net/api/v2
SMTP_HOST= SMTP_USER= SMTP_PASS= MAIL_FROM="Dinar DZ <no-reply@dinar.dz>"
IP_HASH_SALT=<random>
NOTIFICATIONS_WS_ENABLED=true
NOTIFICATIONS_WS_PATH=/api/v1/ws/notifications
WS_ALLOWED_ORIGINS=https://dinar.dz
WS_MAX_CONNECTIONS_PER_USER=3
WS_PING_INTERVAL_SECONDS=25
WS_READ_TIMEOUT_SECONDS=45
WS_WRITE_TIMEOUT_SECONDS=10
WS_EVENT_MAX_BYTES=16384
```
- The application fails at startup when a required variable is missing. Do not provide secret defaults.
- `/readyz` fails when PostgreSQL or Redis is unavailable while rate limiting or locks are enabled. Public cache reads may degrade safely with an alert.
- Never store card data or payment secrets in Redis; store only non-sensitive transaction references in PostgreSQL.
- MinIO is required for file access; `/readyz` fails when required buckets cannot be reached.
- In development, set `CHARGILY_BASE_URL` to the test environment URL after checking the official documentation.

### 3.1 MinIO S3

MinIO is the backend file store, accessed through the S3 API and `minio-go/v7`:

- `dinar-public`: processed offer images, public logos, and ad images only.
- `dinar-private`: merchant documents and payment proofs only; no anonymous read and no public ACL.
- File keys are random and scoped: `offers/{offer_id}/{media_id}_{400|800|1600}.webp` and `merchants/{merchant_id}/{type}/{uuid}.{ext}`.
- Public files use `Content-Type` and `Cache-Control: public, max-age=31536000, immutable`.
- Private files are read through five-minute presigned URLs, and every access is recorded in `audit_logs`.
- Never return an access key, secret key, or private bucket key to the browser.
- Create buckets and policies through an idempotent init job; do not let the application change the production public policy automatically.
- If MinIO is exposed outside Docker, use TLS (`MINIO_USE_SSL=true`) with MinIO certificates. Local development may use `MINIO_USE_SSL=false` only on an isolated network with no public MinIO port; external access must go through Caddy HTTPS.

### 3.1.1 Local demo product seed

The local catalogue can be populated with nine consumer demo products sourced
from public product images. Run the idempotent seed script from the repository
root:

```bash
ENV_FILE=.env.example ./scripts/seed-demo-products.sh
```

The script downloads each source image to a temporary directory, uploads the
JPEG object to `dinar-public/demo-products/{external_id}.jpg`, and upserts the
product plus its media link in PostgreSQL. It also seeds the eight consumer
categories, 30 subcategories, and 90 global-brand mappings. The API then exposes the
MinIO public URL through the same public-offer response used by the web app.
The active source manifests are embedded under `apps/api/internal/guest/data/`;
the older `apps/api/db/seed/demo-products.json` is retained as historical demo
reference only. These files are only demo content and must not be treated as
verified merchant inventory. Re-running the script replaces only the
`demo-products` MinIO prefix and demo PostgreSQL rows. The beauty category
includes both makeup and fragrance examples. Each demo product also carries
`audience_gender` (`men`, `women`, or `unisex`) and an inclusive `age_min` /
`age_max` range for client-side and API filtering.

Example MinIO client initialization:

```go
client, err := minio.New(cfg.Endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
    Secure: cfg.UseSSL,
    Region: cfg.Region,
})
```

## 4. API conventions

- **Base:** `/api/v1`, JSON, UTF-8.
- **Naming:** `snake_case` field names, plural resources, action verbs such as `POST /offers/:id/submit`.
- **Identifiers:** UUID.
- **Pagination:** `?page=1&per_page=24` (maximum 48). Responses include `meta: {page, per_page, total}`.
- **Dates:** ISO 8601 UTC.
- **Success:** `{ "data": ... , "meta": ... }`
- **Error:**
```json
{ "error": { "code": "validation_failed", "message": "Check the fields", "fields": { "title": "Minimum 5 characters" }, "request_id": "..." } }
```
- Error codes: `unauthorized` 401, `forbidden` 403, `not_found` 404, `gone` 410, `conflict` 409, `validation_failed` 422, `rate_limited` 429, and `internal` 500.
- User-facing error messages are localized; internal details stay in logs only.
- Every request receives an `X-Request-ID` that is logged and returned.

### 4.1 Routes

#### WebSocket for live notifications

The API provides one WebSocket channel for user-specific notifications:

| Item | Value |
|---|---|
| Path | `GET /api/v1/ws/notifications` |
| Authentication | access JWT in `Authorization`; the handshake uses a one-use WebSocket ticket cookie |
| Transport between API instances | Redis Pub/Sub through `REDIS_CHANNEL_NOTIFICATIONS` |
| Source of truth | `notifications` rows in PostgreSQL, not Redis or WebSocket |
| Proxy | Caddy forwards the WebSocket upgrade to `api:8080` |

Notification sequence:

1. A service changes payment, merchant, or offer state inside a PostgreSQL transaction.
2. It creates a notification row with `user_id`, `dedupe_key`, `type`, `payload`, and `created_at`.
3. After commit, the worker or API publishes a small event to Redis Pub/Sub.
4. Each API instance receives the event and sends it only to sockets for the matching `user_id`.
5. Redis is not used to recover missed events; the frontend reconnects and requests `/me/notifications?since_id=...` or the full list to repair gaps.

The event contains no secrets or card data:

```json
{
  "type": "notification.created",
  "id": "notification-uuid",
  "user_id": "user-uuid",
  "notification_type": "payment_paid",
  "title_key": "notifications.paymentPaid.title",
  "body_key": "notifications.paymentPaid.body",
  "data": { "payment_id": "payment-uuid" },
  "created_at": "2026-10-09T14:00:00Z"
}
```

Implementation and security rules:

- Verify the WebSocket ticket and user ID on the server during the handshake; never accept `user_id` or JWT from the query string.
- Validate `Origin` against `WS_ALLOWED_ORIGINS` and reject unknown origins.
- Limit connections per user, enforce ping/pong and timeouts, and limit event size.
- Do not use Redis Pub/Sub as durable storage or treat WebSocket delivery as a read receipt; reads are stored through REST in PostgreSQL.
- When a deploy closes a socket, the frontend reconnects with exponential backoff and jitter and resynchronizes notifications through REST.
- Never log cookies, tokens, or sensitive notification content.

**Public (no authentication):**
| Method | Path | Description |
|---|---|---|
| GET | `/offers` | Search and filtering: `q, wilaya, city, category, subcategory, brand, gender, age_from, age_to, min_discount, ending_within, sort, page`; gender accepts `men`, `women`, `unisex`, and unisex products match men/women filters; age filters use inclusive range overlap |
| GET | `/offers/:id` | Details, images, and shop card |
| GET | `/offers/:id/similar` | Similar offers (same category and wilaya) |
| POST | `/offers/:id/view` | Record a view (no body) |
| POST | `/offers/:id/whatsapp` | Record a click and return a `wa.me` URL |
| POST | `/offers/:id/report` | Submit a report |
| GET | `/shops/:slug` | Shop page and offers |
| GET | `/categories` · `/brands` · `/wilayas` | Reference lists; `/brands` accepts `category` and `subcategory` |
| GET | `/ads?slot=&wilaya=&category=` | Matching ad for a slot |
| POST | `/ads/:id/click` | Record a click and redirect |
| GET | `/sitemap/offers` | Offer sitemap links |

**Authentication:**
| GET | `/auth/google/login` | Starts OAuth (stores state and PKCE in a temporary cookie) |
| GET | `/auth/google/callback` | Verifies Google, issues a refresh cookie, and redirects to the frontend |
| POST | `/auth/refresh` | Rotates the refresh JWT and returns an access JWT as JSON |
| POST | `/auth/ws-ticket` | Creates a one-use ticket for WebSocket access |
| POST | `/auth/logout` | Logs out and revokes the refresh family |
| GET | `/me` | User, merchant status, and access status |

**Merchant (requires a merchant session):**
| PUT | `/merchant/profile` | Personal information step |
| PUT | `/merchant/shop` | Shop information step |
| POST | `/merchant/documents` | multipart: `type`, `file` |
| POST | `/merchant/submit` | Submit for review (requires consent and three documents) |
| GET/POST | `/merchant/offers` | List / create draft |
| GET/PUT/DELETE | `/merchant/offers/:id` | Details / edit / delete draft |
| POST | `/merchant/offers/:id/media` | Upload image |
| PATCH | `/merchant/offers/:id/media/order` | Set `[ids]` order |
| DELETE | `/merchant/offers/:id/media/:mid` | Delete image |
| POST | `/merchant/offers/:id/submit` | Submit for review |
| POST | `/merchant/offers/:id/archive` · `/duplicate` | Archive or duplicate |
| GET | `/merchant/stats` | Summary and statistics |
| POST | `/merchant/billing/checkout` | Creates Chargily payment and returns `checkout_url` |
| POST | `/merchant/billing/manual` | multipart: proof + `proof_reference` |
| GET | `/merchant/billing/payments` | Payment history |
| GET | `/merchant/billing/payments/:id` | Payment status after returning from Chargily |

**Admin (Casbin permissions under admin roles):**
| GET | `/admin/dashboard` | Metrics |
| GET | `/admin/merchants?status=` | Review queue |
| GET | `/admin/merchants/:id` | Details |
| GET | `/admin/documents/:id/url` | Five-minute signed URL (audited) |
| POST | `/admin/merchants/:id/approve` · `/reject` · `/suspend` · `/reactivate` | |
| GET | `/admin/offers?status=pending` | Offer review queue |
| POST | `/admin/offers/:id/approve` · `/reject` · `/remove` | |
| POST | `/admin/offers/bulk` | `{action, ids[]}` |
| GET | `/admin/payments?status=under_review` | |
| POST | `/admin/payments/:id/approve` · `/reject` | |
| CRUD | `/admin/categories` · `/admin/ads` | |
| GET | `/admin/reports` · POST `/admin/reports/:id/resolve` | |
| GET/PUT | `/admin/settings` | |
| GET | `/admin/audit-logs` | |

**Webhooks:** `POST /webhooks/chargily` (no session; signature-protected).
**Health:** `GET /healthz` (liveness) and `GET /readyz` (database and storage readiness).

## 5. Authentication and sessions

### 5.1 Google OAuth and JWT flow
1. `/auth/google/login`: generate `state` and `code_verifier` (PKCE), store them in a
   short-lived signed cookie (10 minutes, `HttpOnly`), and redirect to Google with scopes:
   `openid email profile`.
2. `/auth/google/callback`: validate `state`, exchange the code, validate the `id_token`
   (`iss`, `aud`, `exp`, `email_verified == true`), then upsert `users`.
3. Create a short-lived access JWT and refresh JWT with `jti` and `family_id`. Store only
   the refresh hash in `refresh_tokens`, and set
   `__Host-dinar-refresh; HttpOnly; Secure; SameSite=Strict; Path=/` without a Domain.
4. Set a separate CSRF cookie and redirect to the frontend. The frontend calls
   `POST /auth/refresh` to obtain an access JWT in memory only.
5. Redirect to `/dashboard`, `/admin`, or the appropriate onboarding step.

Never put an access JWT in a cookie or localStorage. Every protected request sends:
`Authorization: Bearer <access-jwt>`.

### 5.2 Issuing and rotating JWTs

- `POST /auth/refresh` reads the refresh cookie, verifies its signature, claims, hash,
  state, and expiry, then rotates it inside one transaction.
- The old token receives `revoked_at` and `replaced_by_hash`; reuse revokes the entire
  `family_id` and returns `401`.
- `POST /auth/logout` revokes the current family and clears refresh/CSRF cookies.
- There is no permanent access-JWT blacklist; its lifetime is 10 minutes. Critical
  incidents may use a short Redis `jti` denylist while PostgreSQL remains refresh truth.
- Never accept role claims from the client; `sub` identifies the user and Casbin/DB sets the role.

Short middleware example:

```go
func RequireAuth(jwtVerifier *auth.JWTVerifier) fiber.Handler {
    return func(c *fiber.Ctx) error {
        raw := strings.TrimSpace(strings.TrimPrefix(c.Get("Authorization"), "Bearer "))
        claims, err := jwtVerifier.VerifyAccess(raw) // alg, signature, iss, aud, exp, nbf, jti
        if err != nil || !uuid.IsValid(claims.Subject) {
            return response.Unauthorized(c)
        }
        c.Locals("user_id", claims.Subject)
        return c.Next()
    }
}
```

Handlers must not use SQLBoiler models directly. JWT issuance and refresh rotation go through
the service/repository, while transactions and uniqueness constraints remain in PostgreSQL.

### 5.3 Middleware and trusted origin

- General middleware order: `requestid → recover → logger → secure headers → host/origin → ratelimit → csrf-for-cookie-routes → jwt → routes`.
- `Origin` and `Host` must match the production allowlist. Do not treat `User-Agent` or
  a custom header as proof that a request came from the frontend.
- CORS must not use `*` with credentials. Unsafe browser requests need a valid `Origin`
  and CSRF header; bearer-only requests must not rely on CSRF alone.
- Reject missing Authorization or multiple schemes and return generic errors without revealing whether a user exists.
- Security headers include `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy`, and CSP.

### 5.4 Authorization and roles with Casbin

Authentication proves identity; Casbin decides whether the user may perform an action on a resource. Never read the role from the frontend, a cookie, or a client-supplied claim.

- Use `github.com/casbin/casbin/v2` with RBAC and `keyMatch2` for API routes.
- PostgreSQL is the production source of truth for role assignments and policies. A custom repository/adapter implements the Casbin adapter with `database/sql` and SQLBoiler; `policy.csv` is for development and tests only.
- Core roles are `user`, `merchant`, and `admin`. Optional `admin_reviewer` and `admin_support` roles must have narrow policies; do not grant admin access from email on every request.
- `ADMIN_EMAILS` is used only to bootstrap the first verified admin; later assignments come from PostgreSQL through an audited admin action.
- The default policy is **deny**. Every admin route declares explicit `object` and `action` permissions; do not use `admin/*` as a blanket exception.
- Use `SyncedEnforcer` or a suitable lock and a Redis watcher such as `dinar:casbin:policy:changed` to reload policies on every API instance.
- Do not cache authorization decisions in Redis without version/revocation; reevaluate after role changes or session revocation.

Suggested Casbin model:

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

Policy examples:

```csv
p, admin, /api/v1/admin/*, (GET|POST|PUT|PATCH|DELETE), allow
p, admin_reviewer, /api/v1/admin/merchants/*, (GET|POST), allow
p, admin_reviewer, /api/v1/admin/offers/*, (GET|POST), allow
p, admin_support, /api/v1/admin/merchants/*, GET, allow
g, <user-uuid>, admin
```

The authorization middleware takes `user.ID` from the access JWT verified by `RequireAuth` and passes `c.Path()` and `c.Method()` to Casbin. On an enforcer or adapter error, return `500` and alert; never turn a policy-loading error into allow.

```go
func RequirePermission(e *casbin.SyncedEnforcer) fiber.Handler {
    return func(c *fiber.Ctx) error {
        user, ok := c.Locals("user").(*auth.User)
        if !ok { return response.Unauthorized(c) }
        allowed, err := e.Enforce(user.ID.String(), c.Path(), c.Method())
        if err != nil { return response.Internal(c, err) }
        if !allowed { return response.Forbidden(c) }
        return c.Next()
    }
}
```

Register routes explicitly:

```go
admin := app.Group("/api/v1/admin", RequireAuth(auth), RequirePermission(enforcer))
admin.Post("/merchants/:id/approve", adminHandler.ApproveMerchant)
```

### 5.5 Rate limits
| Scope | Limit |
|---|---|
| Public per IP | 120/minute |
| `/auth/*` | 20/10 minutes |
| `/offers/:id/report` | 5/hour |
| `/offers/:id/whatsapp` and `/view` | 60/minute (optional per-offer hourly check) |
| File uploads | 30/hour per user |
Read the real IP only from trusted `X-Forwarded-For` supplied by Caddy and enable `EnableTrustedProxyCheck`.

Implement limits through Redis, not counters in one process's memory:

- Window key: `dinar:ratelimit:{scope}:{sha256(ip_hash_salt + ip)}:{window_start}`.
- Use atomic `INCR` + `EXPIRE` through Lua or a transaction so keys cannot remain without expiry.
- Never store raw IP addresses; use the same `IP_HASH_SALT` as logging.
- Authentication, upload, report, and payment routes fail closed when Redis is unavailable; public read routes may degrade temporarily with monitoring.
- Rate limiting is not the only control: authorization, CSRF, and webhook signature verification remain mandatory.

## 6. Merchant module

- Validate phone numbers with `phone.NormalizeDZ(s) (string, error)` and store E.164.
- `PUT /merchant/shop`: create a safe slug from the name, adding a numeric suffix on collision.
- `POST /merchant/submit`: require both steps, all three documents, non-empty `consent_accepted_at`, and `register_expires_on > today`; then set `verification_status='pending'` and create an admin notification job.
- Only an `approved` merchant may create or edit an offer. A `suspended` merchant cannot submit an offer for review.

### 6.1 Document uploads
1. Enforce an 8 MB limit with `BodyLimit` on the route.
2. Read the first 512 bytes and use `http.DetectContentType` for JPEG/PNG/WebP/PDF. Reject other content regardless of extension.
3. Compute `sha256`; use `merchants/{merchant_id}/{type}/{uuid}.{ext}` as the key.
4. Upload to the **private bucket** without a public ACL.
5. Insert the row into `merchant_documents`.
6. Never return `storage_key` to the client. Give an admin a signed URL:
```go
url, _ := minioClient.PresignedGetObject(ctx, privateBucket, key, 5*time.Minute, nil)
audit.Log(ctx, adminID, "document.view", "merchant_document", docID)
```

## 7. Offers module

### 7.1 Service validation rules
- Title is 5–120 characters; description is at most 2,000 raw-text characters. Reject `<` and `>` or escape them on output.
- If both prices exist, require `after < before` and calculate `discount_percent`. Otherwise require a value from 1 to 99.
- Require `starts_at < ends_at ≤ starts_at + 180 days` and `ends_at > now()`.
- The category must be active and the wilaya must exist.
- Active offers must not exceed `max_active_offers`.
- Editing an `approved` or `pending` offer returns it to `pending`, hides it, and clears `reviewed_*`.
- Update `search_text` on every save.

### 7.2 Image upload and processing
```go
// imaging.Process: returns three WebP sizes
func Process(src []byte) (variants map[int][]byte, w, h int, err error) {
    img, err := vips.NewImageFromBuffer(src)       // decodes the image and ignores hidden payloads
    if err != nil { return }
    defer img.Close()
    _ = img.AutoRotate()                          // respects EXIF and then removes it
    w, h = img.Width(), img.Height()
    if w < 400 || h < 400 { return nil, 0, 0, ErrTooSmall }
    variants = map[int][]byte{}
    for _, size := range []int{400, 800, 1600} {
        cp, _ := img.Copy()
        _ = cp.Thumbnail(size, size*4, vips.InterestingNone)   // maximum width = size
        ep := vips.NewWebpExportParams(); ep.Quality = 78; ep.StripMetadata = true
        buf, _, e := cp.ExportWebp(ep)
        cp.Close()
        if e != nil { return nil, 0, 0, e }
        variants[size] = buf
    }
    return
}
```
- Keys are `offers/{offer_id}/{media_id}_{400|800|1600}.webp` in the public bucket.
- Use `Cache-Control: public, max-age=31536000, immutable` because names are unique.
- Delete all three sizes when an image is deleted.
- Processing is synchronous in V1, with at most three concurrent processors through a semaphore; move it to `jobs` if load grows.

### 7.3 Public queries
- Use `public_offers` exclusively (see `database.md`).
- Build filters with SQLBoiler modifiers such as `qm.Where`, `qm.InnerJoin`, `qm.OrderBy`, `qm.Limit`, and `qm.Offset`, using allowlists for sort and fields; **never concatenate user text into SQL**.
- Search or `FOR UPDATE` queries that SQLBoiler cannot express cleanly stay in the repository and use parameterized raw SQL; do not add a second ORM or model generator.
- A parent category includes its children (`category_id = ANY($ids)`), whose IDs may come from the Redis cache.
- Use Redis cache-aside for reference lists (`categories`, `wilayas`) for 10 minutes and offer lists for 60 seconds. Keys include locale, filters, and page.
- On data change, approval, or archive, invalidate affected keys or use a version key. Never expose stale Redis data after a sensitive admin action.
- Use `Cache-Control: public, s-maxage=60, stale-while-revalidate=120` for lists and 60 seconds for details.

### 7.4 Views and clicks
- `POST /offers/:id/view` increments a Redis counter (`INCRBY dinar:stats:views:{yyyy-mm-dd}:{offer_id}`), flushed to PostgreSQL every 30 seconds by the worker and during graceful shutdown.
- Deduplicate with `SET dinar:dedupe:view:{hash(ip+ua+offer)} 1 NX EX 3600`. A Redis error must not block the offer response; record a metric.
- The PostgreSQL flush is idempotent and remains the reporting and billing source of truth. Do not delete the Redis counter until the write succeeds or it is saved to a retry queue.
- A short Redis lock (`SET NX PX`) may prevent duplicate flush workers for the same period.
- `POST /offers/:id/whatsapp` verifies visibility and builds the link:
```go
text := url.QueryEscape(fmt.Sprintf("Hello, I would like to ask about your offer: %s\n%s/offers/%s", o.Title, appURL, o.ID))
link := fmt.Sprintf("https://wa.me/%s?text=%s", strings.TrimPrefix(o.WhatsappE164, "+"), text)
```

## 8. Billing (Chargily + manual)

### 8.1 Create a payment
1. Create a `payments` row (`method=chargily`, `status=pending`, `reference`, `amount=settings.subscription_price_dzd`).
2. Call Chargily to create checkout with amount, `dzd`, frontend `success_url` and `failure_url`, `webhook_endpoint=APP_URL/api/v1/webhooks/chargily`, and `metadata={payment_id, merchant_id}`.
3. Store `provider_checkout_id` and return `checkout_url`.
> **Verify request fields, response fields, and event names in the official Chargily Pay v2 documentation at implementation time.** Do not rely on memory.

### 8.2 Webhook
```go
func (h *Handler) Chargily(c *fiber.Ctx) error {
    body := c.Body()                                   // raw body before parsing
    sig := c.Get("signature")
    mac := hmac.New(sha256.New, []byte(h.secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    if !hmac.Equal([]byte(expected), []byte(sig)) {
        return c.SendStatus(fiber.StatusForbidden)
    }
    var ev chargily.Event
    if err := json.Unmarshal(body, &ev); err != nil { return c.SendStatus(400) }
    fresh, err := h.svc.RecordEvent(c.Context(), "chargily", ev.ID, ev.Type, body) // SQLBoiler + unique constraint
    if err != nil { return c.SendStatus(500) }         // makes Chargily retry
    if !fresh { return c.SendStatus(200) }             // duplicate
    if err := h.svc.HandleEvent(c.Context(), ev); err != nil { return c.SendStatus(500) }
    return c.SendStatus(200)
}
```
- On success, find the payment by `provider_checkout_id`, **compare the amount** with the stored value, and run the activation transaction (see `database.md`).
- On failure, cancellation, or expiry, update status only.
- Never trust `success_url` to activate a subscription.
- A short Redis replay lock (`SETNX` with TTL) is only a performance optimization; PostgreSQL uniqueness and the SQLBoiler transaction are the source of truth for duplicates.
- Do not retain the raw webhook body when it contains unnecessary sensitive data. Store only audit minimums such as `event_id`, `type`, and a body hash when needed.

### 8.3 Manual payment
- `POST /merchant/billing/manual`: upload proof to the private bucket using document rules, create `payments` with `under_review`, and notify admins.
- Admin `approve` runs the same activation transaction; `reject` records the reason and notifies the merchant.

### 8.4 Daily cron job
- At 03:00 `Africa/Algiers`, create reminder notifications for merchants whose `access_until` is within 14/7/1/0 days, using `dedupe_key`.
- Clean expired sessions, completed `jobs`, and Chargily `payments` pending for more than 48 hours.

## 9. Admin module

- Every endpoint under `/admin` passes through `RequireAuth` and Casbin `RequirePermission`; handler role checks are not a replacement for Casbin.
- Read and write actions may need separate policies (`GET` does not automatically imply `POST` or `DELETE`). Bulk actions cannot bypass per-action permission.
- Every sensitive action calls `audit.Log(ctx, actor, action, entityType, entityID, meta)` **inside the same transaction**.
- Role or policy changes require high-level admin permission, record actor/reason/old/new values in `audit_logs`, and publish invalidation through the Redis watcher.
- Merchant approval sets `verification_status='approved'` and `verified_at/by`, and approves all documents. Rejection requires a reason and may identify rejected documents.
- Suspension immediately hides all merchant offers through `public_offers`.
- Bulk offer actions run in one transaction and return a result for each item.

## 10. Background jobs (Worker)

- Keep the `jobs` queue durable in PostgreSQL and claim work inside a SQLBoiler transaction. Because `FOR UPDATE SKIP LOCKED` may not be expressible as a query modifier, parameterized raw SQL is allowed only in this repository:
```sql
UPDATE jobs SET locked_at = now(), attempts = attempts + 1
WHERE id = (SELECT id FROM jobs WHERE done_at IS NULL AND run_at <= now()
            AND (locked_at IS NULL OR locked_at < now() - interval '5 minutes')
            ORDER BY run_at FOR UPDATE SKIP LOCKED LIMIT 1)
RETURNING *;
```
- Redis is optional for a distributed worker lock and short dedupe, such as `dinar:lock:jobs:{type}` with a TTL; it never replaces the `jobs` row.
- Retry with exponential delays (1, 5, 30 minutes) up to five attempts, then keep `last_error`.
- Types include `send_email`, `notify_admin_new_merchant`, and `cleanup_media`.
- Email templates support Arabic RTL with simple HTML and plain text.

## 11. Errors and logs

```go
type AppError struct { Code string; Status int; Message string; Fields map[string]string; Err error }
func (e *AppError) Error() string { return e.Code }
```
- A single `fiber.Config.ErrorHandler` maps `AppError` and PostgreSQL SQLSTATE errors from `pgx/stdlib` (`*pgconn.PgError`: `23505` → `conflict`, `23514` → `validation_failed`) to the standard response.
- Redis timeout and `redis.Nil` errors become defined service degradation; never return connection details or Redis keys to the client.
- JSON logs contain `request_id, method, path, status, latency_ms, user_id`. **Never log** bodies, tokens, complete phone numbers, or document paths.
- WebSocket tests cover JWT/ticket authentication, `Origin`, user isolation, reconnect/resync, heartbeat, connection limits, and Redis outage without losing the PostgreSQL notification row.
- Send 5xx errors to monitoring such as Sentry.

## 12. Testing

- **Unit:** `phone`, `search.Normalize`, discount calculation, `access_until`, date validation, and Chargily signatures.
- **Integration:** use real PostgreSQL, Redis, and MinIO testcontainers, apply all migrations, create buckets, and test:
  - every visibility condition in `public_offers`;
  - first-offer trial activation exactly once under concurrency;
  - duplicate webhook activation producing one subscription;
  - subscription extension without losing trial days;
  - blocked publication by unapproved merchants and cross-merchant document access;
  - cache hit/miss and invalidation after offer approval or edit;
  - rate limit, idempotency, and Redis outage without privilege bypass or duplicate payment;
  - Casbin deny-by-default, role permission matrix, client-role forgery prevention, and Redis policy reload;
  - MinIO image/document uploads, private presigned URLs, anonymous-read rejection, and media cleanup;
  - one-time view-counter flush to PostgreSQL across worker restarts.
- **E2E:** complete merchant and visitor journeys with the frontend.
- Target at least 80% service coverage for critical modules such as offers, billing, and merchants.

## 13. Deployment

- Use a multi-stage `Dockerfile` with libvips in the runtime image and a non-root user.
- `docker compose` services are `postgres`, `redis`, `minio`, `minio-init`, `api`, `worker`, `web`, and `caddy`.
- MinIO exposes S3 internally on `9000` and its console on `9001`; never publish these ports publicly. `minio-init` creates `dinar-public` and `dinar-private` and applies public-read only to the public bucket.
- Check MinIO with `mc alias ls`/`mc ready local` or `BucketExists` through `/readyz`; check Redis with `redis-cli ping`.
- MinIO is durable storage for V1 images and documents. Use a separate volume and off-host backups/replication; never delete the volume during redeployment.
- `Caddyfile`:
```
dinar.dz {
  encode zstd gzip
  handle /api/* { reverse_proxy api:8080 }
  handle        { reverse_proxy web:3000 }
}
```
- Run migrations as a separate step before deploying a new version.
- Graceful shutdown (`SIGTERM`): stop accepting requests, flush Redis counters with retry/lock, wait for active MinIO operations, and close PostgreSQL, Redis, and MinIO clients.

## 14. Backend pull-request checklist

- [ ] No SQL or business rules in handlers.
- [ ] All CRUD uses generated SQLBoiler models/queries; regenerate after every migration and never edit `db/models` manually.
- [ ] PostgreSQL is the source of truth; every Redis key has an owner, TTL, invalidation rule, and outage test.
- [ ] Every route has explicit `RequireAuth` and Casbin permission; default policy is deny.
- [ ] Access JWTs are not accepted from cookies or query strings; refresh rotation and reuse detection are tested.
- [ ] Admin roles are never granted from client claims; role and policy changes are audited.
- [ ] Every input is validated and every error uses the standard response.
- [ ] Public queries use `public_offers`.
- [ ] Sensitive actions run in one transaction with `audit_logs`.
- [ ] Logs contain no secrets or unnecessary personal data.
- [ ] New tests cover the change, and `go vet` and `golangci-lint` are clean.
- [ ] Swagger comments and generated frontend types are updated.
