# Dinar DZ — Chargily Pay Integration and Secure Payment Flow

The general billing behavior is defined in [`../userstory.md`](../userstory.md).
This document covers the merchant subscription only: 4,000 DZD for 365 days.
Dinar DZ never collects product prices or processes customer-to-merchant product
payments.

Chargily's account documentation is authoritative for field and event names;
verify every request against the real account before launch.

## 1. Mandatory security principle

Card data must never enter Dinar DZ servers. Redirect the merchant to Chargily's
hosted Checkout and keep card data inside Chargily's environment.

Never store, log, or send:

- Full PAN.
- CVV/CVC, PIN, or OTP.
- Card expiry date.
- Card images or payment-page screenshots.
- CIB/EDAHABIA passwords or credentials.
- Raw webhook bodies when they contain unnecessary sensitive fields.

Store only the transaction record: internal ID, merchant ID, Chargily Checkout
ID, amount, currency, plan, status, timestamps, and the verified event ID. V1
does not need a card table or Customer Card Token.

## 2. Chargily configuration

| Mode | Base URL |
|---|---|
| Test | `https://pay.chargily.net/test/api/v2` |
| Live | `https://pay.chargily.net/api/v2` |

The secret key is sent by the server only:

```http
Authorization: Bearer <CHARGILY_SECRET_KEY>
```

Key rules:

1. Store Test and Live keys in a secret manager or separate environment files.
2. Never expose keys in Git, images, logs, JavaScript, or `NEXT_PUBLIC_*`.
3. Never use a Test key with Live or the reverse; verify `livemode` on every webhook.
4. Rotate keys after exposure or team membership changes, with a short overlap plan.
5. Use HTTPS and verify TLS for every Chargily connection.

