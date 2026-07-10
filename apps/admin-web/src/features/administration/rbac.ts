import type { AgencyUserRole } from "@nadaa/shared-types";

export const adminRoles: AgencyUserRole[] = ["system_admin", "agency_admin"];

export function hasAdminAccess(role: AgencyUserRole, mfaEnabled: boolean) {
  return adminRoles.includes(role) && mfaEnabled;
}
