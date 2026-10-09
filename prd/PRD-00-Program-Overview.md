# Program Overview

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Dinar DZ Product Program                                         |
| **Stories**  | All product stories and platform enablers                         |
| **Priority** | Critical                                                         |
| **Status**   | ✅ Done                                                           |
| **Surfaces** | `prd/`, `userstory.md`, `skills/`, CI/CD                         |
| **API**      | All `/api/v1` endpoints                                          |

## 1. Summary

Dinar DZ is an Algeria-first marketplace for merchant discounts. Guests browse and contact merchants; verified merchants publish moderated offers; admins operate trust, catalogue, ads, and payments.

## 2. Problem

The product needs a staged implementation plan that keeps data, security, API contracts, frontend, operations, and business rules aligned.

## 3. Goals

- Deliver the P0 path: guest discovery, merchant verification, offers, review, trial, and payment.
- Keep public visibility governed by one database rule.
- Ship mobile-first Arabic RTL UX with secure operations.

### Non-goals

Customer accounts, product checkout, video, native mobile apps, French content completion, and automated ads are V2 or later.

## 4. Users

Guests, merchants, admins, developers, and operators.

## 5. Requirements

| # | Requirement | PRD |
|---|---|---|
| R1 | Build data/API foundation first. | PRD-08 |
| R2 | Build guest and merchant journeys on the foundation. | PRD-01, PRD-02, PRD-03 |
| R3 | Add trial, Chargily/manual payments, and access enforcement. | PRD-04 |
| R4 | Add moderation, catalogue, reports, ads, and background jobs. | PRD-05, PRD-06, PRD-07 |
| R5 | Apply brand, accessibility, deployment, security, and quality gates to every PRD. | PRD-09, PRD-10 |

## 6. Flow

```text
Foundation → Identity → Offers → Review → Trial/Payment → Public growth
     │          │          │          │          │
     └──────────┴──────────┴──────────┴──────────┴── UX + Operations gates
```

## 7. Technical notes

- Source documents live in `../userstory.md` and `../skills/`.
- `prd.md` is the master portfolio; `PRD-00` to `PRD-10` are the execution slices.
- A story becomes shipped only after tests, security review, observability, and production verification.

## 8. Subtasks

- [x] T00.1 Confirm scope, legal assumptions, providers, and release owner. See [`decision-log.md`](../decision-log.md); unresolved production choices remain explicit release gates.
- [x] T00.2 Create project board from PRD subtasks and dependencies. See [`project-board.md`](../project-board.md).
- [x] T00.3 Define P0 release gate and pilot launch with 20–30 merchants. See [`release-gate.md`](../release-gate.md).
- [x] T00.4 Review every PRD after implementation decisions change the source specifications. See [`prd-review-matrix.md`](prd-review-matrix.md).

## 9. Success metrics

- All 42 user stories mapped to an owning PRD.
- P0 journey works end-to-end in production-like staging.
- No unresolved Critical security or data-integrity issue at launch.

## 10. Risks & open questions

- Final wilaya dataset, legal privacy obligations, Chargily live fields, hosting providers, and tax/receipt requirements remain release gates before production. They do not block completion of the EPIC-00 planning deliverables.

## 11. EPIC-00 completion evidence

The confirmed V1 scope, technical provider baseline, legal assumptions, release
ownership model, open decisions, and release gate are maintained in the
[project decision log](../decision-log.md). The complete execution evidence is:

- [Decision log](../decision-log.md) — T00.1 scope and provider baseline.
- [Project board](../project-board.md) — T00.2 work sequence and dependencies.
- [P0 release gate](../release-gate.md) — T00.3 pilot exit criteria and rollout.
- [PRD review matrix](prd-review-matrix.md) — T00.4 cross-PRD alignment review.

---

