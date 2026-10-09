# Trial, Billing & Payments

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 4 — Trial, Billing & Payments                                    |
| **Stories**  | US-B01 … US-B08 (8)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/billing`, `apps/api/internal/billing`, worker |
| **API**      | Billing checkout/manual/history/status and `POST /webhooks/chargily` |

## 1. Summary

The first approved offer starts one 60-day trial. Merchants then pay 4000 DZD for 365 days through hosted Chargily Checkout or submit manual proof for admin approval.

## 2. Problem

Money and access state must remain correct through redirects, duplicate webhooks, expiry, retries, and manual review while card data stays outside Dinar DZ.

## 3. Goals

- Start trial atomically and only once.
- Activate paid access only after verified webhook/admin review.
- Never store PAN, CVV, expiry, screenshots, or raw sensitive payloads.
- Hide and restore offers based on access.

### Non-goals

Product payments, stored cards, recurring card charges, automated refunds, or customer wallets.

## 4. Users

Verified merchant and admin payment reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | First approved offer starts one 60-day trial in the same transaction. | US-B01 |
| R2 | Dashboard shows countdown and warning at 14 days or less. | US-B02 |
| R3 | Reminder jobs at 14/7/1/0 days are deduplicated. | US-B03 |
| R4 | Server creates a 4000 DZD hosted Chargily Checkout. | US-B04 |
| R5 | Success redirect never activates access; signed HMAC webhook is required. | US-B04 |
| R6 | Manual proof/reference enters under-review until admin decision. | US-B05 |
| R7 | Payment history and receipts contain no card data. | US-B06 |
| R8 | Expiry hides offers; renewal starts at `max(now, access_until)`. | US-B07, US-B08 |

## 6. Flow

```text
first approval → trial → hosted Checkout/manual proof → verified payment → subscribed → expiry/renewal
```

## 7. Technical notes

See `../skills/payment.md`. Verify raw-body HMAC before parsing; validate event ID, checkout ID, amount, currency, mode, metadata, ownership, and status; use unique constraints, row locks, and idempotent transitions.

## 8. Subtasks

- [ ] T4.1 Implement access/trial/subscription calculations and atomic transitions.
- [ ] T4.2 Build billing UI, history, manual proof, success/failure, and pending states.
- [ ] T4.3 Integrate Chargily Test/Live Checkout and signed webhook processing.
- [ ] T4.4 Implement manual review, reminders, expiry, receipts, reconciliation, and tests.

## 9. Success metrics

Payment accuracy; zero duplicate activations; trial-to-paid conversion; zero card fields in DB/logs/backups/client payloads.

## 10. Risks & open questions

Confirm Chargily live event/field names, tax/receipt requirements, and manual review capacity.
