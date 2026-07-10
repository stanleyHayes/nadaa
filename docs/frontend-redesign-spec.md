# NADAA Frontend Redesign Spec — "Navy Command"

Synthesised from UI/UX studies of two reference apps (AURA — Next + Tailwind +
shadcn institutional booking console; UPOSA — React + Tailwind + DaisyUI +
framer-motion) and re-expressed for NADAA's stack (React + MUI + Emotion +
Leaflet + lucide-react, `@nadaa/brand` `createNadaaTheme`). This is the
consistency contract for the marketing site and all operational dashboards.

## Design direction

One **"Navy Command"** identity, the way maroon anchors AURA and navy anchors
UPOSA:

- **Navy `#0D1B3D`** is the sole brand anchor (sidebar rail, hero bands, topbar).
- **Green `#118D4E`** = operational accent (dashboards, active/live states).
- **Gold `#F4C20D`** = public/citizen accent + hairline top-strip.
- **Red `#E53935`** = alert/severe only, via the AA-safe `severityRoles`/`hazardRoles`.
- Keep NADAA's **rounded 8–12px** MUI surfaces (do **not** adopt UPOSA's
  zero-radius) — a calm, legible emergency UI reads better rounded. Only the
  small 4px button radius is "square".
- Tokens-not-hex: every colour is a `var(--nadaa-*)` or a `color-mix()` over one.
- Recurring **icon-chip** motif (brand mark, nav-group chips, page-header plaque,
  stat-card chips) + a gold hairline on sidebar/hero/menus, so citizen marketing
  and operational dashboards read as one system. Severity is proportioned by
  colour: navy = authority, green = operational, gold = public hope, red = urgency.

## Shared dashboard shell (authority / dispatcher / agency / admin)

A single shell, first shipped in `apps/authority-dashboard`
(`features/command-center/`), reused across the operational apps:

- **Sign-in + MFA gate** (`app/session.ts` live store + `SignInScreen`), session
  persisted to localStorage; RBAC/MFA gating becomes an in-shell authorization state.
- **Collapsible navy sidebar** (`Sidebar`): grouped nav, lucide icons, active
  `aria-current` + gold edge marker, collapse persisted, live count badges,
  mobile MUI Drawer.
- **Glass topbar** (`Topbar`): brand, current-view title, notifications bell,
  user menu (name/role/agency + sign out).
- **Multi-view layout** (`CommandCenterShell` + `navigation.ts` + `views/*`): the
  sidebar switches `activeView`; an **Overview** landing view leads with
  tone-driven **StatCards** (tabular-nums) and, for authority, a severity-triage board.
- Green operational accent; per-section `PageHeader`; Leaflet re-skinned to tokens.
- Every public-safety alert stays **human-approved**.

## Marketing site (multi-page, shipped)

`apps/marketing-web` runs `BrowserRouter` + `SiteLayout` over Home / Platforms /
HowItWorks / Trust / Signup / Contact. Home: parallax navy hero, gilded stats
band (`AnimatedCounter`), `Reveal variant="3d"` cards on a stagger grid, response
loop, role teasers, navy trust band, gradient CTA. Signup = citizen registration
(name / +233 phone / region / language / consent) with inline validation and a
success state. Enrichments still open: `.stagger` grid timing, a View-Transitions
circular theme toggle, and wiring signup to the citizen/notification endpoint.

## Motion system (reduced-motion-aware, CSS + tiny hooks)

Proven in marketing-web; promote to a shared set and reuse everywhere. No heavy
animation library.

- **Tokens:** `--nadaa-ease-out-quart: cubic-bezier(.22,.7,.2,1)`,
  `--nadaa-ease-overshoot: cubic-bezier(.34,1.56,.64,1)`,
  `--nadaa-dur-reveal: .6s`, `--nadaa-dur-micro: .15s`.
- **Utilities:** `.reveal` + `.reveal--up|--3d|--scale` (toggled `.is-visible` by
  IntersectionObserver), `.stagger > *` nth-child delay (55ms; `.stagger-fast` 30ms),
  `.card-lift:hover`, `.btn-press:active`, `.count-up` (tabular-nums), `.pulse-ring`.
- **Hooks/components:** `useInView`, `useParallax`, `useCountUp`, `<Reveal>`,
  `<AnimatedCounter>` — all no-op/jump-to-final under reduced motion. Optional
  `useViewTransitionTheme()` for a circular light/dark reveal.
- Two reduced-motion layers already exist: the `brand.css` global reset and
  `createNadaaTheme({reducedMotion})` collapsing `transitions.duration.*` to 0.
- Dashboards use the same CSS: metrics via `<AnimatedCounter>`, panels wrapped in
  `<Reveal>`, table rows via `.stagger` `fadeInUp`, buttons `.btn-press`, loading
  via a 3-dot wave with the label held behind (`aria-busy`) so width never shifts.

## Per-app notes

- **citizen-web** — public and mostly anonymous, so keep it **lighter than the
  ops shell**: a top-nav (marketing's segmented pill), `accent: "public"` (gold),
  panels wrapped in `<Reveal>` with a gold-plaque `PageHeader`, a prominent
  "Check your risk" hero + a 112 emergency band, and a token-themed Leaflet map
  (navy tiles, severity-coloured markers). Optional light sign-in unlocks saved
  reports/claims; unauthenticated stays fully usable.
- **authority-dashboard** — flagship; shell shipped. Groups: Overview, Operations,
  Alerts, Resources, Field intel, Community.
- **dispatcher-web** — same shell, dispatch-tuned: Overview, Triage (AI triage as
  the hero, "suggestion, human decides"), Incidents, Alerts, Hospitals, Relief;
  ⌘K jump-to-incident; per-row table motion.
- **agency-web** — convert existing tabs to sidebar groups: Overview, Incidents,
  Hospitals, Shelters, Aid requests, Relief; export actions in a topbar action slot.
- **admin-web** — closest to AURA's admin console. Groups: Governance, Security,
  Config; showcase the rich user menu, notifications bell, settings-card, and
  shimmer skeletons on audit views.