# PRD-01 · Guest Experience

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 1 — Guest Experience                                             |
| **Stories**  | US-V01 … US-V07 (7)                                               |
| **Priority** | High                                                             |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(public)/page.tsx`, `offers`, `offers/[id]`, `shop/[slug]` |
| **API**      | `GET /offers`, `GET /offers/:id`, `GET /offers/:id/similar`, `POST /offers/:id/view`, `POST /offers/:id/whatsapp`, `POST /offers/:id/report`, `GET /shops/:slug`, `GET /categories`, `GET /wilayas`, `GET /ads` |

## 1. Summary

Guests browse current discounts, filter them using Arabic search, open details, view a shop, share an offer, report abuse, and contact a merchant through WhatsApp without registration.

## 2. Problem

Visitors must see real value before signup. Stale, hidden, or inaccessible offers reduce trust for both guests and merchants.

## 3. Goals

- Show only `public_offers` results.
- Reach detail in one click and keep filters shareable.
- Make WhatsApp contact one tap on mobile.

### Non-goals

Customer accounts, cart, product payment, favorites, and native mobile apps.

## 4. Users

Guest visitor arriving from search, a shared link, or a local merchant promotion.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Home shows search, categories, latest, ending-soon, and highest-discount sections. | US-V01 |
| R2 | Cards show cover, title, prices, discount, shop, wilaya, and countdown. | US-V01 |
| R3 | Search supports Arabic normalization, filters, sorting, URL state, empty state, and pagination. | US-V02 |
| R4 | Detail shows gallery, description, dates, shop, verified badge, similar offers, OG metadata, 404/410 states. | US-V03 |
| R5 | WhatsApp/phone actions record safe, deduplicated analytics. | US-V04 |
| R6 | Shop page shows public shop details and active offers only. | US-V05 |
| R7 | Reports enforce reason, message limit, IP rate limit, and duplicate protection. | US-V06 |
| R8 | Share uses Web Share API or clipboard. | US-V07 |

## 6. Flow

```text
Home → search/filter → offers → detail → WhatsApp/report/share
                         └────→ shop page
```

## 7. Technical notes

SSR the first view; use `public_offers`; normalize Arabic on the server; lazy-load secondary images and ads; never expose hidden content in 404/410 responses.

## 8. Subtasks

- [ ] T1.1 Build public routes, cards, filters, detail gallery, shop page, and states.
- [ ] T1.2 Implement public queries, Arabic search, pagination, and URL filters.
- [ ] T1.3 Implement WhatsApp, share, view/click deduplication, reports, and rate limits.
- [ ] T1.4 Add SEO metadata and Playwright guest journey.

## 9. Success metrics

LCP under 2.5s; offer-detail CTR; WhatsApp CTR; guest share of public offers; report SLA.

## 10. Risks & open questions

Search quality depends on Arabic seed data. Confirm official wilaya data and define similar-offer ranking before launch.

---

# PRD-02 · Merchant Identity & Verification

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 2 — Merchant Identity & Verification                             |
| **Stories**  | US-M01 … US-M05 (5)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(auth)/login`, `onboarding/personal`, `shop`, `documents`, `review` |
| **API**      | Google auth, `/me`, `PUT /merchant/profile`, `PUT /merchant/shop`, `POST /merchant/documents`, `POST /merchant/submit` |

## 1. Summary

A merchant signs in with Google, completes personal/shop data, uploads identity and commercial-register documents, accepts privacy consent, and follows verification status.

## 2. Problem

The marketplace needs trustworthy merchants while protecting sensitive identity documents and keeping onboarding simple on mobile.

## 3. Goals

- Secure Google OAuth with verified email, state, PKCE, and HttpOnly sessions.
- Autosave onboarding and resume the last incomplete step.
- Keep documents private and review status actionable.

### Non-goals

Passwords, customer accounts, automatic document approval, and third-party KYC in V1.

## 4. Users

Prospective merchant and admin reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Google OAuth creates/resolves a merchant session and rejects unverified email. | US-M01 |
| R2 | Personal information validates Algerian phone, wilaya, address, and autosaves. | US-M02 |
| R3 | Shop information captures identity, location, contact channels, logo, and register expiry. | US-M03 |
| R4 | Slug is unique and editable once before verification. | US-M03 |
| R5 | Three documents accept only verified content types up to 8MB in private storage. | US-M04 |
| R6 | Consent timestamp is stored; status supports incomplete/pending/approved/rejected/suspended. | US-M04, US-M05 |

## 6. Flow

```text
Google login → personal → shop → documents → review → pending → approved/rejected
```

## 7. Technical notes

Store only session hash; use secure cookies; detect document MIME by magic bytes; return signed five-minute document URLs only to admins; audit every document view; apply CSRF except the external OAuth callback.

