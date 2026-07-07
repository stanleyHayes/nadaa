import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  AppBar,
  Box,
  Button,
  Container,
  Grid,
  Paper,
  Stack,
  Tab,
  Tabs,
  Toolbar,
  Typography,
} from "@mui/material";
import {
  Activity,
  Building2,
  ClipboardList,
  LayoutDashboard,
  PackageCheck,
  Phone,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CreateReliefPointRequest,
  HospitalCapacityRecord,
  HospitalCapacityResponse,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatus,
  IncidentStatusUpdateRequest,
  ReliefPointListResponse,
  ReliefPointNearbyResponse,
  ReliefPointRecord,
  ReliefPointStockHistoryResponse,
  ReliefStockHistoryEntry,
  RoadClosureListResponse,
  RoadClosureRecord,
  ShelterRecord,
  UpdateReliefPointRequest,
} from "@nadaa/shared-types";
import { agencyHeaders, agencyRoles, agencySession } from "../../app/session";
import {
  INCIDENT_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import {
  EmptyState,
  ErrorState,
  HospitalCapacityCard,
  HospitalCapacityUpdateForm,
  IncidentDetail,
  IncidentFilters,
  IncidentListItem,
  LoadingState,
  MetricCard,
  ReliefPointCard,
  ReliefPointForm,
  ReliefStockHistoryList,
  ShelterCard,
  ShelterOccupancyForm,
  StatusUpdateForm,
} from "./components";
import {
  fallbackHospitals,
  fallbackIncidents,
  fallbackReliefPoints,
  fallbackShelters,
  initialHospitalCapacityForm,
  initialReliefPointForm,
  initialShelterOccupancyForm,
  initialStatusForm,
} from "./data";
import type {
  AgencyTab,
  HospitalCapacityFormState,
  IncidentFilterState,
  ReliefPointFormState,
  ShelterOccupancyFormState,
  StatusFormState,
  UpdateLoadState,
} from "./types";
import {
  matchesFilters,
  parseStockCategories,
  reliefPointToForm,
} from "./utils";

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

async function fetchAssignedIncidents(): Promise<IncidentRecord[]> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents?assignedToMe=true`,
    { headers: agencyHeaders() },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as IncidentListResponse;
  return payload.incidents;
}

async function updateIncidentStatus(
  incidentId: string,
  request: IncidentStatusUpdateRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/status`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as IncidentRecord;
}

async function fetchShelters(
  lat: number,
  lng: number,
): Promise<ShelterRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/shelters/nearby?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as { shelters: ShelterRecord[] };
    return payload.shelters;
  } catch {
    return fallbackShelters;
  }
}

async function fetchHospitalCapacity(
  lat: number,
  lng: number,
): Promise<HospitalCapacityRecord[]> {
  try {
    const params = new URLSearchParams({
      includeStale: "true",
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/hospitals/capacity?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as HospitalCapacityResponse;
    return payload.facilities;
  } catch {
    return fallbackHospitals;
  }
}

async function fetchRoadClosures(
  lat: number,
  lng: number,
): Promise<RoadClosureRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${ROAD_CLOSURE_API_BASE}/road-closures?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as RoadClosureListResponse;
    return payload.closures;
  } catch {
    return [];
  }
}

async function fetchNearbyReliefPoints(
  lat: number,
  lng: number,
): Promise<ReliefPointRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/relief-points/nearby?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as ReliefPointNearbyResponse;
    return payload.reliefPoints;
  } catch {
    return fallbackReliefPoints;
  }
}

async function fetchReliefPoints(): Promise<ReliefPointRecord[]> {
  const response = await fetch(`${SHELTER_API_BASE}/relief-points?limit=20`);
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as ReliefPointListResponse;
  return payload.reliefPoints;
}

