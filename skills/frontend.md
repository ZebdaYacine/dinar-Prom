---
name: dinar-dz-frontend
description: Architecture, routing, data fetching, forms, auth, SEO, i18n, and performance rules for the Dinar DZ Next.js 15 frontend (TypeScript, Tailwind 4, DaisyUI 5, next-intl, Arabic/French RTL/LTR). Use this skill whenever building or changing any page, component, hook, form, route, layout, or API client in the Dinar DZ web app.
---

# Dinar DZ — Frontend (Next.js + TypeScript)

The main behavior reference is [`../userstory.md`](../userstory.md). The API
contract is in `backend.md`, and visual rules are in `ui.md`.

## 1. Technology

| Purpose | Tool |
|---|---|
| Framework | Next.js 15 App Router + React 19 |
| Language | Strict TypeScript (`strict`, `noUncheckedIndexedAccess`) |
| Styling | Tailwind CSS 4 + DaisyUI 5 |
| Client data | TanStack Query v5 |
| Forms | React Hook Form + Zod (`@hookform/resolvers`) |
| API types | `openapi-typescript` generated from the backend OpenAPI document |
| Icons | `lucide-react` |
| Drag and drop | `@dnd-kit/sortable` |
| Browser image compression | `browser-image-compression` |
| Dates | `date-fns` + `date-fns-tz` (`Africa/Algiers`) |
| Toasts | `sonner` |
| Live notifications | Browser WebSocket API through Caddy, REST for resync, TanStack Query for state |
| i18n | `next-intl` with only `ar` and `fr`, using matching JSON message keys |
| Testing | Vitest, Testing Library, and Playwright |
| Quality | ESLint, Prettier, and `tsc --noEmit` in CI |

## 2. Project structure

```text
apps/web/
  src/
    app/[locale]/
      (public)/                 # home, offers, shops, legal pages
      (auth)/login/page.tsx
      (merchant)/               # onboarding, dashboard, billing, profile
      (admin)/admin/             # review, offers, payments, reports, settings
    components/
      ui/ offers/ merchant/ admin/ ads/ layout/
    lib/
      api/                       # client, server, generated types, endpoints
      auth/                      # JWT memory store, refresh, guards
      format/ validation/ seo/
    hooks/
    i18n/                        # routing, request, navigation
    styles/globals.css
    messages/ar.json
    messages/fr.json
  public/ next.config.ts middleware.ts
```

Rules:

- Keep `page.tsx` files thin: fetch data and compose domain components.
- Move a component to `components/<domain>` when more than one page uses it.
- Business rules such as eligibility, access, and official discounts belong to
  the API. The frontend may validate for feedback, but the backend is the authority.
- Every route is under `app/[locale]`; supported locales are `ar` and `fr`, with `ar` as default.
- Never hardcode translated user-facing text. Add matching keys to both message files.

## 3. Rendering and data

| Page | Strategy |
|---|---|
| Home, lists, details, shop | Server Components + `fetch`, `revalidate: 60` |
| Filter search | SSR reads `searchParams`; filters update the URL with `router.replace` |
| Merchant/admin dashboards | Client Components + TanStack Query behind auth guards |
| Countdown and payment status | `useQuery` with an appropriate `refetchInterval` |

### 3.1 API client

- `lib/api/server.ts` fetches from `API_INTERNAL_URL` for server-side requests and uses a request-scoped access token.
- `lib/api/client.ts` calls `/api/v1` on the same origin, sends a Bearer access JWT from memory, and uses `credentials: 'include'` only for refresh/logout flows.
- Every response goes through an envelope parser for `{data, meta}` and throws `ApiError(code, message, fields)` on errors.
- Generate types with `pnpm gen:api`; do not write response types manually.
- Centralize TanStack Query keys such as `qk.offers.list(params)` and `qk.merchant.me()`.

### 3.2 Caching

- Public pages use `revalidate = 60` and do not depend on cookies.
- Reference lists such as categories and wilayas use `revalidate: 3600`.
- After a merchant mutation, invalidate the relevant dashboard query only.