## 8. Subtasks

- [ ] T2.1 Implement OAuth, sessions, cookies, logout, and role middleware.
- [ ] T2.2 Build onboarding wizard, autosave, validation, resume, and consent.
- [ ] T2.3 Implement private upload/storage, replacement, hashing, and rejection flow.
- [ ] T2.4 Add status emails and authorization/integration tests.

## 9. Success metrics

Onboarding completion rate; time to submission; rejection reasons; zero unauthorized document reads.

## 10. Risks & open questions

Confirm Algerian privacy/legal retention obligations and the final admin-role management policy.

---

# PRD-03 · Merchant Offers

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 3 — Merchant Offers                                               |
| **Stories**  | US-O01 … US-O07 (7)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/offers`, `offers/new`, `offers/[id]/edit`, `stats` |
| **API**      | `/merchant/offers`, `/merchant/offers/:id`, media endpoints, submit/archive/duplicate, `/merchant/stats` |

## 1. Summary

Approved merchants create mobile-first offer drafts, upload up to ten optimized images, submit them for review, manage their lifecycle, and inspect privacy-safe statistics.

## 2. Problem

Publishing must be fast for merchants but strict enough to protect customers from incomplete, expired, misleading, or oversized content.

## 3. Goals

- Make a valid offer possible from a phone.
- Autosave drafts and enforce all business rules on the server.
- Optimize images and preserve review history.

### Non-goals

Video, product checkout, unlimited active offers, and permanent deletion of approved offers.

## 4. Users

Approved merchant and admin reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Validate title, description, category, prices/discount, dates, location, and WhatsApp. | US-O01 |
| R2 | Autosave every 10 seconds; enforce configurable active-offer limit. | US-O01 |
| R3 | Accept 1–10 JPG/PNG/WebP images, max 8MB, min 400×400. | US-O02 |
| R4 | Convert server-side to WebP 400/800/1600 and remove EXIF. | US-O02 |
| R5 | Submit moves to pending; any approved-offer edit hides it and returns it to pending. | US-O03, US-O04 |
| R6 | Support pause, archive, draft delete, and duplicate. | US-O05, US-O06 |
| R7 | Show views, WhatsApp clicks, 30-day series, and totals. | US-O07 |

## 6. Flow

```text
dashboard → draft editor → media → submit → pending → admin approval → public
                                      └──────────────→ rejection → correction
```

## 7. Technical notes

Server calculates discount when both prices exist; validates 1–99 otherwise; limits duration to 180 days; checks ownership for every mutation; uses `public_offers` for public reads.

## 8. Subtasks

- [ ] T3.1 Implement offer state machine, CRUD, ownership, validation, and autosave.
- [ ] T3.2 Build editor, media uploader, ordering, cover selection, and rejection UI.
- [ ] T3.3 Implement image processing, cleanup, duplicate, archive, and stats aggregation.
- [ ] T3.4 Add merchant and moderation E2E flows.

## 9. Success metrics

Draft-to-submit completion; median creation time; image failure rate; invalid/rejected offer rate.

## 10. Risks & open questions

Image processing can consume CPU. Confirm the active-offer limit and whether processing should be worker-backed.

---

# PRD-04 · Trial, Billing & Payments

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 4 — Trial, Billing & Payments                                    |
| **Stories**  | US-B01 … US-B08 (8)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/billing`, `apps/api/internal/billing`, worker |
| **API**      | Billing checkout/manual/history/status and `POST /webhooks/chargily` |

## 1. Summary

The first approved offer starts one 60-day trial. Merchants then pay 4000 DZD for 365 days through hosted Chargily Checkout or submit a manual proof for admin approval.

## 2. Problem

Money and access state must remain correct through redirects, duplicate webhooks, expiry, retries, and manual review while card data stays outside Dinar DZ.

## 3. Goals

- Start trial atomically and only once.
- Activate paid access only after verified webhook/admin review.
- Never store PAN, CVV, expiry, screenshots, or raw sensitive payloads.
- Hide and restore offers based on access state.

### Non-goals

Product payments, stored cards, recurring card charges, automated refunds, or customer wallets.

## 4. Users

