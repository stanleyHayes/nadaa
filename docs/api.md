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
  "location": "Accra Metropolitan",
  "overallRisk": "high",
  "risks": [
    {
      "type": "flood",
      "level": "high",
      "probability": 0.72,
      "reason": "Within 1523m of a severe flood zone and 0m of a recent high flood report."
    }
  ],
  "nearestShelters": [
    {
      "id": "00000000-0000-0000-0000-000000000301",
      "name": "Accra Metro Assembly Shelter",
      "location": { "lat": 5.56, "lng": -0.2 },
      "capacity": 450,
      "currentOccupancy": 116,
      "contact": "112",
      "distanceMeters": 5068,
      "status": "open",
      "facilities": ["water", "first_aid", "accessible_entry", "family_area"]
    }
  ],
  "nearbyFacilities": [
    {
      "id": "00000000-0000-0000-0000-000000000102",
      "name": "Ghana National Fire Service Accra",
      "type": "fire",
      "location": { "lat": 5.565, "lng": -0.185 },
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "contact": "112",
      "distanceMeters": 4308
    }
  ],
  "recommendedActions": [
    "Avoid known flood-prone roads and monitor official NADMO updates.",
    "Check the route to the nearest open shelter before rainfall intensifies.",
    "Keep phones charged and report blocked drains or rising water early."
  ]
}
```

Rules:

- `lat` and `lng` are required and must be valid coordinates.
- The MVP baseline uses seed-aligned fixtures for flood zones, fire zones, shelters, facilities, and one recent flood report.
- Flood scoring is rule-based: inside the flood zone returns `severe`, near both the flood zone and recent report returns `high`, near only a recent report returns `moderate`, and locations outside fixture coverage return `low`.
- Nearby shelters and facilities are returned within 30 km and sorted by distance.

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

Requires `Authorization: Bearer <token>` and returns the citizen or agency-user profile for the token.

### Agency Auth

`POST /api/v1/auth/agency-users`

Requires an agency `system_admin` or `agency_admin` bearer token with MFA completed. Agency admins can create users only inside their own agency.

```json
{
  "name": "Dispatcher One",
  "email": "dispatcher@nadaa.example",
  "phone": "+233200000001",
  "agencyId": "00000000-0000-0000-0000-000000000101",
  "role": "dispatcher"
}
```

Response:

```json
{
  "user": {
    "id": "usr_...",
    "name": "Dispatcher One",
    "email": "dispatcher@nadaa.example",
    "phone": "+233200000001",
    "role": "dispatcher",
    "agency": {
      "id": "00000000-0000-0000-0000-000000000101",
      "name": "NADMO Accra Metro",
      "type": "nadmo",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "contactNumber": "112"
    },
    "mfaRequired": true,
    "mfaEnabled": false,
    "createdAt": "2026-07-06T12:00:00Z",
    "updatedAt": "2026-07-06T12:00:00Z"
  },
  "temporaryPassword": "tmp_...",
  "mfaSetupRequired": true
}
```

`POST /api/v1/auth/agency-users/{id}/mfa/setup`

```json
{
  "email": "dispatcher@nadaa.example",
  "temporaryPassword": "tmp_..."
}
```

Response:

```json
{
  "userId": "usr_...",
  "challengeId": "mfa_...",
  "method": "mock_totp",
  "secret": "mfa_secret_...",
  "expiresAt": "2026-07-06T12:10:00Z"
}
```

In local development, `NADAA_AUTH_MOCK_OTP=123456` can force the MFA challenge code and `NADAA_AUTH_EXPOSE_DEV_OTP=true` includes `devCode` in the setup response.

`POST /api/v1/auth/agency-users/{id}/mfa/verify`

```json
{
  "email": "dispatcher@nadaa.example",
  "temporaryPassword": "tmp_...",
  "code": "123456"
}
```

Response:

```json
{
  "user": {
    "id": "usr_...",
    "name": "Dispatcher One",
    "email": "dispatcher@nadaa.example",
    "phone": "+233200000001",
    "role": "dispatcher",
    "agency": {
      "id": "00000000-0000-0000-0000-000000000101",
      "name": "NADMO Accra Metro",
      "type": "nadmo",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "contactNumber": "112"
    },
    "mfaRequired": true,
    "mfaEnabled": true,
    "createdAt": "2026-07-06T12:00:00Z",
    "updatedAt": "2026-07-06T12:03:00Z"
  }
}
```

`POST /api/v1/auth/agency/login`

```json
{
  "email": "dispatcher@nadaa.example",
  "password": "tmp_...",
  "mfaCode": "123456"
}
```

Response:

```json
{
  "accessToken": "nadaa....",
  "tokenType": "Bearer",
  "expiresAt": "2026-07-07T00:00:00Z",
  "user": {
    "id": "usr_...",
    "name": "Dispatcher One",
    "email": "dispatcher@nadaa.example",
    "phone": "+233200000001",
    "role": "dispatcher",
    "agency": {
      "id": "00000000-0000-0000-0000-000000000101",
      "name": "NADMO Accra Metro",
      "type": "nadmo",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "contactNumber": "112"
    },
    "mfaRequired": true,
    "mfaEnabled": true,
    "createdAt": "2026-07-06T12:00:00Z",
    "updatedAt": "2026-07-06T12:03:00Z"
  }
}
```

Agency login returns `mfa_setup_required` until setup and verification are complete, and `mfa_required` when a verified agency user omits the MFA code.

### Audit Logs

`GET /api/v1/audit/logs?limit=50`

Requires a `system_admin` bearer token with MFA completed. Returns the latest audit logs first. `limit` defaults to 50 and is capped at 100.

Response:

```json
{
  "logs": [
    {
      "id": "aud_...",
      "actorUserId": "usr_...",
      "actorAgencyId": "00000000-0000-0000-0000-000000000101",
      "actorRole": "system_admin",
      "action": "auth.agency_user.created",
      "targetType": "agency_user",
      "targetId": "usr_...",
      "requestId": "req-create-dispatcher",
      "ipAddress": "203.0.113.10",
      "userAgent": "nadaa-test/1.0",
      "after": {
        "id": "usr_...",
        "role": "dispatcher",
        "agencyId": "00000000-0000-0000-0000-000000000101",
        "mfaRequired": true,
        "mfaEnabled": false
      },
      "createdAt": "2026-07-06T12:00:00Z"
    }
  ]
}
```

Current auth-service audit actions include citizen registration/login, agency login, agency-user creation, MFA setup/verification, RBAC denial, and audit-log viewing. Audit snapshots must not include OTPs, MFA codes, temporary passwords, tokens, or provider secrets.

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
  "duplicateCandidates": [
    {
      "incidentId": "inc_...",
      "reference": "INC-000001",
      "score": 0.92,
      "distanceMeters": 48,
      "minutesApart": 12,
      "reasons": [
        "same_hazard",
        "nearby_location",
        "recent_report",
        "similar_description"
      ]
    }
  ]
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
- Duplicate candidates are review hints only. The starter baseline compares same-hazard reports within 750 meters and 3 hours using distance, time, and description similarity.
- No duplicate candidate is automatically merged, hidden, deleted, or downgraded.

`GET /api/v1/incidents`

Starter list endpoint for development and authority command-map wiring. Incident records include `location`, `severity`, `status`, `type`, `createdAt`, `duplicateCandidates`, `verifiedBy`, `verifiedAt`, `statusReason`, `resolutionNotes`, and `closedAt` so the dashboard can render map markers, synchronized queue rows, filters, status controls, and duplicate review prompts. Backend authority filtering and agency scoping land in later service stories; the current dashboard applies client-side command filters while the API contract stabilizes.

### Media Upload

`POST /api/v1/media/uploads`

```json
{
  "purpose": "incident_media",
  "fileName": "flooded-road.jpg",
  "contentType": "image/jpeg",
  "sizeBytes": 820000,
  "uploadedBy": "usr_..."
}
```

Returns a signed upload URL or controlled upload target.

```json
{
  "mediaId": "media_...",
  "uploadUrl": "/dev/uploads/media_.../flooded-road.jpg",
  "method": "PUT",
  "headers": {
    "Content-Type": "image/jpeg"
  },
  "expiresAt": "2026-07-06T12:15:00Z",
  "maxSizeBytes": 10485760,
  "access": "private"
}
```

Starter service rules:

- Supported types: `image/jpeg`, `image/png`, `image/webp`, `video/mp4`, `video/quicktime`, `audio/mpeg`, `audio/mp4`, and `audio/wav`.
- Images are limited to 10 MB.
- Audio is limited to 25 MB.
- Video is limited to 100 MB.
- Media is private by default.
- Incident reports must reference media IDs created through this endpoint.
- When an incident is created, known media IDs are marked as linked to that incident.

`GET /api/v1/media`

Starter development endpoint for inspecting private media metadata and incident linkage. Authority RBAC lands in later stories.

### Incident Command

`GET /api/v1/incidents?hazard=flood&district=ama&status=reported`

Authority workflow endpoints require these headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

`POST /api/v1/incidents/{id}/verify`

```json
{
  "note": "Confirmed through duplicate reports and dispatcher call"
}
```

Allowed verification roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. Verification moves `reported` or `under_review` incidents to `verified`, stores verifier metadata, and writes an `incident.verified` audit event.

`PATCH /api/v1/incidents/{id}/status`

```json
{
  "status": "response_en_route",
  "note": "Fire crew dispatched from Circle station"
}
```

Supported statuses are `reported`, `under_review`, `verified`, `assigned`, `response_en_route`, `on_scene`, `contained`, `recovery_ongoing`, `closed`, and `false_report`. Status values may be sent with spaces or hyphens; the API normalizes them to underscore form.

Rules:

- Allowed workflow roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, and `responder`.
- Status changes must follow the configured transition graph; for example, `reported` can move to `under_review`, `verified`, or `false_report`, but not directly to `closed`.
- `closed` and `false_report` are terminal.
- `resolutionNotes` are required for `closed` and `false_report`.
- Every accepted status change writes an incident audit event with before/after snapshots.

`GET /api/v1/incidents/audit?limit=50`

Returns latest incident workflow audit events for `system_admin`, `agency_admin`, and `nadmo_officer` roles. `limit` defaults to 50 and is capped at 100.

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

Requires authority headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

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

Response:

```json
{
  "id": "alert_000001",
  "title": "Severe Flood Warning",
  "hazardType": "flood",
  "severity": "severe_warning",
  "message": "Avoid low-lying roads and move to higher ground.",
  "target": {
    "type": "district",
    "ids": ["ama", "tema"],
    "label": "Accra Metropolitan and Tema"
  },
  "startsAt": "2026-07-06T16:00:00Z",
  "expiresAt": "2026-07-07T18:00:00Z",
  "recommendedAction": "Prepare to evacuate if instructed.",
  "evacuationRequired": false,
  "shelterIds": ["shelter-ama-001"],
  "issuingAgencyId": "00000000-0000-0000-0000-000000000101",
  "issuedBy": "usr_...",
  "status": "draft",
  "emergencyOverride": false,
  "createdAt": "2026-07-06T12:00:00Z",
  "updatedAt": "2026-07-06T12:00:00Z"
}
```

`PATCH /api/v1/alerts/{id}`

Updates a `draft` or `rejected` alert using the same body as create.

`POST /api/v1/alerts/{id}/submit`

`POST /api/v1/alerts/{id}/approve`

```json
{
  "note": "Reviewed target, timing, and recommended action."
}
```

`POST /api/v1/alerts/{id}/reject`

```json
{
  "reason": "Target area is too broad for the current incident."
}
```

`POST /api/v1/alerts/{id}/emergency-override`

```json
{
  "reason": "Immediate life-safety warning approved by NADMO officer."
}
```

`GET /api/v1/alerts?current=true&lat=5.6037&lng=-0.1870`

Without authority headers, alert listing returns only approved or published public alerts. With authority headers, the service returns draft, submitted, approved, and rejected workflow records. `status` may filter by `draft`, `submitted`, `approved`, `rejected`, `published`, `expired`, or `cancelled`.

`GET /api/v1/alerts/audit?limit=50`

Returns alert workflow audit events for authorized approvers.

Rules:

- Draft/update/submit actions allow `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.
- Approve/reject actions allow `system_admin`, `agency_admin`, and `nadmo_officer`.
- Non-system approvers cannot approve their own draft.
- Emergency override allows only `system_admin` and `nadmo_officer`, requires a reason, marks the alert approved, and creates an `alert.emergency_override` audit event.
- Alert expiry is mandatory and must be after `startsAt`.
- Delivery and notification logs land in NADAA-052; this endpoint stops at approved workflow state.

