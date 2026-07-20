# Route Service

Evacuation route planner for NADAA. Computes decision-support routes from an
origin to a destination, shelter, or higher-ground waypoint while attempting to
avoid reported road closures and severe risk zones.

## Local run

```bash
cd services/route-service
go run ./cmd/server
```

The service listens on `:8096` by default. Override with `PORT`.

## Endpoints

- `GET /health`
- `GET /routes/options` – returns supported `waypointType` values
- `POST /routes/plan` – plan an evacuation route

## Environment variables

| Variable                     | Default                 | Description                                   |
| ---------------------------- | ----------------------- | --------------------------------------------- |
| `PORT`                       | `8096`                  | HTTP listen address                           |
| `ROAD_CLOSURE_SERVICE_URL`   | `http://localhost:8095` | Base URL for road-closure-service             |
| `SHELTER_SERVICE_URL`        | `http://localhost:8093` | Base URL for shelter-service                  |
| `RISK_SERVICE_URL`           | `http://localhost:8081` | Base URL for risk-service                     |
| `NADAA_ALLOWED_ORIGINS`      | `*`                     | Comma-separated CORS origins                  |
| `NADAA_ENV`                  |                         | `development` allows localhost CORS origins   |
| `NADAA_AUTH_TOKEN_SECRET`    |                         | HMAC secret verifying auth-service tokens     |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS` |                       | `true` honors legacy `X-NADAA-Actor-*` headers (only allowed with `NADAA_ENV=development`) |

## Route planning

`POST /routes/plan` accepts a JSON body with `origin`, optional `destination`,
`waypointType` (`shelter`, `higher_ground`, `manual`), `avoidRiskLevels`, and
`closureBufferMeters`.

If `destination` is omitted and `waypointType` is `shelter`, the service queries
the shelter-service for the nearest open shelter. A shelter route is never
planned to a fabricated waypoint: when the lookup fails the service responds
`502` (`shelter_lookup_failed`), and when no open shelter is found it responds
`404` (`no_shelter_available`).

Closure avoidance queries road-closure-service with a bounding box covering the
whole origin→destination corridor. Risk avoidance samples risk-service
(`GET /api/v1/risk`) concurrently at points along the route and flags sampled
zones whose `overallRisk` is in `avoidRiskLevels`. When an upstream lookup
fails, the route is still returned but flagged `degraded: true` with an
`enrichmentStatus` naming the failed lookup (`closure_lookup_failed`,
`risk_lookup_failed`) — an empty `avoidedClosures`/`avoidedRiskZones` then
means missing data, not a verified hazard-free corridor.

The returned polyline is re-sampled against all hazards after every detour
insertion (bounded passes, trying both detour sides); a hazard is listed as
avoided only when the final polyline actually clears it.

A caller `Authorization` header is forwarded to upstream services when present.
`POST /routes/plan` bodies are limited to 1 MiB.

The response includes a polyline of waypoints, estimated distance and walking
duration, and is always marked as `decisionSupport: true` with a disclaimer to
follow official emergency instructions.

## Tests

```bash
go test ./...
```
