# Dinar DZ Project Board

This board is the execution view for the Jira backlog in
[`jira.md`](jira.md). It turns the PRD dependency graph into an
ordered delivery plan while keeping unresolved production decisions visible.

## Board rules

| Column | Meaning |
|---|---|
| Backlog | Accepted scope that is not ready to start |
| Ready | Dependencies and acceptance criteria are clear |
| In Progress | An implementation owner is actively working |
| Code Review | Implementation is complete and awaiting review |
| QA | Functional, security, accessibility, and regression verification |
| Done | Definition of Done is met and the release evidence is linked |
| Blocked | A dependency, legal decision, provider, or incident prevents progress |

## Epic delivery order

| Release | Epic | Scope | Dependency | Current status |
|---|---|---|---|---|
| R0 | EPIC-00 | Program decisions, board, release gate, PRD alignment | — | Done |
| R0 | EPIC-08 | PostgreSQL, SQLBoiler, Redis, API contracts, auth foundation | EPIC-00 | Backlog |
| R0 | EPIC-09 | Next.js shell, Arabic RTL, branding, splash, responsive UI | EPIC-00 | Backlog |
| R0 | EPIC-10 | Docker, Caddy, CI, security, backups, smoke tests | EPIC-08 | Backlog |
| R1 | EPIC-02 | Google identity, merchant onboarding, private documents | EPIC-08, EPIC-09 | Done |
| R2 | EPIC-01 | Guest discovery, search, offer details, reporting, sharing | EPIC-08, EPIC-09 | In Progress |
| R2 | EPIC-03 | Merchant offers, MinIO media, moderation lifecycle | EPIC-02, EPIC-08 | Backlog |
| R2 | EPIC-05 | Admin review, role guards, audit | EPIC-02, EPIC-03, EPIC-08 | Backlog |
| R3 | EPIC-04 | Trial, Chargily/manual payment, access enforcement | EPIC-03, EPIC-08 | Backlog |
| R3 | EPIC-07 | Workers, notifications, expiry, SEO, backup/restore | EPIC-08, EPIC-04 | Backlog |
| R4 | EPIC-06 | Categories, reports, direct ads, settings, dashboards | EPIC-05, EPIC-08 | Backlog |

The complete story and task dependency list remains authoritative in
[`jira.md`](jira.md); this board is the operational ordering and status
view, not a second source of acceptance criteria.

## EPIC-00 completed cards

| Card | Outcome | Evidence | Status |
|---|---|---|---|
| T00.1 | V1 scope, legal assumptions, providers, ownership model, and release blockers recorded | [`decision-log.md`](decision-log.md) | Done |
| T00.2 | Board columns, release order, epic dependencies, and execution rules recorded | This file | Done |
| T00.3 | P0 gate, pilot scope, rollout, rollback, and success metrics recorded | [`release-gate.md`](release-gate.md) | Done |
| T00.4 | Every PRD reviewed against the current decisions and implementation baseline | [`prd/prd-review-matrix.md`](prd/prd-review-matrix.md) | Done |

## EPIC-01 implementation checkpoint

The current guest slice is running locally at `http://127.0.0.1:8180/ar`. It includes the public SSR routes, offer filters, Arabic search normalization, shop/detail pages, share and WhatsApp actions, report protection, and API smoke coverage. The catalog is currently an isolated in-memory adapter so that EPIC-08 can replace its data source with the PostgreSQL `public_offers` view without changing the public contract. EPIC-01 therefore remains In Progress until durable persistence, query-time visibility from PostgreSQL, and Playwright coverage are complete.

Evidence is maintained in `jira.md` under **EPIC-01 execution status**.

## EPIC-02 completion checkpoint

Merchant identity now runs on PostgreSQL and private MinIO storage, with durable refresh-token rotation, database role policies, audited five-minute document links, transactional status-email delivery, ten-second draft autosave, resume routing, and rejected-document replacement. API/web checks and live PostgreSQL/MinIO smoke tests pass against the local Docker stack.

Evidence is maintained in `jira.md` under **EPIC-02 execution status**.

## Next ready queue

The first implementation queue after EPIC-00 is deliberately small:

1. `T8.1` — establish migrations, enums, extensions, reference data, and
   seeds.
2. `T9.1` — establish the web shell, locale routing, theme tokens, and RTL.
3. `T10.1` — establish Docker, Compose, Caddy routing, health checks, and
   persistent volumes.
4. `T8.2` and `T8.3` — add the schema, public visibility rule, repositories,
   errors, pagination, and request IDs.

Each card moves through `Ready → In Progress → Code Review → QA → Done` only
after its listed dependencies and the cross-PRD Definition of Done in
`jira.md` are satisfied.

## Production blockers carried from EPIC-00

These are not hidden by marking the planning epic Done:

- named production release owner and backup;
- hosting region, registry, monitoring, and encrypted off-host backup;
- SMTP/Resend selection and the live Chargily contract;
- official wilaya reference data;
- legal review for privacy, retention, tax, and receipts.

They are tracked as release gates in [`release-gate.md`](release-gate.md).
