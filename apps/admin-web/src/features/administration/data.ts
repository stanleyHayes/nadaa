import type { AdminUserFormState } from "./types";

/**
 * Static defaults for the create-user form. Field definitions only — no
 * fabricated backend records. The operator picks an agency from the live
 * directory before submitting.
 */
export const defaultUserForm: AdminUserFormState = {
  name: "",
  email: "",
  phone: "+233",
  agencyId: "",
  role: "dispatcher",
};
