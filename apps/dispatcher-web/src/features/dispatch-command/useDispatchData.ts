import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import type { SelectChangeEvent } from "@mui/material/Select";
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
  IncidentTriageResponse,
  IncidentTriageReviewRequest,
  IncidentTriageReviewResponse,
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
  FIXTURE_DATA_ENABLED,
  INCIDENT_API_BASE,
  ML_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "@/app/config";
import { dispatcherHeaders } from "@/app/session";
import {
  defaultFilters,
  defaultHospitalCapacityFilters,
  fallbackMLPredictions,
  fallbackTriageSuggestion,
  predictionReviewPoints,
  assignmentAgencyOptions,
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
  TriageLoadState,
  TriageSuggestionFormState,
  TriageSuggestionReview,
} from "./types";
import {
  alertDatesError,
  alertStatusLabel,
  abuseDecisionLabel,
  assignmentForIncident,
  buildAlertTarget,
  buildDefaultAbuseReviewForm,
  buildDefaultAlertForm,
  buildDefaultAssignmentForm,
  buildDefaultStatusForm,
  buildDefaultTriageForm,
  buildAlertRequestFromPrediction,
  buildFilterOptions,
  buildQueueMetrics,
  duplicateReviewCandidatesFor,
  enrichIncidentFromAPI,
  hazardLabel,
  matchesFilters,
  predictionResponseToReview,
  statusLabel,
  triageAcceptRequest,
  triageOverrideRequestFromForm,
  triagePopulationError,
  triageReasonError,
  triageSuggestionFromResponse,
} from "./utils";

/**
 * Central dispatch-console state container. Holds every incident, alert, ML
 * prediction, triage suggestion, hospital capacity, road closure, and relief
 * workflow that the console views depend on, so the shell mounts data once and
 * routes between views without losing state or re-fetching.
 */
