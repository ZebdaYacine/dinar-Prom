# Dinar DZ — PRD Subject Index

> Source of truth: `userstory.md` plus the specifications in `skills/`. This master PRD file is stored in `prd/`.
> Status is `🟡 Planned` because the current workspace contains the product specification, not the application implementation.
> All API paths below are relative to `/api/v1` unless stated otherwise.

## Portfolio

| PRD | Epic | Stories | Priority | Status | Primary surfaces |
|---|---|---|---|---|---|
| PRD-01 | Guest Experience | US-V01 … US-V07 (7) | High | ✅ Shipped | Public Next.js pages |
| PRD-02 | Merchant Identity & Verification | US-M01 … US-M05 (5) | Critical | ✅ Shipped | Auth and onboarding |
| PRD-03 | Merchant Offers | US-O01 … US-O07 (7) | Critical | 🟡 Planned | Merchant dashboard |
| PRD-04 | Trial, Billing & Payments | US-B01 … US-B08 (8) | Critical | 🟡 Planned | Billing and Chargily |
| PRD-05 | Admin Review & Audit | US-A01 … US-A04, US-A10 (5) | Critical | 🟡 Planned | Admin review queues |
| PRD-06 | Admin Catalog, Reports & Ads | US-A05 … US-A09, US-D01 (6) | Medium | 🟡 Planned | Admin operations and ad slots |
| PRD-07 | System Lifecycle | US-S01 … US-S04 (4) | Critical | 🟡 Planned | Worker, SEO, backups, notifications |
| PRD-08 | Data & API Foundation | Enabler for all stories | Critical | 🟡 Planned | PostgreSQL and Go API |
| PRD-09 | Web Shell, Brand & Splash | Cross-cutting | High | 🟡 Planned | Next.js shell and design system |
| PRD-10 | Delivery, Security & Quality | Enabler for all stories | Critical | 🟡 Planned | Docker, Caddy, CI, observability |

## Individual PRD files

- [Program Overview — PRD-00](PRD-00-Program-Overview.md)
- [Guest Experience — PRD-01](PRD-01-Guest-Experience.md)
- [Merchant Identity and Verification — PRD-02](PRD-02-Merchant-Identity-and-Verification.md)
- [Merchant Offers — PRD-03](PRD-03-Merchant-Offers.md)
- [Trial, Billing and Payments — PRD-04](PRD-04-Trial-Billing-and-Payments.md)
- [Admin Review and Audit — PRD-05](PRD-05-Admin-Review-and-Audit.md)
- [Admin Catalog, Reports and Ads — PRD-06](PRD-06-Admin-Catalog-Reports-and-Ads.md)
- [System Lifecycle — PRD-07](PRD-07-System-Lifecycle.md)
- [Data and API Foundation — PRD-08](PRD-08-Data-and-API-Foundation.md)
- [Web Shell, Brand and Splash — PRD-09](PRD-09-Web-Shell-Brand-and-Splash.md)
- [Delivery, Security and Quality — PRD-10](PRD-10-Delivery-Security-and-Quality.md)

## Delivery order

```text
PRD-08 Data/API
       |
       +--> PRD-02 Merchant identity ----+
       |                                 |
       +--> PRD-01 Guest experience      +--> PRD-03 Offers
       |                                 |
       +---------------------------------+--> PRD-04 Billing
                                             |
PRD-09 Web shell ----------------------------+
                                             v
                              PRD-05 Admin --> PRD-06 Commercial ops
                                             |
                              PRD-07 System + PRD-10 Delivery
```

---

# PRD-01 · Guest Experience

|              |                                                                 |
| ------------ | ----------------------------------------------------------------------- |
| **Epic**     | 1 — Guest Experience                                                  |
| **Stories**  | US-V01 … US-V07 (7)                                                    |
| **Priority** | High                                                                    |
| **Status**   | ✅ Shipped                                                              |
| **Surfaces** | `apps/web/app/(public)/page.tsx`, `offers`, `offers/[id]`, `shop/[slug]` |
| **API**      | `GET /offers`, `GET /offers/:id`, `GET /offers/:id/similar`, `POST /offers/:id/view`, `POST /offers/:id/whatsapp`, `POST /offers/:id/report`, `GET /shops/:slug`, `GET /categories`, `GET /wilayas`, `GET /ads` |

## 1. Summary

An unauthenticated visitor can discover current discounts, filter them by Arabic search, open offer details, view a shop, and contact the merchant through WhatsApp without creating an account.

## 2. Problem

The platform must demonstrate real local inventory before asking a visitor to register. Showing stale, hidden, or inaccessible offers would destroy trust and reduce merchant value.

## 3. Goals

- Render real, currently visible offers from the single `public_offers` rule.
- Reach offer details in one click from any offer card.
- Keep search and filters shareable through the URL.
- Make WhatsApp contact one tap on mobile.

### Non-goals

- Product checkout, cart, customer accounts, favorites, or native mobile apps in V1.
- Showing expired, unapproved, removed, or inaccessible merchant offers.

## 4. Users

