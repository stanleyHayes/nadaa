# open-data-service

The NADAA National Open Disaster Data Portal backend. Exposes approved, anonymized disaster datasets for public awareness, research, planning, and accountability.

## Responsibilities

- Catalog public datasets with metadata, license, update frequency, and privacy review status.
- Serve dataset detail with sample rows and column descriptions for approved datasets.
- Rate-limit and audit dataset downloads (audit events persisted locally, forwarded best-effort).
- Accept and manage access requests for restricted datasets.
- Provide admin-only request review and audit-trail endpoints.

## Environment variables

| Variable                     | Default                 | Description                                                       |
| ---------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `PORT`                       | `:8102`                 | HTTP bind address                                                 |
| `AUDIT_LOG_SERVICE_URL`      | `http://localhost:8080` | Base URL of the audit log service (forwarding is best-effort)     |
| `NADAA_INTERNAL_SERVICE_TOKEN` | _(empty)_             | Service token sent as `X-NADAA-Service-Token` on audit forwarding  |
| `NADAA_ALLOWED_ORIGINS`      | `*`                     | Comma-separated CORS allowlist                                    |
| `RATE_LIMIT_REQUESTS`        | `10`                    | Requests allowed per IP per window                                |
| `RATE_LIMIT_WINDOW_SECONDS`  | `60`                    | Rate limit window in seconds                                      |
| `NADAA_AUTH_TOKEN_SECRET`    | _(empty)_               | HMAC secret verifying auth-service bearer tokens for admin routes |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS` | `false`               | Honor legacy `X-NADAA-Actor-*` headers (local dev/smoke only)     |
| `NADAA_TRUST_PROXY_HEADERS`  | `false`                 | Honor `X-Forwarded-For`/`X-Real-Ip` for client IP rate limiting   |
| `NADAA_ENV`                  | _(empty)_               | `development` allows localhost CORS origins alongside the allowlist |

## Routes

- `GET /health` — service health
- `GET /api/v1/open-data/datasets` — catalog (optional `category` filter; anonymous callers see approved datasets only, verified admins may also filter by `privacyReviewStatus`)
- `GET /api/v1/open-data/datasets/{id}` — dataset detail (non-approved datasets are visible to verified admins only, with sample rows and columns stripped)
- `GET /api/v1/open-data/datasets/{id}/download?format=csv|json` — download the dataset's actual rows as a file (`Content-Disposition: attachment`); rate-limited and audited. The download record id, rate limit state, and local-audit outcome are reported via `X-NADAA-Download-Id`, `X-RateLimit-*`, and `X-NADAA-Audit-Logged` response headers.
- `POST /api/v1/open-data/requests` — request access to a restricted dataset (rate-limited)
- `GET /api/v1/open-data/requests` — list requests (admin: verified bearer token with `system_admin`, `agency_admin`, or `nadmo_officer` role and MFA)
- `POST /api/v1/open-data/requests/{id}/approve` — approve or reject a request (admin; `reviewedBy` is set from the verified token subject)
- `GET /api/v1/open-data/audit` — list locally persisted download-audit events (admin)

## Safety

- Datasets must have `privacyReviewStatus: approved` before download.
- Anonymous catalog and detail reads expose approved datasets only; sample rows and column layout of non-approved datasets are never served.
- Only aggregated or anonymized data is exposed in approved public datasets.
- Downloads and access-request creation are rate-limited per client IP using a token bucket; expired buckets are evicted.
- Download audit events are persisted to a local audit list (queryable by admins) and forwarded best-effort to the audit log service, authenticated with `NADAA_INTERNAL_SERVICE_TOKEN` (`X-NADAA-Service-Token`); forwarding failures are logged, never claimed as success.
- Admin routes require a verified auth-service bearer token (`typ=agency`, MFA completed, allowed role); legacy actor headers work only with `NADAA_AUTH_ALLOW_MOCK_ACTORS=true`.
- Every review decision (approve/reject) is audit-logged with the verified admin actor, and a request that has already been decided cannot be re-reviewed (`409 Conflict`).
- Requester emails are sanitized but not stored as PII beyond the access request.

## Development

```bash
cd services/open-data-service
go test ./...
go build ./cmd/server
./open-data-service
```
