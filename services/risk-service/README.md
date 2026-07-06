# Risk Service

The risk service returns hazard risk summaries, nearby shelters, and recommended actions.

Current starter endpoints:

- `GET /healthz`
- `GET /api/v1/risk?lat=5.6037&lng=-0.1870`

The first implementation uses sample flood-focused responses. Later stories will connect PostGIS risk zones, weather/hydrology imports, shelter records, and ML predictions.

