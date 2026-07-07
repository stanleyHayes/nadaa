# Flood Logistic Baseline Evaluation

Model version: `flood-logistic-baseline-0.1.0`

Feature set: `flood-risk-features.v1`

Evaluation scope: `training_fixture_resubstitution`

## Metrics

| Metric              | Value  |
| ------------------- | ------ |
| Accuracy            | 1      |
| Precision high-risk | 1      |
| Recall high-risk    | 1      |
| F1 high-risk        | 1      |
| Brier score         | 0.0002 |

## Confusion Matrix

| True positive | False positive | True negative | False negative |
| ------------- | -------------- | ------------- | -------------- |
| 3             | 0              | 2             | 0              |

## Sample Predictions

| Cell                   | Community          | Probability | Severity | Confidence | Expected onset      |
| ---------------------- | ------------------ | ----------- | -------- | ---------- | ------------------- |
| grid-accra-central-001 | Accra Central      | 0.9993      | severe   | medium     | within_24h          |
| grid-accra-north-002   | Accra North        | 0.9828      | high     | medium     | 24_to_48h           |
| grid-osu-003           | Osu                | 0.9906      | high     | medium     | 24_to_48h           |
| grid-tema-004          | Tema Community One | 0.0264      | low      | medium     | not_expected_in_72h |
| grid-kumasi-005        | Adum               | 0.0008      | low      | medium     | not_expected_in_72h |

## False Positive Review

Review each predicted high/severe cell against rainfall, hydrology, report recency, and local officer context before drafting an alert.

Current fixture false positives: 0.

## False Negative Review

Review each actual high/severe fixture predicted below threshold and add missing rainfall, blocked-drain, or incident-report signals before model promotion.

Current fixture false negatives: 0.

## Limitations

- The fixture training set has only five seed-aligned rows and is suitable for product/API integration, not production accuracy claims.
- Labels are derived from the NADAA-070 fixture severity labels; they are not independently observed event outcomes.
- Model output must remain decision support and cannot publish public alerts without human review.
- Replace fixture rows with official rainfall, hydrology, terrain, population, and historical disaster records before production model promotion.
