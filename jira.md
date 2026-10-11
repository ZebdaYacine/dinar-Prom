# Dinar DZ — Jira Backlog

This file converts [`prd/prd.md`](prd/prd.md), [`userstory.md`](userstory.md), and the documents in [`skills/`](skills/) into a Jira-ready backlog. The IDs below are import references; Jira will assign its own issue keys after import.

## 1. Jira conventions

| Field | Values |
|---|---|
| Issue types | Epic, Story, Task, Bug, Spike |
| Statuses | Backlog → Ready → In Progress → Code Review → QA → Done; Blocked when dependency or decision is missing |
| Priorities | Critical, High, Medium, Low |
| Components | `web`, `api`, `database`, `worker`, `admin`, `billing`, `security`, `infra`, `qa` |
| Labels | `dinar-dz`, `v1`, `rtl`, `p0`, `p1`, `security`, `migration` |
| Definition of Done | Tests pass, authorization verified, no sensitive-data leak, docs updated, observability added, production smoke test passes |

## 2. Epic backlog

| External ID | Epic summary | PRD | Priority | Depends on | Status |
|---|---|---|---|---|---|
| EPIC-00 | Product program and release planning | PRD-00 | Critical | — | Done |
| EPIC-01 | Guest experience | PRD-01 | High | EPIC-08, EPIC-09 | Done |
| EPIC-02 | Merchant identity and verification | PRD-02 | Critical | EPIC-08, EPIC-09 | In Progress |
| EPIC-03 | Merchant offers and media | PRD-03 | Critical | EPIC-02, EPIC-08 | Backlog |
| EPIC-04 | Trial, billing, and payments | PRD-04 | Critical | EPIC-03, EPIC-08 | Backlog |
| EPIC-05 | Admin review and audit | PRD-05 | Critical | EPIC-02, EPIC-03, EPIC-08 | Backlog |
| EPIC-06 | Admin catalog, reports, and ads | PRD-06 | Medium | EPIC-05, EPIC-08 | Backlog |
| EPIC-07 | System lifecycle and background jobs | PRD-07 | Critical | EPIC-08, EPIC-04 | Backlog |
| EPIC-08 | Data and API foundation | PRD-08 | Critical | EPIC-00 | Backlog |
| EPIC-09 | Web shell, brand, and Splash | PRD-09 | High | EPIC-00 | Backlog |
| EPIC-10 | Delivery, security, and quality | PRD-10 | Critical | EPIC-08 | Backlog |

## 2.1 EPIC-00 execution status

| ID | Task | Status | Evidence |
|---|---|---|---|
| T00.1 | Confirm scope, legal assumptions, providers, and release owner | Done | [`decision-log.md`](decision-log.md) |
| T00.2 | Create project board from PRD subtasks and dependencies | Done | [`project-board.md`](project-board.md) |
| T00.3 | Define P0 release gate and pilot launch | Done | [`release-gate.md`](release-gate.md) |
| T00.4 | Keep PRDs synchronized with implementation decisions | Done | [`prd-review-matrix.md`](prd/prd-review-matrix.md) |
| EPIC-00 | Product program and release planning | Done | All four tasks above are complete as planning deliverables |

## 2.2 EPIC-01 execution status

The guest experience is implemented and deployed locally. Public catalogue reads remain deterministic on the current scaffold, while guest view/click/report state is persisted in PostgreSQL through migration `000008_guest_analytics`; the adapter fallback is retained only for isolated unit tests without a database.