async function fetchReliefPointHistory(
  reliefPointId: string,
): Promise<ReliefStockHistoryEntry[]> {
  const response = await fetch(
    `${SHELTER_API_BASE}/relief-points/${encodeURIComponent(reliefPointId)}/stock-history`,
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as ReliefPointStockHistoryResponse;
  return payload.history;
}

async function createReliefPoint(
  request: CreateReliefPointRequest,
): Promise<ReliefPointRecord> {
  const response = await fetch(`${SHELTER_API_BASE}/relief-points`, {
    body: JSON.stringify(request),
    headers: agencyHeaders(),
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as ReliefPointRecord;
}

async function updateReliefPoint(
  reliefPointId: string,
  request: UpdateReliefPointRequest,
): Promise<ReliefPointRecord> {
  const response = await fetch(
    `${SHELTER_API_BASE}/relief-points/${encodeURIComponent(reliefPointId)}`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as ReliefPointRecord;
}

export function AgencyApp() {
  const [activeTab, setActiveTab] = useState<AgencyTab>("dashboard");
  const [incidents, setIncidents] = useState<IncidentRecord[]>([]);
  const [incidentLoadState, setIncidentLoadState] = useState<
    "loading" | "ready" | "fallback" | "empty" | "error"
  >("loading");
  const [incidentError, setIncidentError] = useState<string | null>(null);
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(
    null,
  );
  const [filters, setFilters] = useState<IncidentFilterState>({
    hazard: "all",
    severity: "all",
    status: "all",
  });
  const [statusForm, setStatusForm] =
    useState<StatusFormState>(initialStatusForm);
  const [statusUpdateState, setStatusUpdateState] =
    useState<UpdateLoadState>("idle");
  const [statusUpdateError, setStatusUpdateError] = useState<string | null>(
    null,
  );
  const [shelters, setShelters] = useState<ShelterRecord[]>([]);
  const [shelterForm, setShelterForm] = useState<ShelterOccupancyFormState>(
    initialShelterOccupancyForm,
  );
  const [hospitals, setHospitals] = useState<HospitalCapacityRecord[]>([]);
  const [hospitalForm, setHospitalForm] = useState<HospitalCapacityFormState>(
    initialHospitalCapacityForm,
  );
  const [capacityLoadState, setCapacityLoadState] = useState<
    "loading" | "ready" | "fallback" | "empty" | "error"
  >("loading");
  const [roadClosures, setRoadClosures] = useState<RoadClosureRecord[]>([]);
  const [nearbyReliefPoints, setNearbyReliefPoints] = useState<
    ReliefPointRecord[]
  >([]);
  const [reliefPoints, setReliefPoints] = useState<ReliefPointRecord[]>([]);
  const [selectedReliefPointId, setSelectedReliefPointId] = useState<
    string | null
  >(null);
  const [reliefForm, setReliefForm] = useState<ReliefPointFormState>(
    initialReliefPointForm,
  );
  const [reliefHistory, setReliefHistory] = useState<ReliefStockHistoryEntry[]>(
    [],
  );
  const [reliefLoadState, setReliefLoadState] = useState<
    "loading" | "ready" | "fallback" | "empty" | "error"
  >("loading");
  const [reliefUpdateState, setReliefUpdateState] =
    useState<UpdateLoadState>("idle");
  const [reliefError, setReliefError] = useState<string | null>(null);

  const hasAccess =
    agencySession.mfaEnabled && agencyRoles.includes(agencySession.role);

  useEffect(() => {
    if (!hasAccess) return;
    void loadIncidents();
    void loadReliefPoints();
  }, [hasAccess]);

  async function loadIncidents() {
    setIncidentLoadState("loading");
    setIncidentError(null);
    try {
      const data = await fetchAssignedIncidents();
      setIncidents(data.length > 0 ? data : fallbackIncidents);
      setIncidentLoadState(data.length > 0 ? "ready" : "fallback");
    } catch (error) {
      setIncidents(fallbackIncidents);
      setIncidentError(
        error instanceof Error
          ? error.message
          : "Could not load assigned incidents.",
      );
      setIncidentLoadState("fallback");
    }
  }

  async function loadReliefPoints() {
    setReliefLoadState("loading");
    setReliefError(null);
    try {
      const data = await fetchReliefPoints();
      const nextPoints = data.length > 0 ? data : fallbackReliefPoints;
      setReliefPoints(nextPoints);
      setReliefLoadState(data.length > 0 ? "ready" : "fallback");
      setSelectedReliefPointId(
        (current) => current ?? nextPoints[0]?.id ?? null,
      );
    } catch (error) {
      setReliefPoints(fallbackReliefPoints);
      setSelectedReliefPointId(
        (current) => current ?? fallbackReliefPoints[0]?.id ?? null,
      );
      setReliefError(
        error instanceof Error
          ? error.message
          : "Could not load relief distribution points.",
      );
      setReliefLoadState("fallback");
    }
  }

  async function loadReliefHistory(reliefPointId: string) {
    try {
      const history = await fetchReliefPointHistory(reliefPointId);
      setReliefHistory(history);
    } catch (error) {
      setReliefHistory([]);
      setReliefError(
        error instanceof Error
          ? error.message
          : "Could not load relief stock history.",
      );
    }
  }

  const filteredIncidents = useMemo(() => {
    return incidents.filter((incident) => matchesFilters(incident, filters));
  }, [incidents, filters]);

  const selectedIncident = useMemo(
    () =>
      incidents.find((incident) => incident.id === selectedIncidentId) ?? null,
    [incidents, selectedIncidentId],
  );

  const selectedReliefPoint = useMemo(
    () =>
      reliefPoints.find((point) => point.id === selectedReliefPointId) ?? null,
    [reliefPoints, selectedReliefPointId],
  );

  useEffect(() => {
    if (!selectedIncident) return;
    setStatusForm({ ...initialStatusForm, status: selectedIncident.status });
    setCapacityLoadState("loading");
    void loadCapacity(
      selectedIncident.location.lat,
      selectedIncident.location.lng,
    );
  }, [selectedIncident]);

  useEffect(() => {
    if (!selectedReliefPoint) {
      setReliefHistory([]);
      return;
    }
    setReliefForm(reliefPointToForm(selectedReliefPoint));
    void loadReliefHistory(selectedReliefPoint.id);
  }, [selectedReliefPoint]);

  async function loadCapacity(lat: number, lng: number) {
    const [nearbyShelters, nearbyHospitals, nearbyClosures, nearbyRelief] =
      await Promise.all([
        fetchShelters(lat, lng),
        fetchHospitalCapacity(lat, lng),
        fetchRoadClosures(lat, lng),
        fetchNearbyReliefPoints(lat, lng),
      ]);
    setShelters(nearbyShelters);
    setHospitals(nearbyHospitals);
    setRoadClosures(nearbyClosures);
    setNearbyReliefPoints(nearbyRelief);
    setCapacityLoadState(
      nearbyShelters.length > 0 ||
        nearbyHospitals.length > 0 ||
        nearbyClosures.length > 0 ||
        nearbyRelief.length > 0
        ? "ready"
        : "empty",
    );
  }

  async function handleSaveReliefPoint() {
    setReliefUpdateState("loading");
    setReliefError(null);
    try {
      const location = {
        lat: Number.parseFloat(reliefForm.lat),
        lng: Number.parseFloat(reliefForm.lng),
      };
      const stockCategories = parseStockCategories(reliefForm.stockCategories);
      const request = {
        address: reliefForm.address.trim(),
        contact: reliefForm.contact.trim(),
        district: reliefForm.district.trim(),
        eligibility: reliefForm.eligibility.trim(),
        location,
        name: reliefForm.name.trim(),
        operatingHours: reliefForm.operatingHours.trim(),
        region: reliefForm.region.trim(),
        schedule: reliefForm.schedule.trim(),
        sourceRef: agencySession.agencyId,
        status: reliefForm.status,
        stockCategories,
        type: reliefForm.type,
      };
      const saved = selectedReliefPoint
        ? await updateReliefPoint(
            selectedReliefPoint.id,
            request satisfies UpdateReliefPointRequest,
          )
        : await createReliefPoint({
            ...request,
            source: "manual",
          } satisfies CreateReliefPointRequest);
      setReliefPoints((current) =>
        current.some((point) => point.id === saved.id)
          ? current.map((point) => (point.id === saved.id ? saved : point))
          : [saved, ...current],
      );
      setSelectedReliefPointId(saved.id);
      setReliefForm(reliefPointToForm(saved));
      await loadReliefHistory(saved.id);
      setReliefUpdateState("success");
    } catch (error) {
      setReliefError(
        error instanceof Error ? error.message : "Relief point update failed.",
      );
      setReliefUpdateState("error");
    }
  }

  function handleNewReliefPoint() {
    setSelectedReliefPointId(null);
    setReliefForm(initialReliefPointForm);
    setReliefHistory([]);
    setReliefUpdateState("idle");
    setReliefError(null);
  }

  async function handleStatusUpdate() {
    if (!selectedIncident) return;
    setStatusUpdateState("loading");
    setStatusUpdateError(null);
    try {
      const updated = await updateIncidentStatus(selectedIncident.id, {
        note: statusForm.note,
        resolutionNotes: statusForm.resolutionNotes,
        status: statusForm.status,
      });
      setIncidents((current) =>
        current.map((incident) =>
          incident.id === updated.id ? updated : incident,
        ),
      );
      setStatusForm({ ...initialStatusForm, status: updated.status });
      setStatusUpdateState("success");
    } catch (error) {
      setStatusUpdateError(
        error instanceof Error ? error.message : "Status update failed.",
      );
      setStatusUpdateState("error");
    }
  }

  function handleSelectIncident(incidentId: string) {
    setSelectedIncidentId(incidentId);
    setActiveTab("incident");
    setStatusUpdateState("idle");
    setStatusUpdateError(null);
  }

  const metrics = useMemo(() => {
    return {
      assigned: incidents.filter((incident) => incident.status === "assigned")
        .length,
      enRoute: incidents.filter(
        (incident) =>
          incident.status === "response_en_route" ||
          incident.status === "on_scene",
      ).length,
      priority: incidents.filter((incident) => incident.priorityReview).length,
      open: incidents.filter(
        (incident) => !["closed", "false_report"].includes(incident.status),
      ).length,
      reliefPoints: reliefPoints.filter((point) =>
        ["open", "limited"].includes(point.status),
      ).length,
    };
  }, [incidents, reliefPoints]);

  if (!hasAccess) {
    return (
      <Container maxWidth="md" sx={{ py: 8 }}>
        <Alert severity="error">
          Agency access requires MFA and a valid agency role. Please contact
          your administrator.
        </Alert>
      </Container>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100%" }}>
      <AppBar
        elevation={0}
        position="static"
        sx={{ borderBottom: 4, borderColor: "secondary.main" }}
      >
        <Toolbar>
          <Stack
            alignItems="center"
            direction="row"
            spacing={2}
            sx={{ flexGrow: 1 }}
          >
            <Box
              alt="NADAA"
              component="img"
              src="/brand/nadaa-logo.png"
              sx={{ height: 44, width: 44 }}
            />
            <Box>
              <Typography fontWeight={800} variant="h6">
                NADAA Agency Operations
              </Typography>
              <Typography color="rgba(255,255,255,0.74)" variant="caption">
                {agencySession.agency}
              </Typography>
            </Box>
          </Stack>
          <Button
            color="inherit"
            href={`tel:${nadaaBrand.supportLine}`}
            startIcon={<Phone size={18} />}
            variant="outlined"
          >
            {nadaaBrand.supportLine}
          </Button>
        </Toolbar>
      </AppBar>

      <Box
        sx={{
          borderBottom: 1,
          borderColor: "divider",
          bgcolor: "background.paper",
        }}
      >
        <Tabs
          onChange={(_event, value: AgencyTab) => setActiveTab(value)}
          textColor="primary"
          value={activeTab}
          variant="scrollable"
        >
          <Tab
            icon={<LayoutDashboard size={18} />}
            iconPosition="start"
            label="Dashboard"
            value="dashboard"
          />
          <Tab
            icon={<ClipboardList size={18} />}
            iconPosition="start"
            label="Incident"
            value="incident"
          />
          <Tab
            icon={<Building2 size={18} />}
            iconPosition="start"
            label="Capacity"
            value="capacity"
          />
          <Tab
            icon={<PackageCheck size={18} />}
            iconPosition="start"
            label="Relief"
            value="relief"
          />
        </Tabs>
      </Box>

      <Container maxWidth="xl" sx={{ flex: 1, py: 3 }}>
        {activeTab === "dashboard" ? (
          <Stack spacing={3}>
            <Typography fontWeight={800} variant="h4">
              Assigned incidents
            </Typography>

            <Grid container spacing={2}>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard
                  icon={ClipboardList}
                  label="Assigned"
                  value={metrics.assigned}
                />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard
                  icon={Activity}
                  label="En route / On scene"
                  value={metrics.enRoute}
                />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard
                  icon={ClipboardList}
                  label="Priority"
                  value={metrics.priority}
                />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard
                  icon={ClipboardList}
                  label="Open"
                  value={metrics.open}
                />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard
                  icon={PackageCheck}
                  label="Relief points"
                  value={metrics.reliefPoints}
                />
              </Grid>
            </Grid>

            <IncidentFilters filters={filters} onChange={setFilters} />

            {incidentLoadState === "loading" ? (
              <LoadingState message="Loading assigned incidents" />
            ) : incidentLoadState === "error" && !incidents.length ? (
              <ErrorState
                message={incidentError ?? "Could not load incidents"}
                onRetry={loadIncidents}
              />
            ) : filteredIncidents.length === 0 ? (
              <EmptyState message="No incidents match the current filters." />
            ) : (
              <Grid container spacing={2}>
                {filteredIncidents.map((incident) => (
                  <Grid key={incident.id} size={{ xs: 12, md: 6, lg: 4 }}>
                    <IncidentListItem
                      incident={incident}
                      onClick={() => handleSelectIncident(incident.id)}
                      selected={selectedIncidentId === incident.id}
                    />
                  </Grid>
                ))}
              </Grid>
            )}

            {incidentError && incidents.length > 0 ? (
              <Alert severity="warning" sx={{ mt: 2 }}>
                {incidentError} Showing fallback data.
              </Alert>
            ) : null}
          </Stack>
        ) : null}

        {activeTab === "incident" ? (
          <Stack spacing={3}>
            <Typography fontWeight={800} variant="h4">
              Incident detail
            </Typography>

            {!selectedIncident ? (
              <EmptyState message="Select an incident from the Dashboard tab to view details and update status." />
            ) : (
              <Grid container spacing={3}>
                <Grid size={{ xs: 12, lg: 7 }}>
                  <Paper sx={{ p: 3 }}>
                    <IncidentDetail incident={selectedIncident} />
                  </Paper>
                </Grid>
                <Grid size={{ xs: 12, lg: 5 }}>
                  <Paper sx={{ p: 3 }}>
                    <Typography fontWeight={700} gutterBottom variant="h6">
                      Update status
                    </Typography>
                    <StatusUpdateForm
                      currentStatus={selectedIncident.status}
                      form={statusForm}
                      onChange={setStatusForm}
                      onSubmit={handleStatusUpdate}
                      submitLabel="Update status"
                    />
                    {statusUpdateState === "success" ? (
                      <Alert severity="success" sx={{ mt: 2 }}>
                        Status updated successfully.
                      </Alert>
                    ) : null}
                    {statusUpdateState === "error" && statusUpdateError ? (
                      <Alert severity="error" sx={{ mt: 2 }}>
                        {statusUpdateError}
                      </Alert>
                    ) : null}
                  </Paper>
                </Grid>
              </Grid>
            )}
          </Stack>
        ) : null}

        {activeTab === "capacity" ? (
          <Stack spacing={3}>
            <Typography fontWeight={800} variant="h4">
              Capacity context
            </Typography>

            {!selectedIncident ? (
              <EmptyState message="Select an incident from the Dashboard tab to see nearby shelter and hospital capacity." />
            ) : capacityLoadState === "loading" ? (
              <LoadingState message="Loading nearby capacity" />
            ) : (
              <>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Typography fontWeight={700} gutterBottom variant="h6">
                      Nearby shelters
                    </Typography>
                    {shelters.length === 0 ? (
                      <EmptyState message="No shelters found nearby." />
                    ) : (
                      <Stack spacing={2}>
                        {shelters.map((shelter) => (
                          <ShelterCard key={shelter.id} shelter={shelter} />
                        ))}
                      </Stack>
                    )}
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Typography fontWeight={700} gutterBottom variant="h6">
                      Hospital capacity
                    </Typography>
                    {hospitals.length === 0 ? (
                      <EmptyState message="No hospitals found nearby." />
                    ) : (
                      <Stack spacing={2}>
                        {hospitals.map((facility) => (
                          <HospitalCapacityCard
                            key={facility.id}
                            facility={facility}
                          />
                        ))}
                      </Stack>
                    )}
                  </Grid>
                </Grid>

                {roadClosures.length > 0 && (
                  <Grid container spacing={3}>
                    <Grid size={{ xs: 12 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        Nearby road closures
                      </Typography>
                      <Stack spacing={2}>
                        {roadClosures.map((closure) => (
                          <Paper key={closure.id} sx={{ p: 2 }}>
                            <Stack spacing={0.5}>
                              <Typography fontWeight={700}>
                                {closure.roadName}
                              </Typography>
                              <Typography
                                color="text.secondary"
                                variant="body2"
                              >
                                {closure.reason ?? "Road closure"} ·{" "}
                                {closure.severity} · {closure.status}
                              </Typography>
                              {closure.detourNote ? (
                                <Typography variant="body2">
                                  Detour: {closure.detourNote}
                                </Typography>
                              ) : null}
                            </Stack>
                          </Paper>
                        ))}
                      </Stack>
                    </Grid>
                  </Grid>
                )}

                {nearbyReliefPoints.length > 0 && (
                  <Grid container spacing={3}>
                    <Grid size={{ xs: 12 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        Nearby relief distribution points
                      </Typography>
                      <Stack spacing={2}>
                        {nearbyReliefPoints.map((point) => (
                          <ReliefPointCard
                            key={point.id}
                            onSelect={() => {
                              setSelectedReliefPointId(point.id);
                              setActiveTab("relief");
                            }}
                            point={point}
                            selected={selectedReliefPointId === point.id}
                          />
                        ))}
                      </Stack>
                    </Grid>
                  </Grid>
                )}

                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Paper sx={{ p: 3 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        Update shelter occupancy
                      </Typography>
                      <ShelterOccupancyForm
                        form={shelterForm}
                        onChange={setShelterForm}
                        onSubmit={() => {
                          setShelterForm(initialShelterOccupancyForm);
                        }}
                      />
                    </Paper>
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Paper sx={{ p: 3 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        Update hospital capacity
                      </Typography>
                      <HospitalCapacityUpdateForm
                        form={hospitalForm}
                        onChange={setHospitalForm}
                        onSubmit={() => {
                          setHospitalForm(initialHospitalCapacityForm);
                        }}
                      />
                    </Paper>
                  </Grid>
                </Grid>
              </>
            )}
          </Stack>
        ) : null}

        {activeTab === "relief" ? (
          <Stack spacing={3}>
            <Stack
              alignItems={{ xs: "flex-start", sm: "center" }}
              direction={{ xs: "column", sm: "row" }}
              justifyContent="space-between"
              spacing={2}
            >
              <Box>
                <Typography fontWeight={800} variant="h4">
                  Relief distribution
                </Typography>
                <Typography color="text.secondary" variant="body2">
                  Publish distribution points, update stock, and keep
                  eligibility notes current.
                </Typography>
              </Box>
              <Button
                onClick={handleNewReliefPoint}
                startIcon={<PackageCheck size={18} />}
                variant="contained"
              >
                New point
              </Button>
            </Stack>

            {reliefError && reliefLoadState === "fallback" ? (
              <Alert severity="warning">
                {reliefError} Showing fixture relief distribution points.
              </Alert>
            ) : null}

            {reliefLoadState === "loading" ? (
              <LoadingState message="Loading relief distribution points" />
            ) : reliefLoadState === "error" && !reliefPoints.length ? (
              <ErrorState
                message={reliefError ?? "Could not load relief points"}
                onRetry={loadReliefPoints}
              />
            ) : (
              <Grid container spacing={3}>
                <Grid size={{ xs: 12, lg: 5 }}>
                  <Stack spacing={2}>
                    {reliefPoints.length === 0 ? (
                      <EmptyState message="No relief distribution points have been published yet." />
                    ) : (
                      reliefPoints.map((point) => (
                        <ReliefPointCard
                          key={point.id}
                          onSelect={() => {
                            setSelectedReliefPointId(point.id);
                            setReliefUpdateState("idle");
                            setReliefError(null);
                          }}
                          point={point}
                          selected={selectedReliefPointId === point.id}
                        />
                      ))
                    )}
                  </Stack>
                </Grid>
                <Grid size={{ xs: 12, lg: 7 }}>
                  <Stack spacing={3}>
                    <Paper sx={{ p: 3 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        {selectedReliefPoint
                          ? "Manage distribution point"
                          : "Create distribution point"}
                      </Typography>
                      <ReliefPointForm
                        form={reliefForm}
                        onChange={setReliefForm}
                        onSubmit={handleSaveReliefPoint}
                        submitLabel={
                          reliefUpdateState === "loading"
                            ? "Saving..."
                            : selectedReliefPoint
                              ? "Update point"
                              : "Create point"
                        }
                      />
                      {reliefUpdateState === "success" ? (
                        <Alert severity="success" sx={{ mt: 2 }}>
                          Relief distribution point saved.
                        </Alert>
                      ) : null}
                      {reliefUpdateState === "error" && reliefError ? (
                        <Alert severity="error" sx={{ mt: 2 }}>
                          {reliefError}
                        </Alert>
                      ) : null}
                    </Paper>
                    <Paper sx={{ p: 3 }}>
                      <Typography fontWeight={700} gutterBottom variant="h6">
                        Stock history
                      </Typography>
                      <ReliefStockHistoryList history={reliefHistory} />
                    </Paper>
                  </Stack>
                </Grid>
              </Grid>
            )}
          </Stack>
        ) : null}
      </Container>
    </Box>
  );
}
