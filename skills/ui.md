---
name: dinar-dz-ui
description: Visual identity and UI design system for Dinar DZ, an Algerian discounts platform, including logo with Amazigh Yaz symbol, light/dark themes, icon rules, Arabic/French language switching, notifications, Arabic RTL typography, responsive mobile/desktop layouts, reusable components, offer cards, and ad slots. Use this skill whenever designing or styling any screen, component, logo, icon, theme token, notification, language switcher, or responsive layout for Dinar DZ.
---

# Dinar DZ — Visual Identity and Design System

Reference: [`../userstory.md`](../userstory.md) (behavior) and `frontend.md` (implementation). This file defines **how the product looks**.

## 1. Identity

- **Name:** Dinar DZ — "Dinar" in Arabic. Dinar is Algeria's currency and "DZ" anchors the national identity.
- **Promise:** "All your city's discounts in one place."
- **Personality:** Trustworthy, warm, Algerian, and energetic. The base is elegant (burgundy and green), while offers are energetic (orange-red and gold).
- **Voice:** Simple, clear Arabic with direct verbs and no exaggeration. Buttons include "Browse offers", "Contact on WhatsApp", and "Publish your offer". Errors explain what happened and what the user can do.

### 1.1 Visual direction inspired by the reference

Visual reference: [reference storefront](http://167.86.79.16:3033/en). We take only composition and rhythm principles; we do not copy its name, logo, text, or assets:

- A thin full-width announcement strip for seasonal messages or platform status.
- A quiet, slim header with few links, a clear centered desktop logo, and search/theme/account/language tools at the edges.
- An asymmetric two-column hero: calm copy area and a large image extending to the bottom of the viewport.
- Generous whitespace, thin borders, small radii, and almost no shadows; image and message are the focus.
- A large multi-line display headline, one short description, and only two clear buttons.
- A visual transition from hero to content groups through horizontal rules and editorial headings, not dense clusters of cards.

Apply the direction to Dinar DZ:

- Use copy such as **"All your city's discounts in one place."** with `Browse offers` and `Publish your offer` CTAs.
- Replace generic art with a collage or Algerian lifestyle image related to the merchant/offer, with a subtle burgundy or cream overlay when needed.
- The circular dinar and ⵣ logo remains central; never use the reference brand or imitate its exact style.
- Use dark burgundy for the top strip, cream for the hero, orange-red for discount badges, and gold only for verification and distinction.
- Arabic mirrors the column order and aligns text right; French uses LTR. Never mirror text inside images.
- The home page sells trust and catalogue: after the hero use `FeaturedOffers`, `EndingSoon`, `CategoryRail`, then a merchant CTA instead of a static product gallery.

#### Home page wireframe

```text
┌───────────────────────────────────────────────────────────────┐
│ announcement strip: New offers from Algerian shops            │
├───────────────┬───────────────────────────┬───────────────────┤
│ navigation     │        DINAR DZ           │ search · bell ·  │
│ offers · shops │      logo + ⵣ             │ lang · theme      │
├───────────────┴───────────────────────────┴───────────────────┤
│  cream editorial copy       │   full-bleed offer/shop image   │
│  All your city's discounts │   offer collage + subtle grain   │
│  [Browse offers] [Publish] │   [Featured]                     │
├───────────────────────────────────────────────────────────────┤
│  Latest offers       [cards with one clear discount accent]    │
│  Ending soon         [cards with countdown]                   │
│  Categories & areas  [horizontal category rail]               │
└───────────────────────────────────────────────────────────────┘
```

Important: never make the hero a decorative image without a usable offer, and do not crowd the header with tools. A visitor should reach an offer or search within two clicks.

## 2. Logo

**Concept:** a circular coin (the dinar) with the **Yaz ⵣ** character from Tifinagh at its center, representing "free people" in Amazigh culture.

**Construction:**
1. An outer circle with a coin-like toothed edge, using 24 small teeth.
2. A thin inner ring separates the edge from the disk.
3. The disk is **burgundy** `#6C151E`; the centered Yaz is **gold** `#D4A24C`.
4. The word **Dinar** uses a bold burgundy face, with **DZ** in dark green or a small badge.
5. The Arabic version uses heavy Cairo for "Dinar" with Latin DZ below or beside it.

**Required variants:**
| Variant | Use |
|---|---|
| Full horizontal (icon + Dinar DZ) | Site header and email |
| Vertical (icon above name) | Empty states and print |
| Icon only | favicon, PWA, profile image |
| Monochrome (cream on burgundy, burgundy on cream) | Colored backgrounds |

**Rules:**
- Safe space around the logo equals the height of the D in the wordmark.
- Minimum size: 24px icon and 96px horizontal width. At small sizes remove teeth and use a smooth circle.
- Do not rotate the logo, change approved colors, add shadows/gradients, or place it on a low-contrast background.
- Do not use a local currency abbreviation, "AED", or another country's currency symbol. The logo is a mark, not legal currency.
- Convert ⵣ to an **outline path** in the final asset rather than relying on a font. During development, **Noto Sans Tifinagh** is acceptable.

**Initial SVG (to be replaced by a designer later):**
```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96" role="img" aria-label="Dinar DZ">
  <circle cx="48" cy="48" r="46" fill="#6C151E"/>
  <circle cx="48" cy="48" r="38" fill="none" stroke="#D4A24C" stroke-width="2"/>
  <text x="48" y="64" text-anchor="middle" font-family="Noto Sans Tifinagh, sans-serif"
        font-size="46" fill="#D4A24C">ⵣ</text>
</svg>
```

## 3. Colors

### 3.1 Core colors
| Name | HEX | Role |
|---|---|---|
| Burgundy | `#6C151E` | **primary**: primary buttons, top strip, important links |
| Dark green | `#0F3D3A` | **secondary**: footer, categories, strong positive states |
| Cream | `#F5DABF` | featured cards, warm dividers, text on burgundy/green |
| Orange-red | `#E8452C` | **accent**: discount badges, countdown, "ending soon" |
| Gold | `#D4A24C` | logo, featured offer, verified merchant badge |
| Warm white | `#FFF9F2` | page background (`base-100`) |
| Text dark | `#2B1B1D` | primary text |

### 3.2 Usage rules
- **60/30/10:** 60% warm white and cream, 30% burgundy and green, 10% orange and gold. Do not flood the page with dark colors.
- **Orange-red is for discounts only** (percent badge and countdown). Other uses weaken its meaning.
- **Gold is for trust and distinction only:** verification, featured offer, and logo.
- Contrast (WCAG AA): normal text ≥ 4.5:1, large/bold text ≥ 3:1.
  - Cream on burgundy and green: excellent.
  - White on `#E8452C`: about 3.9:1, **use only for bold text at 14px or larger** (badges). For small orange text use darker `#B52E1A`.
  - `#3A2A0B` text on gold is acceptable.
- States: success `#2E7D5B`, warning = gold, error `#B3261E`, info `#2F6B8A`.

## 4. DaisyUI 5 theme (Tailwind 4)

Place it in `src/styles/globals.css`:

```css
@import "tailwindcss";
@plugin "daisyui";

@plugin "daisyui/theme" {
  name: "dinar";
  default: true;
  color-scheme: light;

  --color-base-100: #FFF9F2;
  --color-base-200: #FBEBDA;
  --color-base-300: #F5DABF;
  --color-base-content: #2B1B1D;

  --color-primary: #6C151E;
  --color-primary-content: #F5DABF;
  --color-secondary: #0F3D3A;
  --color-secondary-content: #F5DABF;
  --color-accent: #E8452C;
  --color-accent-content: #FFFFFF;
  --color-neutral: #2B1B1D;
  --color-neutral-content: #FFF9F2;

  --color-info: #2F6B8A;     --color-info-content: #FFFFFF;
  --color-success: #2E7D5B;  --color-success-content: #FFFFFF;
  --color-warning: #D4A24C;  --color-warning-content: #3A2A0B;
  --color-error: #B3261E;    --color-error-content: #FFFFFF;

  --radius-selector: 1rem;
  --radius-field: 0.625rem;
  --radius-box: 0.875rem;
  --size-selector: 0.25rem;
  --size-field: 0.25rem;
  --border: 1px;
  --depth: 0;
  --noise: 0;
}

@theme {
  --color-gold: #D4A24C;
  --color-cream: #F5DABF;
  --color-accent-700: #B52E1A;
  --font-sans: var(--font-tajawal), system-ui, sans-serif;
  --font-display: var(--font-cairo), var(--font-tajawal), sans-serif;
}

html { font-family: var(--font-sans); }
```

### 4.1 Light and dark modes

The site supports two official modes: **light** and **dark**. Light is the default for a new visitor; the theme button in the header switches modes. Store the preference in `localStorage` under `dinar_theme`, and use `prefers-color-scheme` when no preference exists.

Only `light` and `dark` are allowed. Components must use theme tokens rather than changing colors directly so contrast stays consistent.

```html
<html data-theme="dinar" class="theme-light">
<!-- or -->
<html data-theme="dinar-dark" class="theme-dark">
```

Suggested dark theme:

```css
@plugin "daisyui/theme" {
  name: "dinar-dark";
  prefersdark: true;
  color-scheme: dark;

  --color-base-100: #191416;
  --color-base-200: #251C1F;
  --color-base-300: #352629;
  --color-base-content: #FFF9F2;
  --color-primary: #A63A45;
  --color-primary-content: #FFF9F2;
  --color-secondary: #3D7B73;
  --color-secondary-content: #FFF9F2;
  --color-accent: #FF704D;
  --color-accent-content: #24100C;
  --color-neutral: #FFF9F2;
  --color-neutral-content: #191416;
  --color-info: #6FA8C2;
  --color-success: #69B58C;
  --color-warning: #E1B865;
  --color-error: #F27B72;
  --radius-selector: 1rem;
  --radius-field: 0.625rem;
  --radius-box: 0.875rem;
  --depth: 0;
  --noise: 0;
}
```

Theme rules:

- The toggle shows the current state (`light` or `dark`) and its `aria-label` describes the next state.
- Apply the theme before first paint to avoid a color flash; do not wait until after render to read `localStorage`.
- Images and logos use the variant with sufficient contrast; the logo shape does not change.
- In dark mode avoid pure black and strong shadows; borders and `base-*` tones separate cards.
- Test contrast in both modes for body text, large text, states, buttons, and alerts.

## 5. Typography and scale

- **Body:** Tajawal (400/500/700). **Headings and large numbers:** Cairo (700/800).
- Sizes (rem): `xs .75`, `sm .875`, `base 1`, `lg 1.125`, `xl 1.25`, `2xl 1.5`, `3xl 1.875`, `4xl 2.25`.
- Arabic line height: **1.7** for body and **1.35** for headings.
- Minimum body text is 15–16px on mobile; small notes are at least 12px.
- Do not use caps or letter spacing for Arabic. Only English Dinar DZ headings may have slight tracking.
- Prices use bold Cairo Latin numerals: `12 500 DZD`; old prices are struck through and muted.

## 6. Spacing, grid, and radii

- Use a 4px unit. Common gaps: 8, 12, 16, 24, 32, and 48.
- Container: `max-w-7xl`, side margins 16px mobile, 24px tablet, and 32px desktop.
- Grid: one column at 320–479px, two at ≥480px, three at ≥768px, and four at ≥1280px.
- Radii: cards `0.875rem`, fields `0.625rem`, badges `9999px`, card images `0.75rem`.
- Shadows are very light (DaisyUI depth = 0). Use 1px `base-300` borders rather than heavy shadows.
- Touch targets are at least 44px.

## 7. Components

### 7.1 Offer card `OfferCard` (primary component)
```
┌───────────────────────┐
│ [-40%]        [❤ Later]│  discount badge at the reading start (right in RTL)
│       4:3 image       │
│                       │
├───────────────────────┤
│ Original sneakers      │  title limited to two lines (line-clamp-2)
│ 4 500 DZD ~~7 500~~    │  new price bold primary, old price struck through
│ ◔ 3 days 4 hours left  │  countdown (orange when < 48 hours)
│ 📍 Oran · Noor Shop    │  wilaya and shop name
└───────────────────────┘
```
- Image ratio is `4/3` with `object-cover` and a `base-300` loading background.
- Badge: wide `badge badge-accent` in Cairo 800 at `top-2 start-2`.
- "Ending soon" (<48 hours): subtle pulsing dot and accent countdown.
- "Featured offer": 1.5px gold border and a small gold "Featured" badge.
- Without a price, enlarge the discount and hide the price row.
- States: default, hover (2px lift and primary border), visible focus, and loading skeleton.
- The whole card is one `<a>` to avoid nested interactive elements.

### 7.2 Header and navigation
- **Mobile:** simple header (logo + search icon + menu) and fixed `BottomNav`: home, categories, wilaya, account/publish. Hide `BottomNav` on offer details so it does not conflict with the WhatsApp bar.
- **Desktop:** `primary` header background, cream logo, large centered search, "Publish your offer" button (gold/cream), notification center, language switcher, theme toggle, and login/account link.
- Horizontally scrollable category rail below the header: icon + name, active category with an accent underline.
- Keep a visible wilaya selector ("Oran ▾") and save its choice in localStorage.

#### Header notification center `NotificationCenter`

- Show the notification bell in the desktop header and at the top of mobile beside the logo and search.
- A small badge shows unread count up to `99+`; never rely on color alone.
- Click opens a desktop popover or full-width mobile drawer.
- Each notification shows type icon, title, short text, relative time, and read/unread state.
- Actions: open destination, mark one read, and mark all read.
- The list is paginated or lazy-loaded with no more than 20 notifications per batch.
- Component states: skeleton loading, empty, retryable error, and a new notification arriving while closed.
- Never show sensitive documents, card numbers, or secrets; content is plain text, not HTML.

Suggested frontend contract:

```text
GET  /me/notifications?unread_only=false&page=1&per_page=20
POST /me/notifications/:id/read
POST /me/notifications/read-all
```

Protect these routes with the current user's session and return only that user's notifications. If the API is not ready, hide the list behind a feature flag and never use mock data in production.

#### Language switcher `LanguageSwitcher`

- Only Arabic `ar` and French `fr` are supported.
- Render flags as SVG or stable local assets: Algeria `DZ` for Arabic and France `FR` for French; do not rely on emoji because rendering varies by device.
- Show the language name with the flag, not the flag alone: `Arabic` / `Français`.
- Preserve the path and query string when switching, for example `/ar/offers?...` to `/fr/offers?...`.
- Arabic uses `dir="rtl"`; French uses `dir="ltr"`. Update `lang` and `dir` on the server before paint.
- Do not add languages without updating product scope and translation files.
- If a French translation is missing, show a clear fallback and do not mix text directions inside the component.
- Text comes from `messages/ar.json` and `messages/fr.json` with identical keys; do not add UI text directly inside a component.

### 7.3 Search and filters
- Place the main search row above the category rail. Keep it calm: search input with icon, sort combobox, filter trigger, and clear action; do not add a separate Search button.
- Search results update as the visitor types or changes a filter, with no submit step and no page reload.
- The filter trigger opens a responsive modal containing Wilaya, dependent Commune, consumer gender, and a dual-handle age range slider from 0 to 120.
- Selecting a Wilaya enables its Commune list; applying the modal updates local results without navigation or page reload.
- Category selection reveals the subcategory rail; selecting a subcategory reveals its brand rail.
- Brand names are always rendered using their French localized label, including in Arabic mode.
- Applied filters show a small active indicator on the filter trigger and remain removable through the clear action.
- Empty state: simple illustration + "No matching offers" + suggestions (change wilaya or remove a filter).

### 7.4 Details page
- Gallery: large 4:3 image with dots/count (3/10), thumbnails, and fullscreen zoom.
- Price block: large discount badge, new price at 28–32px, old price struck through.
- Clear countdown with units (day · hour · minute).
- Shop card: logo, name, gold "Verified merchant" badge, wilaya, and "All shop offers" link.
- **Fixed WhatsApp bar** (mobile): green `#25D366` button with contrast-safe text, "Contact on WhatsApp", and a secondary call button. WhatsApp green is allowed here as a recognizable brand exception.
- Small quiet "Report offer" button at the bottom.

### 7.5 Merchant dashboard
- Layout: sidebar on desktop or bottom tabs on mobile: overview, my offers, subscription, shop.
- **Status card** at the top, colored by state:
  - Pending review: light gold + clock icon.
  - Rejected: light red + reason + "Edit and resubmit" button.
  - Trial: light green + "N days left" + progress bar + "Subscribe" button.
  - Expired: orange + "Your offers are hidden" + "Renew now" button.
- Metric cards (active offers, views, WhatsApp clicks) use large Cairo numbers.
- Offer table on desktop or cards on mobile with draft, pending, published, rejected, and expired badges.
- Onboarding wizard: three-step progress, large fields, and one primary action per screen.

### 7.6 Admin dashboard
- Higher density: compact tables, top filters, and queue counters in the sidebar.
- Review screen: data in the first column, documents/preview in the second, and a fixed action bar: **Approve** (success) · **Reject** (error) · **Suspend**.
- Keep the same colors without decorative elements.

### 7.7 Core components (DaisyUI)
| Element | Class |
|---|---|
| Primary button | `btn btn-primary` |
| Secondary button | `btn btn-outline btn-primary` |
| Discount/energetic CTA | `btn btn-accent` (rarely, once per screen) |
| Tertiary button | `btn btn-ghost` |
| Field | `input input-bordered w-full`, label above |
| Select | `select select-bordered` |
| Badges | `badge badge-accent` / `badge-warning` / `badge-success` / `badge-outline` |
| Alert | `alert` + icon + text |
| Modal | `modal` with focus trap |
| Loading | `skeleton` for cards, `loading loading-spinner` for buttons |
| Tabs | `tabs tabs-border` |

Rule: one prominent `btn-primary` per area; keep other buttons quieter.

## 8. Icons and images
- Use `lucide-react` icons with 1.75–2px stroke, 20px by default and 24px for navigation.
- Do not use emoji icons in the official UI.
- Offer images are 4:3 for cards and 1:1 or 4:3 in galleries. Use `object-cover` with a centered focal point.
- Empty-state illustrations are simple burgundy and cream line art with subtle Amazigh geometry (diamonds and triangles) at 6–10% opacity in large backgrounds. Do not overuse it or place it behind text.

### 8.1 Icon options

| Option | Use | Rule |
|---|---|---|
| `lucide-react` outline | Buttons, navigation, fields, states, tables | Default; 1.75–2px stroke |
| Custom SVG | Dinar DZ logo, language flags, WhatsApp, brand states | Trusted local assets with suitable `aria-hidden` or `aria-label` |
| CSS shape | Divider or simple indicator | Decorative only; never for functional meaning |

Core icon dictionary:

| Meaning | Suggested icon |
|---|---|
| Search | `Search` |
| Home | `House` |
| Categories | `LayoutGrid` |
| Wilaya | `MapPin` |
| Account | `UserRound` |
| Notifications | `Bell` |
| Settings | `Settings` |
| Language | `Languages` with flag |
| Theme | `Sun` and `Moon` |
| Back/forward | `ArrowStart` and `ArrowEnd` with RTL-aware mapping |
| Verified | `CircleCheck` |
| Warning | `TriangleAlert` |
| Error | `CircleX` |
| Time/expiry | `Clock3` |
| Share | `Share2` |
| Image | `Image` |
| Upload | `Upload` |
| Delete | `Trash2` |
| Edit | `Pencil` |
| Close | `X` |

Icon rules:

- A standalone icon button must have an `aria-label`; an icon beside text is `aria-hidden="true"`.
- Do not use an icon alone when its meaning is unclear; add text or a desktop tooltip.
- Arrows follow direction: in RTL, `ArrowStart` points visually right and `ArrowEnd` left.
- Do not mix filled and outline styles in one group except to distinguish active state.
- Never communicate state through icon or color alone; show both text and icon.

## 9. Motion
- Use 150–200ms for interactions (hover and opening menus) with `ease-out`.
- Use one prominent motion only: a subtle pulse for the "Ending soon" badge. Do not animate every section into view.
- Respect `prefers-reduced-motion` by disabling pulses and large transitions.

## 10. Ad sizes
| Slot | Dimensions (px) | Devices |
|---|---|---|
| `home_top` | 728×90 (mobile: 320×100) | Home top |
| `list_inline` | 336×280 (grid-width card) | Every 8 offers |
| `detail_sidebar` | 300×250 | Desktop only; below shop card on mobile |
| `footer_banner` | 728×90 (mobile: 320×100) | Before footer |
- Small muted "Advertisement" label (12px) above or inside the corner.
- Reserve space with `aspect-ratio` before loading.
- Ads must not exceed 30% of the first mobile viewport height.

## 11. Accessibility
- Use the contrast rules in section 3.2 and run automated checks.
- Focus state: a 2px `accent` ring with 2px offset on every interactive element.
- Do not communicate meaning with color alone; states include text and icon.
- Text must scale to 200% without breaking layout.

## 12. Design checklist
- [ ] Theme colors come only from the table; no random HEX values.
- [ ] Text contrast is at least AA, especially text on orange.
- [ ] Arabic line height is appropriate and glyphs are not clipped.
- [ ] RTL works for badges, arrows, icons, and tables.
- [ ] Discount badge is the strongest card element; orange is not used elsewhere.
- [ ] Touch targets are at least 44px and interactive elements do not overlap.
- [ ] Loading, empty, error, rejected, and expired states are designed for every screen.
- [ ] Logo variant and size are correct with safe space.

Splash screen, dynamic identity, and the database proverb are documented in [`splash screen.md`](splash%20screen.md).

## 13. Website component catalogue

Every component has loading, empty, error, success, disabled, focus, and light/dark states where relevant. Proposed names are shared by design and frontend.

### 13.1 General components

| Component | Function |
|---|---|
| `AppShell` | Connects theme, locale, metadata, splash, header, content, and footer |
| `Logo` | Renders the appropriate Dinar DZ logo variant |
| `Header` | Logo, search, notifications, language, theme, account/login |
| `NotificationCenter` | Badge, popover/drawer, read actions, empty/error states |
| `LanguageSwitcher` | Arabic/French with flag and `lang/dir` updates |
| `ThemeToggle` | Switches light/dark and saves the preference |
| `SearchBar` | Global search with submit, clear, loading, and keyboard support |
| `WilayaSelector` | Searchable wilaya selection and persistence |
| `CategoryNav` | Scrollable categories and active state |
| `BottomNav` | Mobile navigation: home, categories, wilaya, account/publish |
| `Footer` | Terms, privacy, contact, language, and brand links |
| `Breadcrumbs` | Desktop page trail, shortened on mobile |
| `AdSlot` | Fixed-size ad space with label, lazy loading, and tracking |
| `SplashScreen` | Logo, database proverb, progress, hold, and fallback |

### 13.2 State and interaction components

| Component | Function |
|---|---|
| `Button` | Primary, secondary, accent, ghost, loading, disabled |
| `IconButton` | Icon button with label and visible focus |
| `FormField` | Label, hint, input, error, counter, `aria-describedby` |
| `Modal` | Esc support, focus trap, close button, mobile full/near-full |
| `Drawer` | Bottom drawer on mobile and side drawer on desktop |
| `ConfirmDialog` | Clear confirmation for delete/archive/reject |
| `Toast` | Short result; never the only place for an important decision |
| `Alert` | Info/success/warning/error with icon, text, and action |
| `Skeleton` | Placeholder that does not cause layout shift |
| `EmptyState` | Description, reason, and suggested action |
| `ErrorState` | Safe description, optional request ID, retry |
| `Pagination` | Current page, total, previous/next, keyboard support |
| `StatusBadge` | Label + icon for states, not color alone |
| `Countdown` | Remaining time with hydration-safe updates |

### 13.3 Visitor experience components

| Component | Function |
|---|---|
| `OfferCard` | Image, discount, title, price, countdown, shop/wilaya |
| `OfferGrid` | Responsive grid with loading/empty/error states |
| `FilterDrawer` | Mobile filters and `Show N results` action |
| `FilterChips` | Applied filters with removal |
| `OfferGallery` | Main image, thumbnails, zoom, touch gestures |
| `OfferDetails` | Description, prices, dates, shop, similar offers |
| `ShopCard` | Shop identity, verification, location, offers link |
| `ContactActions` | WhatsApp, phone, share, report |
| `ReportDialog` | Reason, message, validation, success/error |

### 13.4 Merchant components

| Component | Function |
|---|---|
| `MerchantLayout` | Desktop sidebar, mobile tabs, guard, breadcrumbs |
| `AccessStatusCard` | incomplete/pending/trialing/subscribed/expired |
| `OnboardingStepper` | personal/shop/documents/review steps |
| `MerchantProfileForm` | Personal data and autosave |
| `ShopForm` | Shop data, slug, contact, register |
| `DocumentUploader` | Three documents, progress, validation, privacy notice |
| `OfferEditor` | title, description, prices, dates, category, location |
| `ImageUploader` | Compression hint, preview, ordering, cover, delete |
| `MerchantOfferList` | Cards/table by viewport, filters, actions, statuses |
| `StatsCards` and `StatsChart` | Views/clicks, totals, 30-day series |
| `BillingPanel` | Plan, checkout, manual proof, pending/paid/failed |

### 13.5 Admin components

| Component | Function |
|---|---|
| `AdminLayout` | Navigation, queue counters, role guard |
| `MetricCard` | Indicator, value, period, trend, accessible label |
| `QueueTable` | Search, filters, pagination, bulk select, responsive fallback |
| `DocumentViewer` | Signed URL, image/PDF, expiry warning, audit notice |
| `ReviewPanel` | Preview, approve, reject reason, suspend/reactivate |
| `BulkActionBar` | Action, selection count, confirmation, partial failures |
| `CategoryEditor` | Hierarchy, order, visibility, icon/slug |
| `AdEditor` | Advertiser, media, slot, dates, targeting, priority |
| `ReportsTable` | Grouped reports, unique IP count, resolution actions |
| `SettingsForm` | Trial, pricing, limits, manual payment details |
| `AuditTable` | Actor, action, resource, time, IP, request ID, read-only |

## 14. Responsive mobile/desktop specification

### 14.1 Breakpoints

| Range | Name | Layout |
|---|---|---|
| 320–479px | compact mobile | 1-column content, compact header, bottom nav, bottom drawers |
| 480–767px | large mobile | 2-column offer grid when card minimum width allows, bottom nav |
| 768–1023px | tablet | 3-column grids where possible, collapsible side navigation |
| 1024–1279px | desktop | desktop header, side navigation for dashboard, 3-column content |
| 1280px+ | wide desktop | max-width container, 4-column offer grid, persistent sidebars where useful |

### 14.2 Mobile rules

- Design for 360px first; no horizontal page scroll.
- Header is compact; search opens as a full-width row or dedicated screen.
- `BottomNav` is fixed, with safe-area padding and no overlap with WhatsApp CTA.
- Tables become cards or horizontal-scroll regions only when the data cannot be simplified.
- Drawers open from the bottom, have a visible close action, and support swipe only as an enhancement.
- All actions have at least 44×44px touch area; never rely on hover.
- File upload supports camera/gallery where the browser allows it and clearly shows size/type errors.

### 14.3 Desktop rules

- Header includes logo, search, notifications, language, theme, and account actions.
- Merchant/admin pages use a logical side navigation; its visual side changes with RTL/LTR.
- Tables, bulk actions, document preview, and admin review can use two-column or three-column layouts.
- Hover may add context, but focus and click must provide the same function.
- Keep content inside `max-w-7xl`; do not stretch reading text across the full viewport.

### 14.4 Shared responsive rules

- Use CSS logical properties: `ms`, `me`, `ps`, `pe`, `start`, `end`; avoid `ml/mr/left/right` for layout.
- Test Arabic RTL and French LTR separately; do not assume an icon or arrow has the same visual direction.
- Test light/dark in every breakpoint; text and borders must remain AA-compliant.
- Preserve keyboard order and focus when a desktop modal becomes a mobile drawer.

### 14.5 Responsive acceptance matrix

| Surface | Mobile 360px | Tablet 768px | Desktop 1280px |
|---|:---:|:---:|:---:|
| Public header/search | ✅ | ✅ | ✅ |
| Offer grid/detail | ✅ | ✅ | ✅ |
| Merchant onboarding | ✅ | ✅ | ✅ |
| Image uploader | ✅ | ✅ | ✅ |
| Admin queue/review | cards/drawer | compact table | full table + preview |
| Notifications | drawer | popover/drawer | popover |
| Language/theme controls | compact menu | inline/menu | inline |

## 15. Extended design checklist

- [ ] Every component has documented content, loading, empty, error, disabled, focus, light, and dark states where relevant.
- [ ] Header contains notifications, theme toggle, and only Arabic/French language switch with accessible flags.
- [ ] No emoji is used as a functional icon; use Lucide or approved SVG assets.
- [ ] Mobile 360px, tablet 768px, desktop 1280px, RTL, LTR, light, dark, keyboard, and reduced motion are tested.
- [ ] Tables and drawers have intentional responsive behavior, not accidental overflow.
- [ ] All icon-only controls have accessible labels.
- [ ] The frontend API contract for notifications is added to the backend before enabling the production UI.
