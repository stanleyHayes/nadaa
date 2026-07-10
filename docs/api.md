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
  "mlPrediction": {
    "id": "pred_grid-accra-north-002",
    "modelVersion": "flood-logistic-baseline-0.1.0",
    "hazardType": "flood",
    "predictionTime": "2026-07-06T10:00:00Z",
    "targetTime": "2026-07-06T12:00:00Z",
    "cellId": "grid-accra-north-002",
    "region": "Greater Accra",
    "district": "Accra Metropolitan",
    "community": "Accra North",
    "probability": 0.9828,
    "severity": "high",
    "expectedOnset": "24_to_48h",
    "confidence": "medium",
    "explanationFactors": [
      {
        "feature": "vulnerable_population_pct",
        "label": "vulnerable population share",
        "value": 21,
        "contribution": 0.8183,
        "direction": "increases_risk"
      }
    ],
    "inputFeatureSetVersion": "flood-risk-features.v1",
    "predictionLogId": "ml_log_20260706100000_grid_accra_north_002",
    "humanReviewRequired": true,
    "autoPublishAllowed": false,
    "source": "baseline_fixture_model"
  },
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
- When `NADAA_ML_API_URL` is configured, risk-service attaches `mlPrediction` from the ML service as decision support. The response keeps `humanReviewRequired=true` and `autoPublishAllowed=false`; no public alert can be published from model output without the alert approval workflow.
- Dispatcher web can create an alert draft from a reviewed prediction by sending the normal `POST /api/v1/alerts` request with `sourcePrediction` metadata. The draft remains in the standard approval workflow and is audited as `alert.created`.

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
  "abuseSignals": [],
  "abuseScore": 0,
  "abuseReviewRequired": false,
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
- If `contactPermission` is false, reporter phone is not retained in the incident record and authority views hide reporter identity.
- Starter service rate-limits repeated reports by client address.
- Suspicious report signals are stored as transparent `abuseSignals` with a weighted `abuseScore` and `abuseReviewRequired` flag for dispatchers.
- Automated suspicion does not block report creation or suppress life-threatening reports.
- Duplicate candidates are review hints only. The starter baseline compares same-hazard reports within 750 meters and 3 hours using distance, time, and description similarity.
- No duplicate candidate is automatically merged, hidden, deleted, or downgraded.

`GET /api/v1/incidents`

Authority-only incident list endpoint for command-map wiring. Calls require authority actor, role, agency, MFA-completed, and request-id headers. Incident records include `location`, `severity`, `status`, `type`, `createdAt`, `abuseSignals`, `abuseScore`, `abuseReviewRequired`, `duplicateCandidates`, `mergedIncidentIds`, `mergedIntoId`, `verifiedBy`, `verifiedAt`, `statusReason`, `resolutionNotes`, `closedAt`, and `privacy` so dashboards can render map markers, synchronized queue rows, filters, status controls, assignment controls, safety review controls, duplicate review prompts, and privacy indicators.

Privacy metadata is returned with each authority incident view:

```json
{
  "privacy": {
    "reporterIdentityVisible": true,
    "reporterContactVisible": true,
    "locationPrecision": "exact",
    "locationUse": "emergency_response",
    "disclosure": "Location is used to route emergency response, detect duplicates, and coordinate verified authority actions.",
    "notes": [
      "Exact incident location is available only to MFA-verified authority users for emergency response coordination."
    ]
  }
}
```

Privacy rules:

- Reporter identity and contact are visible only when the report is not anonymous, contact permission is granted, reporter details exist, and the authority role is allowed to view contact details.
- `responder` and `agency_viewer` receive standard operational incident views without `reportedBy`.
- Exact location is available only through MFA-verified authority incident endpoints for emergency response, duplicate detection, and verified authority coordination.
- Duplicate review and merge response payloads apply the same incident privacy rules to primary and candidate records.

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

Authority actors can filter to incidents assigned to their agency:

`GET /api/v1/incidents?assignedToMe=true`

`GET /api/v1/incidents?assignedAgencyId=00000000-0000-0000-0000-000000000201`

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

`POST /api/v1/incidents/{id}/abuse-review`

```json
{
  "decision": "clear",
  "note": "Dispatcher confirmed caller and location details"
}
```

