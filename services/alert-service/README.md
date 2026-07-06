# Alert Service

The alert service owns public-warning drafts, geofenced targeting, submission, approval, rejection, emergency override, and alert audit events.

Current NADAA-050/NADAA-051 endpoints:

- `GET /healthz`
- `POST /api/v1/alerts`
- `GET /api/v1/alerts`
- `POST /api/v1/alerts/targets/preview`
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

## Geofenced Targets

Targets support `national`, `region`, `district`, `radius`, `community`, and `custom`.

- `region`, `district`, and `community` targets resolve against the starter catalog and return approximate geometry, area, and population metadata.
- `radius` targets require `center` and `radiusMeters`.
- `custom` targets require a closed polygon geometry.
- `POST /api/v1/alerts/targets/preview` validates and enriches a target before an alert is created.
- `GET /api/v1/alerts?targetType=district&targetId=accra-metropolitan` filters alerts by target.

## Run

```bash
go run .
```

The service listens on `:8089` by default. Override with `NADAA_ALERT_ADDR`.

## Test

```bash
go test ./...
```

Run live smoke checks after starting the service on `:8089`:

```bash
pnpm smoke:alert
pnpm smoke:alert-geofence
```

## Notes

The current implementation uses an in-memory store to lock in the workflow and target geometry contracts. PostGIS persistence, official district boundaries, delivery logs, and citizen alert feed behavior land in later stories.

Related stories:

- NADAA-050
- NADAA-051
- NADAA-052
