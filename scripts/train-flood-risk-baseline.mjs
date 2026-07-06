import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const featurePath = path.join(
  root,
  "data",
  "flood-risk",
  "generated",
  "features.v1.json",
);
const modelDir = path.join(root, "data", "flood-risk", "models");
const modelPath = path.join(modelDir, "baseline-logistic.v1.json");
const predictionsPath = path.join(modelDir, "sample-predictions.v1.json");
const evaluationPath = path.join(modelDir, "evaluation.v1.json");
const reportPath = path.join(modelDir, "evaluation-report.v1.md");

const modelVersion = "flood-logistic-baseline-0.1.0";
const generatedAt = "2026-07-06T10:00:00Z";
const targetTime = "2026-07-06T12:00:00Z";
const featureColumns = [
  "rainfall_intensity_score",
  "exposure_score",
  "drainage_pressure_score",
  "historical_signal_score",
  "rainfall_forecast_24h_mm",
  "water_level_trend_cm",
  "distance_to_drain_m",
  "distance_to_river_m",
  "impervious_surface_pct",
  "population_density_per_sq_km",
  "vulnerable_population_pct",
  "inside_known_flood_zone",
  "low_lying_area",
];
const highRiskLabels = new Set(["high", "severe"]);
const hyperparameters = {
  learningRate: 0.18,
  epochs: 5000,
  l2: 0.015,
  classificationThreshold: 0.5,
  severityThresholds: {
    severe: 0.78,
    high: 0.58,
    moderate: 0.35,
  },
  severeRule:
    "probability >= 0.78, inside known flood zone, and rainfall_forecast_24h_mm >= 55",
};
const featureLabels = {
  rainfall_intensity_score: "rainfall intensity",
  exposure_score: "population and land-use exposure",
  drainage_pressure_score: "drainage and terrain pressure",
  historical_signal_score: "recent historical flood signal",
  rainfall_forecast_24h_mm: "24-hour rainfall forecast",
  water_level_trend_cm: "water-level trend",
  distance_to_drain_m: "distance to nearest drain",
  distance_to_river_m: "distance to nearest river",
  impervious_surface_pct: "impervious surface share",
  population_density_per_sq_km: "population density",
  vulnerable_population_pct: "vulnerable population share",
  inside_known_flood_zone: "inside known flood zone",
  low_lying_area: "low-lying area",
};

const featurePayload = JSON.parse(await readFile(featurePath, "utf8"));
const rows = featurePayload.rows;

if (!Array.isArray(rows) || rows.length === 0) {
  throw new Error("Feature payload must include at least one row.");
}

const labels = rows.map((row) =>
  highRiskLabels.has(row.label_severity) ? 1 : 0,
);
const preprocessing = buildStandardization(rows, featureColumns);
const trainingRows = rows.map((row) =>
  featureColumns.map((column) =>
    standardize(row[column], preprocessing[column]),
  ),
);
const weights = trainLogisticRegression(trainingRows, labels, hyperparameters);
const model = buildModel(weights, preprocessing, featurePayload);
const predictions = buildPredictions(model, rows, preprocessing);
const evaluation = buildEvaluation(
  predictions.predictions,
  labels,
  rows,
  model,
);
const report = buildEvaluationReport(evaluation, model, predictions);

model.trainingMetrics = evaluation.metrics;
model.evaluationArtifact = path.relative(root, evaluationPath);
model.predictionArtifact = path.relative(root, predictionsPath);

await mkdir(modelDir, { recursive: true });

const modelText = `${JSON.stringify(model, null, 2)}\n`;
const predictionsText = `${JSON.stringify(predictions, null, 2)}\n`;
const evaluationText = `${JSON.stringify(evaluation, null, 2)}\n`;
const reportText = `${report}\n`;

await writeFile(modelPath, modelText);
await writeFile(predictionsPath, predictionsText);
await writeFile(evaluationPath, evaluationText);
await writeFile(reportPath, reportText);

console.log(
  `Trained ${modelVersion} with ${rows.length} rows and wrote ${path.relative(
    root,
    modelDir,
  )}.`,
);