Allowed review decisions are `clear`, `monitor`, and `false_report`. Allowed review roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`; MFA is required. `false_report` requires `resolutionNotes`, moves the incident to terminal `false_report`, and writes an `incident.false_reported` audit event. `clear` and `monitor` update review metadata and write `incident.abuse_cleared` or `incident.abuse_monitored` audit events.

`GET /api/v1/incidents/audit?limit=50`

Returns latest incident workflow audit events for `system_admin`, `agency_admin`, and `nadmo_officer` roles. `limit` defaults to 50 and is capped at 100.

`POST /api/v1/incidents/{id}/assignments`

```json
{
  "agencyId": "00000000-0000-0000-0000-000000000201",
  "agencyName": "Ghana National Fire Service",
  "agencyType": "fire",
  "priority": "high",
  "instructions": "Check trapped vehicles near the market road",
  "responderLead": "Station Officer Mensah"
}
```

Allowed assignment roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. `agency_admin` actors can assign only to their own agency. Incidents must be at least `verified`; `reported`, `under_review`, `closed`, and `false_report` incidents reject assignment.

Accepted assignments append an active assignment record, move a `verified` incident to `assigned`, preserve later operational statuses, append an `incident.assigned` timeline event, and write an `incident.assigned` audit event with assignment counts and assigned agency ids.

### Community Volunteers

`POST /api/v1/volunteers`

Registers a citizen as a community volunteer candidate and joins them to a district/community response group. Profiles start as `pending` until an authorized officer verifies them.

```json
{
  "citizenUserId": "usr_volunteer_001",
  "name": "Ama Volunteer",
  "phone": "+233200000111",
  "region": "Greater Accra",
  "district": "Accra Metropolitan",
  "community": "Jamestown",
  "skills": ["first aid", "community alerts"],
  "languages": ["en", "tw"],
  "availabilityStatus": "available"
}
```

`GET /api/v1/volunteers?status=verified&district=Accra%20Metropolitan`

Authority-only volunteer list. Calls use the same authority headers as incident command endpoints. Supported filters are `status` and `district`.

`POST /api/v1/volunteers/{id}/verify`

```json
{
  "decision": "verify",
  "note": "District officer checked ID and community lead reference."
}
```

Allowed decisions are `verify`, `reject`, and `suspend`. Verification writes `volunteer.verified`, `volunteer.rejected`, or `volunteer.suspended` audit events against `targetType: volunteer_profile`.

`POST /api/v1/incidents/{id}/volunteer-tasks`

Assigns a verified, available volunteer to an incident-linked support task.

```json
{
  "volunteerId": "vol_000001",
  "type": "welfare_check",
  "priority": "high",
  "instructions": "Check whether households near the shelter approach need water or transport. Stay on public roads.",
  "locationLabel": "Jamestown shelter approach"
}
```

Allowed task types are `welfare_check`, `shelter_support`, `supply_distribution`, `damage_observation`, `route_observation`, and `community_alerting`. Incidents must be at least `verified`; closed and false-report incidents reject volunteer tasks. The API rejects instructions that tell civilians to enter floodwater, fight fires, conduct rescues, enter collapsed structures, approach armed/violent scenes, or direct highway traffic. Accepted assignments append `incident.volunteer_assigned` to the incident timeline and write both incident and volunteer-task audit events.

`GET /api/v1/volunteers/{id}/tasks`

Returns assigned volunteer tasks for the mobile/PWA task view.

`PATCH /api/v1/volunteer-tasks/{id}/status`

```json
{
  "volunteerId": "vol_000001",
  "status": "accepted",
  "note": "I can check from the public road.",
  "safetyStatus": "safe",
  "location": { "lat": 5.56, "lng": -0.2 }
}
```

Allowed statuses are `accepted`, `en_route`, `on_scene`, `completed`, `cancelled`, and `needs_escalation`. `needs_escalation`, `unsafe`, or `needs_authority` updates append escalation timeline events.

`POST /api/v1/volunteer-tasks/{id}/observations`

```json
{
  "volunteerId": "vol_000001",
  "observation": "Water is rising near the footbridge and families are waiting for authority transport.",
  "safetyStatus": "needs_authority",
  "location": { "lat": 5.562, "lng": -0.202 },
  "escalationRequested": true,
  "media": ["media_volunteer_photo_001"]
}
```

Volunteer observations append `incident.volunteer_observation` to the incident timeline. Escalated observations also append `incident.volunteer_escalation` so dispatchers and agency users can see field risk without switching tools.

`GET /api/v1/incidents/{id}/duplicates`

Returns the selected incident plus full incident records for open duplicate candidates. Allowed reader roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, `responder`, and `agency_viewer`; MFA is required.

`POST /api/v1/incidents/{id}/merge`

```json
{
  "duplicateIncidentIds": ["inc_02H...", "inc_03H..."],
  "note": "Same flood location and reports within 10 minutes"
}
```

Allowed merge roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. Each duplicate id must already be a duplicate candidate for the primary incident. The primary incident remains the operational record; merged duplicate records are closed with `mergedIntoId`, `mergedBy`, `mergedAt`, and `mergeReason` trace fields. Accepted merges append `incident.merged` and `incident.merged_into` timeline events and create audit events for the primary and duplicate records.

`GET /api/v1/incidents/{id}/triage`

Returns an explainable AI triage suggestion for an incident. Allowed reader roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher`, `responder`, and `agency_viewer`; MFA is required. Every call logs an `incident.triage_suggested` audit event containing the suggestion snapshot, so all model outputs are reviewable. Duplicate signals only count open candidates: incidents already merged or marked as false reports are excluded, matching the duplicate-review endpoint.

```json
{
  "suggestion": {
    "suggestionId": "trs_1f6c...",
    "severity": "high",
    "duplicateLikelihood": 0.82,
    "topDuplicateIncidentIds": ["inc_02H..."],
    "affectedPopulation": 31,
    "suggestedAgency": {
      "agencyType": "nadmo",
      "agencyId": "00000000-0000-0000-0000-000000000101",
      "name": "NADMO Accra Metro",
      "reason": "NADMO coordinates multi-hazard disaster response for this report."
    },
    "confidence": "high",
    "modelVersion": "incident-triage-rules-0.1.0",
    "featureSetVersion": "incident-features.v1",
    "explanationFactors": [
      {
        "feature": "urgency",
        "label": "Reported urgency",
        "value": "high",
        "contribution": 0.6,
        "direction": "increases_risk"
      }
    ],
    "humanReviewRequired": true,
    "autoPublishAllowed": false
  }
}
```

The suggestion is decision-support only: it never verifies, assigns, closes, or alerts the incident. `humanReviewRequired` is always `true` and `autoPublishAllowed` is always `false`.

`POST /api/v1/incidents/{id}/triage-review`

Records dispatcher acceptance or override of the triage suggestion.

```json
{
  "accepted": false,
  "suggestionId": "trs_1f6c...",
  "overriddenFields": {
    "severity": "emergency",
    "affectedPopulation": 60,
    "suggestedAgencyType": "fire",
    "suggestedAgencyId": "00000000-0000-0000-0000-000000000201"
  },
  "reason": "Dispatcher callback confirmed trapped vehicles and upgraded severity."
}
```

Allowed review roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. A reason is required when `accepted` is `false` or `overriddenFields` is supplied. `overriddenFields` is sparse: send only the fields the dispatcher actually edited, an empty object is rejected with `empty_override`, and the audit snapshot records only the supplied fields so an unedited field is never logged as an explicit zero. When `suggestionId` references one of the last five logged suggestions for the incident, the audit entry pairs the dispatcher decision with that exact suggestion (`triageSuggestionSource: "logged"`); an unknown `suggestionId` is rejected with `unknown_suggestion`, and omitting it falls back to a fresh recomputation (`"recomputed"`). The endpoint appends an `incident.triage_accepted` or `incident.triage_overridden` timeline event and audit log entry with the reviewed model suggestion and dispatcher values. It does not modify incident status or create assignments automatically.

### Road Closures

`GET /api/v1/road-closures?status=active&lat=5.570&lng=-0.200&radius=30000&bbox=-0.30,5.50,-0.15,5.60&limit=50`

Public read endpoint returning active road closures by default. Filters include `status` (`active`, `scheduled`, `lifted`, `cancelled`), point/radius search, bounding box (`minLng,minLat,maxLng,maxLat`), `limit`, and `includeExpired`.

`POST /api/v1/road-closures`

Requires authority headers:

- `X-NADAA-Actor-ID`
- `X-NADAA-Actor-Role`
- `X-NADAA-Agency-ID`
- `X-NADAA-MFA-Completed: true`
- `X-NADAA-Request-ID`

```json
{
  "roadName": "Accra New Town Road",
  "reason": "Flooding",
  "status": "active",
  "severity": "high",
  "geometry": {
    "type": "LineString",
    "coordinates": [
      [-0.205, 5.57],
      [-0.19, 5.58]
    ]
  },
  "validTo": "2026-07-07T18:00:00Z",
  "detourNote": "Use Kanda Highway"
}
```

Allowed create/update roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. Status values are `active`, `scheduled`, `lifted`, and `cancelled`. Severity values are `low`, `moderate`, `high`, `severe`, and `emergency`.

