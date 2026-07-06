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

