import { useEffect, useMemo, useRef, useState } from "react";
import { AppState } from "react-native";
import AsyncStorage from "@react-native-async-storage/async-storage";
import type {
  HospitalCapacityRecord,
  IncidentRecord,
  IncidentStatus,
} from "@nadaa/shared-types";
import {
  assignmentAgencyOptions,
  defaultCapacityFilters,
  defaultFilters,
  initialAssignmentForm,
  initialPermissions,
  initialStatusForm,
  initialTimelineNoteForm,
  statusLabel,
} from "./data";
import {
  agencyLogin,
  ApiError,
  assignIncident,
  fetchHospitalCapacity,
  fetchIncidentQueue,
  registerPushToken,
  updateIncidentStatus,
} from "./api";
import {
  clearSession,
  readCapacityCache,
  readIncidentCache,
  readSelectedIncidentId,
  readSession,
  writeCapacityCache,
  writeIncidentCache,
  writeSelectedIncidentId,
  writeSession,
  type KeyValueStorage,
} from "./offline";
import { nextPermissionStatus, permissionMessage } from "./permissions";
import {
  configureIncidentNotifications,
  getIncidentPermission,
  notifyNewIncidents,
  requestIncidentPermission,
} from "./incidentNotifications";
import type {
  AuthFormState,
  AssignmentFormState,
  CapacityFilterState,
  DispatcherPermissionState,
  DispatcherSession,
  IncidentFilterState,
  MobileLoadState,
  PushRegistrationState,
  StatusFormState,
  TimelineNoteFormState,
} from "./types";

// Session, incident queue, and capacity caches persist across cold starts
// (same AsyncStorage pattern as the citizen app).
const storage: KeyValueStorage = AsyncStorage;
const QUEUE_POLL_INTERVAL_MS = 30_000;