**Guest visitor** — a person browsing from a phone, search engine, or shared offer link who wants a useful discount quickly.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Home shows search, primary categories, latest offers, ending-soon offers, and highest-discount offers. | US-V01 |
| R2 | Offer cards show cover, title, before/after prices, discount, shop, wilaya, and remaining time. | US-V01 |
| R3 | Search normalizes Arabic characters and supports wilaya, category, discount, expiry window, sort, and pagination. | US-V02 |
| R4 | Detail shows up to 10 images, description, price, dates, shop card, verified badge, similar offers, Open Graph metadata, and correct `404/410` behavior. | US-V03 |
| R5 | WhatsApp opens a prefilled message and records one deduplicated click per offer and visitor window. | US-V04 |
| R6 | Shop page shows public shop data and active offers only. | US-V05 |
| R7 | Report form accepts a reason and optional message with IP rate limits and duplicate protection. | US-V06 |
| R8 | Share uses Web Share API on mobile and clipboard on desktop. | US-V07 |

## 6. Flow

```text
Home (/) ──search/filter──▶ /offers ──click card──▶ /offers/:id
   │                              │                         │
   └──click shop──────────────────┴──▶ /shop/:slug           ├── WhatsApp
                                                             ├── Report
                                                             └── Share
```

## 7. Technical notes

- All public queries use `public_offers`; never duplicate the visibility predicate in the frontend.
- SSR the first public view; lazy-load secondary images, ads, and analytics.
- Use `navigator.sendBeacon` for delayed view recording and open WhatsApp only after the click event.
- Public endpoints are unauthenticated but rate-limited. Report is capped at 5 per hour per IP and one report per offer per IP.
- Arabic text normalization is server-side and indexed with PostgreSQL full-text search.

## 8. Subtasks

- [x] T1.1 Build public route group, RTL layout, SEO metadata, and loading/empty/error states.
- [x] T1.2 Implement public offer query, pagination, filters, sorting, and URL state.
- [x] T1.3 Build `OfferCard`, countdown, gallery, shop card, WhatsApp, share, and report components.
- [x] T1.4 Add view/click deduplication and rate-limit middleware.
- [x] T1.5 Add Playwright flow: home → search → detail → WhatsApp/report.

## 9. Success metrics

- LCP under 2.5s on a mid-range phone and 4G.
- Guest offer-detail click-through rate.
- WhatsApp click-through rate per visible offer.
- Report abuse rate and percentage of reports resolved within 24 hours.

## 10. Risks & open questions

- Search quality depends on Arabic normalization and seed quality; validate with real Algerian queries.
- A 410 page must still provide similar offers without accidentally exposing hidden content.
- Confirm the final wilaya seed list before production.

---

# PRD-02 · Merchant Identity & Verification

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 2 — Merchant Identity & Verification                            |
| **Stories**  | US-M01 … US-M05 (5)                                             |
| **Priority** | Critical                                                        |
| **Status**   | ✅ Shipped                                                      |
| **Surfaces** | `apps/web/app/(auth)/login`, `onboarding/personal`, `shop`, `documents`, `review` |
| **API**      | `GET /auth/google/login`, `GET /auth/google/callback`, `POST /auth/logout`, `GET /me`, `PUT /merchant/profile`, `PUT /merchant/shop`, `POST /merchant/documents`, `POST /merchant/submit` |

## 1. Summary

A merchant signs in with Google, completes personal and shop information, uploads three verification documents, accepts the privacy notice, and follows the review status from one guided onboarding flow.

## 2. Problem

The marketplace needs trustworthy merchant identity without forcing customers to create accounts. The onboarding must collect enough legal data while protecting sensitive identity documents.

## 3. Goals

- Complete onboarding on a phone with autosaved steps.
- Use Google OAuth with state, PKCE, verified email, and secure sessions.
- Store documents privately and expose them only to the merchant and admins.
- Make every verification status and rejection reason understandable.

### Non-goals

- Customer registration or password login in V1.
- Automatic document approval or identity verification by a third party.

## 4. Users

**Prospective merchant** — a shop owner who wants to publish offers and is willing to provide legal verification.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Google OAuth creates or resolves a merchant session and rejects unverified email. | US-M01 |
| R2 | Personal form validates Algerian phone format, wilaya, address, and autosaves as a draft. | US-M02 |
| R3 | Shop form captures name, category, description, logo, location, contact channels, commercial register number, and expiry date. | US-M03 |
| R4 | Slug is unique and editable once before verification. | US-M03 |
| R5 | ID front, ID back, and valid commercial register accept only allowed content types up to 8MB in a private bucket. | US-M04 |
| R6 | Explicit privacy consent is stored with timestamp before submission. | US-M04 |
| R7 | Status is `incomplete`, `pending`, `approved`, `rejected`, or `suspended`; rejection identifies reason and document where relevant. | US-M05 |

## 6. Flow

```text
/login ──Google OAuth──▶ /onboarding/personal
                              │
                              v
                         /onboarding/shop
                              │
                              v
                       /onboarding/documents
                              │
                              v
                         /onboarding/review
                    submit ──▶ pending ──▶ approved/rejected
```

## 7. Technical notes

- Session token is random in the cookie; only its SHA-256 hash is stored.
- The refresh cookie is `HttpOnly`, `Secure` in production, `SameSite=Strict`, and rotated within a 30-day lifetime; the short-lived OAuth state cookie uses `SameSite=Lax`.
- Document content type is detected from magic bytes, not filename extension.
- Private documents never return a storage key. Admin access uses a five-minute signed URL and audit log.
- CSRF double-submit validation protects refresh and logout cookie operations; merchant writes require the in-memory bearer access token, and the OAuth callback validates state and PKCE.