| ID | Task | Status | Evidence |
|---|---|---|---|
| T1.1 | Build public routes and guest UI states | Done | SSR home, offers, offer detail, shop pages, responsive cards, icon category rail, nested subcategory chips, 58-wilaya/commune explorer, search controls, empty states, sharing and reporting in `apps/web/src` |
| T1.2 | Implement public queries, filters, Arabic search, and pagination | Done | Public catalogue adapter in `apps/api/internal/guest/catalog.go`; nested category filters, 58-wilaya/1,541-commune catalogue, city endpoint, discount/expiry filters, sorting, Arabic normalization, client pagination, and URL state covered by API/web checks |
| T1.3 | Implement share, WhatsApp, analytics, reports, and rate limits | Done | Web Share/clipboard fallback, phone fallback, hashed visitor deduplication, PostgreSQL event/report persistence through `000008_guest_analytics`, report validation, duplicate protection, and rate limits verified by live smoke checks |
| T1.4 | Add guest SEO and Playwright flow | Done | Canonical/alternate/OG/Twitter metadata, Offer JSON-LD, `sitemap.xml`, `robots.txt`, and `apps/web/e2e/guest-experience.spec.ts`; both browser journeys pass against local Docker |
| EPIC-01 | Guest experience | Done | `go test ./...`, web typecheck/lint/Vitest/build, Playwright (2 passed), PostgreSQL migration verification, and local health/SEO/guest smoke checks pass |

## 2.3 EPIC-02 execution status

The first merchant identity vertical slice is implemented and verified locally. Google OAuth state/PKCE construction, verified-email handling, rotating HttpOnly refresh sessions, CSRF protection, JWT access validation, merchant profile/shop validation, private-document validation, consent, and onboarding routes are present. Production completion still requires wiring the repository to PostgreSQL/MinIO, loading Casbin policies from the database, and delivering status emails through the worker/SMTP module.

| ID | Task | Status | Evidence |
|---|---|---|---|
| T2.1 | Implement Google OAuth, sessions, cookies, logout, and roles | In Progress | `apps/api/internal/auth/{service.go,http.go}`: state + PKCE, verified Google userinfo, RS256 access JWT, HttpOnly refresh cookie rotation, CSRF, logout, merchant role scaffold; unit tests pass |
| T2.2 | Build onboarding wizard, autosave, validation, and consent | In Progress | `apps/web/src/app/[locale]/onboarding/*`, merchant shell, bilingual messages, API validation, consent and pending state; 10-second draft autosave and persistent repository wiring remain |
| T2.3 | Implement private document storage and validation | In Progress | `000009_merchant_identity` migration, private object-key model, JPEG/PNG/WebP/PDF magic-byte checks, SHA-256, replacement, and 8MB limit; MinIO adapter and DB persistence remain |
| T2.4 | Add status emails and onboarding authorization tests | Backlog | Submission status is exposed and unit-covered; SMTP templates, worker delivery logs, Casbin-backed role enforcement, and integration authorization tests remain |
| EPIC-02 | Merchant identity and verification | In Progress | API/web unit and build checks pass; production persistence and email/authorization gates remain open |

## 3. Story backlog

Create each row as a Jira Story under the listed Epic. The acceptance criteria are the minimum definition for the story; the full contract remains in `userstory.md` and the owning PRD.

### EPIC-01 — Guest experience

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-V01 | Show home page with current offers | P0 | Search, categories, latest, ending-soon, and highest-discount sections; hidden/expired offers never appear. |
| US-V02 | Search and filter offers | P0 | Arabic normalization, wilaya/category/discount/expiry filters, sorting, URL state, empty state, pagination. |
| US-V03 | Show offer details | P0 | Gallery, prices, dates, shop, verified badge, similar offers, OG metadata, 404/410 behavior. |
| US-V04 | Contact merchant through WhatsApp | P0 | Prefilled `wa.me` message, phone fallback, deduplicated click metric, no personal tracking. |
| US-V05 | Show shop page | P1 | `/shop/:slug` shows shop data and active public offers only. |
| US-V06 | Report an offer | P0 | Reason and optional message, IP limits, one report per offer/IP, admin review signal. |
| US-V07 | Share an offer | P1 | Web Share API on mobile and clipboard fallback on desktop. |

