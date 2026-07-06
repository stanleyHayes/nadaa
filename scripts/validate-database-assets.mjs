import { readFileSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();
const migration = readFileSync(
  join(root, "database/migrations/001_core_geospatial_schema.sql"),
  "utf8"
);
const seed = readFileSync(join(root, "database/seeds/001_ghana_mvp_seed.sql"), "utf8");

const requiredTables = [
  "agencies",
  "users",
  "incidents",
  "incident_media",
  "incident_timeline_events",
  "alerts",
  "risk_zones",
  "shelters",
  "emergency_guides",
  "ml_predictions",
  "weather_observations",
  "audit_logs"
];

const requiredGeometryIndexes = [
  "idx_agencies_service_area_geometry",
  "idx_users_home_location",
  "idx_incidents_location_geometry",
  "idx_alerts_target_geometry",
  "idx_risk_zones_geometry",
  "idx_shelters_location_geometry",
  "idx_ml_predictions_geometry",
  "idx_weather_observations_location"
];

const requiredEnums = [
  "user_role",
  "agency_type",
  "hazard_type",
  "risk_level",
  "incident_status",
  "alert_severity",
  "alert_status",
  "guide_stage"
];

for (const enumName of requiredEnums) {
  if (!migration.includes(`CREATE TYPE ${enumName}`)) {
    throw new Error(`Missing enum: ${enumName}`);
  }
}

for (const table of requiredTables) {
  if (!migration.includes(`CREATE TABLE IF NOT EXISTS ${table}`)) {
    throw new Error(`Missing table: ${table}`);
  }
}

for (const index of requiredGeometryIndexes) {
  if (!migration.includes(index)) {
    throw new Error(`Missing geospatial index: ${index}`);
  }
}

for (const table of ["agencies", "users", "shelters", "risk_zones", "emergency_guides", "incidents", "alerts", "ml_predictions"]) {
  if (!seed.includes(`INSERT INTO ${table}`)) {
    throw new Error(`Missing seed insert for table: ${table}`);
  }
}

if (!migration.includes("CREATE EXTENSION IF NOT EXISTS postgis")) {
  throw new Error("Migration must enable PostGIS");
}

console.log(
  `Validated database assets: ${requiredTables.length} tables, ${requiredEnums.length} enums, ${requiredGeometryIndexes.length} geometry indexes.`
);

