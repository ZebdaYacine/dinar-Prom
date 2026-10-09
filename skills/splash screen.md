# Dinar DZ — Splash Screen Specification

## 1. Goal

The splash screen is a temporary brand layer shown while the application shell
loads. It must:

1. Introduce **Dinar DZ — Dinar**.
2. Show the official logo and the active proverb stored in the database under the site name and locale.
3. Show honest progress for real loading stages.
4. Transition to the page after it is ready and remain visible for exactly three additional seconds.

The page loads behind the splash in parallel. The splash must never delay images
or JavaScript that the page needs and must be removable as a temporary overlay.

## 2. Brand identity

- Latin name: **Dinar DZ**.
- Arabic brand name: **Dinar**.
- Promise: **All your city's discounts in one place**.
- Logo: circular coin-inspired mark with the Amazigh Yaz `ⵣ`, burgundy disc `#6C151E`, and gold `#D4A24C` mark and border.
- Preferred splash lockup: vertical icon above `Dinar DZ` and `Dinar`.
- Provide a clear SVG and a PNG fallback with `aria-label="Dinar DZ"`.
- Do not rotate, add gradients, use heavy shadows, or change logo colors.
- Safe space around the mark equals the height of the `D`.
- Icon size: 72–96px on mobile and 96–128px on desktop.
- Use warm white `#FFF9F2` or cream `#F5DABF`; add a small burgundy panel only when contrast requires it.

## 3. Database proverb

The proverb must not be hardcoded in the component. Read it from active branding
data by site name and locale.

### `site_branding`

```json
{
  "site": "dinar_dz",
  "locale": "ar-DZ",
  "proverb": "All your city's discounts in one place",
  "logo_url": "/brand/dinar-dz-mark.svg",
  "logo_alt": "Dinar DZ logo",
  "is_active": true
}
```

Admins may change the proverb without rebuilding the frontend. Store plain text
only, with no HTML, Markdown, or script, and enforce a 180-character limit. Use
only `is_active=true`. If no active record exists, show the local logo, brand
name, and progress bar, report a monitoring error, and do not fail the page.

### Public endpoint

```http
GET /api/v1/site/branding?site=dinar_dz&locale=ar-DZ
```

The API must allowlist `site` and `locale`, reject table names, SQL, and client
logo URLs, and return text that the frontend renders safely.

## 4. Layout

```text
│                                    │
│              [ ⵣ ]                 │  <- logo
│            Dinar DZ                │  <- name
│              Dinar                 │
│  All your city's discounts...     │  <- database proverb
│                                    │
│  Preparing offers...               │
│  ████████████████░░░░  82%         │  <- progress bar
│  Preparing home page               │  <- current stage
│                                    │
```

| Element | Specification |
|---|---|
| Background | Warm white `#FFF9F2`, no large image or video |
| Logo | Branding SVG with local PNG fallback |
| Name | Bold Cairo/Sans `Dinar DZ`, with `Dinar` below or beside it |
| Proverb | Database text, centered, at most two mobile lines |
| Status | `Preparing the site...` followed by the current stage |
| Progress bar | 8px; cream track, burgundy fill, optional small gold marker |
| Percentage | Latin numerals `0–100%` |
| Motion | 150–200ms ease-out; no irritating continuous rotation |

## 5. Loading and progress

Progress must represent actual stages and may only move forward:

| Stage | Target | Condition |
|---|---:|---|
| Application start | 5% | Create splash and read local configuration |
| Branding | 25% | Branding response arrives or a short fallback timeout ends |
| Shell | 45% | Layout and base fonts are ready |
| Page data | 75% | First page data or SSR hydration is ready |
| Critical images | 88% | Logo and priority image are ready |
| Page ready | 100% | `page_ready=true` and no blocking request remains |

Rules:

- Never decrease progress.
- Do not exceed 92% before `page_ready` unless every critical resource is ready.
- If the API is slow, move slowly to 92% with a clear message rather than appearing stuck.
- Move to 100% within 200ms after `page_ready`.
- Do not wait for lazy images, ads, or analytics to hide the splash.
- On branding failure, use the local logo fallback and continue loading.

