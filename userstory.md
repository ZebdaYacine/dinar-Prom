# Dinar DZ — User Stories and Technical Specification

> Version: 1.0 — Status: Approved draft for implementation
> This is the **highest-level reference**. `frontend.md`, `backend.md`, `database.md`, and `ui.md` detail its parts; when documents conflict, this document wins.

---

## 1. Vision

**Dinar DZ** is an open Algerian platform that gathers local shop and merchant discounts in one place. Merchants publish their own offers (images, price, and duration); visitors browse and search without registration and contact merchants directly through WhatsApp. The platform displays and redirects traffic; it does not sell products or collect product payments.

**Revenue:** a yearly merchant subscription (4,000 DZD) after a two-month free trial, plus direct advertising.

**Design principle:** mobile-first speed for 3G/4G visitors, trust through document-verified merchants, and simple publishing.

### 1.1 Out of scope for V1
- No shopping cart or product payments inside the platform.
- No video uploads; the database leaves room for a future media type.
- No customer accounts; browsing is open.
- No native mobile app; the website is PWA-ready.

---

## 2. Terms

| Term | Meaning |
|---|---|
| Visitor | Anyone browsing without an account |
| Merchant | A verified shop or brand owner who publishes offers |
| Admin | System administrator who reviews and decides approvals |
| Offer | One time-limited discount on a product or service, with images |
| Trial | 60 free days beginning when the merchant's first offer is approved |
| Access | The right for a merchant's offers to appear publicly through a valid trial or subscription |
| Subscription | 4,000 DZD for one year (365 days) |

---

## 3. Roles and permissions

| Action | Visitor | Merchant | Admin |
|---|:--:|:--:|:--:|
| Browse and search offers | ✅ | ✅ | ✅ |
| Open merchant WhatsApp | ✅ | ✅ | ✅ |
| Report an offer | ✅ | ✅ | ✅ |
| Create/edit own offers | ❌ | ✅ after verification | ✅ |
| Upload activation documents | ❌ | ✅ | ❌ |
| Approve/reject merchants and offers | ❌ | ❌ | ✅ |
| View sensitive documents | ❌ | Own documents only | ✅ with audit record |
| Manage categories, ads, and settings | ❌ | ❌ | ✅ |

---

## 4. System architecture

```
                    ┌────────────────────────────┐
   Browser/phone ─▶│  Caddy (HTTPS, reverse proxy)│
                    └──────┬──────────────┬──────┘
                    /      │              │   /api/*
                           ▼              ▼
                ┌────────────────┐  ┌──────────────────┐
                │ Next.js (web)  │  │ Go Fiber (api)   │
                │ SSR + React    │─▶│ REST /api/v1     │
                └────────────────┘  └───┬────┬────┬────┘
                                        │    │    │
                       ┌────────────────┘    │    └───────────────┐
                       ▼                     ▼                    ▼
               ┌──────────────┐     ┌────────────────┐    ┌───────────────┐
                       │ PostgreSQL 16│     │ S3 storage (R2)│    │ Chargily Pay  │
                       │ + jobs table │     │ public + private│    │ (webhook)     │
               └──────────────┘     └────────────────┘    └───────────────┘
                       ▲
               ┌───────┴────────┐        ┌───────────────┐
               │ worker (cron)  │───────▶│ Email (SMTP)  │
               └────────────────┘        └───────────────┘
```

### 4.1 Components
| Component | Technology | Responsibility |
|---|---|---|
| web | Next.js 15 App Router + TypeScript + Tailwind 4 + DaisyUI 5 | Public UI, merchant dashboard, and admin dashboard |
| api | Go + Fiber v2 | Business logic, JWT auth, authorization, uploads, and payments |
| worker | Same binary in `worker` mode | Expiration, reminders, and cleanup jobs |
| db | PostgreSQL 16 | Durable data, full-text search, and job queue |
| storage | MinIO locally or Cloudflare R2 | Public image bucket and **private** document bucket |
| proxy | Caddy | Automatic TLS, `/api` to Go, everything else to Next |
| payments | Chargily Pay v2 | Electronic payments (CIB / EDAHABIA) |

### 4.2 Fixed architectural decisions
1. **Single source of truth:** all business logic lives in Go; Next.js never connects directly to the database.
2. **JWT authentication:** Go handles Google OAuth and issues memory-only access JWTs plus rotated HttpOnly refresh cookies.
3. **Same origin:** use `dinar.dz` for the UI and `dinar.dz/api` for the API to minimize CORS and cookie problems.
4. **Integer money:** store DZD as integer values with no fractions.
5. **Time:** store all timestamps as UTC `timestamptz` and display them in `Africa/Algiers`.
6. **Images:** process images on the server into three WebP sizes; never publish private documents.
7. **Public visibility** is computed by one database rule in one place; see section 6.1.

---

## 5. User stories

> Format: **As a** [role] **I want** [action] **so that** [benefit], followed by acceptance criteria (AC).
> Priority: P0 = required for launch, P1 = immediately after launch, P2 = later.

### Epic 1 — Offer discovery (visitor)