### EPIC-02 — Merchant identity

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-M01 | Sign in with Google | P0 | OAuth state + PKCE, verified email, secure session, correct first/return routing. |
| US-M02 | Save personal information | P0 | Valid Algerian phone, wilaya/address, autosave, resume later. |
| US-M03 | Save shop information | P0 | Shop/category/location/contact/register fields, unique slug, one pre-verification slug edit. |
| US-M04 | Upload verification documents | P0 | Three required private documents, magic-byte validation, 8MB limit, explicit privacy consent. |
| US-M05 | Track verification status | P0 | Clear status, rejection reason/document, re-upload, status email, draft-only behavior before approval. |

### EPIC-03 — Merchant offers

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-O01 | Create and autosave offer draft | P0 | Required fields, valid price logic, date limit, defaults from shop, 10-second autosave. |
| US-O02 | Upload and process offer images | P0 | 1–10 images, type/size/dimension limits, WebP 400/800/1600, EXIF removed, reorder/delete. |
| US-O03 | Submit offer for review | P0 | Completeness validation, at least one image, `pending` state, edit lock, merchant notification. |
| US-O04 | Edit approved offer safely | P0 | Any material edit hides offer and returns it to `pending` with a warning. |
| US-O05 | Pause, archive, or delete offer | P0 | Drafts may be deleted; approved offers are archived; paused offers are hidden. |
| US-O06 | Duplicate offer | P1 | New draft copies approved data/images and requires new dates before submission. |
| US-O07 | Show merchant statistics | P1 | Views, WhatsApp clicks, 30-day series, dashboard totals, no personal data. |

### EPIC-04 — Trial and billing

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-B01 | Start one trial on first offer approval | P0 | 60 days starts inside approval transaction and never restarts. |
| US-B02 | Show access countdown | P0 | Trial/subscription state, days/date, warning at 14 days or less. |
| US-B03 | Send access reminders | P0 | 14/7/1/0-day notifications, deduplicated by notification key. |
| US-B04 | Pay annual subscription through Chargily | P0 | 4000 DZD hosted Checkout, no card data in Dinar DZ, HMAC webhook only, idempotent activation. |
| US-B05 | Submit manual payment proof | P0 | Proof/reference upload, `under_review`, admin decision, reason on rejection. |
| US-B06 | View payment history | P1 | Status, method, amount, date, reference, and safe receipt metadata. |
| US-B07 | Hide offers after access expiry | P0 | Offers disappear immediately from public query but remain editable in dashboard. |
| US-B08 | Renew without losing access | P0 | New subscription starts at `max(now, access_until)`. |

### EPIC-05 — Admin review

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-A01 | Secure admin login | P0 | Google login plus allowlisted admin email; UI and API role guards. |
| US-A02 | Review merchant verification queue | P0 | Oldest-first queue, detail, signed document URL, approve/reject/suspend/reactivate, audit. |
| US-A03 | Review offers | P0 | Preview, approve/reject/remove, bulk actions, first approval starts trial. |
| US-A04 | Review manual payments | P0 | Proof, merchant, amount, approve/reject, required reason, subscription activation. |
| US-A10 | View audit log | P1 | Append-only sensitive action records with actor, action, time, IP, resource, request ID. |

### EPIC-06 — Commercial operations

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-A05 | Manage categories | P0 | Two levels, add/edit/order/hide, no deletion when referenced. |
| US-A06 | Resolve reports | P1 | Group by offer, show distinct IP count, ignore/hide/suspend actions. |
| US-A07 | Manage direct ads | P1 | Advertiser/media/link/slot/date/priority/targeting and metrics. |
| US-A08 | Show admin dashboard indicators | P1 | Merchant, offer, access, revenue, and top-category metrics. |
| US-A09 | Manage operational settings | P1 | Trial, price, subscription, limits, and manual-payment settings with audit. |
| US-D01 | Display public ads safely | P1 | Fixed slots, reserved dimensions, “Advertisement” label, sponsored link, daily metrics. |

### EPIC-07 — System lifecycle

| ID | Summary | Priority | Acceptance criteria |
|---|---|---|---|
| US-S01 | Expire and archive offers | P0 | Query-time expiry, nightly archive/cleanup, safe media retention. |
| US-S02 | Send system notifications | P0 | Merchant/admin event emails, deduplication, delivery/error logs. |
| US-S03 | Generate SEO assets | P0 | SSR, sitemap, robots, canonical URLs, Offer JSON-LD, public-only data. |
| US-S04 | Back up and restore database | P0 | Daily encrypted off-host dump, 14-day retention, monthly restore test. |

