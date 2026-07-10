import type { AgencyUserRole } from "@nadaa/shared-types";

export const adminSession = {
  id: "usr_system_admin",
  name: "NADAA System Admin",
  role: "system_admin" as AgencyUserRole,
  agencyId: "00000000-0000-0000-0000-000000000101",
  agency: "NADMO National Operations",
  mfaEnabled: true,
  token: "local-admin-token",
};

export function adminHeaders() {
  return {
    Authorization: `Bearer ${adminSession.token}`,
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": adminSession.id,
    "X-NADAA-Actor-Role": adminSession.role,
    "X-NADAA-Agency-ID": adminSession.agencyId,
    "X-NADAA-MFA-Completed": adminSession.mfaEnabled ? "true" : "false",
    "X-NADAA-Request-ID": `admin-web-${Date.now()}`,
  };
}
