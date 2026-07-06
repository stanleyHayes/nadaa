# API

Base path: `/api/v1`

## Current Starter Endpoint

### Check Area Risk

`GET /api/v1/risk?lat=5.6037&lng=-0.1870`

Returns:

```json
{
  "location": "Accra Central",
  "overallRisk": "high",
  "risks": [
    {
      "type": "flood",
      "level": "severe",
      "probability": 0.82,
      "reason": "Heavy rainfall forecast, low elevation, and historical flood reports nearby."
    }
  ],
  "nearestShelters": [],
  "recommendedActions": []
}
```

## Planned MVP Endpoints

- `POST /api/v1/auth/citizens/register`
- `POST /api/v1/auth/citizens/login`
- `POST /api/v1/incidents`
- `GET /api/v1/incidents`
- `PATCH /api/v1/incidents/{id}/status`
- `POST /api/v1/incidents/{id}/assignments`
- `POST /api/v1/alerts`
- `POST /api/v1/alerts/{id}/submit`
- `POST /api/v1/alerts/{id}/approve`
- `GET /api/v1/alerts`
- `GET /api/v1/guides`
- `GET /api/v1/shelters/nearby`

