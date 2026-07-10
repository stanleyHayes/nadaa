import type { AgencyUserRole } from "@nadaa/shared-types";

export const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

export const authoritySession = {
  id: "usr_nadmo_accra",
  name: "NADMO Officer",
  role: "nadmo_officer" as AgencyUserRole,
  agencyId: "00000000-0000-0000-0000-000000000101",
  agency: "NADMO Accra Metro",
  district: "Accra Metropolitan",
  mfaEnabled: true,
};

export function authorityHeaders() {
  return {
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": authoritySession.id,
    "X-NADAA-Actor-Role": authoritySession.role,
    "X-NADAA-Agency-ID": authoritySession.agencyId,
    "X-NADAA-Actor-District": authoritySession.district,
    "X-NADAA-MFA-Completed": authoritySession.mfaEnabled ? "true" : "false",
    "X-NADAA-Request-ID": `authority-ui-${Date.now()}`,
  };
}