## 4. Platform task backlog

Create each row as a Jira Task under the matching Epic. A task can be split into subtasks by the implementation team without changing the PRD contract.

| ID | Task summary | Epic | Priority | Depends on |
|---|---|---|---|---|
| T00.1 | Confirm scope, legal assumptions, providers, and release owner | EPIC-00 | Critical | — |
| T00.2 | Create project board from PRD dependencies | EPIC-00 | High | T00.1 |
| T00.3 | Define P0 release gate and pilot launch | EPIC-00 | Critical | T00.1 |
| T00.4 | Keep PRDs synchronized with implementation decisions | EPIC-00 | Medium | T00.2 |
| T1.1 | Build public routes and guest UI states | EPIC-01 | High | T8.3, T9.1 |
| T1.2 | Implement public queries, filters, Arabic search, and pagination | EPIC-01 | High | T8.2, T8.5 |
| T1.3 | Implement share, WhatsApp, analytics, reports, and rate limits | EPIC-01 | High | T8.4, T1.1 |
| T1.4 | Add guest SEO and Playwright flow | EPIC-01 | High | T1.1, T1.2 |
| T2.1 | Implement Google OAuth, sessions, cookies, logout, and roles | EPIC-02 | Critical | T8.3, T8.4 |
| T2.2 | Build onboarding wizard, autosave, validation, and consent | EPIC-02 | Critical | T2.1, T9.1 |
| T2.3 | Implement private document storage and validation | EPIC-02 | Critical | T8.2, T2.1 |
| T2.4 | Add status emails and onboarding authorization tests | EPIC-02 | High | T2.2, T2.3, T7.3 |
| T3.1 | Implement offer state machine, CRUD, ownership, and validation | EPIC-03 | Critical | T2.2, T8.2 |
| T3.2 | Build editor, media uploader, ordering, cover, and rejection UI | EPIC-03 | Critical | T3.1, T9.2 |
| T3.3 | Implement image processing, cleanup, duplicate, archive, stats | EPIC-03 | High | T3.1, T7.1 |
| T3.4 | Add merchant/moderation E2E flow | EPIC-03 | High | T3.1, T5.3 |
| T4.1 | Implement access, trial, and subscription transactions | EPIC-04 | Critical | T3.1, T8.5 |
| T4.2 | Build billing UI, history, proof, success/failure, pending states | EPIC-04 | Critical | T9.2, T4.1 |
| T4.3 | Integrate Chargily Checkout and signed webhooks | EPIC-04 | Critical | T4.1, T8.5 |
| T4.4 | Implement manual review, reminders, expiry, receipts, reconciliation | EPIC-04 | Critical | T4.1, T4.3, T7.2 |
| T5.1 | Build admin shell, guards, metrics, queues, filters, pagination | EPIC-05 | Critical | T2.1, T9.1 |
| T5.2 | Build merchant/document review and signed viewer | EPIC-05 | Critical | T2.3, T5.1 |
| T5.3 | Build offer/payment review and bulk operations | EPIC-05 | Critical | T3.1, T4.1, T5.1 |
| T5.4 | Build audit queries, notifications, and authorization tests | EPIC-05 | High | T5.2, T5.3, T7.3 |
| T6.1 | Build category CRUD, hierarchy, ordering, visibility, and seeds | EPIC-06 | High | T8.2, T5.1 |
| T6.2 | Build report aggregation, resolution actions, and side effects | EPIC-06 | High | T1.3, T5.3 |
| T6.3 | Build ad CRUD, media, targeting, slots, and metrics | EPIC-06 | Medium | T6.1, T9.2 |
| T6.4 | Build dashboard/settings UI and API | EPIC-06 | Medium | T5.1, T6.1 |
| T7.1 | Implement job table, worker loop, locks, retries, and dead letters | EPIC-07 | Critical | T8.2 |
| T7.2 | Implement expiry, cleanup, reminders, and notification deduplication | EPIC-07 | Critical | T7.1, T4.1 |
| T7.3 | Implement SMTP templates, delivery logs, and alerts | EPIC-07 | High | T7.1 |
| T7.4 | Implement sitemap, robots, JSON-LD, backup, and restore drill | EPIC-07 | Critical | T7.1, T8.5 |
| T8.1 | Add migration runner, extensions, enums, reference data, and seeds | EPIC-08 | Critical | T00.1 |
| T8.2 | Add schema, constraints, triggers, public view, and sqlc queries | EPIC-08 | Critical | T8.1 |
| T8.3 | Add service/repository conventions, response/errors/pagination/request IDs | EPIC-08 | Critical | T8.2 |
| T8.4 | Add auth, CSRF, rate limits, recovery, headers, and proxy checks | EPIC-08 | Critical | T8.3 |
| T8.5 | Add Arabic search, critical transactions, indexes, and integration tests | EPIC-08 | Critical | T8.2, T8.3 |
| T9.1 | Build root layout, fonts, theme tokens, RTL, metadata, navigation | EPIC-09 | High | T00.1 |
| T9.2 | Build reusable design components and visual states | EPIC-09 | High | T9.1 |
| T9.3 | Add branding endpoint, cache, fallback, admin update, audit hook | EPIC-09 | High | T8.3, T5.4 |
| T9.4 | Implement Splash progress, three-second hold, fade, timeout, accessibility | EPIC-09 | High | T9.1, T9.3 |
| T9.5 | Run responsive and visual QA at 360px, desktop, reduced motion, 200% text | EPIC-09 | High | T9.2, T9.4 |
| T10.1 | Add Dockerfiles, Compose, Caddy routing, health checks, and volumes | EPIC-10 | Critical | T8.1 |
| T10.2 | Add CI lint/test/migration/build/security pipeline and artifacts | EPIC-10 | Critical | T8.3 |
| T10.3 | Validate redeploy, environment, migration, rollback, and smoke tests | EPIC-10 | Critical | T10.1, T10.2 |
| T10.4 | Add secrets policy, redaction, CSP/HSTS, dependency scanning, alerts | EPIC-10 | Critical | T8.4, T10.1 |
| T10.5 | Add encrypted backup/restore, monitoring, incident runbook, release checklist | EPIC-10 | Critical | T7.4, T10.3 |

