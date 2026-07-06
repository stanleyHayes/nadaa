# Risk Service

The risk service returns hazard risk summaries, nearby shelters, and recommended actions.

Current starter endpoints:

- `GET /healthz`
- `GET /api/v1/risk?lat=5.6037&lng=-0.1870`

The MVP baseline uses seed-aligned in-memory fixtures for:

- Accra flood and fire risk zones.
- Accra shelters.
- NADMO, fire, and ambulance facilities.
- A recent flood report near Accra Central.

`GET /api/v1/risk` validates coordinates, returns low/high/severe flood risk bands, includes nearby shelters and response facilities within 30 km, and emits recommended citizen actions. This locks in the public API contract while service-level PostGIS persistence, weather/hydrology imports, and ML predictions are added later.

Set `NADAA_RISK_ADDR` to override the default `:8081` bind address.