### 3.3 `next-intl`

```ts
// src/i18n/routing.ts
import { defineRouting } from 'next-intl/routing';

export const routing = defineRouting({
  locales: ['ar', 'fr'],
  defaultLocale: 'ar',
  localePrefix: 'always',
});
```

- `request.ts` validates `ar`/`fr` and loads `messages/{locale}.json`.
- Middleware preserves locale, `next`, and query parameters.
- `/` redirects to `/ar`; public links always include `/ar/...` or `/fr/...`.
- The locale layout sets `lang`, `dir`, and `NextIntlClientProvider`.
- Use `useTranslations` in client components and `getTranslations` in server components and metadata.
- Use the locale-aware `Link`, `useRouter`, and `usePathname` exports.
- The header switcher shows the Algeria flag with `AR` and the France flag with `FR`, and never shows a third language.
- Switching language preserves the page and query string and changes `rtl`/`ltr` direction.
- Keep shared keys in namespaces such as `common`, `header`, `notifications`, `offers`, `auth`, `merchant`, `admin`, `splash`, and `footer`.
- CI must fail when message keys differ between `ar.json` and `fr.json`.

## 4. Authentication and guards

- Keep the access JWT in memory inside an `AuthProvider` or module store only.
- Never use `localStorage`, `sessionStorage`, IndexedDB, or URLs for tokens.
- The API keeps the refresh JWT in `HttpOnly; Secure; SameSite=Strict` cookie and exposes it only through `POST /api/v1/auth/refresh`.
- Middleware must not decode JWTs or treat a cookie's presence as proof of validity; the backend or BFF performs the real check.
- `getCurrentUser()` uses a request-scoped access token and never reuses a token between users.
- The client adds `Authorization: Bearer`, allows only one concurrent refresh, and retries a failed request once after a successful refresh.
- On refresh failure, clear the in-memory access token and redirect to `/{locale}/login?next=...`.
- Refresh and logout use `credentials: include` and `X-CSRF-Token`; no token belongs in a body or query string.
- After login, route by `/me`: admins to `/admin`, incomplete merchants to onboarding, pending/rejected merchants to their status page, approved merchants to `/dashboard`, and suspended merchants to the suspension page.
- Login is a normal link to `/api/v1/auth/google/login?next=...`; do not add a Google client library.
- The admin layout must return `notFound()` for users without the admin permission.

## 5. Forms and validation

- Zod schemas in `lib/validation` provide client feedback only; the backend is authoritative. Map `error.fields` through `setError`.
- Phone input accepts `05xx` or `+2135xx` and sends a normalized E.164 value from the backend.
- Date ranges use `Africa/Algiers` for display and send UTC ISO values.
- Autosave drafts every 10 seconds with a visible saved state.
- Disable submit buttons while `isSubmitting` to prevent duplicates.
- Error messages come from the `errors` namespace, never from hardcoded strings.

### 5.1 Merchant onboarding

- Three independently navigable steps with a progress bar.
- Step three uploads three documents, records privacy consent, and submits the review.
- After submission, show the pending state, expected review time, and contact action.

### 5.2 Offer form and images

- Offer form sections: basic information, pricing/discount, dates, images, location, and contact.
- Show a live desktop preview and a tabbed mobile preview.
- `ImageUploader` accepts JPEG/PNG/WebP from drag-and-drop, gallery, or camera.
- Compress to a 2000px longest side at quality 0.85 before upload and use a web worker.
- Upload at most three images in parallel, show per-image progress, and retry failures.
- Allow at most 10 images; support drag sorting, cover selection, and confirmed deletion.
- Reject unsupported formats, oversized files, images smaller than 400px, and network failures with clear messages.

## 6. Public pages

