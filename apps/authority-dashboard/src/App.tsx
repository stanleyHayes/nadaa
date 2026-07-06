import { useEffect, useMemo, useRef, useState } from "react";
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
  Grid,
  InputLabel,
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  ThemeProvider,
  Toolbar,
  Typography,
  createTheme
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
  Truck
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AgencyUserRole,
  HazardType,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatus,
  RiskLevel
} from "@nadaa/shared-types";

const INCIDENT_API_BASE = import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";

const commandRoles: AgencyUserRole[] = [
  "system_admin",
  "agency_admin",
  "nadmo_officer",
  "district_officer",
  "dispatcher",
  "responder",
  "agency_viewer"
];

const authoritySession = {
  name: "Accra Dispatcher",
  role: "dispatcher" as AgencyUserRole,
  agency: "NADMO Accra Metro",
  mfaEnabled: true
};

const theme = createTheme({
  palette: {
    primary: { main: nadaaBrand.colors.navy },
    secondary: { main: nadaaBrand.colors.green },
    error: { main: nadaaBrand.colors.red },
    warning: { main: nadaaBrand.colors.gold },
    background: { default: "#F3F6FA", paper: "#FFFFFF" },
    text: { primary: nadaaBrand.colors.ink, secondary: nadaaBrand.colors.slate }
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
    button: { textTransform: "none", fontWeight: 800 }
  }
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

const fallbackIncidents: CommandIncident[] = [
  {
    id: "inc_accra_flood_0241",
    reference: "INC-0241",
    type: "flood",
    severity: "severe",
    status: "under_review",
    description: "Water is rising near a low-lying road and vehicles are trapped.",
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
        reasons: ["same_hazard", "nearby_location", "recent_report"]
      }
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
      "NADMO AMA dispatcher reviewing severity"
    ],
    source: "fixture"
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
      "Motorway patrol requested lane control"
    ],
    source: "fixture"
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
      "Resident contact hidden due anonymous report"
    ],
    source: "fixture"
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
      "Engine crew en route"
    ],
    source: "fixture"
  }
];

const defaultFilters: FilterState = {
  hazard: "all",
  regionDistrict: "all",
  severity: "all",
  status: "all",
  time: "all"
};

const severityOrder: Record<RiskLevel, number> = {
  emergency: 5,
  severe: 4,
  high: 3,
  moderate: 2,
  low: 1
};

const severityColors: Record<RiskLevel, string> = {
  emergency: "#7F1D1D",
  severe: nadaaBrand.colors.red,
  high: "#D97706",
  moderate: nadaaBrand.colors.gold,
  low: nadaaBrand.colors.green
};