function buildStandardization(records, columns) {
  const result = {};
  for (const column of columns) {
    const values = records.map((row) => numericValue(row[column]));
    const mean = average(values);
    const variance = average(values.map((value) => (value - mean) ** 2));
    result[column] = {
      mean: round(mean, 6),
      std: round(Math.sqrt(variance) || 1, 6),
    };
  }
  return result;
}

function trainLogisticRegression(xRows, yValues, params) {
  const weights = new Array(featureColumns.length + 1).fill(0);

  for (let epoch = 0; epoch < params.epochs; epoch++) {
    const gradients = new Array(weights.length).fill(0);

    for (const [index, xRow] of xRows.entries()) {
      const x = [1, ...xRow];
      const prediction = sigmoid(dot(weights, x));
      const error = prediction - yValues[index];

      for (let weightIndex = 0; weightIndex < gradients.length; weightIndex++) {
        gradients[weightIndex] += error * x[weightIndex];
      }
    }

    for (let weightIndex = 0; weightIndex < weights.length; weightIndex++) {
      const l2Penalty =
        weightIndex === 0 ? 0 : params.l2 * weights[weightIndex];
      weights[weightIndex] -=
        params.learningRate *
        (gradients[weightIndex] / xRows.length + l2Penalty);
    }
  }

  return weights.map((weight) => round(weight, 8));
}

function buildModel(weights, standardization, featurePayload) {
  const coefficients = {
    intercept: weights[0],
  };
  for (const [index, column] of featureColumns.entries()) {
    coefficients[column] = weights[index + 1];
  }

  return {
    modelVersion,
    modelFamily: "logistic_regression",
    hazardType: "flood",
    generatedAt,
    trainingFeatureSetVersion: featurePayload.featureSetVersion,
    trainingSchemaVersion: featurePayload.schemaVersion,
    trainingSource: featurePayload.source,
    trainingSourceUpdatedAt: featurePayload.sourceUpdatedAt,
    target: {
      name: "high_or_severe_flood_risk",
      positiveLabels: ["high", "severe"],
      negativeLabels: ["low", "moderate"],
    },
    featureColumns,
    preprocessing: {
      booleanEncoding: "false=0,true=1",
      numericStandardization: standardization,
    },
    hyperparameters,
    coefficients,
    trainingRows: rows.length,
    outputContract: {
      probability: "0..1 high-or-severe flood probability",
      severity: "low, moderate, high, or severe",
      confidence:
        "low, medium, or high review confidence; fixture-trained models are capped at medium",
      expectedOnset: "fixture onset bucket only; not a hydrodynamic forecast",
      explanationFactors:
        "top signed feature contributions from the logistic model",
    },
    limitations: [
      "The fixture training set has only five seed-aligned rows and is suitable for product/API integration, not production accuracy claims.",
      "Labels are derived from the NADAA-070 fixture severity labels; they are not independently observed event outcomes.",
      "Model output must remain decision support and cannot publish public alerts without human review.",
      "Replace fixture rows with official rainfall, hydrology, terrain, population, and historical disaster records before production model promotion.",
    ],
  };
}

function buildPredictions(model, featureRows, standardization) {
  const predictions = featureRows.map((row) => {
    const probability = predictProbability(model, row, standardization);
    const severity = severityFromProbability(probability, row);
    return {
      id: `pred_${row.cell_id}`,
      modelVersion: model.modelVersion,
      hazardType: "flood",
      predictionTime: generatedAt,
      targetTime,
      cellId: row.cell_id,
      region: row.region,
      district: row.district,
      community: row.community,
      geometry: row.geometry,
      probability,
      severity,
      expectedOnset: expectedOnset(row, severity),
      confidence: confidenceFor(probability, row),
      explanationFactors: explanationFactors(model, row, standardization),
      inputFeatureSetVersion: model.trainingFeatureSetVersion,
      sourceFeatureRow: row.cell_id,
    };
  });

  return {
    modelVersion: model.modelVersion,
    featureSetVersion: model.trainingFeatureSetVersion,
    generatedAt,
    targetTime,
    predictionCount: predictions.length,
    predictions,
  };
}

