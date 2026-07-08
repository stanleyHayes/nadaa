# NADAA Design System

This document describes the shared visual system for NADAA web apps. It covers design tokens, the canonical MUI theme, accessibility expectations, and per-app conventions.

## Goals

- Keep all web apps visually consistent and on-brand.
- Reduce duplicated CSS and ad-hoc hex values.
- Meet WCAG 2.1 AA for color contrast, keyboard navigation, and screen-reader support.
- Stay mobile-first and resilient under poor connectivity.

## Design Tokens

Tokens live in `packages/brand/src/tokens.ts` and are mirrored as CSS custom properties in `packages/brand/src/brand.css`.

### Core palette

| Token           | Hex       | Meaning              |
| --------------- | --------- | -------------------- |
| `--nadaa-navy`  | `#0D1B3D` | Trust & authority    |
| `--nadaa-green` | `#118D4E` | Safety & growth      |
| `--nadaa-red`   | `#E53935` | Alert & urgency      |
| `--nadaa-gold`  | `#F4C20D` | Hope & optimism      |
| `--nadaa-slate` | `#555B66` | Stability & strength |
| `--nadaa-white` | `#FFFFFF` | Elevated surfaces    |
| `--nadaa-mist`  | `#F5F8FC` | Page background      |
| `--nadaa-ink`   | `#101828` | Primary text         |

### Semantic roles

| Token                      | Default   | Usage                        |
| -------------------------- | --------- | ---------------------------- |
| `--nadaa-surface`          | mist      | Page background              |
| `--nadaa-surface-elevated` | white     | Cards, sheets, dialogs       |
| `--nadaa-border`           | `#DFEAF2` | Dividers and borders         |
| `--nadaa-text-primary`     | ink       | Headings, body text          |
| `--nadaa-text-secondary`   | slate     | Captions, metadata           |
| `--nadaa-primary`          | navy      | Primary actions, topbars     |
| `--nadaa-secondary`        | green     | Success, operational accents |
| `--nadaa-accent`           | gold      | Public/marketing accents     |
| `--nadaa-danger`           | red       | Errors, severe alerts        |
| `--nadaa-warning`          | gold      | Warnings                     |
| `--nadaa-info`             | `#0B6FB8` | Information                  |

### Spacing

The spacing scale is based on 4 px increments:

| Token              | Value |
| ------------------ | ----- |
| `--nadaa-space-1`  | 4 px  |
| `--nadaa-space-2`  | 8 px  |
| `--nadaa-space-3`  | 12 px |
| `--nadaa-space-4`  | 16 px |
| `--nadaa-space-5`  | 20 px |
| `--nadaa-space-6`  | 24 px |
| `--nadaa-space-8`  | 32 px |
| `--nadaa-space-10` | 40 px |
| `--nadaa-space-12` | 48 px |
| `--nadaa-space-16` | 64 px |

### Elevation

| Token               | Shadow                               |
| ------------------- | ------------------------------------ |
| `--nadaa-shadow-sm` | `0 1px 2px rgba(13, 27, 61, 0.06)`   |
| `--nadaa-shadow-md` | `0 4px 12px rgba(13, 27, 61, 0.08)`  |
| `--nadaa-shadow-lg` | `0 8px 24px rgba(13, 27, 61, 0.10)`  |
| `--nadaa-shadow-xl` | `0 18px 48px rgba(13, 27, 61, 0.12)` |

### Border radius

| Token               | Value |
| ------------------- | ----- |
| `--nadaa-radius-sm` | 4 px  |
| `--nadaa-radius-md` | 8 px  |
| `--nadaa-radius-lg` | 12 px |
| `--nadaa-radius-xl` | 16 px |

## Canonical MUI Theme

MUI-based web apps must import the theme factory from `@nadaa/brand/theme`:

```tsx
import { createNadaaTheme } from "@nadaa/brand/theme";

const theme = createNadaaTheme({ accent: "operational" });
```

Options:

- `accent: "public"` — gold bottom accent for public-facing apps (`citizen-web`, `marketing-web`).
- `accent: "operational"` — green bottom accent for operational apps (`dispatcher-web`, `authority-dashboard`, `agency-web`, `admin-web`).
- `reducedMotion: true` — disables transitions (useful for users who prefer reduced motion).

The theme sets:

- Outfit font family.
- Navy primary, green secondary, red error, gold warning.
- Consistent heading weights and type scale.
- `borderRadius: 8`.
- Paper `backgroundImage: 'none'`.
- Button `minHeight: 42px`.
- Visible `:focus-visible` outline.
- `prefers-reduced-motion` support.

Do not create new local `theme.ts` files. If an app genuinely needs an override, document it in this file and keep it minimal.

## CSS Conventions

Every web app stylesheet should:

1. Import brand tokens first:

   ```css
   @import "@nadaa/brand/brand.css";
   ```

2. Set the topbar accent for the app:

   ```css
   :root {
     --nadaa-topbar-accent: var(--nadaa-green); /* or gold for public apps */
   }
   ```

3. Use token variables for colors, spacing, radius, and shadows instead of hard-coded values.
4. Include a `.skip-link` class.
5. Avoid `!important` unless overriding a third-party library.

## Severity & Hazard Color Roles

Use the role objects from `@nadaa/brand` so severity/hazard indicators are accessible and consistent.

```tsx
import { severityRoles, hazardRoles } from "@nadaa/brand";

<SeverityChip role={severityRoles.high} label="High" />
<HazardChip role={hazardRoles.flood} label="Flood" />
```

Each role provides `background`, `foreground`, and `border` colors that meet WCAG 2.1 AA contrast. Always pair color with an icon and visible text.

## Accessibility Checklist (WCAG 2.1 AA)

- [ ] Color is not the only way to convey information (add icons + text).
- [ ] Contrast ratios meet 4.5:1 for normal text and 3:1 for large text/UI components.
- [ ] All interactive elements have an accessible name (`aria-label` for icon-only controls).
- [ ] Form inputs have associated labels and `aria-describedby` linking to error text.
- [ ] Invalid fields use `aria-invalid="true"`.
- [ ] Focus indicators are visible (`:focus-visible` is configured in the theme).
- [ ] A skip link to `#main-content` is present.
- [ ] Semantic landmarks (`<header>`, `<main>`, `<nav>`, `<footer>`) are used.
- [ ] Data tables are horizontally scrollable on small screens with `tabIndex={0}` and an `aria-label`.
- [ ] Tap targets are at least 44×44 px.
- [ ] Motion respects `prefers-reduced-motion`.

## Per-App Accents

| App                        | Accent              | Theme import                                  |
| -------------------------- | ------------------- | --------------------------------------------- |
| `apps/citizen-web`         | Gold (public)       | `createNadaaTheme({ accent: "public" })`      |
| `apps/marketing-web`       | Gold (public)       | No MUI; import `@nadaa/brand/brand.css`       |
| `apps/dispatcher-web`      | Green (operational) | `createNadaaTheme({ accent: "operational" })` |
| `apps/authority-dashboard` | Green (operational) | `createNadaaTheme({ accent: "operational" })` |
| `apps/agency-web`          | Green (operational) | `createNadaaTheme({ accent: "operational" })` |
| `apps/admin-web`           | Green (operational) | `createNadaaTheme({ accent: "operational" })` |

## Mobile-First Rules

- Use MUI Grid `xs={12}` as the default, then scale up with `md`/`lg`.
- Keep horizontal overflow hidden; allow tables to scroll.
- Use fluid widths (`min()`, `clamp()`) for hero/marketing sections.
- Test at 360 px width.

## Future Work

- Build a shared `@nadaa/ui` package with common primitives (`PageShell`, `TopBar`, `MetricCard`, `EmptyState`, etc.) once the token system is proven across apps.
- Consider a full dark-mode theme.
- Add automated axe-core accessibility scanning to CI.
