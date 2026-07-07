import type { useCitizenMobileState } from "../useCitizenMobileState";

export type CitizenMobileController = ReturnType<typeof useCitizenMobileState>;
export type CitizenMobileState = CitizenMobileController["state"];
export type CitizenMobileActions = CitizenMobileController["actions"];

export type CitizenScreenProps = {
  actions: CitizenMobileActions;
  state: CitizenMobileState;
};