### Guidance And Shelters

`GET /api/v1/guides?hazard=flood&stage=before&language=en`

Returns emergency guide records for the requested hazard, stage, and language.

```json
{
  "guides": [
    {
      "id": "guide_flood_before_en",
      "hazardType": "flood",
      "stage": "before",
      "title": "Prepare before flooding",
      "body": "Know your nearest shelter, keep documents dry, clear drains safely, prepare drinking water, and agree on a family meeting point.",
      "language": "en",
      "offlineAvailable": true,
      "sortOrder": 10,
      "createdAt": "2026-07-06T12:00:00Z",
      "updatedAt": "2026-07-06T12:00:00Z"
    }
  ]
}
```

Rules:

- `hazard` must be a supported NADAA hazard type when provided.
- `stage` must be `before`, `during`, `after`, or `recovery` when provided.
- `language` defaults to `en`.
- If a requested non-English language has no exact match, the guide service falls back to English for the same filters.
- `offline=true` returns only guides marked as offline available.
- Initial content covers floods, fire safety, road crash response, electrical hazard safety, disease prevention, safe evacuation, emergency bag checklist, family emergency planning, and contacting 112.
- General preparedness topics use hazard type `other`.

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

### Integrations

`GET /api/v1/integrations/contracts?domain=weather&direction=inbound`