`PATCH /api/v1/road-closures/{id}`

Accepts the same fields as create; all are optional. Status changes are source-attributed and reviewable.

`POST /api/v1/road-closures/imports/adapter`

Accepts an adapter payload (for example from `police-road-closure-feed`) with a WKT `LINESTRING` geometry:

```json
{
  "source": "ghana-police",
  "sourceRef": "police-feed-001",
  "roadName": "Sample Market Road",
  "status": "active",
  "reason": "Flooding",
  "geometry": "LINESTRING(-0.20 5.56, -0.19 5.57)",
  "validFrom": "2026-07-07T12:00:00Z",
  "validTo": "2026-07-08T12:00:00Z",
  "detour": "Use Independence Avenue"
}
```

The integration-service also exposes `POST /api/v1/integrations/road-closures/imports` to receive adapter records from partner feeds. It validates the payload, forwards it to the `road-closure-service` adapter endpoint using `NADAA_ROAD_CLOSURE_SERVICE_URL` (default `http://localhost:8095`), and records the import locally for observability.

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
    "ids": ["accra-metropolitan"],
    "label": "Accra Metropolitan"
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
    "ids": ["accra-metropolitan"],
    "label": "Accra Metropolitan",
    "center": { "lat": 5.56, "lng": -0.2 },
    "radiusMeters": 9000,
    "geometry": {
      "type": "Polygon",
      "coordinates": [
        [
          [-0.281, 5.479],
          [-0.119, 5.479],
          [-0.119, 5.641],
          [-0.281, 5.641],
          [-0.281, 5.479]
        ]
      ]
    },
    "areaSqKm": 61,
    "estimatedPopulation": 284000
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

Targets support `national`, `region`, `district`, `radius`, `community`, and `custom`. `region`, `district`, and `community` targets currently resolve against the starter catalog. `radius` targets require `center` and `radiusMeters`. `custom` targets require a closed GeoJSON-style polygon geometry. Reviewed ML drafts may include optional `sourcePrediction` metadata; this does not change the draft status or approval workflow.

`POST /api/v1/alerts/targets/preview`

Returns the normalized target with geometry, approximate area, estimated population, summary text, and warnings before the alert is created.

`GET /api/v1/alerts?current=true&targetType=district&targetId=accra-metropolitan`

Lists current alerts queryable by target type and target id.

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
- Delivery and notification logs are owned by the notification service; this endpoint stops at approved workflow state.

### Notification Delivery

`GET /api/v1/notifications/alerts?includeExpired=true`

Returns citizen-facing alert feed records. The notification service reads approved/published alerts from alert-service when available and includes fixture fallback records for local development.

```json
{
  "alerts": [
    {
      "id": "alert_feed_current_flood",
      "title": "Severe flood warning",
      "hazardType": "flood",
      "severity": "severe_warning",
      "message": "Heavy rainfall and rising drains may flood low-lying parts of Accra Metro and Tema.",
      "targetLabel": "Accra Metro and Tema",
      "startsAt": "2026-07-06T11:30:00Z",
      "expiresAt": "2026-07-06T17:00:00Z",
      "status": "current",
      "recommendedAction": "Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed.",
      "evacuationRequired": true,
      "shelterIds": ["shelter-ama-001"],
      "source": "fixture",
      "updatedAt": "2026-07-06T11:40:00Z"
    }
  ],
  "generatedAt": "2026-07-06T12:00:00Z",
  "source": "alert-service+fixture"
}
```

Supported filters: `status=current|expired|upcoming|all`, `includeExpired=true`, `hazard`, `severity`, `targetType`, and `targetId`.

`POST /api/v1/notifications/alerts/{id}/deliver`

Creates push and/or SMS delivery attempts through the configured provider abstraction.

```json
{
  "recipientId": "usr_demo_citizen",
  "phone": "+233200000000",
  "pushToken": "ExponentPushToken-demo",
  "channels": ["push", "sms"],
  "language": "en"
}
```

Response:

```json
{
  "attempts": [
    {
      "id": "delivery_000001",
      "alertId": "alert_feed_current_flood",
      "alertTitle": "Severe flood warning",
      "channel": "push",
      "provider": "mock_push",
      "recipientRef": "usr_demo_citizen",
      "status": "delivered",
      "messageId": "mock_push_alert_feed_current_flood_1783346400",
      "attemptedAt": "2026-07-06T12:00:00Z"
    }
  ]
}
```

`GET /api/v1/notifications/delivery-logs?alertId=alert_feed_current_flood&channel=sms`

Returns logged delivery attempts. Supported filters include `channel=push|sms|voice`. `NADAA_SMS_ENABLED=false` keeps SMS disabled in development and logs attempts as `skipped`; mock providers are used by default when push/SMS are enabled.

`POST /api/v1/notifications/voice-alerts`

Generates multilingual voice alert assets from a current or upcoming citizen alert feed item. Voice is deliberately excluded from the generic alert delivery endpoint; callers must generate and approve a voice asset before delivery.

```json
{
  "alertId": "alert_feed_current_flood",
  "languages": ["en", "tw", "ga", "ee", "dag", "ha"],
  "workflowRequestedBy": "dispatcher_001",
  "source": "tts_sandbox"
}
```

Response:

```json
{
  "asset": {
    "id": "voice_alert_000001",
    "alertId": "alert_feed_current_flood",
    "alertTitle": "Severe flood warning",
    "hazardType": "flood",
    "severity": "severe_warning",
    "targetLabel": "Accra Metro and Tema",
    "status": "generated",
    "reviewStatus": "pending_review",
    "source": "tts_sandbox",
    "workflowRequestedBy": "dispatcher_001",
    "variants": [
      {
        "id": "voice_variant_000001",
        "language": "en",
        "locale": "en-GH",
        "voiceName": "nadaa-english-sandbox",
        "messageText": "NADAA alert. Severe flood warning. Area: Accra Metro and Tema. Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed. Call 112 if life is in danger.",
        "audioUrl": "voice://tts_sandbox/alert_feed_current_flood/en.mp3",
        "durationSeconds": 16,
        "status": "generated",
        "reviewStatus": "pending_review",
        "accessibilityChecks": [
          "plain_language",
          "action_oriented",
          "target_area_included",
          "includes_112_guidance",
          "low_literacy_length"
        ]
      }
    ]
  }
}
```

Supported languages are `en`, `tw`, `ga`, `ee`, `dag`, and `ha`. `source` may be `tts_sandbox` or `recorded_audio`.

`GET /api/v1/notifications/voice-alerts`

Lists generated voice alert assets, latest first.

`POST /api/v1/notifications/voice-alerts/{id}/review`

