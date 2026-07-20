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

// Enum values the services write that 001 did not include (added by 007).
const requiredEnumValues = [
  ["notification_channel", "voice"],
  ["notification_channel", "cell_broadcast"],
  ["notification_delivery_status", "broadcast"],
  ["notification_delivery_status", "simulated"],
  ["notification_delivery_status", "sent"],
];

// Columns added or renamed by 007 to match what the services persist.
const requiredServiceColumns = [
  "submitted_at",
  "approved_at",
  "rejected_by",
  "rejected_at",
  "status_reason",
  "emergency_override",
  "source_prediction_id",
  "urgency",
  "abuse_score",
];

// Columns added by 008 so incident-service can persist abuse-review and
// status-lifecycle fields (IncidentRecord in services/incident-service).
const requiredIncidentLifecycleColumns = [
  "abuse_review_required",
  "abuse_review_reason",
  "abuse_review_decision",
  "abuse_reviewed_by",
  "abuse_reviewed_at",
  "verified_at",
  "status_updated_by",
  "resolution_notes",
  "closed_at",
  "merged_into_id",
  "merged_by",
  "merged_at",
  "merge_reason",
];

// TEXT ID conversions from 008 for service-owned string IDs (usr_*, inc_*,
// road_closure_*, relief_*, aid_*) that can never cast to UUID. The child
// columns are included because PostgreSQL refuses ALTER COLUMN ... TYPE TEXT
// on a PK while a typed FK references it, so both sides must convert.
const requiredTextIdStatements = [
  "ALTER TABLE IF EXISTS incidents ALTER COLUMN id TYPE TEXT",
  "ALTER TABLE IF EXISTS users ALTER COLUMN id TYPE TEXT",
  "ALTER TABLE IF EXISTS road_closures ALTER COLUMN id TYPE TEXT",
  "ALTER TABLE IF EXISTS relief_points ALTER COLUMN id TYPE TEXT",
  "ALTER TABLE IF EXISTS aid_requests ALTER COLUMN id TYPE TEXT",
  "ALTER TABLE IF EXISTS incident_media ALTER COLUMN incident_id TYPE TEXT",
  "ALTER TABLE IF EXISTS incident_timeline_events ALTER COLUMN incident_id TYPE TEXT",
  "ALTER TABLE IF EXISTS relief_point_stock_history ALTER COLUMN relief_point_id TYPE TEXT",
  "ALTER TABLE IF EXISTS aid_pledges ALTER COLUMN aid_request_id TYPE TEXT",
];

// Seed agency IDs must match the incident-service triage fixtures
// (triageSuggestedAgency in services/incident-service/internal/store).
const requiredSeedAgencyIds = [
  "00000000-0000-0000-0000-000000000201",
  "00000000-0000-0000-0000-000000000202",
  "00000000-0000-0000-0000-000000000203",
  "00000000-0000-0000-0000-000000000204",
];

for (const enumName of requiredEnums) {
  if (!migration.includes(`CREATE TYPE ${enumName}`)) {
    throw new Error(`Missing enum: ${enumName}`);
  }
}

for (const [enumName, value] of requiredEnumValues) {
  if (
    !migration.includes(
      `ALTER TYPE ${enumName} ADD VALUE IF NOT EXISTS '${value}'`,
    )
  ) {
    throw new Error(`Missing enum value: ${enumName} '${value}'`);
  }
}

for (const column of requiredServiceColumns) {
  if (!migration.includes(column)) {
    throw new Error(`Missing service-alignment column: ${column}`);
  }
}

for (const column of requiredIncidentLifecycleColumns) {
  if (!migration.includes(column)) {
    throw new Error(`Missing incident lifecycle column: ${column}`);
  }
}

for (const statement of requiredTextIdStatements) {
  if (!migration.includes(statement)) {
    throw new Error(`Missing TEXT ID alignment: ${statement}`);
  }
}

for (const id of requiredSeedAgencyIds) {
  if (!seed.includes(`'${id}'`)) {
    throw new Error(`Missing seed agency matching triage fixture: ${id}`);
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
