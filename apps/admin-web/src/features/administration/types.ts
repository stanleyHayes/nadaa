import type {
  AgencySummary,
  AgencyType,
  AgencyUserProfile,
  AgencyUserRole,
  AlertSeverity,
  AlertStatus,
  AlertTargetType,
  AuditLogRecord,
  IntegrationContract,
  IntegrationDomain,
  IntegrationDirection,
} from "@nadaa/shared-types";

export type AdminLoadState = "loading" | "ready" | "error";

export type AgencyOperationalStatus = "active" | "pilot" | "review";

export interface ManagedAgency extends AgencySummary {
  status: AgencyOperationalStatus;
  users: number;
  openAssignments: number;
  mfaCoverage: number;
  dataScope: string;
  lastAuditAt: string;
}

export interface ManagedAgencyUser extends AgencyUserProfile {
  status: "active" | "mfa_pending" | "review";
  lastLoginAt?: string;
  accessScope: string;
}

export interface AdminMetric {
  label: string;
  value: string;
  detail: string;
  tone: "navy" | "green" | "gold" | "red";
}

export interface AdminUserFormState {
  name: string;
  email: string;
  phone: string;
  agencyId: string;
  role: AgencyUserRole;
}

export interface AdminActionResult {
  severity: "success" | "info" | "warning" | "error";
  message: string;
}

export interface DataSourceSummary {
  id: string;
  partner: string;
  domain: IntegrationDomain;
  direction: IntegrationDirection;
  status: IntegrationContract["status"];
  cadence: string;
  freshnessWindowMinutes: number;
  pii: "none" | "minimal_operational" | "aggregate_only";
  authenticationMode: string;
  secretScope?: string;
  owner: string;
  manualFallback: string;
  updatedAt: string;
}

export interface AlertRuleSummary {
  id: string;
  name: string;
  scope: string;
  targetType: AlertTargetType;
  severity: AlertSeverity;
  status: AlertStatus;
  approverRoles: AgencyUserRole[];
  emergencyOverrideRoles: AgencyUserRole[];
  mfaRequired: boolean;
  auditAction: string;
  lastReviewedAt: string;
}

export interface AdminOverviewData {
  agencies: ManagedAgency[];
  users: ManagedAgencyUser[];
  auditLogs: AuditLogRecord[];
  dataSources: DataSourceSummary[];
  alertRules: AlertRuleSummary[];
}

export type AgencyTypeLabelMap = Record<AgencyType, string>;
