# Integrations

NADAA integrations are contract-first. Official APIs may arrive gradually, so MVP adapters must keep partner-specific authentication, payload shape, cadence, retries, and failure handling isolated from incident response and alert approval workflows.

## Contract Matrix

| ID                           | Partner                         | Domain            | Direction     | Data Owner                            | Cadence                                                 | Auth                           | Failure Behavior                                                       |
| ---------------------------- | ------------------------------- | ----------------- | ------------- | ------------------------------------- | ------------------------------------------------------- | ------------------------------ | ---------------------------------------------------------------------- |
| `gmet-rainfall-nowcast`      | Ghana Meteorological Agency     | Weather           | Inbound       | GMet                                  | Every 15 minutes during watch/warning periods           | API key + signed source header | Retry, dead-letter, continue manual risk review                        |
| `hydro-water-level-feed`     | Ghana Hydrological Authority    | Hydrology         | Inbound       | Ghana Hydrological Authority          | Every 15 minutes during rainy season, hourly otherwise  | mTLS + source header           | Retry, dead-letter, continue manual risk review                        |
| `nadmo-incident-sync`        | NADMO National Operations       | Incident sync     | Outbound      | NADAA platform operator               | Near real time on verification, assignment, and closure | Signed webhook                 | Retry, then manual dispatcher handoff                                  |
| `nadmo-alert-sync`           | NADMO National Operations       | Alert sync        | Outbound      | NADAA platform operator               | Near real time after human approval                     | Signed webhook                 | Retry, public alert workflow remains authoritative                     |
| `police-road-closure-feed`   | Ghana Police Service            | Road closure      | Bidirectional | Ghana Police Service                  | On change with hourly reconciliation                    | Signed webhook                 | Import as source-attributed context; do not auto-route without review  |
| `fire-incident-dispatch`     | Ghana National Fire Service     | Incident sync     | Outbound      | NADAA platform operator               | Near real time for fire/rescue assignments              | Signed webhook                 | Dispatcher calls 112 and records manual handoff                        |
| `ambulance-medical-dispatch` | National Ambulance Service      | Incident sync     | Outbound      | NADAA platform operator               | Near real time for injury/medical assignments           | Signed webhook                 | Dispatcher calls 112 and keeps incident manual                         |
| `district-shelter-status`    | District Assemblies             | Shelter status    | Bidirectional | District Assembly or shelter operator | Every 30 minutes during response, daily otherwise       | API key                        | Treat updates as advisory until authorized confirmation                |
| `district-relief-inventory`  | District Assemblies/NGOs        | Relief inventory  | Bidirectional | District Assembly or relief operator  | Every 30 minutes during response, daily otherwise       | API key + source header        | Treat stock as advisory until an authorized agency update confirms it  |
| `hospital-capacity-feed`     | Hospitals and health facilities | Hospital capacity | Inbound       | Participating health facility         | Every 30 minutes during active incidents                | API key                        | Restrict visibility; retry without blocking dispatch                   |
| `utility-outage-feed`        | Utilities and power providers   | Utility outage    | Inbound       | Originating utility                   | On change with hourly reconciliation                    | Signed webhook                 | Enrich dispatcher context; never suppress citizen reports              |
| `sms-ussd-inclusive-access`  | SMS/USSD provider               | Citizen access    | Inbound       | NADAA platform operator               | Near real time for citizen requests and reports         | Signed webhook or API key      | Queue report locally, log provider error, and advise caller to use 112 |
| `whatsapp-emergency-chatbot` | WhatsApp Business API provider  | Citizen access    | Inbound       | NADAA platform operator               | Near real time for citizen requests, media, and reports | Signed webhook or API key      | Keep conversation state locally, queue report locally, and advise 112  |
| `voice-alert-provider`       | Voice call/TTS provider         | Citizen access    | Outbound      | NADAA platform operator               | Near real time after voice asset approval               | API key + signed callbacks     | Skip unapproved variants, log delivery status, and fall back to SMS    |

## Common Rules

- Every import must preserve `source`, `observedAt` or `updatedAt`, validity window, source license/usage constraints when known, and contact point.
- Every outbound sync must include a `correlationId` for idempotency and traceability.
- Adapter failures must be retryable and dead-lettered, but they must not block manual incident response or human-approved alerts.
- API keys, mTLS credentials, OAuth clients, webhook signing secrets, and SFTP credentials must live in environment secrets or a secret manager.
- Operationally sensitive data such as hospital capacity should be role-restricted and excluded from public exports.
- Relief inventory imports should preserve eligibility notes, schedule, stock category units, source references, and update actor context where available; beneficiary identity must never be included.
- SMS/USSD/WhatsApp provider adapters should normalize provider-specific payloads into notification-service webhooks, log `providerError` when signature validation or provider delivery fails, and avoid storing raw phone numbers, full message bodies, or media captions in runtime logs.
- WhatsApp adapters should preserve provider message IDs, location pins, and media IDs/URLs while allowing notification-service to store only privacy-safe transcript summaries with a 90-day retention timestamp.
- Voice alert adapters should send only approved `voiceAlertAsset` variants, retain provider message IDs in delivery logs, and avoid sending voice for expired or unreviewed alerts.

