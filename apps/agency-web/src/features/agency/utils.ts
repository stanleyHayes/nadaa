import type {
  AidRequestPriority,
  AidRequestRecord,
  AidRequestStatus,
  HospitalCapacityRecord,
  HospitalCapacityStatus,
  IncidentRecord,
  IncidentStatus,
  ReliefPointRecord,
  ReliefPointStatus,
  ReliefStockCategory,
  RiskLevel,
} from "@nadaa/shared-types";
import { incidentTransitionOptions, severityOrder } from "./data";
import type {
  AidRequestFormState,
  IncidentFilterState,
  ReliefPointFormState,
} from "./types";

export function allowedTransitions(status: IncidentStatus) {
  return incidentTransitionOptions[status] ?? [];
}

export function matchesFilters(
  incident: IncidentRecord,
  filters: IncidentFilterState,
) {
  if (filters.hazard !== "all" && incident.type !== filters.hazard) {
    return false;
  }
  if (filters.severity !== "all" && incident.severity !== filters.severity) {
    return false;
  }
  if (filters.status !== "all" && incident.status !== filters.status) {
    return false;
  }
  return true;
}

export function severityColor(severity: RiskLevel) {
  switch (severity) {
    case "emergency":
    case "severe":
      return "error";
    case "high":
      return "warning";
    case "moderate":
      return "info";
    case "low":
    default:
      return "success";
  }
}

export function hospitalCapacityColor(capacity: HospitalCapacityStatus) {
  switch (capacity) {
    case "available":
      return "success";
    case "limited":
      return "warning";
    case "full":
      return "error";
    default:
      return "default";
  }
}

export function hospitalBedPercent(facility: HospitalCapacityRecord) {
  if (!facility.totalBeds) return 0;
  return Math.round(
    ((facility.totalBeds - facility.availableBeds) / facility.totalBeds) * 100,
  );
}

export function reliefStatusColor(status: ReliefPointStatus) {
  switch (status) {
    case "open":
      return "success";
    case "limited":
      return "warning";
    case "closed":
      return "error";
    default:
      return "default";
  }
}

export function reliefLabel(value: string) {
  return value
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function aidLabel(value: string) {
  return reliefLabel(value);
}

export function aidStatusColor(status: AidRequestStatus) {
  switch (status) {
    case "pending_review":
      return "warning";
    case "approved":
    case "open":
    case "partially_matched":
      return "success";
    case "fulfilled":
      return "info";
    case "rejected":
    case "closed":
      return "error";
    default:
      return "default";
  }
}

export function aidPriorityColor(priority: AidRequestPriority) {
  switch (priority) {
    case "urgent":
      return "error";
    case "high":
      return "warning";
    case "medium":
      return "info";
    default:
      return "default";
  }
}

export function aidProgressPercent(request: AidRequestRecord) {
  if (!request.quantityNeeded) return 0;
  return Math.min(
    100,
    Math.round((request.quantityPledged / request.quantityNeeded) * 100),
  );
}

export function stockSummary(categories: ReliefStockCategory[]) {
  if (!categories.length) {
    return "No stock categories recorded";
  }
  return categories
    .map(
      (item) =>
        `${reliefLabel(item.category)}: ${item.quantity.toLocaleString("en-GH")} ${item.unit}`,
    )
    .join(" · ");
}

export function parseStockCategories(value: string): ReliefStockCategory[] {
  const now = new Date().toISOString();
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      const [category, quantityText, unitText] = item
        .split(":")
        .map((part) => part.trim());
      const quantity = Number.parseInt(quantityText ?? "0", 10);
      return {
        category,
        quantity:
          Number.isFinite(quantity) && !Number.isNaN(quantity)
            ? Math.max(0, quantity)
            : 0,
        unit: unitText || "units",
        lastUpdated: now,
      };
    })
    .filter((item) => item.category);
}

export function serializeStockCategories(categories: ReliefStockCategory[]) {
  return categories
    .map((item) => `${item.category}:${item.quantity}:${item.unit}`)
    .join(", ");
}

export function reliefPointToForm(
  point: ReliefPointRecord,
): ReliefPointFormState {
  return {
    name: point.name,
    type: point.type,
    region: point.region,
    district: point.district,
    address: point.address,
    lat: point.location.lat.toString(),
    lng: point.location.lng.toString(),
    contact: point.contact,
    operatingHours: point.operatingHours,
    eligibility: point.eligibility,
    schedule: point.schedule,
    status: point.status,
    stockCategories: serializeStockCategories(point.stockCategories),
  };
}

export function aidRequestToForm(
  request: AidRequestRecord,
): AidRequestFormState {
  return {
    title: request.title,
    category: request.category,
    priority: request.priority,
    region: request.region,
    district: request.district,
    lat: request.location.lat.toString(),
    lng: request.location.lng.toString(),
    receivingOrganization: request.receivingOrganization,
    contact: request.contact,
    quantityNeeded: request.quantityNeeded.toString(),
    quantityUnit: request.quantityUnit,
    description: request.description,
    neededBy: request.neededBy.slice(0, 16),
    visibility: request.visibility,
    sourceReliefPointId: request.sourceReliefPointId ?? "",
  };
}
