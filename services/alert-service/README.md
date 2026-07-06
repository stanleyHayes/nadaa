# Alert Service

The alert service owns public-warning drafts, submission, approval, rejection, emergency override, and alert audit events.

Current NADAA-050 endpoints:

- `GET /healthz`
- `POST /api/v1/alerts`
- `GET /api/v1/alerts`
- `PATCH /api/v1/alerts/{id}`
- `POST /api/v1/alerts/{id}/submit`
- `POST /api/v1/alerts/{id}/approve`
- `POST /api/v1/alerts/{id}/reject`
- `POST /api/v1/alerts/{id}/emergency-override`
- `GET /api/v1/alerts/audit`

## Authority Headers

Write and audit endpoints require authority context headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Draft/update/submit actions allow `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.

Approve/reject actions allow `system_admin`, `agency_admin`, and `nadmo_officer`. Non-system approvers cannot approve their own draft.

Emergency override allows only `system_admin` and `nadmo_officer`, requires a reason, marks the alert approved, and writes an audit event.

## Run

```bash
go run .
```

The service listens on `:8089` by default. Override with `NADAA_ALERT_ADDR`.

## Test

```bash
go test ./...
```

## Notes

The current implementation uses an in-memory store to lock in the workflow contract. PostGIS persistence, geofenced target geometry, delivery logs, and citizen alert feed behavior are planned in later stories.

Related stories:

- NADAA-050
- NADAA-051
- NADAA-052
