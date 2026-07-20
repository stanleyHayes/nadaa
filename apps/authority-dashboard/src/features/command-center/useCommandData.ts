import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import type { SelectChangeEvent } from "@mui/material/Select";
import type {
  AssignIncidentRequest,
  AlertListResponse,
  AuthorityAlertRecord,
  CreateAlertRequest,
  CreateReliefPointRequest,
  DuplicateReviewCandidate,
  DuplicateReviewResponse,
  ImageryGeoJSONFeatureCollection,
  IncidentAbuseReviewRequest,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatusUpdateRequest,
  MergeIncidentsRequest,
  MergeIncidentsResponse,
  ReliefPointListResponse,
  ReliefPointRecord,
  ReliefPointStockHistoryResponse,
  ReliefStockCategory,
  ShelterListResponse,
  ShelterOccupancyUpdateRequest,
  ShelterRecord,
  ShelterUpdateResponse,
  UpdateReliefPointRequest,
} from "@nadaa/shared-types";
import {
  ALERT_API_BASE,
  INCIDENT_API_BASE,
  SHELTER_API_BASE,
} from "@/app/config";
import { authorityHeaders, useAuthoritySession } from "@/app/session";
import { playCommandAlarm } from "@/app/alarm";
import { defaultFilters, assignmentAgencyOptions } from "./data";
import type {
  AbuseReviewFormState,
  AlertFormState,
  AlertLoadState,
  AssignmentFormState,
  CommandIncident,
  FilterState,
  IncidentLoadState,
  IncidentStatusFormState,
  ReliefPointFormState,
  ShelterFormState,
} from "./types";
import {
  alertStatusLabel,
  abuseDecisionLabel,
  assignmentForIncident,
  buildAlertTarget,
  buildDefaultAbuseReviewForm,
  buildDefaultAlertForm,
  buildDefaultAssignmentForm,
  buildDefaultReliefPointForm,
  buildDefaultShelterForm,
  buildDefaultStatusForm,
  buildFilterOptions,
  buildQueueMetrics,
  duplicateReviewCandidatesFor,
  enrichIncidentFromAPI,
  matchesFilters,
  parseReliefStockCategories,
  statusLabel,
} from "./utils";

/**
 * Read a service error body (`{message}` or `{error:{message}}`) so mutation
 * failures surface what the API actually said instead of a generic
 * "API not running" guess.
 */
async function extractError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as {
      message?: string;
      error?: string | { message?: string };
    };
    if (body.message) return body.message;
    if (typeof body.error === "string") return body.error;
    if (body.error?.message) return body.error.message;
    return `Request failed (${response.status})`;
  } catch {
    return `Request failed (${response.status})`;
  }
}

/**
 * Central command-center state container. Holds every incident, alert,
 * shelter, and relief workflow that the dashboard views depend on so the
 * app shell can mount data once and route between views without losing state.
 */
