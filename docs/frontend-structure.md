# Frontend Structure

This document defines the standard `src/` layout and conventions for the NADAA
Vite + React + TypeScript + MUI web apps:

- `apps/marketing-web`
- `apps/citizen-web`
- `apps/authority-dashboard`
- `apps/dispatcher-web`
- `apps/admin-web`
- `apps/agency-web`

It does **not** apply to the Expo / React Native apps (`apps/citizen-mobile`,
`apps/dispatcher-mobile`), which follow their own conventions.

## Standard layout

```
src/
  main.tsx              # entry: mounts <App/>, imports fonts + global styles
  App.tsx               # thin entry that re-exports the feature root
  vite-env.d.ts
  styles/
    global.css          # global stylesheet
  app/                  # app-wide composition (cross-feature, app singletons)
    config.ts           # runtime/env configuration (import.meta.env, API bases)
    theme.ts            # MUI theme (omitted for apps without MUI)
    session.ts          # session / auth headers helpers (where applicable)
  features/<feature>/   # one folder per domain feature
    <Feature>App.tsx    # feature root (owns layout + composition)
    components/         # feature-scoped presentational components
      shared.tsx        # primitives, chips, maps, small shared helpers
      <Panel>.tsx       # one file per focused panel / component group
      index.ts          # barrel re-exporting the feature's components
    data.ts             # fixtures / seed data (where present)
    types.ts            # feature-local types (where present)
    utils.ts            # feature-local pure helpers (where present)
    api.ts              # feature API client (where present)
    index.ts            # feature barrel: `export { default } from "./<Feature>App"`
  components/           # shared cross-feature presentational components (only if needed)
  hooks/                # shared hooks (only if needed)
  lib/                  # shared pure helpers / clients (only if needed)
  types/                # shared local types (only if needed)
```

Only create the shared `components/`, `hooks/`, `lib/`, and `types/` folders when
something is genuinely shared across features. Domain contracts stay in
`@nadaa/shared-types`; brand constants stay in `@nadaa/brand`.

## Path alias

Every app defines a single alias `@/*` -> `src/*`:

- `tsconfig.json` — `compilerOptions.baseUrl: "."` and
  `compilerOptions.paths: { "@/*": ["src/*"] }`.
- `vite.config.ts` — `resolve.alias["@"] = fileURLToPath(new URL("./src", import.meta.url))`.

Import conventions:

- Use the `@/` alias for cross-layer imports, e.g. a feature importing
  `@/app/config`, `@/app/theme`, `@/app/session`, `@/lib/*`, or `@/components/*`.
- Use **relative** imports within a feature (`./data`, `./types`, `./utils`,
  `./components`) and within a folder (`./shared`).

## Entry and barrels

- `src/App.tsx` stays a thin entry: `export { default } from "./features/<feature>";`
  (no JSX, no logic). This satisfies the "root `App.tsx` is a thin entrypoint"
  rule.
- Each feature exposes a default via `features/<feature>/index.ts`.
- Each feature's `components/index.ts` barrel re-exports the feature's components
  so the feature root imports them from a single `./components` path.
- The MUI `ThemeProvider` is applied by the feature root (or, for `agency-web`,
  in `main.tsx`); the theme itself always lives in `app/theme.ts`.

## Splitting oversized files

- A feature that grows past one focused concern is split into `data.ts`,
  `types.ts`, `utils.ts`, and focused `*.tsx` components before adding behavior.
- Large grab-bag component files are split into `components/shared.tsx` (leaf
  primitives, chips, maps, and the private helpers they use) plus one file per
  focused panel or cohesive component group. Keep the dependency direction
  acyclic: `shared` is a leaf that panels import from; panels never import each
  other in a cycle.

## Per-app feature names

| App                 | Feature folder              | Feature root           |
| ------------------- | --------------------------- | ---------------------- |
| marketing-web       | `features/marketing`        | `MarketingApp`         |
| citizen-web         | `features/citizen`          | `CitizenApp`           |
| authority-dashboard | `features/command-center`   | `CommandCenterApp`     |
| dispatcher-web      | `features/dispatch-command` | `DispatcherCommandApp` |
| admin-web           | `features/administration`   | `AdminConsoleApp`      |
| agency-web          | `features/agency`           | `AgencyApp`            |

## Verification

Each app must pass, from the repo root:

```
pnpm --filter @nadaa/<app> typecheck
pnpm --filter @nadaa/<app> lint
pnpm --filter @nadaa/<app> build
```

and changed files must satisfy `pnpm exec prettier --check`.
