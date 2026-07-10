import { type ChangeEvent, useEffect, useState } from "react";
import type { CreateAgencyUserResponse } from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";
import { adminHeaders } from "@/app/session";
import { fetchAlertRules, fetchAuditLogs, fetchDataSources } from "./api";
import {
  defaultUserForm,
  fallbackAgencies,
  fallbackAlertRules,
  fallbackAuditLogs,
  fallbackDataSources,
  fallbackUsers,
} from "./data";
import type {
  AdminActionResult,
  AdminLoadState,
  AdminUserFormState,
  AlertRuleSummary,
  DataSourceSummary,
  ManagedAgency,
  ManagedAgencyUser,
} from "./types";
import { managedUserFromCreateResponse, validateUserForm } from "./utils";

export type AdminData = ReturnType<typeof useAdminData>;

/**
 * Central governance-console state container. Loads every agency, user, audit,
 * integration, and alert-rule surface the admin views depend on so the shell
 * can mount data once and route between views without losing state. Preserves
 * the fixture-fallback behaviour: any API surface that is unavailable falls
 * back to safe admin fixtures instead of blanking the console.
 */
export function useAdminData() {
  const [loadState, setLoadState] = useState<AdminLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading governance data");
  const [agencies] = useState<ManagedAgency[]>(fallbackAgencies);
  const [users, setUsers] = useState<ManagedAgencyUser[]>(fallbackUsers);
  const [auditLogs, setAuditLogs] = useState(fallbackAuditLogs);
  const [dataSources, setDataSources] =
    useState<DataSourceSummary[]>(fallbackDataSources);
  const [alertRules, setAlertRules] =
    useState<AlertRuleSummary[]>(fallbackAlertRules);
  const [userForm, setUserForm] = useState<AdminUserFormState>(defaultUserForm);
  const [createBusy, setCreateBusy] = useState(false);
  const [actionResult, setActionResult] = useState<AdminActionResult>();

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading governance data");

    const [auditResult, sourceResult, alertResult] = await Promise.allSettled([
      fetchAuditLogs(signal),
      fetchDataSources(signal),
      fetchAlertRules(signal),
    ]);

    if (signal?.aborted) {
      return;
    }

    let fallbackCount = 0;
    if (auditResult.status === "fulfilled") {
      setAuditLogs(auditResult.value);
    } else {
      fallbackCount += 1;
      setAuditLogs(fallbackAuditLogs);
    }

    if (sourceResult.status === "fulfilled") {
      setDataSources(sourceResult.value);
    } else {
      fallbackCount += 1;
      setDataSources(fallbackDataSources);
    }

    if (alertResult.status === "fulfilled") {
      setAlertRules(alertResult.value);
    } else {
      fallbackCount += 1;
      setAlertRules(fallbackAlertRules);
    }

    if (fallbackCount === 0) {
      setLoadState("ready");
      setLoadMessage("Governance APIs connected.");
      return;
    }

    setLoadState("fallback");
    setLoadMessage(
      `${fallbackCount} governance API surface${fallbackCount === 1 ? "" : "s"} unavailable. Showing safe admin fixture data.`,
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
      if (!response.ok) {
        throw new Error(`auth API returned ${response.status}`);
      }

      const payload = (await response.json()) as CreateAgencyUserResponse;
      setUsers((current) => [managedUserFromCreateResponse(payload), ...current]);
      setUserForm(defaultUserForm);
      setActionResult({
        severity: "success",
        message:
          "Authority user created. MFA setup is required before the user can sign in.",
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
    onFieldChange,
    createUser,
  };
}
