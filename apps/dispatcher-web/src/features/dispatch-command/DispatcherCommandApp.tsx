import { type ChangeEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  AppBar,
  Box,
  Button,
  Chip,
  Container,
  CssBaseline,
  Grid,
  LinearProgress,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  ThemeProvider,
  Toolbar,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  Eye,
  Filter,
  LocateFixed,
  MapPinned,
  RadioTower,
  RefreshCw,
  ShieldAlert,
  Truck,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AssignIncidentRequest,
  AlertListResponse,
  AuthorityAlertRecord,
  CreateAlertRequest,
  DuplicateReviewCandidate,
  DuplicateReviewResponse,
  HospitalCapacityRecord,
  HospitalCapacityResponse,
  IncidentAbuseReviewRequest,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatusUpdateRequest,
  MergeIncidentsRequest,
  MergeIncidentsResponse,
  MLPredictionResponse,
  ReliefPointListResponse,
  ReliefPointRecord,
  RoadClosureListResponse,
  RoadClosureRecord,
} from "@nadaa/shared-types";
import {
  ALERT_API_BASE,
  INCIDENT_API_BASE,
  ML_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import {
  dispatcherHeaders,
  dispatcherSession,
  commandRoles,
} from "../../app/session";
import { dispatcherTheme } from "../../app/theme";
import {
  AlertWorkflowPanel,
  CommandSelect,
  EmptyState,
  HospitalCapacityPanel,
  IncidentDetailPanel,
  IncidentMap,
  MLPredictionReviewPanel,
  PrivacyChip,
  ReliefPointPanel,
  StatusLine,
} from "./components";
import {
  defaultFilters,
  defaultHospitalCapacityFilters,
  fallbackAlerts,
  fallbackHospitalFacilities,
  fallbackIncidents,
  fallbackMLPredictions,
  predictionReviewPoints,
  assignmentAgencyOptions,
  severityColors,
} from "./data";
import type {
  AbuseReviewFormState,
  AlertFormState,
  AlertLoadState,
  AssignmentFormState,
  CapacityLoadState,
  CommandIncident,
  FilterState,
  HospitalCapacityFilterState,
  IncidentLoadState,
  IncidentStatusFormState,
  MLPredictionReview,
  MLReviewLoadState,
} from "./types";
import {
  alertStatusLabel,
  abuseDecisionLabel,
  assignmentForIncident,
  buildAlertTarget,
  buildDefaultAbuseReviewForm,
  buildDefaultAlertForm,
  buildDefaultAssignmentForm,
  buildDefaultStatusForm,
  buildAlertRequestFromPrediction,
  buildFilterOptions,
  buildQueueMetrics,
  duplicateReviewCandidatesFor,
  enrichIncidentFromAPI,
  formatIncidentAge,
  hazardLabel,
  matchesFilters,
  predictionResponseToReview,
  severityLabel,
  statusLabel,
} from "./utils";

function DispatcherCommandApp() {
  const hasCommandAccess =
    commandRoles.includes(dispatcherSession.role) &&
    dispatcherSession.mfaEnabled;
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
  const [mlPredictions, setMlPredictions] = useState<MLPredictionReview[]>(
    fallbackMLPredictions,
  );
  const [mlReviewLoadState, setMlReviewLoadState] =
    useState<MLReviewLoadState>("loading");
  const [mlReviewMessage, setMlReviewMessage] = useState(
    "Loading ML flood predictions",
  );
  const [selectedPredictionId, setSelectedPredictionId] = useState(
    fallbackMLPredictions[0]?.id ?? "",
  );
  const [mlDraftBusy, setMlDraftBusy] = useState(false);
  const [mlDraftFeedback, setMlDraftFeedback] = useState("");
  const [predictionReviewNotes, setPredictionReviewNotes] = useState<
    Record<string, string>
  >({});
  const [hospitalFacilities, setHospitalFacilities] = useState<
    HospitalCapacityRecord[]
  >(fallbackHospitalFacilities);
  const [hospitalLoadState, setHospitalLoadState] =
    useState<CapacityLoadState>("loading");
  const [hospitalMessage, setHospitalMessage] = useState(
    "Loading hospital capacity",
  );
  const [hospitalFilters, setHospitalFilters] =
    useState<HospitalCapacityFilterState>(defaultHospitalCapacityFilters);
  const [roadClosures, setRoadClosures] = useState<RoadClosureRecord[]>([]);
  const [roadClosureLoadState, setRoadClosureLoadState] =
    useState<CapacityLoadState>("loading");
  const [roadClosureMessage, setRoadClosureMessage] = useState(
    "Loading road closures",
  );
  const [reliefPoints, setReliefPoints] = useState<ReliefPointRecord[]>([]);
  const [reliefPointLoadState, setReliefPointLoadState] =
    useState<CapacityLoadState>("loading");
  const [reliefPointMessage, setReliefPointMessage] = useState(
    "Loading relief distribution points",
  );

  const refreshIncidents = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading incident feed");

    try {
      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
        headers: dispatcherHeaders(),
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
        headers: dispatcherHeaders(),
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

  const refreshMLPredictions = async (signal?: AbortSignal) => {
    setMlReviewLoadState("loading");
    setMlReviewMessage("Loading ML flood predictions");

    try {
      const predictions = await Promise.all(
        predictionReviewPoints.map(async (point) => {
          const response = await fetch(`${ML_API_BASE}/ml/flood/predictions`, {
            method: "POST",
            headers: dispatcherHeaders(),
            signal,
            body: JSON.stringify({
              location: point.location,
              requestedBy: "dispatcher-web",
              correlationId: `ml-review-${point.id}`,
            }),
          });
          if (!response.ok) {
            throw new Error(`ML API returned ${response.status}`);
          }
          const payload = (await response.json()) as MLPredictionResponse;
          return predictionResponseToReview(payload, point);
        }),
      );

      setMlPredictions(predictions);
      setSelectedPredictionId(predictions[0]?.id ?? "");
      setMlReviewLoadState("ready");
      setMlReviewMessage("Live ML predictions connected.");
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setMlPredictions(fallbackMLPredictions);
      setSelectedPredictionId(fallbackMLPredictions[0]?.id ?? "");
      setMlReviewLoadState("fallback");
      setMlReviewMessage(
        "ML service unavailable. Showing baseline prediction fixture data.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshMLPredictions(controller.signal);
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

  const selectedPrediction = useMemo(() => {
    if (!mlPredictions.length) {
      return undefined;
    }
    return (
      mlPredictions.find(
        (prediction) => prediction.id === selectedPredictionId,
      ) ?? mlPredictions[0]
    );
  }, [mlPredictions, selectedPredictionId]);

  const filteredFallbackHospitalFacilities = useMemo(
    () =>
      fallbackHospitalFacilities.filter((facility) =>
        matchesHospitalCapacityFilters(facility, hospitalFilters),
      ),
    [hospitalFilters],
  );

  const refreshHospitalCapacity = async (signal?: AbortSignal) => {
    const anchor = selectedIncident?.location ??
      fallbackIncidents[0]?.location ?? { lat: 5.56, lng: -0.2 };
    const params = new URLSearchParams({
      includeStale: String(hospitalFilters.includeStale),
      lat: String(anchor.lat),
      limit: "6",
      lng: String(anchor.lng),
    });
    const minAvailableBeds = Number.parseInt(
      hospitalFilters.minAvailableBeds,
      10,
    );
    if (
      Number.isFinite(minAvailableBeds) &&
      !Number.isNaN(minAvailableBeds) &&
      minAvailableBeds > 0
    ) {
      params.set("minAvailableBeds", String(minAvailableBeds));
    }
    if (hospitalFilters.service !== "all") {
      params.set("service", hospitalFilters.service);
    }
    if (hospitalFilters.emergencyCapacity !== "all") {
      params.set("emergencyCapacity", hospitalFilters.emergencyCapacity);
    }

    setHospitalLoadState("loading");
    setHospitalMessage("Loading hospital capacity");
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/hospitals/capacity?${params.toString()}`,
        {
          headers: dispatcherHeaders(),
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`hospital capacity API returned ${response.status}`);
      }

      const payload = (await response.json()) as HospitalCapacityResponse;
      setHospitalFacilities(payload.facilities);
      setHospitalLoadState(payload.facilities.length ? "ready" : "empty");
      setHospitalMessage(
        payload.facilities.length
          ? `Live hospital capacity connected. Stale threshold ${payload.staleThresholdMinutes} minutes.`
          : "No hospitals match the current capacity filters.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }

      setHospitalFacilities(filteredFallbackHospitalFacilities);
      setHospitalLoadState("fallback");
      setHospitalMessage(
        "Hospital capacity API unavailable. Showing facility capacity fixtures.",
      );
    }
  };

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

  useEffect(() => {
    if (!mlPredictions.length) {
      setSelectedPredictionId("");
      return;
    }

    if (
      !mlPredictions.some(
        (prediction) => prediction.id === selectedPredictionId,
      )
    ) {
      setSelectedPredictionId(mlPredictions[0].id);
    }
  }, [mlPredictions, selectedPredictionId]);

  useEffect(() => {
    const controller = new AbortController();
    void refreshHospitalCapacity(controller.signal);
    return () => controller.abort();
  }, [
    filteredFallbackHospitalFacilities,
    hospitalFilters.emergencyCapacity,
    hospitalFilters.includeStale,
    hospitalFilters.minAvailableBeds,
    hospitalFilters.service,
    selectedIncident?.id,
  ]);

  const refreshRoadClosures = async (signal?: AbortSignal) => {
    setRoadClosureLoadState("loading");
    setRoadClosureMessage("Loading road closures");
    try {
      const response = await fetch(`${ROAD_CLOSURE_API_BASE}/road-closures`, {
        headers: dispatcherHeaders(),
        signal,
      });
      if (!response.ok) {
        throw new Error(`road closure API returned ${response.status}`);
      }
      const payload = (await response.json()) as RoadClosureListResponse;
      setRoadClosures(payload.closures);
      setRoadClosureLoadState(payload.closures.length ? "ready" : "empty");
      setRoadClosureMessage(
        payload.closures.length
          ? "Live road closures connected."
          : "No active road closures reported.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setRoadClosures([]);
      setRoadClosureLoadState("fallback");
      setRoadClosureMessage(
        "Road closure API unavailable. Map layer disabled.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshRoadClosures(controller.signal);
    return () => controller.abort();
  }, []);

  const refreshReliefPoints = async (signal?: AbortSignal) => {
    setReliefPointLoadState("loading");
    setReliefPointMessage("Loading relief distribution points");
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/relief-points?limit=50`,
        {
          headers: dispatcherHeaders(),
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`relief point API returned ${response.status}`);
      }
      const payload = (await response.json()) as ReliefPointListResponse;
      setReliefPoints(payload.reliefPoints);
      setReliefPointLoadState(payload.reliefPoints.length ? "ready" : "empty");
      setReliefPointMessage(
        payload.reliefPoints.length
          ? "Live relief points connected."
          : "No relief distribution points reported.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setReliefPoints([]);
      setReliefPointLoadState("fallback");
      setReliefPointMessage(
        "Relief point API unavailable. Map layer disabled.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshReliefPoints(controller.signal);
    return () => controller.abort();
  }, []);

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
  }, [selectedIncident?.id]);

  const updateFilter =
    (key: keyof FilterState) => (event: SelectChangeEvent) => {
      setFilters((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateHospitalCapacityFilter = (event: SelectChangeEvent) => {
    setHospitalFilters((current) => ({
      ...current,
      emergencyCapacity: event.target
        .value as HospitalCapacityFilterState["emergencyCapacity"],
    }));
  };

  const updateHospitalIncludeStale = (checked: boolean) => {
    setHospitalFilters((current) => ({ ...current, includeStale: checked }));
  };

  const updateHospitalMinBeds = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    setHospitalFilters((current) => ({
      ...current,
      minAvailableBeds: event.target.value,
    }));
  };

  const updateHospitalServiceFilter = (event: SelectChangeEvent) => {
    setHospitalFilters((current) => ({
      ...current,
      service: event.target.value as HospitalCapacityFilterState["service"],
    }));
  };

  const updateAlertForm =
    (key: keyof AlertFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      const value =
        "checked" in event.target && typeof event.target.checked === "boolean"
          ? event.target.checked
          : event.target.value;
      setAlertForm((current) => ({ ...current, [key]: value }));
    };

  const updatePredictionReviewNote = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    if (!selectedPrediction) {
      return;
    }
    setPredictionReviewNotes((current) => ({
      ...current,
      [selectedPrediction.id]: event.target.value,
    }));
  };

  const updateStatusForm =
    (key: keyof IncidentStatusFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      setStatusForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateAbuseForm =
    (key: keyof AbuseReviewFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      setAbuseForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateAssignmentForm =
    (key: keyof AssignmentFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
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

  const refreshDuplicateReview = async (
    incidentId: string,
    signal?: AbortSignal,
  ) => {
    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${incidentId}/duplicates`,
        {
          headers: dispatcherHeaders(),
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
          headers: dispatcherHeaders(),
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
          headers: dispatcherHeaders(),
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
          headers: dispatcherHeaders(),
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
        "Report safety review needs a live incident-service API and valid dispatcher transition.",
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
          headers: dispatcherHeaders(),
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
          headers: dispatcherHeaders(),
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

  const createAlertDraft = async () => {
    setAlertBusy(true);
    setAlertFeedback("");
    try {
      const response = await fetch(`${ALERT_API_BASE}/alerts`, {
        method: "POST",
        headers: dispatcherHeaders(),
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

  const createAlertDraftFromPrediction = async () => {
    if (!selectedPrediction) {
      return;
    }

    const reviewNote =
      predictionReviewNotes[selectedPrediction.id] ??
      `Reviewed ${selectedPrediction.modelVersion} prediction for ${selectedPrediction.community}.`;

    setMlDraftBusy(true);
    setMlDraftFeedback("");
    try {
      const response = await fetch(`${ALERT_API_BASE}/alerts`, {
        method: "POST",
        headers: dispatcherHeaders(),
        body: JSON.stringify(
          buildAlertRequestFromPrediction(selectedPrediction, reviewNote),
        ),
      });
      if (!response.ok) {
        throw new Error(`alert API returned ${response.status}`);
      }

      const alert = (await response.json()) as AuthorityAlertRecord;
      setAlerts((current) => [
        alert,
        ...current.filter((item) => item.id !== alert.id),
      ]);
      setMlPredictions((current) =>
        current.map((prediction) =>
          prediction.id === selectedPrediction.id
            ? { ...prediction, reviewStatus: "draft_created" }
            : prediction,
        ),
      );
      setAlertLoadState("ready");
      setMlDraftFeedback(
        `Draft ${alert.id} created from ${selectedPrediction.community} prediction.`,
      );
      setAlertFeedback("ML-reviewed draft created. Submit it for approval.");
    } catch (error) {
      setMlDraftFeedback(
        "Alert API unavailable. Start alert-service to create an ML-reviewed draft.",
      );
    } finally {
      setMlDraftBusy(false);
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
        ? { reason: "Rejected from dispatcher command review queue." }
        : action === "emergency-override"
          ? { reason: "Emergency override from dispatcher command console." }
          : { note: `Dispatcher command console ${action}.` };

    try {
      const response = await fetch(
        `${ALERT_API_BASE}/alerts/${alert.id}/${action}`,
        {
          method: "POST",
          headers: dispatcherHeaders(),
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

  if (!hasCommandAccess) {
    return (
      <ThemeProvider theme={dispatcherTheme}>
        <CssBaseline />
        <Container maxWidth="sm" className="access-shell">
          <Paper className="surface access-panel">
            <ShieldAlert size={38} color={nadaaBrand.colors.red} />
            <Typography variant="h5">Dispatcher access required</Typography>
            <Typography color="text.secondary">
              Incident command requires an agency account, an approved role, and
              completed MFA.
            </Typography>
          </Paper>
        </Container>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={dispatcherTheme}>
      <CssBaseline />
      <AppBar position="sticky" elevation={0} className="topbar">
        <Toolbar className="toolbar">
          <Stack
            direction="row"
            spacing={1.5}
            alignItems="center"
            className="brand-lockup"
          >
            <Box
              component="img"
              src="/brand/nadaa-logo.png"
              alt="NADAA shield"
              className="brand-logo"
            />
            <Box>
              <Typography variant="h6">NADAA Dispatch Command</Typography>
              <Typography variant="caption">
                National Disaster Alert & Response Platform
              </Typography>
            </Box>
          </Stack>
          <Stack direction="row" spacing={1} className="topbar-actions">
            <Chip
              size="small"
              color="success"
              label={`${dispatcherSession.name} / MFA`}
              className="session-chip"
            />
            <Button
              color="inherit"
              variant="outlined"
              startIcon={<RadioTower size={17} />}
            >
              Issue alert
            </Button>
            <Button
              color="secondary"
              variant="contained"
              startIcon={<Truck size={17} />}
            >
              Assign team
            </Button>
          </Stack>
        </Toolbar>
      </AppBar>

      <Container maxWidth="xl" className="dashboard-shell">
        <Stack
          direction={{ xs: "column", md: "row" }}
          justifyContent="space-between"
          gap={2}
          className="page-heading"
        >
          <Box>
            <Typography variant="overline" color="secondary">
              Incident command
            </Typography>
            <Typography variant="h4">
              Live Greater Accra operations map
            </Typography>
            <Typography color="text.secondary">
              Monitor emergencies by place, severity, hazard, time, and response
              status.
            </Typography>
          </Box>
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            flexWrap="wrap"
          >
            <Chip
              icon={<Eye size={16} />}
              label={
                loadState === "ready"
                  ? "Live API"
                  : loadState === "empty"
                    ? "No active incidents"
                    : "Fixture mode"
              }
              color={
                loadState === "ready"
                  ? "success"
                  : loadState === "empty"
                    ? "default"
                    : "warning"
              }
            />
            <Button
              variant="outlined"
              startIcon={<RefreshCw size={17} />}
              onClick={() => void refreshIncidents()}
              disabled={loadState === "loading"}
            >
              Refresh
            </Button>
          </Stack>
        </Stack>

        {loadState === "fallback" ||
        loadState === "error" ||
        loadState === "empty" ? (
          <Alert
            severity={loadState === "empty" ? "info" : "warning"}
            className="feed-alert"
          >
            {loadMessage}
          </Alert>
        ) : null}
        {loadState === "loading" ? (
          <LinearProgress className="feed-progress" />
        ) : null}

        <Grid container spacing={2.5}>
          {metrics.map((item) => {
            const Icon = item.icon;
            return (
              <Grid size={{ xs: 12, sm: 6, lg: 3 }} key={item.label}>
                <Paper className="metric-card">
                  <Stack
                    direction="row"
                    justifyContent="space-between"
                    alignItems="center"
                  >
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        {item.label}
                      </Typography>
                      <Typography variant="h3">{item.value}</Typography>
                    </Box>
                    <Box className="metric-icon" style={{ color: item.color }}>
                      <Icon size={28} />
                    </Box>
                  </Stack>
                </Paper>
              </Grid>
            );
          })}
        </Grid>

        <MLPredictionReviewPanel
          busy={mlDraftBusy}
          feedback={mlDraftFeedback}
          loadMessage={mlReviewMessage}
          loadState={mlReviewLoadState}
          onCreateDraft={() => void createAlertDraftFromPrediction()}
          onRefresh={() => void refreshMLPredictions()}
          onSelectPrediction={setSelectedPredictionId}
          onUpdateReviewNote={updatePredictionReviewNote}
          predictions={mlPredictions}
          reviewNote={
            selectedPrediction
              ? (predictionReviewNotes[selectedPrediction.id] ??
                `Reviewed ${selectedPrediction.modelVersion} prediction for ${selectedPrediction.community}.`)
              : ""
          }
          selectedPrediction={selectedPrediction}
          selectedPredictionId={selectedPredictionId}
        />

        <HospitalCapacityPanel
          facilities={hospitalFacilities}
          filters={hospitalFilters}
          loadMessage={hospitalMessage}
          loadState={hospitalLoadState}
          onRefresh={() => void refreshHospitalCapacity()}
          onUpdateCapacity={updateHospitalCapacityFilter}
          onUpdateIncludeStale={updateHospitalIncludeStale}
          onUpdateMinBeds={updateHospitalMinBeds}
          onUpdateService={updateHospitalServiceFilter}
        />

        <ReliefPointPanel
          loadMessage={reliefPointMessage}
          loadState={reliefPointLoadState}
          onRefresh={() => void refreshReliefPoints()}
          reliefPoints={reliefPoints}
        />

        <Paper className="surface filter-surface">
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            className="section-heading"
          >
            <Filter size={20} color={nadaaBrand.colors.navy} />
            <Typography variant="h6">Filters</Typography>
          </Stack>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect
                label="Hazard"
                value={filters.hazard}
                onChange={updateFilter("hazard")}
              >
                <MenuItem value="all">All hazards</MenuItem>
                {filterOptions.hazards.map((hazard) => (
                  <MenuItem value={hazard} key={hazard}>
                    {hazardLabel(hazard)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect
                label="Region / district"
                value={filters.regionDistrict}
                onChange={updateFilter("regionDistrict")}
              >
                <MenuItem value="all">All districts</MenuItem>
                {filterOptions.regionDistricts.map((district) => (
                  <MenuItem value={district} key={district}>
                    {district}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect
                label="Severity"
                value={filters.severity}
                onChange={updateFilter("severity")}
              >
                <MenuItem value="all">All severities</MenuItem>
                {filterOptions.severities.map((severity) => (
                  <MenuItem value={severity} key={severity}>
                    {severityLabel(severity)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect
                label="Status"
                value={filters.status}
                onChange={updateFilter("status")}
              >
                <MenuItem value="all">All statuses</MenuItem>
                {filterOptions.statuses.map((status) => (
                  <MenuItem value={status} key={status}>
                    {statusLabel(status)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect
                label="Time"
                value={filters.time}
                onChange={updateFilter("time")}
              >
                <MenuItem value="all">Any time</MenuItem>
                <MenuItem value="1h">Last hour</MenuItem>
                <MenuItem value="6h">Last 6 hours</MenuItem>
                <MenuItem value="24h">Last 24 hours</MenuItem>
              </CommandSelect>
            </Grid>
          </Grid>
        </Paper>

        <Grid container spacing={2.5} className="main-grid">
          <Grid size={{ xs: 12, lg: 8 }}>
            <Paper className="surface map-surface">
              <Stack
                direction={{ xs: "column", md: "row" }}
                justifyContent="space-between"
                spacing={2}
              >
                <Box>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <MapPinned size={21} color={nadaaBrand.colors.navy} />
                    <Typography variant="h5">Incident map</Typography>
                  </Stack>
                  <Typography color="text.secondary">
                    {filteredIncidents.length} visible of {incidents.length}{" "}
                    incidents
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  {filterOptions.hazards.slice(0, 4).map((hazard) => (
                    <Chip
                      key={hazard}
                      label={hazardLabel(hazard)}
                      size="small"
                    />
                  ))}
                </Stack>
              </Stack>

              <IncidentMap
                incidents={filteredIncidents}
                selectedIncidentId={selectedIncident?.id}
                onSelect={setSelectedIncidentId}
                closures={roadClosures}
                reliefPoints={reliefPoints}
              />
            </Paper>

            <Paper className="surface incident-table">
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                className="section-heading"
              >
                <LocateFixed size={21} color={nadaaBrand.colors.navy} />
                <Typography variant="h6">Incident queue</Typography>
              </Stack>
              {filteredIncidents.length ? (
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Reference</TableCell>
                      <TableCell>Hazard</TableCell>
                      <TableCell>District</TableCell>
                      <TableCell>Severity</TableCell>
                      <TableCell>Status</TableCell>
                      <TableCell>Privacy</TableCell>
                      <TableCell>Assigned</TableCell>
                      <TableCell>Age</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredIncidents.map((incident) => (
                      <TableRow
                        key={incident.id}
                        hover
                        selected={incident.id === selectedIncident?.id}
                        onClick={() => setSelectedIncidentId(incident.id)}
                        className="incident-row"
                      >
                        <TableCell>
                          <Typography variant="subtitle2">
                            {incident.reference}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {incident.locality}
                          </Typography>
                        </TableCell>
                        <TableCell>{hazardLabel(incident.type)}</TableCell>
                        <TableCell>{incident.district}</TableCell>
                        <TableCell>
                          <Chip
                            size="small"
                            label={severityLabel(incident.severity)}
                            className="severity-chip"
                            style={{
                              backgroundColor:
                                severityColors[incident.severity],
                              color: "#FFFFFF",
                            }}
                          />
                        </TableCell>
                        <TableCell>{statusLabel(incident.status)}</TableCell>
                        <TableCell>
                          <PrivacyChip incident={incident} />
                        </TableCell>
                        <TableCell>{incident.assignedAgency}</TableCell>
                        <TableCell>
                          {formatIncidentAge(incident.createdAt)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <EmptyState
                  title="No incidents match these filters"
                  detail="Adjust filters or refresh the feed."
                />
              )}
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, lg: 4 }}>
            <Stack spacing={2.5}>
              <IncidentDetailPanel
                abuseBusy={abuseBusy}
                abuseFeedback={abuseFeedback}
                abuseForm={abuseForm}
                assignmentBusy={assignmentBusy}
                assignmentFeedback={assignmentFeedback}
                assignmentForm={assignmentForm}
                busy={statusBusy}
                duplicateCandidates={duplicateReviewCandidates}
                feedback={statusFeedback}
                form={statusForm}
                incident={selectedIncident}
                mergeBusy={mergeBusy}
                mergeFeedback={mergeFeedback}
                onAssign={assignSelectedIncident}
                onMergeDuplicates={mergeSelectedDuplicates}
                onReviewAbuse={reviewSelectedIncidentAbuse}
                onToggleDuplicate={toggleDuplicateSelection}
                onUpdateAbuseForm={updateAbuseForm}
                onUpdateAssignmentForm={updateAssignmentForm}
                onUpdateForm={updateStatusForm}
                onUpdateStatus={updateIncidentStatus}
                onVerify={verifySelectedIncident}
                selectedDuplicateIds={selectedDuplicateIds}
              />

              <AlertWorkflowPanel
                alerts={alerts}
                busy={alertBusy}
                feedback={alertFeedback || alertMessage}
                form={alertForm}
                loadState={alertLoadState}
                onCreateDraft={createAlertDraft}
                onRunAction={runAlertAction}
                onUpdateForm={updateAlertForm}
                selectedIncident={selectedIncident}
              />

              <Paper className="surface">
                <Typography variant="h6" className="section-heading">
                  Operating posture
                </Typography>
                <Stack spacing={1.25}>
                  <StatusLine
                    label="Incident feed"
                    value={loadState === "ready" ? "Live" : "Fixture"}
                    color={loadState === "ready" ? "success" : "warning"}
                  />
                  <StatusLine
                    label="Dispatcher session"
                    value={dispatcherSession.agency}
                    color="success"
                  />
                  <StatusLine
                    label="ML alerts"
                    value="Human review"
                    color="warning"
                  />
                  <StatusLine
                    label="Hospital capacity"
                    value={
                      hospitalLoadState === "ready"
                        ? "Live"
                        : hospitalLoadState === "empty"
                          ? "No match"
                          : hospitalLoadState === "loading"
                            ? "Loading"
                            : "Fixture"
                    }
                    color={
                      hospitalLoadState === "ready"
                        ? "success"
                        : hospitalLoadState === "empty"
                          ? "default"
                          : "warning"
                    }
                  />
                  <StatusLine
                    label="Audit trail"
                    value="Required"
                    color="success"
                  />
                </Stack>
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      </Container>
    </ThemeProvider>
  );
}

function matchesHospitalCapacityFilters(
  facility: HospitalCapacityRecord,
  filters: HospitalCapacityFilterState,
) {
  if (!filters.includeStale && facility.stale) {
    return false;
  }
  if (
    filters.emergencyCapacity !== "all" &&
    facility.emergencyCapacity !== filters.emergencyCapacity
  ) {
    return false;
  }
  if (
    filters.service !== "all" &&
    !facility.services.some((service) => service === filters.service)
  ) {
    return false;
  }
  const minAvailableBeds = Number.parseInt(filters.minAvailableBeds, 10);
  if (
    Number.isFinite(minAvailableBeds) &&
    !Number.isNaN(minAvailableBeds) &&
    minAvailableBeds > 0 &&
    facility.availableBeds < minAvailableBeds
  ) {
    return false;
  }
  return true;
}

export default DispatcherCommandApp;
