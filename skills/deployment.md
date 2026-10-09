# Dinar DZ — Docker Deployment Plan

This document describes production deployment and redeployment after code
changes. Use `docker-compose.dev.yml` for development and
`docker-compose.prod.yml` for production.

## 1. Architecture

```text
Internet -> Caddy HTTPS -> web (Next.js)
                      \-> api (Go Fiber) -> postgres
                                           -> Redis
                                           -> MinIO S3
worker (Go) -> postgres + SMTP + jobs
```

| Service | Responsibility | Notes |
|---|---|---|
| `caddy` | HTTPS and routing | Only public service |
| `web` | Next.js SSR and UI | Never connects directly to PostgreSQL |
| `api` | Go Fiber and REST API | Only owner of business logic |
| `worker` | Expiration, email, cleanup | May use the API image |
| `postgres` | PostgreSQL 16 | Persistent volume and backups |
| `minio` | MinIO S3 | Public images and private documents |

### Development ports

These ports avoid the existing Aisha, Hodhod, 50lab, and Northstar containers:

| Service | Host port | Container port |
|---|---:|---:|
| Caddy HTTP | `8180` | `80` |
| Caddy HTTPS | `8443` | `443` |
| Next.js web | `3143` | `3000` |
| Go API | `8188` | `8080` |
| PostgreSQL | `5543` | `5432` |
| Redis | `6381` | `6379` |
| MinIO API | `9100` | `9000` |
| MinIO Console | `9101` | `9001` |

Inside Docker, use `postgres:5432`, `redis:6379`, and `minio:9000`. Do not expose
PostgreSQL, Redis, or the MinIO Console publicly.

## 2. Requirements

- Linux server with enough memory for Next.js, Go, PostgreSQL, Redis, and MinIO.
- Docker Engine and Docker Compose v2.
- DNS records for `dinar.dz` and `media.dinar.dz` in production.
- Protected `.env.production` with mode `600` and JWT PEM files outside Git.
- SMTP and MinIO credentials.
- Firewall allowing only ports 80, 443, and restricted SSH.

Validate the files before starting:

```bash
docker compose --env-file .env -f docker-compose.dev.yml config --quiet
docker compose --env-file .env.production -f docker-compose.prod.yml config --quiet
```

## 3. Shared environment

The root `.env` file is the shared configuration source for the stack. Create it
from [`.env.example`](/home/zedyacine/Desktop/dinar/.env.example). Compose uses
`SHARED_ENV_FILE` to pass the same file to the API, worker, and MinIO initializer.
Use a secret manager or protected `.env.production` in production; never commit
real secrets.

| Use | Inside Docker | On the host |
|---|---|---|
| API | `http://api:8080` | `http://localhost:8188` |
| PostgreSQL | `postgres:5432` | `localhost:5543` |
| Redis | `redis:6379` | `localhost:6381` |
| MinIO | `minio:9000` | `localhost:9100` |

Do not pass backend secrets to the web container. Only `NEXT_PUBLIC_*` values may
reach the browser. MinIO uses `MINIO_USE_SSL=false` only because it is private
inside Docker and public HTTPS terminates at Caddy. Enable MinIO TLS if it moves
outside that private network.

## 4. First build and startup

1. Copy the project and protected environment file to the server.
2. Generate JWT keys as described in `secrets/README.md`.
3. Verify persistent PostgreSQL, MinIO, and Caddy volumes plus `minio-init`.
4. Build and start production services:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml build --pull
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --remove-orphans
```

5. Run migrations before public traffic:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml run --rm --entrypoint /app/dinar-api api migrate up
```

6. Check readiness:

```bash
curl --fail https://dinar.dz/healthz
curl --fail https://dinar.dz/readyz
docker compose --env-file .env.production -f docker-compose.prod.yml ps
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=100 api web worker caddy
```

## 5. Dockerfile and Caddy rules

- Build the Go API in a separate stage and run the runtime image as non-root.
- Use one API image with separate API and worker commands.
- Build Next.js with `output: "standalone"`; never place backend secrets in `NEXT_PUBLIC_*`.
- Caddy is the only public service and routes `/api/*` to `api:8080` and other paths to `web:3000`.
- Caddy `reverse_proxy` automatically forwards WebSocket upgrades for `/api/v1/ws/notifications`.
- Keep Caddy certificates and configuration on persistent volumes.

## 6. Redeployment

`redeploy.sh` rebuilds images, starts services, and shows their status:

```bash
chmod +x redeploy.sh
./redeploy.sh
```

Run migrations explicitly after a build:

```bash
MIGRATION_SERVICE=api MIGRATION_COMMAND='/app/dinar-api migrate up' ./redeploy.sh
```

Run `PULL_SOURCE=1 ./redeploy.sh` only when the server is intentionally managed
through Git. Never expose Docker or a redeploy endpoint publicly.

For local development, the shorter entry point is:

```bash
chmod +x script.sh
./script.sh
```

It validates Compose, rebuilds the API/web/worker images, starts the stack, and
waits for the API and web URLs. For the production-style Compose file, run:

```bash
COMPOSE_FILE=docker-compose.prod.yml ENV_FILE=.env.production ./script.sh
```

Set `BUILD_PULL=1` when base images must be refreshed.

## 7. Backups, rollback, and monitoring

- Never delete the PostgreSQL volume during redeployment.
- Create encrypted daily `pg_dump` backups outside the server and retain at least 14 days.
- Test restoration monthly in a separate PostgreSQL instance.
- Use immutable image tags such as `api:2026-10-09-abc123`; never rely on `latest` for rollback.
- Roll back application images only; never automatically reverse a database migration.
- Monitor 5xx responses, API latency, disk usage, PostgreSQL health, webhook failures, email failures, and WebSocket disconnects.
- Never log cookies, tokens, document numbers, or file contents.

## 8. Production checklist

- [ ] `docker compose config --quiet` succeeds.
- [ ] Secrets are in a secret manager or protected environment file, not Git.
- [ ] JWT rotation, refresh reuse detection, CSRF, CORS, and CSP are enabled.
- [ ] `healthz` and `readyz` succeed.
- [ ] HTTPS works and HTTP redirects.
- [ ] OAuth callback and Chargily webhook use the correct domain.
- [ ] The document bucket is private and has no public ACL.
- [ ] Backup restoration has been tested.
- [ ] Visitor, merchant, payment, and WebSocket notification smoke tests pass.
- [ ] A rollback method and previous release are documented.
