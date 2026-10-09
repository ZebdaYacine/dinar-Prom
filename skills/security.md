# Dinar DZ — Security and JWT Policy

This file is the shared security reference for the backend, frontend, and Docker
stack. The goal is revocable and rotating authentication without giving browser
JavaScript a permanent place to store access or refresh tokens.

## 1. Token model

| Token | Location | Lifetime | Use |
|---|---|---:|---|
| Access JWT | Application memory only (`React context`/module memory) | 10 minutes | `Authorization: Bearer` for API requests |
| Refresh JWT | `HttpOnly; Secure; SameSite=Strict` cookie | 30 days with rotation | Issue a new access token only |
| CSRF token | Non-sensitive readable cookie plus header | Browser session | Protect cookie-authenticated endpoints |
| WebSocket ticket | One-use `HttpOnly` cookie | 30 seconds | Open WebSocket without putting a JWT in the URL |

Never store access or refresh tokens in `localStorage`, `sessionStorage`,
IndexedDB, query strings, URL fragments, or logs. After a page reload, the
frontend requests a new access token from `/api/v1/auth/refresh` using the
refresh cookie; it does not restore a token from persistent storage.

The recommended design uses RSA-256 or EdDSA with asymmetric keys. Do not use
HS256 across multiple services unless the secret is intentionally shared through
a protected secret manager. Production JWT keys must never be stored in an image
or Git.

## 2. Claims and validation

Every access JWT must contain:

```json
{
  "iss": "dinar-api",
  "aud": "dinar-web",
  "sub": "user-uuid",
  "jti": "token-uuid",
  "iat": 1730000000,
  "nbf": 1730000000,
  "exp": 1730000600,
  "token_use": "access"
}
```

Backend validation is mandatory:

- Accept exactly the algorithm configured by the server; reject `none` and every unexpected algorithm.
- Validate the signature, `iss`, `aud`, UUID-shaped `sub`, `exp`, `nbf`, `iat`, and `jti`.
- Allow no more than 30 seconds of clock skew and reject expired access tokens immediately.
- Never use a client-provided `role` or `is_admin` claim for authorization. Load roles and policies from PostgreSQL/Casbin; `sub` identifies the user only.
- Use `kid` for key rotation and retain the previous key for at least the maximum access-token lifetime.
- Limit the Authorization header and JSON/input sizes before parsing. Return a generic `401 unauthorized` without revealing signature details or whether a user exists.
- Never log JWTs, Authorization headers, refresh cookies, or sensitive request bodies.

## 3. Login and refresh flow

1. Google OAuth uses `state`, PKCE, and an HTTPS callback with an exact redirect URI.
2. After OAuth succeeds, the API creates a random refresh token with a `jti`, stores only its hash in PostgreSQL, and sends the refresh JWT in the `__Host-dinar-refresh` cookie in production.
3. The API creates a CSRF cookie and redirects to the frontend. The frontend reads only the CSRF value and sends it in `X-CSRF-Token` for `refresh`, `logout`, and other cookie-sensitive operations.
4. The frontend calls `POST /api/v1/auth/refresh` with `credentials: include`. The API returns an access JWT as JSON, and the frontend keeps it in memory only.
5. Protected requests send `Authorization: Bearer <access>`. After a `401`, the client performs one refresh through a mutex and retries the request once; it never enters a refresh loop.
6. The refresh endpoint rotates the refresh token on every use and revokes the old one. Reuse of an old refresh token revokes the entire token family and requires login again.
7. Logout revokes the `jti`/family in PostgreSQL and clears cookies; clearing only the frontend memory variable is insufficient.

## 4. Blocking untrusted communication

The server cannot prove that a request came from frontend code using `User-Agent`
or a custom header; any client can copy those values. Production enforcement must
use layered controls:

