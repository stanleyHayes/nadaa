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

## Safety Rules

- ML predictions must not automatically send public alerts.
- Mass alerts require authority approval and audit logging.
- Authority users require role-based access and MFA.
- Anonymous citizen reports must preserve privacy unless policy and authorization permit disclosure.
- Life-threatening reports must never be hidden solely by an automated spam or suspicion score.

