# Data & API Foundation

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — data and API                                  |
| **Stories**  | Enabler for US-V01 … US-S04                                      |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/api/internal`, `apps/api/db/migrations`, `apps/api/db/queries`, PostgreSQL 16 |
| **API**      | All `/api/v1` endpoints, `GET /healthz`, `GET /readyz`      |

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

Confirm wilaya seeds, search ranking quality, and retention policies for high-volume events and stats.
