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
2. Feature schema, sample data, generated outputs, and validation script.
3. Baseline model using logistic regression, random forest, or XGBoost.
4. FastAPI model serving.
5. Prediction logs in `ml_predictions`.
6. Risk API integration.
7. Authority review UI before alert drafting or approval.
