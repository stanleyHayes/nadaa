# ML Service

Planned service for flood risk models, prediction serving, model metadata, explainability, simulation, incident triage, and image verification.

## Current Baseline Artifacts

NADAA-071 adds the first fixture-trained flood model artifacts:

- `data/flood-risk/models/baseline-logistic.v1.json` - logistic-regression model metadata, preprocessing, coefficients, output contract, and limitations.
- `data/flood-risk/models/sample-predictions.v1.json` - sample predictions for the five MVP feature cells.
- `data/flood-risk/models/evaluation.v1.json` - metrics, confusion matrix, calibration buckets, and review records.
- `data/flood-risk/models/evaluation-report.v1.md` - human-readable evaluation report.

Run:

```bash
pnpm ml:flood:train
pnpm validate:ml
```

The baseline is intentionally dependency-free and deterministic. It is useful for NADAA-072 prediction serving contracts and authority-review UI work, but it is not production-ready because it is trained on five development fixture rows.

## MVP Service

NADAA-072 adds the MVP HTTP service:

- `GET /healthz`
- `POST /api/v1/ml/flood/predictions`
- `GET /api/v1/ml/prediction-logs?limit=&offset=`
- `POST /api/v1/ml/flood/simulations`
- `GET /api/v1/ml/flood/simulations?limit=&offset=` and `GET /api/v1/ml/flood/simulations/{id}`
- `POST /api/v1/cv/analyze`
- `GET /api/v1/cv/results?limit=&offset=` and `GET /api/v1/cv/results/{imageId}`
- `PATCH /api/v1/cv/results/{id}/review`
- `GET /api/v1/forecasts` family and `GET /api/v1/staging-suggestions`

The service listens on `:8094` by default. Set `NADAA_ML_ADDR` to override the bind address and `NADAA_ML_MODEL_DIR` when model artifacts are mounted somewhere other than the repository defaults.

All endpoints except `GET /healthz` require either a verified NADAA bearer token (`Authorization: Bearer nadaa.<payload>.<sig>`, verified with `NADAA_AUTH_TOKEN_SECRET`) whose claims carry an agency/dispatcher role (`system_admin`, `agency_admin`, `nadmo_officer`, `district_officer`, `dispatcher` — verified tokens without one, such as citizen tokens, get `403`), or the `X-NADAA-Service-Token` header matching `NADAA_INTERNAL_SERVICE_TOKEN`. When `NADAA_INTERNAL_SERVICE_TOKEN` is unset the service-token path fails closed (the header is ignored) and a warning is logged at startup. Legacy `X-NADAA-Actor-*` mock actor headers are honored only when `NADAA_AUTH_ALLOW_MOCK_ACTORS=true`, and that setting is rejected at startup unless `NADAA_ENV=development`.

`PATCH /api/v1/cv/results/{id}/review` records a human review decision on a CV result (looked up by result ID or image ID): body `{"decision":"approved"|"rejected", "note"?}`, `200` returns the updated result (`{"result": {...}}` with `reviewStatus`, `reviewedBy`, `reviewedAt`), `400 invalid_decision` for other decisions, `404` for unknown IDs. The reviewer identity always comes from a verified agency bearer token — service tokens and mock headers are not accepted for review.

List endpoints paginate with `limit` (default 50, max 200) and `offset`, and report `total`, `limit`, and `offset`. In-memory collections are bounded with FIFO eviction: 500 prediction logs, 100 simulation runs, and 500 CV results. Simulation creation rejects non-positive `durationHours`/`timeStepHours` (previously silently defaulted) and names over 200 characters.

At startup the service verifies `checksums.sha256` in the model directory (sha256 of the model artifacts and the generated features file) and refuses to load on mismatch; set `NADAA_ML_SKIP_INTEGRITY_CHECK=true` to bypass. Regenerate the manifest after retraining or regenerating features:

```bash
cd data/flood-risk/models
shasum -a 256 baseline-logistic.v1.json evaluation-report.v1.md evaluation.v1.json sample-predictions.v1.json ../generated/features.v1.csv ../generated/features.v1.json ../generated/manifest.v1.json > checksums.sha256
```

Run locally:

```bash
cd services/ml-service
go run .
```

Run the smoke check:

```bash
pnpm smoke:ml
```

Risk-service can attach ML decision support when `NADAA_ML_API_URL` points to the ML API base URL, for example `http://127.0.0.1:8094/api/v1`.

Dispatcher-web can load the review panel from this service when `VITE_ML_API_URL` points to the same ML API base URL. The reviewed alert-draft smoke path is `pnpm smoke:ml-review` after starting both ML service and alert-service.

The service loads `baseline-logistic.v1.json`, accepts a location, returns the sample-prediction shape, records an in-memory prediction log aligned to the future `ml_predictions` table, and keeps prediction output human-reviewed. Public alerts must never be auto-published from model output.

Related stories:

- NADAA-070
- NADAA-071
- NADAA-072
- NADAA-073
- NADAA-150
- NADAA-151
- NADAA-152
