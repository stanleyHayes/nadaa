import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  Alert,
  AppBar,
  Box,
  Button,
  Chip,
  Container,
  CssBaseline,
  Divider,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  ThemeProvider,
  TextField,
  Toolbar,
  Typography,
  createTheme,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
  AlertTriangle,
  BellRing,
  CheckCheck,
  Crosshair,
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
  AgencyUserRole,
  AlertListResponse,
  AlertSeverity,
  AlertStatus,
  AuthorityAlertRecord,
  CreateAlertRequest,
  HazardType,
  IncidentStatusUpdateRequest,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatus,
  RiskLevel,
} from "@nadaa/shared-types";

const INCIDENT_API_BASE =
  import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
const ALERT_API_BASE =
  import.meta.env.VITE_ALERT_API_URL ?? "http://localhost:8089/api/v1";

const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer",
];

const authoritySession = {
  id: "usr_nadmo_accra",
  name: "NADMO Officer",
  role: "nadmo_officer" as AgencyUserRole,
  agencyId: "00000000-0000-0000-0000-000000000101",
  agency: "NADMO Accra Metro",
  mfaEnabled: true,
};

const theme = createTheme({
  palette: {
    primary: { main: nadaaBrand.colors.navy },
    secondary: { main: nadaaBrand.colors.green },
    error: { main: nadaaBrand.colors.red },
    warning: { main: nadaaBrand.colors.gold },
    background: { default: "#F3F6FA", paper: "#FFFFFF" },
    text: {
      primary: nadaaBrand.colors.ink,
      secondary: nadaaBrand.colors.slate,
    },
  },
  shape: { borderRadius: 8 },
  typography: {
    fontFamily:
      'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    h1: { fontWeight: 800 },
    h2: { fontWeight: 800 },
    h3: { fontWeight: 800 },
    h4: { fontWeight: 800 },
    h5: { fontWeight: 800 },
    h6: { fontWeight: 800 },
    button: { textTransform: "none", fontWeight: 800 },
  },
});

type CommandIncident = IncidentRecord & {
  region: string;
  district: string;
  locality: string;
  assignedAgency: string;
  responderEta: string;
  timeline: string[];
  source: "api" | "fixture";
};

type FilterState = {
  hazard: "all" | HazardType;
  regionDistrict: "all" | string;
  severity: "all" | RiskLevel;
  status: "all" | IncidentStatus;
  time: "all" | "1h" | "6h" | "24h";
};

type IncidentLoadState = "loading" | "ready" | "fallback" | "empty" | "error";
type AlertLoadState = "loading" | "ready" | "fallback" | "error";

type AlertFormState = {
  title: string;
  severity: AlertSeverity;
  message: string;
  targetLabel: string;
  startsAt: string;
  expiresAt: string;
  recommendedAction: string;
  evacuationRequired: boolean;
  shelterIds: string;
};

type IncidentStatusFormState = {
  status: IncidentStatus;
  note: string;
  resolutionNotes: string;
};