Approves or rejects all variants by default, or only the listed `languages`.

```json
{
  "action": "approve",
  "reviewer": "nadmo_voice_reviewer",
  "note": "Checked language, 112 guidance, and low-literacy script.",
  "languages": ["en", "tw"]
}
```

`POST /api/v1/notifications/voice-alerts/{id}/deliver`

Delivers an approved voice alert asset and writes normal notification delivery logs with `channel=voice`, `voiceAssetId`, `language`, and `audioUrl`.

```json
{
  "recipients": [
    {
      "phone": "+233200000000",
      "language": "en"
    },
    {
      "recipientId": "usr_voice_002",
      "phone": "+233200000001",
      "language": "tw"
    }
  ]
}
```

Response:

```json
{
  "attempts": [
    {
      "id": "delivery_000001",
      "alertId": "alert_feed_current_flood",
      "alertTitle": "Severe flood warning",
      "channel": "voice",
      "provider": "mock_voice",
      "recipientRef": "phone:...0000",
      "status": "delivered",
      "voiceAssetId": "voice_alert_000001",
      "language": "en",
      "audioUrl": "voice://tts_sandbox/alert_feed_current_flood/en.mp3",
      "attemptedAt": "2026-07-06T12:00:00Z"
    }
  ]
}
```

`POST /api/v1/notifications/ussd`

Accepts a sandbox/provider USSD session webhook. The MVP-compatible JSON contract keeps provider adapters isolated while still modelling the menu tree required for Phase 2 inclusive access.

```json
{
  "sessionId": "ussd_001",
  "phone": "+233200000000",
  "serviceCode": "*920#",
  "text": "1*2*1*3",
  "language": "en",
  "profileId": "usr_...",
  "linkProfile": true,
  "location": {
    "lat": 5.579,
    "lng": -0.212
  }
}
```

USSD menu path:

- Empty `text`: language menu.
- `1`: English main menu.
- `1*1`: current alerts.
- `1*2`: report emergency hazard menu.
- `1*2*1`: flood report urgency menu.
- `1*2*1*3`: submit high-urgency flood report.
- `1*3`: shelter lookup summary.
- `1*4`: 112 guidance.

Response:

```json
{
  "sessionId": "ussd_001",
  "action": "end",
  "message": "NADAA report received: access_report_000001. Call 112 if life is in immediate danger.",
  "language": "en",
  "log": {
    "id": "access_000001",
    "channel": "ussd",
    "provider": "ussd_sandbox",
    "phoneRef": "phone:...0000",
    "profileId": "usr_...",
    "linkedProfile": true,
    "language": "en",
    "intent": "report_emergency",
    "status": "queued",
    "createdAt": "2026-07-07T12:00:00Z"
  },
  "report": {
    "id": "access_report_000001",
    "channel": "ussd",
    "type": "flood",
    "urgency": "high",
    "description": "USSD emergency report: flood with high urgency.",
    "phoneRef": "phone:...0000",
    "linkedProfile": true,
    "status": "queued"
  }
}
```

`POST /api/v1/notifications/sms/inbound`

Accepts inbound SMS commands from a sandbox/provider adapter.

Supported commands:

- `ALERTS`: current alert summary.
- `SHELTER`: nearby shelter guidance.
- `HELP` or `112`: emergency-call guidance.
- `REPORT FLOOD HIGH <location/details>`: basic incident report.

Reports are mapped into incident-service when `NADAA_INCIDENT_SERVICE_URL` is configured. Otherwise the notification service stores a queued inclusive-access report for manual/provider handoff. Raw phone numbers are not stored in access logs; logs retain masked `phoneRef`. `profileId` is retained only when `linkProfile` is true.

Provider adapters can send `providerError` to record failed inbound delivery or signature checks without creating a report.

Runtime visibility: notification-service emits structured `INFO`, `WARN`, and `ERROR` logs for SMS/USSD/WhatsApp request receipt, voice asset generation/review/delivery, validation failures, menu transitions, provider errors, report creation, incident-service handoff, queue fallback, and inclusive access log/report persistence. Logs use masked `phoneRef`, command names, path depth, counts, statuses, and IDs; raw phone numbers, full message bodies, and full report descriptions should not be logged.

`POST /api/v1/notifications/whatsapp/webhook`

Accepts WhatsApp Business API or sandbox adapter messages. `/api/v1/notifications/whatsapp/inbound` is also supported for adapters that use an inbound naming convention.

```json
{
  "from": "+233200000000",
  "body": "REPORT FLOOD HIGH water entering homes near Odaw",
  "language": "en",
  "provider": "whatsapp_sandbox",
  "providerMessageId": "wamid.demo",
  "profileId": "usr_...",
  "linkProfile": true,
  "location": {
    "lat": 5.579,
    "lng": -0.212
  },
  "media": [
    {
      "id": "wa_media_001",
      "contentType": "image/jpeg"
    }
  ]
}
```

Supported commands:

- `ALERTS`: authority-approved current alert summary from notification-service alert feed.
- `RISK`: location-aware risk guidance using the current alert signal until a dedicated risk-service chat adapter is added.
- `REPORT`: starts a multi-message incident report flow.
- `REPORT FLOOD HIGH <location/details>`: submits a direct incident report.
- `SHELTER`: nearby shelter guidance.
- `GUIDE FLOOD`: emergency guide snippet.
- `HELP` or `112`: emergency-call guidance.

Conversation state supports incomplete reports across messages: `awaiting_report_hazard`, `awaiting_report_urgency`, and `awaiting_report_location`. Location pins and media IDs are accepted during the final report step and included in the optional incident-service handoff. WhatsApp transcripts are stored as privacy-safe summaries with command/length, direction, intent, state, masked `phoneRef`, media count, and a 90-day retention timestamp; raw message text, raw phone numbers, and captions should not be logged.

`GET /api/v1/notifications/access-logs?channel=sms&intent=report_emergency`

Returns inclusive access logs for SMS, USSD, and WhatsApp sessions. Supported filters are `channel=sms|ussd|whatsapp`, `intent=language_menu|main_menu|current_alerts|report_emergency|risk_check|emergency_guides|shelter_lookup|guidance_112|provider_error|invalid_selection`, and `status=handled|failed|queued|submitted`.

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
- The citizen web app stores offline-available guides in browser storage and registers a small service worker for the app shell and guide responses.

`GET /api/v1/shelters/nearby?lat=5.6037&lng=-0.1870`

Returns nearby shelters and recovery support locations sorted by distance.