Verified merchant and admin payment reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | First approved offer starts one 60-day trial in the same transaction. | US-B01 |
| R2 | Dashboard shows trial/subscription countdown and 14-day warning. | US-B02 |
| R3 | Reminder jobs at 14/7/1/0 days are deduplicated. | US-B03 |
| R4 | Server creates a 4000 DZD hosted Chargily Checkout. | US-B04 |
| R5 | Success redirect never activates access; signed HMAC webhook is required. | US-B04 |
| R6 | Manual proof/reference enters under-review until admin decision. | US-B05 |
| R7 | Payment history and receipt metadata contain no card data. | US-B06 |
| R8 | Expiry hides offers; renewal starts at `max(now, access_until)`. | US-B07, US-B08 |

## 6. Flow

```text
first approval → trial → hosted Checkout/manual proof → verified payment → subscribed → expiry/renewal
```

## 7. Technical notes

See `../skills/payment.md`. Verify raw-body HMAC before parsing; validate event ID, checkout ID, amount, currency, mode, metadata, ownership, and status; use unique constraints, row locks, and idempotent transitions.

## 8. Subtasks

- [ ] T4.1 Implement access/trial/subscription calculations and atomic transitions.
- [ ] T4.2 Build billing UI, history, manual proof, success/failure, and pending states.
- [ ] T4.3 Integrate Chargily Test/Live Checkout and signed webhook processing.
- [ ] T4.4 Implement manual review, reminders, expiry, receipts, reconciliation, and tests.

## 9. Success metrics

Payment confirmation accuracy; zero duplicate activations; trial-to-paid conversion; zero card fields in DB/logs/backups/client payloads.

## 10. Risks & open questions

Confirm Chargily live event and field names, tax/receipt requirements, and manual review capacity.

---

# PRD-05 · Admin Review & Audit

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 5 — Admin Review & Audit                                         |
| **Stories**  | US-A01 … US-A04, US-A10 (5)                                      |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(admin)/admin`, `merchants`, `offers`, `payments`, `audit-logs` |
| **API**      | Admin dashboard, merchant/offer/payment queues, signed document URL, actions, audit logs |

## 1. Summary

Admins review merchants, documents, offers, and payments through searchable queues with mandatory reasons and auditable sensitive actions.

## 2. Problem

Trust decisions must be explainable and protected. A hidden frontend button is not sufficient authorization.

## 3. Goals

- Provide age-ordered queues and complete detail views.
- Make document access temporary and auditable.
- Make every rejection actionable and every action traceable.

### Non-goals

Delegated admin roles, customer moderation accounts, and editable audit history.

## 4. Users

Platform admin/operator.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Admin role requires Google login and configured admin email. | US-A01 |
| R2 | Merchant queue supports search, status, oldest-first ordering, and full detail. | US-A02 |
| R3 | Documents use five-minute signed URLs and audit events. | US-A02 |
| R4 | Admin can approve, reject with reason/document, suspend, and reactivate. | US-A02 |
| R5 | Offer queue supports preview, approve, reject, remove, and bulk actions. | US-A03 |
| R6 | Payment queue supports proof review and reasoned approve/reject. | US-A04 |
| R7 | Audit records capture actor, action, resource, time, IP, and request ID. | US-A10 |

## 6. Flow

```text
admin login → queue → detail → decision → notification + audit
```

## 7. Technical notes

Use API role middleware; signed URLs never persist in page state; bulk responses include per-ID failures; audit writes are append-only from the application UI.

## 8. Subtasks

- [ ] T5.1 Build admin shell, guards, dashboard metrics, queues, search, and pagination.
- [ ] T5.2 Build merchant/document review and signed viewer.
- [ ] T5.3 Build offer/payment review and bulk operations.
- [ ] T5.4 Build audit queries, notifications, and authorization tests.

## 9. Success metrics

Review age; 24-hour SLA; actionable rejection rate; zero unauthorized admin actions.

## 10. Risks & open questions

An email allowlist may not scale. Confirm future admin role management and document retention policy.

---

# PRD-06 · Admin Catalog, Reports & Ads

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 6 — Admin Catalog, Reports & Ads                                 |
| **Stories**  | US-A05 … US-A09, US-D01 (6)                                      |
| **Priority** | Medium                                                          |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(admin)/categories`, `ads`, `reports`, `settings`, public `AdSlot` |
| **API**      | CRUD categories/ads, reports list/resolve, settings, public ads/click |