## 8. Subtasks

- [x] T2.1 Configure Google OAuth, callback, session table, cookie, logout, and role middleware.
- [x] T2.2 Build onboarding wizard with autosave, validation, resume step, and mobile UX.
- [x] T2.3 Implement private document upload, size/type checks, hashing, and storage adapter.
- [x] T2.4 Implement consent, merchant submission transaction, rejection/re-upload behavior, and email jobs.
- [x] T2.5 Add authorization and document privacy integration tests.

## 9. Success metrics

- Onboarding completion rate.
- Median time from OAuth login to submitted verification.
- Document rejection rate by reason.
- Zero unauthorized document reads in security tests.

## 10. Risks & open questions

- Data-protection obligations for Algerian identity documents require legal review before launch.
- Confirm accepted commercial-register formats and retention period with the legal owner.
- Confirm whether admin role is managed only through `ADMIN_EMAILS` or a future database workflow.

---

# PRD-03 · Merchant Offers

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 3 — Merchant Offers                                             |
| **Stories**  | US-O01 … US-O07 (7)                                             |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/offers`, `offers/new`, `offers/[id]/edit`, `stats` |
| **API**      | `GET/POST /merchant/offers`, `GET/PUT/DELETE /merchant/offers/:id`, `POST /merchant/offers/:id/media`, `PATCH /merchant/offers/:id/media/order`, `DELETE /merchant/offers/:id/media/:mid`, `POST /merchant/offers/:id/submit`, `POST /merchant/offers/:id/archive`, `POST /merchant/offers/:id/duplicate`, `GET /merchant/stats` |

## 1. Summary

An approved merchant can create a mobile-first offer draft, upload and order up to ten optimized images, submit it for review, manage its lifecycle, and inspect views and WhatsApp clicks.

## 2. Problem

Merchants need a fast publishing workflow, but every public offer must be complete, safe, time-bounded, image-optimized, and reviewed before visibility.

## 3. Goals

- Make a complete offer possible from a phone.
- Preserve drafts and prevent accidental data loss.
- Enforce price, date, image, ownership, and review rules server-side.
- Provide useful, privacy-safe performance statistics.

### Non-goals

- Video upload, customer checkout, or unlimited public publishing without moderation.
- Permanent deletion of an approved offer.

## 4. Users

**Approved merchant** — owns a shop, wants to publish and measure discounts.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Offer draft validates title, description, category, prices/discount, dates, wilaya, address, and WhatsApp. | US-O01 |
| R2 | Autosave runs every 10 seconds and supports a maximum of 50 active offers by setting. | US-O01 |
| R3 | One to ten JPG/PNG/WebP images are accepted; minimum 400×400 and maximum 8MB each. | US-O02 |
| R4 | Server converts images to WebP at 400/800/1600 widths and removes EXIF. | US-O02 |
| R5 | Submit validates completeness and moves draft to `pending`; approved-offer edits return it to `pending` and hide it. | US-O03, US-O04 |
| R6 | Merchant can pause, archive, duplicate, and delete drafts; approved offers are archived instead of hard-deleted. | US-O05, US-O06 |
| R7 | Dashboard shows per-offer views, WhatsApp clicks, 30-day series, and totals. | US-O07 |

## 6. Flow

```text
/dashboard/offers ──new──▶ draft editor ──images──▶ submit
       ▲                         │                    │
       └──archive/duplicate──────┴────────────── pending
                                                     │
                                           admin approve/reject
                                                     │
                                              public/hidden
```

## 7. Technical notes

- Only an approved merchant can create or submit offers; service layer repeats this check.
- If both prices exist, the server calculates the percentage. If not, percentage 1–99 is required.
- End date must follow start date and the offer may not exceed 180 days.
- Image ownership is checked before every media mutation.
- Public visibility remains a database view rule: approved offer, active dates, approved merchant, active access.

## 8. Subtasks

- [ ] T3.1 Build offer CRUD service, ownership checks, validation, and state machine.
- [ ] T3.2 Build mobile offer editor, autosave, validation messages, and unsaved-change handling.
- [ ] T3.3 Build image upload/processing pipeline, ordering, cover selection, and cleanup jobs.
- [ ] T3.4 Build merchant offer list, bulk-friendly actions, duplicate flow, and rejection feedback.
- [ ] T3.5 Build event aggregation and stats dashboard.
- [ ] T3.6 Add integration and E2E tests for draft → submit → review → public.

## 9. Success metrics

- Draft-to-submission completion rate.
- Median time to submit a valid offer.
- Image processing failure rate.
- Review rejection rate caused by validation versus merchant content.

## 10. Risks & open questions

- Large image processing can exhaust CPU; apply per-user limits and worker processing where needed.
- Changing an approved offer hides it until review, which may surprise merchants; the warning must be explicit.
- Confirm whether active-offer limit is global or configurable per merchant tier.

---

# PRD-04 · Trial, Billing & Payments

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 4 — Trial, Billing & Payments                                  |
| **Stories**  | US-B01 … US-B08 (8)                                             |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/billing`, payment success/failure, `apps/api/internal/billing`, worker |
| **API**      | `POST /merchant/billing/checkout`, `POST /merchant/billing/manual`, `GET /merchant/billing/payments`, `GET /merchant/billing/payments/:id`, `POST /webhooks/chargily` |

