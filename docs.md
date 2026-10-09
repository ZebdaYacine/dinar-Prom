# Dinar DZ — Comprehensive Project Documentation

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Business Model and Scope](#2-business-model-and-scope)
3. [Technology Stack](#3-technology-stack)
4. [Project Architecture](#4-project-architecture)
5. [Directory Structure](#5-directory-structure)
6. [Users and Domain Models](#6-users-and-domain-models)
7. [Feature Modules](#7-feature-modules)
8. [Pages and Routing](#8-pages-and-routing)
9. [Core UI Infrastructure](#9-core-ui-infrastructure)
10. [Data and Persistence](#10-data-and-persistence)
11. [API Layer](#11-api-layer)
12. [Authentication and Authorization](#12-authentication-and-authorization)
13. [Internationalization, Theme, and Responsive UI](#13-internationalization-theme-and-responsive-ui)
14. [Payments and Billing](#14-payments-and-billing)
15. [Notifications and Background Jobs](#15-notifications-and-background-jobs)
16. [Security and Error Handling](#16-security-and-error-handling)
17. [Testing and Quality](#17-testing-and-quality)
18. [Getting Started and Deployment](#18-getting-started-and-deployment)
19. [Operational Checklist](#19-operational-checklist)

---

## 1. Project Overview

**Dinar DZ** (Dinar) is a bilingual Algerian marketplace for discovering local discounts. Merchants publish time-limited offers with images, prices, location, and WhatsApp contact details. Visitors browse and search without creating an account, then contact the merchant directly.

The platform does not sell products, hold customer funds, or process product orders. Its revenue comes from merchant subscriptions and direct advertising.

The system has three main experiences:

- **Visitor experience** — browse, search, filter, share, report, and contact shops without registration.
- **Merchant experience** — sign in with Google, complete verification, create offers, monitor engagement, and pay for annual access.
- **Administrator experience** — review merchants and offers, manage categories and ads, resolve reports, manage settings, and inspect the audit trail.

Project-level decisions for scope, legal assumptions, providers, release
ownership, and production blockers are recorded in
[`decision-log.md`](decision-log.md).

EPIC-00 delivery evidence is maintained in the
[`project board`](project-board.md), [`P0 release gate`](release-gate.md), and
[`PRD review matrix`](prd/prd-review-matrix.md). The Jira-ready status is in
[`jira.md`](jira.md).

The product is mobile-first because the main audience is expected to use Algerian 3G/4G connections and phones.

---

## 2. Business Model and Scope

### 2.1 Business rules

| Rule | Description |
|---|---|
| Trial | 60 days, starts when the merchant's first offer is approved |
| Subscription | 4,000 DZD for 365 days in V1 |
| Public access | A merchant's offers are public only while verification and access are valid |
| Customer account | Not required in V1 |
| Product payment | Not supported; customers contact merchants directly |
| Merchant payment | Chargily hosted checkout or reviewed manual transfer |
| Offer images | 1–10 images, processed to WebP at 400/800/1600px |
| Supported UI languages | Arabic (`ar`) and French (`fr`) |

### 2.2 In scope for V1

- Public offer catalogue with Arabic search and filtering.
- Merchant onboarding and private document verification.
- Offer creation, image processing, moderation, archive, and duplicate flow.
- Trial, Chargily subscription, manual payment, and payment history.
- Admin review, categories, ads, reports, settings, and audit logs.
- Notification list in the header and email notifications for important status changes.
- Responsive light/dark UI and Arabic/French language switching with `next-intl`.
- Docker deployment with PostgreSQL, Redis, worker, web, API, and Caddy.

### 2.3 Out of scope for V1

- Shopping cart, checkout, delivery, or customer-to-platform product payments.
- Customer accounts, customer wishlists, or customer order history.
- Native iOS/Android applications.
- Public document URLs.
- Video uploads; the database leaves room for a future media type.
- Automatic identity verification by a third party.

---

## 3. Technology Stack

| Layer | Technology | Responsibility |
|---|---|---|
| Web framework | Next.js 15 App Router + React 19 | SSR, pages, dashboard, SEO |
| Language | TypeScript strict mode | Frontend type safety |
| UI | Tailwind CSS 4 + DaisyUI 5 | Responsive design and themes |
| i18n | `next-intl` | Arabic/French messages and locale routing |
| API | Go + Fiber v2 | Business logic and REST API |
| Database | PostgreSQL 16 | Durable source of truth |
| ORM/code generation | SQLBoiler v4 | Generated models and query modifiers |
| Cache/coordination | Redis 7 + `go-redis/v9` | Cache, rate limits, counters, short locks, Casbin invalidation |
| Authorization | Casbin v2 | RBAC and route/action permissions |
| Object storage | MinIO S3 | Public media and private documents |
| Payments | Chargily Pay v2 | Hosted merchant subscription checkout |
| Email | SMTP or Resend | Status, trial, and payment messages |
| Reverse proxy | Caddy | HTTPS and `/api` routing |
| Real-time transport | WebSocket + Redis Pub/Sub | Live user notifications across API replicas |
| Migrations | `golang-migrate` | Versioned PostgreSQL changes |
| Images | libvips via `govips` | Resize, EXIF stripping, WebP conversion |
| Tests | Vitest/Testing Library, Playwright, Go testing, Testcontainers | Unit, integration, and E2E quality |

### Fixed architectural decisions

1. The Go API is the only owner of business rules.
2. The Next.js app never connects directly to PostgreSQL or Redis.
3. PostgreSQL is the source of truth for users, merchants, offers, payments, subscriptions, policies, jobs, and audit logs.
4. Redis is disposable infrastructure; losing Redis must not authorize a request or duplicate a payment.
5. Access JWTs are short-lived and memory-only in the browser; refresh JWTs are
   rotated in `HttpOnly; Secure; SameSite=Strict` cookies and never enter storage.
6. All money is stored as integer DZD, never floating point.
7. All database timestamps are UTC; display time uses `Africa/Algiers`.

---

## 4. Project Architecture

```text
                         Internet
                            |
                       Caddy HTTPS
                 _________/    \_________
                /                       \
       Next.js web                    Go Fiber API
       SSR + React                    /api/v1
                \_______________________/
                   |       |       |       |
             PostgreSQL  Redis   MinIO  Chargily
                   |
             Go background worker ---- SMTP
```

### 4.1 Backend layers

```text
HTTP handler
    -> service/use case
        -> repository interface
            -> SQLBoiler/database/sql
                -> PostgreSQL
```

- Handlers parse input, call services, and return the standard response envelope.
- Services enforce business rules and transaction boundaries.
- Repositories own SQLBoiler and parameterized PostgreSQL queries.
- Middleware owns authentication, Casbin authorization, CSRF, request IDs, rate limits, and recovery.
- Workers process durable PostgreSQL jobs and use Redis only for short coordination.

### 4.2 Frontend layers

```text
Next.js route/layout
    -> page-level server/client component
        -> domain component and hook
            -> API client / TanStack Query
                -> Go API
```

Public pages are server-rendered for SEO. Merchant and admin dashboards use client components behind session guards.

---

## 5. Directory Structure

```text
dinar/
├── apps/
│   ├── web/
│   │   ├── app/
│   │   │   └── [locale]/
│   │   │       ├── (public)/
│   │   │       ├── (auth)/login/
│   │   │       ├── (merchant)/
│   │   │       └── (admin)/admin/
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   ├── layout/
│   │   │   ├── offers/
│   │   │   ├── merchant/
│   │   │   └── admin/
│   │   ├── lib/api/
│   │   ├── lib/auth/
│   │   ├── lib/validation/
│   │   ├── i18n/routing.ts
│   │   ├── i18n/request.ts
│   │   ├── i18n/navigation.ts
│   │   └── messages/ar.json, fr.json
│   └── api/
│       ├── cmd/api/
│       ├── cmd/worker/
│       ├── internal/config/
│       ├── internal/platform/db/
│       ├── internal/platform/cache/
│       ├── internal/platform/authz/
│       ├── internal/http/middleware/
│       ├── internal/auth/
│       ├── internal/merchants/
│       ├── internal/offers/
│       ├── internal/billing/
│       ├── internal/admin/
│       ├── internal/jobs/
│       ├── db/migrations/
│       ├── db/models/               # generated SQLBoiler models
│       ├── db/queries/
│       └── db/sqlboiler.toml
├── messages/ar.json
├── messages/fr.json
├── docker/
├── secrets/README.md
├── skills/
├── prd/
├── redeploy.sh
├── script.sh
├── docker-compose.dev.yml
├── docker-compose.prod.yml
└── docs.md
```

Generated SQLBoiler files are never edited manually. Documentation under `skills/` contains the implementation rules; `prd/` contains product requirements and delivery tasks.

---

## 6. Users and Domain Models

### 6.1 Visitor

An unauthenticated visitor can:

- view public offers and shop pages;
- search and filter by text, category, wilaya, discount, expiry, price, and sort order;
- share an offer;
- open WhatsApp or phone contact;
- report an offer.

No visitor profile is stored in V1. Anonymous analytics use salted hashes and aggregated counters only.

### 6.2 Merchant

```ts
type Merchant = {
  id: string;
  userId: string;
  shopName: string | null;
  slug: string | null;
  verificationStatus:
    | 'incomplete' | 'pending' | 'approved' | 'rejected' | 'suspended';
  wilayaId: number | null;
  categoryId: number | null;
  accessUntil: string | null;
  trialEndsAt: string | null;
};
```

The merchant can draft offers before approval but cannot publish or submit them until verification is complete.

### 6.3 Offer

```ts
type Offer = {
  id: string;
  merchantId: string;
  categoryId: number;
  wilayaId: number;
  title: string;
  description: string;
  priceBeforeDzd: number | null;
  priceAfterDzd: number | null;
  discountPercent: number;
  startsAt: string;
  endsAt: string;
  status: 'draft' | 'pending' | 'approved' | 'rejected' | 'archived' | 'removed';
  media: OfferMedia[];
};
```

An approved offer becomes hidden and returns to `pending` after any material edit.

### 6.4 Payment and access

The `payments` record tracks the provider transaction. The `subscriptions` record tracks the access period. The merchant's `access_until` is updated transactionally and is used by the `public_offers` view.

---

## 7. Feature Modules

### 7.1 Public catalogue

- Home page with latest, expiring-soon, featured, and highest-discount sections.
- Offer list with URL-based search and filters.
- Offer detail with image gallery, countdown, shop card, share, WhatsApp, phone, and report actions.
- Shop page with verified merchant identity and current offers.
- Empty, loading, 404, 410, and error states.

### 7.2 Merchant onboarding

Three autosaved steps:

1. Personal identity and phone.
2. Shop identity, category, location, contacts, and commercial register.
3. Three private documents, consent, and submission.

Status changes are visible to the merchant and generate notifications/email where configured.

### 7.3 Offer management

- Draft create/edit and autosave.
- Price or discount validation.
- Up to 10 compressed images.
- Image reorder, cover selection, and delete.
- Submit for moderation.
- Archive, suspend, duplicate, and view statistics.

### 7.4 Billing

- Annual Chargily checkout.
- Manual payment proof upload.
- Payment list and detail status.
- Subscription/access countdown.
- Trial and expiry reminder notifications.

### 7.5 Administration

- Dashboard metrics.
- Merchant/document review.
- Offer moderation and bulk actions.
- Payment/manual proof review.
- Category and ad management.
- Report resolution.
- Site settings, notification review, Casbin roles/policies, and audit logs.

---

## 8. Pages and Routing

All routes use a locale prefix: `/ar/...` or `/fr/...`. `/` redirects to `/ar`.

| Route | Access | Description |
|---|---|---|
| `/{locale}` | Public | Home, search entry, showcase, splash fallback |
| `/{locale}/offers` | Public | Search, filters, sorting, pagination |
| `/{locale}/offers/{id}` | Public | Offer details and actions |
| `/{locale}/shop/{slug}` | Public | Shop profile and active offers |
| `/{locale}/about` | Public | About Dinar |
| `/{locale}/terms` | Public | Terms |
| `/{locale}/privacy` | Public | Privacy and document policy |
| `/{locale}/contact` | Public | Contact information |
| `/{locale}/login` | Public | Google sign-in |
| `/{locale}/onboarding/*` | Merchant | Personal, shop, documents, pending, rejected |
| `/{locale}/dashboard` | Merchant | Offer and access overview |
| `/{locale}/dashboard/offers` | Merchant | Offer list |
| `/{locale}/dashboard/offers/new` | Merchant | New offer |
| `/{locale}/dashboard/offers/{id}/edit` | Merchant | Edit offer |
| `/{locale}/dashboard/billing` | Merchant | Checkout, payments, access |
| `/{locale}/dashboard/profile` | Merchant | Merchant/shop profile |
| `/{locale}/admin` | Admin | Admin dashboard |
| `/{locale}/admin/merchants` | Admin | Verification queue |
| `/{locale}/admin/offers` | Admin | Offer moderation queue |
| `/{locale}/admin/payments` | Admin | Manual/payment review |
| `/{locale}/admin/categories` | Admin | Category CRUD |
| `/{locale}/admin/ads` | Admin | Advertisement CRUD |
| `/{locale}/admin/reports` | Admin | Report queue |
| `/{locale}/admin/settings` | Admin | Settings and policies |
| `/{locale}/admin/audit-logs` | Admin | Append-only audit view |

### Navigation flows

```text
Visitor /ar
   ├── search/filter ──▶ /ar/offers ──▶ /ar/offers/:id
   ├── shop card ──────▶ /ar/shop/:slug
   └── join merchant ──▶ /ar/login ──▶ /ar/onboarding/*

Merchant login
   ├── incomplete ─────▶ /:locale/onboarding/...
   ├── pending ────────▶ /:locale/onboarding/pending
   └── approved ───────▶ /:locale/dashboard

Admin login ───────────▶ /:locale/admin
```

---

## 9. Core UI Infrastructure

### Header

The header contains:

- Dinar logo and brand identity.
- Search entry on public pages.
- Notification list with unread count.
- `LanguageSwitcher` using `next-intl`, Arabic/Algerian flag and French flag.
- Light/dark theme toggle.
- Login, profile, or dashboard action.

The homepage follows an editorial, reference-inspired composition: a thin announcement strip, a quiet navigation bar, a split hero with a cream copy panel and a full-bleed Algerian offer/shop image, then curated offer sections. The composition is adapted to Dinar's burgundy, green, cream, gold, and discount-orange identity; it does not reuse the reference brand, copy, logo, or assets. The detailed visual rules live in [`skills/ui.md`](skills/ui.md).

### Language switcher

```tsx
const nextLocale = locale === 'ar' ? 'fr' : 'ar';
router.replace({ pathname, query }, { locale: nextLocale });
```

The switcher preserves the current page and query string. It changes document `lang` and `dir` through the `[locale]` layout.

### Theme

- DaisyUI themes are `dinar` and `dinar-dark`.
- Persist the preference under `dinar_theme`.
- Apply the theme before first paint to avoid a flash.
- Respect `prefers-reduced-motion`.

### Splash screen

The splash screen displays the Dinar logo, a proverb selected from the configured data source, and a loading progress bar. It remains visible for approximately three seconds after the initial page load while never blocking server rendering or hiding a fatal error.

### Responsive behavior

| Surface | Mobile | Tablet | Desktop |
|---|---|---|---|
| Offer grid | 2 columns | 3 columns | 4 columns |
| Filters | Bottom drawer and chips | Side sheet | Persistent sidebar |
| Offer detail | Sticky WhatsApp bar | Inline action bar | Detail sidebar |
| Merchant form | Single column | Two logical groups | Form + live preview |
| Admin tables | Cards/compact rows | Horizontal scroll | Full data table |

Use logical CSS properties (`ms`, `me`, `ps`, `pe`, `start`, `end`) so Arabic RTL and French LTR remain correct.

---

## 10. Data and Persistence

The canonical schema is documented in [`skills/database.md`](skills/database.md). Core tables are:

- reference: `wilayas`, `categories`, `settings`;
- identity: `users`, `refresh_tokens`, `merchants`, `merchant_documents`;
- offers: `offers`, `offer_media`;
- billing: `subscriptions`, `payments`, `webhook_events`;
- engagement: `reports`, `ad_stats_daily`, `offer_stats_daily`, `notifications`;
- operations: `jobs`, `audit_logs`, `casbin_rule`.

### Visibility rule

All public offer reads use `public_offers`, which requires:

```text
offer approved
AND offer currently active
AND merchant approved
AND merchant access_until > now()
```

No frontend or handler may recreate this predicate independently.

### SQLBoiler workflow

```bash
sqlboiler psql --config apps/api/db/sqlboiler.toml
```

Generated models are regenerated after migrations, reviewed in code review, and never edited manually. Complex PostgreSQL operations such as `FOR UPDATE SKIP LOCKED` may use parameterized SQL inside a repository.

### Redis boundary

Redis handles:

- reference and short-lived public caches;
- atomic rate windows;
- view/click counters before flush;
- short locks and deduplication;
- Casbin policy change notifications.

Redis does not replace PostgreSQL for payments, refresh tokens, subscriptions,
roles, audit logs, or jobs.

---

## 11. API Layer

### 11.1 Conventions

- Base path: `/api/v1`.
- JSON and UTF-8.
- UUID identifiers.
- `snake_case` JSON fields.
- Pagination: `page` and `per_page`, maximum 48.
- Dates: ISO 8601 UTC.
- Success envelope: `{ "data": ..., "meta": ... }`.
- Error envelope:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Check the fields",
    "fields": { "title": "Minimum 5 characters" },
    "request_id": "..."
  }
}
```

Every response carries or logs `X-Request-ID`.

### 11.2 Public endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/offers` | Search/filter/paginate visible offers |
| `GET` | `/offers/:id` | Offer details |
| `GET` | `/offers/:id/similar` | Similar visible offers |
| `POST` | `/offers/:id/view` | Deduplicated view counter |
| `POST` | `/offers/:id/whatsapp` | Record click and return `wa.me` link |
| `POST` | `/offers/:id/report` | Submit a report |
| `GET` | `/shops/:slug` | Shop and active offers |
| `GET` | `/categories` | Category reference data |
| `GET` | `/wilayas` | Wilaya reference data |
| `GET` | `/ads` | Select an active ad |
| `POST` | `/ads/:id/click` | Record ad click |
| `GET` | `/sitemap/offers` | Sitemap offer URLs |

### 11.3 Authentication and account endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/auth/google/login` | Start Google OAuth with state and PKCE |
| `GET` | `/auth/google/callback` | Verify Google identity and issue JWT refresh cookie |
| `POST` | `/auth/refresh` | Rotate refresh JWT and return short-lived access JWT |
| `POST` | `/auth/ws-ticket` | Create one-time HttpOnly ticket for notifications WebSocket |
| `POST` | `/auth/logout` | Revoke current refresh-token family |
| `GET` | `/me` | Current user, merchant status, and access |
| `GET` | `/me/notifications` | Notification list and unread count |
| `POST` | `/me/notifications/:id/read` | Mark one notification read |
| `POST` | `/me/notifications/read-all` | Mark all notifications read |

### 11.4 Live notification channel

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` (WebSocket upgrade) | `/api/v1/ws/notifications` | Stream notifications for the authenticated user |

The API persists every notification in PostgreSQL first, then publishes a small
`notification.created` event to Redis Pub/Sub. Each API instance forwards the
event only to sockets belonging to the event's server-side `user_id`. Redis is
transport only; reconnecting clients resync through `/me/notifications` so an
outage or deploy cannot lose durable notifications.

The browser first calls `/auth/ws-ticket` with its in-memory access JWT. The API
sets a short-lived, one-use `HttpOnly` ticket cookie; the browser then opens the
WebSocket without putting a JWT, ticket, or `user_id` in the URL. Caddy proxies
the upgrade to `api:8080`; the backend validates `Origin`, consumes the ticket,
limits connections per user, sends ping/pong heartbeats, caps event size, and
closes stale connections.

The frontend notification center loads the initial list through REST, listens
for live events, updates the unread badge, uses REST for read/read-all actions,
and reconnects with exponential backoff plus a REST resync after reconnect.

### 11.5 Merchant endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `PUT` | `/merchant/profile` | Save personal onboarding step |
| `PUT` | `/merchant/shop` | Save shop information |
| `POST` | `/merchant/documents` | Upload private document |
| `POST` | `/merchant/submit` | Submit merchant review |
| `GET/POST` | `/merchant/offers` | List/create offers |
| `GET/PUT/DELETE` | `/merchant/offers/:id` | Read/edit/delete draft |
| `POST` | `/merchant/offers/:id/media` | Upload image |
| `PATCH` | `/merchant/offers/:id/media/order` | Reorder images |
| `DELETE` | `/merchant/offers/:id/media/:mid` | Delete image |
| `POST` | `/merchant/offers/:id/submit` | Submit offer review |
| `POST` | `/merchant/offers/:id/archive` | Archive offer |
| `POST` | `/merchant/offers/:id/duplicate` | Create a draft copy |
| `GET` | `/merchant/stats` | Merchant engagement summary |
| `POST` | `/merchant/billing/checkout` | Create Chargily checkout |
| `POST` | `/merchant/billing/manual` | Upload manual proof |
| `GET` | `/merchant/billing/payments` | Payment history |
| `GET` | `/merchant/billing/payments/:id` | Payment state |

### 11.6 Admin endpoints

All admin routes require `RequireAuth` and a Casbin permission.

| Method | Endpoint | Permission area |
|---|---|---|
| `GET` | `/admin/dashboard` | Metrics |
| `GET` | `/admin/merchants` | Review queue |
| `GET` | `/admin/merchants/:id` | Merchant details |
| `GET` | `/admin/documents/:id/url` | Signed private document URL |
| `POST` | `/admin/merchants/:id/approve` | Merchant approval |
| `POST` | `/admin/merchants/:id/reject` | Merchant rejection |
| `POST` | `/admin/merchants/:id/suspend` | Suspension |
| `POST` | `/admin/merchants/:id/reactivate` | Reactivation |
| `GET` | `/admin/offers` | Offer review queue |
| `POST` | `/admin/offers/:id/approve` | Offer approval |
| `POST` | `/admin/offers/:id/reject` | Offer rejection |
| `POST` | `/admin/offers/:id/remove` | Removal |
| `POST` | `/admin/offers/bulk` | Bulk moderation |
| `GET/POST` | `/admin/payments` | Manual/payment review |
| `POST` | `/admin/payments/:id/approve` | Activate access |
| `POST` | `/admin/payments/:id/reject` | Reject payment |
| `CRUD` | `/admin/categories` | Category management |
| `CRUD` | `/admin/ads` | Advertisement management |
| `GET` | `/admin/reports` | Report queue |
| `POST` | `/admin/reports/:id/resolve` | Resolve report |
| `GET/PUT` | `/admin/settings` | Settings |
| `GET` | `/admin/audit-logs` | Audit history |
| `GET/POST` | `/admin/authz/policies` | Casbin policy management |

### 11.7 System endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/webhooks/chargily` | Signed payment provider webhook |
| `GET` | `/healthz` | Process is alive |
| `GET` | `/readyz` | PostgreSQL, Redis, and storage readiness |

---

## 12. Authentication and Authorization

### 12.1 Google OAuth and JWT

1. The API generates `state` and PKCE `code_verifier`.
2. Both are stored in a short-lived signed `HttpOnly` cookie.
3. Google returns an authorization code.
4. The API validates state, exchanges the code, verifies issuer/audience, and requires `email_verified`.
5. The API upserts the user by `google_sub`.
6. The API creates a short-lived access JWT and a rotated refresh JWT. Only the
   refresh token hash, `jti`, and family metadata are stored in PostgreSQL.
7. The refresh JWT is sent as `__Host-dinar-refresh` with `HttpOnly`, `Secure`,
   `SameSite=Strict`, and a 30-day absolute expiry. The access JWT is returned by
   `/auth/refresh` and kept in browser memory only.
8. Refresh reuse revokes the whole token family; logout revokes the family and
   clears cookies. No token is written to localStorage, sessionStorage, or URLs.

### 12.2 Request authentication and origin policy

Protected API requests require `Authorization: Bearer <access-jwt>`. The API
validates the configured algorithm, signature, issuer, audience, subject, time
claims, `jti`, and `token_use=access`. It ignores role claims from the client and
asks Casbin/PostgreSQL for authorization. Production exposes only Caddy; the API,
database, Redis, and MinIO have no public host ports. Caddy, CORS, `Origin`, CSRF,
rate limits, CSP, and exact host allowlists reject requests that do not belong to
the approved web origin. A spoofable `User-Agent` or custom frontend header is
never treated as proof that a request came from the frontend.

### 12.3 Casbin authorization

Authentication and authorization are separate. Casbin evaluates:

```text
subject = authenticated user ID
object  = request route
action  = HTTP method
```

Example policy:

```csv
p, admin, /api/v1/admin/*, (GET|POST|PUT|PATCH|DELETE), allow
p, admin_reviewer, /api/v1/admin/merchants/*, (GET|POST), allow
p, admin_support, /api/v1/admin/merchants/*, GET, allow
g, <user-uuid>, admin
```

Authorization is deny-by-default. Roles are not read from request bodies, browser
storage, or untrusted client claims. PostgreSQL stores persistent policy rows;
Redis publishes policy invalidation events.

### 12.4 Merchant ownership

Every merchant repository method checks the authenticated merchant ID. A merchant can access only its own profile, documents, offers, payments, and statistics. Admin document access creates a signed URL for five minutes and an audit record.

---

## 13. Internationalization, Theme, and Responsive UI

### 13.1 `next-intl`

```ts
export const routing = defineRouting({
  locales: ['ar', 'fr'],
  defaultLocale: 'ar',
  localePrefix: 'always',
});
```

- Messages live in `messages/ar.json` and `messages/fr.json`.
- Both files must have identical leaf keys.
- Server components use `getTranslations`; client components use `useTranslations`.
- Navigation uses the generated `Link`, `useRouter`, and `usePathname` from `i18n/navigation.ts`.
- The root redirect is `/` → `/ar`.
- The `[locale]` layout sets `lang` and `dir`: Arabic is RTL; French is LTR.
- `generateMetadata` creates localized titles, descriptions, canonical URLs, and `hreflang` links.

### 13.2 Translation rules

Never hardcode user-facing strings inside components. Use namespaces such as `common`, `header`, `notifications`, `offers`, `auth`, `merchant`, `billing`, `admin`, `splash`, `errors`, and `footer`.

API data must provide locale-aware fields such as `name_ar` and `name_fr`, or use a documented locale parameter. Do not silently mix Arabic and French fallback content.

### 13.3 Theme and icon rules

- Use local SVG/logo assets for brand identity and flags.
- Use `lucide-react` for standard actions.
- Every icon-only action has an accessible label.
- Light and dark themes use the same semantic tokens, not duplicated arbitrary colors.

---

## 14. Payments and Billing

### 14.1 Hosted checkout

The platform never receives card details. The merchant is redirected to Chargily's hosted checkout. Dinar stores only the internal payment reference, provider checkout ID, amount, currency, status, and safe provider metadata.

Client request:

```json
{ "plan": "annual" }
```

The API calculates the amount from server settings. It never trusts an amount or currency supplied by the browser.

### 14.2 Webhook processing

```text
raw body
  -> verify provider signature
  -> validate event schema and livemode
  -> insert provider/event_id idempotency record
  -> lock payment and merchant
  -> compare amount/currency/checkout
  -> activate subscription in one transaction
  -> audit and notify
```

The success redirect only displays status. It never activates access. A duplicate webhook returns success without applying the business operation twice.

### 14.3 Manual payment

Manual payment proof is uploaded to the private bucket, linked to a payment in `under_review`, and reviewed by an authorized admin. Approval uses the same subscription activation transaction as Chargily.

### 14.4 Payment restrictions

Never store or log:

- full card number;
- CVV/CVC, PIN, OTP, or expiry;
- screenshots of a payment page;
- raw sensitive provider payloads;
- provider secrets or API keys.

---

## 15. Notifications and Background Jobs

### Notifications

The header notification list reads from `/me/notifications`. Notifications are deduplicated by `(user_id, dedupe_key)` and support read-one/read-all actions.

Important events include:

- merchant submitted, approved, rejected, suspended, or reactivated;
- offer approved, rejected, removed, or reported;
- trial ending in 14, 7, 1, or 0 days;
- payment pending, paid, failed, or rejected;
- admin review required.

### Worker jobs

The worker claims jobs from PostgreSQL using `FOR UPDATE SKIP LOCKED` and retries with exponential delays. Job types include:

- `send_email`;
- `notify_admin_new_merchant`;
- `notify_offer_status`;
- `cleanup_media`;
- `flush_offer_stats`;
- `expire_pending_payment`;
- `create_access_expiry_notifications`.

Redis may provide a short worker lock, but the durable job row remains in PostgreSQL.

---

## 16. Security and Error Handling

The complete token, browser, origin, input-validation, WebSocket, and incident
response rules are maintained in [`skills/security.md`](/home/zedyacine/Desktop/dinar/skills/security.md).

### Security controls

- HTTPS is mandatory in production.
- Caddy is the only public container.
- CSRF checks compare `Origin`/`Referer` to `APP_URL` for cookie-authenticated
  endpoints such as refresh and logout; Bearer-authenticated API calls still
  require exact origin/CORS policy where applicable.
- Access JWTs are short-lived and memory-only in the browser. Refresh JWTs are
  rotated, stored only in an HttpOnly cookie, and hashed in PostgreSQL.
- The production API is reachable only through Caddy; the API rejects invalid
  JWTs, origins, hosts, CSRF tokens, content types, body sizes, and rate limits.
- Rate limits use Redis atomic counters and salted IP hashes.
- Private documents use private storage and five-minute signed URLs.
- Uploads use magic-byte MIME detection, size limits, SHA-256, and safe generated keys.
- Images are decoded and re-encoded with libvips; EXIF is stripped.
- No secrets are exposed through `NEXT_PUBLIC_*` or logs.
- Admin actions, policy changes, document views, payment decisions, and merchant status changes are audited.

### Standard error codes

| Code | HTTP | Meaning |
|---|---:|---|
| `unauthorized` | 401 | No valid session |
| `forbidden` | 403 | Authenticated JWT subject lacks Casbin permission |
| `not_found` | 404 | Resource does not exist |
| `gone` | 410 | Offer existed but is no longer available |
| `conflict` | 409 | Unique/state conflict |
| `validation_failed` | 422 | Invalid input |
| `rate_limited` | 429 | Rate limit exceeded |
| `internal` | 500 | Unexpected server error |

Internal details go to structured logs and monitoring only. Responses contain the request ID and safe localized messages.

---

## 17. Testing and Quality

### Backend tests

- Unit: phone normalization, Arabic search normalization, discount calculations, date rules, webhook signature verification, and Casbin enforcement.
- Integration: PostgreSQL and Redis Testcontainers with all migrations.
- Transaction tests: first approved offer starts trial once; repeated webhook activates once; concurrent subscription updates do not lose days.
- Security tests: merchant ownership, private document access, deny-by-default Casbin, rate-limit failure modes, upload MIME checks, and no card data persistence.

### Frontend tests

- Unit: formatting, countdown, locale message loading, and validation schemas.
- Component: offer cards, filters, gallery, uploader, notification center, theme toggle, and language switcher.
- E2E: visitor search/detail/WhatsApp/report, merchant onboarding/offer/payment, admin review, Arabic RTL, French LTR, light/dark mode, and mobile 360px layouts.
- Accessibility: automated `axe` checks and keyboard navigation for header, forms, dialogs, and tables.

### Quality gates

```text
go test ./...
go vet ./...
golangci-lint run
pnpm lint
pnpm typecheck
pnpm test
pnpm test:e2e
```

Target coverage is at least 80% for critical backend services and all critical payment/auth/authorization paths.

---

## 18. Getting Started and Deployment

### Prerequisites

- Docker Engine and Docker Compose v2.
- Go toolchain for backend development.
- Node.js and pnpm for frontend development.
- PostgreSQL 16 and Redis 7, or the Docker services.
- Google OAuth credentials, MinIO credentials, Chargily test credentials, and SMTP credentials.

### Environment

Use the root [`.env.example`](/home/zedyacine/Desktop/dinar/.env.example) as the
single shared template for Compose, API, worker, web, and MinIO initialization.
Copy it to `.env` locally; use a secret manager or a protected
`.env.production` in production. Never expose backend secrets through
`NEXT_PUBLIC_*`.

```dotenv
APP_ENV=development
APP_URL=http://localhost:8180
API_ADDR=:8080
DATABASE_URL=postgres://dinar:dinar@postgres:5432/dinar?sslmode=disable
REDIS_URL=redis://redis:6379/0
REDIS_PREFIX=dinar:
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8180/api/v1/auth/google/callback
# If web runs inside Docker use http://api:8080; if Next.js runs on the host use http://localhost:8188.
API_INTERNAL_URL=http://localhost:8188
JWT_ISSUER=dinar-api
JWT_AUDIENCE=dinar-web
JWT_ALGORITHM=RS256
JWT_ACCESS_TTL_SECONDS=600
JWT_REFRESH_TTL_SECONDS=2592000
JWT_PRIVATE_KEY_FILE=/run/secrets/jwt_private_key
JWT_PUBLIC_KEY_FILE=/run/secrets/jwt_public_key
REFRESH_TOKEN_COOKIE_NAME=dinar_refresh
COOKIE_SECURE=false
CSRF_COOKIE_NAME=dinar_csrf
CSRF_HEADER_NAME=X-CSRF-Token
ALLOWED_ORIGINS=http://localhost:8180,http://localhost:3143
TRUSTED_HOSTS=localhost,127.0.0.1
NEXT_PUBLIC_WS_URL=
NOTIFICATIONS_WS_PATH=/api/v1/ws/notifications
NOTIFICATIONS_WS_ENABLED=true
REDIS_CHANNEL_NOTIFICATIONS=dinar:notifications
MINIO_ENDPOINT=minio:9000
MINIO_REGION=us-east-1
MINIO_USE_SSL=false
MINIO_ACCESS_KEY=...
MINIO_SECRET_KEY=...
MINIO_BUCKET_PUBLIC=dinar-public
MINIO_BUCKET_PRIVATE=dinar-private
CHARGILY_BASE_URL=https://pay.chargily.net/test/api/v2
CHARGILY_API_KEY=...
CHARGILY_SECRET=...
SMTP_HOST=...
SMTP_USER=...
SMTP_PASS=...
IP_HASH_SALT=...
```

### First run

```bash
cp .env.example .env
openssl genrsa -out secrets/jwt_private.pem 4096
openssl rsa -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem
chmod 600 secrets/jwt_private.pem
chmod 644 secrets/jwt_public.pem
docker compose --env-file .env -f docker-compose.dev.yml up -d postgres redis minio minio-init api
docker compose --env-file .env -f docker-compose.dev.yml run --rm api go run ./cmd/api migrate up
cd apps/web
pnpm install
pnpm dev -- --port 3143
```

If you run migration or SQLBoiler tools on the host instead of inside the API
container, use `localhost:5543` for PostgreSQL and `localhost:9100` for MinIO in
that tool environment. Inside Docker, use internal service names and ports such
as `postgres:5432` and `minio:9000`.

### Local ports on the development machine

To avoid the existing Aisha, Hodhod, 50lab, and Northstar containers, Dinar uses:

| Service | Host port | Container port |
|---|---:|---:|
| Caddy HTTP | `8180` | `80` |
| Caddy HTTPS | `8443` | `443` |
| Next.js web | `3143` | `3000` |
| Go API | `8188` | `8080` |
| PostgreSQL | `5543` | `5432` |
| Redis | `6381` | `6379` |
| MinIO API | `9100` | `9000` |
| MinIO Console | `9101` | `9001` |

Open the local site at `http://localhost:8180` when Caddy is running, or
`http://localhost:3143` when Next.js is running directly on the host. Containers
continue to use service DNS and internal ports, for example `postgres:5432`,
`redis:6379`, and `minio:9000`.

When Caddy is the entry point, WebSocket notifications use
`ws://localhost:8180/api/v1/ws/notifications` locally and
`wss://dinar.dz/api/v1/ws/notifications` in production. Do not expose a separate
WebSocket port; reuse the authenticated HTTPS origin.

### Production deployment

1. Copy the protected production environment file and set `SHARED_ENV_FILE=.env.production`.
2. Build versioned API, worker, and web images with `docker-compose.prod.yml`.
3. Validate `docker compose --env-file .env.production -f docker-compose.prod.yml config --quiet`.
4. Run PostgreSQL migrations as a separate step with `--entrypoint /app/dinar-api api migrate up`.
5. Start API, worker, web, Redis, PostgreSQL, and Caddy.
6. Verify `/healthz`, `/readyz`, OAuth callback, storage access, WebSocket notifications, and Chargily test webhook.
7. Run smoke tests before opening traffic.

For repeatable redeployment:

```bash
chmod +x redeploy.sh
./redeploy.sh
```

The script must not delete database volumes or automatically reverse migrations. See [`skills/deployment.md`](skills/deployment.md) for the full plan.

---

## 19. Operational Checklist

- [ ] PostgreSQL backups are encrypted, tested, and stored outside the host.
- [ ] Redis is configured with health checks and is not the only store for business data.
- [ ] All production secrets are outside Git and Docker images.
- [ ] `healthz` and `readyz` pass before traffic is enabled.
- [ ] Caddy exposes only ports 80/443; PostgreSQL and Redis are private.
- [ ] `public_offers` remains the only public visibility source.
- [ ] SQLBoiler models match the latest migration.
- [ ] Casbin has deny-by-default policies and an audited bootstrap admin.
- [ ] Chargily webhook signature and idempotency tests pass.
- [ ] No card information, raw document keys, cookies, or secrets appear in logs.
- [ ] Arabic/French messages have identical keys.
- [ ] Responsive tests pass at 360px, tablet, and desktop widths.
- [ ] Rollback uses a known previous image tag and does not blindly reverse migrations.