## 1. Summary

Admins operate taxonomy, reports, direct-sold ads, dashboard indicators, and business settings without code changes.

## 2. Problem

Operational configuration changes more frequently than application releases. Hard-coded categories and ads make moderation and monetization slow.

## 3. Goals

- Manage two-level categories safely.
- Resolve reports with explicit side effects.
- Run measurable, labeled, layout-stable direct ads.
- Expose configurable limits and business indicators.

### Non-goals

AdSense, ad auctions, and more than two category levels.

## 4. Users

Admin/operator and guest ad viewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Add/edit/reorder/hide categories; prevent deleting referenced categories. | US-A05 |
| R2 | Group reports by offer and expose reason, message, unique IP count, and status. | US-A06 |
| R3 | Ignore, hide offer, or suspend merchant from report resolution. | US-A06 |
| R4 | Ads support advertiser, media, URL, slot, dates, priority, targeting, impressions, and clicks. | US-A07 |
| R5 | Dashboard shows merchant, offer, access, revenue, and top-category indicators. | US-A08 |
| R6 | Settings control trial, price, subscription days, limits, and manual payment data. | US-A09 |
| R7 | Public ad slots reserve space, show “Advertisement”, and use sponsored links. | US-D01 |

## 6. Flow

```text
admin config → validation/audit → cached public API → ad/report/dashboard result
```

## 7. Technical notes

Validate HTTPS targets; use slot/date/targeting filters; aggregate metrics daily; reserve aspect ratio to protect CLS; audit settings changes.

## 8. Subtasks

- [ ] T6.1 Build category CRUD, hierarchy, ordering, visibility, and seeds.
- [ ] T6.2 Build report aggregation, resolution actions, and side effects.
- [ ] T6.3 Build ad CRUD, media, targeting, slot rendering, and metrics.
- [ ] T6.4 Build dashboard/settings UI and API.

## 9. Success metrics

Report resolution SLA; ad fill/viewability/CTR; zero ad-caused CLS; category completeness.

## 10. Risks & open questions

Define ad contracts, payment reconciliation, dashboard timezone, and metrics aggregation windows.

---

# PRD-07 · System Lifecycle

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 7 — System Lifecycle                                             |
| **Stories**  | US-S01 … US-S04 (4)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/api/internal/jobs`, `apps/api/cmd/worker`, SEO routes, backup runner |
| **API**      | `GET /healthz`, `GET /readyz`, `GET /sitemap.xml`, `GET /robots.txt`, `GET /sitemap/offers` |

## 1. Summary

Worker jobs, notifications, SEO generation, and backups keep the marketplace current, discoverable, and recoverable.

## 2. Problem

Expiry, reminders, email, sitemap updates, and backups cannot depend on a user keeping a page open.

## 3. Goals

- Make scheduled behavior idempotent and observable.
- Keep public SEO limited to visible content.
- Produce and test encrypted off-host backups.

### Non-goals

Push notifications, WhatsApp notifications, and video processing.

## 4. Users

Merchants, admins, search engines, and operators.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Query-time expiry hides offers; jobs archive old offers and clean media. | US-S01 |
| R2 | Review, offer, payment, and access emails are deduplicated. | US-S02 |
| R3 | Sitemap, robots, canonical URLs, and Offer JSON-LD are generated from public data. | US-S03 |
| R4 | Daily encrypted `pg_dump` is stored off-host for at least 14 days and restored monthly. | US-S04 |

## 6. Flow

```text
cron/queue → claim job → execute → record result → retry/alert on failure
```

## 7. Technical notes

Use a jobs table with attempts, lock time, next run, and idempotency key. Expiry correctness stays in `public_offers`; jobs handle archival/cleanup. Worker uses the same Go binary in a separate container.

## 8. Subtasks

- [ ] T7.1 Implement job table, worker loop, locks, retries, and dead letters.
- [ ] T7.2 Implement expiry, cleanup, access reminders, and notification deduplication.
- [ ] T7.3 Implement SMTP templates, delivery logs, and failure alerts.
- [ ] T7.4 Implement sitemap/robots/JSON-LD and encrypted backup/restore drill.

## 9. Success metrics

Job success/retry age; duplicate notifications; backup completion; restore-test success; sitemap freshness.

## 10. Risks & open questions

Confirm SMTP limits, legal retention, and geographically separate backup storage.

---

# PRD-08 · Data & API Foundation

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — data and API                                  |
| **Stories**  | Enabler for US-V01 … US-S04                                      |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/api/internal`, `apps/api/db/migrations`, `apps/api/db/queries`, PostgreSQL 16 |
| **API**      | All `/api/v1` endpoints, `GET /healthz`, `GET /readyz`            |

