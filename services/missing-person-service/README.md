# NADAA Missing Person Service

Privacy-sensitive missing persons workflow for disaster response and family reunification.

## Endpoints

- `GET /healthz` - service health.
- `POST /api/v1/missing-persons` - public intake. Creates a private `pending_review` record.
- `GET /api/v1/missing-persons` - public approved search/list.
- `GET /api/v1/missing-persons/{id}` - public approved record lookup.
- `GET /api/v1/authority/missing-persons` - authority full sensitive queue.
- `GET /api/v1/authority/missing-persons/{id}` - authority full sensitive record.
- `PATCH /api/v1/authority/missing-persons/{id}/review` - approve public, approve private, or reject.
- `PATCH /api/v1/authority/missing-persons/{id}/close` - close, reunite, mark located, withdraw, duplicate, or deceased.
- `GET /api/v1/authority/missing-persons/{id}/audit` - authority audit history.

## Authority Authentication

Authority endpoints require a verified `Authorization: Bearer nadaa.<payload>.<sig>`
token issued by auth-service (validated with `NADAA_AUTH_TOKEN_SECRET`). The verified
claims supply the actor id (`sub`), role, agency id, district, and MFA flag.

For local development and smoke tests, set `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` to honor
legacy self-asserted headers instead:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

Without a valid token (and without mock actors enabled), authority endpoints return `401`.
Public endpoints (`POST/GET /api/v1/missing-persons`) stay public. Approving public
visibility for a record whose reporter declined `consentToPublicShare` requires an
explicit `"consentOverride": true` in the review payload, which is recorded in the
audit trail. The localhost CORS exception (when `NADAA_ALLOWED_ORIGINS` is set) only
applies with `NADAA_ENV=development`.

## Local Development

```bash
cd services/missing-person-service
go test ./...
go run ./cmd/server
```

The service listens on `:8101` by default. Set `PORT=:18101` to override.

Run the smoke flow from the repository root:

```bash
pnpm smoke:missing-person
```

## Privacy Notes

Public records never expose reporter contact details, authority review notes, or closure notes. New reports stay private until an authority review explicitly approves public visibility. Closure and reunification move records back to private visibility and write an audit entry.
