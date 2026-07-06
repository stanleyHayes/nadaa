# Guide Service

The guide service owns emergency preparedness, response, and recovery guidance content.

Current NADAA-060 endpoints:

- `GET /healthz`
- `GET /api/v1/guides`

## Guide Lookup

`GET /api/v1/guides` supports:

- `hazard` - any NADAA hazard type, such as `flood`, `fire`, `road_crash`, `electrical_hazard`, `disease_outbreak`, or `other`.
- `stage` - `before`, `during`, `after`, or `recovery`.
- `language` - defaults to `en`; if a requested language has no exact match, the service falls back to English for the same filters.
- `offline` - `true` or `false`, filtering by offline availability.

The starter fixtures cover floods, fire safety, road crash response, electrical hazard safety, disease prevention, safe evacuation, emergency bag checklist, family emergency planning, and contacting 112. General preparedness topics use hazard type `other`.

## Editor And CMS Notes

Until a CMS lands, guide content should be updated in the database seed and service fixture together. Future CMS records should preserve the same model: hazard type, stage, title, body, language, offline availability, sort order, created timestamp, and updated timestamp.

Editors should keep instructions short, action-oriented, and safe for offline use. Emergency contact guidance should use 112 for life-threatening emergencies in Ghana.

## Run

```bash
go run .
```

The service listens on `:8086` by default. Override with `NADAA_GUIDE_ADDR`.

## Test

```bash
go test ./...
```

## Notes

The current implementation uses an in-memory store to lock in the API contract and content model for the offline-first citizen guidance UI. The citizen app caches offline-available guide records and serves a small service worker for the app shell and guide responses. Verify that integration with `pnpm smoke:citizen-guides` when guide-service and citizen-web are running.

PostGIS persistence and CMS publishing workflow can be added later without changing the lookup contract.
