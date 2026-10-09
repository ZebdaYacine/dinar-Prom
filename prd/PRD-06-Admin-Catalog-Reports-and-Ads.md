# Admin Catalog, Reports & Ads

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | 6 — Admin Catalog, Reports & Ads                                 |
| **Stories**  | US-A05 … US-A09, US-D01 (6)                                      |
| **Priority** | Medium                                                          |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/(admin)/categories`, `ads`, `reports`, `settings`, public `AdSlot` |
| **API**      | CRUD categories/ads, reports list/resolve, settings, public ads/click |

## 1. Summary

Admins operate taxonomy, reports, direct-sold ads, dashboard indicators, and business settings without code changes.

## 2. Problem

Operational configuration changes more frequently than application releases. Hard-coded categories and ads make moderation and monetization slow.

## 3. Goals

- Manage two-level categories safely.
- Resolve reports with explicit side effects.
- Run measurable, labeled, layout-stable direct ads.
- Expose configurable limits and business indicators.

### Non-goals

AdSense, ad auctions, and more than two category levels.

## 4. Users

Admin/operator and guest ad viewer.

## 5. Requirements

| # | Requirement | Story |
|---|---|---|
| R1 | Add/edit/reorder/hide categories; prevent deleting referenced categories. | US-A05 |
| R2 | Group reports by offer and expose reason, message, unique IP count, and status. | US-A06 |
| R3 | Ignore, hide offer, or suspend merchant from report resolution. | US-A06 |
| R4 | Ads support advertiser, media, URL, slot, dates, priority, targeting, impressions, and clicks. | US-A07 |
| R5 | Dashboard shows merchant, offer, access, revenue, and top-category indicators. | US-A08 |
| R6 | Settings control trial, price, subscription days, limits, and manual payment data. | US-A09 |
| R7 | Public ad slots reserve space, show “Advertisement”, and use sponsored links. | US-D01 |

## 6. Flow

```text
admin config → validation/audit → cached public API → ad/report/dashboard result
```

## 7. Technical notes

Validate HTTPS targets; filter ads by slot/date/targeting; aggregate metrics daily; reserve aspect ratio to protect CLS; audit settings changes.

## 8. Subtasks

- [ ] T6.1 Build category CRUD, hierarchy, ordering, visibility, and seeds.
- [ ] T6.2 Build report aggregation, resolution actions, and side effects.
- [ ] T6.3 Build ad CRUD, media, targeting, slot rendering, and metrics.
- [ ] T6.4 Build dashboard/settings UI and API.

## 9. Success metrics

Report resolution SLA; ad fill/viewability/CTR; zero ad-caused CLS; category completeness.

## 10. Risks & open questions

Define ad contracts, payment reconciliation, dashboard timezone, and metrics aggregation windows.
