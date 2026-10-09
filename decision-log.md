# Dinar DZ Decision Log

This decision log records project-level decisions that must remain aligned with
the PRDs, user stories, and implementation skills.

## T00.1 — Scope, legal assumptions, providers, and release ownership

**Status:** Confirmed for planning; production launch remains gated by the open
items listed below.

**Reference:** `EPIC-00 / T00.1`, [PRD-00](prd/PRD-00-Program-Overview.md),
[userstory.md](userstory.md), and [docs.md](docs.md).

### Confirmed V1 scope

- Guests browse, search, filter, share, report, and contact merchants through
  WhatsApp or phone without a customer account.
- Merchants authenticate with Google, complete onboarding, upload private
  verification documents, create moderated offers, and manage access status.
- Admins review merchants, documents, offers, payments, reports, categories,
  ads, settings, and audit records.
- Revenue is an annual merchant subscription of 4,000 DZD after one 60-day
  trial, plus direct-sold advertising.
- V1 does not include customer accounts, product checkout, stored cards, video,
  a native mobile app, or automated ad networks.
- The pilot target is one city with 20–30 manually onboarded merchants.

### Confirmed technical providers

| Capability | V1 decision | Boundary |
|---|---|---|
| Database | PostgreSQL 16 | Durable source of truth; UTC `timestamptz` |
| ORM/model generation | SQLBoiler | Generated models are never edited manually |
| Cache/coordination | Redis 7 | Cache, rate limits, counters, short locks, Pub/Sub only |
| Object storage | MinIO S3 | Public media bucket and private document/payment bucket |
| Reverse proxy | Caddy | HTTPS termination and `/api`/web routing |
| Backend | Go + Fiber | Only owner of business logic |
| Frontend | Next.js + TypeScript | SSR public pages; no direct database access |
| Authorization | Casbin | Deny-by-default RBAC and route/action policies |
| Authentication | Google OAuth + JWT | Memory-only access JWT and rotated HttpOnly refresh cookie |
| Payments | Chargily Pay v2 hosted Checkout | No card data enters or is stored by Dinar DZ |
| Email | SMTP or Resend adapter | Provider remains selectable through configuration |
| Runtime | Docker Compose | Development and single-host production baseline |

### Legal and compliance assumptions

- Merchant identity documents and payment proofs are personal or sensitive
  data and remain private in MinIO.
- The platform does not process or store card numbers, CVV, PIN, expiry data,
  or full hosted-checkout credentials.
- Explicit consent, purpose, retention, access audit, and deletion procedures
  are required for merchant documents.
- Algerian personal-data obligations, retention limits, required declarations
  or authorizations, tax treatment, and invoice requirements must be reviewed
  by qualified local counsel before production launch.
- The legal review is a release gate, not a substitute for technical security
  controls.

### Release ownership

- **Release owner role:** Dinar DZ product owner.
- **Implementation owner:** engineering lead assigned in Jira.
- **Operational approver:** production operator assigned before deployment.
- A named person and contact channel are still required in Jira before the R0
  release can be marked Ready. The repository does not invent a person's name.

### Open decisions blocking production

| Decision | Current state | Required action |
|---|---|---|
| Named release owner | Not recorded | Assign a named Jira owner and backup |
| Hosting provider/region | Not selected | Choose host, region, and data-residency position |
| Container registry | Not selected | Choose registry and retention policy |
| Monitoring/alerting | Not selected | Choose monitoring/logging provider and on-call route |
| Backup destination | Not selected | Choose encrypted off-host storage and restore owner |
| SMTP vs Resend | Open | Select provider and confirm sending limits |
| Chargily live contract | Test base URL configured | Verify live endpoint, events, signature header, and merchant account |
| Wilayas reference data | Open | Confirm official dataset and final count |
| Tax and receipts | Open | Obtain legal/accounting decision |

### Release gate

The scope, legal assumptions, provider baseline, and ownership model are now
recorded for planning, so T00.1 is complete as a program-management task.
The named production owner and the provider/legal choices below remain explicit
R0 release gates and must be resolved before production deployment.
