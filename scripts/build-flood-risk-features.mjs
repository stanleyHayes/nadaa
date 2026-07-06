import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const dataDir = path.join(root, "data", "flood-risk");
const generatedDir = path.join(dataDir, "generated");
const fixturesPath = path.join(dataDir, "source-fixtures.v1.json");
const schemaPath = path.join(dataDir, "feature-schema.v1.json");
const jsonOutputPath = path.join(generatedDir, "features.v1.json");
const csvOutputPath = path.join(generatedDir, "features.v1.csv");
const manifestOutputPath = path.join(generatedDir, "manifest.v1.json");

const fixtures = JSON.parse(await readFile(fixturesPath, "utf8"));
const schema = JSON.parse(await readFile(schemaPath, "utf8"));
const columns = schema.columns.map((column) => column.name);

if (fixtures.datasetVersion !== schema.featureSetVersion) {
  throw new Error(
    `Fixture dataset version ${fixtures.datasetVersion} does not match schema ${schema.featureSetVersion}`,
  );
}

const rainfallByCell = indexByCell(fixtures.rainfallObservations);
const terrainByCell = indexByCell(fixtures.terrainSamples);
const hydrologyByCell = indexByCell(fixtures.hydrologySamples);
const landUseByCell = indexByCell(fixtures.landUseSamples);
const populationByCell = indexByCell(fixtures.populationSamples);
const reportsByCell = indexByCell(fixtures.historicalReports);
const knownZonesByID = new Map(
  fixtures.knownFloodZones.map((zone) => [zone.id, zone]),
);

const rows = fixtures.gridCells.map((cell) => {
  const rainfall = rainfallByCell.get(cell.cellId);
  const terrain = terrainByCell.get(cell.cellId);
  const hydrology = hydrologyByCell.get(cell.cellId);
  const landUse = landUseByCell.get(cell.cellId);
  const population = populationByCell.get(cell.cellId);
  const reports = reportsByCell.get(cell.cellId);
  const knownZone = cell.knownFloodZoneId
    ? knownZonesByID.get(cell.knownFloodZoneId)
    : null;

  const missingRainfall = !rainfall;
  const missingTerrain = !terrain;
  const missingHydrology = !hydrology;
  const missingLandUse = !landUse;
  const missingPopulation = !population;
  const missingHistoricalReports = !reports;
  const missingGeometry = !cell.geometry;

  const rainfall24hMm = rainfall?.rainfall24hMm ?? 0;
  const rainfall72hMm = rainfall?.rainfall72hMm ?? 0;
  const rainfallForecast24hMm = rainfall?.rainfallForecast24hMm ?? 0;
  const elevationM = terrain?.elevationM ?? 0;
  const slopeDegrees = terrain?.slopeDegrees ?? 0;
  const distanceToDrainM = hydrology?.distanceToDrainM ?? 5000;
  const distanceToRiverM = hydrology?.distanceToRiverM ?? 5000;
  const waterLevelTrendCm = hydrology?.waterLevelTrendCm ?? 0;
  const drainageDensityKmPerSqKm = hydrology?.drainageDensityKmPerSqKm ?? 0;
  const imperviousSurfacePct = landUse?.imperviousSurfacePct ?? 0;
  const populationDensityPerSqKm = population?.populationDensityPerSqKm ?? 0;
  const vulnerablePopulationPct = population?.vulnerablePopulationPct ?? 0;
  const historicalFloodReports30d = reports?.floodReports30d ?? 0;
  const daysSinceLastFloodReport = reports?.daysSinceLastFloodReport ?? 999;
  const insideKnownFloodZone = Boolean(knownZone);
  const lowLyingArea = Boolean(terrain?.lowLyingArea);

  const rainfallIntensityScore = roundScore(
    0.45 * normalize(rainfall24hMm, 0, 100) +
      0.25 * normalize(rainfall72hMm, 0, 180) +
      0.3 * normalize(rainfallForecast24hMm, 0, 120),
  );
  const exposureScore = roundScore(
    0.45 * normalize(populationDensityPerSqKm, 0, 16000) +
      0.35 * normalize(imperviousSurfacePct, 0, 100) +
      0.2 * normalize(vulnerablePopulationPct, 0, 30),
  );
  const drainagePressureScore = roundScore(
    0.22 * inverseNormalize(elevationM, 0, 80) +
      0.14 * inverseNormalize(slopeDegrees, 0, 12) +
      0.2 * inverseNormalize(distanceToDrainM, 0, 1200) +
      0.16 * inverseNormalize(distanceToRiverM, 0, 3500) +
      0.18 * normalize(waterLevelTrendCm, 0, 35) +
      0.1 * inverseNormalize(drainageDensityKmPerSqKm, 0, 4),
  );
  const historicalSignalScore = roundScore(
    0.65 * normalize(historicalFloodReports30d, 0, 5) +
      0.35 * inverseNormalize(daysSinceLastFloodReport, 0, 30),
  );
  const zoneBoost = insideKnownFloodZone ? 0.08 : 0;
  const compositeRuleScore = roundScore(
    Math.min(
      1,
      0.35 * rainfallIntensityScore +
        0.3 * drainagePressureScore +
        0.2 * exposureScore +
        0.15 * historicalSignalScore +
        zoneBoost,
    ),
  );

  return orderRow({
    feature_set_version: fixtures.datasetVersion,
    source: fixtures.source,
    source_updated_at: fixtures.sourceUpdatedAt,
    generated_at: fixtures.generatedAt,
    valid_from: fixtures.validFrom,
    valid_to: fixtures.validTo,
    cell_id: cell.cellId,
    region: cell.region,
    district: cell.district,
    community: cell.community,
    centroid_lat: cell.centroid.lat,
    centroid_lng: cell.centroid.lng,
    geometry: cell.geometry,
    district_geometry_source: cell.districtGeometrySource,
    rainfall_24h_mm: rainfall24hMm,
    rainfall_72h_mm: rainfall72hMm,
    rainfall_forecast_24h_mm: rainfallForecast24hMm,
    elevation_m: elevationM,
    slope_degrees: slopeDegrees,
    distance_to_drain_m: distanceToDrainM,
    distance_to_river_m: distanceToRiverM,
    water_level_trend_cm: waterLevelTrendCm,
    drainage_density_km_per_sq_km: drainageDensityKmPerSqKm,
    dominant_land_use: landUse?.dominantLandUse ?? "unknown",
    impervious_surface_pct: imperviousSurfacePct,
    population_density_per_sq_km: populationDensityPerSqKm,
    vulnerable_population_pct: vulnerablePopulationPct,
    historical_flood_reports_30d: historicalFloodReports30d,
    days_since_last_flood_report: daysSinceLastFloodReport,
    inside_known_flood_zone: insideKnownFloodZone,
    low_lying_area: lowLyingArea,
    rainfall_intensity_score: rainfallIntensityScore,
    exposure_score: exposureScore,
    drainage_pressure_score: drainagePressureScore,
    historical_signal_score: historicalSignalScore,
    composite_rule_score: compositeRuleScore,
    label_severity: severityFromScore(compositeRuleScore),
    missing_rainfall: missingRainfall,
    missing_terrain: missingTerrain,
    missing_hydrology: missingHydrology,
    missing_land_use: missingLandUse,
    missing_population: missingPopulation,
    missing_historical_reports: missingHistoricalReports,
    missing_geometry: missingGeometry,
  });
});