export function useDispatcherMobileState() {
  const [capacity, setCapacity] = useState<HospitalCapacityRecord[]>([]);
  const [capacityFilters, setCapacityFilters] = useState<CapacityFilterState>(
    defaultCapacityFilters,
  );
  const [filters, setFilters] = useState<IncidentFilterState>(defaultFilters);
  const [incidents, setIncidents] = useState<IncidentRecord[]>([]);
  const [loadState, setLoadState] = useState<MobileLoadState>({
    status: "idle",
    message: "Dispatcher mobile ready.",
  });
  const [permissions, setPermissions] =
    useState<DispatcherPermissionState>(initialPermissions);
  const [pushState, setPushState] = useState<PushRegistrationState>({
    status: "permission_needed",
    message: "Allow notifications to receive critical incident escalation.",
  });
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(
    null,
  );
  const [session, setSession] = useState<DispatcherSession | null>(null);
  const [authForm, setAuthForm] = useState<AuthFormState>({
    email: "",
    mfaCode: "",
    password: "",
  });
  const [statusForm, setStatusForm] =
    useState<StatusFormState>(initialStatusForm);
  const [assignmentForm, setAssignmentForm] = useState<AssignmentFormState>(
    initialAssignmentForm,
  );
  const [timelineNoteForm, setTimelineNoteForm] =
    useState<TimelineNoteFormState>(initialTimelineNoteForm);

  const seenIncidentIds = useRef<Set<string>>(new Set());
  const incidentsSeeded = useRef(false);
  // Refs mirror the latest state so the foreground poller (a mount-once
  // interval) never reads a stale session, permission, or selection closure.
  const sessionRef = useRef(session);
  const permissionsRef = useRef(permissions);
  const selectedIncidentIdRef = useRef<string | null>(null);

  useEffect(() => {
    sessionRef.current = session;
    permissionsRef.current = permissions;
  });

  useEffect(() => {
    void hydrate().catch(() => undefined);
    void configureIncidentNotifications().catch(() => undefined);
  }, []);

  // Poll the queue while the app is foregrounded so a newly-arrived
  // life-threatening incident notifies without waiting for a manual refresh.
  useEffect(() => {
    const poll = () => {
      // Never fire a request with a dead or missing token — after auth expiry
      // the session is cleared and polling becomes a no-op until re-login.
      if (!sessionRef.current?.accessToken) {
        return;
      }
      void refreshQueue().catch(() => undefined);
    };
    let interval: ReturnType<typeof setInterval> | null = null;
    const startPolling = () => {
      if (interval == null) {
        interval = setInterval(poll, QUEUE_POLL_INTERVAL_MS);
      }
    };
    const stopPolling = () => {
      if (interval != null) {
        clearInterval(interval);
        interval = null;
      }
    };
    const subscription = AppState.addEventListener("change", (nextState) => {
      if (nextState === "active") {
        // The OS permission may have changed in Settings while backgrounded.
        void syncPushPermission().catch(() => undefined);
        poll();
        startPolling();
      } else {
        stopPolling();
      }
    });
    if (AppState.currentState === "active") {
      startPolling();
    }
    return () => {
      stopPolling();
      subscription.remove();
    };
  }, []);

  const selectedIncident = useMemo(
    () =>
      incidents.find((incident) => incident.id === selectedIncidentId) ?? null,
    [incidents, selectedIncidentId],
  );

  const filteredIncidents = useMemo(() => {
    return incidents.filter((incident) => {
      if (filters.hazard !== "all" && incident.type !== filters.hazard) {
        return false;
      }
      if (
        filters.severity !== "all" &&
        incident.severity !== filters.severity
      ) {
        return false;
      }
      if (filters.status !== "all" && incident.status !== filters.status) {
        return false;
      }
      if (filters.time !== "all") {
        const windowMinutes =
          filters.time === "1h" ? 60 : filters.time === "6h" ? 360 : 1440;
        const createdAt = new Date(incident.createdAt).getTime();
        if (Number.isNaN(createdAt)) {
          return true;
        }
        if (Date.now() - createdAt > windowMinutes * 60 * 1000) {
          return false;
        }
      }
      return true;
    });
  }, [incidents, filters]);

  const filteredCapacity = useMemo(() => {
    return capacity.filter((facility) => {
      if (
        capacityFilters.emergencyCapacity !== "all" &&
        facility.emergencyCapacity !== capacityFilters.emergencyCapacity
      ) {
        return false;
      }
      if (
        facility.availableBeds < Number(capacityFilters.minAvailableBeds || 0)
      ) {
        return false;
      }
      if (
        capacityFilters.service !== "all" &&
        !facility.services.includes(capacityFilters.service)
      ) {
        return false;
      }
      if (!capacityFilters.includeStale && facility.stale) {
        return false;
      }
      return true;
    });
  }, [capacity, capacityFilters]);

  const queueMetrics = useMemo(() => {
    const urgent = incidents.filter(
      (incident) =>
        incident.urgency === "life_threatening" || incident.priorityReview,
    ).length;
    const open = incidents.filter(
      (incident) => !["closed", "false_report"].includes(incident.status),
    ).length;
    return { open, total: incidents.length, urgent };
  }, [incidents]);

  async function hydrate() {
    setLoadState({ status: "loading", message: "Loading dispatcher state" });
    const [savedSession, incidentCache, savedIncidentId, capacityCache] =
      await Promise.all([
        readSession(storage),
        readIncidentCache(storage),
        readSelectedIncidentId(storage),
        readCapacityCache(storage),
      ]);
    setSession(savedSession);
    setIncidents(incidentCache.incidents as IncidentRecord[]);
    selectedIncidentIdRef.current = savedIncidentId;
    setSelectedIncidentId(savedIncidentId);
    setCapacity(capacityCache.facilities as HospitalCapacityRecord[]);
    setLoadState(
      savedSession
        ? {
            status: "success",
            message: `Queue loaded: ${(incidentCache.incidents as IncidentRecord[]).length} incidents.`,
          }
        : {
            status: "idle",
            message: "Sign in from the Profile tab to load the live queue.",
          },
    );
  }

  function isAuthError(error: unknown): boolean {
    return (
      error instanceof ApiError &&
      (error.status === 401 || error.status === 403)
    );
  }

  // A 401/403 means the token is dead: drop the stored session so the poller
  // stops firing, and route the dispatcher back to sign-in.
  async function handleAuthExpired() {
    setSession(null);
    sessionRef.current = null;
    await clearSession(storage).catch(() => undefined);
    setLoadState({
      status: "auth_expired",
      message: "Session expired. Sign in again from the Profile tab.",
    });
  }

  async function refreshQueue(sessionOverride?: DispatcherSession) {
    // Callers that just changed the session (login) pass it in; everyone else
    // — including the foreground poller — reads the latest via the ref.
    const activeSession = sessionOverride ?? sessionRef.current;
    if (!activeSession?.accessToken) {
      // Signed out: keep the cached queue and say so instead of firing an
      // unauthenticated request that would fail as "offline".
      setLoadState({
        status: "idle",
        message: "Sign in from the Profile tab to load the live queue.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Refreshing incident queue" });
    try {
      const nextIncidents = await fetchIncidentQueue(activeSession);
      setIncidents(nextIncidents);
      // Notify for newly-arrived active incidents (never on the first load).
      // The OS channel handles DND; a life-threatening incident overrides it.
      if (
        incidentsSeeded.current &&
        permissionsRef.current.push === "granted"
      ) {
        void notifyNewIncidents(nextIncidents, seenIncidentIds.current).catch(
          () => undefined,
        );
      } else {
        nextIncidents.forEach((incident) =>
          seenIncidentIds.current.add(incident.id),
        );
      }
      incidentsSeeded.current = true;
      await writeIncidentCache(storage, {
        cachedAt: new Date().toISOString(),
        incidents: nextIncidents,
      });
      setLoadState({
        status: "success",
        message: `${nextIncidents.length} incidents refreshed.`,
      });
    } catch (error) {
      if (isAuthError(error)) {
        await handleAuthExpired();
        return;
      }
      setLoadState({
        status: "offline",
        message:
          error instanceof Error
            ? error.message
            : "Network unavailable. Showing cached queue.",
      });
    }
  }

  async function selectIncident(id: string | null) {
    selectedIncidentIdRef.current = id;
    setSelectedIncidentId(id);
    await writeSelectedIncidentId(storage, id);
    if (id == null) {
      return;
    }
    const incident = incidents.find((item) => item.id === id);
    if (incident) {
      setStatusForm({ ...initialStatusForm, status: incident.status });
      await refreshCapacityForIncident(incident);
    }
  }

  async function refreshCapacityForIncident(incident: IncidentRecord) {
    const activeSession = sessionRef.current;
    if (!activeSession?.accessToken) {
      setLoadState({
        status: "idle",
        message: "Sign in from the Profile tab to look up live capacity.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Loading nearby capacity" });
    try {
      const response = await fetchHospitalCapacity(
        activeSession,
        incident.location.lat,
        incident.location.lng,
      );
      // A different incident was selected while loading — ignore this
      // response so rapid selection cannot mix up another incident's capacity.
      if (selectedIncidentIdRef.current !== incident.id) {
        return;
      }
      setCapacity(response.facilities);
      await writeCapacityCache(storage, {
        cachedAt: response.generatedAt,
        facilities: response.facilities,
      });
      setLoadState({
        status: "success",
        message: `${response.facilities.length} facilities nearby.`,
      });
    } catch (error) {
      if (selectedIncidentIdRef.current !== incident.id) {
        return;
      }
      if (isAuthError(error)) {
        await handleAuthExpired();
        return;
      }
      setLoadState({
        status: "offline",
        message:
          error instanceof Error
            ? error.message
            : "Capacity lookup failed. Showing cached data.",
      });
    }
  }

  async function submitStatusUpdate() {
    if (!selectedIncident) {
      setLoadState({ status: "error", message: "Select an incident first." });
      return;
    }
    if (!session?.accessToken) {
      setLoadState({ status: "error", message: "Sign in to update incidents." });
      return;
    }
    const { note, resolutionNotes, status } = statusForm;
    setLoadState({
      status: "loading",
      message: `Updating status to ${status}`,
    });
    try {
      const updated = await updateIncidentStatus(session, selectedIncident.id, {
        note,
        resolutionNotes,
        status,
      });
      await replaceIncident(updated);
      setStatusForm(initialStatusForm);
      setLoadState({
        status: "success",
        message: `Status updated to ${statusLabel(status)}.`,
      });
    } catch (error) {
      if (isAuthError(error)) {
        await handleAuthExpired();
        return;
      }
      setLoadState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Status update failed.",
      });
    }
  }

  async function submitAssignment() {
    if (!selectedIncident) {
      setLoadState({ status: "error", message: "Select an incident first." });
      return;
    }
    if (!session?.accessToken) {
      setLoadState({ status: "error", message: "Sign in to assign incidents." });
      return;
    }
    if (!assignmentForm.agencyId) {
      setLoadState({
        status: "error",
        message: "Choose an agency to assign.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Assigning incident" });
    try {
      const updated = await assignIncident(session, selectedIncident.id, {
        agencyId: assignmentForm.agencyId,
        agencyName: assignmentForm.agencyName,
        agencyType: assignmentForm.agencyType,
        instructions: assignmentForm.instructions,
        priority: assignmentForm.priority,
        responderLead: assignmentForm.responderLead,
      });
      await replaceIncident(updated);
      setAssignmentForm(initialAssignmentForm);
      setLoadState({ status: "success", message: "Incident assigned." });
    } catch (error) {
      if (isAuthError(error)) {
        await handleAuthExpired();
        return;
      }
      setLoadState({
        status: "error",
        message: error instanceof Error ? error.message : "Assignment failed.",
      });
    }
  }

  async function submitTimelineNote() {
    if (!selectedIncident) {
      setLoadState({ status: "error", message: "Select an incident first." });
      return;
    }
    if (!session?.accessToken) {
      setLoadState({ status: "error", message: "Sign in to add notes." });
      return;
    }
    if (!timelineNoteForm.note.trim()) {
      setLoadState({ status: "error", message: "Enter a note." });
      return;
    }
    setLoadState({ status: "loading", message: "Adding timeline note" });
    try {
      const updated = await updateIncidentStatus(session, selectedIncident.id, {
        note: timelineNoteForm.note,
        status: selectedIncident.status,
      });
      await replaceIncident(updated);
      setTimelineNoteForm(initialTimelineNoteForm);
      setLoadState({ status: "success", message: "Timeline note added." });
    } catch (error) {
      if (isAuthError(error)) {
        await handleAuthExpired();
        return;
      }
      setLoadState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Timeline note failed.",
      });
    }
  }

  async function replaceIncident(updated: IncidentRecord) {
    const nextIncidents = [
      updated,
      ...incidents.filter((incident) => incident.id !== updated.id),
    ];
    setIncidents(nextIncidents);
    await writeIncidentCache(storage, {
      cachedAt: new Date().toISOString(),
      incidents: nextIncidents,
    });
  }

  async function login() {
    setLoadState({ status: "loading", message: "Signing in" });
    try {
      const response = await agencyLogin({
        email: authForm.email,
        mfaCode: authForm.mfaCode,
        password: authForm.password,
      });
      const nextSession: DispatcherSession = {
        accessToken: response.accessToken,
        agencyId: response.user.agency?.id ?? "",
        agencyName: response.user.agency?.name ?? "",
        mfaCompleted: response.user.mfaEnabled,
        role: response.user.role,
        userId: response.user.id,
        userName: response.user.name,
      };
      setSession(nextSession);
      sessionRef.current = nextSession;
      await writeSession(storage, nextSession);
      setAuthForm({ email: "", mfaCode: "", password: "" });
      setLoadState({
        status: "success",
        message: `Signed in as ${nextSession.userName}.`,
      });
      // Refresh with the NEW session — the state closure still holds the
      // pre-login (signed-out) session at this point.
      await refreshQueue(nextSession);
    } catch (error) {
      setLoadState({
        status: "error",
        message: error instanceof Error ? error.message : "Sign in failed.",
      });
    }
  }

  function updateFilter<Key extends keyof IncidentFilterState>(
    key: Key,
    value: IncidentFilterState[Key],
  ) {
    setFilters((current) => ({ ...current, [key]: value }));
  }

  function updateCapacityFilter<Key extends keyof CapacityFilterState>(
    key: Key,
    value: CapacityFilterState[Key],
  ) {
    setCapacityFilters((current) => ({ ...current, [key]: value }));
  }

  function updateAuthForm(values: Partial<AuthFormState>) {
    setAuthForm((current) => ({ ...current, ...values }));
  }

  function updateStatusForm(values: Partial<StatusFormState>) {
    setStatusForm((current) => ({ ...current, ...values }));
  }

  function updateAssignmentForm(values: Partial<AssignmentFormState>) {
    setAssignmentForm((current) => ({ ...current, ...values }));
  }

  function chooseAssignmentAgency(agencyId: string) {
    const agency = assignmentAgencyOptions.find(
      (option) => option.id === agencyId,
    );
    if (agency) {
      setAssignmentForm((current) => ({
        ...current,
        agencyId: agency.id,
        agencyName: agency.name,
        agencyType: agency.type,
        responderLead: agency.responderLead,
      }));
    }
  }

  function updateTimelineNoteForm(values: Partial<TimelineNoteFormState>) {
    setTimelineNoteForm((current) => ({ ...current, ...values }));
  }

  // The OS notification permission is the source of truth; mirror it into the
  // in-app state so a change made in device Settings is never mislabeled.
  async function syncPushPermission(): Promise<boolean> {
    const granted = await getIncidentPermission();
    setPermissions((current) => {
      // "unknown" means the user has not been asked yet — keep the prompt copy
      // instead of pre-labeling a denial.
      if (!granted && current.push === "unknown") {
        return current;
      }
      const resolved = granted ? "granted" : "denied";
      if (current.push === resolved) {
        return current;
      }
      return { ...current, push: resolved };
    });
    return granted;
  }

  async function togglePermission(key: keyof DispatcherPermissionState) {
    if (key === "push") {
      // Tapping while granted cannot revoke the OS permission from inside the
      // app — re-sync from the OS and say so. Any other tap is an attempt to
      // enable and must ALWAYS consult the OS: a user who re-enabled
      // notifications in OS Settings must not stay "denied" in-app.
      const enabling = permissions.push !== "granted";
      if (enabling) {
        // requestIncidentPermission checks getPermissionsAsync() first and
        // only prompts when the OS has not already granted.
        await requestIncidentPermission();
      }
      const granted = await syncPushPermission();
      setLoadState({
        status: granted ? "success" : "idle",
        message: granted
          ? permissionMessage("push", "granted")
          : enabling
            ? permissionMessage("push", "denied")
            : "Notifications stay on until they are disabled in device Settings.",
      });
      setPushState(await registerPushToken(granted));
      return;
    }
    const nextStatus = nextPermissionStatus(permissions[key]);
    const nextPermissions = { ...permissions, [key]: nextStatus };
    setPermissions(nextPermissions);
    setLoadState({
      status: nextStatus === "granted" ? "success" : "idle",
      message: permissionMessage(key, nextStatus),
    });
  }

  function setStatusForTransition(status: IncidentStatus) {
    setStatusForm((current) => ({ ...current, status }));
  }

  return {
    actions: {
      chooseAssignmentAgency,
      login,
      refreshCapacityForIncident,
      refreshQueue,
      selectIncident,
      setStatusForTransition,
      submitAssignment,
      submitStatusUpdate,
      submitTimelineNote,
      togglePermission,
      updateAssignmentForm,
      updateAuthForm,
      updateCapacityFilter,
      updateFilter,
      updateStatusForm,
      updateTimelineNoteForm,
    },
    state: {
      assignmentForm,
      authForm,
      capacity,
      capacityFilters,
      filteredCapacity,
      filteredIncidents,
      filters,
      incidents,
      loadState,
      permissions,
      pushState,
      queueMetrics,
      selectedIncident,
      selectedIncidentId,
      session,
      statusForm,
      timelineNoteForm,
    },
  };
}