## 1. Summary

Every merchant receives one 60-day trial when the first offer is approved. After that, the merchant can pay 4000 DZD for 365 days through hosted Chargily Checkout or submit a manual transfer proof for admin review.

## 2. Problem

Access controls must be financially correct and resilient to retries, redirects, duplicate webhooks, failed payments, and expired subscriptions. Card data must never enter Dinar DZ.

## 3. Goals

- Start trial atomically with approval of the first offer.
- Activate a paid subscription only from a verified, idempotent payment event.
- Keep card/PAN/CVV/expiry data entirely outside Dinar DZ.
- Hide public offers immediately when access expires and restore them after valid activation.

### Non-goals

- Product payments, refunds automation, stored cards, customer wallets, or recurring card charges in V1.

## 4. Users

**Verified merchant** — wants uninterrupted access to publish offers and see a clear payment status.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | First approved offer starts a single 60-day trial inside the approval transaction. | US-B01 |
| R2 | Dashboard shows trial/subscription status and warning at 14 days or less. | US-B02 |
| R3 | Worker sends deduplicated reminders at 14, 7, 1, and 0 days. | US-B03 |
| R4 | Annual Chargily Checkout is created server-side for exactly 4000 DZD. | US-B04 |
| R5 | Success redirect never activates access; only a valid signed webhook can do so. | US-B04 |
| R6 | Manual payment accepts an 8MB proof and reference, then waits for admin approval. | US-B05 |
| R7 | Merchant sees payment history and receipt metadata without sensitive payment data. | US-B06 |
| R8 | Expiry hides offers; renewal starts at `max(now, access_until)` and restores valid offers. | US-B07, US-B08 |

## 6. Flow

```text
first offer approved ──▶ trialing (60d)
                              │
                ┌─────────────┴─────────────┐
                v                           v
       Chargily Checkout              manual proof
                │                           │
          signed webhook                 admin review
                └─────────────┬─────────────┘
                              v
                     subscribed (365d)
                              │
                         expired/renewed
```

## 7. Technical notes

- `payment.md` is the detailed payment security contract.
- Store transaction identifiers and statuses only; never PAN, CVV, expiry, screenshots, or raw sensitive webhook payloads.
- Verify Chargily HMAC on the raw request body before JSON parsing; then validate event, checkout ID, amount, currency, mode, metadata, and ownership.
- Use unique provider event/checkout constraints, row locks, and idempotent state transitions.
- Valid unknown events may be acknowledged and recorded as ignored; transient database failures return 500 for retry.
- Trial and subscription updates must update `access_until` atomically.

## 8. Subtasks

- [ ] T4.1 Implement access-status calculation and trial start transaction.
- [ ] T4.2 Build billing page, status banner, checkout CTA, manual-payment form, and payment history.
- [ ] T4.3 Integrate Chargily Test/Live configuration and hosted Checkout creation.
- [ ] T4.4 Implement signed webhook receiver, replay protection, reconciliation, and audit events.
- [ ] T4.5 Implement manual-payment review and subscription activation.
- [ ] T4.6 Implement reminders, expiry behavior, receipts, and billing E2E tests.

## 9. Success metrics

- Payment confirmation accuracy and webhook processing success rate.
- Duplicate-webhook rate that causes zero duplicate activations.
- Trial-to-paid conversion rate.
- Number of sensitive card fields found in DB, logs, backups, or client payloads: target zero.

## 10. Risks & open questions

- Chargily event names and request field names must be confirmed in the live account before release.
- Legal/tax requirements for merchant subscription receipts remain open.
- Manual payment review capacity may become a bottleneck as merchant count grows.

---

# PRD-05 · Admin Review & Audit

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 5 — Admin Review & Audit                                        |
| **Stories**  | US-A01 … US-A04, US-A10 (5)                                     |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/web/app/(admin)/admin`, `merchants`, `merchants/[id]`, `offers`, `payments`, `audit-logs` |
| **API**      | `GET /admin/dashboard`, `GET /admin/merchants`, `GET /admin/merchants/:id`, `GET /admin/documents/:id/url`, merchant actions, offer review actions, `GET /admin/audit-logs` |

## 1. Summary

Admins review merchant identity, offers, and payments through queues ordered by age, with mandatory reasons for rejection and a tamper-resistant audit trail for sensitive actions.

## 2. Problem

The platform depends on human trust decisions. Admin actions must be explicit, reversible where possible, permissioned, and explainable to merchants.

## 3. Goals

- Give admins a complete queue with filters, search, pagination, and detail pages.
- Make document viewing temporary and auditable.
- Make rejection reasons actionable for merchants.
- Ensure every sensitive decision is recorded with actor, time, IP, and request ID.

### Non-goals

- Customer moderation accounts or delegated admin roles in V1.
- Permanent deletion of audit records through the UI.

## 4. Users

**Platform admin** — reviews and governs merchants, offers, payments, and sensitive documents.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Admin access requires Google login plus an email in `ADMIN_EMAILS`; API and UI both enforce role. | US-A01 |
| R2 | Merchant queue shows oldest pending first with status/search filters and complete detail page. | US-A02 |
| R3 | Documents open through five-minute signed URLs; each access creates an audit event. | US-A02 |
| R4 | Admin can approve, reject with mandatory reason/document, suspend, and reactivate merchants. | US-A02 |
| R5 | Offer queue supports preview, approve, reject, remove, and bulk actions; first approval can start trial. | US-A03 |
| R6 | Manual payments show proof, merchant, amount, and approve/reject reason workflow. | US-A04 |
| R7 | Audit log is append-only from the application perspective and cannot be edited in the UI. | US-A10 |

## 6. Flow

```text
/admin ──▶ queue ──▶ detail ──▶ approve/reject/suspend
   │          │          │
   │          └── bulk actions
   └──────────────────────▶ audit log