- Do not expose the API, PostgreSQL, Redis, or MinIO directly. Caddy is the only public entrypoint, and the production API has no host port in Compose.
- Allow only known `Host` and `Origin` values (`https://dinar.dz` in production); reject unknown origins for WebSocket, cookie requests, and unsafe operations.
- Enable CORS with an exact origin allowlist; never use `*` with credentials.
- Require Bearer authorization for protected APIs and a CSRF header for refresh/logout; never accept an access token in a cookie or URL.
- Keep service-to-service traffic on the private Docker network. Use mTLS or another private network when a service leaves Docker.
- Apply separate rate limits to login, refresh, upload, admin, webhook, and WebSocket operations.
- Add `Content-Security-Policy` that restricts script and connection origins, with `frame-ancestors 'none'`, HSTS, `nosniff`, and Referrer-Policy.
- Validate method, content type, body size, and request ID before the handler, and reject unencrypted HTTP in production through Caddy.

## 5. Refresh-token storage

PostgreSQL stores only the fields required for refresh revocation:

| Field | Rule |
|---|---|
| `token_hash` | SHA-256 or HMAC hash, UNIQUE; never the raw token |
| `user_id` | FK to the user |
| `family_id` | Identify reuse and revoke a family |
| `jti` | UNIQUE, for auditing and revocation |
| `expires_at` | Mandatory expiry |
| `revoked_at` | Nullable, with a revocation reason |
| `replaced_by_hash` | Link rotation without storing the new token |
| `user_agent`, `ip` | Restricted metadata, redacted in logs |

Redis is not the authentication source of truth. It may provide a short refresh
lock or rate limit, but Redis loss must never authorize a request or reuse a
refresh token.

## 6. WebSocket notifications

Because the browser WebSocket API does not provide a normal `Authorization`
header, the frontend uses this sequence:

1. Send `POST /api/v1/auth/ws-ticket` with the in-memory access Bearer token.
2. The API validates the JWT, stores a one-use ticket hash in Redis with a 30-second TTL, and sends the ticket in a `HttpOnly; Secure; SameSite=Strict` cookie with a restricted path.
3. Open `wss://dinar.dz/api/v1/ws/notifications` without a token in the URL.
4. The backend consumes the ticket once during the handshake, validates Origin and user ID, and subscribes to the user's Redis Pub/Sub notifications.

The WebSocket does not change read state and is not a durable data source. After
reconnect, the frontend requests `/me/notifications?since_id=...` through REST.
Every notification is stored in PostgreSQL before it is published to Redis.

## 7. Input validation

- Use Zod in the frontend for user feedback only; Go validator/schema is authoritative.
- Reject unknown JSON fields on sensitive endpoints or apply an explicit field allowlist.
- Apply limits to strings, arrays, pagination, uploads, headers, and nested depth.
- Normalize email, phone, and locale before validation. Never use raw input in SQL, shell commands, HTML, or object keys.
- Use parameterized SQL/SQLBoiler and allowlist sort/filter values.
- Return `422 validation_failed` for input, `401` for token failures, `403` for authorization, and `429` at a rate limit; do not mix authentication errors with authorization details.

## 8. Security headers and browser rules

Production Caddy/API must add:

```text
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Content-Security-Policy: default-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; connect-src 'self' https://dinar.dz wss://dinar.dz
```

Do not use `dangerouslySetInnerHTML` for user-provided content, and never expose
private MinIO objects directly. Presigned URLs must be short-lived and
operation-scoped; every administrative view must create an audit record.

## 9. Logging, monitoring, and incident response

- Redact `Authorization`, cookies, sensitive JWT claims, payment payloads, and private document paths from logs and traces.
- Monitor 401/403/429 rates, refresh reuse, signature failures, WebSocket disconnects, Casbin changes, and rejected origins.
- If a JWT key is exposed, disable the current key, publish a new key with a new `kid`, revoke affected refresh families, and restart after readiness checks.
- If a refresh token is exposed, revoke only its family and require login again. A per-access blacklist is normally unnecessary because access tokens are short-lived, except for a high-severity incident.
- Test logout, rotation, reuse detection, key rotation, CSRF, CORS, WebSocket isolation, input limits, payment webhook replay, and private MinIO access in CI.