export function useCommandData() {
  const { session } = useAuthoritySession();
  // Destructive delete actions are limited to admin roles. The backend also
  // enforces this (DELETE returns 403 otherwise); this is the matching UI gate.
  const canDelete = Boolean(
    session && ["system_admin", "agency_admin"].includes(session.role),
  );
  // Shelter/relief capacity updates are limited to these roles server-side
  // (shelter-service ShelterUpdateRoles -> 403 otherwise). Gate the edit/create
  // UI to match, so a read-only role sees a view instead of a failing action.
  const canManage = Boolean(
    session &&
      [
        "system_admin",
        "agency_admin",
        "nadmo_officer",
        "district_officer",
        "dispatcher",
      ].includes(session.role),
  );

  const [incidents, setIncidents] = useState<CommandIncident[]>([]);
  const [loadState, setLoadState] = useState<IncidentLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading incident feed");
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [selectedIncidentId, setSelectedIncidentId] = useState("");
  const [statusBusy, setStatusBusy] = useState(false);
  const [statusFeedback, setStatusFeedback] = useState("");
  const [statusForm, setStatusForm] = useState<IncidentStatusFormState>(
    buildDefaultStatusForm(),
  );
  const [abuseBusy, setAbuseBusy] = useState(false);
  const [abuseFeedback, setAbuseFeedback] = useState("");
  const [abuseForm, setAbuseForm] = useState<AbuseReviewFormState>(
    buildDefaultAbuseReviewForm(),
  );
  const [assignmentBusy, setAssignmentBusy] = useState(false);
  const [assignmentFeedback, setAssignmentFeedback] = useState("");
  const [assignmentForm, setAssignmentForm] = useState<AssignmentFormState>(
    buildDefaultAssignmentForm(),
  );
  const [duplicateReviewCandidates, setDuplicateReviewCandidates] = useState<
    DuplicateReviewCandidate[]
  >([]);
  const [selectedDuplicateIds, setSelectedDuplicateIds] = useState<string[]>(
    [],
  );
  const [mergeBusy, setMergeBusy] = useState(false);
  const [mergeFeedback, setMergeFeedback] = useState("");
  const [alerts, setAlerts] = useState<AuthorityAlertRecord[]>([]);
  const [alertLoadState, setAlertLoadState] =
    useState<AlertLoadState>("loading");
  const [alertMessage, setAlertMessage] = useState("Loading alert workflow");
  const [alertBusy, setAlertBusy] = useState(false);
  const [alertFeedback, setAlertFeedback] = useState("");
  const [alertForm, setAlertForm] = useState<AlertFormState>(
    buildDefaultAlertForm(),
  );
  const [shelters, setShelters] = useState<ShelterRecord[]>([]);
  const [shelterLoadState, setShelterLoadState] =
    useState<IncidentLoadState>("loading");
  const [shelterFeedback, setShelterFeedback] = useState(
    "Loading shelter capacity",
  );
  const [shelterBusy, setShelterBusy] = useState(false);
  const [shelterForm, setShelterForm] = useState<ShelterFormState>(
    buildDefaultShelterForm(),
  );
  const [reliefPoints, setReliefPoints] = useState<ReliefPointRecord[]>([]);
  const [reliefLoadState, setReliefLoadState] =
    useState<IncidentLoadState>("loading");
  const [reliefFeedback, setReliefFeedback] = useState("Loading relief points");
  const [reliefBusy, setReliefBusy] = useState(false);
  const [reliefForm, setReliefForm] = useState<ReliefPointFormState>(
    buildDefaultReliefPointForm(),
  );
  // Mirror the latest relief form so async refreshes resolve the operator's
  // current selection instead of a stale closure value.
  const reliefFormRef = useRef(reliefForm);
  reliefFormRef.current = reliefForm;
  const [reliefHistory, setReliefHistory] = useState<
    ReliefPointStockHistoryResponse["history"]
  >([]);
  const [showImageryOverlay, setShowImageryOverlay] = useState(false);
  const [imageryFeatures, setImageryFeatures] = useState<
    ImageryGeoJSONFeatureCollection | undefined
  >(undefined);

  const refreshIncidents = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading incident feed");

    try {
      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
        headers: authorityHeaders(),
        signal,
      });
      if (!response.ok) {
        throw new Error(`incident API returned ${response.status}`);
      }

      const payload = (await response.json()) as IncidentListResponse;
      if (!payload.incidents.length) {
        setIncidents([]);
        setSelectedIncidentId("");
        setLoadState("empty");
        setLoadMessage("No incidents are currently in the command queue.");
        return;
      }

      const nextIncidents = payload.incidents.map(enrichIncidentFromAPI);
      setIncidents(nextIncidents);
      setSelectedIncidentId(nextIncidents[0]?.id ?? "");
      setLoadState("ready");
      setLoadMessage("Live incident feed connected.");
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setIncidents([]);
      setSelectedIncidentId("");
      setLoadState("error");
      setLoadMessage(
        "Incident feed unavailable. Reconnect the incident-service to load the queue.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshIncidents(controller.signal);
    return () => controller.abort();
  }, []);

  const refreshAlerts = async (signal?: AbortSignal) => {
    setAlertLoadState("loading");
    setAlertMessage("Loading alert workflow");

    try {
      const response = await fetch(`${ALERT_API_BASE}/alerts`, {
        headers: authorityHeaders(),
        signal,
      });
      if (!response.ok) {
        throw new Error(`alert API returned ${response.status}`);
      }

      const payload = (await response.json()) as AlertListResponse;
      setAlerts(payload.alerts.length ? payload.alerts : []);
      setAlertLoadState("ready");
      setAlertMessage("Alert workflow API connected.");
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setAlerts([]);
      setAlertLoadState("error");
      setAlertMessage(
        "Alert workflow unavailable. Reconnect the alert-service to load approvals.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshAlerts(controller.signal);
    return () => controller.abort();
  }, []);

  const refreshShelters = async (signal?: AbortSignal) => {
    setShelterLoadState("loading");
    setShelterFeedback("Loading shelter capacity");

    try {
      const response = await fetch(`${SHELTER_API_BASE}/shelters`, {
        signal,
      });
      if (!response.ok) {
        throw new Error(`shelter API returned ${response.status}`);
      }

      const payload = (await response.json()) as ShelterListResponse;
      const nextShelters = payload.shelters;
      setShelters(nextShelters);
      setShelterForm((current) => {
        const selected =
          nextShelters.find((shelter) => shelter.id === current.shelterId) ??
          nextShelters[0];
        return buildDefaultShelterForm(selected);
      });
      setShelterLoadState(nextShelters.length ? "ready" : "empty");
      setShelterFeedback(
        nextShelters.length
          ? "Shelter capacity API connected."
          : "No shelters are currently registered.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setShelters([]);
      setShelterForm(buildDefaultShelterForm());
      setShelterLoadState("error");
      setShelterFeedback(
        "Shelter capacity unavailable. Reconnect the shelter-service.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshShelters(controller.signal);
    return () => controller.abort();
  }, []);

  const refreshReliefHistory = async (
    reliefPointId?: string,
    signal?: AbortSignal,
  ) => {
    if (!reliefPointId || reliefPointId === "__new__") {
      setReliefHistory([]);
      return;
    }

    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/relief-points/${reliefPointId}/stock-history`,
        { signal },
      );
      if (!response.ok) {
        throw new Error(`relief history API returned ${response.status}`);
      }
      const payload =
        (await response.json()) as ReliefPointStockHistoryResponse;
      setReliefHistory(payload.history);
    } catch (error) {
      if (!signal?.aborted) {
        setReliefHistory([]);
      }
    }
  };

  const refreshReliefPoints = async (signal?: AbortSignal) => {
    setReliefLoadState("loading");
    setReliefFeedback("Loading relief distribution points");

    try {
      const response = await fetch(`${SHELTER_API_BASE}/relief-points`, {
        signal,
      });
      if (!response.ok) {
        throw new Error(`relief point API returned ${response.status}`);
      }

      const payload = (await response.json()) as ReliefPointListResponse;
      const nextReliefPoints = payload.reliefPoints;
      setReliefPoints(nextReliefPoints);
      // Keep the operator's current point selected when it still exists and
      // load the stock history for that same point — not blindly the first.
      const selected =
        nextReliefPoints.find(
          (point) => point.id === reliefFormRef.current.reliefPointId,
        ) ?? nextReliefPoints[0];
      setReliefForm(buildDefaultReliefPointForm(selected));
      setReliefLoadState(nextReliefPoints.length ? "ready" : "empty");
      setReliefFeedback(
        nextReliefPoints.length
          ? "Relief point API connected."
          : "No relief distribution points are currently published.",
      );
      void refreshReliefHistory(selected?.id, signal);
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setReliefPoints([]);
      setReliefForm(buildDefaultReliefPointForm());
      setReliefHistory([]);
      setReliefLoadState("error");
      setReliefFeedback(
        "Relief points unavailable. Reconnect the shelter-service.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshReliefPoints(controller.signal);
    return () => controller.abort();
  }, []);

  const filteredIncidents = useMemo(
    () => incidents.filter((incident) => matchesFilters(incident, filters)),
    [filters, incidents],
  );

  const selectedIncident = useMemo(() => {
    if (!filteredIncidents.length) {
      return undefined;
    }
    return (
      filteredIncidents.find(
        (incident) => incident.id === selectedIncidentId,
      ) ?? filteredIncidents[0]
    );
  }, [filteredIncidents, selectedIncidentId]);

  const selectedShelter = useMemo(
    () =>
      shelters.find((shelter) => shelter.id === shelterForm.shelterId) ??
      shelters[0],
    [shelterForm.shelterId, shelters],
  );

  const selectedReliefPoint = useMemo(
    () =>
      reliefPoints.find((point) => point.id === reliefForm.reliefPointId) ??
      reliefPoints[0],
    [reliefForm.reliefPointId, reliefPoints],
  );

  useEffect(() => {
    if (!filteredIncidents.length) {
      setSelectedIncidentId("");
      return;
    }

    if (
      !filteredIncidents.some((incident) => incident.id === selectedIncidentId)
    ) {
      setSelectedIncidentId(filteredIncidents[0].id);
    }
  }, [filteredIncidents, selectedIncidentId]);

  const metrics = useMemo(() => buildQueueMetrics(incidents), [incidents]);
  const filterOptions = useMemo(
    () => buildFilterOptions(incidents),
    [incidents],
  );

  useEffect(() => {
    setAlertForm(buildDefaultAlertForm(selectedIncident));
    setStatusForm(buildDefaultStatusForm(selectedIncident));
    setAbuseForm(buildDefaultAbuseReviewForm(selectedIncident));
    setAssignmentForm(buildDefaultAssignmentForm(selectedIncident));
    const localCandidates = duplicateReviewCandidatesFor(
      selectedIncident,
      incidents,
    );
    setDuplicateReviewCandidates(localCandidates);
    setSelectedDuplicateIds(
      localCandidates.map((candidate) => candidate.incident.id),
    );
    setStatusFeedback("");
    setAbuseFeedback("");
    setAssignmentFeedback("");
    setMergeFeedback("");

    if (
      !selectedIncident ||
      selectedIncident.source !== "api" ||
      !selectedIncident.duplicateCandidates.length
    ) {
      return;
    }

    const controller = new AbortController();
    void refreshDuplicateReview(selectedIncident.id, controller.signal);
    return () => controller.abort();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIncident?.id]);

  const updateFilter =
    (key: keyof FilterState) => (event: SelectChangeEvent) => {
      setFilters((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateAlertForm =
    (key: keyof AlertFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      const value =
        "checked" in event.target && typeof event.target.checked === "boolean"
          ? event.target.checked
          : event.target.value;
      setAlertForm((current) => ({ ...current, [key]: value }));
    };

  const updateStatusForm =
    (key: keyof IncidentStatusFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      setStatusForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateAbuseForm =
    (key: keyof AbuseReviewFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      setAbuseForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateAssignmentForm =
    (key: keyof AssignmentFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setAssignmentForm((current) => {
        if (key === "agencyId") {
          const agency = assignmentAgencyOptions.find(
            (item) => item.id === value,
          );
          if (!agency) {
            return current;
          }
          return {
            ...current,
            agencyId: agency.id,
            agencyName: agency.name,
            agencyType: agency.type,
            responderLead: agency.responderLead,
          };
        }
        return { ...current, [key]: value };
      });
    };

  const updateShelterForm =
    (key: keyof ShelterFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setShelterForm((current) => ({ ...current, [key]: value }));
    };

  // Prime the shelter form with an existing record so the edit dialog opens
  // pre-filled. Shelters have no create endpoint, so there is no draft helper.
  const editShelter = (shelter: ShelterRecord) => {
    setShelterFeedback("");
    setShelterForm(buildDefaultShelterForm(shelter));
  };

  const updateReliefForm =
    (key: keyof ReliefPointFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setReliefForm((current) => ({ ...current, [key]: value }));
    };

  // Reset the relief form to an empty draft (`reliefPointId === "__new__"`)
  // so `saveReliefPoint` performs a POST for the add dialog.
  const startReliefPointDraft = () => {
    setReliefFeedback("");
    setReliefHistory([]);
    setReliefForm(buildDefaultReliefPointForm());
  };

  // Prime the relief form + stock history for an existing point so the edit
  // dialog opens pre-filled and `saveReliefPoint` performs a PATCH.
  const editReliefPoint = (point: ReliefPointRecord) => {
    setReliefFeedback("");
    setReliefForm(buildDefaultReliefPointForm(point));
    void refreshReliefHistory(point.id);
  };

  const refreshDuplicateReview = async (
    incidentId: string,
    signal?: AbortSignal,
  ) => {
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${incidentId}/duplicates`,
        {
          headers: authorityHeaders(),
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`incident API returned ${response.status}`);
      }
      const payload = (await response.json()) as DuplicateReviewResponse;
      setDuplicateReviewCandidates(payload.candidates);
      setSelectedDuplicateIds(
        payload.candidates.map((candidate) => candidate.incident.id),
      );
    } catch (error) {
      if (!signal?.aborted) {
        setMergeFeedback("Duplicate review details need a live incident API.");
      }
    }
  };

  const applyIncidentUpdates = (updates: IncidentRecord[]) => {
    const enrichedUpdates = updates.map(enrichIncidentFromAPI);
    const updateIds = new Set(enrichedUpdates.map((incident) => incident.id));
    setIncidents((current) => {
      const remaining = current.filter((item) => !updateIds.has(item.id));
      return [...enrichedUpdates, ...remaining].sort(
        (a, b) =>
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
      );
    });
    if (enrichedUpdates[0]) {
      setSelectedIncidentId(enrichedUpdates[0].id);
      setStatusForm(buildDefaultStatusForm(enrichedUpdates[0]));
      setAbuseForm(buildDefaultAbuseReviewForm(enrichedUpdates[0]));
      setAssignmentForm(buildDefaultAssignmentForm(enrichedUpdates[0]));
    }
    setLoadState("ready");
  };

  const applyIncidentUpdate = (incident: IncidentRecord) => {
    applyIncidentUpdates([incident]);
  };

  const verifySelectedIncident = async () => {
    if (!selectedIncident) {
      return;
    }

    setStatusBusy(true);
    setStatusFeedback("");
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${selectedIncident.id}/verify`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify({ note: statusForm.note }),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setStatusFeedback(`${statusLabel(incident.status)} status saved.`);
    } catch (error) {
      setStatusFeedback(
        error instanceof Error
          ? error.message
          : "Incident workflow action needs the incident-service API running with this incident.",
      );
    } finally {
      setStatusBusy(false);
    }
  };

  const updateIncidentStatus = async () => {
    if (!selectedIncident) {
      return;
    }

    const request: IncidentStatusUpdateRequest = {
      status: statusForm.status,
      note: statusForm.note,
      resolutionNotes: statusForm.resolutionNotes,
    };

    setStatusBusy(true);
    setStatusFeedback("");
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${selectedIncident.id}/status`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setStatusFeedback(`${statusLabel(incident.status)} status saved.`);
    } catch (error) {
      setStatusFeedback(
        error instanceof Error
          ? error.message
          : "Incident workflow action needs a valid live incident transition.",
      );
    } finally {
      setStatusBusy(false);
    }
  };

  const reviewSelectedIncidentAbuse = async () => {
    if (!selectedIncident) {
      return;
    }

    const request: IncidentAbuseReviewRequest = {
      decision: abuseForm.decision,
      note: abuseForm.note,
      resolutionNotes: abuseForm.resolutionNotes,
    };

    setAbuseBusy(true);
    setAbuseFeedback("");
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${selectedIncident.id}/abuse-review`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setAbuseFeedback(
        `${abuseDecisionLabel(request.decision)} review saved for ${incident.reference}.`,
      );
    } catch (error) {
      setAbuseFeedback(
        error instanceof Error
          ? error.message
          : "Report safety review needs a live incident-service API and valid authority transition.",
      );
    } finally {
      setAbuseBusy(false);
    }
  };

  const assignSelectedIncident = async () => {
    if (!selectedIncident) {
      return;
    }

    const request: AssignIncidentRequest = {
      agencyId: assignmentForm.agencyId,
      agencyName: assignmentForm.agencyName,
      agencyType: assignmentForm.agencyType,
      priority: assignmentForm.priority,
      instructions: assignmentForm.instructions.trim(),
      responderLead: assignmentForm.responderLead.trim() || undefined,
    };

    setAssignmentBusy(true);
    setAssignmentFeedback("");
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${selectedIncident.id}/assignments`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setAssignmentFeedback(`Assigned to ${assignmentForIncident(incident)}.`);
    } catch (error) {
      setAssignmentFeedback(
        error instanceof Error
          ? error.message
          : "Assignment needs a verified live incident and incident-service API.",
      );
    } finally {
      setAssignmentBusy(false);
    }
  };

  const toggleDuplicateSelection = (incidentId: string) => {
    setSelectedDuplicateIds((current) =>
      current.includes(incidentId)
        ? current.filter((id) => id !== incidentId)
        : [...current, incidentId],
    );
  };

  const mergeSelectedDuplicates = async () => {
    if (!selectedIncident || !selectedDuplicateIds.length) {
      return;
    }

    const request: MergeIncidentsRequest = {
      duplicateIncidentIds: selectedDuplicateIds,
      note: `Merged duplicate reports into ${selectedIncident.reference}.`,
    };

    setMergeBusy(true);
    setMergeFeedback("");
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${selectedIncident.id}/merge`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const payload = (await response.json()) as MergeIncidentsResponse;
      applyIncidentUpdates([payload.incident, ...payload.mergedIncidents]);
      setDuplicateReviewCandidates([]);
      setSelectedDuplicateIds([]);
      setMergeFeedback(
        `${payload.mergedIncidents.length} duplicate report${
          payload.mergedIncidents.length === 1 ? "" : "s"
        } merged into ${payload.incident.reference}.`,
      );
    } catch (error) {
      setMergeFeedback(
        error instanceof Error
          ? error.message
          : "Merge needs a live duplicate candidate and incident-service API.",
      );
    } finally {
      setMergeBusy(false);
    }
  };

  const buildAlertRequest = (): CreateAlertRequest => ({
    title: alertForm.title,
    hazardType: selectedIncident?.type ?? "flood",
    severity: alertForm.severity,
    message: alertForm.message,
    target: buildAlertTarget(alertForm),
    startsAt: new Date(alertForm.startsAt).toISOString(),
    expiresAt: new Date(alertForm.expiresAt).toISOString(),
    recommendedAction: alertForm.recommendedAction,
    evacuationRequired: alertForm.evacuationRequired,
    shelterIds: alertForm.shelterIds
      .split(",")
      .map((shelterId) => shelterId.trim())
      .filter(Boolean),
  });

  const updateShelterCapacity = async (): Promise<boolean> => {
    if (!selectedShelter) {
      return false;
    }

    const capacity = Number(shelterForm.capacity);
    const currentOccupancy = Number(shelterForm.currentOccupancy);
    if (
      !Number.isFinite(capacity) ||
      !Number.isFinite(currentOccupancy) ||
      capacity < 0 ||
      currentOccupancy < 0 ||
      currentOccupancy > capacity
    ) {
      setShelterFeedback(
        "Capacity and occupancy must be valid numbers, and occupancy cannot exceed capacity.",
      );
      return false;
    }

    const request: ShelterOccupancyUpdateRequest = {
      capacity,
      currentOccupancy,
      status: shelterForm.status,
      notes: shelterForm.notes.trim(),
    };

    setShelterBusy(true);
    setShelterFeedback("");
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/shelters/${selectedShelter.id}/occupancy`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }

      const payload = (await response.json()) as ShelterUpdateResponse;
      setShelters((current) =>
        current.map((shelter) =>
          shelter.id === payload.shelter.id ? payload.shelter : shelter,
        ),
      );
      setShelterForm(buildDefaultShelterForm(payload.shelter));
      setShelterLoadState("ready");
      setShelterFeedback(`${payload.shelter.name} capacity updated.`);
      return true;
    } catch (error) {
      setShelterFeedback(
        error instanceof Error
          ? error.message
          : "Shelter update needs a live shelter-service API and authority session.",
      );
      return false;
    } finally {
      setShelterBusy(false);
    }
  };

  const deleteShelter = async (shelter: ShelterRecord): Promise<boolean> => {
    setShelterBusy(true);
    setShelterFeedback("");
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/shelters/${shelter.id}`,
        {
          method: "DELETE",
          headers: authorityHeaders(),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }

      setShelters((current) =>
        current.filter((item) => item.id !== shelter.id),
      );
      setShelterForm((current) =>
        current.shelterId === shelter.id
          ? buildDefaultShelterForm()
          : current,
      );
      setShelterFeedback(`${shelter.name} removed from the shelter register.`);
      return true;
    } catch (error) {
      setShelterFeedback(
        error instanceof Error
          ? error.message
          : "Shelter delete needs a live shelter-service API and an admin session.",
      );
      return false;
    } finally {
      setShelterBusy(false);
    }
  };

  const saveReliefPoint = async (): Promise<boolean> => {
    const latitude = Number(reliefForm.latitude);
    const longitude = Number(reliefForm.longitude);
    if (
      !reliefForm.name.trim() ||
      !Number.isFinite(latitude) ||
      !Number.isFinite(longitude) ||
      latitude < -90 ||
      latitude > 90 ||
      longitude < -180 ||
      longitude > 180
    ) {
      setReliefFeedback(
        "Relief point name and valid latitude/longitude are required.",
      );
      return false;
    }

    let stockCategories: ReliefStockCategory[];
    try {
      stockCategories = parseReliefStockCategories(reliefForm.stockCategories);
    } catch (error) {
      setReliefFeedback(
        error instanceof Error
          ? error.message
          : "Stock categories must be valid.",
      );
      return false;
    }

    const payload = {
      address: reliefForm.address.trim(),
      contact: reliefForm.contact.trim(),
      district: reliefForm.district.trim(),
      eligibility: reliefForm.eligibility.trim(),
      location: { lat: latitude, lng: longitude },
      name: reliefForm.name.trim(),
      operatingHours: reliefForm.operatingHours.trim(),
      region: reliefForm.region.trim(),
      schedule: reliefForm.schedule.trim(),
      sourceRef: reliefForm.sourceRef.trim() || "authority-dashboard",
      status: reliefForm.status,
      stockCategories,
      type: reliefForm.type,
    };
    const creating = reliefForm.reliefPointId === "__new__";

    setReliefBusy(true);
    setReliefFeedback("");
    try {
      const response = await fetch(
        creating
          ? `${SHELTER_API_BASE}/relief-points`
          : `${SHELTER_API_BASE}/relief-points/${reliefForm.reliefPointId}`,
        {
          method: creating ? "POST" : "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(
            creating
              ? ({
                  ...payload,
                  source: "manual",
                } satisfies CreateReliefPointRequest)
              : (payload satisfies UpdateReliefPointRequest),
          ),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }

      const reliefPoint = (await response.json()) as ReliefPointRecord;
      setReliefPoints((current) => {
        const exists = current.some((item) => item.id === reliefPoint.id);
        return exists
          ? current.map((item) =>
              item.id === reliefPoint.id ? reliefPoint : item,
            )
          : [reliefPoint, ...current];
      });
      setReliefForm(buildDefaultReliefPointForm(reliefPoint));
      setReliefLoadState("ready");
      setReliefFeedback(
        `${reliefPoint.name} ${creating ? "published" : "updated"}.`,
      );
      void refreshReliefHistory(reliefPoint.id);
      return true;
    } catch (error) {
      setReliefFeedback(
        error instanceof Error
          ? error.message
          : "Relief point save needs a live shelter-service API and authority session.",
      );
      return false;
    } finally {
      setReliefBusy(false);
    }
  };

  const deleteReliefPoint = async (
    point: ReliefPointRecord,
  ): Promise<boolean> => {
    setReliefBusy(true);
    setReliefFeedback("");
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/relief-points/${point.id}`,
        {
          method: "DELETE",
          headers: authorityHeaders(),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }

      setReliefPoints((current) =>
        current.filter((item) => item.id !== point.id),
      );
      setReliefForm((current) =>
        current.reliefPointId === point.id
          ? buildDefaultReliefPointForm()
          : current,
      );
      setReliefHistory([]);
      setReliefFeedback(`${point.name} removed from relief distribution.`);
      return true;
    } catch (error) {
      setReliefFeedback(
        error instanceof Error
          ? error.message
          : "Relief point delete needs a live shelter-service API and an admin session.",
      );
      return false;
    } finally {
      setReliefBusy(false);
    }
  };

  const createAlertDraft = async () => {
    const startsAt = new Date(alertForm.startsAt);
    const expiresAt = new Date(alertForm.expiresAt);
    if (
      !alertForm.startsAt ||
      !alertForm.expiresAt ||
      Number.isNaN(startsAt.getTime()) ||
      Number.isNaN(expiresAt.getTime())
    ) {
      setAlertFeedback(
        "Set valid start and expiry date/times before drafting an alert.",
      );
      return;
    }
    setAlertBusy(true);
    setAlertFeedback("");
    try {
      const response = await fetch(`${ALERT_API_BASE}/alerts`, {
        method: "POST",
        headers: authorityHeaders(),
        body: JSON.stringify(buildAlertRequest()),
      });
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const alert = (await response.json()) as AuthorityAlertRecord;
      setAlerts((current) => [
        alert,
        ...current.filter((item) => item.id !== alert.id),
      ]);
      setAlertLoadState("ready");
      setAlertFeedback("Draft created.");
    } catch (error) {
      setAlertFeedback(
        error instanceof Error
          ? error.message
          : "Alert API unavailable. Start alert-service to create drafts.",
      );
    } finally {
      setAlertBusy(false);
    }
  };

  const runAlertAction = async (
    alert: AuthorityAlertRecord,
    action: "submit" | "approve" | "reject" | "emergency-override",
  ) => {
    setAlertBusy(true);
    setAlertFeedback("");
    const body =
      action === "reject"
        ? { reason: "Rejected from authority dashboard review queue." }
        : action === "emergency-override"
          ? { reason: "Emergency override from authority dashboard." }
          : { note: `Authority dashboard ${action}.` };

    try {
      const response = await fetch(
        `${ALERT_API_BASE}/alerts/${alert.id}/${action}`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(body),
        },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const updatedAlert = (await response.json()) as AuthorityAlertRecord;
      setAlerts((current) =>
        current.map((item) =>
          item.id === updatedAlert.id ? updatedAlert : item,
        ),
      );
      setAlertLoadState("ready");
      setAlertFeedback(`${alertStatusLabel(updatedAlert.status)} alert saved.`);
      if (updatedAlert.status === "approved") {
        // An approved warning just went out to citizens — sound the alarm.
        playCommandAlarm();
      }
    } catch (error) {
      setAlertFeedback(
        error instanceof Error
          ? error.message
          : "Alert action needs the alert-service API running.",
      );
    } finally {
      setAlertBusy(false);
    }
  };

  return {
    // Session-derived permissions
    canDelete,
    canManage,
    // Incident feed
    incidents,
    filteredIncidents,
    loadState,
    loadMessage,
    metrics,
    filters,
    filterOptions,
    updateFilter,
    selectedIncident,
    selectedIncidentId,
    setSelectedIncidentId,
    refreshIncidents,
    // Incident workflow
    statusBusy,
    statusFeedback,
    statusForm,
    updateStatusForm,
    verifySelectedIncident,
    updateIncidentStatus,
    abuseBusy,
    abuseFeedback,
    abuseForm,
    updateAbuseForm,
    reviewSelectedIncidentAbuse,
    assignmentBusy,
    assignmentFeedback,
    assignmentForm,
    updateAssignmentForm,
    assignSelectedIncident,
    duplicateReviewCandidates,
    selectedDuplicateIds,
    toggleDuplicateSelection,
    mergeBusy,
    mergeFeedback,
    mergeSelectedDuplicates,
    // Alerts
    alerts,
    alertLoadState,
    alertMessage,
    alertBusy,
    alertFeedback,
    alertForm,
    updateAlertForm,
    createAlertDraft,
    runAlertAction,
    // Shelters
    shelters,
    shelterLoadState,
    shelterFeedback,
    shelterBusy,
    shelterForm,
    updateShelterForm,
    selectedShelter,
    refreshShelters,
    updateShelterCapacity,
    editShelter,
    deleteShelter,
    // Relief
    reliefPoints,
    reliefLoadState,
    reliefFeedback,
    reliefBusy,
    reliefForm,
    updateReliefForm,
    selectedReliefPoint,
    reliefHistory,
    refreshReliefPoints,
    saveReliefPoint,
    startReliefPointDraft,
    editReliefPoint,
    deleteReliefPoint,
    // Imagery overlay bridge
    showImageryOverlay,
    setShowImageryOverlay,
    imageryFeatures,
    setImageryFeatures,
  };
}

export type CommandData = ReturnType<typeof useCommandData>;
