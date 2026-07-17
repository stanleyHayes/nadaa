import { AlertTriangle, CheckCheck, ShieldAlert, Truck } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyType,
  AgencyUserRole,
  AlertSeverity,
  AlertStatus,
  AlertTarget,
  AlertTargetGeometry,
  AlertTargetType,
  Coordinates,
  DuplicateReviewCandidate,
  HazardType,
  IncidentAbuseReviewDecision,
  IncidentAssignmentPriority,
  IncidentRecord,
  IncidentStatus,
  IncidentTimelineEvent,
  ReliefPointRecord,
  ReliefStockCategory,
  RiskLevel,
  ShelterRecord,
} from "@nadaa/shared-types";
import {
  alertTargetCatalog,
  assignmentAgencyOptions,
  incidentTransitionOptions,
  severityOrder,
  type AlertTargetCatalogItem,
} from "./data";
import type {
  AbuseReviewFormState,
  AlertFormState,
  AssignmentFormState,
  CommandIncident,
  FilterState,
  IncidentStatusFormState,
  ReliefPointFormState,
  ShelterFormState,
} from "./types";

export function buildQueueMetrics(incidents: CommandIncident[]) {
  return [
    {
      label: "New reports",
      value: incidents.filter(
        (incident) =>
          incident.status === "reported" || incident.status === "under_review",
      ).length,
      icon: ShieldAlert,
      color: nadaaBrand.colors.red,
    },
    {
      label: "Verified",
      value: incidents.filter(
        (incident) =>
          incident.status === "verified" || incident.status === "assigned",
      ).length,
      icon: CheckCheck,
      color: nadaaBrand.colors.green,
    },
    {
      label: "Teams en route",
      value: incidents.filter(
        (incident) =>
          incident.status === "response_en_route" ||
          incident.status === "on_scene",
      ).length,
      icon: Truck,
      color: "#0B6FB8",
    },
    {
      label: "Priority review",
      value: incidents.filter((incident) => incident.priorityReview).length,
      icon: AlertTriangle,
      color: nadaaBrand.colors.gold,
    },
  ];
}

export function buildFilterOptions(incidents: CommandIncident[]) {
  return {
    hazards: uniqueSorted(incidents.map((incident) => incident.type)),
    regionDistricts: uniqueSorted(
      incidents.map((incident) => `${incident.region} / ${incident.district}`),
    ),
    severities: uniqueSorted(
      incidents.map((incident) => incident.severity),
    ).sort((a, b) => severityOrder[b] - severityOrder[a]),
    statuses: uniqueSorted(incidents.map((incident) => incident.status)),
  };
}

export function matchesFilters(
  incident: CommandIncident,
  filters: FilterState,
) {
  if (filters.hazard !== "all" && incident.type !== filters.hazard) {
    return false;
  }
  if (
    filters.regionDistrict !== "all" &&
    `${incident.region} / ${incident.district}` !== filters.regionDistrict
  ) {
    return false;
  }
  if (filters.severity !== "all" && incident.severity !== filters.severity) {
    return false;
  }
  if (filters.status !== "all" && incident.status !== filters.status) {
    return false;
  }
  if (
    filters.time !== "all" &&
    !withinTimeWindow(incident.createdAt, filters.time)
  ) {
    return false;
  }
  return true;
}

export function withinTimeWindow(
  createdAt: string,
  timeFilter: FilterState["time"],
) {
  const hours =
    timeFilter === "1h"
      ? 1
      : timeFilter === "6h"
        ? 6
        : timeFilter === "24h"
          ? 24
          : 0;
  if (!hours) {
    return true;
  }
  const incidentTime = new Date(createdAt).getTime();
  return Date.now() - incidentTime <= hours * 60 * 60 * 1000;
}

export function duplicateReviewCandidatesFor(
  incident: CommandIncident | undefined,
  incidents: CommandIncident[],
): DuplicateReviewCandidate[] {
  if (!incident?.duplicateCandidates.length) {
    return [];
  }

  const reviewCandidates: DuplicateReviewCandidate[] = [];
  for (const candidate of incident.duplicateCandidates) {
    const candidateIncident = incidents.find(
      (item) =>
        item.id === candidate.incidentId &&
        !item.mergedIntoId &&
        item.status !== "false_report",
    );
    if (!candidateIncident) {
      continue;
    }
    reviewCandidates.push({
      candidate,
      incident: candidateIncident,
    });
  }
  return reviewCandidates;
}

