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
  TextField,
  ThemeProvider,
  Toolbar,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  Eye,
  Filter,
  LifeBuoy,
  Loader2,
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
  CreateReliefPointRequest,
  DuplicateReviewCandidate,
  DuplicateReviewResponse,
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
  ReliefStockHistoryEntry,
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
} from "../../app/config";
import {
  authorityHeaders,
  authoritySession,
  commandRoles,
} from "../../app/session";
import { authorityTheme } from "../../app/theme";
import {
  AlertWorkflowPanel,
  CommandSelect,
  EmptyState,
  HazardChip,
  IncidentDetailPanel,
  IncidentMap,
  PrivacyChip,
  ScrollableTable,
  SeverityChip,
  StatusLine,
} from "./components";
import { RoutePlannerPanel } from "./RoutePlannerPanel";
import { FloodSimulationPanel } from "./FloodSimulationPanel";
import {
  defaultFilters,
  fallbackAlerts,
  fallbackIncidents,
  fallbackReliefPoints,
  fallbackShelters,
  assignmentAgencyOptions,
  severityColors,
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
  buildDefaultStatusForm,
  buildFilterOptions,
  buildQueueMetrics,
  duplicateReviewCandidatesFor,
  enrichIncidentFromAPI,
  formatIncidentAge,
  hazardLabel,
  matchesFilters,
  severityLabel,
  statusLabel,
} from "./utils";

