# AGENTS.md

## Mission

Build NADAA as a public-safety platform for Ghana. Keep implementation aligned with `spec.md` and the delivery workflow in `agent_plan.md`.

## Agent Workflow

1. Read `agent_plan.md` before starting work.
2. Claim the relevant story in the Active Work Board.
3. Keep changes scoped to the claimed story.
4. Update the Master Story Tracker and Plan Ledger when work moves status.
5. Do not commit secrets or real client/agency credentials.

## Branch And Commit Standards

- Branch: `feature/NADAA-123-short-name`
- Commit: `NADAA-123 implement short name`
- PR: `NADAA-123 Short Name`

## Go Service Structure

Each Go service under `services/` is organized as a multi-package Go module:

```
services/<service>/
├── cmd/server/main.go       # dependency wiring and ListenAndServe
└── internal/
    ├── config/              # env-based configuration
    ├── models/              # exported request/response/record structs
    ├── store/               # Store interface + in-memory implementation + seed data
    ├── utils/               # JSON, CORS, security headers, validation, env helpers
    └── handlers/            # HTTP server, routes, middleware, resource handlers, tests
```

- Keep `cmd/server/main.go` small: only load config, create the store, build the handler server, and start the HTTP listener.
- Put domain types in `internal/models` and export them.
- Define a `Store` interface in `internal/store` so handlers stay decoupled from storage details.
- Keep shared helpers (JSON encoding, CORS, env parsing, validation, coordinates) in `internal/utils`.
- Place handlers, middleware, and route registration in `internal/handlers`; split files by domain resource.
- Move service tests into `internal/handlers/*_test.go` under `package handlers` so they can call unexported server methods.
- Authority endpoints verify `Authorization: Bearer nadaa.<payload>.<sig>` tokens (HMAC-SHA256, shared `NADAA_AUTH_TOKEN_SECRET`) and build the actor context from verified claims (`internal/handlers/auth.go`). Self-asserted `X-NADAA-Actor-*` headers are only honored when `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` (local dev/smoke). Service-to-service calls use `X-NADAA-Service-Token` (`NADAA_INTERNAL_SERVICE_TOKEN`) where supported.
- Preserve env vars, defaults, routes, CORS behavior, security headers, and observable behavior when refactoring.

## Safety Rules

- ML predictions must not automatically send public alerts.
- Mass alerts require authority approval and audit logging.
- Authority users require role-based access and MFA.
- Anonymous citizen reports must preserve privacy unless policy and authorization permit disclosure.
- Life-threatening reports must never be hidden solely by an automated spam or suspicion score.
