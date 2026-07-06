import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const dataDir = path.join(root, "data", "flood-risk");
const generatedDir = path.join(dataDir, "generated");
const fixturesPath = path.join(dataDir, "source-fixtures.v1.json");
const schemaPath = path.join(dataDir, "feature-schema.v1.json");
const featuresPath = path.join(generatedDir, "features.v1.json");
const csvPath = path.join(generatedDir, "features.v1.csv");
const manifestPath = path.join(generatedDir, "manifest.v1.json");

const fixtures = JSON.parse(await readFile(fixturesPath, "utf8"));
const schema = JSON.parse(await readFile(schemaPath, "utf8"));
const featuresText = await readFile(featuresPath, "utf8");
const csvText = await readFile(csvPath, "utf8");
const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
const features = JSON.parse(featuresText);

const columns = schema.columns.map((column) => column.name);
const columnRules = new Map(
  schema.columns.map((column) => [column.name, column]),
);

assert(
  fixtures.datasetVersion === schema.featureSetVersion,
  "Fixture dataset version must match schema feature set version.",
);
assert(
  features.featureSetVersion === schema.featureSetVersion,
  "Feature output version must match schema feature set version.",
);
assert(
  features.schemaVersion === schema.schemaVersion,
  "Feature output schema version must match schema.",
);
assert(
  manifest.featureSetVersion === features.featureSetVersion,
  "Manifest feature set version must match output.",
);
assert(
  manifest.rowCount === features.rows.length,
  "Manifest row count must match output rows.",
);
assert(
  features.rows.length >= 5,
  "Feature output must include at least 5 rows.",
);

const csvHeader = csvText.split("\n")[0].split(",");
assert(
  JSON.stringify(csvHeader) === JSON.stringify(columns),
  "CSV header must match schema columns.",
);

for (const [index, row] of features.rows.entries()) {
  const keys = Object.keys(row);
  assert(
    JSON.stringify(keys) === JSON.stringify(columns),
    `Row ${index} must include schema columns in order.`,
  );
  assert(
    row.feature_set_version === features.featureSetVersion,
    `Row ${index} feature_set_version must match output version.`,
  );

  for (const column of columns) {
    validateValue(index, column, row[column], columnRules.get(column));
  }
}

const outputChecksums = new Map(
  manifest.outputs.map((output) => [output.path, output.sha256]),
);
assert(
  outputChecksums.get("data/flood-risk/generated/features.v1.json") ===
    sha256(featuresText),
  "Manifest checksum for features.v1.json is stale.",
);
assert(
  outputChecksums.get("data/flood-risk/generated/features.v1.csv") ===
    sha256(csvText),
  "Manifest checksum for features.v1.csv is stale.",
);

console.log(
  `Validated ${features.rows.length} flood-risk feature rows across ${columns.length} columns.`,
);

function validateValue(rowIndex, column, value, rule) {
  assert(rule, `Unknown schema column: ${column}`);

  if (rule.required) {
    assert(
      value !== undefined && value !== null && value !== "",
      `Row ${rowIndex} ${column} is required.`,
    );
  }

  switch (rule.type) {
    case "string":
      assert(
        typeof value === "string",
        `Row ${rowIndex} ${column} must be a string.`,
      );
      break;
    case "datetime":
      assert(
        typeof value === "string" && !Number.isNaN(Date.parse(value)),
        `Row ${rowIndex} ${column} must be an ISO datetime string.`,
      );
      break;
    case "number":
      assert(
        typeof value === "number" && Number.isFinite(value),
        `Row ${rowIndex} ${column} must be a finite number.`,
      );
      validateRange(rowIndex, column, value, rule);
      break;
    case "integer":
      assert(
        Number.isInteger(value),
        `Row ${rowIndex} ${column} must be an integer.`,
      );
      validateRange(rowIndex, column, value, rule);
      break;
    case "boolean":
      assert(
        typeof value === "boolean",
        `Row ${rowIndex} ${column} must be a boolean.`,
      );
      break;
    case "geojson":
      assert(
        value && value.type === "Polygon" && Array.isArray(value.coordinates),
        `Row ${rowIndex} ${column} must be GeoJSON polygon geometry.`,
      );
      break;
    default:
      throw new Error(`Unsupported schema type: ${rule.type}`);
  }

  if (rule.enum) {
    assert(
      rule.enum.includes(value),
      `Row ${rowIndex} ${column} must be one of ${rule.enum.join(", ")}.`,
    );
  }
}

function validateRange(rowIndex, column, value, rule) {
  if ("min" in rule) {
    assert(value >= rule.min, `Row ${rowIndex} ${column} is below min.`);
  }
  if ("max" in rule) {
    assert(value <= rule.max, `Row ${rowIndex} ${column} is above max.`);
  }
}

function sha256(content) {
  return createHash("sha256").update(content).digest("hex");
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