## 6. Timing and state machine

The timer starts when the splash becomes visible. After `page_ready`, set progress
to 100%, keep the splash visible for another `3000ms`, then remove it with a
200ms fade. If the page becomes ready before branding, wait for branding or its
fallback before starting the three-second period.

Suggested states:

```text
mount -> loading_branding -> loading_page -> page_ready
                         \-> fallback_branding
page_ready -> wait_3000ms -> fade_200ms -> removed
critical_error -> error_retry
```

Do not rely only on `window.onload` in an SSR app. Emit `page_ready` after
hydration and critical data are ready. Keep all timing constants in configuration,
not in user input or database data.

## 7. Responsive behavior

### Mobile

- Centered vertical layout.
- Logo 80–96px.
- Proverb width 80% with at most two lines.
- Progress width 78% and at least 240px.
- Content must fit within the viewport height.

### Desktop

- Logo 96–128px.
- Centered content with a maximum width of 520px.
- An extremely subtle Amazigh geometric pattern may use 6–10% opacity.
- No advertisement panel or video during the splash.

## 8. Accessibility and RTL

- Use `<html lang="ar" dir="rtl">` for Arabic.
- Use `role="status"` and `aria-live="polite"` on the status container.
- The progress bar uses `role="progressbar"`, `aria-valuemin`, `aria-valuemax`, and `aria-valuenow`.
- Never communicate progress by color alone; show the percentage and text.
- Logo `alt` comes from sanitized database branding.
- Respect `prefers-reduced-motion` while retaining the required three-second timing.
- Do not use sound or vibration.
- The splash must not block screen-reader access to the page heading during a long load.
- After removal, restore focus to the page or the element that started the action; do not leave a permanent focus trap.

## 9. Security

- Branding reads are public; changing the proverb or logo requires the admin permission.
- Do not accept HTML or inline SVG from the admin panel; `logo_url` must match a local or trusted CDN allowlist.
- Reject `javascript:`, `data:text/html`, and non-HTTPS external logo URLs.
- Never put user data, JWTs, or payment status in the proverb or branding response.
- Apply CSP to logo images, such as `img-src 'self' https://media.dinar.dz`.
- Audit branding changes with actor and timestamp without storing JWTs or refresh cookies.
- Cache public branding and invalidate it when `version` changes.

## 10. Performance

- SVG logo below 20KB and PNG fallback below 80KB.
- Do not load offer images or ads before hiding the splash.
- Do not block SSR or LCP; the main page is sent behind the overlay.
- Cache the branding endpoint with a short timeout and local fallback.
- Use CSS transitions or the Web Animations API rather than a large animation library.
- Keep the fixed splash container from causing CLS.

## 11. Error states and acceptance

| State | Behavior |
|---|---|
| Branding API unavailable | Local logo/name fallback, metric, continue site loading |
| Invalid JSON | Ignore response, never render HTML, use fallback |
| Logo fails | Local PNG/SVG fallback without breaking layout |
| Page takes over 15 seconds | Show “We are still preparing the page” and a retry button |
| First page fails | Show the complete error after timeout; do not hide it behind splash |
| Offline | Use cached branding if available and explain that new data needs a connection |

Acceptance checklist:

- [ ] `Dinar DZ`, `Dinar`, and the correct logo are visible.
- [ ] The active database proverb is displayed for `dinar_dz` and `ar-DZ`.
- [ ] A database proverb change appears after cache invalidation without a rebuild.
- [ ] Progress starts at 0–5%, moves forward, and never moves backward.
- [ ] Progress never reaches 100% before `page_ready`.
- [ ] The splash remains for 3000ms after `page_ready` and fades for 200ms.
- [ ] Ads, lazy images, and analytics do not delay removal.
- [ ] Card data and authentication tokens never appear in branding, DOM, or logs.
- [ ] Mobile 360px, RTL, reduced motion, branding failure, and page timeout work.
- [ ] Lighthouse shows no CLS caused by the splash and LCP remains within budget.
