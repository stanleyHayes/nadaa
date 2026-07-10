# open-data-service

The NADAA National Open Disaster Data Portal backend. Exposes approved, anonymized disaster datasets for public awareness, research, planning, and accountability.

## Responsibilities

- Catalog public datasets with metadata, license, update frequency, and privacy review status.
- Serve dataset detail with sample rows and column descriptions.
- Rate-limit and audit dataset downloads.
- Accept and manage access requests for restricted datasets.
- Provide admin-only request review endpoints.

## Environment variables

| Variable                    | Default                 | Description                        |
| --------------------------- | ----------------------- | ---------------------------------- |
| `PORT`                      | `:8102`                 | HTTP bind address                  |
| `AUDIT_LOG_SERVICE_URL`     | `http://localhost:8080` | Base URL of the audit log service  |
| `NADAA_ALLOWED_ORIGINS`     | `*`                     | Comma-separated CORS allowlist     |
| `RATE_LIMIT_REQUESTS`       | `10`                    | Requests allowed per IP per window |
| `RATE_LIMIT_WINDOW_SECONDS` | `60`                    | Rate limit window in seconds       |

## Routes

- `GET /health` — service health
- `GET /api/v1/open-data/datasets` — public catalog (optional `category`, `privacyReviewStatus` filters)
- `GET /api/v1/open-data/datasets/{id}` — dataset detail
- `GET /api/v1/open-data/datasets/{id}/download?format=csv|json|parquet` — download (rate-limited, audited)
- `POST /api/v1/open-data/requests` — request access to a restricted dataset
- `GET /api/v1/open-data/requests` — list requests (`system_admin`, `agency_admin`, or `nadmo_officer` role required)
- `POST /api/v1/open-data/requests/{id}/approve` — approve or reject a request (admin role required)

## Safety

- Datasets must have `privacyReviewStatus: approved` before download.
- Only aggregated or anonymized data is exposed in approved public datasets.
- Downloads are rate-limited per IP using a token bucket and logged as audit events.
- Requester emails are sanitized but not stored as PII beyond the access request.

## Development

```bash
cd services/open-data-service
go test ./...
go build ./cmd/server
./open-data-service
```
