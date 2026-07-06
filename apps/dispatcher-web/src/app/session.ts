import type { AgencyUserRole } from "@nadaa/shared-types";

export const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
];

export const dispatcherSession = {
  id: "usr_dispatch_accra",
  name: "Accra Dispatcher",
  role: "dispatcher" as AgencyUserRole,
  agencyId: "00000000-0000-0000-0000-000000000101",
  agency: "NADMO Accra Dispatch",
  mfaEnabled: true,
};

export function dispatcherHeaders() {
  return {
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": dispatcherSession.id,
    "X-NADAA-Actor-Role": dispatcherSession.role,
    "X-NADAA-Agency-ID": dispatcherSession.agencyId,
    "X-NADAA-MFA-Completed": dispatcherSession.mfaEnabled ? "true" : "false",
    "X-NADAA-Request-ID": `dispatcher-web-${Date.now()}`,
  };
}