## Payload Examples

### Weather Observation

```json
{
  "source": "gmet-accra-nowcast",
  "observedAt": "2026-07-06T11:45:00Z",
  "validFrom": "2026-07-06T11:45:00Z",
  "validTo": "2026-07-06T12:15:00Z",
  "location": { "lat": 5.6037, "lng": -0.187 },
  "stationId": "GHA-ACC-RAIN-001",
  "rainfallMm": 34.2
}
```

### Hydrology Observation

```json
{
  "source": "hydro-odaw-level",
  "observedAt": "2026-07-06T11:48:00Z",
  "validFrom": "2026-07-06T11:48:00Z",
  "validTo": "2026-07-06T12:18:00Z",
  "location": { "lat": 5.575, "lng": -0.205 },
  "stationId": "GHA-ODAW-LVL-001",
  "waterLevelM": 1.76
}
```

### Weather/Hydrology Import Job

```json
{
  "adapterId": "mock-weather-hydrology-adapter",
  "metric": "rainfall_mm",
  "requestedBy": "scheduler",
  "correlationId": "import-20260706-001"
}
```

Import logs keep status, trigger (`manual`, `scheduled`, or `retry`), attempts, retryability, imported and failed counts, errors, and `nextRetryAt` when retry is possible. Imported observations are normalized for the `weather_observations` storage target while preserving source, station, timestamp, point location, validity window, metadata, and source record.

### Incident Sync

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

### Alert Sync

```json
{
  "type": "alert",
  "sourceId": "alert_001",
  "reference": "ALT-000001",
  "hazardType": "flood",
  "severity": "warning",
  "title": "Flood Watch",
  "message": "Avoid low-lying routes.",
  "targetLabel": "Accra Metropolitan",
  "targetAgencyIds": ["00000000-0000-0000-0000-000000000101"],
  "correlationId": "corr_alert_001"
}
```

### Road Closure

The integration-service `POST /api/v1/integrations/road-closures/imports` endpoint validates an inbound road closure record and forwards it to the `road-closure-service` at `NADAA_ROAD_CLOSURE_SERVICE_URL` (default `http://localhost:8095`). The record is also stored as an integration import for observability.

```json
{
  "source": "ghana-police",
  "sourceRef": "police-run-20260707-1000",
  "roadName": "Sample Market Road",
  "status": "active",
  "geometry": "LINESTRING(-0.20 5.56, -0.19 5.57)",
  "validFrom": "2026-07-06T12:00:00Z",
  "reason": "Flooding",
  "detour": "Use alternate bypass via Ring Road Central."
}
```

### Shelter Status

```json
{
  "shelterId": "00000000-0000-0000-0000-000000000301",
  "status": "open",
  "capacity": 450,
  "currentOccupancy": 116,
  "updatedAt": "2026-07-06T12:00:00Z"
}
```

### Relief Inventory

```json
{
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
    { "category": "rice_kg", "quantity": 420, "unit": "kg" },
    { "category": "water_sachets", "quantity": 1800, "unit": "sachets" }
  ],
  "status": "open",
  "source": "district-relief-inventory",
  "sourceRef": "ama-relief-20260707-1000"
}
```

Relief inventory adapters should submit records into shelter-service relief point create/update workflows through an authorized agency context. Stock changes create stock-history snapshots so agencies can see when quantities changed and who confirmed them.

### Hospital Capacity

```json
{
  "facilityId": "hospital_001",
  "source": "hospital-capacity-feed",
  "sourceRef": "feed-run-20260707-1015",
  "observedAt": "2026-07-07T10:15:00Z",
  "totalBeds": 820,
  "availableBeds": 14,
  "icuBedsAvailable": 2,
  "maternityBedsAvailable": 5,
  "pediatricBedsAvailable": 4,
  "isolationBedsAvailable": 1,
  "emergencyCapacity": "limited",
  "emergencyUnitStatus": "busy",
  "ambulancesAvailable": 1,
  "oxygenAvailable": true,
  "notes": "Capacity confirmed by hospital emergency desk."
}
```

### Utility Outage

```json
{
  "source": "utility-provider",
  "utilityType": "electricity",
  "status": "outage_reported",
  "area": "Accra Central sample zone",
  "validFrom": "2026-07-06T12:00:00Z",
  "hazardType": "electrical_hazard"
}
```