```json
{
  "shelters": [
    {
      "id": "00000000-0000-0000-0000-000000000301",
      "name": "Accra Metro Assembly Shelter",
      "type": "evacuation_shelter",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "address": "Accra Metropolitan Assembly Hall",
      "location": {
        "lat": 5.56,
        "lng": -0.2
      },
      "capacity": 450,
      "currentOccupancy": 116,
      "status": "open",
      "contact": "112",
      "facilities": ["water", "first_aid", "accessible_entry"],
      "distanceMeters": 0,
      "updatedAt": "2026-07-06T12:00:00Z"
    }
  ],
  "recoverySupport": [
    {
      "id": "recovery_ama_relief_001",
      "name": "AMA Relief Distribution Point",
      "type": "relief_point",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "address": "Independence Avenue recovery desk",
      "location": {
        "lat": 5.558,
        "lng": -0.197
      },
      "contact": "112",
      "services": ["food", "water", "blankets"],
      "hours": "08:00-20:00",
      "status": "open",
      "distanceMeters": 420,
      "updatedAt": "2026-07-06T12:00:00Z"
    }
  ],
  "generatedAt": "2026-07-06T12:00:00Z"
}
```

`GET /api/v1/recovery-support/nearby?lat=5.6037&lng=-0.1870`

Returns only recovery support locations for relief distribution, medical support, recovery registration, water points, and family reunification.

`GET /api/v1/shelters`

Returns all shelter records for command console capacity views.

`PATCH /api/v1/shelters/{id}/occupancy`

Authority only. Requires authority actor, role, agency, MFA-completed, and request-id headers.

```json
{
  "currentOccupancy": 116,
  "capacity": 450,
  "status": "open",
  "notes": "Capacity confirmed by district shelter desk."
}
```

Allowed update roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`. `currentOccupancy` cannot exceed `capacity`. `status` must be `open`, `full`, `closed`, or `unknown`.

`GET /api/v1/relief-points?status=open&type=food&lat=5.5600&lng=-0.2000&radius=10000&limit=12`

Returns relief distribution points for citizen and authority views. Results can be filtered by status, relief type, nearby radius, bounding box, and limit. When coordinates are supplied, responses include `distanceMeters` and are sorted by operational status, distance, and name.

```json
{
  "reliefPoints": [
    {
      "id": "relief_ama_food_001",
      "name": "AMA Central Food Distribution",
      "type": "food",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "address": "Independence Avenue recovery desk",
      "location": { "lat": 5.558, "lng": -0.197 },
      "contact": "112",
      "operatingHours": "08:00-20:00",
      "eligibility": "Households affected by verified flooding.",
      "schedule": "Daily while stocks last",
      "stockCategories": [
        {
          "category": "rice_kg",
          "quantity": 420,
          "unit": "kg",
          "lastUpdated": "2026-07-07T10:10:00Z"
        }
      ],
      "status": "open",
      "source": "manual",
      "sourceRef": "district-relief-desk",
      "updatedBy": "district_officer_ama",
      "updatedAt": "2026-07-07T10:10:00Z"
    }
  ],
  "generatedAt": "2026-07-07T10:17:00Z"
}
```

`GET /api/v1/relief-points/nearby?lat=5.5600&lng=-0.2000`

Returns nearby open or limited relief points sorted by distance. Citizen web uses this endpoint to show food, water, medical, hygiene, blanket, cash, and mixed relief points with stock, hours, schedule, contact, and eligibility notes.

`POST /api/v1/relief-points`

Authority only. Requires authority actor, role, agency, MFA-completed, and request-id headers.

```json
{
  "name": "Smoke Relief Distribution Point",
  "type": "mixed",
  "region": "Greater Accra",
  "district": "Accra Metropolitan",
  "address": "Smoke Test Relief Desk",
  "location": { "lat": 5.552, "lng": -0.203 },
  "contact": "112",
  "operatingHours": "08:00-17:00",
  "eligibility": "Affected households.",
  "schedule": "Daily",
  "stockCategories": [
    { "category": "rice_kg", "quantity": 90, "unit": "kg" },
    { "category": "water_bottles", "quantity": 240, "unit": "bottles" }
  ],
  "status": "open",
  "sourceRef": "district-relief-desk"
}
```

`source` is set to `manual` by the service for authority-created points.

`PATCH /api/v1/relief-points/{id}`

Authority only. Updates name, type, location, contact, operating hours, eligibility, schedule, status, source reference, and stock categories. Stock changes write a stock-history entry with `changedBy`, `changedAt`, and the full stock snapshot.

`GET /api/v1/relief-points/{id}/stock-history`

Returns the latest stock snapshots for an authority relief point detail panel.

```json
{
  "reliefPointId": "relief_ama_food_001",
  "history": [
    {
      "changedBy": "district_officer_ama",
      "changedAt": "2026-07-07T10:10:00Z",
      "stockCategories": [
        {
          "category": "rice_kg",
          "quantity": 420,
          "unit": "kg",
          "lastUpdated": "2026-07-07T10:10:00Z"
        }
      ]
    }
  ]
}
```

The shelter service emits `INFO` logs for relief point list, nearby, create, update, and stock-history reads; `WARN` logs for invalid JSON, unauthorized authority context, failed validation, and missing records; and `ERROR` logs for response encoding failures.

`GET /api/v1/aid-requests?category=hygiene&priority=high&limit=20`

Returns public/partner donation and aid needs that have been approved for donor visibility. Authority users can add `includePrivate=true` with the standard authority headers to see pending, paused, rejected, closed, and partner-only needs.

```json
{
  "aidRequests": [
    {
      "id": "aid_ama_hygiene_001",
      "title": "Hygiene kits for displaced households",
      "category": "hygiene",
      "priority": "high",
      "status": "open",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "location": { "lat": 5.56, "lng": -0.2 },
      "receivingOrganization": "AMA Central Food Distribution",
      "quantityNeeded": 300,
      "quantityUnit": "kits",
      "quantityPledged": 80,
      "visibility": "public",
      "pledges": []
    }
  ],
  "generatedAt": "2026-07-07T13:00:00Z"
}
```

`POST /api/v1/aid-requests`

Authority only. Creates an aid need in `pending_review` so agencies can verify the receiving organization, category, contact, quantity, and location before public listing.

`PATCH /api/v1/aid-requests/{id}/review`

Authority only. Approves, opens, pauses, closes, or rejects an aid need. `approvalNotes` are required when approving. `antiFraudNotes` capture verification context without changing incident state.

`POST /api/v1/aid-requests/{id}/pledges`

Public/partner endpoint. Donors can pledge support only against approved/open aid needs. The pledge starts as `pledged` with `reviewStatus=pending_review` so agencies can verify the donor before counting it as cleared.

`GET /api/v1/aid-requests/{id}/pledges`

Authority only. Returns pledges for agency review.

`PATCH /api/v1/aid-requests/{id}/pledges/{pledgeId}/review`

Authority only. Updates pledge `status`, `reviewStatus`, and `fraudReviewNotes`.

`GET /api/v1/aid-requests/report.csv`

Authority only. Exports aid needs, pledge totals, pledge counts, priorities, districts, and needed-by dates for donor coordination reports.

The shelter service emits `INFO` logs for aid request list/create/review/export and pledge create/list/review success; `WARN` logs for invalid JSON, unsupported filters, unauthorized authority context, validation failures, pending request pledge attempts, and missing records; and `ERROR` logs for response encoding or CSV write failures. Logs must not include raw donor phone numbers, private donor notes, or sensitive beneficiary details.

### Missing Persons And Family Reunification

Base URL: `http://localhost:8101/api/v1`