Returns integration contracts for official agencies and fixture/mock adapters.

```json
{
  "contracts": [
    {
      "id": "gmet-rainfall-nowcast",
      "partner": "Ghana Meteorological Agency",
      "partnerType": "meteorological",
      "domain": "weather",
      "direction": "inbound",
      "dataOwner": "GMet",
      "cadence": "Every 15 minutes during watch/warning periods",
      "authentication": {
        "mode": "api_key",
        "requiredHeaders": ["X-NADAA-Source", "X-NADAA-Signature"],
        "secretScope": "environment_secret_manager"
      },
      "payloads": [
        {
          "name": "WeatherObservation",
          "contentType": "application/json",
          "requiredFields": [
            "source",
            "observedAt",
            "validFrom",
            "validTo",
            "location.lat",
            "location.lng",
            "rainfallMm"
          ],
          "pii": "none",
          "geometry": "Point WGS84",
          "exampleRef": "docs/integrations.md#weather-observation"
        }
      ],
      "freshnessWindowMinutes": 30,
      "status": "mock_contract"
    }
  ]
}
```

Rules:

- `domain` may be `weather`, `hydrology`, `incident_sync`, `alert_sync`, `road_closure`, `hospital_capacity`, `utility_outage`, or `shelter_status`.
- `direction` may be `inbound`, `outbound`, or `bidirectional`.
- Contracts define data ownership, cadence, expected payloads, authentication, freshness windows, retry/dead-letter behavior, and manual fallback.
- Mock contracts do not imply official production access or credentials.

`GET /api/v1/integrations/mock/weather-hydrology/observations?metric=rainfall_mm`

Returns fixture weather and hydrology observations for importer development.

`POST /api/v1/integrations/mock/sync-events`

Accepts mock outbound incident or alert sync events and returns `202 Accepted`.

```json
{
  "type": "incident",
  "sourceId": "inc_001",
  "reference": "INC-000001",
  "hazardType": "flood",
  "status": "verified",
  "severity": "high",
  "summary": "Flooded road near market",
  "location": { "lat": 5.6037, "lng": -0.187 },
  "targetAgencyIds": ["00000000-0000-0000-0000-000000000101"],
  "correlationId": "corr_001"
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
