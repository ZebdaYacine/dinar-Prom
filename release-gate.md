# Dinar DZ P0 Release Gate

This is the release contract for the first Dinar DZ pilot. It is owned by the
product owner, executed by the engineering lead, and approved by the
production operator. The pilot is one Algerian city with 20–30 manually
onboarded merchants.

## P0 exit criteria

| Gate | Required evidence | Related work | Owner |
|---|---|---|---|
| Scope and legal | Decision log is current; privacy, retention, tax, and receipt decisions are reviewed locally | T00.1, T00.3 | Product owner / legal reviewer |
| Foundation | PostgreSQL migrations, SQLBoiler models, Redis, API errors, request IDs, health checks, and seeds pass in staging | T8.1–T8.5 | Engineering |
| Identity and access | Google OAuth, JWT access/rotated refresh handling, Casbin policy tests, secure cookies, and logout pass | T2.1, T8.4 | Engineering / security |
| Private data | Merchant documents and payment proofs are private MinIO objects; signed URLs expire; access is audited; no card data is stored | T2.3, T4.3, T5.2 | Engineering / security |
| Guest journey | Search, filters, offer detail, shop, WhatsApp, report, share, SEO, 404/410, mobile and desktop flows pass | T1.1–T1.4, T9.5 | Product / QA |
| Offer trust | Merchant submission, moderation, public visibility, expiry, pause, archive, and material-edit rules pass | T3.1–T3.4, T5.3, T7.2 | Engineering / operations |
| Revenue | 60-day first-approval trial, 4,000 DZD subscription, Chargily signed/idempotent webhook, manual proof workflow, expiry, and reconciliation pass | T4.1–T4.4 | Billing / engineering |
| Operations | Admin role guards, audit trail, notifications, worker retries/dead letters, backup, restore drill, alerts, and incident runbook pass | T5.1–T5.4, T7.1–T7.4, T10.5 | Operations |
| Delivery | Docker Compose staging/prod, Caddy HTTPS, secrets policy, CI, dependency scan, migration/rollback rehearsal, and smoke tests pass | T10.1–T10.5 | Engineering / operations |

## Pilot launch checklist

Before inviting merchants:

- Record the named release owner, backup owner, hosting region, registry,
  monitoring route, backup destination, email provider, live Chargily details,
  official wilaya data, and tax/receipt decision in
  [`decision-log.md`](decision-log.md).
- Run migrations against a fresh staging database and verify the public offer
  query cannot return hidden, paused, expired, or unapproved offers.
- Create test merchant, admin, and guest accounts; exercise the full happy path
  and all payment failure/replay paths.
- Verify that access and refresh tokens are not placed in browser local storage,
  and that logs redact tokens, document URLs, payment proofs, and provider
  secrets.
- Run the backup, restore, rollback, and incident exercises and record their
  timestamps and owners.
- Complete responsive checks at 360px and desktop widths, Arabic RTL checks,
  keyboard navigation, reduced-motion behavior, and loading/error/empty states.

## Rollout

1. Deploy the tagged release to staging and run migrations plus smoke tests.
2. Invite five internal/test merchants and observe onboarding, approval,
   offer publication, notifications, and payment reconciliation.
3. Expand to 20–30 merchants in one city after the pilot checklist is signed.
4. Review daily errors, failed jobs, payment mismatches, document-access audit
   events, offer reports, and support tickets for the first seven days.

## Rollback and stop conditions

Stop the pilot and disable new submissions or payments if any of the following
occurs: unauthorized private-object access, card data enters Dinar DZ systems,
duplicate payment activation, broken public visibility enforcement, migration
data loss, unrecoverable backup failure, or a Critical security finding.

Rollback uses the previous application image and the documented database
rollback/forward-fix procedure. Never delete production data to roll back.

## Pilot success metrics

- 20–30 merchants onboarded with at least 90% completing required steps.
- At least 80% of submitted offers reach an admin decision within one business
  day.
- Zero unauthorized document reads, card-data storage events, or duplicate
  subscription activations.
- Restore drill completes within the agreed recovery objective.
- Guest pages remain usable on mobile and achieve the PRD-01 performance target.
