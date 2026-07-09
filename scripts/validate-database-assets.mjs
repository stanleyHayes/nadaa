import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();
const migrationDir = join(root, "database/migrations");
const migration = readdirSync(migrationDir)
  .filter((file) => file.endsWith(".sql"))
  .sort()
  .map((file) => readFileSync(join(migrationDir, file), "utf8"))
  .join("\n");
const seed = readFileSync(
  join(root, "database/seeds/001_ghana_mvp_seed.sql"),
  "utf8",
);

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
  "weather_import_jobs",
  "audit_logs",
  "notification_delivery_logs",
  "road_closures",
  "relief_points",
  "relief_point_stock_history",
  "aid_requests",
  "aid_pledges",
  "missing_person_reports",
  "missing_person_audit_entries",
];

const requiredGeometryIndexes = [
  "idx_agencies_service_area_geometry",
  "idx_users_home_location",
  "idx_incidents_location_geometry",
  "idx_alerts_target_geometry",
  "idx_risk_zones_geometry",
  "idx_shelters_location_geometry",
  "idx_ml_predictions_geometry",
  "idx_weather_observations_location",
  "idx_road_closures_geometry",
  "idx_relief_points_geometry",
  "idx_aid_requests_geometry",
  "idx_missing_person_reports_geometry",
];

const requiredEnums = [
  "user_role",
  "agency_type",
  "hazard_type",
  "risk_level",
  "incident_status",
  "alert_severity",
  "alert_status",
  "guide_stage",
  "notification_channel",
  "notification_delivery_status",
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

for (const table of [
  "agencies",
  "users",
  "shelters",
  "risk_zones",
  "emergency_guides",
  "incidents",
  "alerts",
  "ml_predictions",
]) {
  if (!seed.includes(`INSERT INTO ${table}`)) {
    throw new Error(`Missing seed insert for table: ${table}`);
  }
}

if (!migration.includes("CREATE EXTENSION IF NOT EXISTS postgis")) {
  throw new Error("Migration must enable PostGIS");
}

console.log(
  `Validated database assets: ${requiredTables.length} tables, ${requiredEnums.length} enums, ${requiredGeometryIndexes.length} geometry indexes.`,
);
