# API

Base path: `/api/v1`

All request and response bodies are JSON unless a media upload flow explicitly uses signed upload URLs or multipart form data.

## Common Conventions

### Authentication

- Citizen endpoints accept citizen access tokens where identity is required.
- Public safety read endpoints may allow anonymous access when no personal data is returned.
- Authority endpoints require agency-user access tokens and role checks.
- MFA is required before authority users can perform sensitive actions.

### Error Shape

```json
{
  "error": {
    "code": "invalid_coordinates",
    "message": "lat and lng query parameters are required",
    "requestId": "req_01H..."
  }
}
```

### Pagination

List endpoints should support:

- `limit`
- `cursor`
- `sort`
- filter-specific query parameters

### Audit Headers

Authority write endpoints should preserve:

- actor user id
- agency id
- request id
- IP/device metadata where available
- target resource id

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

## MVP API Contracts

### Citizen Auth

`POST /api/v1/auth/citizens/register`

```json
{
  "name": "Ama Mensah",
  "phone": "+233200000000",
  "preferredLanguage": "en",
  "homeLocation": {
    "lat": 5.6037,
    "lng": -0.187
  },
  "contactPermission": true
}
```

Response:

```json
{
  "userId": "usr_...",
  "phone": "+233200000000",
  "challengeId": "otp_...",
  "otpDelivery": "mock"
}
```

In local development, `NADAA_AUTH_MOCK_OTP=123456` can force a known OTP. `NADAA_AUTH_EXPOSE_DEV_OTP=true` includes `devOtp` in the registration response and must not be enabled in production.

`POST /api/v1/auth/citizens/login`

```json
{
  "phone": "+233200000000",
  "otp": "123456"
}
```

Response:

```json
{
  "accessToken": "nadaa....",
  "tokenType": "Bearer",
  "expiresAt": "2026-07-07T12:00:00Z",
  "user": {
    "id": "usr_...",
    "name": "Ama Mensah",
    "phone": "+233200000000",
    "role": "citizen",
    "preferredLanguage": "en",
    "homeLocation": {
      "lat": 5.6037,
      "lng": -0.187
    },
    "contactPermission": true,
    "createdAt": "2026-07-06T12:00:00Z"
  }
}
```

`GET /api/v1/auth/me`

Requires `Authorization: Bearer <token>` and returns the citizen profile.

### Agency Auth

`POST /api/v1/auth/agency-users`

Authority/admin only.

```json
{
  "name": "Dispatcher One",
  "email": "dispatcher@nadaa.example",
  "phone": "+233200000001",
  "agencyId": "agency_nadmo_ama",
  "role": "dispatcher"
}
```

`POST /api/v1/auth/agency-users/{id}/mfa/setup`

`POST /api/v1/auth/agency-users/{id}/mfa/verify`

### Incident Reporting

`POST /api/v1/incidents`

```json
{
  "type": "flood",
  "description": "Road is flooded and vehicles are trapped",
  "location": {
    "lat": 5.579,
    "lng": -0.212
  },
  "peopleAffected": 12,
  "injuriesReported": false,
  "urgency": "high",
  "anonymous": false,
  "contactPermission": true,
  "accessibilityNeeds": "Elderly person needs help evacuating",
  "media": ["media_001"],
  "reporter": {
    "userId": "usr_...",
    "phone": "+233200000000"
  }
}
```

Response:

```json
{
  "id": "inc_...",
  "reference": "INC-000001",
  "status": "reported",
  "severity": "high",
  "priorityReview": false,
  "duplicateCandidates": []
}
```

Rules:

- `type` must be a supported hazard.
- `location.lat` and `location.lng` must be valid coordinates.
- `description` must be 5 to 2000 safe characters.
- `urgency` must be `low`, `moderate`, `high`, or `life_threatening`.
- `life_threatening` or injury reports are flagged for priority review.
- Anonymous reports do not retain `reportedBy` in standard incident records.
- If `contactPermission` is false, reporter phone is not retained in the incident record.
- Starter service rate-limits repeated reports by client address.

`GET /api/v1/incidents`

Starter list endpoint for development and dashboard wiring. Authority filtering and RBAC land in later stories.

### Media Upload

`POST /api/v1/media/uploads`

```json
{
  "purpose": "incident_media",
  "fileName": "flooded-road.jpg",
  "contentType": "image/jpeg",
  "sizeBytes": 820000
}
```

Returns a signed upload URL or controlled upload target.

### Incident Command

`GET /api/v1/incidents?hazard=flood&district=ama&status=reported`

`PATCH /api/v1/incidents/{id}/status`

```json
{
  "status": "verified",
  "note": "Confirmed through duplicate reports and dispatcher call"
}
```

`POST /api/v1/incidents/{id}/assignments`

```json
{
  "agencyId": "agency_fire_ama",
  "priority": "high",
  "instructions": "Check trapped vehicles near the market road"
}
```

`POST /api/v1/incidents/{id}/merge`

```json
{
  "duplicateIncidentIds": ["inc_02H...", "inc_03H..."],
  "reason": "Same flood location and reports within 10 minutes"
}
```

### Alerts

`POST /api/v1/alerts`

```json
{
  "title": "Severe Flood Warning",
  "hazardType": "flood",
  "severity": "severe_warning",
  "message": "Avoid low-lying roads and move to higher ground.",
  "target": {
    "type": "district",
    "ids": ["ama", "tema"]
  },
  "startsAt": "2026-07-06T16:00:00Z",
  "expiresAt": "2026-07-07T18:00:00Z",
  "recommendedAction": "Prepare to evacuate if instructed.",
  "evacuationRequired": false,
  "shelterIds": ["shelter-ama-001"]
}
```

`POST /api/v1/alerts/{id}/submit`

`POST /api/v1/alerts/{id}/approve`

`POST /api/v1/alerts/{id}/reject`

`GET /api/v1/alerts?current=true&lat=5.6037&lng=-0.1870`

### Guidance And Shelters

`GET /api/v1/guides?hazard=flood&stage=before&language=en`

`GET /api/v1/shelters/nearby?lat=5.6037&lng=-0.1870`

`PATCH /api/v1/shelters/{id}/occupancy`

Authority only.

```json
{
  "currentOccupancy": 116,
  "capacity": 450,
  "status": "open"
}
```

### ML Predictions

`GET /api/v1/ml/predictions/flood?lat=5.6037&lng=-0.1870`

```json
{
  "hazardType": "flood",
  "modelVersion": "flood-baseline-0.1.0",
  "predictionTime": "2026-07-06T15:00:00Z",
  "targetTime": "2026-07-06T18:00:00Z",
  "probability": 0.82,
  "severity": "severe",
  "confidence": "medium",
  "explanation": [
    "Heavy rainfall forecast",
    "Low elevation",
    "Historical flood zone"
  ]
}
```

`POST /api/v1/alerts/from-prediction/{predictionId}`

Creates an alert draft only. It must still pass approval.

## Phase 2 API Areas

- SMS/USSD inbound webhook.
- WhatsApp inbound webhook.
- Voice alert asset review.
- Volunteer registration and assignment.
- Hospital capacity updates.
- Relief distribution point management.
- Donation/aid request and pledge management.
- Road closure management.
- Evacuation route planning.
- Missing persons intake and review.
- Property damage report/export.
- Drone/satellite imagery ingestion.

## Phase 3 API Areas

- Flood simulation jobs and map layers.
- AI incident triage suggestions and overrides.
- Computer vision evidence review.
- Predictive resource positioning.
- School preparedness profiles and drill logs.
- Campaign publishing.
- Open data catalog and exports.
- Cell broadcast adapter and simulator.
