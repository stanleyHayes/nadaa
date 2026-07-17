import { type ChangeEvent, useEffect, useState } from "react";
import type {
  AuditLogRecord,
  CreateAgencyUserResponse,
} from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";
import { adminHeaders } from "@/app/session";
import { handleUnauthorized } from "@/app/http";
import {
  fetchAgencies,
  fetchAlertRules,
  fetchAuditLogs,
  fetchDataSources,
} from "./api";
import { defaultUserForm } from "./data";
import type {
  AdminActionResult,
  AdminLoadState,
  AdminUserFormState,
  AlertRuleSummary,
  CreatedUserCredentials,
  DataSourceSummary,
  ManagedAgency,
  ManagedAgencyUser,
} from "./types";
import { managedUserFromCreateResponse, validateUserForm } from "./utils";

export type AdminData = ReturnType<typeof useAdminData>;

/**
 * Central governance-console state container. Loads the agency directory and
 * the audit, integration, and alert-rule surfaces the admin views depend on
 * from the live backends so the shell can mount data once and route between
 * views without losing state. Collections start empty; a failed surface stays
 * empty and the console surfaces a concise error state instead of
 * substituting fixture data.
 */
export function useAdminData() {
  const [loadState, setLoadState] = useState<AdminLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading governance data");
  const [agencies, setAgencies] = useState<ManagedAgency[]>([]);
  const [users, setUsers] = useState<ManagedAgencyUser[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLogRecord[]>([]);
  const [dataSources, setDataSources] = useState<DataSourceSummary[]>([]);
  const [alertRules, setAlertRules] = useState<AlertRuleSummary[]>([]);
  const [userForm, setUserForm] = useState<AdminUserFormState>(defaultUserForm);
  const [createBusy, setCreateBusy] = useState(false);
  const [actionResult, setActionResult] = useState<AdminActionResult>();
  const [createdCredentials, setCreatedCredentials] =
    useState<CreatedUserCredentials | null>(null);

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading governance data");

    const [agencyResult, auditResult, sourceResult, alertResult] =
      await Promise.allSettled([
        fetchAgencies(signal),
        fetchAuditLogs(signal),
        fetchDataSources(signal),
        fetchAlertRules(signal),
      ]);

    if (signal?.aborted) {
      return;
    }

    let failureCount = 0;
    if (agencyResult.status === "fulfilled") {
      setAgencies(agencyResult.value);
    } else {
      failureCount += 1;
      setAgencies([]);
    }

    if (auditResult.status === "fulfilled") {
      setAuditLogs(auditResult.value);
    } else {
      failureCount += 1;
      setAuditLogs([]);
    }

    if (sourceResult.status === "fulfilled") {
      setDataSources(sourceResult.value);
    } else {
      failureCount += 1;
      setDataSources([]);
    }

    if (alertResult.status === "fulfilled") {
      setAlertRules(alertResult.value);
    } else {
      failureCount += 1;
      setAlertRules([]);
    }

    if (failureCount === 0) {
      setLoadState("ready");
      setLoadMessage("Governance APIs connected.");
      return;
    }

    setLoadState("error");
    setLoadMessage(
      `${failureCount} governance API surface${failureCount === 1 ? "" : "s"} unavailable. Retry once the services are reachable.`,
    );
  };

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    return () => controller.abort();
  }, []);

  const onFieldChange = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = event.target;
    setUserForm((current) => ({ ...current, [name]: value }));
  };

  const createUser = async () => {
    const validationMessage = validateUserForm(userForm);
    if (validationMessage) {
      setActionResult({ severity: "warning", message: validationMessage });
      return;
    }

    setCreateBusy(true);
    setActionResult(undefined);
    try {
      const response = await fetch(`${AUTH_API_BASE}/auth/agency-users`, {
        method: "POST",
        headers: adminHeaders(),
        body: JSON.stringify(userForm),
      });
      handleUnauthorized(response);
      if (!response.ok) {
        throw new Error(`auth API returned ${response.status}`);
      }

      const payload = (await response.json()) as CreateAgencyUserResponse;
      setUsers((current) => [managedUserFromCreateResponse(payload), ...current]);
      setUserForm(defaultUserForm);
      // The temporary password is only returned here, once; hold it for the
      // success dialog and drop it when the dialog closes.
      setCreatedCredentials({
        userId: payload.user.id,
        name: payload.user.name,
        email: payload.user.email,
        temporaryPassword: payload.temporaryPassword,
      });
      setActionResult({
        severity: "success",
        message:
          "Authority user created. Hand the temporary password to the user — it is shown once and required for MFA setup.",
      });
    } catch {
      setActionResult({
        severity: "error",
        message:
          "User was not created. The auth API is unavailable or rejected the current admin session.",
      });
    } finally {
      setCreateBusy(false);
    }
  };

  return {
    agencies,
    users,
    auditLogs,
    dataSources,
    alertRules,
    loadState,
    loadMessage,
    refresh: () => void refresh(),
    userForm,
    createBusy,
    actionResult,
    createdCredentials,
    dismissCreatedCredentials: () => setCreatedCredentials(null),
    onFieldChange,
    createUser,
  };
}