See Chargily's [API keys](https://dev.chargily.com/pay-v2/api-keys) and
[authentication](https://dev.chargily.com/pay-v2/api-reference/authentication)
documentation.

## 3. Stored data

### 3.1 `payments`

```text
id                         uuid primary key
merchant_id                uuid not null
provider                   text not null              -- chargily
provider_checkout_id       text unique not null
plan                       text not null              -- annual
amount_dzd                 integer not null           -- 4000 from server config
currency                   text not null              -- dzd
status                     text not null              -- pending/paid/failed/canceled
livemode                   boolean not null
provider_event_id          text unique null
created_at                 timestamptz not null
updated_at                 timestamptz not null
paid_at                    timestamptz null
failure_code               text null                  -- non-sensitive code only
```

### 3.2 `payment_events`

```text
id                         uuid primary key
provider                   text not null
provider_event_id          text unique not null
provider_checkout_id       text null
event_type                 text not null
signature_verified         boolean not null
payload_sha256             bytea not null
processing_status          text not null              -- received/processed/ignored/failed
received_at                timestamptz not null
processed_at               timestamptz null
processing_error_code      text null
```

Do not store `raw_payload` by default. If an investigation requires it, store an
encrypted redacted copy with payment-method and customer-sensitive fields
removed and a short retention period. `payload_sha256` is sufficient to prove
that a payload was unchanged without retaining its contents.

If Chargily supports `metadata`, store only:

```json
{
  "payment_id": "internal-uuid",
  "purpose": "merchant_annual_subscription"
}
```

Never put email, phone, merchant name, documents, or card data in metadata. The
local database remains the source of truth for amount and ownership.

## 4. Complete payment flow

```text
Merchant clicks subscribe
        |
        v
Server calculates amount and creates payment=pending
        |
        v
Server creates a Chargily Checkout with the secret key
        |
        v
Frontend redirects to the hosted checkout_url
        |
        v
Chargily processes payment and sends redirect + webhook
        |
        +--> redirect: display status only; never activate access
        |
        +--> webhook: read raw body
                    -> verify HMAC signature
                    -> validate schema, checkout, amount, and mode
                    -> idempotency inside a transaction
                    -> mark paid and activate subscription
```

### Stage A — Create the local payment

`POST /api/v1/merchant/billing/checkout` requires an authenticated merchant.

1. Verify the merchant is approved and not suspended.
2. Calculate the price from server configuration, not the client body; V1 is 4,000 DZD.
3. Create `payments` with `pending` status and an initially empty provider checkout ID.
4. Create an internal idempotency key to prevent duplicate Checkouts after double-clicks or retries.
5. Never accept `amount_dzd` or `currency` from the client as the source of truth.

Request:

```json
{ "plan": "annual" }
```

### Stage B — Create the Chargily Checkout

```http
POST https://pay.chargily.net/test/api/v2/checkouts
Authorization: Bearer <secret>
Content-Type: application/json
```

Example Test payload:

```json
{
  "amount": 4000,
  "currency": "dzd",
  "success_url": "https://dinar.dz/dashboard/billing/success",
  "failure_url": "https://dinar.dz/dashboard/billing/failure",
  "webhook_endpoint": "https://dinar.dz/api/v1/webhooks/chargily",
  "description": "Dinar DZ annual merchant subscription",
  "locale": "ar",
  "metadata": {
    "payment_id": "internal-uuid",
    "purpose": "merchant_annual_subscription"
  }
}
```

Use the exact webhook field required by the Chargily API version in the account;
never guess provider fields in production.

Validate at least:

- Non-empty `id`.
- HTTPS `checkout_url` matching the expected Chargily domain.
- `amount == 4000` and `currency == "dzd"`.
- `status == "pending"` at creation.
- `livemode` matching the Dinar DZ environment.

After validation, save the Checkout ID and return only:

```json
{
  "data": {
    "payment_id": "internal-uuid",
    "checkout_url": "https://pay.chargily.dz/test/checkouts/.../pay",
    "amount_dzd": 4000,
    "currency": "dzd",
    "status": "pending"
  }
}
```

Never return the full Chargily response, secret key, or internal metadata to the browser.

### Stage C — Hosted payment page

Redirect directly to `checkout_url`. Never build a card form inside Dinar DZ and
never use an iframe or JavaScript that reads card fields.

### Stage D — Return to Dinar DZ

`success_url` and `failure_url` are display-only:

- Success: show “Payment confirmation pending” and query `GET /api/v1/merchant/billing/payments/:id`.
- Failure: show the failure and allow a new Checkout.
- Never change `payments.status` to `paid` from a redirect, query string, or JavaScript result.
- Never accept a `payment_id` belonging to another merchant; always bind it to the authenticated JWT subject.

The webhook is the payment confirmation source. A redirect never replaces the
verified webhook.

## 5. Chargily webhook endpoint

```http
POST /api/v1/webhooks/chargily
Content-Type: application/json
signature: <hex-hmac-sha256>
```

This public endpoint has no user JWT, but accepts only POST. Disable CSRF for this
external provider endpoint specifically while keeping every other protection.

Mandatory processing order:

1. Validate method, content type, content length, and body limit.
2. Read the raw body bytes before JSON parsing or reserialization.
3. Read the `signature` header.
4. Calculate `expected = hex(HMAC-SHA256(CHARGILY_SECRET_KEY, raw_body))`.
5. Compare signature and expected using constant-time comparison such as Go's `hmac.Equal`.
6. If missing or invalid, return `400`/`403` without parsing JSON or changing the database.
7. Only after signature verification, decode JSON and validate the schema.
8. Validate `type`, `id`, `data.id`, `data.status`, `data.amount`, `data.currency`, `data.livemode`, and metadata.
9. Link `data.id` to `payments.provider_checkout_id`; never trust a client merchant ID.
10. Confirm amount, currency, mode, and plan match the local record.

For a valid event, use a transaction and a unique `(provider, provider_event_id)`
constraint. Duplicate events must produce the same result without extending the
subscription twice or sending duplicate mail. If the Checkout amount differs,
record `security_mismatch` and alert the team.

On a successful payment, lock the payment row, verify the provider Checkout ID,
mark the payment paid once, activate the subscription once, create the durable
notification, and publish a live notification only after commit. Return `200` only
after the event is safely recorded; return `5xx` for retryable processing failures.

## 6. API security checklist

- [ ] Hosted Checkout only; Dinar DZ never processes card data.
- [ ] Test and Live keys are separated and stored as secrets.
- [ ] HTTPS and TLS verification are enabled.
- [ ] No secrets, raw webhook body, or card data appear in logs.
- [ ] HMAC is verified against the raw body with constant-time comparison.
- [ ] Amount, currency, mode, Checkout ownership, and event type are validated.
- [ ] Idempotency, unique constraints, and row locks are implemented.
- [ ] Success redirects and query parameters never activate subscriptions.
- [ ] The client cannot set the price.
- [ ] A merchant cannot read another merchant's payment.
- [ ] Rate limits, body limits, and timeouts protect the webhook.
- [ ] Audit records contain safe success/failure and mismatch details only.

## 7. Frontend checklist

- [ ] Do not store Checkout URLs or sensitive payment status in local storage.
- [ ] Do not collect card fields with JavaScript.
- [ ] Show “paid” only after reading the local API state `paid`.
- [ ] Require an HTTPS Chargily Checkout URL from the expected domain.
- [ ] Ignore `success`, `paid`, and `status` query parameters.
- [ ] Show a clear pending state while waiting for the webhook.
