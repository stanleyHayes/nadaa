# ML

The first ML priority is flood risk prediction. ML output is decision support only and must remain human-reviewed before any public alert is sent.

## MVP Model Goal

Predict flood probability and expected severity for a location, district, community, or grid cell, then provide explanation factors that help NADMO/district officers decide whether to draft or approve an alert.

## Inputs

- Rainfall forecast.
- Historical rainfall.
- Drainage density.
- Elevation and slope.
- Soil type.
- Distance to rivers and drains.
- Land use.
- Historical flood reports.
- Dam spillage notices.
- Population density.
- Road network.
- Citizen reports.
- Satellite rainfall data.
- GMet alerts.
- Hydrological water-level data.

## Outputs

- Flood probability by grid cell or target area.
- Expected severity.
- Expected onset time where available.
- Affected communities.
- Recommended alert level.
- Explanation factors.
- Model version and confidence level.

## Model Stages

### Stage 0: Rule-Based Risk Score

Purpose: unblock MVP risk checker before enough official data is available.

Inputs:

- Known flood-prone area fixture.
- Rainfall fixture/import.
- Elevation or low-lying area fixture.
- Recent citizen reports.
- Nearby shelter and response facility fixtures for preparedness context.

Output:

- Low, moderate, high, severe, or emergency risk level.
- Recommended citizen actions and nearest response resources.

Current MVP implementation:

- `services/risk-service` serves `GET /api/v1/risk`.
- The baseline mirrors the development seed data with in-memory fixtures until service-level PostGIS persistence is added.
- Severe risk is returned inside the Accra flood fixture, high risk near the flood fixture and recent report, moderate risk near only recent reports, and low risk outside fixture coverage.

### NADAA-070 Feature Pipeline

The first repeatable feature pipeline is fixture-backed and versioned so Stage 1 model work can start without waiting for every official source integration.

Artifacts:

- `data/flood-risk/source-fixtures.v1.json` stores seed-aligned rainfall, elevation, slope, hydrology, land-use, population, historical report, and simplified geometry inputs.
- `data/flood-risk/feature-schema.v1.json` defines the 44-column feature contract.
- `data/flood-risk/generated/features.v1.json` stores feature rows with GeoJSON geometry.
- `data/flood-risk/generated/features.v1.csv` stores the same rows for notebooks and model experiments.
- `data/flood-risk/generated/manifest.v1.json` records row count, column count, validity window, limitations, and output checksums.

Commands:

```bash
pnpm features:flood
pnpm validate:features
```

Feature groups:

- Identity, time validity, source freshness, and geometry provenance.
- Rainfall: 24-hour observed, 72-hour observed, and 24-hour forecast aggregate.
- Terrain: elevation, slope, and low-lying flag.
- Hydrology: distance to drain, distance to river, water-level trend, and drainage density proxy.
- Exposure: land use, impervious surface, population density, and vulnerable population share.
- Historical signal: recent flood report count and days since latest flood report.
- Derived signals: rainfall intensity, exposure, drainage pressure, historical signal, composite rule score, and severity label.
- Missing-data flags for rainfall, terrain, hydrology, land use, population, historical reports, and geometry.

Candidate production sources to evaluate after the MVP fixture pass:

- Ghana Meteorological Agency rainfall nowcasts for official rainfall inputs.
- Ghana Hydrological Authority water-level feeds for official hydrology inputs.
- [NASA GPM IMERG](https://gpm.nasa.gov/data/imerg) for satellite rainfall gap filling.
- [NASA SRTM](https://www.earthdata.nasa.gov/data/instruments/srtm) for elevation and slope derivation.
- [OpenStreetMap](https://welcome.openstreetmap.org/working-with-osm-data/downloading-and-using/) for road, waterway, and drainage proxies where official geometry is unavailable.
- [WorldPop Ghana](https://hub.worldpop.org/geodata/country?iso3=GHA) for gridded population density and exposure features.

Limitations:

- Fixture values are development data and must not be treated as official Ghana hazard data.
- Rainfall aggregates combine mock integration observations with hand-authored development aggregates.
- Terrain, drainage, land-use, and population values are proxies until production source ingestion, licensing, and quality review are complete.
- District/community geometries are simplified bounding boxes and must be replaced with authoritative boundaries before production model training.
- Historical report counts are intentionally small and biased toward current MVP seed records.

### Stage 1: Baseline ML

Candidate models:

- Logistic regression.
- Random forest.
- XGBoost.

Current MVP baseline:

- `scripts/train-flood-risk-baseline.mjs` trains `flood-logistic-baseline-0.1.0` using deterministic logistic regression over `data/flood-risk/generated/features.v1.json`.
- `data/flood-risk/models/baseline-logistic.v1.json` stores model metadata, feature columns, preprocessing statistics, hyperparameters, coefficients, output contract, limitations, and training metrics.
- `data/flood-risk/models/sample-predictions.v1.json` stores five sample predictions with probability, severity, expected onset bucket, medium fixture confidence, and explanation factors.
- `data/flood-risk/models/evaluation.v1.json` stores precision, recall, F1, Brier score, confusion matrix, calibration buckets, and false-positive/false-negative review records.
- `data/flood-risk/models/evaluation-report.v1.md` is the human-readable evaluation report for authority and ML review.
- `scripts/validate-flood-risk-model.mjs` verifies the model, prediction, evaluation, and report artifacts.

Commands:

```bash
pnpm ml:flood:train
pnpm validate:ml
```

The MVP model predicts a binary high-or-severe flood risk probability, then maps that output into `low`, `moderate`, `high`, or `severe` with an additional conservative severe rule. Confidence is capped at `medium` for fixture-trained outputs because the current training set has five seed-aligned rows and no independent outcome history.

Minimum evaluation:

- Precision/recall for high-risk flood events.
- Calibration by probability band.
- False positive review.
- False negative review.
- District/community breakdown when data permits.

Current fixture evaluation:

- Scope: training fixture resubstitution only.
- Rows: 5.
- Positive high-risk labels: 3.
- Negative labels: 2.
- Precision/recall/F1 on the fixture: 1.0.
- Brier score on the fixture: 0.0002.
- False positives: 0.
- False negatives: 0.

These metrics prove artifact and contract wiring, not production readiness. The next data step is to add official rainfall/hydrology and independent historical outcome labels, then rerun evaluation with temporal or district holdout splits.

### Stage 2: Spatial And Temporal Improvement

Candidate approaches:

- Gradient boosting with geospatial features.
- Time-series models for rainfall and water-level trends.
- Drainage/road graph features.

### Stage 3: Advanced Simulation

Phase 3 only:

- Real-time flood simulation.
- Scenario modeling.
- Flood depth/severity bands where data permits.

## Feature Store Contract

Every feature output should include:

- `featureSetVersion`
- `source`
- `sourceUpdatedAt`
- `generatedAt`
- `geometry`
- `validFrom`
- `validTo`
- feature values
- missing-data flags

## Prediction Log Contract

Every prediction should include:

- `id`
- `hazardType`
- `modelVersion`
- `predictionTime`
- `targetTime`
- `geometry`
- `probability`
- `severity`
- `confidence`
- `explanation`
- `inputFeatureSetVersion`
- `createdBy`

Current MVP serving implementation:

- `services/ml-service` exposes `POST /api/v1/ml/flood/predictions` and `GET /api/v1/ml/prediction-logs`.
- The service loads `data/flood-risk/models/baseline-logistic.v1.json` and `data/flood-risk/models/sample-predictions.v1.json`.
- `NADAA_ML_MODEL_DIR` can point the service to a mounted model artifact directory.
- Prediction logs are in-memory during the MVP but use `storageTarget: "ml_predictions"` and carry `modelVersion` plus `inputFeatureSetVersion` for the future database writer.
- `services/risk-service` attaches `mlPrediction` when `NADAA_ML_API_URL` points to the ML API base URL.
- Every served prediction has `humanReviewRequired: true` and `autoPublishAllowed: false`.

## Human Review Rules

- Predictions can inform risk maps.
- Predictions can create alert drafts.
- Predictions cannot publish alerts.
- Authority users must see confidence, explanation, and model version.
- Overrides and alert decisions should be captured for later model evaluation.

## Incident Triage Suggestions (NADAA-151)

The incident service serves explainable, rules-based triage suggestions for dispatchers. The model identifier is `incident-triage-rules-0.1.0` over feature set `incident-features.v1`.

### Triage Labels

- Severity: `low`, `moderate`, `high`, `emergency`.
- Duplicate likelihood: `0.0` to `1.0`, taken from the strongest scored duplicate candidate, with up to three linked incident ids.
- Affected population: integer estimate derived from reported people affected, urgency, and duplicate report volume, capped at 1,000,000.
- Suggested agency routing: one of the supported agency types (`fire`, `police`, `ambulance`, `district_assembly`, `nadmo`) with a plain-language reason.
- Confidence: `low`, `medium`, `high` based on input completeness.

### Triage Features And Evaluation Data

- Inputs: reported urgency, people affected, injuries reported, hazard type, duplicate candidate scores (NADAA-033), and abuse review signals (NADAA-091).
- Duplicate signals only count open candidates: incidents already merged or marked as false reports are excluded, matching the duplicate-review endpoint, so triage never scores reports dispatchers have already dismissed.
- Every suggestion returns per-feature explanation factors with contribution weights and direction so dispatchers can see why a value was suggested.
- Evaluation data: every suggestion exposure is logged as an `incident.triage_suggested` audit event with a unique `suggestionId`, and triage review events pair the dispatcher decision (accept or override) with the exact logged suggestion the dispatcher saw, forming the labeled dataset for future model training and threshold tuning.

### Human Oversight Rules

- `humanReviewRequired` is always `true` and `autoPublishAllowed` is always `false`.
- Suggestions never verify, close, assign, or merge incidents and never create alerts.
- Dispatchers can edit severity, affected population, and agency routing before acting; overrides require a written reason and record only the fields the dispatcher actually changed.
- Every suggestion exposure is audit-logged (`incident.triage_suggested`), and every acceptance or override is logged as an `incident.triage_accepted` or `incident.triage_overridden` timeline event plus an audit event capturing the reviewed suggestion (matched by `suggestionId`), dispatcher values, model version, and feature set version.

### Bias And Error Review Process

- Review cadence: NADMO reviews triage audit logs at least once per sprint and after every major incident surge.
- Override rate: track the share of suggestions overridden overall and segmented by hazard type, region/district, reporter anonymity, and accessibility needs. A segment whose override rate diverges by more than 15 percentage points from the platform average triggers a rules review.
- Error taxonomy: classify overrides as severity under-call, severity over-call, wrong agency routing, wrong population estimate, or duplicate mis-scoring; record the classification in the review notes.
- Under-call escalation: any suggestion overridden from `low`/`moderate` to `emergency` is reviewed within the same sprint because under-calls carry life-safety risk.
- Data bias checks: compare triage outcomes for anonymous versus identified reporters and for low-connectivity districts to detect systematic under-weighting of underserved communities.
- Change control: rule or weight changes bump `TriageModelVersion`, are documented in this file, and require review sign-off before deployment; suggestions produced by prior versions remain attributable through the logged model version.

## Data Risks

- Official weather and hydrology access may lag delivery.
- Historical disaster records may be incomplete or inconsistent.
- Citizen reports may be biased toward connected communities.
- Satellite products may have latency or licensing constraints.
- Model performance may vary by region and season.

## Delivery Path

1. Rule-based flood risk score in the risk service.
2. Feature schema, sample data, generated outputs, and validation script.
3. Baseline logistic-regression model, sample predictions, and fixture evaluation report.
4. MVP ML HTTP model serving and risk-service decision-support integration.
5. MVP prediction logs aligned to the `ml_predictions` table shape.
6. Durable prediction persistence.
7. Authority review UI before alert drafting or approval.

### Computer Vision Image Verification (NADAA-152)

Purpose: decision-support image analysis for flood and fire evidence detection.

Current implementation:

- `services/ml-service` exposes `POST /api/v1/cv/analyze`, `GET /api/v1/cv/results/{imageId}`, and `GET /api/v1/cv/results`.
- The first-pass engine is a deterministic rule-based mock that parses filenames for hints.
- Supported labels: `flood_evidence`, `fire_evidence`, `smoke_evidence`, `no_evidence`, `unclear`, `sensitive`, `person_in_distress`.
- Confidence threshold for human review: 0.7. Sensitive labels always require review.
- Model version: `cv-mock-rule-engine-0.1.0`.
- Limitations are returned with every result to remind users this is not real inference.
- Results are cached in-memory by `imageId`.
- Safety policy: `autoPublishAllowed=false`, `humanReviewRequired` based on confidence/labels.

Future integration path:

1. Replace mock engine with ONNX/TensorFlow Lite model serving.
2. Add real bounding box coordinates from object detection.
3. Add image preprocessing pipeline (resize, normalize, augment).
4. Add model versioning and A/B testing.
5. Add batch analysis endpoint for multiple images.
6. Integrate with incident-service media ingestion pipeline (NADAA-140).

### Predictive Resource Positioning (NADAA-153)

Purpose: decision-support demand forecasting and staging suggestions so agencies can pre-position ambulances and fire units before high-risk periods.

Current implementation:

- `services/ml-service` exposes `GET /api/v1/forecasts` (optional `?region=`), `GET /api/v1/forecasts/{region}`, `GET /api/v1/staging-suggestions` (optional `?agencyType=`), and `POST /api/v1/forecasts/compare`.
- Model version: `resource-forecast-rules-0.1.0`. Fully deterministic — no randomness or wall-clock in the scoring; time windows derive from the injected request clock.
- Demand features per district, aggregated from the NADAA-070 flood-risk feature grid: historical flood reports (30d), 24h rainfall forecast, composite flood-risk score, and vulnerable-population percentage. `predictedIncidentCount = round((histReports * historicalWeight + composite * 4) * (0.7 + rainfall/150) * (windowHours/24))`.
- Confidence score reflects data completeness and signal decisiveness (historical presence, cell coverage, composite extremity) and maps to `low`/`medium`/`high` bands.
- Staging suggestions match a fixed set of candidate bases (fire, ambulance, NADMO) to the nearest demand district by haversine distance, sizing `recommendedUnits` by predicted demand and agency type, with per-agency operational constraints and a 5 km coverage radius.
- Scenario comparison returns a baseline ("Current conditions") and an adjusted scenario applying `historicalWeight`, `capacityFactor`, `riskLevel`, `hazardTypes`, and `timeWindowHours`, each with a summary of total predicted incidents and average confidence.

Human oversight and safety:

- All endpoints are read-only decision support; there are no deployment-action endpoints, so no automatic dispatch is possible.
- Forecasts and staging suggestions include confidence and operational constraints; agency leadership retains final deployment authority.

Bias and error review process:

- Track forecast accuracy by comparing `predictedIncidentCount` against realized incidents per district and window; review divergence by region and hazard each sprint.
- Watch for under-service of low-connectivity districts whose historical report counts may understate true demand; weight reviews toward vulnerable-population segments.
- Model or weight changes bump `resource-forecast-rules-0.1.0` and are documented here before deployment.

Future integration path:

1. Replace the rules baseline with a trained demand model (Poisson/gradient-boosted) over historical incident time series.
2. Incorporate live weather/hydrology feeds and real response-time telemetry.
3. Source candidate staging positions from an agency facility registry instead of fixtures.
4. Add capacity-aware optimization against real-time ambulance/hospital availability (NADAA-121).