function CommandCenterApp() {
  const hasCommandAccess =
    commandRoles.includes(authoritySession.role) && authoritySession.mfaEnabled;
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
  const [reliefHistory, setReliefHistory] = useState<ReliefStockHistoryEntry[]>(
    [],
  );

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
  }, [selectedIncident?.id]);

  const updateFilter =
    (key: keyof FilterState) => (event: SelectChangeEvent) => {
      setFilters((current) => ({ ...current, [key]: event.target.value }));
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

  const updateShelterForm =
    (key: keyof ShelterFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
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
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
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

  if (!hasCommandAccess) {
    return (
      <ThemeProvider theme={authorityTheme}>
        <CssBaseline />
        <a href="#main-content" className="skip-link">
          Skip to main content
        </a>
        <Container
          component="main"
          id="main-content"
          maxWidth="sm"
          className="access-shell"
        >
          <Paper className="surface access-panel">
            <ShieldAlert size={38} color={nadaaBrand.colors.red} />
            <Typography variant="h5">Authority access required</Typography>
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
    <ThemeProvider theme={authorityTheme}>
      <CssBaseline />
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
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
              <Typography variant="h6">NADAA Command</Typography>
              <Typography variant="caption">
                National Disaster Alert & Response Platform
              </Typography>
            </Box>
          </Stack>
          <Stack direction="row" spacing={1} className="topbar-actions">
            <Chip
              size="small"
              color="success"
              label={`${authoritySession.name} / MFA`}
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

      <Container
        component="main"
        id="main-content"
        maxWidth="xl"
        className="dashboard-shell"
      >
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
                    <HazardChip key={hazard} hazard={hazard} />
                  ))}
                </Stack>
              </Stack>

              <IncidentMap
                incidents={filteredIncidents}
                selectedIncidentId={selectedIncident?.id}
                onSelect={setSelectedIncidentId}
              />
            </Paper>

            <Paper className="surface">
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
                <ScrollableTable label="Incident queue table">
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
                            <Typography
                              variant="caption"
                              color="text.secondary"
                            >
                              {incident.locality}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <HazardChip hazard={incident.type} />
                          </TableCell>
                          <TableCell>{incident.district}</TableCell>
                          <TableCell>
                            <SeverityChip severity={incident.severity} />
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
                </ScrollableTable>
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

              <Paper className="surface shelter-panel">
                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  spacing={1}
                  justifyContent="space-between"
                  alignItems={{ xs: "stretch", sm: "center" }}
                  className="section-heading"
                >
                  <Stack direction="row" spacing={1} alignItems="center">
                    <LifeBuoy size={21} color={nadaaBrand.colors.green} />
                    <Box>
                      <Typography variant="h6">Shelter capacity</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Update occupancy and operating status
                      </Typography>
                    </Box>
                  </Stack>
                  <Button
                    type="button"
                    variant="outlined"
                    size="small"
                    startIcon={
                      shelterLoadState === "loading" ? (
                        <Loader2 size={16} className="spin-icon" />
                      ) : (
                        <RefreshCw size={16} />
                      )
                    }
                    onClick={() => void refreshShelters()}
                    disabled={shelterLoadState === "loading"}
                  >
                    Refresh
                  </Button>
                </Stack>

                {shelterFeedback ? (
                  <Alert
                    severity={
                      shelterLoadState === "ready" ? "success" : "warning"
                    }
                    className="feed-alert"
                  >
                    {shelterFeedback}
                  </Alert>
                ) : null}

                <Stack spacing={1.5}>
                  <CommandSelect
                    label="Shelter"
                    value={shelterForm.shelterId}
                    onChange={updateShelterForm("shelterId")}
                  >
                    {shelters.map((shelter) => (
                      <MenuItem value={shelter.id} key={shelter.id}>
                        {shelter.name}
                      </MenuItem>
                    ))}
                  </CommandSelect>

                  {selectedShelter ? (
                    <Box className="shelter-capacity-summary">
                      <Stack direction="row" spacing={1} flexWrap="wrap">
                        <Chip
                          size="small"
                          label={selectedShelter.status}
                          color={
                            selectedShelter.status === "open"
                              ? "success"
                              : "warning"
                          }
                        />
                        <Chip
                          size="small"
                          variant="outlined"
                          label={`${selectedShelter.currentOccupancy}/${selectedShelter.capacity} occupied`}
                        />
                      </Stack>
                      <Typography variant="caption" color="text.secondary">
                        {selectedShelter.district} · {selectedShelter.address}
                      </Typography>
                    </Box>
                  ) : null}

                  <Grid container spacing={1}>
                    <Grid size={6}>
                      <TextField
                        label="Capacity"
                        size="small"
                        fullWidth
                        required
                        value={shelterForm.capacity}
                        onChange={updateShelterForm("capacity")}
                        inputProps={{ inputMode: "numeric" }}
                        error={
                          Boolean(shelterForm.capacity) &&
                          !Number.isFinite(Number(shelterForm.capacity))
                        }
                        helperText={
                          Boolean(shelterForm.capacity) &&
                          !Number.isFinite(Number(shelterForm.capacity))
                            ? "Capacity must be a number"
                            : ""
                        }
                      />
                    </Grid>
                    <Grid size={6}>
                      <TextField
                        label="Occupancy"
                        size="small"
                        fullWidth
                        required
                        value={shelterForm.currentOccupancy}
                        onChange={updateShelterForm("currentOccupancy")}
                        inputProps={{ inputMode: "numeric" }}
                        error={
                          Boolean(shelterForm.currentOccupancy) &&
                          !Number.isFinite(Number(shelterForm.currentOccupancy))
                        }
                        helperText={
                          Boolean(shelterForm.currentOccupancy) &&
                          !Number.isFinite(Number(shelterForm.currentOccupancy))
                            ? "Occupancy must be a number"
                            : ""
                        }
                      />
                    </Grid>
                    <Grid size={12}>
                      <CommandSelect
                        label="Status"
                        value={shelterForm.status}
                        onChange={updateShelterForm("status")}
                      >
                        <MenuItem value="open">Open</MenuItem>
                        <MenuItem value="full">Full</MenuItem>
                        <MenuItem value="closed">Closed</MenuItem>
                        <MenuItem value="unknown">Unknown</MenuItem>
                      </CommandSelect>
                    </Grid>
                    <Grid size={12}>
                      <TextField
                        label="Operational note"
                        size="small"
                        fullWidth
                        multiline
                        minRows={2}
                        value={shelterForm.notes}
                        onChange={updateShelterForm("notes")}
                      />
                    </Grid>
                  </Grid>

                  <Button
                    type="button"
                    variant="contained"
                    startIcon={<LifeBuoy size={17} />}
                    onClick={() => void updateShelterCapacity()}
                    disabled={shelterBusy}
                  >
                    {shelterBusy ? "Updating" : "Update capacity"}
                  </Button>
                </Stack>
              </Paper>

              <Paper className="surface relief-panel">
                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  spacing={1}
                  justifyContent="space-between"
                  alignItems={{ xs: "stretch", sm: "center" }}
                  className="section-heading"
                >
                  <Stack direction="row" spacing={1} alignItems="center">
                    <LifeBuoy size={21} color={nadaaBrand.colors.gold} />
                    <Box>
                      <Typography variant="h6">Relief distribution</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Publish points, stock and eligibility
                      </Typography>
                    </Box>
                  </Stack>
                  <Button
                    type="button"
                    variant="outlined"
                    size="small"
                    startIcon={
                      reliefLoadState === "loading" ? (
                        <Loader2 size={16} className="spin-icon" />
                      ) : (
                        <RefreshCw size={16} />
                      )
                    }
                    onClick={() => void refreshReliefPoints()}
                    disabled={reliefLoadState === "loading"}
                  >
                    Refresh
                  </Button>
                </Stack>

                {reliefFeedback ? (
                  <Alert
                    severity={
                      reliefLoadState === "ready" ? "success" : "warning"
                    }
                    className="feed-alert"
                  >
                    {reliefFeedback}
                  </Alert>
                ) : null}

                <Stack spacing={1.5}>
                  <CommandSelect
                    label="Relief point"
                    value={reliefForm.reliefPointId}
                    onChange={updateReliefForm("reliefPointId")}
                  >
                    <MenuItem value="__new__">New relief point</MenuItem>
                    {reliefPoints.map((point) => (
                      <MenuItem value={point.id} key={point.id}>
                        {point.name}
                      </MenuItem>
                    ))}
                  </CommandSelect>

                  {selectedReliefPoint &&
                  reliefForm.reliefPointId !== "__new__" ? (
                    <Box className="shelter-capacity-summary">
                      <Stack direction="row" spacing={1} flexWrap="wrap">
                        <Chip
                          size="small"
                          label={selectedReliefPoint.status}
                          color={
                            selectedReliefPoint.status === "open"
                              ? "success"
                              : "warning"
                          }
                        />
                        <Chip
                          size="small"
                          variant="outlined"
                          label={`${selectedReliefPoint.stockCategories.length} stock lines`}
                        />
                      </Stack>
                      <Typography variant="caption" color="text.secondary">
                        {selectedReliefPoint.district} ·{" "}
                        {selectedReliefPoint.address}
                      </Typography>
                    </Box>
                  ) : null}

                  <Grid container spacing={1}>
                    <Grid size={{ xs: 12, sm: 7 }}>
                      <TextField
                        label="Name"
                        size="small"
                        fullWidth
                        required
                        value={reliefForm.name}
                        onChange={updateReliefForm("name")}
                      />
                    </Grid>
                    <Grid size={{ xs: 6, sm: 5 }}>
                      <CommandSelect
                        label="Type"
                        value={reliefForm.type}
                        onChange={updateReliefForm("type")}
                      >
                        <MenuItem value="food">Food</MenuItem>
                        <MenuItem value="water">Water</MenuItem>
                        <MenuItem value="medical">Medical</MenuItem>
                        <MenuItem value="hygiene">Hygiene</MenuItem>
                        <MenuItem value="blankets">Blankets</MenuItem>
                        <MenuItem value="cash">Cash</MenuItem>
                        <MenuItem value="mixed">Mixed</MenuItem>
                      </CommandSelect>
                    </Grid>
                    <Grid size={{ xs: 6, sm: 4 }}>
                      <CommandSelect
                        label="Status"
                        value={reliefForm.status}
                        onChange={updateReliefForm("status")}
                      >
                        <MenuItem value="open">Open</MenuItem>
                        <MenuItem value="limited">Limited</MenuItem>
                        <MenuItem value="paused">Paused</MenuItem>
                        <MenuItem value="closed">Closed</MenuItem>
                      </CommandSelect>
                    </Grid>
                    <Grid size={{ xs: 12, sm: 8 }}>
                      <TextField
                        label="Address"
                        size="small"
                        fullWidth
                        value={reliefForm.address}
                        onChange={updateReliefForm("address")}
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="Latitude"
                        size="small"
                        fullWidth
                        required
                        value={reliefForm.latitude}
                        onChange={updateReliefForm("latitude")}
                        inputProps={{ inputMode: "decimal" }}
                        error={
                          Boolean(reliefForm.latitude) &&
                          !Number.isFinite(Number(reliefForm.latitude))
                        }
                        helperText={
                          Boolean(reliefForm.latitude) &&
                          !Number.isFinite(Number(reliefForm.latitude))
                            ? "Latitude must be a number"
                            : ""
                        }
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="Longitude"
                        size="small"
                        fullWidth
                        required
                        value={reliefForm.longitude}
                        onChange={updateReliefForm("longitude")}
                        inputProps={{ inputMode: "decimal" }}
                        error={
                          Boolean(reliefForm.longitude) &&
                          !Number.isFinite(Number(reliefForm.longitude))
                        }
                        helperText={
                          Boolean(reliefForm.longitude) &&
                          !Number.isFinite(Number(reliefForm.longitude))
                            ? "Longitude must be a number"
                            : ""
                        }
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="District"
                        size="small"
                        fullWidth
                        value={reliefForm.district}
                        onChange={updateReliefForm("district")}
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="Contact"
                        size="small"
                        fullWidth
                        value={reliefForm.contact}
                        onChange={updateReliefForm("contact")}
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="Hours"
                        size="small"
                        fullWidth
                        value={reliefForm.operatingHours}
                        onChange={updateReliefForm("operatingHours")}
                      />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                      <TextField
                        label="Schedule"
                        size="small"
                        fullWidth
                        value={reliefForm.schedule}
                        onChange={updateReliefForm("schedule")}
                      />
                    </Grid>
                    <Grid size={12}>
                      <TextField
                        label="Eligibility"
                        size="small"
                        fullWidth
                        multiline
                        minRows={2}
                        value={reliefForm.eligibility}
                        onChange={updateReliefForm("eligibility")}
                      />
                    </Grid>
                    <Grid size={12}>
                      <TextField
                        label="Stock lines"
                        size="small"
                        fullWidth
                        multiline
                        minRows={3}
                        value={reliefForm.stockCategories}
                        onChange={updateReliefForm("stockCategories")}
                      />
                    </Grid>
                  </Grid>

                  {reliefHistory.length ? (
                    <Box className="relief-history">
                      <Typography variant="caption" color="text.secondary">
                        Recent stock history
                      </Typography>
                      {reliefHistory.slice(0, 2).map((entry) => (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          key={entry.id}
                        >
                          {formatShortDate(entry.changedAt)} ·{" "}
                          {entry.stockCategories.length} stock lines ·{" "}
                          {entry.changedBy}
                        </Typography>
                      ))}
                    </Box>
                  ) : null}

                  <Button
                    type="button"
                    variant="contained"
                    startIcon={<LifeBuoy size={17} />}
                    onClick={() => void saveReliefPoint()}
                    disabled={reliefBusy}
                  >
                    {reliefBusy
                      ? "Saving"
                      : reliefForm.reliefPointId === "__new__"
                        ? "Publish relief point"
                        : "Update relief point"}
                  </Button>
                </Stack>
              </Paper>

              <RoutePlannerPanel
                selectedIncident={
                  selectedIncident
                    ? {
                        id: selectedIncident.id,
                        reference: selectedIncident.reference,
                        location: selectedIncident.location,
                      }
                    : undefined
                }
              />

              <FloodSimulationPanel />

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
                    label="Authority session"
                    value={authoritySession.agency}
                    color="success"
                  />
                  <StatusLine
                    label="ML alerts"
                    value="Human review"
                    color="warning"
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

function buildDefaultShelterForm(shelter: ShelterRecord): ShelterFormState {
  return {
    shelterId: shelter.id,
    capacity: `${shelter.capacity}`,
    currentOccupancy: `${shelter.currentOccupancy}`,
    status: shelter.status,
    notes: shelter.notes ?? "",
  };
}

function buildDefaultReliefPointForm(
  point?: ReliefPointRecord,
): ReliefPointFormState {
  return {
    reliefPointId: point?.id ?? "__new__",
    name: point?.name ?? "",
    type: point?.type ?? "food",
    status: point?.status ?? "open",
    region: point?.region ?? "Greater Accra",
    district: point?.district ?? "",
    address: point?.address ?? "",
    latitude: point ? `${point.location.lat}` : "5.5600",
    longitude: point ? `${point.location.lng}` : "-0.2000",
    contact: point?.contact ?? "112",
    operatingHours: point?.operatingHours ?? "08:00-18:00",
    eligibility: point?.eligibility ?? "",
    schedule: point?.schedule ?? "Daily",
    stockCategories: formatReliefStockLines(point?.stockCategories ?? []),
    sourceRef: point?.sourceRef ?? "authority-dashboard",
  };
}

function formatReliefStockLines(categories: ReliefStockCategory[]) {
  return categories
    .map((category) =>
      [category.category, category.quantity, category.unit].join(", "),
    )
    .join("\n");
}

function parseReliefStockCategories(value: string): ReliefStockCategory[] {
  const now = new Date().toISOString();
  const categories = value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [category, quantityText, unit = "units"] = line
        .split(",")
        .map((part) => part.trim());
      const quantity = Number(quantityText);
      if (!category || !Number.isFinite(quantity) || quantity < 0) {
        throw new Error(
          "Stock lines must use category, quantity, unit per line.",
        );
      }
      return {
        category,
        quantity,
        unit: unit || "units",
        lastUpdated: now,
      };
    });

  if (!categories.length) {
    throw new Error("At least one stock line is required.");
  }
  return categories;
}

function formatShortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-GH", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export default CommandCenterApp;
