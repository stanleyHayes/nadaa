import type { useDispatcherMobileState } from "../useDispatcherMobileState";

export type DispatcherMobileController = ReturnType<
  typeof useDispatcherMobileState
>;
export type DispatcherMobileState = DispatcherMobileController["state"];
export type DispatcherMobileActions = DispatcherMobileController["actions"];

export type DispatcherScreenProps = {
  actions: DispatcherMobileActions;
  state: DispatcherMobileState;
};
