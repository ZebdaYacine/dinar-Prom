# Web Shell, Brand & Splash

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **Epic**     | Platform enabler — web experience and identity                   |
| **Stories**  | Cross-cutting support for US-V01 … US-S04                         |
| **Priority** | High                                                             |
| **Status**   | 🟡 Planned                                                       |
| **Surfaces** | `apps/web/app/layout.tsx`, route groups, components, global styles, `site/branding` |
| **API**      | `GET /site/branding` and public APIs used by route groups       |

## 1. Summary

The shell establishes Arabic RTL, Dinar DZ identity, reusable components, loading states, and a database-driven Splash screen.

## 2. Problem

Feature screens can drift in design, RTL behavior, accessibility, loading, and brand expression without a shared shell.

## 3. Goals

- Apply the design system consistently.
- Provide responsive, accessible, Arabic-first navigation.
- Show logo, site name, proverb, and truthful progress in Splash.

### Non-goals

Final logo redesign, complete French content, and native mobile app.

## 4. Users

All guests, merchants, admins, and brand administrators.

## 5. Requirements

| # | Requirement | Source |
|---|---|---|
| R1 | Use Dinar DZ name, approved logo, burgundy/green/cream/orange/gold palette. | `skills/ui.md` |
| R2 | Use RTL, logical properties, approved fonts, visible focus, and 44px touch targets. | `skills/ui.md`, `skills/frontend.md` |
| R3 | Build shared header, bottom nav, cards, forms, modals, states, and typography. | `skills/ui.md` |
| R4 | Splash reads active `dinar_dz` branding/proverb and holds three seconds after `page_ready`. | `skills/splash screen.md` |
| R5 | Cover fallback, reduced motion, timeout, error, and no-CLS cases. | `skills/splash screen.md` |
| R6 | SSR public routes and isolate merchant/admin bundles and guards. | `skills/frontend.md` |

## 6. Flow

```text
app boot → branding/page data → page_ready → 3s hold → fade → usable page
```

## 7. Technical notes

Keep Splash fixed and non-blocking to SSR; store branding in `site_branding`; sanitize text and allowlist logo URL; expose progress text and `role="progressbar"`; honor reduced motion.

## 8. Subtasks

- [ ] T9.1 Build root layout, fonts, theme tokens, RTL, metadata, and navigation.
- [ ] T9.2 Build reusable components and visual states.
- [ ] T9.3 Add branding endpoint, cache, fallback, admin update, and audit hook.
- [ ] T9.4 Implement Splash progress, three-second hold, fade, timeout, and accessibility.
- [ ] T9.5 Run responsive/visual QA at 360px, desktop, reduced motion, and 200% text scale.

## 9. Success metrics

CLS under 0.1; axe pass; brand consistency; no Splash-induced LCP regression.

## 10. Risks & open questions

Measure whether the three-second post-ready hold harms bounce; confirm final SVG outline and brand legal availability.