export function enrichIncidentFromAPI(
  incident: IncidentRecord,
): CommandIncident {
  const normalizedIncident: IncidentRecord = {
    ...incident,
    duplicateCandidates: incident.duplicateCandidates ?? [],
    mergedIncidentIds: incident.mergedIncidentIds ?? [],
    assignments: incident.assignments ?? [],
    timeline: incident.timeline ?? [],
    media: incident.media ?? [],
    abuseSignals: incident.abuseSignals ?? [],
    abuseScore: incident.abuseScore ?? 0,
    abuseReviewRequired: incident.abuseReviewRequired ?? false,
  };
  const district = districtFromCoordinates(incident.location);
  return {
    ...normalizedIncident,
    region: district.region,
    district: district.district,
    locality: district.locality,
    assignedAgency: assignmentForIncident(normalizedIncident),
    responderEta: etaForIncident(normalizedIncident),
    timelineEntries: timelineEntriesForIncident(normalizedIncident),
    source: "api",
  };
}

export function districtFromCoordinates(location: {
  lat: number;
  lng: number;
}) {
  if (location.lng > -0.08) {
    return {
      region: "Greater Accra",
      district: "Tema Metropolitan",
      locality: "Tema",
    };
  }
  if (location.lng < -0.25) {
    return {
      region: "Greater Accra",
      district: "Ablekuma West",
      locality: "Ablekuma",
    };
  }
  if (location.lat < 5.56) {
    return {
      region: "Greater Accra",
      district: "Accra Metropolitan",
      locality: "Korle Gonno",
    };
  }
  return {
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Accra Central",
  };
}

export function assignmentForIncident(incident: IncidentRecord) {
  const activeAssignments = (incident.assignments ?? []).filter(
    (assignment) => assignment.status === "active",
  );
  if (activeAssignments.length) {
    return activeAssignments
      .map((assignment) => assignment.agencyName)
      .join(" + ");
  }
  if (
    incident.status === "reported" ||
    incident.status === "under_review" ||
    incident.status === "verified"
  ) {
    return incident.status === "verified"
      ? "Ready for assignment"
      : "Unassigned";
  }
  if (incident.type === "fire") {
    return "Ghana National Fire Service";
  }
  if (incident.type === "road_crash" || incident.type === "medical_emergency") {
    return "Ambulance + Police";
  }
  if (incident.type === "blocked_drain") {
    return "District Assembly";
  }
  return "NADMO District Desk";
}

export function timelineEntriesForIncident(incident: IncidentRecord) {
  if (incident.timeline?.length) {
    return [...incident.timeline]
      .sort(
        (a, b) =>
          new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
      )
      .map(formatTimelineEvent);
  }

  return [
    `${hazardLabel(incident.type)} report received from incident service`,
    `${statusLabel(incident.status)} status synchronized`,
    incident.verifiedAt
      ? `Verified by ${incident.verifiedBy || "authority user"}`
      : "",
    incident.statusReason ? `Latest note: ${incident.statusReason}` : "",
    incident.resolutionNotes ? `Resolution: ${incident.resolutionNotes}` : "",
    incident.abuseReviewRequired
      ? `Safety review: ${incident.abuseReviewReason || "dispatcher review required"}`
      : "",
    incident.priorityReview
      ? "Priority review flag is active"
      : "Dispatcher monitoring normal queue",
  ].filter(Boolean);
}

export function formatTimelineEvent(event: IncidentTimelineEvent) {
  const actor = event.actorRole ? ` / ${roleLabel(event.actorRole)}` : "";
  return `${formatShortTime(event.createdAt)} ${event.message}${actor}`;
}

export function etaForIncident(incident: IncidentRecord) {
  if (incident.severity === "emergency" || incident.severity === "severe") {
    return "5 min";
  }
  if (incident.priorityReview) {
    return "12 min";
  }
  return "30 min";
}

export function alertReadiness(incident: CommandIncident) {
  const severityWeight = severityOrder[incident.severity] * 14;
  const duplicateWeight = incident.duplicateCandidates.length ? 12 : 0;
  const mediaWeight = incident.media.length ? 8 : 0;
  return Math.min(95, 30 + severityWeight + duplicateWeight + mediaWeight);
}

