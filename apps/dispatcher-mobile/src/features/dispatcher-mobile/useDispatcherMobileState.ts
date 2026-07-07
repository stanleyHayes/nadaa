import { useEffect, useMemo, useState } from "react";
import type {
  HospitalCapacityRecord,
  IncidentRecord,
  IncidentStatus,
} from "@nadaa/shared-types";
import {
  assignmentAgencyOptions,
  defaultCapacityFilters,
  defaultFilters,
  fixtureDispatcherSession,
  initialAssignmentForm,
  initialPermissions,
  initialStatusForm,
  initialTimelineNoteForm,
  statusLabel,
} from "./data";
import {
  agencyLogin,
  assignIncident,
  fetchHospitalCapacity,
  fetchIncidentQueue,
  registerPushToken,
  updateIncidentStatus,
} from "./api";
import {
  createMemoryStorage,
  readCapacityCache,
  readIncidentCache,
  readSelectedIncidentId,
  readSession,
  writeCapacityCache,
  writeIncidentCache,
  writeSelectedIncidentId,
  writeSession,
} from "./offline";
import { nextPermissionStatus, permissionMessage } from "./permissions";
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

const storage = createMemoryStorage();

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
  const [session, setSession] = useState<DispatcherSession>(
    fixtureDispatcherSession,
  );
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

  useEffect(() => {
    void hydrate();
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
    setSelectedIncidentId(savedIncidentId);
    setCapacity(capacityCache.facilities as HospitalCapacityRecord[]);
    setLoadState({
      status: "success",
      message: `Queue loaded: ${(incidentCache.incidents as IncidentRecord[]).length} incidents.`,
    });
  }

  async function refreshQueue() {
    setLoadState({ status: "loading", message: "Refreshing incident queue" });
    try {
      const nextIncidents = await fetchIncidentQueue(session);
      setIncidents(nextIncidents);
      await writeIncidentCache(storage, {
        cachedAt: new Date().toISOString(),
        incidents: nextIncidents,
      });
      setLoadState({
        status: "success",
        message: `${nextIncidents.length} incidents refreshed.`,
      });
    } catch (error) {
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
    setLoadState({ status: "loading", message: "Loading nearby capacity" });
    try {
      const response = await fetchHospitalCapacity(
        session,
        incident.location.lat,
        incident.location.lng,
      );
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
        agencyId: response.user.agency?.id ?? fixtureDispatcherSession.agencyId,
        agencyName:
          response.user.agency?.name ?? fixtureDispatcherSession.agencyName,
        mfaCompleted: true,
        role: response.user.role,
        userId: response.user.id,
        userName: response.user.name,
      };
      setSession(nextSession);
      await writeSession(storage, nextSession);
      setAuthForm({ email: "", mfaCode: "", password: "" });
      setLoadState({
        status: "success",
        message: `Signed in as ${nextSession.userName}.`,
      });
      await refreshQueue();
    } catch (error) {
      setLoadState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Sign in failed. Dev fixture session is active.",
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

  async function togglePermission(key: keyof DispatcherPermissionState) {
    const nextStatus = nextPermissionStatus(permissions[key]);
    const nextPermissions = { ...permissions, [key]: nextStatus };
    setPermissions(nextPermissions);
    setLoadState({
      status: nextStatus === "granted" ? "success" : "idle",
      message: permissionMessage(key, nextStatus),
    });
    if (key === "push") {
      setPushState(await registerPushToken(nextStatus === "granted"));
    }
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