```

## 7. Technical notes

- Admin endpoints use `RequireAuth` and `RequireRole(admin)`; frontend guards are not security boundaries.
- Signed document URLs expire after five minutes and never appear in email or permanent page state.
- Rejection reasons are required for merchant, offer, and payment rejection.
- Bulk operations return `processed` and per-ID failures instead of silently dropping records.
- Audit events cover approvals, rejections, suspensions, document views, settings changes, and payment actions.

## 8. Subtasks

- [ ] T5.1 Build admin shell, role guard, dashboard metrics, queues, filters, and pagination.
- [ ] T5.2 Build merchant detail and document viewer with signed URL and audit event.
- [ ] T5.3 Build merchant review actions and notification jobs.
- [ ] T5.4 Build offer review queue, bulk actions, and trial-start integration.
- [ ] T5.5 Build audit log queries, append-only policy, and authorization tests.

## 9. Success metrics

- Median pending-review age.
- Percentage of decisions with actionable reasons.
- Review SLA: target within 24 hours.
- Zero unauthorized admin actions in security testing.

## 10. Risks & open questions

- A single admin email list may not scale; role management may need a dedicated admin table later.
- Document privacy rules and legal retention must be approved before real ID uploads.

---

# PRD-06 · Admin Catalog, Reports & Ads

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 6 — Admin Catalog, Reports & Ads                                |
| **Stories**  | US-A05 … US-A09, US-D01 (6)                                     |
| **Priority** | Medium                                                          |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/web/app/(admin)/categories`, `ads`, `reports`, `settings`, public `AdSlot` |
| **API**      | CRUD `/admin/categories`, CRUD `/admin/ads`, `GET /admin/reports`, `POST /admin/reports/:id/resolve`, `GET/PUT /admin/settings`, `GET /ads`, `POST /ads/:id/click` |

## 1. Summary

Admins manage the taxonomy, direct-sold advertisements, abuse reports, and operational settings that control marketplace behavior without code deployment.

## 2. Problem

Categories, ad placements, reports, and product limits change operationally. Hard-coding them slows moderation and makes the public catalogue inconsistent.

## 3. Goals

- Provide safe CRUD for categories and ads.
- Keep category hierarchy to two levels and prevent deleting referenced categories.
- Resolve reports with explicit actions and notes.
- Make ad display measurable and layout-stable.

### Non-goals

- AdSense or automated ad auctions in V1.
- More than two category levels.

## 4. Users

**Admin/operator** — configures the marketplace and responds to reports. **Guest** — sees clearly labeled direct ads.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Admin can add, edit, reorder, and hide two-level categories; referenced categories cannot be deleted. | US-A05 |
| R2 | Reports are grouped by offer and show reason, message, distinct IP count, and status. | US-A06 |
| R3 | Admin can ignore a report, hide the offer, or suspend the merchant. | US-A06 |
| R4 | Ads support advertiser, image, target URL, slot, dates, priority, wilaya/category targeting, impressions, and clicks. | US-A07 |
| R5 | Dashboard provides merchant, offer, access, revenue, and top-category indicators. | US-A08 |
| R6 | Settings control trial days, annual price, subscription days, image/offer limits, and manual payment details. | US-A09 |
| R7 | Public ads use fixed slots, reserved aspect ratio, “Advertisement” label, and sponsored links. | US-D01 |

## 6. Flow

```text
admin categories/settings/ads/reports
          │
          ├── publish configuration ──▶ cached public API
          ├── resolve report ─────────▶ offer/merchant action
          └── inspect metrics ────────▶ dashboard
```

## 7. Technical notes

- Public ad selection filters by slot, date, wilaya, category, visibility, and priority.
- `AdSlot` reserves space before loading to protect CLS and records one viewport impression.
- Target links use `rel="sponsored noopener"` and are validated as HTTPS URLs.
- Settings changes are audited and validated against safe bounds.

## 8. Subtasks

- [ ] T6.1 Build category CRUD, hierarchy validation, ordering, visibility, and seed data.
- [ ] T6.2 Build report aggregation, status transitions, actions, and merchant/offer side effects.
- [ ] T6.3 Build ad CRUD, media validation, targeting, slots, impression/click aggregation.
- [ ] T6.4 Build admin dashboard metrics and settings UI/API.
- [ ] T6.5 Build public `AdSlot` with lazy loading, reserved dimensions, and sponsored tracking.

## 9. Success metrics

- Percentage of reports resolved within target SLA.
- Ad fill rate, viewability, CTR, and zero layout shift caused by ads.
- Category search success and percentage of offers with valid taxonomy.

## 10. Risks & open questions