export function buildDefaultStatusForm(
  incident?: CommandIncident,
): IncidentStatusFormState {
  const status = nextIncidentStatus(incident?.status ?? "reported");
  return {
    status,
    note: incident
      ? `${statusLabel(status)} update for ${incident.reference}.`
      : "Authority status update.",
    resolutionNotes: "",
  };
}

export function buildDefaultAbuseReviewForm(
  incident?: CommandIncident,
): AbuseReviewFormState {
  const decision = incident?.abuseReviewRequired ? "clear" : "monitor";
  return {
    decision,
    note: incident
      ? `${abuseDecisionLabel(decision)} safety review for ${incident.reference}.`
      : "Dispatcher safety review.",
    resolutionNotes: "",
  };
}

export function buildDefaultAssignmentForm(
  incident?: CommandIncident,
): AssignmentFormState {
  const activeAssignment = latestActiveAssignment(incident);
  if (activeAssignment) {
    return {
      agencyId: activeAssignment.agencyId,
      agencyName: activeAssignment.agencyName,
      agencyType: activeAssignment.agencyType,
      priority: activeAssignment.priority,
      instructions: activeAssignment.instructions,
      responderLead: activeAssignment.responderLead ?? "",
    };
  }

  const agency = suggestedAgencyForIncident(incident);
  const priority = assignmentPriorityForIncident(incident);
  return {
    agencyId: agency.id,
    agencyName: agency.name,
    agencyType: agency.type,
    priority,
    instructions: incident
      ? `Respond to ${hazardLabel(incident.type).toLowerCase()} incident ${incident.reference}. ${incident.description}`
      : "Respond to the selected incident and report field status.",
    responderLead: agency.responderLead,
  };
}

export function latestActiveAssignment(incident?: IncidentRecord) {
  const assignments = incident?.assignments ?? [];
  for (let index = assignments.length - 1; index >= 0; index -= 1) {
    const assignment = assignments[index];
    if (assignment?.status === "active") {
      return assignment;
    }
  }
  return undefined;
}

export function suggestedAgencyForIncident(incident?: IncidentRecord) {
  if (incident?.type === "fire" || incident?.type === "electrical_hazard") {
    return agencyOptionByType("fire");
  }
  if (
    incident?.type === "road_crash" ||
    incident?.type === "medical_emergency"
  ) {
    return agencyOptionByType("ambulance");
  }
  if (incident?.type === "blocked_drain") {
    return agencyOptionByType("district_assembly");
  }
  return assignmentAgencyOptions[0]!;
}

export function agencyOptionByType(type: AgencyType) {
  return (
    assignmentAgencyOptions.find((agency) => agency.type === type) ??
    assignmentAgencyOptions[0]!
  );
}

export function assignmentPriorityForIncident(
  incident?: IncidentRecord,
): IncidentAssignmentPriority {
  if (incident?.severity === "emergency" || incident?.severity === "severe") {
    return "urgent";
  }
  if (incident?.priorityReview || incident?.severity === "high") {
    return "high";
  }
  return "normal";
}

export function nextIncidentStatus(status: IncidentStatus): IncidentStatus {
  return incidentTransitionOptions[status][0] ?? status;
}

export function requiresIncidentResolution(status: IncidentStatus) {
  return status === "closed" || status === "false_report";
}

export function canAssignIncident(status: IncidentStatus) {
  return !["reported", "under_review", "closed", "false_report"].includes(
    status,
  );
}