## 1. Summary

The foundation provides PostgreSQL schema, migrations, transactions, Arabic search, API envelopes, middleware, and the single public visibility rule.

## 2. Problem

Duplicated business rules across UI, handlers, and jobs create inconsistent visibility, access, and payment behavior.

## 3. Goals

- Enforce invariants in database constraints and service transactions.
- Keep handler, service, and repository responsibilities separate.
- Standardize errors, pagination, request IDs, auth, and rate limits.

### Non-goals

Analytics warehouse, multi-region DB, and customer account persistence.

## 4. Users

All product users depend on this layer; developers are its direct users.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Use UUIDs, UTC timestamps, integer DZD amounts, enums, constraints, and forward-only migrations. | `skills/database.md` |
| R2 | Model users, sessions, merchants, documents, shops, offers/media, subscriptions/payments, reports/ads, jobs, notifications, and audit logs. | `skills/database.md` |
| R3 | Implement `public_offers` with status, date, merchant verification, and access conditions. | `userstory.md` |
| R4 | Implement Arabic normalization, search vectors, filters, indexes, and pagination. | `skills/database.md` |
| R5 | Implement JSON envelopes, errors, request IDs, auth/role/CSRF/rate limits, and health/readiness. | `skills/backend.md` |
| R6 | Make trial, subscription, and webhook processing atomic/idempotent. | `skills/database.md`, `skills/payment.md` |

## 6. Flow

```text
request → middleware → handler → service transaction → repository/sqlc → response
```

## 7. Technical notes

Use `pgxpool`, sqlc, migrations, `SELECT FOR UPDATE` for state transitions, indexes on public filters/time/status, and never return private storage keys or session hashes.

## 8. Subtasks

- [ ] T8.1 Add migration runner, extensions, enums, reference data, and seeds.
- [ ] T8.2 Add schema, constraints, triggers, `public_offers`, and sqlc queries.
- [ ] T8.3 Add service/repository conventions, response/errors/pagination/request IDs.
- [ ] T8.4 Add auth, roles, CSRF, rate limits, recovery, secure headers, and trusted proxy checks.
- [ ] T8.5 Add Arabic search, critical transactions, indexes, and integration tests.

## 9. Success metrics

Public-list p95 under 300ms; zero duplicate activation in concurrency tests; clean/populated migration success; consistent error contract.

## 10. Risks & open questions

Confirm wilaya seeds, search ranking quality, and retention policies for high-volume events/stats.

---

