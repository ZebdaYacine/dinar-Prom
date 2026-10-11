# Merchant Offers

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 3 — Merchant Offers                                               |
| **Stories**  | US-O01 … US-O07 (7)                                               |
| **Priority** | Critical                                                        |
| **Status**   | ✅ Shipped                                                       |
| **Surfaces** | `apps/web/app/(merchant)/dashboard/offers`, `offers/new`, `offers/[id]/edit`, `stats` |
| **API**      | `/merchant/offers`, `/merchant/offers/:id`, media endpoints, submit/archive/duplicate, `/merchant/stats` |

## 1. Summary

Approved merchants create mobile-first offer drafts, upload up to ten optimized images, submit for review, manage lifecycle, and inspect privacy-safe statistics.

## 2. Problem

Publishing must be fast but strict enough to protect customers from incomplete, expired, misleading, or oversized content.

## 3. Goals

- Make a valid offer possible from a phone.
- Autosave drafts and enforce business rules server-side.
- Optimize images and preserve moderation history.

### Non-goals

Video, product checkout, unlimited active offers, and permanent deletion of approved offers.

## 4. Users

Approved merchant and admin reviewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Validate title, description, category, prices/discount, dates, location, and WhatsApp. | US-O01 |
| R2 | Autosave every 10 seconds and enforce configurable active-offer limit. | US-O01 |
| R3 | Accept 1–10 JPG/PNG/WebP images, max 8MB, min 400×400. | US-O02 |
| R4 | Convert server-side to WebP 400/800/1600 and remove EXIF. | US-O02 |
| R5 | Submit moves to pending; approved-offer edits hide it and return it to pending. | US-O03, US-O04 |
| R6 | Support pause, archive, draft delete, and duplicate. | US-O05, US-O06 |
| R7 | Show views, WhatsApp clicks, 30-day series, and totals. | US-O07 |

## 6. Flow

```text
dashboard → draft editor → media → submit → pending → admin approval → public
                                      └──────────────→ rejection → correction
```

## 7. Technical notes

The server calculates discounts, validates 1–99 percentages and 180-day maximum duration, checks ownership for every mutation, and uses `public_offers` for public reads.

## 8. Subtasks

- [x] T3.1 Implement offer state machine, CRUD, ownership, validation, and autosave.
- [x] T3.2 Build editor, media uploader, ordering, cover selection, and rejection UI.
- [x] T3.3 Implement image processing, cleanup, duplicate, archive, and stats aggregation.
- [x] T3.4 Add merchant and moderation E2E flows.

## 9. Success metrics

Draft-to-submit completion; median creation time; image failure rate; invalid/rejected offer rate.

## 10. Risks & open questions

Image processing can consume CPU. Confirm the active-offer limit and whether processing should be worker-backed.
