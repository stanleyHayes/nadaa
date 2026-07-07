import type { AgencyUserRole } from "@nadaa/shared-types";

export const agencyRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

export const agencySession = {
  id: "usr_agency_responder_001",
  name: "NADMO Responder",
  role: "responder" as AgencyUserRole,
  agencyId: "00000000-0000-0000-0000-000000000101",
  agency: "NADMO Accra Metro",
  mfaEnabled: true,
  token: "local-agency-token",
};

export function agencyHeaders() {
  return {
    Authorization: `Bearer ${agencySession.token}`,
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": agencySession.id,
    "X-NADAA-Actor-Role": agencySession.role,
    "X-NADAA-Agency-ID": agencySession.agencyId,
    "X-NADAA-MFA-Completed": agencySession.mfaEnabled ? "true" : "false",
    "X-NADAA-Request-ID": `agency-web-${Date.now()}`,
  };
}