function buildEvaluation(predictions, labelValues, featureRows, model) {
  const classified = predictions.map((prediction, index) => ({
    prediction,
    actual: labelValues[index],
    predicted:
      prediction.probability >= model.hyperparameters.classificationThreshold
        ? 1
        : 0,
    labelSeverity: featureRows[index].label_severity,
  }));
  const tp = classified.filter(
    (item) => item.actual === 1 && item.predicted === 1,
  ).length;
  const tn = classified.filter(
    (item) => item.actual === 0 && item.predicted === 0,
  ).length;
  const fp = classified.filter(
    (item) => item.actual === 0 && item.predicted === 1,
  );
  const fn = classified.filter(
    (item) => item.actual === 1 && item.predicted === 0,
  );
  const precision = safeDivide(tp, tp + fp.length);
  const recall = safeDivide(tp, tp + fn.length);
  const f1 = safeDivide(2 * precision * recall, precision + recall);
  const accuracy = safeDivide(tp + tn, classified.length);
  const brierScore = average(
    classified.map((item) => (item.prediction.probability - item.actual) ** 2),
  );

  return {
    modelVersion: model.modelVersion,
    featureSetVersion: model.trainingFeatureSetVersion,
    generatedAt,
    evaluationScope: "training_fixture_resubstitution",
    rowCount: classified.length,
    positiveLabelCount: labelValues.filter(Boolean).length,
    negativeLabelCount: labelValues.filter((label) => !label).length,
    metrics: {
      accuracy: round(accuracy, 4),
      precisionHighRisk: round(precision, 4),
      recallHighRisk: round(recall, 4),
      f1HighRisk: round(f1, 4),
      brierScore: round(brierScore, 4),
    },
    confusionMatrix: {
      truePositive: tp,
      falsePositive: fp.length,
      trueNegative: tn,
      falseNegative: fn.length,
    },
    calibrationBuckets: buildCalibrationBuckets(classified),
    falsePositiveReview: {
      count: fp.length,
      cases: fp.map((item) => reviewCase(item)),
      process:
        "Review each predicted high/severe cell against rainfall, hydrology, report recency, and local officer context before drafting an alert.",
    },
    falseNegativeReview: {
      count: fn.length,
      cases: fn.map((item) => reviewCase(item)),
      process:
        "Review each actual high/severe fixture predicted below threshold and add missing rainfall, blocked-drain, or incident-report signals before model promotion.",
    },
    limitations: model.limitations,
  };
}

function buildEvaluationReport(evaluation, model, predictions) {
  const rows = predictions.predictions
    .map(
      (prediction) =>
        `| ${prediction.cellId} | ${prediction.community} | ${prediction.probability.toFixed(
          4,
        )} | ${prediction.severity} | ${prediction.confidence} | ${prediction.expectedOnset} |`,
    )
    .join("\n");

  return `# Flood Logistic Baseline Evaluation

Model version: \`${model.modelVersion}\`

Feature set: \`${model.trainingFeatureSetVersion}\`

Evaluation scope: \`${evaluation.evaluationScope}\`

## Metrics

| Metric | Value |
| ------ | ----- |
| Accuracy | ${evaluation.metrics.accuracy} |
| Precision high-risk | ${evaluation.metrics.precisionHighRisk} |
| Recall high-risk | ${evaluation.metrics.recallHighRisk} |
| F1 high-risk | ${evaluation.metrics.f1HighRisk} |
| Brier score | ${evaluation.metrics.brierScore} |

## Confusion Matrix

| True positive | False positive | True negative | False negative |
| ------------- | -------------- | ------------- | -------------- |
| ${evaluation.confusionMatrix.truePositive} | ${evaluation.confusionMatrix.falsePositive} | ${evaluation.confusionMatrix.trueNegative} | ${evaluation.confusionMatrix.falseNegative} |

## Sample Predictions

| Cell | Community | Probability | Severity | Confidence | Expected onset |
| ---- | --------- | ----------- | -------- | ---------- | -------------- |
${rows}

## False Positive Review

${evaluation.falsePositiveReview.process}

Current fixture false positives: ${evaluation.falsePositiveReview.count}.

## False Negative Review

${evaluation.falseNegativeReview.process}

Current fixture false negatives: ${evaluation.falseNegativeReview.count}.

## Limitations

${model.limitations.map((limitation) => `- ${limitation}`).join("\n")}
`;
}