## 5. Suggested Jira releases

| Release | Scope | Exit criteria |
|---|---|---|
| R0 — Foundation | EPIC-00, EPIC-08, EPIC-09, EPIC-10 foundation tasks | Migrations, API contract, shell, CI, health checks, Docker baseline |
| R1 — Accounts | EPIC-02 plus admin authentication | Merchant can submit verification; private docs and audit pass |
| R2 — Offers | EPIC-01, EPIC-03, EPIC-05 review | Guest can discover; merchant can submit; admin can approve; public visibility works |
| R3 — Revenue | EPIC-04, EPIC-07 notifications | Trial/payment/manual review/expiry are idempotent and observable |
| R4 — Operations | EPIC-06, remaining EPIC-07/10 | Categories, ads, reports, backups, SEO, monitoring, and restore drill |

## 6. Jira import notes

- Import Epics first, then Stories, then Tasks so Epic Link values can resolve.
- Map `External ID` to a custom field such as `External Reference`.
- Convert `P0` to Highest/Critical, `P1` to High, and `P2` to Medium if the Jira instance has no custom priority names.
- Keep `Depends on` as issue links, not plain text, after import.
- Use `dinar-dz` on every issue and add component labels from the conventions table.
- For implementation stories and platform tasks, do not mark an issue `Done`
  from documentation alone; require the cross-PRD Definition of Done in
  `prd/prd.md`. `EPIC-00` is the exception because its scope is program planning
  and its completion evidence is the decision log, board, release gate, and
  PRD review matrix.
