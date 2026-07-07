import { readFileSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();
const schema = JSON.parse(
  readFileSync(
    join(root, "docs/project-dashboard/contract.schema.json"),
    "utf8",
  ),
);
const records = JSON.parse(
  readFileSync(
    join(root, "docs/project-dashboard/sample-records.json"),
    "utf8",
  ),
);

const required = schema.required;
const statusValues = new Set(schema.properties.status.enum);
const phaseValues = new Set(schema.properties.phase.enum);

if (!Array.isArray(records)) {
  throw new Error("sample-records.json must be an array");
}

for (const [index, record] of records.entries()) {
  for (const field of required) {
    if (!(field in record)) {
      throw new Error(`Record ${index} is missing required field: ${field}`);
    }
  }

  if (!/^NADAA-[0-9]{3}$/.test(record.jiraKey)) {
    throw new Error(`Record ${index} has invalid jiraKey: ${record.jiraKey}`);
  }

  if (!statusValues.has(record.status)) {
    throw new Error(`Record ${index} has invalid status: ${record.status}`);
  }

  if (!phaseValues.has(record.phase)) {
    throw new Error(`Record ${index} has invalid phase: ${record.phase}`);
  }

  if (!Number.isInteger(record.progressPercentage)) {
    throw new Error(`Record ${index} progressPercentage must be an integer`);
  }

  if (record.progressPercentage < 0 || record.progressPercentage > 100) {
    throw new Error(
      `Record ${index} progressPercentage must be between 0 and 100`,
    );
  }

  for (const listField of ["dependencies", "blockers", "verification"]) {
    if (!Array.isArray(record[listField])) {
      throw new Error(`Record ${index} ${listField} must be an array`);
    }
  }
}

console.log(`Validated ${records.length} dashboard records.`);
