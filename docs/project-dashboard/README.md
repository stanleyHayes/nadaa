# Project Dashboard Contract

The internal project dashboard should synchronize project status across Jira, GitHub, agent handoffs, and the living `agent_plan.md`.

This contract is intentionally simple for Sprint 0. It gives agents and future automation a stable shape for progress records.

## Files

- `contract.schema.json` - JSON Schema for project dashboard records.
- `sample-records.json` - sample MVP records that match the schema.

## Required Dashboard Fields

- Client.
- Project.
- Epic.
- Story.
- Jira key.
- Branch name.
- Pull request.
- Current status.
- Assigned team member.
- Estimated effort.
- Actual effort.
- Progress percentage.
- Last updated.

## Status Values

- `Todo`
- `In Progress`
- `Blocked`
- `Review`
- `Done`

The dashboard can map these to Jira workflow states later, but the multi-agent board should stay simple and fast to update.

## Agent Usage

1. Claim a story in `agent_plan.md`.
2. Update the dashboard record owner/status/branch.
3. Add blocker notes when blocked.
4. Add verification evidence when moving to `Review` or `Done`.
5. Keep actual effort current when the story closes.

## Validation

Run:

```bash
pnpm validate:dashboard
pnpm validate:features
```

This validates that the sample records contain the required Sprint 0 fields and status values, and that the flood-risk feature artifacts match their schema and manifest.
