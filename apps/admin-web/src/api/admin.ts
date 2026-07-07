import type {
  AlertListResponse,
  AuditLogListResponse,
  IntegrationContractListResponse,
} from "@nadaa/shared-types";
import { adminHeaders } from "../auth/session";
import {
  buildAlertRulesFromAlerts,
  dataSourceFromContract,
} from "../lib/utils";
import { ALERT_API_BASE, AUTH_API_BASE, INTEGRATION_API_BASE } from "./config";

export async function fetchAuditLogs(signal?: AbortSignal) {
  const response = await fetch(`${AUTH_API_BASE}/audit/logs?limit=25`, {
    headers: adminHeaders(),
    signal,
  });
  if (!response.ok) {
    throw new Error(`audit API returned ${response.status}`);
  }

  const payload = (await response.json()) as AuditLogListResponse;
  return payload.logs;
}

export async function fetchDataSources(signal?: AbortSignal) {
  const response = await fetch(
    `${INTEGRATION_API_BASE}/integrations/contracts`,
    {
      signal,
    },
  );
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
  if (!response.ok) {
    throw new Error(`alert API returned ${response.status}`);
  }

  const payload = (await response.json()) as AlertListResponse;
  return buildAlertRulesFromAlerts(payload.alerts);
}
