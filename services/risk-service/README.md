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

`GET /api/v1/risk` validates coordinates, returns low/high/severe flood risk bands, includes nearby shelters and response facilities within 30 km, and emits recommended citizen actions.

When `NADAA_ML_API_URL` is set to the ML API base URL, for example `http://127.0.0.1:8094/api/v1`, the response also includes `mlPrediction` decision support with model version, probability, severity, confidence, explanation factors, prediction log id, `humanReviewRequired=true`, and `autoPublishAllowed=false`. ML failures are logged and do not block rule-based risk responses. Set `NADAA_INTERNAL_SERVICE_TOKEN` to send it as `X-NADAA-Service-Token` on the ML prediction call (unset means no token is sent), and any inbound `Authorization` header is forwarded to the ML service.

Set `NADAA_RISK_ADDR` to override the default `:8081` bind address. When `NADAA_ALLOWED_ORIGINS` is set, localhost origins are only tolerated while `NADAA_ENV=development`.
