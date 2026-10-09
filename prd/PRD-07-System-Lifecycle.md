# System Lifecycle

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
| R3 | Sitemap, robots, canonical URLs, and Offer JSON-LD come from public data. | US-S03 |
| R4 | Daily encrypted `pg_dump` is stored off-host for at least 14 days and restored monthly. | US-S04 |

## 6. Flow

```text
cron/queue → claim job → execute → record result → retry/alert on failure
```

## 7. Technical notes

Use a jobs table with attempts, lock time, next run, and idempotency key. Expiry correctness stays in `public_offers`; jobs handle archival and cleanup. Worker uses the same Go binary in a separate container.

## 8. Subtasks

- [ ] T7.1 Implement job table, worker loop, locks, retries, and dead letters.
- [ ] T7.2 Implement expiry, cleanup, access reminders, and notification deduplication.
- [ ] T7.3 Implement SMTP templates, delivery logs, and failure alerts.
- [ ] T7.4 Implement sitemap/robots/JSON-LD and encrypted backup/restore drill.

## 9. Success metrics

Job success/retry age; duplicate notifications; backup completion; restore-test success; sitemap freshness.

## 10. Risks & open questions

Confirm SMTP limits, legal retention, and geographically separate backup storage.
