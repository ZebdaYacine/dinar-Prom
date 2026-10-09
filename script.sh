#!/usr/bin/env bash
set -Eeuo pipefail

# Rebuild and redeploy the Dinar DZ stack.
# Local development: ./script.sh
# Production-style: COMPOSE_FILE=docker-compose.prod.yml ENV_FILE=.env.production ./script.sh
# Optional: BUILD_PULL=1, HEALTHCHECK_URL=..., WEB_URL=...

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
ENV_FILE="${ENV_FILE:-.env.example}"
BUILD_PULL="${BUILD_PULL:-0}"

if [[ "$COMPOSE_FILE" == "docker-compose.prod.yml" ]]; then
  HEALTHCHECK_URL="${HEALTHCHECK_URL:-http://127.0.0.1/healthz}"
  WEB_URL="${WEB_URL:-http://127.0.0.1/}"
else
  HEALTHCHECK_URL="${HEALTHCHECK_URL:-http://127.0.0.1:8188/healthz}"
  WEB_URL="${WEB_URL:-http://127.0.0.1:8180/ar}"
fi

command -v docker >/dev/null 2>&1 || { echo 'Error: Docker is required.' >&2; exit 1; }
docker info >/dev/null 2>&1 || { echo 'Error: Docker daemon is not available.' >&2; exit 1; }
[[ -f "$COMPOSE_FILE" ]] || { printf 'Error: Compose file not found: %s\n' "$COMPOSE_FILE" >&2; exit 1; }
[[ -f "$ENV_FILE" ]] || { printf 'Error: environment file not found: %s\n' "$ENV_FILE" >&2; exit 1; }
command -v curl >/dev/null 2>&1 || { echo 'Error: curl is required.' >&2; exit 1; }

compose() {
  SHARED_ENV_FILE="$ENV_FILE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

compose config --quiet
if [[ "$BUILD_PULL" == "1" ]]; then
  compose build --pull
else
  compose build
fi
compose up -d --remove-orphans
compose ps

wait_for_url() {
  local url="$1"
  local attempt
  for ((attempt = 1; attempt <= 30; attempt++)); do
    if curl --fail --silent --show-error --max-time 5 "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  printf 'Error: health check failed: %s\n' "$url" >&2
  compose logs --tail=80 api web worker caddy >&2 || true
  return 1
}

wait_for_url "$HEALTHCHECK_URL"
wait_for_url "$WEB_URL"

printf '\nDeployment completed successfully using %s.\n' "$COMPOSE_FILE"
printf 'Web: %s\n' "$WEB_URL"
printf 'API health: %s\n' "$HEALTHCHECK_URL"

if [[ "$COMPOSE_FILE" == "docker-compose.dev.yml" ]]; then
  printf 'Direct Next.js: http://127.0.0.1:3143/ar\n'
  printf 'API through Caddy: http://127.0.0.1:8180/api/v1\n'
  printf 'MinIO console: http://127.0.0.1:9101\n'
fi