**US-V01 — Home page (P0)**
As a visitor, I want to see the latest and soon-to-expire offers so that I can benefit quickly.
- AC1: The page shows a search bar, main categories, "Ending soon", "Latest offers", and "Highest discount".
- AC2: Each offer card shows cover image, title, before/after prices, percentage badge, shop name, wilaya, and remaining-time counter.
- AC3: Expired, unapproved offers and offers from merchants without valid access never appear.
- AC4: The first page reaches LCP < 2.5s on a typical 4G connection.

**US-V02 — Search and filtering (P0)**
As a visitor, I want to search and filter results so that I can find what I need.
- AC1: Arabic text search ignores diacritics and normalizes common letter variants.
- AC2: Filters include wilaya, category and subcategories, minimum discount, and "ending within 24 hours / 7 days".
- AC3: Sort by newest, ending soon, highest discount, or lowest price.
- AC4: Filters are stored in the query string for easy sharing.
- AC5: The no-results state suggests removing a filter or changing wilaya.
- AC6: Paginate with 24 offers per page.

**US-V03 — Offer details (P0)**
As a visitor, I want to see complete offer details before contacting the merchant.
- AC1: Image gallery of up to 10 images with zoom and touch swipe.
- AC2: Show title, full description, both prices and discount, start/end dates, and countdown.
- AC3: Shop card shows name, logo, wilaya, address, shop link, and "Verified merchant" badge.
- AC4: Every offer has a shareable direct link with Open Graph image and title.
- AC5: If expired or hidden, show "This offer is no longer available" with similar offers (HTTP 410 for expired, 404 for missing).

**US-V04 — WhatsApp contact (P0)**
As a visitor, I want one button to message the merchant.
- AC1: A fixed mobile button opens `https://wa.me/213XXXXXXXXX?text=...`.
- AC2: The prefilled message includes the offer title and link.
- AC3: Record one click per offer per visitor per hour for merchant statistics without storing personal data.
- AC4: Provide a secondary `tel:` call button.

**US-V05 — Shop page (P1)**
As a visitor, I want to see all current offers from one shop.
- AC1: `/shop/{slug}` shows public shop information and only its active offers.

**US-V06 — Report an offer (P0)**
As a visitor, I want to report a misleading or expired offer.
- AC1: Form with a reason list and optional message; no account required.
- AC2: Maximum five reports per IP per hour and one report per offer per IP.
- AC3: Three reports from different addresses mark the offer "needs review" in the admin panel; do not hide it automatically.

**US-V07 — Sharing (P1)**
- AC1: Share button uses Web Share API on mobile and copies the link on desktop.

### Epic 2 — Merchant registration and verification

**US-M01 — Google sign-in (P0)**
As a merchant, I want one-click Google sign-in.
- AC1: "Continue with Google" starts OAuth2 (Authorization Code + PKCE + state).
- AC2: First login creates a `merchant` user and opens onboarding.
- AC3: Later login opens the last incomplete step or dashboard.
- AC4: Use a 30-day renewable refresh-token family; logout revokes it in the database.
- AC5: Reject an unverified Google email.

**US-M02 — Personal information (P0)**
- AC1: Fields: full name, valid Algerian phone stored as E.164 (`+213...`), wilaya, and address.
- AC2: Autosave as a draft and allow the merchant to return later.

**US-M03 — Shop information (P0)**
- AC1: Fields: shop name, main category, description, optional logo, wilaya, municipality, address, required WhatsApp, optional phone/social/site, commercial-register number and expiry.
- AC2: Generate a unique `slug` from the shop name; allow one edit before verification.
- AC3: Validate WhatsApp format only; no verification code in V1.

**US-M04 — Activation documents (P0)**
As a merchant, I want to upload my identity and commercial-register documents to activate my account.
- AC1: Three required documents: identity card front, identity card back, and a valid commercial register.
- AC2: JPG/PNG/WebP/PDF up to 8 MB; detect content with magic bytes, not the extension.
- AC3: Store documents in a private bucket with no public URL.
- AC4: After all three are submitted, set status to `pending` and show "Your application is under review".
- AC5: Clearly explain data protection, collection purpose, retention period, and record explicit timestamped checkbox consent.

**US-M05 — Application status (P0)**
- AC1: Dashboard shows incomplete / pending / approved / rejected / suspended.
- AC2: Rejection shows the reason and any rejected document; the merchant can re-upload and resubmit.
- AC3: Send email on every status change.
- AC4: Before approval, the merchant may create drafts but cannot submit them for review.

### Epic 3 — Offer management (merchant)

**US-O01 — Create offer (P0)**
As an approved merchant, I want to create an offer quickly from my phone.
- AC1: Fields:
  - title (5–120 characters)
  - description (up to 2,000 characters)
  - required category
  - optional before/after prices; if present, after < before and discount is calculated automatically
  - required discount from 1–99 when prices are not supplied
  - start/end date and time; end after start and duration ≤ 180 days
  - wilaya and address, prefilled from the shop but editable
  - WhatsApp number, defaulting to the shop number
- AC2: Autosave as a draft every 10 seconds while editing.
- AC3: No total-offer cap in V1, but maximum 50 active offers, configurable.