const fallbackIncidents: CommandIncident[] = [
  {
    id: "inc_accra_flood_0241",
    reference: "INC-0241",
    type: "flood",
    severity: "severe",
    status: "under_review",
    description:
      "Water is rising near a low-lying road and vehicles are trapped.",
    location: { lat: 5.579, lng: -0.212 },
    peopleAffected: 28,
    injuriesReported: false,
    urgency: "life_threatening",
    anonymous: false,
    contactPermission: true,
    media: ["media_flood_photo_001"],
    priorityReview: true,
    duplicateCandidates: [
      {
        incidentId: "inc_accra_flood_0237",
        reference: "INC-0237",
        score: 0.82,
        distanceMeters: 214,
        minutesApart: 16,
        reasons: ["same_hazard", "nearby_location", "recent_report"],
      },
    ],
    reportedBy: { userId: "usr_ama", phone: "+233200000003" },
    createdAt: "2026-07-06T18:42:00Z",
    updatedAt: "2026-07-06T18:48:00Z",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Accra Central",
    assignedAgency: "NADMO AMA",
    responderEta: "7 min",
    timeline: [
      "Citizen report received with photo evidence",
      "Duplicate reports grouped near Accra Central",
      "NADMO AMA dispatcher reviewing severity",
    ],
    source: "fixture",
  },
  {
    id: "inc_tema_crash_0239",
    reference: "INC-0239",
    type: "road_crash",
    severity: "high",
    status: "assigned",
    description: "Three-vehicle crash blocking the Tema motorway shoulder.",
    location: { lat: 5.642, lng: -0.028 },
    peopleAffected: 11,
    injuriesReported: true,
    urgency: "high",
    anonymous: false,
    contactPermission: true,
    media: ["media_crash_photo_002"],
    priorityReview: true,
    duplicateCandidates: [],
    createdAt: "2026-07-06T18:25:00Z",
    updatedAt: "2026-07-06T18:39:00Z",
    region: "Greater Accra",
    district: "Tema Metropolitan",
    locality: "Tema Motorway",
    assignedAgency: "Ambulance + Police",
    responderEta: "12 min",
    timeline: [
      "Dispatcher verified multiple injured persons",
      "Ambulance and police units assigned",
      "Motorway patrol requested lane control",
    ],
    source: "fixture",
  },
  {
    id: "inc_ablekuma_drain_0236",
    reference: "INC-0236",
    type: "blocked_drain",
    severity: "moderate",
    status: "verified",
    description: "Blocked drain backing water into a residential street.",
    location: { lat: 5.601, lng: -0.286 },
    peopleAffected: 14,
    injuriesReported: false,
    urgency: "moderate",
    anonymous: true,
    contactPermission: false,
    media: [],
    priorityReview: false,
    duplicateCandidates: [],
    createdAt: "2026-07-06T17:58:00Z",
    updatedAt: "2026-07-06T18:12:00Z",
    region: "Greater Accra",
    district: "Ablekuma West",
    locality: "Dansoman",
    assignedAgency: "District Assembly",
    responderEta: "31 min",
    timeline: [
      "District officer verified blocked drain",
      "Sanitation crew notified",
      "Resident contact hidden due anonymous report",
    ],
    source: "fixture",
  },
  {
    id: "inc_korle_fire_0232",
    reference: "INC-0232",
    type: "fire",
    severity: "high",
    status: "response_en_route",
    description: "Electrical fire reported behind a market stall.",
    location: { lat: 5.544, lng: -0.213 },
    peopleAffected: 8,
    injuriesReported: false,
    urgency: "high",
    anonymous: false,
    contactPermission: true,
    media: ["media_fire_photo_003"],
    priorityReview: true,
    duplicateCandidates: [],
    createdAt: "2026-07-06T17:41:00Z",
    updatedAt: "2026-07-06T18:19:00Z",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Korle Gonno",
    assignedAgency: "Ghana National Fire Service",
    responderEta: "5 min",
    timeline: [
      "Fire service call confirmed smoke visible",
      "Hydrant access checked by dispatcher",
      "Engine crew en route",
    ],
    source: "fixture",
  },
];

const fallbackAlerts: AuthorityAlertRecord[] = [
  {
    id: "alert_fixture_submitted",
    title: "Accra flood watch",
    hazardType: "flood",
    severity: "warning",
    message: "Heavy rainfall may cause flooding in low-lying communities.",
    target: {
      type: "district",
      ids: ["accra-metropolitan"],
      label: "Accra Metropolitan",
    },
    startsAt: "2026-07-06T19:30:00Z",
    expiresAt: "2026-07-07T07:00:00Z",
    recommendedAction:
      "Avoid flooded roads and prepare to move to higher ground.",
    evacuationRequired: false,
    shelterIds: ["00000000-0000-0000-0000-000000000301"],
    issuingAgencyId: "00000000-0000-0000-0000-000000000101",
    issuedBy: "usr_dispatcher_fixture",
    status: "submitted",
    emergencyOverride: false,
    createdAt: "2026-07-06T18:15:00Z",
    updatedAt: "2026-07-06T18:45:00Z",
    submittedAt: "2026-07-06T18:45:00Z",
  },
];

const defaultFilters: FilterState = {
  hazard: "all",
  regionDistrict: "all",
  severity: "all",
  status: "all",
  time: "all",
};

const severityOrder: Record<RiskLevel, number> = {
  emergency: 5,
  severe: 4,
  high: 3,
  moderate: 2,
  low: 1,
};

const severityColors: Record<RiskLevel, string> = {
  emergency: "#7F1D1D",
  severe: nadaaBrand.colors.red,
  high: "#D97706",
  moderate: nadaaBrand.colors.gold,
  low: nadaaBrand.colors.green,
};