export function useDispatchData() {
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
  const [mlPredictions, setMlPredictions] = useState<MLPredictionReview[]>(
    FIXTURE_DATA_ENABLED ? fallbackMLPredictions : [],
  );
  const [mlReviewLoadState, setMlReviewLoadState] =
    useState<MLReviewLoadState>("loading");
  const [mlReviewMessage, setMlReviewMessage] = useState(
    "Loading ML flood predictions",
  );
  const [selectedPredictionId, setSelectedPredictionId] = useState(
    FIXTURE_DATA_ENABLED ? (fallbackMLPredictions[0]?.id ?? "") : "",
  );
  const [mlDraftBusy, setMlDraftBusy] = useState(false);
  const [mlDraftFeedback, setMlDraftFeedback] = useState("");
  const [predictionReviewNotes, setPredictionReviewNotes] = useState<
    Record<string, string>
  >({});
  const [triageSuggestion, setTriageSuggestion] = useState<
    TriageSuggestionReview | undefined
  >(undefined);
  const [triageLoadState, setTriageLoadState] =
    useState<TriageLoadState>("loading");
  const [triageMessage, setTriageMessage] = useState(
    "Loading AI triage suggestion",
  );
  const [triageBusy, setTriageBusy] = useState(false);
  const [triageFeedback, setTriageFeedback] = useState("");
  const [triageForm, setTriageForm] = useState<TriageSuggestionFormState>(
    buildDefaultTriageForm(),
  );
  const triageAbortRef = useRef<AbortController | null>(null);
  const [hospitalFacilities, setHospitalFacilities] = useState<
    HospitalCapacityRecord[]
  >([]);
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

      setIncidents([]);
      setSelectedIncidentId("");
      setLoadState("error");
      setLoadMessage(
        "Incident API unavailable. Could not load the incident feed.",
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

      setAlerts([]);
      setAlertLoadState("error");
      setAlertMessage(
        "Alert API unavailable. Could not load the alert workflow.",
      );
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

      if (FIXTURE_DATA_ENABLED) {
        setMlPredictions(fallbackMLPredictions);
        setSelectedPredictionId(fallbackMLPredictions[0]?.id ?? "");
        setMlReviewLoadState("fallback");
        setMlReviewMessage(
          "ML service unavailable. Showing baseline prediction fixture data (dev only).",
        );
        return;
      }

      setMlPredictions([]);
      setSelectedPredictionId("");
      setMlReviewLoadState("error");
      setMlReviewMessage(
        "ML service unavailable. Live predictions could not be loaded.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshMLPredictions(controller.signal);
    return () => controller.abort();
  }, []);

  const refreshTriage = async (incidentId: string) => {
    triageAbortRef.current?.abort();
    const controller = new AbortController();
    triageAbortRef.current = controller;

    setTriageLoadState("loading");
    setTriageMessage("Loading AI triage suggestion");

    try {
      const response = await fetch(
        `${INCIDENT_API_BASE}/incidents/${incidentId}/triage`,
        {
          headers: dispatcherHeaders(),
          signal: controller.signal,
        },
      );
      if (!response.ok) {
        throw new Error(`incident API returned ${response.status}`);
      }

      const payload = (await response.json()) as IncidentTriageResponse;
      if (controller.signal.aborted) {
        return;
      }
      const suggestion = triageSuggestionFromResponse(incidentId, payload);
      setTriageSuggestion(suggestion);
      setTriageForm(buildDefaultTriageForm(suggestion));
      setTriageLoadState("ready");
      setTriageMessage("Live AI triage connected.");
    } catch (error) {
      if (controller.signal.aborted) {
        return;
      }

      const incident = incidents.find((item) => item.id === incidentId);
      if (!incident || !FIXTURE_DATA_ENABLED) {
        setTriageSuggestion(undefined);
        setTriageForm(buildDefaultTriageForm());
        setTriageLoadState("error");
        setTriageMessage("Incident triage API unavailable.");
        return;
      }

      const fallback = fallbackTriageSuggestion(incident);
      setTriageSuggestion(fallback);
      setTriageForm(buildDefaultTriageForm(fallback));
      setTriageLoadState("fallback");
      setTriageMessage(
        "Incident triage API unavailable. Showing rule-based suggestion (dev only).",
      );
    }
  };

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

  const refreshHospitalCapacity = async (signal?: AbortSignal) => {
    const anchor = selectedIncident?.location ?? { lat: 5.56, lng: -0.2 };
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

      setHospitalFacilities([]);
      setHospitalLoadState("fallback");
      setHospitalMessage(
        "Hospital capacity API unavailable. Could not load facility capacity.",
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
    setTriageFeedback("");

    if (!selectedIncident) {
      setTriageSuggestion(undefined);
      setTriageForm(buildDefaultTriageForm());
      setTriageLoadState("empty");
      setTriageMessage("Select an incident to see its AI triage suggestion.");
      return;
    }

    if (FIXTURE_DATA_ENABLED) {
      const fallback = fallbackTriageSuggestion(selectedIncident);
      setTriageSuggestion(fallback);
      setTriageForm(buildDefaultTriageForm(fallback));
    } else {
      setTriageSuggestion(undefined);
      setTriageForm(buildDefaultTriageForm());
    }

    if (selectedIncident.source !== "api") {
      if (FIXTURE_DATA_ENABLED) {
        setTriageLoadState("fallback");
        setTriageMessage(
          "Fixture incident: showing the rule-based fixture suggestion. Reviews are not logged for fixture data.",
        );
      } else {
        setTriageSuggestion(undefined);
        setTriageForm(buildDefaultTriageForm());
        setTriageLoadState("error");
        setTriageMessage(
          "Fixture incident data is not available outside development.",
        );
      }
    }

    const controller = new AbortController();
    if (selectedIncident.source === "api") {
      void refreshTriage(selectedIncident.id);
    }
    if (
      selectedIncident.source === "api" &&
      selectedIncident.duplicateCandidates.length
    ) {
      void refreshDuplicateReview(selectedIncident.id, controller.signal);
    }
    return () => {
      controller.abort();
      triageAbortRef.current?.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  const updateTriageForm =
    (key: keyof TriageSuggestionFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setTriageForm((current) => {
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
            agencyType: agency.type,
          };
        }
        return { ...current, [key]: value };
      });
    };

  const copyTriageSuggestionToAssignmentForm = () => {
    if (!triageSuggestion) {
      return;
    }
    const agency = assignmentAgencyOptions.find(
      (item) => item.type === triageSuggestion.suggestedAgency.agencyType,
    );
    if (agency) {
      setAssignmentForm((current) => ({
        ...current,
        agencyId: agency.id,
        agencyName: agency.name,
        agencyType: agency.type,
        responderLead: agency.responderLead,
        instructions: selectedIncident
          ? `Respond to ${hazardLabel(selectedIncident.type).toLowerCase()} incident ${selectedIncident.reference}. ${selectedIncident.description}`
          : current.instructions,
      }));
    }
  };

  const postTriageReview = async (
    incidentId: string,
    request: IncidentTriageReviewRequest,
  ): Promise<IncidentTriageReviewResponse> => {
    const response = await fetch(
      `${INCIDENT_API_BASE}/incidents/${incidentId}/triage-review`,
      {
        method: "POST",
        headers: dispatcherHeaders(),
        body: JSON.stringify(request),
      },
    );
    if (!response.ok) {
      const body = (await response.json().catch(() => null)) as {
        error?: { message?: string };
      } | null;
      throw new Error(
        body?.error?.message ?? `incident API returned ${response.status}`,
      );
    }
    return (await response.json()) as IncidentTriageReviewResponse;
  };

  const triageErrorDetail = (error: unknown) =>
    error instanceof Error && error.message
      ? error.message
      : "The incident-service API must be running for this incident.";

  const acceptTriageSuggestion = async () => {
    if (!selectedIncident || !triageSuggestion) {
      return;
    }

    if (selectedIncident.source !== "api") {
      copyTriageSuggestionToAssignmentForm();
      setTriageFeedback(
        "Suggestion copied to the agency assignment form. Acceptance logging needs the incident-service API for this incident.",
      );
      return;
    }

    setTriageBusy(true);
    setTriageFeedback("");
    try {
      const payload = await postTriageReview(
        selectedIncident.id,
        triageAcceptRequest(triageSuggestion),
      );
      applyIncidentUpdate(payload.incident);
      copyTriageSuggestionToAssignmentForm();
      setTriageFeedback(
        `Triage acceptance logged for ${payload.incident.reference}. Suggestion copied to the agency assignment form. Review before assigning.`,
      );
    } catch (error) {
      setTriageFeedback(
        `Acceptance was not logged, so the suggestion was not applied. ${triageErrorDetail(error)}`,
      );
    } finally {
      setTriageBusy(false);
    }
  };

  const overrideTriageSuggestion = async () => {
    if (!selectedIncident || !triageSuggestion) {
      return;
    }

    if (selectedIncident.source !== "api") {
      setTriageFeedback(
        "Overrides can only be logged for live incidents from the incident-service API.",
      );
      return;
    }

    const request: IncidentTriageReviewRequest = triageOverrideRequestFromForm(
      triageSuggestion,
      triageForm,
    );

    setTriageBusy(true);
    setTriageFeedback("");
    try {
      const payload = await postTriageReview(selectedIncident.id, request);
      const incident = payload.incident;
      applyIncidentUpdate(incident);
      const agency = assignmentAgencyOptions.find(
        (item) => item.id === triageForm.agencyId,
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
      setTriageFeedback(
        `Triage override recorded for ${incident.reference}. Assignment form updated.`,
      );
    } catch (error) {
      setTriageFeedback(
        `Triage override was not recorded. ${triageErrorDetail(error)}`,
      );
    } finally {
      setTriageBusy(false);
    }
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
    // Validate the schedule before any request: invalid dates would throw a
    // RangeError in buildAlertRequest and surface as a fake API failure.
    const datesError = alertDatesError(alertForm);
    if (datesError) {
      setAlertFeedback(datesError);
      return;
    }

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

    // Safety gate: a reviewed draft must never be built from fixture
    // predictions — only live ML output may drive a real alert draft.
    if (mlReviewLoadState !== "ready") {
      setMlDraftFeedback(
        "Live ML predictions unavailable — a reviewed draft cannot be created from fixture data.",
      );
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
        const errorBody = (await response.json().catch(() => null)) as {
          error?: { code?: string; message?: string };
        } | null;
        throw new Error(
          errorBody?.error?.message ?? `alert API returned ${response.status}`,
        );
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
      setAlertFeedback(
        error instanceof Error && error.message
          ? error.message
          : "Alert action needs the alert-service API running.",
      );
    } finally {
      setAlertBusy(false);
    }
  };

  const selectedPredictionReviewNote = selectedPrediction
    ? (predictionReviewNotes[selectedPrediction.id] ??
      `Reviewed ${selectedPrediction.modelVersion} prediction for ${selectedPrediction.community}.`)
    : "";

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
    // AI triage
    triageSuggestion,
    triageLoadState,
    triageMessage,
    triageBusy,
    triageFeedback,
    triageForm,
    updateTriageForm,
    acceptTriageSuggestion,
    overrideTriageSuggestion,
    refreshTriage,
    triagePopulationError,
    triageReasonError,
    // ML review
    mlPredictions,
    mlReviewLoadState,
    mlReviewMessage,
    selectedPrediction,
    selectedPredictionId,
    setSelectedPredictionId,
    mlDraftBusy,
    mlDraftFeedback,
    predictionReviewNotes,
    selectedPredictionReviewNote,
    updatePredictionReviewNote,
    createAlertDraftFromPrediction,
    refreshMLPredictions,
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
    // Hospital capacity
    hospitalFacilities,
    hospitalLoadState,
    hospitalMessage,
    hospitalFilters,
    updateHospitalCapacityFilter,
    updateHospitalIncludeStale,
    updateHospitalMinBeds,
    updateHospitalServiceFilter,
    refreshHospitalCapacity,
    // Map layers + relief
    roadClosures,
    roadClosureLoadState,
    roadClosureMessage,
    reliefPoints,
    reliefPointLoadState,
    reliefPointMessage,
    refreshReliefPoints,
  };
}

export type DispatchData = ReturnType<typeof useDispatchData>;