**US-O02 — Offer images (P0)**
- AC1: One to **10** images; the first is the cover.
- AC2: JPG/PNG/WebP up to 8 MB; the frontend compresses before upload with a 2,000px longest edge.
- AC3: The server converts to WebP at 400 / 800 / 1600 pixels and removes EXIF.
- AC4: Drag reorder, delete, and change cover.
- AC5: Reject images smaller than 400×400 pixels.

**US-O03 — Submit for review (P0)**
- AC1: "Submit for review" validates required fields and at least one image.
- AC2: Set status to `pending` and lock edits to core fields.
- AC3: Admin approval decides publication and notifies the merchant.

**US-O04 — Edit published offer (P0)**
- AC1: **Any edit** to an approved offer (description, images, prices, dates, WhatsApp, address) returns it to `pending` and hides it until admin approval.
- AC2: Warn before save: "The offer will be hidden until it is reviewed."
- AC3: Showing the old version during review is not supported in V1.

**US-O05 — Pause/archive/delete (P0)**
- AC1: Pause (hide), archive (hide but retain history), and permanently delete drafts only.
- AC2: Approved offers cannot be permanently deleted; archive them to preserve statistics and history.

**US-O06 — Duplicate offer (P1)**
- AC1: "Duplicate offer" creates a draft with the same data and images for new dates.

**US-O07 — Statistics (P1)**
- AC1: Per offer, show views and WhatsApp clicks for the last 30 days in a simple chart.
- AC2: Dashboard summary shows active offers, total views, and total clicks.

### Epic 4 — Trial and subscription

**US-B01 — Start trial automatically (P0)**
- AC1: The 60-day trial starts **when the merchant's first offer is approved**; record `trial_started_at` and `trial_ends_at`.
- AC2: Start it in the same approval transaction to avoid inconsistent states.
- AC3: One trial per merchant; deleting offers does not reset it.
- AC4: Tie the trial to phone number and commercial register so the same register cannot receive another trial.

**US-B02 — Trial and subscription counter (P0)**
- AC1: Dashboard shows "N trial days left" or "Your subscription is active until {date}".
- AC2: Show a warning banner at 14 days or less.

**US-B03 — Reminders (P0)**
- AC1: Email 14 days, 7 days, 1 day, and the day of access expiry.
- AC2: Never send the same reminder twice (`notifications` with a unique key).

**US-B04 — Chargily electronic payment (P0)**
- AC1: "Subscribe — 4,000 DZD / year" creates checkout and redirects to Chargily.
- AC2: Success is accepted **only** after a valid signed webhook, not after the browser returns to the success page.
- AC3: The webhook is idempotent; duplicates never extend twice.
- AC4: The success page queries payment status and shows "Confirmation pending" until the webhook arrives.

**US-B05 — Manual payment (P0)**
- AC1: Show manual account details and a unique payment reference.
- AC2: Merchant uploads transfer proof (image/PDF up to 8 MB) and transaction number.
- AC3: Status is `under_review` until admin approval or rejection with a reason.
- AC4: Approval activates the subscription and notifies the merchant.

**US-B06 — Payment history (P1)**
- AC1: List payments and statuses with a simple PDF receipt (number, date, amount, method).

**US-B07 — Expiry behavior (P0)**
- AC1: When access expires, hide all merchant offers from visitors immediately; do not delete them.
- AC2: The merchant can enter the dashboard and edit, but nothing is published.
- AC3: Active offers reappear immediately when subscription is activated.

**US-B08 — Renewal (P0)**
- AC1: The new subscription starts at `max(now, current access end)` so the merchant loses no paid or trial days.

### Epic 5 — Admin dashboard

**US-A01 — Admin sign-in (P0)**
- AC1: Admins also use Google sign-in; `admin` is granted only to configured bootstrap emails.
- AC2: `/admin` is protected in both frontend and API.

**US-A02 — Merchant application queue (P0)**
- AC1: List `pending` applications oldest first with search and status filter.
- AC2: Application page shows all merchant/shop data and three documents.
- AC3: Serve documents through five-minute signed URLs; log every opening in `audit_logs`.
- AC4: Actions: approve, reject (reason required and document-specific), suspend, and reactivate.

**US-A03 — Offer review (P0)**
- AC1: `pending` offer queue with full visitor preview.
- AC2: Approve, reject with reason, or remove. Approving the first offer starts the trial (US-B01).
- AC3: Bulk approve/reject several offers.
- AC4: Review oldest first with a stated 24-hour review target.

**US-A04 — Manual payment review (P0)** — See US-B05, with proof, merchant name, and amount.

**US-A05 — Category management (P0)**
- AC1: Add, edit, order, and hide categories (main and subcategory levels) with icon and slug.
- AC2: A category with offers cannot be deleted; hide it instead.

**US-A06 — Reports (P1)**
- AC1: Reports are grouped by offer with ignore, hide offer, and suspend merchant actions.

**US-A07 — Advertising (P1)**
- AC1: Create ads with advertiser, image, link, slot, dates, priority, and optional wilaya/category filters.
- AC2: Report impressions and clicks per ad.

**US-A08 — Metrics dashboard (P1)**
- AC1: Show new merchants, pending requests, active offers, trial/subscribed/expired merchants, monthly revenue, and top categories.

