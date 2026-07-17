# Database

NADAA uses PostgreSQL with PostGIS for core relational and geospatial data.

## Files

- `migrations/001_core_geospatial_schema.sql` - core MVP schema, enums, PostGIS extension, tables, delivery logs, and indexes.
- `migrations/002_add_guide_sort_order.sql` - idempotent guide content ordering column for existing local databases.
- `migrations/003_add_road_closures.sql` - road closure reporting table with status/severity windows and detour notes.
- `migrations/004_add_relief_points.sql` - relief distribution points and stock history for inventory tracking.
- `migrations/005_add_aid_coordination.sql` - aid requests and donation pledges for recovery coordination.
- `migrations/006_add_missing_persons.sql` - missing person reports and authority audit entries for family reunification.
- `migrations/007_fix_service_alignment.sql` - notification enum extensions, alert workflow and incident urgency/abuse_score columns, and TEXT ID alignment with service-owned string IDs.
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

Apply migrations manually in filename order (all seven files):

```bash
for migration in database/migrations/*.sql; do
  psql "postgres://nadaa:nadaa_dev_password@localhost:5432/nadaa?sslmode=disable" \
    -f "$migration"
done
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
