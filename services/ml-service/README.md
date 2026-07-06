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

## Serving Handoff

The future serving endpoint should load `baseline-logistic.v1.json`, accept the NADAA-070 feature contract, return the sample-prediction shape, and keep prediction output human-reviewed. Public alerts must never be auto-published from model output.

Related stories:

- NADAA-070
- NADAA-071
- NADAA-072
- NADAA-073
- NADAA-150
- NADAA-151
- NADAA-152
