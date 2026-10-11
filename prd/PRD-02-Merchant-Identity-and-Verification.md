# Merchant Identity & Verification

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 2 — Merchant Identity & Verification                             |
| **Stories**  | US-M01 … US-M05 (5)                                               |
| **Priority** | Critical                                                        |
| **Status**   | 🟠 In Progress                                                   |
| **Surfaces** | `apps/web/app/(auth)/login`, `onboarding/personal`, `shop`, `documents`, `review` |
| **API**      | Google auth, `GET /me`, `PUT /merchant/profile`, `PUT /merchant/shop`, `POST /merchant/documents`, `POST /merchant/submit` |

## 1. Summary

A merchant signs in with Google, completes personal/shop data, uploads identity and commercial-register documents, accepts privacy consent, and follows verification status.

## 2. Problem

The marketplace needs trustworthy merchants while protecting sensitive documents and keeping onboarding simple on mobile.

## 3. Goals

- Secure Google OAuth with verified email, state, PKCE, and HttpOnly sessions.
- Autosave onboarding and resume the last incomplete step.
- Keep documents private and rejection status actionable.

### Non-goals

Passwords, customer accounts, automatic document approval, and third-party KYC in V1.

## 4. Users

Prospective merchant and admin reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Google OAuth creates/resolves a merchant session and rejects unverified email. | US-M01 |
| R2 | Personal form validates Algerian phone, wilaya, address, and autosaves. | US-M02 |
| R3 | Shop form captures identity, location, contact channels, logo, and register expiry. | US-M03 |
| R4 | Slug is unique and editable once before verification. | US-M03 |
| R5 | Three documents accept verified content types up to 8MB in private storage. | US-M04 |
| R6 | Consent timestamp is stored; status supports incomplete/pending/approved/rejected/suspended. | US-M04, US-M05 |

## 6. Flow

```text
Google login → personal → shop → documents → review → pending → approved/rejected
```

## 7. Technical notes

Store only session hashes; use secure cookies; detect MIME by magic bytes; return signed five-minute document URLs only to admins; audit every document view; apply CSRF except the external OAuth callback.

## 8. Subtasks

- [~] T2.1 Implement OAuth, sessions, cookies, logout, and role middleware.
- [~] T2.2 Build onboarding wizard, autosave, validation, resume, and consent.
- [~] T2.3 Implement private upload/storage, replacement, hashing, and rejection flow.
- [ ] T2.4 Add status emails and authorization/integration tests.

## 9. Success metrics

Onboarding completion rate; time to submission; rejection reasons; zero unauthorized document reads.

## 10. Risks & open questions

Confirm Algerian privacy/legal retention obligations and the final admin-role management policy.
