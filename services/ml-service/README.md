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
- `GET /api/v1/ml/prediction-logs`

The service listens on `:8094` by default. Set `NADAA_ML_ADDR` to override the bind address and `NADAA_ML_MODEL_DIR` when model artifacts are mounted somewhere other than the repository defaults.

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

The service loads `baseline-logistic.v1.json`, accepts a location, returns the sample-prediction shape, records an in-memory prediction log aligned to the future `ml_predictions` table, and keeps prediction output human-reviewed. Public alerts must never be auto-published from model output.

Related stories:

- NADAA-070
- NADAA-071
- NADAA-072
- NADAA-073
- NADAA-150
- NADAA-151
- NADAA-152