export function buildAlertTarget(form: AlertFormState): AlertTarget {
  const type = form.targetType;
  const ids =
    type === "national" ? ["ghana"] : commaValues(form.targetIds || type);
  const target: AlertTarget = {
    type,
    ids,
    label: form.targetLabel.trim() || alertTargetTypeLabel(type),
  };

  if (type === "national") {
    return {
      ...target,
      center: { lat: 7.9465, lng: -1.0232 },
      radiusMeters: 365000,
      geometry: geometryFromBounds(4.54, -3.26, 11.18, 1.2),
      areaSqKm: 238533,
      estimatedPopulation: 33480000,
    };
  }

  if (type === "region" || type === "district" || type === "community") {
    const catalogItems = ids
      .map((id) => alertTargetCatalog[`${type}:${id}`])
      .filter(Boolean);
    if (catalogItems.length) {
      return enrichCatalogTarget(target, catalogItems);
    }
  }

  if (type === "radius") {
    const center = {
      lat: Number(form.targetLatitude) || 5.56,
      lng: Number(form.targetLongitude) || -0.2,
    };
    const radiusMeters = Number(form.targetRadiusMeters) || 5000;
    const areaSqKm = roundArea(Math.PI * (radiusMeters / 1000) ** 2);
    return {
      ...target,
      center,
      radiusMeters,
      areaSqKm,
      estimatedPopulation: Math.round(areaSqKm * 4500),
    };
  }

  if (type === "custom") {
    const geometry = parseTargetGeometry(form.targetGeometry);
    const center = geometry ? polygonCenter(geometry) : undefined;
    const areaSqKm = geometry ? polygonAreaSqKm(geometry) : 0;
    return {
      ...target,
      geometry,
      center,
      areaSqKm,
      estimatedPopulation: Math.round(areaSqKm * 5000),
    };
  }

  return target;
}

export function enrichCatalogTarget(
  target: AlertTarget,
  items: AlertTargetCatalogItem[],
): AlertTarget {
  const center = {
    lat: roundCoordinate(
      items.reduce((sum, item) => sum + item.center.lat, 0) / items.length,
    ),
    lng: roundCoordinate(
      items.reduce((sum, item) => sum + item.center.lng, 0) / items.length,
    ),
  };
  const areaSqKm = roundArea(
    items.reduce((sum, item) => sum + item.areaSqKm, 0),
  );
  return {
    ...target,
    label:
      target.label === alertTargetTypeLabel(target.type)
        ? items.map((item) => item.label).join(", ")
        : target.label,
    center,
    radiusMeters: Math.max(...items.map((item) => item.radiusMeters)),
    geometry: geometryFromCatalogItems(items),
    areaSqKm,
    estimatedPopulation: items.reduce(
      (sum, item) => sum + item.estimatedPopulation,
      0,
    ),
  };
}

export function geometryFromCatalogItems(
  items: AlertTargetCatalogItem[],
): AlertTargetGeometry {
  let minLat = 90;
  let minLng = 180;
  let maxLat = -90;
  let maxLng = -180;
  for (const item of items) {
    const { latDelta, lngDelta } = degreeDeltas(item.center, item.radiusMeters);
    minLat = Math.min(minLat, item.center.lat - latDelta);
    maxLat = Math.max(maxLat, item.center.lat + latDelta);
    minLng = Math.min(minLng, item.center.lng - lngDelta);
    maxLng = Math.max(maxLng, item.center.lng + lngDelta);
  }
  return geometryFromBounds(minLat, minLng, maxLat, maxLng);
}

export function geometryFromBounds(
  minLat: number,
  minLng: number,
  maxLat: number,
  maxLng: number,
): AlertTargetGeometry {
  return {
    type: "Polygon",
    coordinates: [
      [
        [roundCoordinate(minLng), roundCoordinate(minLat)],
        [roundCoordinate(maxLng), roundCoordinate(minLat)],
        [roundCoordinate(maxLng), roundCoordinate(maxLat)],
        [roundCoordinate(minLng), roundCoordinate(maxLat)],
        [roundCoordinate(minLng), roundCoordinate(minLat)],
      ],
    ],
  };
}

export function customGeometryAround(center: Coordinates): AlertTargetGeometry {
  return geometryFromBounds(
    center.lat - 0.02,
    center.lng - 0.025,
    center.lat + 0.02,
    center.lng + 0.025,
  );
}

export function parseTargetGeometry(
  value: string,
): AlertTargetGeometry | undefined {
  try {
    const parsed = JSON.parse(value) as AlertTargetGeometry;
    if (parsed?.type !== "Polygon" || !Array.isArray(parsed.coordinates)) {
      return undefined;
    }
    return parsed;
  } catch {
    return undefined;
  }
}

export function polygonCenter(
  geometry: AlertTargetGeometry,
): Coordinates | undefined {
  const ring = geometry.coordinates[0];
  if (!ring?.length) {
    return undefined;
  }
  const points = ring.slice(0, -1);
  if (!points.length) {
    return undefined;
  }
  return {
    lat: roundCoordinate(
      points.reduce((sum, point) => sum + point[1], 0) / points.length,
    ),
    lng: roundCoordinate(
      points.reduce((sum, point) => sum + point[0], 0) / points.length,
    ),
  };
}

