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

| Variable                   | Default                 | Description                       |
| -------------------------- | ----------------------- | --------------------------------- |
| `PORT`                     | `8096`                  | HTTP listen address               |
| `ROAD_CLOSURE_SERVICE_URL` | `http://localhost:8095` | Base URL for road-closure-service |
| `SHELTER_SERVICE_URL`      | `http://localhost:8093` | Base URL for shelter-service      |
| `RISK_SERVICE_URL`         | `http://localhost:8082` | Base URL for risk-service         |
| `NADAA_ALLOWED_ORIGINS`    | `*`                     | Comma-separated CORS origins      |

## Route planning

`POST /routes/plan` accepts a JSON body with `origin`, optional `destination`,
`waypointType` (`shelter`, `higher_ground`, `manual`), `avoidRiskLevels`, and
`closureBufferMeters`.

If `destination` is omitted and `waypointType` is `shelter`, the service queries
the shelter-service for the nearest shelter. If the shelter-service is
unavailable or returns no shelters, a fallback higher-ground waypoint is used.

The response includes a polyline of waypoints, estimated distance and walking
duration, and is always marked as `decisionSupport: true` with a disclaimer to
follow official emergency instructions.

## Tests

```bash
go test ./...
```
