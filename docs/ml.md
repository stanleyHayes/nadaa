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

### Stage 1: Baseline ML

Candidate models:

- Logistic regression.
- Random forest.
- XGBoost.

Minimum evaluation:

- Precision/recall for high-risk flood events.
- Calibration by probability band.
- False positive review.
- False negative review.
- District/community breakdown when data permits.

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

## Human Review Rules

- Predictions can inform risk maps.
- Predictions can create alert drafts.
- Predictions cannot publish alerts.
- Authority users must see confidence, explanation, and model version.
- Overrides and alert decisions should be captured for later model evaluation.

## Data Risks

- Official weather and hydrology access may lag delivery.
- Historical disaster records may be incomplete or inconsistent.
- Citizen reports may be biased toward connected communities.
- Satellite products may have latency or licensing constraints.
- Model performance may vary by region and season.

## Delivery Path

1. Rule-based flood risk score in the risk service.
2. Feature schema and sample data.
3. Baseline model using logistic regression, random forest, or XGBoost.
4. FastAPI model serving.
5. Prediction logs in `ml_predictions`.
6. Risk API integration.
7. Authority review UI before alert drafting or approval.
