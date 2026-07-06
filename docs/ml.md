# ML

## MVP Focus

The first ML priority is flood risk prediction.

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

## Delivery Path

1. Rule-based flood risk score in the risk service.
2. Baseline model using logistic regression, random forest, or XGBoost.
3. FastAPI model serving.
4. Risk API integration.
5. Authority review UI before any alert is drafted or approved.

