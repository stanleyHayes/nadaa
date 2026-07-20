import type {
  AgencyType,
  AgencyUserRole,
  HazardType,
  HospitalCapacityStatus,
  IncidentAssignmentPriority,
  IncidentStatus,
  RiskLevel,
} from "@nadaa/shared-types";

export type DispatcherSession = {
  accessToken?: string;
  agencyId: string;
  agencyName: string;
  mfaCompleted: boolean;
  role: AgencyUserRole;
  userId: string;
  userName: string;
};

export type PermissionStatus = "unknown" | "granted" | "denied" | "blocked";

export type DispatcherPermissionState = {
  camera: PermissionStatus;
  location: PermissionStatus;
  push: PermissionStatus;
};

export type MobileLoadState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "offline"; message: string }
  | { status: "error"; message: string }
  | { status: "auth_expired"; message: string }
  | { status: "success"; message: string };

export type PushRegistrationState =
  | { status: "not_configured"; message: string }
  | { status: "permission_needed"; message: string }
  | { status: "registered"; provider: string; token: string }
  | { status: "failed"; message: string };

export type IncidentFilterState = {
  hazard: "all" | HazardType;
  severity: "all" | RiskLevel;
  status: "all" | IncidentStatus;
  time: "all" | "1h" | "6h" | "24h";
};

export type CapacityFilterState = {
  emergencyCapacity: "all" | HospitalCapacityStatus;
  includeStale: boolean;
  minAvailableBeds: string;
  service:
    | "all"
    | "emergency"
    | "trauma"
    | "icu"
    | "maternity"
    | "pediatric"
    | "ambulance"
    | "oxygen";
};

export type AuthFormState = {
  email: string;
  mfaCode: string;
  password: string;
};

export type StatusFormState = {
  note: string;
  resolutionNotes: string;
  status: IncidentStatus;
};

export type AssignmentFormState = {
  agencyId: string;
  agencyName: string;
  agencyType: AgencyType;
  instructions: string;
  priority: IncidentAssignmentPriority;
  responderLead: string;
};

export type TimelineNoteFormState = {
  note: string;
};

export type IncidentCachePayload = {
  cachedAt: string;
  incidents: unknown[];
};

export type CapacityCachePayload = {
  cachedAt: string;
  facilities: unknown[];
};

export type DispatcherMobileSnapshot = {
  capacityFilters: CapacityFilterState;
  filters: IncidentFilterState;
  permissions: DispatcherPermissionState;
  session: DispatcherSession;
};
