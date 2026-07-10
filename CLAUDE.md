# CLAUDE.md

This repository follows the NADAA delivery plan in `agent_plan.md`.

## Core Context

- Product: Ghana National Disaster Alert and Response Platform.
- Slogan: Be Aware. Be Prepared. Be Safe.
- Priority hazard: flood risk and flood response.
- MVP: citizen risk checker, citizen reporting, authority dashboard, approved alerts, emergency guidance, shelters, and baseline flood risk scoring.

## Engineering Guidance

- Prefer the existing monorepo layout.
- Keep shared domain contracts in `packages/shared-types`.
- Keep brand constants in `packages/brand`.
- Keep public-safety decisions human-approved.
- Update docs when API, deployment, security, ML, or workflow behavior changes.

## Frontend Modularity Rules

Applies to the Vite web apps (`marketing-web`, `citizen-web`,
`authority-dashboard`, `dispatcher-web`, `admin-web`, `agency-web`). The Expo
apps are out of scope. Full detail lives in `docs/frontend-structure.md`.

- Keep each web app root `src/App.tsx` a thin entrypoint that re-exports the
  feature root through `src/features/<feature>/index.ts`.
- Put app-wide configuration, theme, and session helpers under `src/app/`.
- Put domain-specific data, types, utilities, api clients, and components under
  `src/features/<feature>/`; keep presentational components in
  `src/features/<feature>/components/` behind an `index.ts` barrel.
- Use the `@/*` -> `src/*` path alias (tsconfig `paths` + vite `resolve.alias`)
  for cross-layer imports such as `@/app/*`; keep intra-feature imports relative.
- Keep the global stylesheet at `src/styles/global.css`.
- Do not add new screens, API orchestration, fixtures, or large JSX surfaces
  directly to root app files.
- If a feature grows beyond one focused concern, split it into `data.ts`,
  `types.ts`, `utils.ts`, and focused `*.tsx` components (a leaf
  `components/shared.tsx` plus one file per panel) before adding more behavior.