function App() {
  const hasCommandAccess = commandRoles.includes(authoritySession.role) && authoritySession.mfaEnabled;
  const [incidents, setIncidents] = useState<CommandIncident[]>(fallbackIncidents);
  const [loadState, setLoadState] = useState<IncidentLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading incident feed");
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [selectedIncidentId, setSelectedIncidentId] = useState(fallbackIncidents[0]?.id ?? "");

  const refreshIncidents = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading incident feed");

    try {
      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, { signal });
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

  const filteredIncidents = useMemo(
    () => incidents.filter((incident) => matchesFilters(incident, filters)),
    [filters, incidents]
  );

  const selectedIncident = useMemo(() => {
    if (!filteredIncidents.length) {
      return undefined;
    }
    return (
      filteredIncidents.find((incident) => incident.id === selectedIncidentId) ??
      filteredIncidents[0]
    );
  }, [filteredIncidents, selectedIncidentId]);

  useEffect(() => {
    if (!filteredIncidents.length) {
      setSelectedIncidentId("");
      return;
    }

    if (!filteredIncidents.some((incident) => incident.id === selectedIncidentId)) {
      setSelectedIncidentId(filteredIncidents[0].id);
    }
  }, [filteredIncidents, selectedIncidentId]);

  const metrics = useMemo(() => buildQueueMetrics(incidents), [incidents]);
  const filterOptions = useMemo(() => buildFilterOptions(incidents), [incidents]);
  const pendingAlertIncident = useMemo(
    () =>
      filteredIncidents.find((incident) => incident.severity === "severe" || incident.severity === "emergency") ??
      filteredIncidents[0],
    [filteredIncidents]
  );

  const updateFilter = (key: keyof FilterState) => (event: SelectChangeEvent) => {
    setFilters((current) => ({ ...current, [key]: event.target.value }));
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
              Incident command requires an agency account, an approved role, and completed MFA.
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
          <Stack direction="row" spacing={1.5} alignItems="center" className="brand-lockup">
            <Box component="img" src="/brand/nadaa-logo.png" alt="NADAA shield" className="brand-logo" />
            <Box>
              <Typography variant="h6">NADAA Command</Typography>
              <Typography variant="caption">National Disaster Alert & Response Platform</Typography>
            </Box>
          </Stack>
          <Stack direction="row" spacing={1} className="topbar-actions">
            <Chip
              size="small"
              color="success"
              label={`${authoritySession.name} / MFA`}
              className="session-chip"
            />
            <Button color="inherit" variant="outlined" startIcon={<RadioTower size={17} />}>
              Issue alert
            </Button>
            <Button color="secondary" variant="contained" startIcon={<Truck size={17} />}>
              Assign team
            </Button>
          </Stack>
        </Toolbar>
      </AppBar>

      <Container maxWidth="xl" className="dashboard-shell">
        <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" gap={2} className="page-heading">
          <Box>
            <Typography variant="overline" color="secondary">
              Incident command
            </Typography>
            <Typography variant="h4">Live Greater Accra operations map</Typography>
            <Typography color="text.secondary">
              Monitor emergencies by place, severity, hazard, time, and response status.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <Chip
              icon={<Eye size={16} />}
              label={loadState === "ready" ? "Live API" : loadState === "empty" ? "No active incidents" : "Fixture mode"}
              color={loadState === "ready" ? "success" : loadState === "empty" ? "default" : "warning"}
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

        {loadState === "fallback" || loadState === "error" || loadState === "empty" ? (
          <Alert severity={loadState === "empty" ? "info" : "warning"} className="feed-alert">
            {loadMessage}
          </Alert>
        ) : null}
        {loadState === "loading" ? <LinearProgress className="feed-progress" /> : null}

        <Grid container spacing={2.5}>
          {metrics.map((item) => {
            const Icon = item.icon;
            return (
              <Grid size={{ xs: 12, sm: 6, lg: 3 }} key={item.label}>
                <Paper className="metric-card">
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
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
          <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
            <Filter size={20} color={nadaaBrand.colors.navy} />
            <Typography variant="h6">Filters</Typography>
          </Stack>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect label="Hazard" value={filters.hazard} onChange={updateFilter("hazard")}>
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
              <CommandSelect label="Severity" value={filters.severity} onChange={updateFilter("severity")}>
                <MenuItem value="all">All severities</MenuItem>
                {filterOptions.severities.map((severity) => (
                  <MenuItem value={severity} key={severity}>
                    {severityLabel(severity)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect label="Status" value={filters.status} onChange={updateFilter("status")}>
                <MenuItem value="all">All statuses</MenuItem>
                {filterOptions.statuses.map((status) => (
                  <MenuItem value={status} key={status}>
                    {statusLabel(status)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 2.4 }}>
              <CommandSelect label="Time" value={filters.time} onChange={updateFilter("time")}>
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
              <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={2}>
                <Box>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <MapPinned size={21} color={nadaaBrand.colors.navy} />
                    <Typography variant="h5">Incident map</Typography>
                  </Stack>
                  <Typography color="text.secondary">
                    {filteredIncidents.length} visible of {incidents.length} incidents
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  {filterOptions.hazards.slice(0, 4).map((hazard) => (
                    <Chip key={hazard} label={hazardLabel(hazard)} size="small" />
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
              <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
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
                          <Typography variant="subtitle2">{incident.reference}</Typography>
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
                              backgroundColor: severityColors[incident.severity],
                              color: "#FFFFFF"
                            }}
                          />
                        </TableCell>
                        <TableCell>{statusLabel(incident.status)}</TableCell>
                        <TableCell>{incident.assignedAgency}</TableCell>
                        <TableCell>{formatIncidentAge(incident.createdAt)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <EmptyState title="No incidents match these filters" detail="Adjust filters or refresh the feed." />
              )}
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, lg: 4 }}>
            <Stack spacing={2.5}>
              <IncidentDetailPanel incident={selectedIncident} />

              <Paper className="surface alert-panel">
                <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                  <BellRing size={21} color={nadaaBrand.colors.red} />
                  <Typography variant="h6">Alert approval</Typography>
                </Stack>
                {pendingAlertIncident ? (
                  <Stack spacing={1.5}>
                    <Box>
                      <Stack direction="row" justifyContent="space-between" gap={1}>
                        <Typography variant="subtitle2">
                          {severityLabel(pendingAlertIncident.severity)} {hazardLabel(pendingAlertIncident.type)} watch
                        </Typography>
                        <Chip size="small" label="Draft" color="warning" />
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {pendingAlertIncident.district} · {pendingAlertIncident.responderEta} responder ETA
                      </Typography>
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={alertReadiness(pendingAlertIncident)}
                      color={pendingAlertIncident.priorityReview ? "error" : "warning"}
                    />
                    <Button variant="contained" color="error" startIcon={<BellRing size={17} />}>
                      Review alert
                    </Button>
                  </Stack>
                ) : (
                  <EmptyState title="No alert candidates" detail="Filtered incidents do not currently require alert review." />
                )}
              </Paper>

              <Paper className="surface">
                <Typography variant="h6" className="section-heading">
                  Operating posture
                </Typography>
                <Stack spacing={1.25}>
                  <StatusLine label="Incident feed" value={loadState === "ready" ? "Live" : "Fixture"} color={loadState === "ready" ? "success" : "warning"} />
                  <StatusLine label="Authority session" value={authoritySession.agency} color="success" />
                  <StatusLine label="ML alerts" value="Human review" color="warning" />
                  <StatusLine label="Audit trail" value="Required" color="success" />
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
  value
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
  selectedIncidentId
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
      scrollWheelZoom: false
    });

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
      maxZoom: 19
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
      const marker = L.circleMarker([incident.location.lat, incident.location.lng], {
        radius: isSelected ? 13 : 9,
        color: "#FFFFFF",
        weight: isSelected ? 4 : 2,
        fillColor: severityColors[incident.severity],
        fillOpacity: isSelected ? 0.95 : 0.78
      });
      marker.bindPopup(
        `<strong>${incident.reference}</strong><br>${hazardLabel(incident.type)} · ${severityLabel(incident.severity)}<br>${incident.locality}`
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
    const selected = incidents.find((incident) => incident.id === selectedIncidentId);
    if (!map || !selected) {
      return;
    }
    map.flyTo([selected.location.lat, selected.location.lng], Math.max(map.getZoom(), 12), {
      animate: true,
      duration: 0.45
    });
  }, [incidents, selectedIncidentId]);

  return (
    <Box className="map-frame">
      <Box ref={containerRef} className="leaflet-command-map" />
      {!incidents.length ? (
        <Box className="map-empty">
          <EmptyState title="No map markers" detail="No incidents match the current command filters." />
        </Box>
      ) : null}
    </Box>
  );
}

function IncidentDetailPanel({ incident }: { incident?: CommandIncident }) {
  if (!incident) {
    return (
      <Paper className="surface">
        <EmptyState title="No incident selected" detail="Choose a map marker or queue row to inspect the incident." />
      </Paper>
    );
  }

  return (
    <Paper className="surface detail-panel">
      <Stack direction="row" justifyContent="space-between" gap={1} className="section-heading">
        <Box>
          <Typography variant="overline" color="secondary">
            Selected incident
          </Typography>
          <Typography variant="h6">{incident.reference}</Typography>
        </Box>
        <Chip
          size="small"
          label={severityLabel(incident.severity)}
          style={{ backgroundColor: severityColors[incident.severity], color: "#FFFFFF" }}
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

      <Stack direction="row" spacing={1} className="detail-actions">
        <Button variant="contained" startIcon={<CheckCheck size={17} />}>
          Verify
        </Button>
        <Button variant="outlined" startIcon={<Truck size={17} />}>
          Assign
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
  value
}: {
  color: "success" | "warning" | "default";
  label: string;
  value: string;
}) {
  return (
    <Stack direction="row" justifyContent="space-between" alignItems="center" gap={1}>
      <Typography variant="body2">{label}</Typography>
      <Chip size="small" label={value} color={color} />
    </Stack>
  );
}

function EmptyState({ detail, title }: { detail: string; title: string }) {
  return (
    <Stack alignItems="center" justifyContent="center" spacing={1} className="empty-state">
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
      value: incidents.filter((incident) => incident.status === "reported" || incident.status === "under_review").length,
      icon: ShieldAlert,
      color: nadaaBrand.colors.red
    },
    {
      label: "Verified",
      value: incidents.filter((incident) => incident.status === "verified" || incident.status === "assigned").length,
      icon: CheckCheck,
      color: nadaaBrand.colors.green
    },
    {
      label: "Teams en route",
      value: incidents.filter((incident) => incident.status === "response_en_route" || incident.status === "on_scene").length,
      icon: Truck,
      color: "#0B6FB8"
    },
    {
      label: "Priority review",
      value: incidents.filter((incident) => incident.priorityReview).length,
      icon: AlertTriangle,
      color: nadaaBrand.colors.gold
    }
  ];
}

function buildFilterOptions(incidents: CommandIncident[]) {
  return {
    hazards: uniqueSorted(incidents.map((incident) => incident.type)),
    regionDistricts: uniqueSorted(incidents.map((incident) => `${incident.region} / ${incident.district}`)),
    severities: uniqueSorted(incidents.map((incident) => incident.severity)).sort(
      (a, b) => severityOrder[b] - severityOrder[a]
    ),
    statuses: uniqueSorted(incidents.map((incident) => incident.status))
  };
}

function matchesFilters(incident: CommandIncident, filters: FilterState) {
  if (filters.hazard !== "all" && incident.type !== filters.hazard) {
    return false;
  }
  if (filters.regionDistrict !== "all" && `${incident.region} / ${incident.district}` !== filters.regionDistrict) {
    return false;
  }
  if (filters.severity !== "all" && incident.severity !== filters.severity) {
    return false;
  }
  if (filters.status !== "all" && incident.status !== filters.status) {
    return false;
  }
  if (filters.time !== "all" && !withinTimeWindow(incident.createdAt, filters.time)) {
    return false;
  }
  return true;
}

function withinTimeWindow(createdAt: string, timeFilter: FilterState["time"]) {
  const hours = timeFilter === "1h" ? 1 : timeFilter === "6h" ? 6 : timeFilter === "24h" ? 24 : 0;
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
      incident.priorityReview ? "Priority review flag is active" : "Dispatcher monitoring normal queue"
    ],
    source: "api"
  };
}

function districtFromCoordinates(location: { lat: number; lng: number }) {
  if (location.lng > -0.08) {
    return { region: "Greater Accra", district: "Tema Metropolitan", locality: "Tema" };
  }
  if (location.lng < -0.25) {
    return { region: "Greater Accra", district: "Ablekuma West", locality: "Ablekuma" };
  }
  if (location.lat < 5.56) {
    return { region: "Greater Accra", district: "Accra Metropolitan", locality: "Korle Gonno" };
  }
  return { region: "Greater Accra", district: "Accra Metropolitan", locality: "Accra Central" };
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

function statusLabel(status: IncidentStatus) {
  return status
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function formatIncidentAge(createdAt: string) {
  const latestFixtureTime = new Date("2026-07-06T19:00:00Z").getTime();
  const minutes = Math.max(1, Math.round((latestFixtureTime - new Date(createdAt).getTime()) / 60000));
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  return `${hours} hr ${minutes % 60} min`;
}

export default App;