**US-A09 — Settings (P1)**
- AC1: Configure trial days (60), subscription price (4,000), subscription days (365), maximum images (10), maximum active offers (50), and manual payment details.

**US-A10 — Audit log (P1)**
- AC1: Read-only list of sensitive actions: actor, action, time, and IP.

### Epic 6 — Advertising

**US-D01 — Display ads (P1)**
- AC1: Fixed slots: `home_top`, `list_inline` (after every 8 offers), `detail_sidebar`, and `footer_banner`.
- AC2: Each slot has defined sizes (see `ui.md`) and reserves space to avoid CLS.
- AC3: Every ad has a clear "Advertisement" label.
- AC4: Impressions and clicks are aggregated daily.
- AC5: V1 supports direct-sold ads only; no AdSense.

### Epic 7 — Platform and background system

**US-S01 — Offer expiry (P0)**
- AC1: An offer disappears at `ends_at` without intervention; visibility queries use time, not a scheduled task.
- AC2: A nightly task archives offers expired for more than 30 days and deletes large images after 90 days.

**US-S02 — Notifications (P0)**
- Email the merchant on account approval/rejection, offer approval/rejection, expiry reminders, and payment confirmation.
- Email admins on a new merchant application and manual payment; use a daily summary when volume is high.

**US-S03 — SEO (P0)**
- AC1: SSR for all public pages, dynamic `sitemap.xml`, `robots.txt`, `Offer` JSON-LD, and canonical links.

**US-S04 — Backups (P0)**
- AC1: Daily encrypted `pg_dump` stored off-server for at least 14 days, with a monthly restore test.

---

## 6. Business rules

### 6.1 Public offer visibility (single reference)

An offer appears to visitors **if and only if** all conditions hold:

```
offer.status            = 'approved'
offer.starts_at        <= now()
offer.ends_at           > now()
merchant.verification   = 'approved'
merchant.access_until   > now()
```

Implement this as one database **VIEW** (`public_offers`) used by every public query. Do not duplicate the rule elsewhere.

### 6.2 Merchant status (`verification_status`)

```
incomplete ─▶ pending ─▶ approved ◀──▶ suspended
                 │
                 └──▶ rejected ─▶ (re-upload) ─▶ pending
```

### 6.3 Offer status (`status`)

```
draft ─▶ pending ─▶ approved ─▶ archived
            │          │
            ▼          └─▶ pending   (after edit)
         rejected ─▶ draft           (for correction)
approved/pending ─▶ removed          (admin removal)
```
"Expired" is not a stored status; derive it from `ends_at < now()`.

### 6.4 Derived access status

| Status | Condition |
|---|---|
| `no_trial` | Verified and no offer has been approved |
| `trialing` | `trial_ends_at > now()` and no subscription |
| `subscribed` | An active subscription ends in the future |
| `expired` | `access_until <= now()` |

`merchants.access_until = max(trial_ends_at, latest subscription.ends_at)` and update it in the same transaction that changes either value.

### 6.5 Trial and subscription
- Trial: 60 days, started only when the **first offer** is approved.
- Subscription: 4,000 DZD for 365 days, starting at `max(now, access_until)`.
- Refunds: no automatic refunds; terms of use define the policy. Admins may cancel a payment and manually adjust access with an audit record.

### 6.6 Prices and discount
- If both prices exist: `discount_percent = round((before - after) * 100 / before)`.
- If only a percentage exists, display it without prices.
- Prevent abuse: after price < before price and percentage from 1 to 99.

### 6.7 WhatsApp number
- Store as E.164, for example `+213550123456`.
- Accept `0550123456`, `550123456`, or `+213550123456`, then normalize on the server.
- V1 supports Algerian mobile numbers only (starting with 5/6/7).

### 6.8 Documents
- Only the owner and admins can view them.
- Delete 12 months after account closure or on merchant request unless law requires longer retention.
- Never display them publicly or send them by email.

---

## 7. Non-functional requirements

| Area | Target |
|---|---|
| Performance | LCP < 2.5s, INP < 200ms, first mobile page < 1 MB, API p95 < 300ms for lists |
| Availability | 99.5% monthly in V1 |
| Capacity (V1) | 500 merchants, 5,000 active offers, 50,000 monthly visits |
| Security | OWASP Top 10, HTTPS only, HSTS, CSP, rate limits, disk encryption, environment secrets |
| Privacy | Data minimization, explicit consent, privacy policy, document-access audit |
| Accessibility | WCAG 2.1 AA: contrast, keyboard navigation, field labels, image alt text |
| Language | Arabic RTL by default, French in V2; architecture ready for i18n |
| Browsers | Latest two Chrome/Safari/Firefox/Samsung Internet versions; Android 8+ |
| Monitoring | JSON logs, `/healthz`, `/readyz`, alerts for 5xx and webhook failures |
| Testing | Business-rule unit tests, API integration tests with a real database, and E2E visitor/merchant journeys |

---

## 8. Security and privacy (summary)