`POST /missing-persons`

Public intake. Creates a private `pending_review` record with reporter contact details, consent flags, last-seen location/time, optional photo URL, optional related incident ID, and person description. Public visibility is never enabled during intake.

`GET /missing-persons?q=kojo&district=Accra`

Public approved search. Returns only records with `reviewStatus=approved`, `publicVisibility=public`, and an active/located status. Reporter contact details, review notes, and closure notes are never returned.

`GET /missing-persons/{id}`

Public approved lookup. Unapproved, private, rejected, closed, and reunited records return `404` to avoid leaking sensitive case state.

`GET /authority/missing-persons`

Authority only. Returns full sensitive records, including reporter contact, review notes, consent flags, and closure metadata.

`PATCH /authority/missing-persons/{id}/review`

Authority only. Supports `decision=approve_public`, `approve_private`, or `reject`. `approve_public` requires `publicSummary`; `reject` requires `reviewNotes`.

`PATCH /authority/missing-persons/{id}/close`

Authority only. Supports `closureType` values `reunited`, `located_safe`, `duplicate`, `withdrawn`, `deceased`, and `other`. Closure moves the record back to private visibility and writes an audit entry.

`GET /authority/missing-persons/{id}/audit`

Authority only. Returns create, review, and closure audit events.

The missing-person-service emits `INFO` logs for public search, intake, authority list, review, closure, and audit reads; `WARN` logs for invalid JSON, validation failures, missing/unauthorized authority context, and missing records; and `ERROR` logs for response encoding or server shutdown failures. Logs must not include reporter phone numbers, full descriptions, closure notes, or private review notes.

`GET /api/v1/hospitals/capacity?lat=5.5600&lng=-0.2000&service=emergency&emergencyCapacity=available&minAvailableBeds=10&includeStale=false`

Returns nearby hospital and emergency facility capacity sorted by freshness, distance, capacity status, and available beds. Dispatchers can omit coordinates for a national list or filter by service, capacity status, minimum available beds, stale-data visibility, and `limit`.

```json
{
  "facilities": [
    {
      "id": "hospital_001",
      "name": "Korle Bu Teaching Hospital",
      "type": "teaching_hospital",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "address": "Korle Bu emergency entrance",
      "location": { "lat": 5.536, "lng": -0.227 },
      "contact": "0302665401",
      "services": [
        "emergency",
        "trauma",
        "icu",
        "maternity",
        "pediatric",
        "oxygen"
      ],
      "totalBeds": 820,
      "availableBeds": 46,
      "icuBedsAvailable": 4,
      "maternityBedsAvailable": 9,
      "pediatricBedsAvailable": 5,
      "isolationBedsAvailable": 3,
      "emergencyCapacity": "available",
      "emergencyUnitStatus": "open",
      "ambulancesAvailable": 3,
      "oxygenAvailable": true,
      "source": "manual",
      "sourceRef": "hospital-capacity-feed",
      "updatedBy": "hospital_korle_bu_desk",
      "updatedAt": "2026-07-07T10:15:00Z",
      "distanceMeters": 4200,
      "stale": false
    }
  ],
  "generatedAt": "2026-07-07T10:17:00Z",
  "staleThresholdMinutes": 30
}
```

Rules:

- `service` matches normalized facility service tags such as `emergency`, `trauma`, `icu`, `maternity`, `pediatric`, `ambulance`, and `oxygen`.
- `emergencyCapacity` must be `available`, `limited`, `full`, `offline`, or `unknown`.
- `includeStale=false` hides facilities whose `updatedAt` is older than the stale threshold.
- Stale records remain visible by default with `stale=true` and `staleReason` so dispatchers know when to call and confirm.
- `lat` and `lng` must be provided together and must be valid coordinates.

`PATCH /api/v1/hospitals/{id}/capacity`