await mkdir(generatedDir, { recursive: true });

const featurePayload = {
  featureSetVersion: fixtures.datasetVersion,
  schemaVersion: schema.schemaVersion,
  source: fixtures.source,
  sourceUpdatedAt: fixtures.sourceUpdatedAt,
  generatedAt: fixtures.generatedAt,
  validFrom: fixtures.validFrom,
  validTo: fixtures.validTo,
  rowCount: rows.length,
  limitations: fixtures.sourceLimitations,
  rows,
};
const jsonContent = `${JSON.stringify(featurePayload, null, 2)}\n`;
const csvContent = buildCSV(rows, columns);

await writeFile(jsonOutputPath, jsonContent);
await writeFile(csvOutputPath, csvContent);

const manifest = {
  featureSetVersion: fixtures.datasetVersion,
  schemaVersion: schema.schemaVersion,
  sourceFixture: path.relative(root, fixturesPath),
  schema: path.relative(root, schemaPath),
  generatedAt: fixtures.generatedAt,
  sourceUpdatedAt: fixtures.sourceUpdatedAt,
  validFrom: fixtures.validFrom,
  validTo: fixtures.validTo,
  rowCount: rows.length,
  columnCount: columns.length,
  columns,
  outputs: [
    {
      path: path.relative(root, jsonOutputPath),
      format: "json",
      sha256: sha256(jsonContent),
    },
    {
      path: path.relative(root, csvOutputPath),
      format: "csv",
      sha256: sha256(csvContent),
    },
  ],
  limitations: fixtures.sourceLimitations,
};

await writeFile(manifestOutputPath, `${JSON.stringify(manifest, null, 2)}\n`);

console.log(
  `Generated ${rows.length} flood-risk feature rows in ${path.relative(
    root,
    generatedDir,
  )}.`,
);

function indexByCell(records) {
  return new Map(records.map((record) => [record.cellId, record]));
}

function normalize(value, min, max) {
  if (max <= min) {
    return 0;
  }
  return clamp((value - min) / (max - min), 0, 1);
}

function inverseNormalize(value, min, max) {
  return 1 - normalize(value, min, max);
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function roundScore(value) {
  return Number(clamp(value, 0, 1).toFixed(4));
}

function severityFromScore(score) {
  if (score >= 0.78) {
    return "severe";
  }
  if (score >= 0.58) {
    return "high";
  }
  if (score >= 0.35) {
    return "moderate";
  }
  return "low";
}

function orderRow(row) {
  const ordered = {};
  for (const column of columns) {
    if (!(column in row)) {
      throw new Error(`Generated row is missing schema column: ${column}`);
    }
    ordered[column] = row[column];
  }

  const unknownColumns = Object.keys(row).filter(
    (column) => !columns.includes(column),
  );
  if (unknownColumns.length > 0) {
    throw new Error(
      `Generated row has unknown columns: ${unknownColumns.join(", ")}`,
    );
  }

  return ordered;
}

function buildCSV(records, header) {
  const lines = [header.join(",")];
  for (const record of records) {
    lines.push(header.map((column) => csvCell(record[column])).join(","));
  }
  return `${lines.join("\n")}\n`;
}

function csvCell(value) {
  if (value === null || value === undefined) {
    return "";
  }
  const text =
    typeof value === "object" ? JSON.stringify(value) : String(value);
  if (/["\n,]/.test(text)) {
    return `"${text.replaceAll('"', '""')}"`;
  }
  return text;
}

function sha256(content) {
  return createHash("sha256").update(content).digest("hex");
}
