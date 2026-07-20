import type {
  AgencySummary,
  AgencyType,
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

/**
 * Operational status of an agency. The directory API does not report one, so
 * agencies load as "unknown" rather than with a fabricated status.
 */
export type AgencyOperationalStatus = "active" | "pilot" | "review" | "unknown";

export interface ManagedAgency extends AgencySummary {
  status: AgencyOperationalStatus;
  /** Authority users in the agency; null while the users directory is unavailable. */
  users: number | null;
  /** Open dispatch assignments; no governance API reports these yet (null). */
  openAssignments: number | null;
  /** Share of the agency's users with MFA enrolled; null when unknown. */
  mfaCoverage: number | null;
  dataScope: string;
  /** Latest audit event for the agency; empty when no audit data is loaded. */
  lastAuditAt: string;
}

/**
 * An authority user row in the console. Built either from the users directory
 * (`GET /auth/agency-users`, which exposes identity fields only) or from the
 * fuller profile returned by user provisioning, so only shared display fields
 * are guaranteed here.
 */
export interface ManagedAgencyUser {
  id: string;
  name: string;
  email: string;
  role: AgencyUserRole;
  agency: Pick<AgencySummary, "id" | "name">;
  mfaEnabled: boolean;
  createdAt?: string;
  /** Set while the account is locked after too many failed attempts. */
  lockedUntil?: string;
  status: "active" | "mfa_pending" | "review" | "locked";
  lastLoginAt?: string;
  accessScope: string;
}

/**
 * Entry of `GET /auth/agency-users` (contract landed with the auth-service
 * directory endpoint; not yet exported from shared-types). system_admin
 * receives every agency's users, agency_admin only their own.
 */
export interface AgencyUserDirectoryEntry {
  id: string;
  name: string;
  email: string;
  role: AgencyUserRole;
  agencyId: string;
  mfaEnabled: boolean;
  lockedUntil?: string;
  createdAt: string;
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

/**
 * One-time credential bundle returned by user provisioning. Shown exactly once
 * in the create-user dialog, then discarded — the auth service never returns
 * it again. The user id is included because MFA setup addresses the account
 * by id.
 */
export interface CreatedUserCredentials {
  userId: string;
  name: string;
  email: string;
  temporaryPassword: string;
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