Authority only. Requires authority actor, role, agency, MFA-completed, and request-id headers. Allowed update roles are `system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, and `dispatcher`.

```json
{
  "availableBeds": 37,
  "icuBedsAvailable": 3,
  "emergencyCapacity": "available",
  "emergencyUnitStatus": "open",
  "ambulancesAvailable": 2,
  "oxygenAvailable": true,
  "notes": "Manual confirmation from hospital desk.",
  "source": "manual",
  "sourceRef": "dispatcher-call-20260707"
}
```

Accepted updates stamp `updatedAt`, `updatedBy`, `source`, and `sourceRef`. Bed and ambulance counts must be zero or greater, and `availableBeds` cannot exceed `totalBeds`.

`POST /api/v1/hospitals/capacity/imports/fixture`

Authority only. Runs the local fixture adapter for integration development. If `records` is omitted, the service imports the built-in fixture feed.

```json
{
  "source": "fixture_adapter",
  "sourceRef": "hospital-capacity-feed",
  "records": [
    {
      "facilityId": "hospital_001",
      "availableBeds": 38,
      "icuBedsAvailable": 3,
      "emergencyCapacity": "available",
      "emergencyUnitStatus": "open",
      "ambulancesAvailable": 2,
      "oxygenAvailable": true,
      "notes": "Fixture adapter update from hospital-capacity-feed."
    }
  ]
}
```

The shelter service emits `INFO` logs for shelter, recovery support, relief point, and hospital capacity list/update/import success; `WARN` logs for invalid input, missing records, stale filters, or unauthorized workflow failures; and `ERROR` logs for response encoding failures.

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

`POST /api/v1/integrations/weather-hydrology/import-jobs`

Starts a weather/hydrology fixture import into the NADAA observation store and returns `202 Accepted`.

```json
{
  "metric": "rainfall_mm",
  "requestedBy": "scheduler",
  "correlationId": "import-20260706-001"
}
```

Response:

```json
{
  "id": "import_manual_20260706120000_001",
  "adapterId": "mock-weather-hydrology-adapter",
  "source": "mock-weather-hydrology-adapter",
  "metric": "rainfall_mm",
  "status": "succeeded",
  "trigger": "manual",
  "attempts": 1,
  "retryable": true,
  "importedCount": 2,
  "failedCount": 0,
  "message": "Imported 2 weather/hydrology observations."
}
```

Rules:

- `metric` may be omitted, `rainfall_mm`, or `water_level_m`.
- Imported records preserve source, station, observed timestamp, validity window, point location, normalized rainfall/water-level fields, metadata, source record, and `weather_observations` storage target.
- A failed job keeps `status=failed`, `retryable=true`, `error`, `failedCount`, and `nextRetryAt`.
- `POST /api/v1/integrations/weather-hydrology/import-jobs/{id}/retry` retries a failed job.
- `GET /api/v1/integrations/weather-hydrology/import-jobs?status=failed` lists import status logs.
- `GET /api/v1/integrations/weather-hydrology/observations?metric=rainfall_mm` lists imported observations.
- The service can run a scheduled importer when `NADAA_IMPORT_SCHEDULER_ENABLED=true`; set `NADAA_IMPORT_SCHEDULER_INTERVAL` to a Go duration such as `15m`.

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

`POST /api/v1/ml/flood/predictions`

```json
{
  "location": { "lat": 5.6037, "lng": -0.187 },
  "requestedBy": "risk-service",
  "correlationId": "risk_5.6037_-0.1870"
}
```

Response:

```json
{
  "prediction": {
    "id": "pred_grid-accra-north-002",
    "modelVersion": "flood-logistic-baseline-0.1.0",
    "hazardType": "flood",
    "predictionTime": "2026-07-06T10:00:00Z",
    "targetTime": "2026-07-06T12:00:00Z",
    "cellId": "grid-accra-north-002",
    "region": "Greater Accra",
    "district": "Accra Metropolitan",
    "community": "Accra North",
    "location": { "lat": 5.6037, "lng": -0.187 },
    "distanceMeters": 0,
    "probability": 0.9828,
    "severity": "high",
    "expectedOnset": "24_to_48h",
    "confidence": "medium",
    "explanationFactors": [
      {
        "feature": "vulnerable_population_pct",
        "label": "vulnerable population share",
        "value": 21,
        "contribution": 0.8183,
        "direction": "increases_risk"
      }
    ],
    "inputFeatureSetVersion": "flood-risk-features.v1",
    "humanReviewRequired": true,
    "autoPublishAllowed": false,
    "source": "baseline_fixture_model"
  },
  "log": {
    "id": "ml_log_20260706100000_grid_accra_north_002",
    "predictionId": "pred_grid-accra-north-002",
    "modelVersion": "flood-logistic-baseline-0.1.0",
    "inputFeatureSetVersion": "flood-risk-features.v1",
    "requestedBy": "risk-service",
    "correlationId": "risk_5.6037_-0.1870",
    "location": { "lat": 5.6037, "lng": -0.187 },
    "storageTarget": "ml_predictions",
    "humanReviewRequired": true,
    "autoPublishAllowed": false,
    "createdAt": "2026-07-06T10:01:00Z"
  },
  "safety": {
    "humanReviewRequired": true,
    "autoPublishAllowed": false,
    "message": "Model output is decision support only and cannot publish alerts without authority review and approval."
  }
}
```

`GET /api/v1/ml/prediction-logs`

Returns in-memory MVP prediction log records. Each record is aligned to the `ml_predictions` storage target and includes model version and input feature set version.

### Flood Simulations

`POST /api/v1/ml/flood/simulations`

Runs a deterministic flood simulation using the baseline logistic model and the NADAA-070 feature grid. Rainfall and water-level overrides are applied linearly across the requested duration.

```json
{
  "name": "Accra +50 mm rainfall scenario",
  "rainfallMmOverride": 50,
  "waterLevelTrendCmOverride": 10,
  "durationHours": 6,
  "timeStepHours": 1
}
```

Response:

```json
{
  "simulation": {
    "id": "sim_20260709100000",
    "reference": "FS-2026-00001",
    "name": "Accra +50 mm rainfall scenario",
    "status": "completed",
    "scenario": {
      "rainfallMmOverride": 50,
      "waterLevelTrendCmOverride": 10,
      "durationHours": 6,
      "timeStepHours": 1
    },
    "frames": [
      {
        "targetTime": "2026-07-09T11:00:00Z",
        "cells": [
          {
            "cellId": "grid-accra-central-001",
            "region": "Greater Accra",
            "district": "Accra Metropolitan",
            "community": "Accra Central",
            "geometry": {
              "type": "Polygon",
              "coordinates": [
                [
                  [-0.21, 5.55],
                  [-0.19, 5.55],
                  [-0.19, 5.57],
                  [-0.21, 5.57],
                  [-0.21, 5.55]
                ]
              ]
            },
            "probability": 0.94,
            "severity": "severe",
            "depthBand": "> 1.2 m",
            "confidence": "medium",
            "explanationFactors": []
          }
        ]
      }
    ],
    "assumptions": [
      "Simulation applies user rainfall and water-level overrides linearly across the requested time window."
    ],
    "limitations": [
      "The fixture training set has only five seed-aligned rows and is suitable for product/API integration, not production accuracy claims."
    ],
    "modelVersion": "flood-logistic-baseline-0.1.0",
    "featureSetVersion": "flood-risk-features.v1",
    "createdAt": "2026-07-09T10:00:00Z",
    "updatedAt": "2026-07-09T10:00:01Z",
    "safety": {
      "humanReviewRequired": true,
      "autoPublishAllowed": false,
      "message": "Simulation output is decision support only and cannot publish alerts without authority review and approval."
    }
  }
}
```

`GET /api/v1/ml/flood/simulations`

Returns the list of simulation jobs sorted newest first.

`GET /api/v1/ml/flood/simulations/{id}`

Returns a single simulation job including all frames.

### ML-Reviewed Alert Drafts

`POST /api/v1/alerts`

NADAA-073 uses the standard alert creation endpoint for reviewed ML predictions. Include `sourcePrediction` to preserve the prediction, model, feature set, probability, severity, confidence, and review note in the draft and `alert.created` audit snapshot.

```json
{
  "title": "ML reviewed Accra Central flood alert",
  "hazardType": "flood",
  "severity": "severe_warning",
  "message": "Reviewed ML prediction estimates 99.9% flood probability for Accra Central.",
  "target": {
    "type": "custom",
    "ids": ["grid-accra-central-001"],
    "label": "Accra Central prediction cell",
    "geometry": {
      "type": "Polygon",
      "coordinates": [
        [
          [-0.21, 5.55],
          [-0.19, 5.55],
          [-0.19, 5.57],
          [-0.21, 5.57],
          [-0.21, 5.55]
        ]
      ]
    }
  },
  "startsAt": "2026-07-06T12:00:00Z",
  "expiresAt": "2026-07-07T00:00:00Z",
  "recommendedAction": "Prepare to evacuate if instructed by authorities and avoid low-lying roads.",
  "evacuationRequired": true,
  "shelterIds": ["00000000-0000-0000-0000-000000000301"],
  "sourcePrediction": {
    "predictionId": "pred_grid-accra-central-001",
    "predictionLogId": "ml_log_20260706100000_grid_accra_central_001",
    "modelVersion": "flood-logistic-baseline-0.1.0",
    "inputFeatureSetVersion": "flood-risk-features.v1",
    "probability": 0.9993,
    "severity": "severe",
    "confidence": "medium",
    "humanReviewRequired": true,
    "autoPublishAllowed": false,
    "reviewNote": "Dispatcher reviewed explanation factors."
  }
}
```

`sourcePrediction.humanReviewRequired` must be `true`, and `sourcePrediction.autoPublishAllowed` must be `false`. The created alert has `status: "draft"` and must still be submitted and approved before any public release. Public alert list responses omit `sourcePrediction`; authority reads and audit logs retain it for traceability.

## Phase 2 API Areas

- SMS/USSD inbound webhook implemented in notification-service.
- WhatsApp inbound webhook implemented in notification-service.
- Voice alert asset generation, review, and approved delivery implemented in notification-service.
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

### Computer Vision Image Verification

`POST /api/v1/cv/analyze`

Accepts an image ID/name/URL and returns computer vision evidence labels with confidence scores.

```json
{
  "imageId": "media_flood_photo_001",
  "imageName": "flooded-road.jpg"
}
```

Response:

```json
{
  "result": {
    "id": "cv_20260706184200_media_flood_photo_001",
    "imageId": "media_flood_photo_001",
    "labels": [
      { "label": "flood_evidence", "confidence": 0.92 },
      { "label": "water_surface", "confidence": 0.88 },
      { "label": "submerged_road", "confidence": 0.76 }
    ],
    "modelVersion": "cv-mock-rule-engine-0.1.0",
    "limitations": "This is a deterministic rule-based mock engine...",
    "humanReviewRequired": false,
    "createdAt": "2026-07-06T18:42:00Z",
    "reviewStatus": "pending"
  },
  "safety": {
    "humanReviewRequired": false,
    "autoPublishAllowed": false,
    "message": "CV output is decision support only..."
  }
}
```

Rules:

- `imageId` is required.
- The mock engine uses filename hints: "flood" → flood_evidence, "fire" → fire_evidence, "injured" → sensitive, etc.
- `humanReviewRequired` is true when any confidence < 0.7 or the label is `sensitive`.
- Results are cached by `imageId`; repeated analyzes return the cached result.
- CV output is decision-support only and cannot trigger alerts or public actions without authority review.

`GET /api/v1/cv/results/{imageId}`

Retrieves a cached CV analysis result by image ID. Returns 404 if not found.

`GET /api/v1/cv/results`

Lists all cached CV analysis results.

## Resource Planning

Served by `ml-service` (default `:8094`). Demand forecasts and staging suggestions are produced by the deterministic `resource-forecast-rules-0.1.0` model over the NADAA-070 flood-risk feature grid (historical reports, rainfall forecast, composite risk, and vulnerable population). All endpoints are read-only decision support; there are no deployment-action endpoints.

### List Demand Forecasts

`GET /api/v1/forecasts`

Query parameters:

- `region` (optional) — filter by region name

Response:

```json
{
  "forecasts": [
    {
      "id": "forecast_001",
      "region": "Greater Accra",
      "district": "Accra Metropolitan",
      "timeWindowStart": "2026-07-09T00:00:00Z",
      "timeWindowEnd": "2026-07-10T00:00:00Z",
      "predictedIncidentCount": 12,
      "hazardType": "flood",
      "confidence": "medium",
      "confidenceScore": 0.72,
      "factors": [
        {
          "name": "historical_incidents",
          "label": "Historical flood reports (30d)",
          "value": 10,
          "weight": 0.35,
          "direction": "increases_demand"
        }
      ],
      "riskLevel": "high",
      "generatedAt": "2026-07-09T12:00:00Z"
    }
  ],
  "generatedAt": "2026-07-09T12:00:00Z"
}
```

### Forecast by Region

`GET /api/v1/forecasts/{region}`

Returns forecasts for a specific region.

Response:

```json
{
  "forecast": { ... },
  "forecasts": [ ... ],
  "generatedAt": "2026-07-09T12:00:00Z"
}
```

### List Staging Suggestions

`GET /api/v1/staging-suggestions`

Query parameters:

- `agencyType` (optional) — filter by agency type (e.g., `fire`, `ambulance`, `nadmo`)

Response:

```json
{
  "suggestions": [
    {
      "id": "staging_001",
      "location": { "lat": 5.6037, "lng": -0.187 },
      "locationLabel": "Accra Central Fire Station",
      "agencyType": "fire",
      "reason": "High predicted fire incidents in Accra Metropolitan",
      "confidence": "medium",
      "confidenceScore": 0.65,
      "operationalConstraints": [
        "Road congestion during rush hours",
        "Limited water tanker availability after 22:00"
      ],
      "recommendedUnits": 3,
      "radiusMeters": 5000,
      "generatedAt": "2026-07-09T12:00:00Z"
    }
  ],
  "generatedAt": "2026-07-09T12:00:00Z"
}
```

### Compare Scenarios

`POST /api/v1/forecasts/compare`

Request body:

```json
{
  "region": "Greater Accra",
  "riskLevel": "severe",
  "historicalWeight": 1.5,
  "capacityFactor": 0.8,
  "timeWindowHours": 12
}
```

`region`, `riskLevel` (a minimum-severity threshold), and `hazardTypes` are scope filters applied to both the baseline and the adjusted scenario so their totals remain comparable. Only the levers `historicalWeight` and `timeWindowHours` differ between the two scenarios. `capacityFactor` is validated and echoed back for forward compatibility with capacity-aware staging but does not change demand forecast counts.

Response:

```json
{
  "scenarios": [
    {
      "name": "Current conditions",
      "parameters": { "region": "Greater Accra", "riskLevel": "severe" },
      "forecasts": [ ... ],
      "summary": {
        "totalPredictedIncidents": 12,
        "averageConfidenceScore": 0.72,
        "highestRiskRegion": "Greater Accra",
        "highestRiskHazard": "flood"
      }
    },
    {
      "name": "Adjusted scenario",
      "parameters": { ... },
      "forecasts": [ ... ],
      "summary": { ... }
    }
  ],
  "generatedAt": "2026-07-09T12:00:00Z"
}
```

Rules:

- All forecasts include confidence levels and operational constraints.
- Predictions are decision-support only; no automatic deployment orders are generated.
- Agency leadership retains final deployment authority.
