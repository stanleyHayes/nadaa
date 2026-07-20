import { type ChangeEvent, useEffect, useState } from "react";
import type {
  AuditLogRecord,
  CreateAgencyUserResponse,
} from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";
import { adminHeaders, type AdminSession } from "@/app/session";
import { handleUnauthorized, SessionExpiredError } from "@/app/http";
import {
  fetchAgencies,
  fetchAgencyUsers,
  fetchAlertRules,
  fetchAuditLogs,
  fetchDataSources,
  GovernanceForbiddenError,
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
import {
  managedUserFromCreateResponse,
  managedUserFromDirectoryEntry,
  validateUserForm,
  withAgencyUserMetrics,
} from "./utils";

export type AdminData = ReturnType<typeof useAdminData>;

/** Best-effort read of the auth service's error body for honest feedback. */
async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as {
      error?: { message?: string };
    };
    if (payload.error?.message) {
      return payload.error.message;
    }
  } catch {
    // Non-JSON error body; fall back to the status line.
  }
  return `The auth API returned ${response.status}.`;
}

/**
 * Central governance-console state container. Loads the users directory, the
 * agency directory, and the audit, integration, and alert-rule surfaces the
 * admin views depend on from the live backends so the shell can mount data
 * once and route between views without losing state. Surfaces the admin's
 * role cannot read (the system_admin-only agency directory and audit trail)
 * degrade to scoped "requires system admin" states; a surface that fails
 * outright stays empty and the console surfaces a concise error state instead
 * of substituting fixture data.
 */
export function useAdminData(session: AdminSession) {
  const [loadState, setLoadState] = useState<AdminLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading governance data");
  const [agencies, setAgencies] = useState<ManagedAgency[]>([]);
  const [users, setUsers] = useState<ManagedAgencyUser[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLogRecord[]>([]);
  const [dataSources, setDataSources] = useState<DataSourceSummary[]>([]);
  const [alertRules, setAlertRules] = useState<AlertRuleSummary[]>([]);
  const [agenciesForbidden, setAgenciesForbidden] = useState(false);
  const [auditForbidden, setAuditForbidden] = useState(false);
  const [usersError, setUsersError] = useState<string | null>(null);

  // Agency admins provision within their own agency only, so the form is
  // pinned to it; system admins pick from the live directory.
  const isAgencyAdmin = session.role === "agency_admin";
  const lockedAgency = isAgencyAdmin
    ? { id: session.agencyId, name: session.agency }
    : undefined;
  const initialUserForm = (): AdminUserFormState =>
    isAgencyAdmin
      ? { ...defaultUserForm, agencyId: session.agencyId }
      : { ...defaultUserForm };

  const [userForm, setUserForm] = useState<AdminUserFormState>(initialUserForm);
  const [createBusy, setCreateBusy] = useState(false);
  const [actionResult, setActionResult] = useState<AdminActionResult>();
  const [createdCredentials, setCreatedCredentials] =
    useState<CreatedUserCredentials | null>(null);

  const agencyNameFor = (agencyId: string, directory: ManagedAgency[]) => {
    const match = directory.find((agency) => agency.id === agencyId);
    if (match) {
      return match.name;
    }
    return agencyId === session.agencyId ? session.agency : "Unknown agency";
  };

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading governance data");

    const [agencyResult, userResult, auditResult, sourceResult, alertResult] =
      await Promise.allSettled([
        fetchAgencies(signal),
        fetchAgencyUsers(signal),
        fetchAuditLogs(signal),
        fetchDataSources(signal),
        fetchAlertRules(signal),
      ]);

    if (signal?.aborted) {
      return;
    }

    let failureCount = 0;

    let loadedAgencies: ManagedAgency[] = [];
    if (agencyResult.status === "fulfilled") {
      loadedAgencies = agencyResult.value;
      setAgenciesForbidden(false);
    } else if (agencyResult.reason instanceof GovernanceForbiddenError) {
      setAgenciesForbidden(true);
    } else {
      failureCount += 1;
      setAgenciesForbidden(false);
    }

    let loadedUsers: ManagedAgencyUser[] | null = null;
    if (userResult.status === "fulfilled") {
      loadedUsers = userResult.value.map((entry) =>
        managedUserFromDirectoryEntry(
          entry,
          agencyNameFor(entry.agencyId, loadedAgencies),
        ),
      );
      setUsersError(null);
    } else {
      failureCount += 1;
      setUsersError(
        "The users directory could not be loaded. Retry once the auth service is reachable.",
      );
    }
    setUsers(loadedUsers ?? []);
    setAgencies(withAgencyUserMetrics(loadedAgencies, loadedUsers));

    if (auditResult.status === "fulfilled") {
      setAuditLogs(auditResult.value);
      setAuditForbidden(false);
    } else if (auditResult.reason instanceof GovernanceForbiddenError) {
      setAuditLogs([]);
      setAuditForbidden(true);
    } else {
      failureCount += 1;
      setAuditLogs([]);
      setAuditForbidden(false);
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
        throw new Error(await readErrorMessage(response));
      }

      const payload = (await response.json()) as CreateAgencyUserResponse;
      setUsers((current) => [managedUserFromCreateResponse(payload), ...current]);
      setUserForm(initialUserForm());
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
    } catch (cause) {
      if (cause instanceof SessionExpiredError) {
        // The 401 guard cleared the session; the app returns to sign-in.
        return;
      }
      setActionResult({
        severity: "error",
        message:
          cause instanceof TypeError
            ? "User was not created. The auth API is unavailable."
            : cause instanceof Error
              ? `User was not created. ${cause.message}`
              : "User was not created. The auth API is unavailable.",
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
    agenciesForbidden,
    auditForbidden,
    usersError,
    lockedAgency,
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
