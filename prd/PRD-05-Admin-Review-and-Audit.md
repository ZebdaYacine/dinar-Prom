# Admin Review & Audit

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 5 — Admin Review & Audit                                         |
| **Stories**  | US-A01 … US-A04, US-A10 (5)                                      |
| **Priority** | Critical                                                        |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(admin)/admin`, `merchants`, `offers`, `payments`, `audit-logs` |
| **API**      | Admin dashboard, merchant/offer/payment queues, signed document URL, actions, audit logs |

## 1. Summary

Admins review merchants, documents, offers, and payments through searchable queues with mandatory reasons and auditable sensitive actions.

## 2. Problem

Trust decisions must be explainable and protected. A hidden frontend button is not sufficient authorization.

## 3. Goals

- Provide age-ordered queues and complete detail views.
- Make document access temporary and auditable.
- Make every rejection actionable and every decision traceable.

### Non-goals

Delegated admin roles, customer moderation accounts, and editable audit history.

## 4. Users

Platform admin/operator.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Admin role requires Google login and configured admin email. | US-A01 |
| R2 | Merchant queue supports search, status, oldest-first ordering, and full detail. | US-A02 |
| R3 | Documents use five-minute signed URLs and audit events. | US-A02 |
| R4 | Admin can approve, reject with reason/document, suspend, and reactivate. | US-A02 |
| R5 | Offer queue supports preview, approve, reject, remove, and bulk actions. | US-A03 |
| R6 | Payment queue supports proof review and reasoned approve/reject. | US-A04 |
| R7 | Audit captures actor, action, resource, time, IP, and request ID. | US-A10 |

## 6. Flow

```text
admin login → queue → detail → decision → notification + audit
```

## 7. Technical notes

Use API role middleware; signed URLs never persist in page state; bulk responses include per-ID failures; audit writes are append-only from the UI.

## 8. Subtasks

- [ ] T5.1 Build admin shell, guards, dashboard metrics, queues, search, and pagination.
- [ ] T5.2 Build merchant/document review and signed viewer.
- [ ] T5.3 Build offer/payment review and bulk operations.
- [ ] T5.4 Build audit queries, notifications, and authorization tests.

## 9. Success metrics

Review age; 24-hour SLA; actionable rejection rate; zero unauthorized admin actions.

## 10. Risks & open questions

An email allowlist may not scale. Confirm future admin role management and document retention policy.
