# Deployment

## Local Development

Start local data services:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

Install dependencies:

```bash
pnpm install
```

Run apps:

```bash
pnpm dev:citizen
pnpm dev:authority
```

Run risk service:

```bash
cd services/risk-service
go run .
```

## Future Environments

- Local: Docker Compose, mock providers, local PostGIS, Redis, and MinIO.
- Staging: production-like services, test notification providers, seeded Ghana geospatial fixtures.
- Production: managed PostGIS, object storage, Redis/queue, observability, backup/restore, and strict secret handling.