const alertSeverityOptions: AlertSeverity[] = [
  "advisory",
  "watch",
  "warning",
  "severe_warning",
  "emergency",
];

const incidentStatusOptions: IncidentStatus[] = [
  "reported",
  "under_review",
  "verified",
  "assigned",
  "response_en_route",
  "on_scene",
  "contained",
  "recovery_ongoing",
  "closed",
  "false_report",
];

const incidentTransitionOptions: Record<IncidentStatus, IncidentStatus[]> = {
  reported: ["under_review", "verified", "false_report"],
  under_review: ["verified", "false_report"],
  verified: ["assigned", "response_en_route", "false_report"],
  assigned: [
    "response_en_route",
    "on_scene",
    "contained",
    "recovery_ongoing",
    "closed",
    "false_report",
  ],
  response_en_route: [
    "on_scene",
    "contained",
    "recovery_ongoing",
    "closed",
    "false_report",
  ],
  on_scene: ["contained", "recovery_ongoing", "closed", "false_report"],
  contained: ["recovery_ongoing", "closed", "false_report"],
  recovery_ongoing: ["closed", "false_report"],
  closed: [],
  false_report: [],
};

function App() {
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
  const [alerts, setAlerts] = useState<AuthorityAlertRecord[]>(fallbackAlerts);
  const [alertLoadState, setAlertLoadState] =
    useState<AlertLoadState>("loading");
  const [alertMessage, setAlertMessage] = useState("Loading alert workflow");
  const [alertBusy, setAlertBusy] = useState(false);
  const [alertFeedback, setAlertFeedback] = useState("");
  const [alertForm, setAlertForm] = useState<AlertFormState>(
    buildDefaultAlertForm(fallbackIncidents[0]),
  );

  const authorityHeaders = () => ({
    "Content-Type": "application/json",
    "X-NADAA-Actor-ID": authoritySession.id,
    "X-NADAA-Actor-Role": authoritySession.role,
    "X-NADAA-Agency-ID": authoritySession.agencyId,
    "X-NADAA-MFA-Completed": authoritySession.mfaEnabled ? "true" : "false",
    "X-NADAA-Request-ID": `authority-ui-${Date.now()}`,
  });

  const refreshIncidents = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading incident feed");

    try {
      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
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
    setStatusFeedback("");
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

  const applyIncidentUpdate = (incident: IncidentRecord) => {
    const enriched = enrichIncidentFromAPI(incident);
    setIncidents((current) => {
      const exists = current.some((item) => item.id === incident.id);
      if (!exists) {
        return [enriched, ...current];
      }
      return current.map((item) => (item.id === incident.id ? enriched : item));
    });
    setSelectedIncidentId(incident.id);
    setStatusForm(buildDefaultStatusForm(enriched));
    setLoadState("ready");
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

  const buildAlertRequest = (): CreateAlertRequest => ({
    title: alertForm.title,
    hazardType: selectedIncident?.type ?? "flood",
    severity: alertForm.severity,
    message: alertForm.message,
    target: {
      type: "district",
      ids: selectedIncident
        ? [districtSlug(selectedIncident.district)]
        : ["accra-metropolitan"],
      label: alertForm.targetLabel,
    },
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
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Container maxWidth="sm" className="access-shell">
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
    <ThemeProvider theme={theme}>
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
                busy={statusBusy}
                feedback={statusFeedback}
                form={statusForm}
                incident={selectedIncident}
                onUpdateForm={updateStatusForm}
                onUpdateStatus={updateIncidentStatus}
                onVerify={verifySelectedIncident}
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

function CommandSelect({
  children,
  label,
  onChange,
  value,
}: {
  children: React.ReactNode;
  label: string;
  onChange: (event: SelectChangeEvent) => void;
  value: string;
}) {
  return (
    <FormControl fullWidth size="small">
      <InputLabel>{label}</InputLabel>
      <Select label={label} value={value} onChange={onChange}>
        {children}
      </Select>
    </FormControl>
  );
}

function IncidentMap({
  incidents,
  onSelect,
  selectedIncidentId,
}: {
  incidents: CommandIncident[];
  onSelect: (incidentId: string) => void;
  selectedIncidentId?: string;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [5.586, -0.18],
      zoom: 11,
      zoomControl: true,
      scrollWheelZoom: false,
    });

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
      maxZoom: 19,
    }).addTo(map);

    mapRef.current = map;
    layerRef.current = L.layerGroup().addTo(map);

    return () => {
      map.remove();
      mapRef.current = null;
      layerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const layer = layerRef.current;
    const map = mapRef.current;
    if (!layer || !map) {
      return;
    }

    layer.clearLayers();
    if (!incidents.length) {
      return;
    }

    const bounds = L.latLngBounds([]);
    incidents.forEach((incident) => {
      const isSelected = incident.id === selectedIncidentId;
      const marker = L.circleMarker(
        [incident.location.lat, incident.location.lng],
        {
          radius: isSelected ? 13 : 9,
          color: "#FFFFFF",
          weight: isSelected ? 4 : 2,
          fillColor: severityColors[incident.severity],
          fillOpacity: isSelected ? 0.95 : 0.78,
        },
      );
      marker.bindPopup(
        `<strong>${incident.reference}</strong><br>${hazardLabel(incident.type)} · ${severityLabel(incident.severity)}<br>${incident.locality}`,
      );
      marker.on("click", () => onSelect(incident.id));
      marker.addTo(layer);
      bounds.extend([incident.location.lat, incident.location.lng]);
    });

    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 13 });
    }
  }, [incidents, onSelect, selectedIncidentId]);

  useEffect(() => {
    const map = mapRef.current;
    const selected = incidents.find(
      (incident) => incident.id === selectedIncidentId,
    );
    if (!map || !selected) {
      return;
    }
    map.flyTo(
      [selected.location.lat, selected.location.lng],
      Math.max(map.getZoom(), 12),
      {
        animate: true,
        duration: 0.45,
      },
    );
  }, [incidents, selectedIncidentId]);

  return (
    <Box className="map-frame">
      <Box ref={containerRef} className="leaflet-command-map" />
      {!incidents.length ? (
        <Box className="map-empty">
          <EmptyState
            title="No map markers"
            detail="No incidents match the current command filters."
          />
        </Box>
      ) : null}
    </Box>
  );
}