- Direct ad inventory requires an owner for advertiser contracts and payment reconciliation.
- Dashboard numbers must define timezone and aggregation window consistently.

---

# PRD-07 · System Lifecycle

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | 7 — System Lifecycle                                             |
| **Stories**  | US-S01 … US-S04 (4)                                             |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/api/internal/jobs`, `apps/api/cmd/worker`, SEO routes, backup runner |
| **API**      | `GET /healthz`, `GET /readyz`, `GET /sitemap.xml`, `GET /robots.txt`, `GET /sitemap/offers` |

## 1. Summary

Background jobs, notifications, SEO generation, and backups keep the marketplace current and recoverable after the request/response path has finished.

## 2. Problem

Offer expiry, reminders, email, sitemap freshness, and backup operations cannot depend on a merchant or admin keeping a page open.

## 3. Goals

- Make time-based behavior deterministic and idempotent.
- Notify the right party without duplicate mail.
- Keep public SEO data limited to visible content.
- Produce and test recoverable database backups.

### Non-goals

- Push notifications, WhatsApp notifications, or video processing in V1.

## 4. Users

Merchants, admins, search engines, and operators who depend on current system state.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Public queries hide offers immediately after `ends_at`; nightly jobs archive old offers and clean large media according to retention. | US-S01 |
| R2 | Merchant/admin emails cover review, offer, payment, and access events with deduplicated notifications. | US-S02 |
| R3 | SSR SEO includes dynamic sitemap, robots rules, canonical URLs, and Offer JSON-LD. | US-S03 |
| R4 | Daily encrypted `pg_dump` is stored off-host for at least 14 days and restored in a monthly drill. | US-S04 |

## 6. Flow

```text
cron/queue ──▶ claim job ──▶ execute ──▶ record result
    │              │             │
    ├── expiry     ├── email     ├── sitemap
    └── backup     └── cleanup   └── alert on failure
```

## 7. Technical notes

- Use a jobs table with idempotency key, attempts, lock time, and next run.
- Expiry visibility is query-based; the job is for archival and cleanup, not correctness.
- Notification uniqueness prevents duplicate reminders for the same merchant/access milestone.
- Worker is the same Go binary in a separate `worker` container.
- Backup files must be encrypted, access-controlled, rotated, and periodically restored.

## 8. Subtasks

- [ ] T7.1 Implement job table, worker loop, retry policy, locks, and dead-letter handling.
- [ ] T7.2 Implement expiry archival, media cleanup, access reminders, and notification deduplication.
- [ ] T7.3 Implement email templates, SMTP provider, delivery logs, and admin alerts.
- [ ] T7.4 Implement sitemap/robots/JSON-LD generation from public visibility.
- [ ] T7.5 Implement encrypted backup, retention, restore drill, and failure alerts.

## 9. Success metrics

- Job success rate and retry age.
- Duplicate notification rate: zero for the same notification key.
- Backup completion and restore-test success rate.
- Sitemap freshness after offer approval/removal.

## 10. Risks & open questions

- SMTP provider limits and bounce handling need an operational owner.
- Confirm legal retention rules before deleting documents or media.
- Backup storage must be geographically and administratively separate from the application host.

---

# PRD-08 · Data & API Foundation

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | Platform enabler — data and API                                  |
| **Stories**  | Enabler for US-V01 … US-S04                                     |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/api/internal`, `apps/api/db/migrations`, `apps/api/db/queries`, PostgreSQL 16 |
| **API**      | All `/api/v1` endpoints, `GET /healthz`, `GET /readyz`           |

## 1. Summary

This foundation provides PostgreSQL schema, migrations, transactions, indexes, Arabic search normalization, API envelopes, authorization middleware, and the single source of truth for public offer visibility.

## 2. Problem

Business rules become unsafe when duplicated across handlers, UI, and jobs. The project needs a stable data contract before feature work can scale.

## 3. Goals

- Make business invariants enforceable in database constraints and service transactions.
- Keep Go service logic independent of Fiber and SQL details.
- Return consistent JSON, errors, pagination, and request IDs.
- Support Arabic search, public visibility, and critical payment/trial transactions.

### Non-goals

- Analytics warehouse, multi-region database, or customer accounts in V1.

## 4. Users

All product users indirectly depend on this layer; developers are its direct users.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Use UUIDs, UTC `timestamptz`, integer DZD amounts, snake_case names, enums, and database constraints. | database.md |
| R2 | Model users, sessions, merchants, documents, shops, offers, media, subscriptions, payments, reports, ads, stats, jobs, notifications, and audit logs. | database.md |
| R3 | Implement `public_offers` so an offer is public only when status, dates, merchant verification, and access are all valid. | userstory.md/database.md |
| R4 | Implement Arabic normalization, generated search vectors, indexes, and pagination limits. | database.md/backend.md |
| R5 | Implement API success/error envelopes, request IDs, auth/role/CSRF/rate-limit middleware, and `401/403/404/410/409/422/429/500`. | backend.md/userstory.md |
| R6 | Keep first-offer trial, subscription activation, and webhook idempotency atomic. | database.md/payment.md |

## 6. Flow

```text
HTTP request ─▶ middleware ─▶ handler ─▶ service/transaction ─▶ repository/sqlc ─▶ response
                                      │
                                      └── audit/job event when required
```

