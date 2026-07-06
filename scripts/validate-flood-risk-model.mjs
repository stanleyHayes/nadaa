import { readFile } from "node:fs/promises";
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

const features = JSON.parse(await readFile(featurePath, "utf8"));
const model = JSON.parse(await readFile(modelPath, "utf8"));
const predictions = JSON.parse(await readFile(predictionsPath, "utf8"));
const evaluation = JSON.parse(await readFile(evaluationPath, "utf8"));
const report = await readFile(reportPath, "utf8");

const severityValues = new Set(["low", "moderate", "high", "severe"]);
const confidenceValues = new Set(["low", "medium", "high"]);

assert(
  model.modelFamily === "logistic_regression",
  "Model must be logistic regression.",
);
assert(model.hazardType === "flood", "Model hazard type must be flood.");
assert(
  model.trainingFeatureSetVersion === features.featureSetVersion,
  "Model feature set version must match generated features.",
);
assert(
  predictions.featureSetVersion === features.featureSetVersion,
  "Predictions feature set version must match generated features.",
);
assert(
  evaluation.featureSetVersion === features.featureSetVersion,
  "Evaluation feature set version must match generated features.",
);
assert(
  predictions.modelVersion === model.modelVersion &&
    evaluation.modelVersion === model.modelVersion,
  "Prediction and evaluation artifacts must use the model version.",
);
assert(
  predictions.predictionCount === features.rows.length,
  "Prediction count must match feature rows.",
);
assert(
  predictions.predictions.length === features.rows.length,
  "Prediction rows must match feature rows.",
);
assert(
  model.trainingRows === features.rows.length,
  "Model trainingRows must match features.",
);
assert(
  Array.isArray(model.featureColumns) && model.featureColumns.length > 0,
  "Model must include feature columns.",
);
assert(
  typeof model.coefficients.intercept === "number" &&
    Number.isFinite(model.coefficients.intercept),
  "Model must include a finite intercept.",
);

for (const column of model.featureColumns) {
  assert(column in model.coefficients, `Missing coefficient for ${column}.`);
  assert(
    Number.isFinite(model.coefficients[column]),
    `Coefficient for ${column} must be finite.`,
  );
  assert(
    model.preprocessing.numericStandardization[column],
    `Missing standardization rule for ${column}.`,
  );
}

const predictionCells = new Set();
for (const prediction of predictions.predictions) {
  assert(
    !predictionCells.has(prediction.cellId),
    `Duplicate prediction for ${prediction.cellId}.`,
  );
  predictionCells.add(prediction.cellId);
  assert(
    typeof prediction.probability === "number" &&
      prediction.probability >= 0 &&
      prediction.probability <= 1,
    `Prediction ${prediction.cellId} probability must be between 0 and 1.`,
  );
  assert(
    severityValues.has(prediction.severity),
    `Invalid severity for ${prediction.cellId}.`,
  );
  assert(
    confidenceValues.has(prediction.confidence),
    `Invalid confidence for ${prediction.cellId}.`,
  );
  assert(
    Array.isArray(prediction.explanationFactors) &&
      prediction.explanationFactors.length > 0,
    `Prediction ${prediction.cellId} must include explanation factors.`,
  );
}

for (const row of features.rows) {
  assert(
    predictionCells.has(row.cell_id),
    `Missing prediction for feature row ${row.cell_id}.`,
  );
}

const matrix = evaluation.confusionMatrix;
const matrixTotal =
  matrix.truePositive +
  matrix.falsePositive +
  matrix.trueNegative +
  matrix.falseNegative;
assert(
  matrixTotal === features.rows.length,
  "Confusion matrix must cover all feature rows.",
);

for (const metric of [
  "accuracy",
  "precisionHighRisk",
  "recallHighRisk",
  "f1HighRisk",
  "brierScore",
]) {
  assert(
    Number.isFinite(evaluation.metrics[metric]) &&
      evaluation.metrics[metric] >= 0 &&
      evaluation.metrics[metric] <= 1,
    `Metric ${metric} must be between 0 and 1.`,
  );
}

assert(
  report.includes("## False Positive Review") &&
    report.includes("## False Negative Review"),
  "Evaluation report must include false positive and false negative review sections.",
);
assert(
  report.includes(model.modelVersion),
  "Evaluation report must include the model version.",
);

console.log(
  `Validated ${model.modelVersion} with ${predictions.predictionCount} sample predictions.`,
);

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
