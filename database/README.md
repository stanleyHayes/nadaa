# Database

NADAA uses PostgreSQL with PostGIS for core relational and geospatial data.

## Files

- `migrations/001_core_geospatial_schema.sql` - core MVP schema, enums, PostGIS extension, tables, and indexes.
- `seeds/001_ghana_mvp_seed.sql` - development seed data for Ghana agencies, shelters, emergency guides, risk zones, alerts, incidents, and ML predictions.

## Local Database

Start local services:

```bash
docker compose -f infra/docker/docker-compose.yml up -d postgres
```

If port `5432` is already in use:

```bash
POSTGRES_PORT=55432 docker compose -f infra/docker/docker-compose.yml up -d postgres
```

Apply migrations manually:

```bash
psql "postgres://nadaa:nadaa_dev_password@localhost:5432/nadaa?sslmode=disable" \
  -f database/migrations/001_core_geospatial_schema.sql
```

Use the same custom port in the connection string when `POSTGRES_PORT` is set.

Apply seed data:

```bash
psql "postgres://nadaa:nadaa_dev_password@localhost:5432/nadaa?sslmode=disable" \
  -f database/seeds/001_ghana_mvp_seed.sql
```

Validate expected database assets are present:

```bash
pnpm validate:database
```

## Geometry Standard

- SRID: `4326`
- Point columns use `geometry(Point, 4326)`.
- Area columns use `geometry(Polygon, 4326)` or `geometry(MultiPolygon, 4326)`.
- Geometry columns should have GiST indexes.