function predictProbability(model, row, standardization) {
  const weights = featureColumns.map((column) => model.coefficients[column]);
  const standardized = featureColumns.map((column) =>
    standardize(row[column], standardization[column]),
  );
  return round(
    sigmoid(model.coefficients.intercept + dot(weights, standardized)),
    4,
  );
}

function explanationFactors(model, row, standardization) {
  return featureColumns
    .map((column) => {
      const standardized = standardize(row[column], standardization[column]);
      const contribution = round(model.coefficients[column] * standardized, 4);
      return {
        feature: column,
        label: featureLabels[column],
        value: row[column],
        contribution,
        direction: contribution >= 0 ? "increases_risk" : "reduces_risk",
      };
    })
    .sort((a, b) => Math.abs(b.contribution) - Math.abs(a.contribution))
    .slice(0, 4);
}

function expectedOnset(row, severity) {
  if (
    severity === "severe" &&
    row.rainfall_forecast_24h_mm >= 55 &&
    row.water_level_trend_cm >= 20
  ) {
    return "within_24h";
  }
  if (severity === "high" && row.rainfall_forecast_24h_mm >= 45) {
    return "24_to_48h";
  }
  if (severity === "moderate") {
    return "48_to_72h";
  }
  return "not_expected_in_72h";
}

function confidenceFor(probability, row) {
  const hasMissing = Object.entries(row).some(
    ([key, value]) => key.startsWith("missing_") && value === true,
  );
  if (hasMissing) {
    return "low";
  }
  const margin = Math.abs(
    probability - hyperparameters.classificationThreshold,
  );
  if (margin >= 0.15) {
    return "medium";
  }
  return "low";
}

function severityFromProbability(probability, row) {
  if (
    probability >= hyperparameters.severityThresholds.severe &&
    row.inside_known_flood_zone &&
    row.rainfall_forecast_24h_mm >= 55
  ) {
    return "severe";
  }
  if (probability >= hyperparameters.severityThresholds.high) {
    return "high";
  }
  if (probability >= hyperparameters.severityThresholds.moderate) {
    return "moderate";
  }
  return "low";
}

function buildCalibrationBuckets(classified) {
  const buckets = [
    { name: "0.00-0.25", min: 0, max: 0.25 },
    { name: "0.25-0.50", min: 0.25, max: 0.5 },
    { name: "0.50-0.75", min: 0.5, max: 0.75 },
    { name: "0.75-1.00", min: 0.75, max: 1.01 },
  ];

  return buckets.map((bucket) => {
    const members = classified.filter(
      (item) =>
        item.prediction.probability >= bucket.min &&
        item.prediction.probability < bucket.max,
    );
    return {
      bucket: bucket.name,
      count: members.length,
      averagePrediction: round(
        members.length
          ? average(members.map((item) => item.prediction.probability))
          : 0,
        4,
      ),
      observedHighRiskRate: round(
        members.length ? average(members.map((item) => item.actual)) : 0,
        4,
      ),
    };
  });
}

function reviewCase(item) {
  return {
    cellId: item.prediction.cellId,
    community: item.prediction.community,
    probability: item.prediction.probability,
    predictedSeverity: item.prediction.severity,
    labelSeverity: item.labelSeverity,
    explanationFactors: item.prediction.explanationFactors,
  };
}

function standardize(value, rule) {
  return round((numericValue(value) - rule.mean) / rule.std, 6);
}

function numericValue(value) {
  if (typeof value === "boolean") {
    return value ? 1 : 0;
  }
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`Expected finite numeric feature value, got ${value}`);
  }
  return value;
}

function dot(left, right) {
  return left.reduce((sum, value, index) => sum + value * right[index], 0);
}

function sigmoid(value) {
  return 1 / (1 + Math.exp(-value));
}

function average(values) {
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function safeDivide(numerator, denominator) {
  return denominator === 0 ? 0 : numerator / denominator;
}

function round(value, digits) {
  return Number(value.toFixed(digits));
}
