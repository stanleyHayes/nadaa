import type {
  AgencySummary,
  AlertListResponse,
  AuditLogListResponse,
  IntegrationContractListResponse,
} from "@nadaa/shared-types";
import { adminHeaders } from "@/app/session";
import { handleUnauthorized } from "@/app/http";
import {
  buildAlertRulesFromAlerts,
  dataSourceFromContract,
  managedAgencyFromSummary,
} from "./utils";
import type { AgencyUserDirectoryEntry } from "./types";
import {
  ALERT_API_BASE,
  AUTH_API_BASE,
  INTEGRATION_API_BASE,
} from "@/app/config";

/** Payload of `GET /auth/agencies` (not yet exported from shared-types). */
interface AgencyListResponse {
  agencies: AgencySummary[];
}

/** Payload of `GET /auth/agency-users` (not yet exported from shared-types). */
interface AgencyUserListResponse {
  users: AgencyUserDirectoryEntry[];
}

/**
 * Raised when the signed-in admin's role is not allowed on a governance
 * surface (HTTP 403). Unlike a transport failure this is expected for scoped
 * roles — e.g. an agency_admin reading the system_admin-only agency directory
 * — so views render a scoped "requires system admin" state instead of a
 * console-wide load error.
 */
export class GovernanceForbiddenError extends Error {
  constructor(surface: string) {
    super(`${surface} requires a system admin session.`);
    this.name = "GovernanceForbiddenError";
  }
}

/** Throw the scoped-role error for 403s on system_admin-only surfaces. */
function throwIfForbidden(response: Response, surface: string): void {
  if (response.status === 403) {
    throw new GovernanceForbiddenError(surface);
  }
}

export async function fetchAuditLogs(signal?: AbortSignal) {
  const response = await fetch(`${AUTH_API_BASE}/audit/logs?limit=25`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  throwIfForbidden(response, "The audit trail");
  if (!response.ok) {
    throw new Error(`audit API returned ${response.status}`);
  }

  const payload = (await response.json()) as AuditLogListResponse;
  return payload.logs;
}

/**
 * Load the agency directory. Restricted to system_admin tokens with MFA, so a
 * lesser role surfaces as a scoped forbidden state rather than empty data.
 */
export async function fetchAgencies(signal?: AbortSignal) {
  const response = await fetch(`${AUTH_API_BASE}/auth/agencies`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  throwIfForbidden(response, "The agency directory");
  if (!response.ok) {
    throw new Error(`agencies API returned ${response.status}`);
  }

  const payload = (await response.json()) as AgencyListResponse;
  return payload.agencies.map(managedAgencyFromSummary);
}

/**
 * Load the authority-users directory. system_admin receives every agency's
 * users; agency_admin receives only their own agency's. Entries carry identity
 * fields only — the caller resolves agency display names.
 */
export async function fetchAgencyUsers(
  signal?: AbortSignal,
): Promise<AgencyUserDirectoryEntry[]> {
  const response = await fetch(`${AUTH_API_BASE}/auth/agency-users`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  if (!response.ok) {
    throw new Error(`users directory API returned ${response.status}`);
  }

  const payload = (await response.json()) as AgencyUserListResponse;
  return payload.users;
}

export async function fetchDataSources(signal?: AbortSignal) {
  const response = await fetch(
    `${INTEGRATION_API_BASE}/integrations/contracts`,
    {
      signal,
    },
  );
  handleUnauthorized(response);
  if (!response.ok) {
    throw new Error(`integration API returned ${response.status}`);
  }

  const payload = (await response.json()) as IntegrationContractListResponse;
  return payload.contracts.map(dataSourceFromContract);
}

export async function fetchAlertRules(signal?: AbortSignal) {
  const response = await fetch(`${ALERT_API_BASE}/alerts`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  if (!response.ok) {
    throw new Error(`alert API returned ${response.status}`);
  }

  const payload = (await response.json()) as AlertListResponse;
  return buildAlertRulesFromAlerts(payload.alerts);
}