1. **OAuth:** `state` + PKCE, `email_verified` validation, and session-identifier renewal at login.
2. **Tokens:** HttpOnly, Secure, SameSite cookies for rotated refresh tokens; store only hashes server-side.
3. **CSRF:** Reject unsafe methods (POST/PUT/PATCH/DELETE) when `Origin` does not match the allowlist.
4. **Uploads:** magic-byte checks, size limits, image re-encoding that removes hidden payloads, and random filenames.
5. **Documents:** private bucket, five-minute signed URL, and an audit record for every access.
6. **Webhook:** HMAC verification over the raw body, then idempotency by event ID.
7. **Rate limiting:** 120 public requests/IP/minute, with stricter reports, WhatsApp, authentication, uploads, and payments.
8. **XSS:** no HTML in descriptions; escape plain text on output and preserve only line breaks.
9. **Audit:** record every approval, rejection, suspension, document view, and settings change.
10. **Compliance:** Algerian personal-data law may require declaration or authorization when processing identity documents. **Verify current procedures before launch** and consult legal counsel.

---

## 9. Roadmap

| Phase | Approximate duration | Content |
|---|---|---|
| **0 — Foundation** | 1 week | Repository, Docker Compose, CI, first migrations, DaisyUI theme, logo |
| **1 — Accounts** | 2 weeks | Google OAuth, merchant wizard, document uploads, admin approval panel |
| **2 — Offers** | 3 weeks | Offer form, image upload, review, home/search/details/WhatsApp |
| **3 — Subscription** | 2 weeks | Trial rules, Chargily, manual payment, expiry hiding, reminders |
| **4 — Ads and optimization** | 2 weeks | Advertising, statistics, reports, SEO, performance |
| **Pilot launch** | — | One city and 20–30 manually onboarded merchants |
| **V2** | Later | Video, customer accounts/favorites, French, WhatsApp notifications |

### Definition of Done for every story
1. All acceptance criteria are met.
2. Unit/integration tests and critical-path E2E tests pass.
3. It works at 360px without RTL layout errors.
4. No console errors or critical accessibility warnings.
5. Upload and authorization paths receive a security review.
6. The migration works on both an empty and populated database.

---

## 10. Assumptions and open questions

1. **Administrative divisions:** confirm the official wilaya count before creating `wilayas` data.
2. **Name and domain:** verify `dinar.dz` availability and trademark safety before investing in identity.
3. **Pre-review:** mandatory for every V1 offer. Later, a "trusted merchant" may publish after 20 accepted offers without rejection.
4. **Chargily data:** verify webhook event names and response fields from official documentation at implementation time.
5. **Invoicing:** determine whether merchants need official invoices based on platform legal status; deferred.
6. **First notification language:** Arabic, with French later.

---

## 11. Executable system catalogue: pages, features, and API endpoints

This section turns the user stories into a clear implementation contract between frontend and API. Every feature must add or update four items: page, UI state, endpoint, and acceptance test.

### 11.1 API contract rules

