# CLAUDE.md

This repository follows the delivery plan in `agent_plan.md`, now moving toward
the product direction in `SUBVENIO_Product_Blueprint_v2.docx` and the identity
in `SUBVENIO_Brand_Motion_Implementation_Guide.docx`.

## Core Context

- Product: SUBVENIO, Ghana's emergency alert and response platform.
- Tagline: Someone is coming.
- Priority hazard: flood risk and flood response.
- MVP: citizen risk checker, citizen reporting, authority dashboard, approved alerts, emergency guidance, shelters, and baseline flood risk scoring.
- In progress: the blueprint's personal-safety modules — Beacon, Circle, Halo,
  Watch, Vault — are not built yet. Incident, Alert, Dispatch and Responder are.

## Brand

- SUBVENIO is a WORKING identity, not a cleared trademark. Do not treat the
  rebrand as final until clearance completes.
- Everything user-facing resolves through `ACTIVE_BRAND` in
  `packages/brand/src/identity.ts`. Changing brands is that one line plus the
  `[data-brand]` block in `brand.css`.
- `X-NADAA-*` headers, `NADAA_*` env vars, the `@nadaa/*` npm scope and the
  `--nadaa-*` CSS properties are wire and internal identifiers. They do NOT
  move with the brand; renaming them breaks service-to-service auth.
- Go services read the name from `SUBVENIO_BRAND_NAME` (they cannot import the
  TypeScript package); keep its default in step with `ACTIVE_BRAND`.

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