export function polygonAreaSqKm(geometry: AlertTargetGeometry): number {
  const center = polygonCenter(geometry);
  const ring = geometry.coordinates[0];
  if (!center || !ring || ring.length < 4) {
    return 0;
  }
  let sum = 0;
  for (let index = 0; index < ring.length - 1; index += 1) {
    const [x1, y1] = lonLatToMeters(ring[index][0], ring[index][1], center.lat);
    const [x2, y2] = lonLatToMeters(
      ring[index + 1][0],
      ring[index + 1][1],
      center.lat,
    );
    sum += x1 * y2 - x2 * y1;
  }
  return roundArea(Math.abs(sum) / 2 / 1_000_000);
}

export function lonLatToMeters(lng: number, lat: number, referenceLat: number) {
  return [
    lng * 111_320 * Math.cos((referenceLat * Math.PI) / 180),
    lat * 110_540,
  ] as const;
}

export function degreeDeltas(center: Coordinates, radiusMeters: number) {
  const latDelta = radiusMeters / 111_320;
  const lngDelta =
    radiusMeters / (111_320 * Math.cos((center.lat * Math.PI) / 180));
  return {
    latDelta,
    lngDelta: Number.isFinite(lngDelta) ? lngDelta : latDelta,
  };
}

export function alertTargetSummary(target: AlertTarget) {
  if (target.type === "radius") {
    return `${metersLabel(target.radiusMeters ?? 0)} radius around ${target.label}`;
  }
  if (target.type === "custom") {
    return `${target.label} custom polygon`;
  }
  return `${target.label} ${alertTargetTypeLabel(target.type).toLowerCase()} target`;
}

export function alertTargetWarnings(target: AlertTarget) {
  const warnings: string[] = [];
  if (target.type === "national") {
    warnings.push(
      "National alerts should be reserved for broad life-safety threats.",
    );
  }
  if ((target.areaSqKm ?? 0) > 1000) {
    warnings.push(
      "Large target area may increase alert fatigue; confirm scope before approval.",
    );
  }
  if (target.type === "custom") {
    warnings.push(
      "Custom geometry should be checked against official boundaries before publishing.",
    );
  }
  return warnings;
}

export function alertTargetTypeLabel(type: AlertTargetType) {
  return type
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function metersLabel(value: number) {
  if (value >= 1000) {
    return `${Math.round(value / 100) / 10} km`;
  }
  return `${Math.round(value)} m`;
}

export function commaValues(value: string) {
  return value
    .split(",")
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean);
}

export function roundCoordinate(value: number) {
  return Math.round(value * 1_000_000) / 1_000_000;
}

export function roundArea(value: number) {
  return Math.round(value * 10) / 10;
}

export function buildDefaultAlertForm(
  incident?: CommandIncident,
): AlertFormState {
  const startsAt = new Date(Date.now() + 30 * 60 * 1000);
  const expiresAt = new Date(Date.now() + 12 * 60 * 60 * 1000);
  const hazard = incident ? hazardLabel(incident.type).toLowerCase() : "flood";
  const district = incident?.district ?? "Accra Metropolitan";
  const districtId = districtSlug(district);
  const severity = riskToAlertSeverity(incident?.severity ?? "high");
  const center = incident?.location ?? { lat: 5.56, lng: -0.2 };

  return {
    title: `${alertSeverityLabel(severity)} ${hazard} alert`,
    severity,
    message: incident
      ? `${incident.description} Avoid the affected area and follow official NADMO instructions.`
      : "Avoid low-lying roads and follow official NADMO instructions.",
    targetType: "district",
    targetIds: districtId,
    targetLabel: district,
    targetLatitude: `${center.lat}`,
    targetLongitude: `${center.lng}`,
    targetRadiusMeters: "5000",
    targetGeometry: JSON.stringify(customGeometryAround(center), null, 2),
    startsAt: formatDateTimeLocal(startsAt),
    expiresAt: formatDateTimeLocal(expiresAt),
    recommendedAction:
      incident?.severity === "emergency" || incident?.severity === "severe"
        ? "Prepare to evacuate if instructed by authorities."
        : "Stay alert, avoid the affected area, and monitor NADAA updates.",
    evacuationRequired: incident?.severity === "emergency",
    shelterIds: "00000000-0000-0000-0000-000000000301",
  };
}

