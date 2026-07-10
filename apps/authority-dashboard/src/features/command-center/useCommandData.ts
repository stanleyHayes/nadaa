import { type ChangeEvent, useEffect, useMemo, useState } from "react";
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
import { authorityHeaders } from "@/app/session";
import {
  defaultFilters,
  fallbackAlerts,
  fallbackIncidents,
  fallbackReliefPoints,
  fallbackShelters,
  assignmentAgencyOptions,
} from "./data";
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
 * Central command-center state container. Holds every incident, alert,
 * shelter, and relief workflow that the dashboard views depend on so the
 * app shell can mount data once and route between views without losing state.
 */
export function useCommandData() {
  const [incidents, setIncidents] =
    useState<CommandIncident[]>(fallbackIncidents);
  const [loadState, setLoadState] = useState<IncidentLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading incident feed");
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [selectedIncidentId, setSelectedIncidentId] = useState(
    fallbackIncidents[0]?.id ?? "",
  );
  const [statusBusy, setStatusBusy] = useState(false);
  const [statusFeedback, setStatusFeedback] = useState("");
  const [statusForm, setStatusForm] = useState<IncidentStatusFormState>(
    buildDefaultStatusForm(fallbackIncidents[0]),
  );
  const [abuseBusy, setAbuseBusy] = useState(false);
  const [abuseFeedback, setAbuseFeedback] = useState("");
  const [abuseForm, setAbuseForm] = useState<AbuseReviewFormState>(
    buildDefaultAbuseReviewForm(fallbackIncidents[0]),
  );
  const [assignmentBusy, setAssignmentBusy] = useState(false);
  const [assignmentFeedback, setAssignmentFeedback] = useState("");
  const [assignmentForm, setAssignmentForm] = useState<AssignmentFormState>(
    buildDefaultAssignmentForm(fallbackIncidents[0]),
  );
  const [duplicateReviewCandidates, setDuplicateReviewCandidates] = useState<
    DuplicateReviewCandidate[]
  >(duplicateReviewCandidatesFor(fallbackIncidents[0], fallbackIncidents));
  const [selectedDuplicateIds, setSelectedDuplicateIds] = useState<string[]>(
    duplicateReviewCandidatesFor(fallbackIncidents[0], fallbackIncidents).map(
      (candidate) => candidate.incident.id,
    ),
  );
  const [mergeBusy, setMergeBusy] = useState(false);
  const [mergeFeedback, setMergeFeedback] = useState("");
  const [alerts, setAlerts] = useState<AuthorityAlertRecord[]>(fallbackAlerts);
  const [alertLoadState, setAlertLoadState] =
    useState<AlertLoadState>("loading");
  const [alertMessage, setAlertMessage] = useState("Loading alert workflow");
  const [alertBusy, setAlertBusy] = useState(false);
  const [alertFeedback, setAlertFeedback] = useState("");
  const [alertForm, setAlertForm] = useState<AlertFormState>(
    buildDefaultAlertForm(fallbackIncidents[0]),
  );
  const [shelters, setShelters] = useState<ShelterRecord[]>(fallbackShelters);
  const [shelterLoadState, setShelterLoadState] =
    useState<IncidentLoadState>("loading");
  const [shelterFeedback, setShelterFeedback] = useState(
    "Loading shelter capacity",
  );
  const [shelterBusy, setShelterBusy] = useState(false);
  const [shelterForm, setShelterForm] = useState<ShelterFormState>(
    buildDefaultShelterForm(fallbackShelters[0]),
  );
  const [reliefPoints, setReliefPoints] =
    useState<ReliefPointRecord[]>(fallbackReliefPoints);
  const [reliefLoadState, setReliefLoadState] =
    useState<IncidentLoadState>("loading");
  const [reliefFeedback, setReliefFeedback] = useState("Loading relief points");
  const [reliefBusy, setReliefBusy] = useState(false);
  const [reliefForm, setReliefForm] = useState<ReliefPointFormState>(
    buildDefaultReliefPointForm(fallbackReliefPoints[0]),
  );
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

      setIncidents(fallbackIncidents);
      setSelectedIncidentId(fallbackIncidents[0]?.id ?? "");
      setLoadState("fallback");
      setLoadMessage("Incident API unavailable. Showing command fixture data.");
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

      setAlerts(fallbackAlerts);
      setAlertLoadState("fallback");
      setAlertMessage("Alert API unavailable. Showing approval fixture data.");
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
      const nextShelters = payload.shelters.length
        ? payload.shelters
        : fallbackShelters;
      setShelters(nextShelters);
      setShelterForm((current) => {
        const selected =
          nextShelters.find((shelter) => shelter.id === current.shelterId) ??
          nextShelters[0];
        return buildDefaultShelterForm(selected);
      });
      setShelterLoadState("ready");
      setShelterFeedback("Shelter capacity API connected.");
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setShelters(fallbackShelters);
      setShelterForm(buildDefaultShelterForm(fallbackShelters[0]));
      setShelterLoadState("fallback");
      setShelterFeedback("Shelter API unavailable. Showing fixture capacity.");
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
      const response = await fetch(
        `${SHELTER_API_BASE}/relief-points?limit=12`,
        {
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`relief point API returned ${response.status}`);
      }

      const payload = (await response.json()) as ReliefPointListResponse;
      const nextReliefPoints = payload.reliefPoints.length
        ? payload.reliefPoints
        : fallbackReliefPoints;
      setReliefPoints(nextReliefPoints);
      setReliefForm((current) => {
        const selected =
          nextReliefPoints.find(
            (point) => point.id === current.reliefPointId,
          ) ?? nextReliefPoints[0];
        return buildDefaultReliefPointForm(selected);
      });
      setReliefLoadState("ready");
      setReliefFeedback("Relief point API connected.");
      void refreshReliefHistory(nextReliefPoints[0]?.id, signal);
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setReliefPoints(fallbackReliefPoints);
      setReliefForm(buildDefaultReliefPointForm(fallbackReliefPoints[0]));
      setReliefHistory([]);
      setReliefLoadState("fallback");
      setReliefFeedback("Relief point API unavailable. Showing fixtures.");
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
      setShelterForm((current) => {
        if (key === "shelterId") {
          const shelter = shelters.find((item) => item.id === value);
          return shelter ? buildDefaultShelterForm(shelter) : current;
        }
        return { ...current, [key]: value };
      });
    };

  const updateReliefForm =
    (key: keyof ReliefPointFormState) =>
    (
      event:
        | ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
        | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setReliefForm((current) => {
        if (key === "reliefPointId") {
          if (value === "__new__") {
            setReliefHistory([]);
            return buildDefaultReliefPointForm();
          }
          const reliefPoint = reliefPoints.find((item) => item.id === value);
          if (reliefPoint) {
            void refreshReliefHistory(reliefPoint.id);
            return buildDefaultReliefPointForm(reliefPoint);
          }
          return current;
        }
        return { ...current, [key]: value };
      });
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
        throw new Error(`incident API returned ${response.status}`);
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setStatusFeedback(`${statusLabel(incident.status)} status saved.`);
    } catch (error) {
      setStatusFeedback(
        "Incident workflow action needs the incident-service API running with this incident.",
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
        throw new Error(`incident API returned ${response.status}`);
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setStatusFeedback(`${statusLabel(incident.status)} status saved.`);
    } catch (error) {
      setStatusFeedback(
        "Incident workflow action needs a valid live incident transition.",
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
        throw new Error(`incident API returned ${response.status}`);
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setAbuseFeedback(
        `${abuseDecisionLabel(request.decision)} review saved for ${incident.reference}.`,
      );
    } catch (error) {
      setAbuseFeedback(
        "Report safety review needs a live incident-service API and valid authority transition.",
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
        throw new Error(`incident API returned ${response.status}`);
      }
      const incident = (await response.json()) as IncidentRecord;
      applyIncidentUpdate(incident);
      setAssignmentFeedback(`Assigned to ${assignmentForIncident(incident)}.`);
    } catch (error) {
      setAssignmentFeedback(
        "Assignment needs a verified live incident and incident-service API.",
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
        throw new Error(`incident API returned ${response.status}`);
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
        "Merge needs a live duplicate candidate and incident-service API.",
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

  const updateShelterCapacity = async () => {
    if (!selectedShelter) {
      return;
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
      return;
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
        throw new Error(`shelter API returned ${response.status}`);
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
    } catch (error) {
      setShelterFeedback(
        "Shelter update needs a live shelter-service API and authority session.",
      );
    } finally {
      setShelterBusy(false);
    }
  };

  const saveReliefPoint = async () => {
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
      return;
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
      return;
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
        throw new Error(`relief point API returned ${response.status}`);
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
    } catch (error) {
      setReliefFeedback(
        "Relief point save needs a live shelter-service API and authority session.",
      );
    } finally {
      setReliefBusy(false);
    }
  };

  const createAlertDraft = async () => {
    setAlertBusy(true);
    setAlertFeedback("");
    try {
      const response = await fetch(`${ALERT_API_BASE}/alerts`, {
        method: "POST",
        headers: authorityHeaders(),
        body: JSON.stringify(buildAlertRequest()),
      });
      if (!response.ok) {
        throw new Error(`alert API returned ${response.status}`);
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
        "Alert API unavailable. Start alert-service to create drafts.",
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
        throw new Error(`alert API returned ${response.status}`);
      }
      const updatedAlert = (await response.json()) as AuthorityAlertRecord;
      setAlerts((current) =>
        current.map((item) =>
          item.id === updatedAlert.id ? updatedAlert : item,
        ),
      );
      setAlertLoadState("ready");
      setAlertFeedback(`${alertStatusLabel(updatedAlert.status)} alert saved.`);
    } catch (error) {
      setAlertFeedback("Alert action needs the alert-service API running.");
    } finally {
      setAlertBusy(false);
    }
  };

  return {
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
    // Imagery overlay bridge
    showImageryOverlay,
    setShowImageryOverlay,
    imageryFeatures,
    setImageryFeatures,
  };
}

export type CommandData = ReturnType<typeof useCommandData>;