## 7. Technical notes

- Layering is handler → service → repository; handlers contain no SQL or business decisions.
- `sqlc` owns generated query code; migrations are forward-only and tested on empty and populated databases.
- Use `SELECT ... FOR UPDATE` for state transitions that affect access or payment.
- Index public filters, `ends_at`, statuses, foreign keys, search vectors, and time-series keys.
- Never return private storage keys, session hashes, or sensitive document metadata.

## 8. Subtasks

- [ ] T8.1 Create migration runner, extensions, enums, base reference data, and seed process.
- [ ] T8.2 Create core schema, constraints, triggers, `public_offers`, and sqlc queries.
- [ ] T8.3 Implement service/repository conventions, response envelopes, errors, pagination, and request IDs.
- [ ] T8.4 Implement middleware: auth, roles, CSRF, rate limits, recovery, secure headers, and trusted proxy handling.
- [ ] T8.5 Implement Arabic search, visibility queries, critical transactions, and database integration tests.

## 9. Success metrics

- API p95 under 300ms for public lists at V1 capacity.
- Zero duplicate trial/subscription activation in concurrency tests.
- Migration success on a fresh database and a representative populated database.
- Consistent error contract across all endpoints.

## 10. Risks & open questions

- Wilaya count and official seed data must be confirmed.
- PostgreSQL full-text search may need a later ranking/tuning pass using real Arabic queries.
- Database growth from events and stats requires retention and aggregation policies.

---

# PRD-09 · Web Shell, Brand & Splash

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | Platform enabler — web experience and identity                      |
| **Stories**  | Cross-cutting support for US-V01 … US-S04                       |
| **Priority** | High                                                            |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `apps/web/app/layout.tsx`, route groups, `components`, `styles/globals.css`, `site/branding` |
| **API**      | `GET /site/branding`, public APIs used by route groups             |

## 1. Summary

The web shell establishes Dinar DZ’s Arabic RTL identity, responsive design system, shared navigation, loading states, and database-driven splash screen before feature pages are assembled.

## 2. Problem

Without a coherent shell, each feature can drift in color, RTL behavior, accessibility, loading, and brand expression. The splash must communicate identity without hiding a slow page indefinitely.

## 3. Goals

- Apply the Dinar DZ visual system consistently across public, merchant, and admin surfaces.
- Deliver an accessible Arabic-first responsive shell.
- Show logo, site name, database proverb, and a truthful progress bar in Splash.
- Keep loading and transitions within the performance budget.

### Non-goals

- Final logo redesign, French content completion, or native mobile application.

## 4. Users

Every visitor, merchant, and admin sees the shell; brand administrators manage the active proverb and logo reference.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Use `Dinar DZ`/`Dinar`, burgundy `#6C151E`, dark green `#0F3D3A`, cream, orange-red accent, and gold according to the design system. | ui.md |
| R2 | Use Arabic RTL, logical CSS properties, Cairo/Tajawal or approved equivalent, and accessible focus/touch targets. | ui.md/frontend.md |
| R3 | Shared header, bottom navigation, buttons, cards, forms, modals, states, and typography are implemented as reusable components. | ui.md/frontend.md |
| R4 | Splash reads active `dinar_dz` branding/proverb from DB, shows progress phases, reaches 100% only at page readiness, and remains 3 seconds afterward. | splash screen.md |
| R5 | Splash fallback, reduced motion, timeout, error state, and no-CLS behavior are covered. | splash screen.md |
| R6 | Public routes are SSR/SEO-aware; merchant/admin route groups are separated for bundle and authorization boundaries. | frontend.md |

## 6. Flow

```text
app boot ─▶ branding + page data ─▶ page_ready ─▶ 3s hold ─▶ fade out ─▶ route usable
```

## 7. Technical notes

- Splash uses `position: fixed` and does not block SSR; deferred images, ads, and analytics do not delay removal.
- `site_branding` stores `site_key`, locale, proverb, logo URL/alt, active flag, and version.
- Sanitize DB text as plain text; allowlist logo URL origins and apply CSP.
- Use `prefers-reduced-motion`, visible progress text, `role="progressbar"`, and `aria-live="polite"`.
- Keep homepage JavaScript under the specified budget and avoid large icon/utility bundles.

## 8. Subtasks

- [ ] T9.1 Build root layout, fonts, theme tokens, RTL, metadata, and shared navigation.
- [ ] T9.2 Build reusable design components and states from `ui.md`.
- [ ] T9.3 Add `site_branding` API, admin-editable proverb, cache, fallback, and audit hook.
- [ ] T9.4 Implement Splash component, progress phases, three-second hold, fade, timeout, and accessibility.
- [ ] T9.5 Add responsive visual QA at 360px, tablet, desktop, reduced-motion, and 200% text scale.

## 9. Success metrics

- CLS under 0.1 and no layout jump during Splash removal.
- Lighthouse accessibility score and automated axe pass for core routes.
- Brand consistency review across public, merchant, and admin screenshots.

## 10. Risks & open questions

- Holding a Splash for three seconds after readiness is a deliberate product choice but can reduce perceived speed; measure it and make the timing configurable.
- Confirm final logo SVG outline and legal brand availability before launch.

---

# PRD-10 · Delivery, Security & Quality

