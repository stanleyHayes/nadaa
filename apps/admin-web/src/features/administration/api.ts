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
import {
  ALERT_API_BASE,
  AUTH_API_BASE,
  INTEGRATION_API_BASE,
} from "@/app/config";

/** Payload of `GET /auth/agencies` (not yet exported from shared-types). */
interface AgencyListResponse {
  agencies: AgencySummary[];
}

export async function fetchAuditLogs(signal?: AbortSignal) {
  const response = await fetch(`${AUTH_API_BASE}/audit/logs?limit=25`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  if (!response.ok) {
    throw new Error(`audit API returned ${response.status}`);
  }

  const payload = (await response.json()) as AuditLogListResponse;
  return payload.logs;
}

/**
 * Load the agency directory. Restricted to system_admin tokens with MFA, so a
 * lesser role surfaces as a failed governance surface rather than empty data.
 */
export async function fetchAgencies(signal?: AbortSignal) {
  const response = await fetch(`${AUTH_API_BASE}/auth/agencies`, {
    headers: adminHeaders(),
    signal,
  });
  handleUnauthorized(response);
  if (!response.ok) {
    throw new Error(`agencies API returned ${response.status}`);
  }

  const payload = (await response.json()) as AgencyListResponse;
  return payload.agencies.map(managedAgencyFromSummary);
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
