# PRD Review Matrix — T00.4

All execution PRDs were reviewed after the current provider, security,
deployment, payment, brand, and localization decisions were recorded in the
project documentation. The review confirms that the source specifications do
not contradict the baseline below.

## Cross-PRD decisions applied

- PostgreSQL 16 is the durable source of truth; SQLBoiler owns generated Go
  models; Redis is used for cache, limits, counters, locks, and Pub/Sub.
- MinIO S3 separates public media from private merchant documents and payment
  proofs; private objects require short-lived signed URLs and audit events.
- Google OAuth plus JWT is used for authentication. Access tokens stay in
  memory and refresh tokens are rotated in secure HttpOnly cookies; browser
  local storage is not used for tokens.
- Casbin is deny-by-default for admin and merchant authorization.
- Chargily hosted Checkout receives card data; Dinar DZ stores only safe
  transaction metadata, verifies webhook signatures, and applies idempotency.
- Next.js uses `next-intl` with Arabic and French locale routing, RTL support,
  light/dark themes, responsive mobile/desktop layouts, and the documented
  splash behavior.
- Docker Compose, Caddy, shared environment configuration, health checks,
  WebSocket notifications, CI, security headers, backups, and smoke tests are
  delivery requirements.

## Review results

| PRD | Reviewed focus | Alignment result | Follow-up |
|---|---|---|---|
| PRD-00 | Scope, ownership, dependencies, P0 gate | Aligned | Open production choices remain release gates |
| PRD-01 | Guest routes, public visibility, search, SEO, reports | Aligned | Implement against the public-offer query only |
| PRD-02 | OAuth, JWT/session boundary, onboarding, private documents | Aligned and implemented | Keep SMTP credentials and legal retention policy as production release gates |
| PRD-03 | Offer state machine, media processing, MinIO boundaries | Aligned | Enforce server-side validation and moderation transitions |
| PRD-04 | Trial, 4,000 DZD subscription, Chargily, manual proof | Aligned | Confirm live Chargily contract before production |
| PRD-05 | Casbin admin roles, queues, signed viewers, audit | Aligned | Keep deny-by-default policies and audit sensitive reads |
| PRD-06 | Categories, reports, direct ads, metrics, settings | Aligned | Keep ad media/public content separate from private objects |
| PRD-07 | Workers, Redis locks, WebSocket/email notifications, backups, SEO | Aligned | Complete provider and restore-owner decisions |
| PRD-08 | PostgreSQL, SQLBoiler, Redis, API conventions, transactions | Aligned | Generate models from migrations; do not hand-edit them |
| PRD-09 | `next-intl`, ar/fr flags, themes, icons, splash, responsive UI | Aligned | Validate RTL, reduced motion, 360px, and desktop layouts |
| PRD-10 | Docker, Caddy, CI, JWT security, secrets, observability, QA | Aligned | Run staging smoke, rollback, security, and restore gates |

## Completion decision

T00.4 is complete: every PRD has a recorded alignment result and a concrete
follow-up where implementation or production decisions are still required.
Future implementation changes must update this matrix and the affected PRD in
the same change set.