# PRD-09 · Web Shell, Brand & Splash

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — web experience and identity                   |
| **Stories**  | Cross-cutting support for US-V01 … US-S04                         |
| **Priority** | High                                                             |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/layout.tsx`, route groups, components, global styles, `site/branding` |
| **API**      | `GET /site/branding` and public APIs used by route groups         |

## 1. Summary

The shell establishes Arabic RTL, Dinar DZ identity, reusable components, loading states, and a database-driven Splash screen.

## 2. Problem

Feature screens can drift in design, RTL behavior, accessibility, loading, and brand expression without a shared shell.

## 3. Goals

- Apply the design system consistently.
- Provide responsive, accessible, Arabic-first navigation.
- Show logo, site name, proverb, and truthful progress in Splash.

### Non-goals

Final logo redesign, complete French content, and native mobile app.

## 4. Users

All guests, merchants, admins, and brand administrators.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Use Dinar DZ name, approved logo, burgundy/green/cream/orange/gold palette. | `skills/ui.md` |
| R2 | Use RTL, logical properties, approved fonts, visible focus, and 44px touch targets. | `skills/ui.md`, `skills/frontend.md` |
| R3 | Build shared header, bottom nav, cards, forms, modals, states, and typography. | `skills/ui.md` |
| R4 | Splash reads active `dinar_dz` branding/proverb and holds three seconds after `page_ready`. | `skills/splash screen.md` |
| R5 | Cover fallback, reduced motion, timeout, error, and no-CLS cases. | `skills/splash screen.md` |
| R6 | SSR public routes and isolate merchant/admin bundles and guards. | `skills/frontend.md` |

## 6. Flow

```text
app boot → branding/page data → page_ready → 3s hold → fade → usable page
```

## 7. Technical notes

Keep Splash fixed and non-blocking to SSR; store branding in `site_branding`; sanitize text and allowlist logo URL; expose progress text and `role="progressbar"`; honor reduced motion.

## 8. Subtasks

- [ ] T9.1 Build root layout, fonts, theme tokens, RTL, metadata, and navigation.
- [ ] T9.2 Build reusable components and visual states.
- [ ] T9.3 Add branding endpoint, cache, fallback, admin update, and audit hook.
- [ ] T9.4 Implement Splash progress, three-second hold, fade, timeout, and accessibility.
- [ ] T9.5 Run responsive/visual QA at 360px, desktop, reduced motion, and 200% text scale.

## 9. Success metrics

CLS under 0.1; axe pass; brand consistency; no Splash-induced LCP regression.

## 10. Risks & open questions

Measure whether the three-second post-ready hold harms bounce; confirm final SVG outline and brand legal availability.

---

# PRD-10 · Delivery, Security & Quality

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — delivery, security, and quality               |
| **Stories**  | Enabler for all P0/P1 stories                                    |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `Dockerfile`, `docker-compose.yml`, `Caddyfile`, `redeploy.sh`, CI, monitoring |
| **API**      | `GET /healthz`, `GET /readyz`; all deployed services             |

## 1. Summary

Docker, Caddy, CI, redeploy, secrets, backups, health checks, observability, and test gates make the specification safely deployable.

## 2. Problem

Features are not production-ready without repeatable builds, secure secrets, rollback, recovery, monitoring, and regression tests.

## 3. Goals

- Build and deploy web, API, worker, database, and proxy reproducibly.
- Protect sessions, documents, payments, secrets, and logs.
- Gate release on unit, integration, E2E, accessibility, security, and health checks.

### Non-goals

Multi-region failover, Kubernetes, and zero-downtime database migrations in V1.

## 4. Users

Developers, operators, admins, and all end users.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Compose runs Caddy, web, API, worker, PostgreSQL, and persistent volumes. | `skills/deployment.md` |
| R2 | Secrets never enter Git, image layers, logs, or client bundles. | deployment/backend/payment |
| R3 | `redeploy.sh` validates Compose, builds, starts, optionally migrates, prints status, and can health-check. | `redeploy.sh` |
| R4 | HTTPS, HSTS, CSP, CSRF, rate limits, private buckets, signed URLs, and webhook HMAC are enabled. | security specs |
| R5 | CI runs unit, integration, Playwright RTL, axe, bundle/performance, migration, and security checks. | frontend/backend/userstory |
| R6 | JSON redacted logs, alerts, encrypted backups, 14-day retention, and monthly restore drill exist. | deployment/userstory |

## 6. Flow

```text
push → CI → versioned image → redeploy.sh → migrate → health checks → release/rollback
```

## 7. Technical notes

Keep `redeploy.sh` in the project root; Caddy is the only public service; use immutable image tags; run migrations forward-only; never log tokens, card data, private documents, or raw payment payloads.

## 8. Subtasks

- [ ] T10.1 Add Dockerfiles, Compose, Caddy routing, health checks, and volumes.
- [ ] T10.2 Add CI lint/test/migration/build/security pipeline and versioned artifacts.
- [ ] T10.3 Validate redeploy, environment, migration, rollback, and smoke-test procedures.
- [ ] T10.4 Add secrets policy, redaction, CSP/HSTS, dependency scanning, and alerts.
- [ ] T10.5 Add encrypted backup/restore, monitoring, incident runbook, and release checklist.

## 9. Success metrics

Repeatable clean-checkout deployment; 99.5% V1 availability; rollback/recovery time; zero leaked secrets/sensitive fields; successful restore drill.

## 10. Risks & open questions

The application/Compose files are not implemented yet. Select hosting, registry, SMTP, storage, monitoring, backup, and alert providers before production.