- Base is `/api/v1`; all fields use `snake_case`.
- Dates are ISO 8601 UTC and amounts are integer Algerian dinars.
- A single success is `{ "data": {...} }`; a list is `{ "data": [...], "meta": {"page": 1, "per_page": 24, "total": 0} }`.
- Errors use:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Check the fields",
    "fields": {"title": "Minimum 5 characters"},
    "request_id": "req_01..."
  }
}
```

- Standard `code` values are `unauthorized`, `forbidden`, `not_found`, `gone`, `conflict`, `validation_failed`, `rate_limited`, and `internal`.
- `401` means a missing or expired session, `403` means the role lacks permission, and `422` means invalid input.
- Protected endpoints use the rotated HttpOnly refresh session; the frontend never sends JWTs or secrets in `localStorage`.
- Every endpoint returns `X-Request-ID`; the same `request_id` appears in logs and error responses.

### 11.2 Shared data models

#### Offer `OfferSummary`

```json
{
  "id": "uuid",
  "slug": "string",
  "title": "string",
  "cover_url": "https://media.dinar.dz/.._800.webp",
  "price_before_dzd": 12500,
  "price_after_dzd": 9900,
  "discount_percent": 21,
  "starts_at": "2026-10-09T08:00:00Z",
  "ends_at": "2026-10-20T22:00:00Z",
  "remaining_seconds": 950400,
  "wilaya": {"id": 16, "name": "Algiers"},
  "category": {"id": "uuid", "name": "Clothing", "slug": "clothing"},
  "shop": {"id": "uuid", "name": "Shop name", "slug": "shop-name", "verified": true}
}
```

#### Offer `OfferDetails`

Contains every `OfferSummary` field plus `description`, `address`, `whatsapp_url`, `phone`, `images[]`, and `similar_offers[]`. Each `images[]` item is `{id, url_400, url_800, url_1600, alt, position}`. Private merchant fields and statistics are not returned on public pages.

#### Merchant status `MerchantStatus`

```json
{
  "verification_status": "incomplete|pending|approved|rejected|suspended",
  "access_status": "no_trial|trialing|subscribed|expired",
  "trial_started_at": "2026-10-09T08:00:00Z",
  "trial_ends_at": "2026-12-08T08:00:00Z",
  "access_until": "2026-12-08T08:00:00Z",
  "rejection_reason": null,
  "next_step": "personal|shop|documents|submit|dashboard"
}
```

### 11.3 Public page map

| Page | Path | Required data | Operations and states |
|---|---|---|---|
| Home | `/` | `latest`, `ending_soon`, `top_discount`, categories, ad | SSR; skeleton; empty; API error; card view after two seconds |
| Offer list | `/offers` | Offers, `meta`, wilaya/category lists | Query-string filters, sorting, pagination, no-results state, shareable URL |
| Offer details | `/offers/[id]` | `OfferDetails`, similar offers, ad | Gallery, countdown, share, WhatsApp, call, expired `410`, missing `404` |
| Shop page | `/shop/[slug]` | Public shop data and active offers | Pagination; missing-shop state; never show expired or hidden offers |
| Login | `/login` | Nothing before login | Google OAuth redirect; unverified-email error; safe internal `next` only |
| Privacy | `/privacy` | Static content | Explain document purpose, retention, and merchant rights |
| Terms | `/terms` | Static content | Explain that the platform displays offers, does not sell products, and does not collect product payments |

#### `/offers` search parameters

| Parameter | Type | Required | Values/constraints |
|---|---|---:|---|
| `q` | string | No | Up to 120 characters, Arabic normalization, excluded from SEO indexing |
| `wilaya` | integer | No | Valid wilaya ID |
| `category` | UUID or slug | No | Main or child category |
| `min_discount` | integer | No | 0–99 |
| `ending_within` | enum | No | `24h` or `7d` |
| `sort` | enum | No | `newest`, `ending_soon`, `discount`, `price_asc` |
| `page` | integer | No | Starts at 1 |
| `per_page` | integer | No | Default 24, maximum 48 |

### 11.4 Merchant page map

Every merchant page requires an authenticated `merchant` role. If the merchant is not verified, the guard displays `next_step` and blocks offer submission.

| Page | Path | Content | Operations |
|---|---|---|---|
| Merchant dashboard | `/dashboard` | `MerchantStatus`, offer count, views, clicks, latest review | Open offer, continue verification, subscribe, warn at 14 days |
| Personal information | `/onboarding/personal` | Name, phone, wilaya, address | Autosave draft, field errors, next to shop information |
| Shop information | `/onboarding/shop` | Name, category, description, logo, location, WhatsApp, register | Create/edit slug once, save, validate phone/register date |
| Documents | `/onboarding/documents` | Front/back identity and register status | Upload, replace rejected document, delete draft, show consent/privacy |
| Submit verification | `/onboarding/review` | Data/document summary and consent status | `submit`; block missing documents or expired register |
| Merchant offers | `/dashboard/offers` | Offers with status, dates, and filters | Create, edit, pause, archive, duplicate, submit for review |
| Create offer | `/dashboard/offers/new` | Offer form and image uploader | Save draft every 10 seconds; validate prices/dates/images; upload/order |
| Edit offer | `/dashboard/offers/[id]/edit` | Same as create form | Owner only; editing approved returns it to `pending` and hides it |
| Statistics | `/dashboard/stats` | Totals and 30-day chart | Switch offer/period; never expose IP or personal data |
| Billing | `/dashboard/billing` | Access status, price, manual account, payment history | Chargily checkout, manual proof upload, payment polling |

#### Create/edit offer model

```json
{
  "title": "string, 5..120",
  "description": "string, max 2000",
  "category_id": "uuid",
  "price_before_dzd": 12500,
  "price_after_dzd": 9900,
  "discount_percent": null,
  "starts_at": "2026-10-09T08:00:00Z",
  "ends_at": "2026-10-20T22:00:00Z",
  "wilaya_id": 16,
  "address": "string",
  "whatsapp": "+213550123456"
}
```

Send `price_before_dzd` and `price_after_dzd` together, or send only `discount_percent`. When both prices exist, the server calculates the percentage and ignores a client-calculated value. `ends_at` must be after `starts_at` and the duration cannot exceed 180 days.

### 11.5 Admin page map

Every admin page requires the `admin` role and returns `403` even when the client hides the button.

| Page | Path | Data and features |
|---|---|---|
| Metrics | `/admin` | New merchants, pending requests, active offers, access states, revenue, top categories |
| Merchants | `/admin/merchants` | `status`/`q`/pagination; approve/reject/suspend/reactivate |
| Merchant details | `/admin/merchants/[id]` | Account, shop, documents, offers, payments, decisions |
| Offer review | `/admin/offers` | Status filter, preview, approve/reject/remove, bulk actions |
| Payments | `/admin/payments` | Chargily/manual payments, proof, approve/reject with required reason |
| Categories | `/admin/categories` | Create, edit, order, hide; cannot delete linked category |
| Ads | `/admin/ads` | Advertiser, image, link, slot, period, priority, wilaya/category, statistics |
| Reports | `/admin/reports` | Group by offer, reason, message, unique IP count, ignore/hide/suspend |
| Settings | `/admin/settings` | Trial, price, subscription duration, limits, manual payment details |
| Audit | `/admin/audit-logs` | Actor, action, time, IP, resource, request ID; read-only |

### 11.6 Public endpoint contract

| Method and path | Inputs | Successful output |
|---|---|---|
| `GET /offers` | query: `q,wilaya,category,min_discount,ending_within,sort,page,per_page` | `data: OfferSummary[]` + `meta` |
| `GET /offers/:id` | path `id` or slug | `data: OfferDetails`; `410` for expired and `404` for missing |
| `GET /offers/:id/similar` | path `id`, optional `limit` up to 12 | `data: OfferSummary[]` |
| `POST /offers/:id/view` | No body; IP/offer rate limit | `{data:{recorded:true}}`; duplicates in the window do not increment twice |
| `POST /offers/:id/whatsapp` | No body | `{data:{whatsapp_url, recorded}}`; server builds URL from normalized number |
| `POST /offers/:id/report` | `{reason, message?}`; reason from allowlist, message up to 1,000 | `{data:{report_id, status:"received"}}` |
| `GET /shops/:slug` | query `page,per_page` | `{data:{shop,offers},meta}` |
| `GET /categories` | Optional query; public callers cannot enable hidden categories | `{data:[{id,name,slug,parent_id,icon_url}]}` |
| `GET /wilayas` | No body | `{data:[{id,name,slug}]}` |
| `GET /ads` | query `slot,wilaya,category` | `{data:{id,image_url,target_url,alt,slot}}` or `data:null` |
| `POST /ads/:id/click` | No body | `{data:{redirect_url}}` with click recorded |
| `GET /sitemap/offers` | No body | XML with active offer links; internal sitemap endpoint |

### 11.7 Authentication contract

| Method and path | Inputs | Output |
|---|---|---|
| `GET /auth/google/login` | Optional `next`; internal path only | `302` to Google with `state` and PKCE in a temporary cookie |
| `GET /auth/google/callback` | `code,state` from Google | `302` to next step and create token cookies; OAuth errors create no session |
| `POST /auth/logout` | No body | `{data:{logged_out:true}}` and revoke token family |
| `GET /me` | Access JWT | `{data:{user:{id,email,name,role},merchant:MerchantStatus|null}}` |

### 11.8 Merchant and owned-offer contract

| Method and path | Inputs | Output |
|---|---|---|
| `PUT /merchant/profile` | `{full_name,phone,wilaya_id,address}` | `{data:{profile,completed:true}}` |
| `PUT /merchant/shop` | Shop profile fields and register data | `{data:{shop,slug,completed:true}}` |
| `POST /merchant/documents` | multipart: `type` = `id_front\|id_back\|commercial_register` and `file` up to 8MB | `{data:{id,type,status:"uploaded",uploaded_at}}`; never return storage key |
| `POST /merchant/submit` | `{consent_accepted:true}`; all data and documents complete | `{data:{verification_status:"pending",submitted_at}}` |
| `GET /merchant/offers` | query `status,page,per_page` | Owned `OfferSummary[]` with `status` and `rejection_reason` |
| `POST /merchant/offers` | Offer model from section 11.4 | `{data:{offer,status:"draft"}}` |
| `GET /merchant/offers/:id` | Merchant-owned path `id` | `{data:{offer,images,validation}}` |
| `PUT /merchant/offers/:id` | Changed offer fields | `{data:{offer,status}}`; approved becomes `pending` |
| `DELETE /merchant/offers/:id` | path `id` | `{data:{deleted:true}}` for drafts; published offers are archived |
| `POST /merchant/offers/:id/media` | multipart `file`, JPG/PNG/WebP, 8MB, at least 400x400 | `{data:{media_id,url_400,url_800,url_1600,position}}` |
| `PATCH /merchant/offers/:id/media/order` | `{ids:["uuid",...]}`; every current image once | `{data:{images:[...]}}` |
| `DELETE /merchant/offers/:id/media/:mid` | path `id,mid` | `{data:{deleted:true}}`; cannot delete the only image on submission |
| `POST /merchant/offers/:id/submit` | No body | `{data:{status:"pending",submitted_at}}` after validation |
| `POST /merchant/offers/:id/archive` | No body | `{data:{status:"archived"}}` |
| `POST /merchant/offers/:id/duplicate` | No body | `{data:{offer,status:"draft"}}` with logical image copies |
| `GET /merchant/stats` | Optional `offer_id,from,to`; default 30 days | `{data:{totals:{views,whatsapp_clicks},series:[{date,views,whatsapp_clicks}]}}` |

### 11.9 Billing and trial contract

Chargily integration, card-data prohibition, and webhook verification are documented in [`payment.md`](payment.md).

| Method and path | Inputs | Output |
|---|---|---|
| `POST /merchant/billing/checkout` | `{plan:"annual"}` | `{data:{payment_id,checkout_url,amount_dzd:4000,expires_at}}` |
| `POST /merchant/billing/manual` | multipart: `proof_file` up to 8MB and `proof_reference` | `{data:{payment_id,status:"under_review",amount_dzd}}` |
| `GET /merchant/billing/payments` | query `page,per_page,status` | `data:[{id,method,amount_dzd,status,reference,created_at,paid_at,receipt_url?}]` + `meta` |
| `GET /merchant/billing/payments/:id` | Merchant-owned path `id` | `{data:{payment,status,subscription_ends_at?}}` |
| `POST /webhooks/chargily` | Raw body + Chargily signature | `{data:{received:true}}`; verify signature and idempotency before activation |

Payment `status` is `pending`, `under_review`, `paid`, `rejected`, or `failed`. The frontend must not treat a Chargily return as success; true success comes from a valid webhook or admin approval of manual payment.

### 11.10 Admin contract

| Method and path | Inputs | Output |
|---|---|---|
| `GET /admin/dashboard` | Optional `from,to` | `{data:{merchants,offers,access,monthly_revenue_dzd,top_categories}}` |
| `GET /admin/merchants` | query `status,q,page,per_page` | `data:[merchant summaries]` + `meta` |
| `GET /admin/merchants/:id` | path `id` | `{data:{merchant,shop,documents_metadata,offers,payments,audit_summary}}` |
| `GET /admin/documents/:id/url` | path `id` | `{data:{signed_url,expires_at}}`; record `document.view` in audit |
| `POST /admin/merchants/:id/approve` | No body or `{note?}` | `{data:{verification_status:"approved"}}` |
| `POST /admin/merchants/:id/reject` | `{reason,document_type?}`; reason required | `{data:{verification_status:"rejected",rejection_reason}}` |
| `POST /admin/merchants/:id/suspend` | `{reason}` | `{data:{verification_status:"suspended"}}` |
| `POST /admin/merchants/:id/reactivate` | No body | `{data:{verification_status:"approved"}}` |
| `GET /admin/offers` | query `status,q,page,per_page` | `data:[offer review rows]` + `meta` |
| `POST /admin/offers/:id/approve` | No body or `{note?}` | `{data:{status:"approved",trial_started_at?,trial_ends_at?}}`; first approval starts trial in one transaction |
| `POST /admin/offers/:id/reject` | `{reason}`; reason required | `{data:{status:"rejected",rejection_reason}}` |
| `POST /admin/offers/:id/remove` | `{reason}` | `{data:{status:"removed"}}` |
| `POST /admin/offers/bulk` | `{action:"approve|reject|remove",ids:[uuid],reason?}` | `{data:{processed,failed:[{id,code}]}}` |
| `GET /admin/payments` | query `status,method,page,per_page` | `data:[payment review rows]` + `meta` |
| `POST /admin/payments/:id/approve` | No body or `{note?}` | `{data:{status:"paid",subscription_ends_at}}` |
| `POST /admin/payments/:id/reject` | `{reason}` | `{data:{status:"rejected",rejection_reason}}` |
| `GET/POST/PUT/DELETE /admin/categories` | category: `{name,slug,parent_id?,icon_url?,sort_order,is_visible}` | Saved resource; cannot delete a category linked to offers |
| `GET/POST/PUT/DELETE /admin/ads` | `{advertiser,image_key,target_url,slot,starts_at,ends_at,priority,wilaya_id?,category_id?}` | Ad with `impressions` and `clicks` on read |
| `GET /admin/reports` | query `status,page,per_page` | `data:[{id,offer_id,reason,message,unique_ip_count,status,created_at}]` + `meta` |
| `POST /admin/reports/:id/resolve` | `{action:"ignore|hide_offer|suspend_merchant",note?}` | `{data:{status:"resolved",action}}` |
| `GET /admin/settings` | No body | `{data:{trial_days,annual_price_dzd,subscription_days,max_images,max_active_offers,manual_payment_details}}` |
| `PUT /admin/settings` | Same settings object with bounds validation | `{data:{settings,updated_at}}` |
| `GET /admin/audit-logs` | query `actor_id,action,resource_type,from,to,page,per_page` | `data:[{id,actor_id,action,resource_type,resource_id,ip,request_id,created_at}]` + `meta` |

### 11.11 Operations and health endpoints

| Path | Purpose | Response |
|---|---|---|
| `GET /healthz` | Is the API process alive? | `200 {"status":"ok"}` without DB check |
| `GET /readyz` | Is the API ready? | `200 {"status":"ready","database":"ok","storage":"ok"}` or `503` |
| `GET /robots.txt` | Search-engine rules | Text disallowing `/dashboard`, `/admin`, `/onboarding`, and `/api` |
| `GET /sitemap.xml` | SEO | Static/dynamic sitemap with public pages only |

### 11.12 Permission matrix and acceptance tests

| Area | Visitor | Merchant | Admin |
|---|:---:|:---:|:---:|
| Read public offers | Yes | Yes | Yes |
| Report/WhatsApp/view | Yes with rate limit | Yes | Yes |
| Edit own profile/offers | No | Yes, own only | Yes |
| View sensitive documents | No | Own only | Yes, temporary URL with audit |
| Approve or reject | No | No | Yes |
| Change settings | No | No | Yes |

Before marking a feature complete, test valid inputs, every validation error, `401/403`, cross-merchant data isolation, API outage, and offer/subscription expiry. Critical E2E paths are: `visitor ← search ← details ← WhatsApp`, `merchant ← verification ← create offer ← admin review ← public visibility`, and `merchant ← payment ← webhook/manual approval ← visibility restored`.