function AlertWorkflowPanel({
  alerts,
  busy,
  feedback,
  form,
  loadState,
  onCreateDraft,
  onRunAction,
  onUpdateForm,
  selectedIncident,
}: {
  alerts: AuthorityAlertRecord[];
  busy: boolean;
  feedback: string;
  form: AlertFormState;
  loadState: AlertLoadState;
  onCreateDraft: () => void;
  onRunAction: (
    alert: AuthorityAlertRecord,
    action: "submit" | "approve" | "reject" | "emergency-override",
  ) => void;
  onUpdateForm: (
    key: keyof AlertFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  selectedIncident?: CommandIncident;
}) {
  const queueAlerts = alerts.filter(
    (alert) => alert.status !== "published" && alert.status !== "expired",
  );

  return (
    <Paper className="surface alert-panel">
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        className="section-heading"
      >
        <BellRing size={21} color={nadaaBrand.colors.red} />
        <Box>
          <Typography variant="h6">Alert workflow</Typography>
          <Typography variant="caption" color="text.secondary">
            Draft, submit, approve, reject, or override with audit.
          </Typography>
        </Box>
      </Stack>

      {selectedIncident ? (
        <Stack spacing={1.5}>
          <Box>
            <Stack direction="row" justifyContent="space-between" gap={1}>
              <Typography variant="subtitle2">
                Draft from {selectedIncident.reference}
              </Typography>
              <Chip
                size="small"
                label={alertSeverityLabel(form.severity)}
                color={form.severity === "emergency" ? "error" : "warning"}
              />
            </Stack>
            <Typography variant="body2" color="text.secondary">
              {hazardLabel(selectedIncident.type)} · {selectedIncident.district}
            </Typography>
          </Box>

          <TextField
            size="small"
            label="Title"
            value={form.title}
            onChange={onUpdateForm("title")}
          />
          <TextField
            size="small"
            label="Message"
            value={form.message}
            onChange={onUpdateForm("message")}
            multiline
            minRows={3}
          />

          <Grid container spacing={1.25}>
            <Grid size={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Severity</InputLabel>
                <Select
                  label="Severity"
                  value={form.severity}
                  onChange={onUpdateForm("severity")}
                >
                  {alertSeverityOptions.map((severity) => (
                    <MenuItem key={severity} value={severity}>
                      {alertSeverityLabel(severity)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={6}>
              <TextField
                size="small"
                label="Target"
                value={form.targetLabel}
                onChange={onUpdateForm("targetLabel")}
                fullWidth
              />
            </Grid>
            <Grid size={6}>
              <TextField
                size="small"
                label="Starts"
                value={form.startsAt}
                onChange={onUpdateForm("startsAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
            <Grid size={6}>
              <TextField
                size="small"
                label="Expires"
                value={form.expiresAt}
                onChange={onUpdateForm("expiresAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
          </Grid>

          <TextField
            size="small"
            label="Recommended action"
            value={form.recommendedAction}
            onChange={onUpdateForm("recommendedAction")}
          />
          <TextField
            size="small"
            label="Shelter IDs"
            value={form.shelterIds}
            onChange={onUpdateForm("shelterIds")}
          />
          <FormControlLabel
            control={
              <Switch
                checked={form.evacuationRequired}
                onChange={onUpdateForm("evacuationRequired")}
              />
            }
            label="Evacuation required"
          />
          <Button
            variant="contained"
            color="error"
            startIcon={<BellRing size={17} />}
            disabled={busy}
            onClick={onCreateDraft}
          >
            Create draft
          </Button>
        </Stack>
      ) : (
        <EmptyState
          title="No incident selected"
          detail="Choose an incident before drafting an alert."
        />
      )}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          gap={1}
        >
          <Typography variant="subtitle2">Approval queue</Typography>
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live"
                : loadState === "loading"
                  ? "Loading"
                  : "Fixture"
            }
            color={loadState === "ready" ? "success" : "warning"}
          />
        </Stack>
        {feedback ? (
          <Alert
            severity={
              loadState === "ready"
                ? "success"
                : loadState === "loading"
                  ? "info"
                  : "warning"
            }
          >
            {feedback}
          </Alert>
        ) : null}
        {queueAlerts.length ? (
          queueAlerts.slice(0, 4).map((alert) => (
            <Box key={alert.id} className="alert-queue-row">
              <Stack direction="row" justifyContent="space-between" gap={1}>
                <Box>
                  <Typography variant="subtitle2">{alert.title}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {alert.target.label} · {alertSeverityLabel(alert.severity)}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={alertStatusLabel(alert.status)}
                  color={alertStatusColor(alert.status)}
                />
              </Stack>
              <Stack
                direction="row"
                spacing={1}
                flexWrap="wrap"
                className="alert-actions"
              >
                {alert.status === "draft" ? (
                  <Button
                    size="small"
                    variant="outlined"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "submit")}
                  >
                    Submit
                  </Button>
                ) : null}
                {alert.status === "submitted" ? (
                  <>
                    <Button
                      size="small"
                      variant="contained"
                      color="success"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "approve")}
                    >
                      Approve
                    </Button>
                    <Button
                      size="small"
                      variant="outlined"
                      color="error"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "reject")}
                    >
                      Reject
                    </Button>
                  </>
                ) : null}
                {alert.status === "draft" ||
                alert.status === "submitted" ||
                alert.status === "rejected" ? (
                  <Button
                    size="small"
                    color="error"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "emergency-override")}
                  >
                    Override
                  </Button>
                ) : null}
              </Stack>
            </Box>
          ))
        ) : (
          <EmptyState
            title="No alerts in queue"
            detail="Create a draft to begin the approval workflow."
          />
        )}
      </Stack>
    </Paper>
  );
}

function IncidentDetailPanel({
  busy,
  feedback,
  form,
  incident,
  onUpdateForm,
  onUpdateStatus,
  onVerify,
}: {
  busy: boolean;
  feedback: string;
  form: IncidentStatusFormState;
  incident?: CommandIncident;
  onUpdateForm: (
    key: keyof IncidentStatusFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateStatus: () => void;
  onVerify: () => void;
}) {
  if (!incident) {
    return (
      <Paper className="surface">
        <EmptyState
          title="No incident selected"
          detail="Choose a map marker or queue row to inspect the incident."
        />
      </Paper>
    );
  }

  const nextStatuses = incidentTransitionOptions[incident.status];
  const terminal = nextStatuses.length === 0;
  const resolutionRequired = requiresIncidentResolution(form.status);
  const canVerify = nextStatuses.includes("verified");

  return (
    <Paper className="surface detail-panel">
      <Stack
        direction="row"
        justifyContent="space-between"
        gap={1}
        className="section-heading"
      >
        <Box>
          <Typography variant="overline" color="secondary">
            Selected incident
          </Typography>
          <Typography variant="h6">{incident.reference}</Typography>
        </Box>
        <Chip
          size="small"
          label={severityLabel(incident.severity)}
          style={{
            backgroundColor: severityColors[incident.severity],
            color: "#FFFFFF",
          }}
        />
      </Stack>

      <Typography variant="body2" color="text.secondary">
        {incident.description}
      </Typography>

      <Divider className="detail-divider" />

      <Grid container spacing={1.5}>
        <Grid size={6}>
          <Fact label="Hazard" value={hazardLabel(incident.type)} />
        </Grid>
        <Grid size={6}>
          <Fact label="Status" value={statusLabel(incident.status)} />
        </Grid>
        <Grid size={6}>
          <Fact label="People" value={`${incident.peopleAffected}`} />
        </Grid>
        <Grid size={6}>
          <Fact label="Responder ETA" value={incident.responderEta} />
        </Grid>
        <Grid size={12}>
          <Fact label="Assigned agency" value={incident.assignedAgency} />
        </Grid>
      </Grid>

      <Divider className="detail-divider" />

      <Stack spacing={1}>
        <Typography variant="subtitle2">Response timeline</Typography>
        {incident.timeline.map((event) => (
          <Box className="timeline-row" key={event}>
            <Typography variant="body2">{event}</Typography>
          </Box>
        ))}
      </Stack>

      {incident.duplicateCandidates.length ? (
        <>
          <Divider className="detail-divider" />
          <Alert severity="warning">
            {incident.duplicateCandidates.length} possible duplicate
            {incident.duplicateCandidates.length > 1 ? "s" : ""} need review.
          </Alert>
        </>
      ) : null}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Status workflow</Typography>
            <Typography variant="caption" color="text.secondary">
              {terminal
                ? "Terminal incident state"
                : "Audited authority action"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={incident.source === "api" ? "Live" : "Fixture"}
            color={incident.source === "api" ? "success" : "warning"}
          />
        </Stack>

        {feedback ? (
          <Alert
            severity={
              feedback.includes("needs") || feedback.includes("valid")
                ? "warning"
                : "success"
            }
          >
            {feedback}
          </Alert>
        ) : null}

        <FormControl fullWidth size="small" disabled={terminal}>
          <InputLabel>Next status</InputLabel>
          <Select
            label="Next status"
            value={form.status}
            onChange={onUpdateForm("status")}
          >
            {(nextStatuses.length ? nextStatuses : [incident.status]).map(
              (status) => (
                <MenuItem value={status} key={status}>
                  {statusLabel(status)}
                </MenuItem>
              ),
            )}
          </Select>
        </FormControl>

        <TextField
          size="small"
          label="Status note"
          value={form.note}
          onChange={onUpdateForm("note")}
          multiline
          minRows={2}
        />

        {resolutionRequired ? (
          <TextField
            size="small"
            label="Resolution notes"
            value={form.resolutionNotes}
            onChange={onUpdateForm("resolutionNotes")}
            multiline
            minRows={3}
          />
        ) : null}
      </Stack>

      <Stack direction="row" spacing={1} className="detail-actions">
        <Button
          variant="contained"
          disabled={busy || !canVerify}
          onClick={onVerify}
          startIcon={<CheckCheck size={17} />}
        >
          Verify
        </Button>
        <Button
          variant="outlined"
          disabled={
            busy ||
            terminal ||
            (resolutionRequired && !form.resolutionNotes.trim())
          }
          onClick={onUpdateStatus}
          startIcon={<Truck size={17} />}
        >
          Update status
        </Button>
      </Stack>
    </Paper>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <Box className="fact">
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="subtitle2">{value}</Typography>
    </Box>
  );
}

function StatusLine({
  color,
  label,
  value,
}: {
  color: "success" | "warning" | "default";
  label: string;
  value: string;
}) {
  return (
    <Stack
      direction="row"
      justifyContent="space-between"
      alignItems="center"
      gap={1}
    >
      <Typography variant="body2">{label}</Typography>
      <Chip size="small" label={value} color={color} />
    </Stack>
  );
}

function EmptyState({ detail, title }: { detail: string; title: string }) {
  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      spacing={1}
      className="empty-state"
    >
      <Crosshair size={28} color={nadaaBrand.colors.slate} />
      <Typography variant="subtitle2">{title}</Typography>
      <Typography variant="body2" color="text.secondary" textAlign="center">
        {detail}
      </Typography>
    </Stack>
  );
}

function buildQueueMetrics(incidents: CommandIncident[]) {
  return [
    {
      label: "New reports",
      value: incidents.filter(
        (incident) =>
          incident.status === "reported" || incident.status === "under_review",
      ).length,
      icon: ShieldAlert,
      color: nadaaBrand.colors.red,
    },
    {
      label: "Verified",
      value: incidents.filter(
        (incident) =>
          incident.status === "verified" || incident.status === "assigned",
      ).length,
      icon: CheckCheck,
      color: nadaaBrand.colors.green,
    },
    {
      label: "Teams en route",
      value: incidents.filter(
        (incident) =>
          incident.status === "response_en_route" ||
          incident.status === "on_scene",
      ).length,
      icon: Truck,
      color: "#0B6FB8",
    },
    {
      label: "Priority review",
      value: incidents.filter((incident) => incident.priorityReview).length,
      icon: AlertTriangle,
      color: nadaaBrand.colors.gold,
    },
  ];
}

function buildFilterOptions(incidents: CommandIncident[]) {
  return {
    hazards: uniqueSorted(incidents.map((incident) => incident.type)),
    regionDistricts: uniqueSorted(
      incidents.map((incident) => `${incident.region} / ${incident.district}`),
    ),
    severities: uniqueSorted(
      incidents.map((incident) => incident.severity),
    ).sort((a, b) => severityOrder[b] - severityOrder[a]),
    statuses: uniqueSorted(incidents.map((incident) => incident.status)),
  };
}

function matchesFilters(incident: CommandIncident, filters: FilterState) {
  if (filters.hazard !== "all" && incident.type !== filters.hazard) {
    return false;
  }
  if (
    filters.regionDistrict !== "all" &&
    `${incident.region} / ${incident.district}` !== filters.regionDistrict
  ) {
    return false;
  }
  if (filters.severity !== "all" && incident.severity !== filters.severity) {
    return false;
  }
  if (filters.status !== "all" && incident.status !== filters.status) {
    return false;
  }
  if (
    filters.time !== "all" &&
    !withinTimeWindow(incident.createdAt, filters.time)
  ) {
    return false;
  }
  return true;
}

function withinTimeWindow(createdAt: string, timeFilter: FilterState["time"]) {
  const hours =
    timeFilter === "1h"
      ? 1
      : timeFilter === "6h"
        ? 6
        : timeFilter === "24h"
          ? 24
          : 0;
  if (!hours) {
    return true;
  }
  const incidentTime = new Date(createdAt).getTime();
  const latestFixtureTime = new Date("2026-07-06T19:00:00Z").getTime();
  return latestFixtureTime - incidentTime <= hours * 60 * 60 * 1000;
}

function enrichIncidentFromAPI(incident: IncidentRecord): CommandIncident {
  const district = districtFromCoordinates(incident.location);
  return {
    ...incident,
    region: district.region,
    district: district.district,
    locality: district.locality,
    assignedAgency: assignmentForIncident(incident),
    responderEta: etaForIncident(incident),
    timeline: [
      `${hazardLabel(incident.type)} report received from incident service`,
      `${statusLabel(incident.status)} status synchronized`,
      incident.verifiedAt
        ? `Verified by ${incident.verifiedBy || "authority user"}`
        : "",
      incident.statusReason ? `Latest note: ${incident.statusReason}` : "",
      incident.resolutionNotes ? `Resolution: ${incident.resolutionNotes}` : "",
      incident.priorityReview
        ? "Priority review flag is active"
        : "Dispatcher monitoring normal queue",
    ].filter(Boolean),
    source: "api",
  };
}

function districtFromCoordinates(location: { lat: number; lng: number }) {
  if (location.lng > -0.08) {
    return {
      region: "Greater Accra",
      district: "Tema Metropolitan",
      locality: "Tema",
    };
  }
  if (location.lng < -0.25) {
    return {
      region: "Greater Accra",
      district: "Ablekuma West",
      locality: "Ablekuma",
    };
  }
  if (location.lat < 5.56) {
    return {
      region: "Greater Accra",
      district: "Accra Metropolitan",
      locality: "Korle Gonno",
    };
  }
  return {
    region: "Greater Accra",
    district: "Accra Metropolitan",
    locality: "Accra Central",
  };
}

function assignmentForIncident(incident: IncidentRecord) {
  if (incident.type === "fire") {
    return "Ghana National Fire Service";
  }
  if (incident.type === "road_crash" || incident.type === "medical_emergency") {
    return "Ambulance + Police";
  }
  if (incident.type === "blocked_drain") {
    return "District Assembly";
  }
  return "NADMO District Desk";
}

function etaForIncident(incident: IncidentRecord) {
  if (incident.severity === "emergency" || incident.severity === "severe") {
    return "5 min";
  }
  if (incident.priorityReview) {
    return "12 min";
  }
  return "30 min";
}

function alertReadiness(incident: CommandIncident) {
  const severityWeight = severityOrder[incident.severity] * 14;
  const duplicateWeight = incident.duplicateCandidates.length ? 12 : 0;
  const mediaWeight = incident.media.length ? 8 : 0;
  return Math.min(95, 30 + severityWeight + duplicateWeight + mediaWeight);
}

function buildDefaultStatusForm(
  incident?: CommandIncident,
): IncidentStatusFormState {
  const status = nextIncidentStatus(incident?.status ?? "reported");
  return {
    status,
    note: incident
      ? `${statusLabel(status)} update for ${incident.reference}.`
      : "Authority status update.",
    resolutionNotes: "",
  };
}

function nextIncidentStatus(status: IncidentStatus): IncidentStatus {
  return incidentTransitionOptions[status][0] ?? status;
}

function requiresIncidentResolution(status: IncidentStatus) {
  return status === "closed" || status === "false_report";
}

function buildDefaultAlertForm(incident?: CommandIncident): AlertFormState {
  const startsAt = new Date(Date.now() + 30 * 60 * 1000);
  const expiresAt = new Date(Date.now() + 12 * 60 * 60 * 1000);
  const hazard = incident ? hazardLabel(incident.type).toLowerCase() : "flood";
  const district = incident?.district ?? "Accra Metropolitan";
  const severity = riskToAlertSeverity(incident?.severity ?? "high");

  return {
    title: `${alertSeverityLabel(severity)} ${hazard} alert`,
    severity,
    message: incident
      ? `${incident.description} Avoid the affected area and follow official NADMO instructions.`
      : "Avoid low-lying roads and follow official NADMO instructions.",
    targetLabel: district,
    startsAt: formatDateTimeLocal(startsAt),
    expiresAt: formatDateTimeLocal(expiresAt),
    recommendedAction:
      incident?.severity === "emergency" || incident?.severity === "severe"
        ? "Prepare to evacuate if instructed by authorities."
        : "Stay alert, avoid the affected area, and monitor NADAA updates.",
    evacuationRequired: incident?.severity === "emergency",
    shelterIds: "00000000-0000-0000-0000-000000000301",
  };
}

function riskToAlertSeverity(severity: RiskLevel): AlertSeverity {
  if (severity === "emergency") {
    return "emergency";
  }
  if (severity === "severe") {
    return "severe_warning";
  }
  if (severity === "high") {
    return "warning";
  }
  if (severity === "moderate") {
    return "watch";
  }
  return "advisory";
}

function formatDateTimeLocal(date: Date) {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

function uniqueSorted<T extends string>(values: T[]) {
  return [...new Set(values)].sort((a, b) => a.localeCompare(b));
}

function hazardLabel(hazard: HazardType) {
  return hazard
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function severityLabel(severity: RiskLevel) {
  return severity[0].toUpperCase() + severity.slice(1);
}

function alertSeverityLabel(severity: AlertSeverity) {
  return severity
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function statusLabel(status: IncidentStatus) {
  return status
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function alertStatusLabel(status: AlertStatus) {
  return status
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function alertStatusColor(
  status: AlertStatus,
): "default" | "warning" | "success" | "error" {
  if (status === "approved" || status === "published") {
    return "success";
  }
  if (status === "submitted" || status === "draft") {
    return "warning";
  }
  if (status === "rejected" || status === "cancelled" || status === "expired") {
    return "error";
  }
  return "default";
}

function districtSlug(district: string) {
  return district
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

function formatIncidentAge(createdAt: string) {
  const latestFixtureTime = new Date("2026-07-06T19:00:00Z").getTime();
  const minutes = Math.max(
    1,
    Math.round((latestFixtureTime - new Date(createdAt).getTime()) / 60000),
  );
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  return `${hours} hr ${minutes % 60} min`;
}

export default App;