- URL filters use `q`, `wilaya`, `category`, `min_discount`, `ending`, `sort`, and `page`.
- On mobile, filters open in a bottom drawer with removable chips.
- Search uses a 300ms debounce.
- `OfferGrid` uses two columns on phones, three on tablets, and four on desktop.
- Pagination uses real `<a href>` links for SEO; “load more” is an optional enhancement.
- Insert a `list_inline` advertisement after every eight cards.
- Details use `generateMetadata`, Open Graph cover images, canonical URLs, JSON-LD, a mobile WhatsApp bar, countdown, view beacon, and similar offers.
- Expired or hidden offers return a 410/404 state with a clear unavailable message.
- Ad slots reserve their aspect ratio, record impressions once with IntersectionObserver, show an `Advertisement` label, and use `rel="sponsored noopener"`.

## 7. Admin UI

- Tables provide search, status filters, pagination, and multi-select bulk actions.
- `DocViewer` requests a short-lived signed URL only when opened, clears it on close, and displays a visible “sensitive document access is audited” warning.
- Rejection forms require a reason from templates or free text.
- Never rely on frontend controls for authorization; every action is enforced by the API.

## 8. Performance

- Target LCP < 2.5s, CLS < 0.1, and INP < 200ms on a mid-range phone over 4G.
- Keep home-page JavaScript below 150KB gzip and split dashboards with route groups and `dynamic()`.
- Use `next/image` with `_400/_800/_1600` CDN variants, correct `sizes`, priority only for the first image, and lazy loading for the rest.
- Load only the Arabic and Latin font subsets and use `display: swap`.
- Avoid heavy libraries such as Moment and full Lodash; import icons individually.
- Run bundle analysis in CI with a failure threshold.

## 9. SEO, RTL, and accessibility

- `sitemap.ts` reads active offer, shop, and category URLs; `robots.ts` blocks dashboards, admin, onboarding, and API routes.
- Set `lang`/`dir` from `[locale]`: Arabic is `ar`/`rtl`, French is `fr`/`ltr`.
- Add alternate language metadata for `/ar/...` and `/fr/...`; use `noindex` for free-text search results.
- Use logical Tailwind properties (`ms-*`, `me-*`, `ps-*`, `pe-*`, `start-*`, `end-*`, `text-start`), never physical left/right utilities.
- Mirror directional icons with `rtl:rotate-180`.
- Format DZD consistently and pass the active locale to all date, number, and currency formatters.
- Associate labels and errors, keep visible focus, support Escape and focus trapping, provide image alt text, and respect `prefers-reduced-motion`.

## 10. Frontend security

- Use `dangerouslySetInnerHTML` only for server-generated JSON-LD with escaped `<`.
- Render offer descriptions as text with `whitespace-pre-line`.
- Use `rel="noopener noreferrer"` for external links and `sponsored` for ads.
- Configure CSP in `next.config.ts`; allow only approved image, API, media, and WebSocket origins.
- Never expose backend secrets through `NEXT_PUBLIC_*`.

## 11. Environment and live notifications

```dotenv
NEXT_PUBLIC_APP_URL=https://dinar.dz
NEXT_PUBLIC_MEDIA_BASE_URL=https://media.dinar.dz
API_INTERNAL_URL=http://api:8080
NEXT_INTL_DEFAULT_LOCALE=ar
NEXT_PUBLIC_WS_URL=
NOTIFICATIONS_WS_PATH=/api/v1/ws/notifications
```

The header notification center first loads `/api/v1/me/notifications`, then
requests a one-time WebSocket ticket and opens the same-origin notification
socket. It updates the unread count from `notification.created`, uses REST for
read/read-all actions, and reconnects with `1s, 2s, 4s, 8s` exponential backoff
with jitter. After reconnect it resynchronizes with `since_id`. Do not open a
WebSocket during SSR, and never put JWTs or tickets in its URL.

## 12. Testing checklist

- Unit-test formatters, countdowns, phone validation, and `next-intl` loading.
- Component-test offer cards, uploaders, filters, language switching, and all loading/empty/error states.
- Playwright-test public browsing, merchant onboarding, admin review, manual payment, `/ar`, `/fr`, RTL/LTR, and accessibility.
- Verify WebSocket reconnect, REST resync, SSR safety, JWT isolation, and that one user's notification cannot appear for another user.
- Run `tsc`, ESLint, component tests, E2E tests, and bundle-size checks in CI.
