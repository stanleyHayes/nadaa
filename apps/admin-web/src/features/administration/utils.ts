import { nadaaBrand, severityRoles, type Severity } from "@nadaa/brand";
import type {
  AgencyType,
  AgencyUserRole,
  AlertSeverity,
  AuditLogRecord,
  AuthorityAlertRecord,
  CreateAgencyUserResponse,
  IntegrationContract,
} from "@nadaa/shared-types";
import type {
  AdminMetric,
  AdminUserFormState,
  AlertRuleSummary,
  DataSourceSummary,
  ManagedAgency,
  ManagedAgencyUser,
} from "./types";

export const roleOptions: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

const roleLabels: Record<AgencyUserRole, string> = {
  system_admin: "System admin",
  agency_admin: "Agency admin",
  nadmo_officer: "NADMO officer",
  district_officer: "District officer",
  dispatcher: "Dispatcher",
  responder: "Responder",
  agency_viewer: "Agency viewer",
};

const agencyTypeLabels: Record<AgencyType, string> = {
  nadmo: "NADMO",
  district_assembly: "District assembly",
  police: "Police",
  fire: "Fire service",
  ambulance: "Ambulance",
  meteorological: "Meteorological",
  hydrological: "Hydrological",
  hospital: "Hospital",
  utility: "Utility",
  ngo: "NGO",
  other: "Other",
};

export const toneColors = {
  navy: nadaaBrand.colors.navy,
  green: nadaaBrand.colors.green,
  gold: nadaaBrand.colors.gold,
  red: nadaaBrand.colors.red,
} as const;

export function roleLabel(role: AgencyUserRole) {
  return roleLabels[role];
}

export function agencyTypeLabel(type: AgencyType) {
  return agencyTypeLabels[type];
}

export function formatDateTime(value?: string) {
  if (!value) {
    return "Not recorded";
  }

  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function formatPercent(value: number) {
  return `${Math.round(value)}%`;
}

export function buildAdminMetrics(
  agencies: ManagedAgency[],
  users: ManagedAgencyUser[],
  auditLogs: AuditLogRecord[],
  dataSources: DataSourceSummary[],
): AdminMetric[] {
  const mfaReady = users.filter((user) => user.mfaEnabled).length;
  const activeDataSources = dataSources.filter(
    (source) => source.status === "pilot" || source.status === "production",
  ).length;

  return [
    {
      label: "Agencies",
      value: `${agencies.length}`,
      detail: `${agencies.filter((agency) => agency.status === "active").length} active, ${agencies.filter((agency) => agency.status === "pilot").length} pilot`,
      tone: "navy",
    },
    {
      label: "Authority users",
      value: `${users.length}`,
      detail: `${mfaReady}/${users.length || 1} MFA ready`,
      tone: "green",
    },
    {
      label: "Audit events",
      value: `${auditLogs.length}`,
      detail: auditLogs[0]
        ? `Latest ${formatDateTime(auditLogs[0].createdAt)}`
        : "No recent audit events",
      tone: "gold",
    },
    {
      label: "Data sources",
      value: `${dataSources.length}`,
      detail: `${activeDataSources} pilot or production contracts`,
      tone: activeDataSources ? "green" : "red",
    },
  ];
}

export function mfaCoverageFor(users: ManagedAgencyUser[], agencyId: string) {
  const agencyUsers = users.filter((user) => user.agency.id === agencyId);
  if (!agencyUsers.length) {
    return 0;
  }

  return (
    (agencyUsers.filter((user) => user.mfaEnabled).length /
      agencyUsers.length) *
    100
  );
}

export function validateUserForm(form: AdminUserFormState) {
  if (!form.name.trim()) {
    return "Name is required.";
  }
  if (!form.email.includes("@")) {
    return "A valid email address is required.";
  }
  if (!form.phone.startsWith("+233") || form.phone.length < 8) {
    return "Use an E.164 Ghana phone number, for example +233200000000.";
  }
  if (!form.agencyId) {
    return "Agency is required.";
  }
  return "";
}

export function managedUserFromCreateResponse(
  response: CreateAgencyUserResponse,
): ManagedAgencyUser {
  return {
    ...response.user,
    status: "mfa_pending",
    accessScope: `${roleLabel(response.user.role)} access`,
  };
}

export function auditTargetSummary(log: AuditLogRecord) {
  if (!log.targetId) {
    return log.targetType;
  }
  return `${log.targetType} ${log.targetId}`;
}

export function auditSnapshotSummary(log: AuditLogRecord) {
  const source = log.after ?? log.before;
  if (!source) {
    return "No snapshot attached";
  }

  const entries = Object.entries(source)
    .filter(([key]) => !/token|password|secret|otp|code/i.test(key))
    .slice(0, 3);

  if (!entries.length) {
    return "Sensitive snapshot hidden";
  }

  return entries.map(([key, value]) => `${key}: ${String(value)}`).join(", ");
}

export function dataSourceFromContract(
  contract: IntegrationContract,
): DataSourceSummary {
  const firstPayload = contract.payloads[0];

  return {
    id: contract.id,
    partner: contract.partner,
    domain: contract.domain,
    direction: contract.direction,
    status: contract.status,
    cadence: contract.cadence,
    freshnessWindowMinutes: contract.freshnessWindowMinutes,
    pii: firstPayload?.pii ?? "none",
    authenticationMode: contract.authentication.mode,
    secretScope: contract.authentication.secretScope,
    owner: contract.dataOwner,
    manualFallback: contract.failureBehavior.manualFallback,
    updatedAt: contract.updatedAt,
  };
}

export function buildAlertRulesFromAlerts(
  alerts: AuthorityAlertRecord[],
): AlertRuleSummary[] {
  return alerts.slice(0, 4).map((alert) => ({
    id: `rule-${alert.id}`,
    name: `${alert.title} governance`,
    scope: alert.target.label,
    targetType: alert.target.type,
    severity: alert.severity,
    status: alert.status,
    approverRoles: ["system_admin", "agency_admin", "nadmo_officer"],
    emergencyOverrideRoles: ["system_admin", "nadmo_officer"],
    mfaRequired: true,
    auditAction: alert.emergencyOverride
      ? "alert.emergency_override"
      : "alert.workflow_review",
    lastReviewedAt: alert.updatedAt,
  }));
}

export function statusColor(status: string) {
  if (
    ["active", "production", "approved", "published", "ready"].includes(status)
  ) {
    return "success";
  }
  if (["pilot", "submitted", "mfa_pending", "mock_contract"].includes(status)) {
    return "warning";
  }
  if (["review", "rejected", "failed"].includes(status)) {
    return "error";
  }
  return "default";
}

const alertSeverityRoleMap: Record<AlertSeverity, Severity> = {
  advisory: "info",
  watch: "low",
  warning: "medium",
  severe_warning: "high",
  emergency: "severe",
};

export function alertSeverityRole(severity: AlertSeverity) {
  return severityRoles[alertSeverityRoleMap[severity]];
}
