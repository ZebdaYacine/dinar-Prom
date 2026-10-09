# Delivery, Security & Quality

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — delivery, security, and quality               |
| **Stories**  | Enabler for all P0/P1 stories                                    |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `Dockerfile`, `docker-compose.yml`, `Caddyfile`, `redeploy.sh`, CI, monitoring |
| **API**      | `GET /healthz`, `GET /readyz`; all deployed services          |

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
| R4 | HTTPS, HSTS, CSP, CSRF, rate limits, private buckets, signed URLs, and webhook HMAC are enabled. | security specifications |
| R5 | CI runs unit, integration, Playwright RTL, axe, bundle/performance, migration, and security checks. | frontend/backend/userstory |
| R6 | Redacted JSON logs, alerts, encrypted backups, 14-day retention, and monthly restore drill exist. | deployment/userstory |

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

Repeatable clean-checkout deployment; 99.5% V1 availability; recovery time; zero leaked secrets/sensitive fields; successful restore drill.

## 10. Risks & open questions

The application and Compose files are not implemented yet. Select hosting, registry, SMTP, storage, monitoring, backup, and alert providers before production.