export function riskToAlertSeverity(severity: RiskLevel): AlertSeverity {
  if (severity === "emergency") {
    return "emergency";
  }
  if (severity === "severe") {
    return "severe_warning";
  }
  if (severity === "high") {
    return "warning";
  }
  if (severity === "moderate") {
    return "watch";
  }
  return "advisory";
}

export function formatDateTimeLocal(date: Date) {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

export function uniqueSorted<T extends string>(values: T[]) {
  return [...new Set(values)].sort((a, b) => a.localeCompare(b));
}

export function hazardLabel(hazard: HazardType) {
  return hazard
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function severityLabel(severity: RiskLevel) {
  return severity[0].toUpperCase() + severity.slice(1);
}

export function alertSeverityLabel(severity: AlertSeverity) {
  return severity
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function statusLabel(status: IncidentStatus) {
  return status
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function abuseDecisionLabel(decision: IncidentAbuseReviewDecision) {
  if (decision === "false_report") {
    return "False report";
  }
  return decision[0].toUpperCase() + decision.slice(1);
}

export function abuseScoreLabel(score: number) {
  if (!score) {
    return "0% score";
  }
  return `${Math.round(score * 100)}% score`;
}

export function alertStatusLabel(status: AlertStatus) {
  return status
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function roleLabel(role: AgencyUserRole) {
  return role
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

export function formatShortTime(createdAt: string) {
  return new Intl.DateTimeFormat("en-GH", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(createdAt));
}

export function alertStatusColor(
  status: AlertStatus,
): "default" | "warning" | "success" | "error" {
  if (status === "approved" || status === "published") {
    return "success";
  }
  if (status === "submitted" || status === "draft") {
    return "warning";
  }
  if (status === "rejected" || status === "cancelled" || status === "expired") {
    return "error";
  }
  return "default";
}

export function districtSlug(district: string) {
  return district
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

export function formatIncidentAge(createdAt: string) {
  const minutes = Math.max(
    1,
    Math.round((Date.now() - new Date(createdAt).getTime()) / 60000),
  );
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  return `${hours} hr ${minutes % 60} min`;
}

export function buildDefaultShelterForm(
  shelter?: ShelterRecord,
): ShelterFormState {
  return {
    shelterId: shelter?.id ?? "",
    capacity: shelter ? `${shelter.capacity}` : "",
    currentOccupancy: shelter ? `${shelter.currentOccupancy}` : "",
    status: shelter?.status ?? "open",
    notes: shelter?.notes ?? "",
  };
}

export function buildDefaultReliefPointForm(
  point?: ReliefPointRecord,
): ReliefPointFormState {
  return {
    reliefPointId: point?.id ?? "__new__",
    name: point?.name ?? "",
    type: point?.type ?? "food",
    status: point?.status ?? "open",
    region: point?.region ?? "Greater Accra",
    district: point?.district ?? "",
    address: point?.address ?? "",
    latitude: point ? `${point.location.lat}` : "5.5600",
    longitude: point ? `${point.location.lng}` : "-0.2000",
    contact: point?.contact ?? "112",
    operatingHours: point?.operatingHours ?? "08:00-18:00",
    eligibility: point?.eligibility ?? "",
    schedule: point?.schedule ?? "Daily",
    stockCategories: formatReliefStockLines(point?.stockCategories ?? []),
    sourceRef: point?.sourceRef ?? "authority-dashboard",
  };
}

export function formatReliefStockLines(categories: ReliefStockCategory[]) {
  return categories
    .map((category) =>
      [category.category, category.quantity, category.unit].join(", "),
    )
    .join("\n");
}

export function parseReliefStockCategories(
  value: string,
): ReliefStockCategory[] {
  const now = new Date().toISOString();
  const categories = value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [category, quantityText, unit = "units"] = line
        .split(",")
        .map((part) => part.trim());
      const quantity = Number(quantityText);
      if (!category || !Number.isFinite(quantity) || quantity < 0) {
        throw new Error(
          "Stock lines must use category, quantity, unit per line.",
        );
      }
      return {
        category,
        quantity,
        unit: unit || "units",
        lastUpdated: now,
      };
    });

  if (!categories.length) {
    throw new Error("At least one stock line is required.");
  }
  return categories;
}

export function formatShortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-GH", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}
