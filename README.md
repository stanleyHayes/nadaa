# NADAA

NADAA is the Ghana National Disaster Alert and Response Platform.

Slogan: **Be Aware. Be Prepared. Be Safe.**

The platform helps citizens, NADMO, district assemblies, dispatchers, and response agencies prepare for, report, monitor, respond to, and recover from disasters. The first implementation phase focuses on flood risk, citizen reporting, authority incident command, approved alerts, emergency guidance, and shelter visibility.

## Repository Layout

```text
apps/
  citizen-web/            Citizen PWA for alerts, risk checks, reports, guides, and shelters
  authority-dashboard/    Agency dashboard for incident command, alerts, assignments, and maps
services/
  auth-service/
  incident-service/
  alert-service/
  risk-service/
  dispatch-service/
  notification-service/
  integration-service/
  ml-service/
packages/
  brand/                  NADAA colors, slogan, feature pillars, and brand constants
  shared-types/           Shared TypeScript domain contracts
  config/                 Shared tool configuration
infra/
  docker/
  kubernetes/
  terraform/
docs/
```

## Getting Started

Install dependencies:

```bash
pnpm install
```

Run the citizen web app:

```bash
pnpm dev:citizen
```

Run the authority dashboard:

```bash
pnpm dev:authority
```

Run both apps:

```bash
pnpm dev
```

Run the Go risk service:

```bash
cd services/risk-service
go run .
```

Run the Go auth service:

```bash
cd services/auth-service
NADAA_AUTH_MOCK_OTP=123456 NADAA_AUTH_EXPOSE_DEV_OTP=true go run .
```

## Project Coordination

Use `agent_plan.md` as the living project board. Before starting work, agents should claim a row in the Active Work Board, update the Master Story Tracker, and record handoff notes when finished or blocked.

## Documentation

- [Product Scope](docs/product.md)
- [Architecture](docs/architecture.md)
- [API](docs/api.md)
- [Security](docs/security.md)
- [ML](docs/ml.md)
- [Deployment](docs/deployment.md)
- [QA Strategy](docs/qa.md)
- [Database](database/README.md)
- [Project Dashboard Contract](docs/project-dashboard/README.md)

## Source Documents

- `spec.md`
- `AI_Native_Software_Engineering_Operations_Manual.docx`
- `AI_Development_Workflow_Training_Manual.docx`
