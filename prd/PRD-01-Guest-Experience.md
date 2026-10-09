# Guest Experience

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 1 — Guest Experience                                             |
| **Stories**  | US-V01 … US-V07 (7)                                               |
| **Priority** | High                                                             |
| **Status**   | 🟠 In Progress                                                    |
| **Surfaces** | `apps/web/app/(public)/page.tsx`, `offers`, `offers/[id]`, `shop/[slug]` |
| **API**      | `GET /offers`, `GET /offers/:id`, `GET /shops/:slug`, `POST /offers/:id/view`, `POST /offers/:id/whatsapp`, `POST /offers/:id/report`, `GET /categories`, `GET /wilayas`, `GET /wilayas/:code/cities` |

## 1. Summary

Guests browse current discounts, filter them with Arabic search, open details, view a shop, share an offer, report abuse, and contact merchants through WhatsApp without registration.

## 2. Problem

Visitors must see real value before signup. Stale, hidden, or inaccessible offers reduce trust for guests and merchants.

## 3. Goals

- Show only results from the single `public_offers` visibility rule.
- Reach offer details in one click and keep filters shareable in the URL.
- Make WhatsApp contact one tap on mobile.

### Non-goals

Customer accounts, cart, product payments, favorites, and native mobile apps.

## 4. Users

**Guest visitor** — arrives from search, a shared link, or a local merchant promotion.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Home shows search, icon-based horizontal categories, nested subcategory chips, latest, ending-soon, and highest-discount sections. | US-V01 |
| R2 | Cards show cover, title, prices, discount, shop, wilaya, and countdown. | US-V01 |
| R3 | Search supports Arabic normalization, category/subcategory filters, 58-wilaya and commune filters, sorting, URL state, empty state, and pagination. | US-V02 |
| R4 | Detail shows gallery, dates, shop, verified badge, similar offers, OG metadata, and 404/410 states. | US-V03 |
| R5 | WhatsApp/phone actions record safe, deduplicated analytics. | US-V04 |
| R6 | Shop page shows public shop details and active offers only. | US-V05 |
| R7 | Reports enforce reason, message limit, IP rate limit, and duplicate protection. | US-V06 |
| R8 | Share uses Web Share API or clipboard. | US-V07 |

## 6. Flow

```text
Home → search/filter → offers → detail → WhatsApp/report/share
                         └────→ shop page
```

## 7. Technical notes

SSR the first view; normalize Arabic on the server; lazy-load secondary images and ads; rate-limit public actions; never expose hidden content in 404/410 responses.

The guest catalogue exposes the complete 58-wilaya / 1,541-commune reference set. The web UI renders horizontal, keyboard-accessible wilaya and commune rails and nested category/subcategory rails with local SVG icons.

## 8. Subtasks

- [x] T1.1 Build public routes, cards, filters, detail gallery, shop page, and states, including responsive category and location explorers.
- [x] T1.2 Implement public queries, Arabic search, pagination, nested category filters, and URL-based wilaya/commune filters.
- [ ] T1.3 Implement WhatsApp, sharing, view/click deduplication, reports, and rate limits.
- [ ] T1.4 Add SEO metadata and Playwright guest journey.

## 9. Success metrics

LCP under 2.5s; offer-detail CTR; WhatsApp CTR; public-offer freshness; report SLA.

## 10. Risks & open questions

Search quality depends on Arabic seed data. Confirm official wilaya data and similar-offer ranking before launch.

## 11. Implementation checkpoint

The current implementation delivers the guest HTTP and SSR experience against an isolated catalog adapter in `apps/api/internal/guest/catalog.go`. The adapter exposes the documented public endpoints, applies the approved/access/date visibility rule, normalizes Arabic search, paginates results, generates WhatsApp links, deduplicates safe visitor actions, and limits reports. The web application provides the home, listing, detail, shop, share, WhatsApp, report, responsive, and localized surfaces under `apps/web/src/`.

Local verification completed:

- `apps/api`: `go test ./...`
- `apps/web`: `pnpm typecheck`, `pnpm lint`, `pnpm test`, `pnpm build`
- Docker smoke checks: `/healthz`, `/api/v1/offers`, Arabic search, Arabic and French listing pages, offer detail, shop page, and missing-offer 404

The epic is intentionally still **In Progress**. Replacing the adapter with the PostgreSQL `public_offers` query, persisting metrics/rate-limit state through the platform data layer, and adding the Playwright guest journey remain required before the PRD can be marked Done.