|              |                                                                 |
| ------------ | --------------------------------------------------------------- |
| **Epic**     | Platform enabler — delivery, security, and quality              |
| **Stories**  | Enabler for all P0/P1 stories                                   |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                      |
| **Surfaces** | `Dockerfile`, `docker-compose.yml`, `Caddyfile`, `redeploy.sh`, CI, logs/monitoring |
| **API**      | `GET /healthz`, `GET /readyz`; all deployed services            |

## 1. Summary

Docker, Caddy, CI, automated redeploy, secrets, backups, health checks, observability, and test gates turn the product specifications into a safely deployable service.

## 2. Problem

Features are not production-ready without repeatable builds, secure secrets, rollback, database recovery, monitoring, and regression tests.

## 3. Goals

- Build web, API, worker, database, and proxy services reproducibly.
- Redeploy after code changes through `redeploy.sh` or CI.
- Protect secrets, private documents, payments, sessions, and logs.
- Gate release on unit, integration, E2E, accessibility, security, and health checks.

### Non-goals

- Multi-region failover, Kubernetes, or zero-downtime database migrations in V1.

## 4. Users

Developers, operators, admins, and end users who depend on availability and data safety.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Compose services run Caddy, Next.js web, Go API, worker, and PostgreSQL with persistent DB/Caddy volumes. | deployment.md |
| R2 | Production secrets are environment/secret-manager values, never source code, image layers, or client bundles. | deployment.md/backend.md/payment.md |
| R3 | `redeploy.sh` validates Compose, builds with `--pull`, starts services, optionally runs migrations, prints status, and can run an HTTP health check. | redeploy.sh/deployment.md |
| R4 | HTTPS, HSTS, secure headers, CSP, CSRF, rate limits, private buckets, signed document URLs, and webhook HMAC are enabled. | userstory.md/backend.md/payment.md |
| R5 | CI runs unit, integration with real PostgreSQL, Playwright RTL mobile flows, axe, bundle/performance checks, and migration checks. | frontend.md/backend.md/userstory.md |
| R6 | Logs are JSON and redacted; alerts cover 5xx, webhook failures, disk, DB, backups, and suspicious payment mismatches. | deployment.md/payment.md |
| R7 | Daily encrypted off-host backup is retained 14 days and restored monthly. | deployment.md/userstory.md |

## 6. Flow

```text
git push ─▶ CI tests/builds ─▶ versioned image ─▶ redeploy.sh
                                                │
                           compose up ─▶ migrate ─▶ health checks
                                                │
                                      success or rollback
```

## 7. Technical notes

- `redeploy.sh` remains in the project root; use `bash redeploy.sh` where executable permission is unavailable.
- Caddy is the only public service; API/database/storage are private network services.
- Image tags are immutable for rollback; do not use `latest` as the only release reference.
- Migrations run forward-only with a backup before risky changes; rollback does not automatically reverse migrations.
- Never log OAuth tokens, session values, card data, raw private documents, or raw payment payloads.
- Health endpoint checks process liveness; readiness checks database and storage.

## 8. Subtasks

- [ ] T10.1 Add Dockerfiles, Compose services, Caddy routing, health checks, and persistent volumes.
- [ ] T10.2 Add CI pipeline for lint, tests, migrations, build, security scan, and artifact tagging.
- [ ] T10.3 Validate and document `redeploy.sh`, environment files, migration command, and rollback procedure.
- [ ] T10.4 Add secrets policy, log redaction, CSP/HSTS/secure headers, dependency updates, and vulnerability scanning.
- [ ] T10.5 Add backup/restore automation, monitoring dashboards, alert routing, and incident runbook.
- [ ] T10.6 Add release checklist and production smoke tests for guest, merchant, admin, and payment paths.

## 9. Success metrics

- Successful repeatable deployment from a clean checkout.
- Health check recovery and rollback time.
- Monthly availability target of 99.5% for V1.
- Zero secrets or sensitive payment/document fields in repository and logs.
- Successful monthly restore drill and passing release smoke tests.

## 10. Risks & open questions

- The current workspace has no application code or Compose file; infrastructure implementation is still required.
- Final hosting provider, registry, SMTP provider, R2/S3 account, and alert destination must be selected.
- A production migration strategy is needed before high-volume data changes.

---

## Cross-PRD Definition of Done

Every subtask is complete only when:

1. Acceptance requirements and authorization rules are implemented server-side.
2. Unit and integration tests cover success, validation, unauthorized, forbidden, not-found, rate-limit, and failure paths.
3. Critical user journeys have Playwright coverage in RTL at 360px and desktop widths.
4. No sensitive data appears in browser storage, logs, error responses, backups, or public URLs.
5. Database migrations work on a clean database and a populated database.
6. Performance, accessibility, SEO, and security budgets are checked in CI.
7. The feature has monitoring signals, a documented failure state, and a rollback/recovery path where applicable.
8. The related user story is updated to `✅ Shipped` only after production verification, not after code merge alone.

## Open decisions before implementation

- Confirm the official wilaya dataset and supported language rollout.
- Confirm Chargily live event names, webhook field names, and account verification.
- Confirm legal retention/consent obligations for Algerian identity documents.
- Select production host, container registry, object storage, SMTP, monitoring, and backup providers.
- Decide whether the requested three-second post-ready Splash hold remains after measuring bounce and perceived performance.
