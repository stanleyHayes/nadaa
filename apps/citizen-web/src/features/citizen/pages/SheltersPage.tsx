import { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import {
  Cross,
  LifeBuoy,
  Loader2,
  RefreshCw,
  TriangleAlert,
} from "lucide-react";
import type {
  AreaRiskResponse,
  NearbyShelterResponse,
  ReliefPointNearbyResponse,
  RoadClosureListResponse,
  RoadClosureRecord,
} from "@nadaa/shared-types";
import {
  RISK_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "@/app/config";
import { PageBanner } from "../components/PageBanner";
import { PageHeader, Reveal, RoutePlanner, SkeletonRows } from "../components";
import { areaPresets } from "../data";
import type { RiskState, ShelterState } from "../types";
import {
  extractAPIError,
  formatDistance,
  formatListLabel,
  formatOccupancy,
  formatReliefStock,
  formatSupportType,
} from "../utils";

/**
 * Self-contained "Shelters, routes & relief" surface migrated from the legacy
 * `#resources` section: the evacuation route planner, nearby shelters, relief
 * distribution, recovery support, nearby responders and active road closures.
 * Shelters, relief, road closures and area risk (for the responders panel) all
 * load live from their backend services for the default Accra preset on mount
 * and can be refreshed.
 */
function ShelterResources() {
  const riskCoordinates = {
    lat: areaPresets[0].lat.toFixed(6),
    lng: areaPresets[0].lng.toFixed(6),
  };
  const [risk, setRisk] = useState<AreaRiskResponse | null>(null);
  const [riskState, setRiskState] = useState<RiskState>({ status: "idle" });
  const [shelterSupport, setShelterSupport] =
    useState<NearbyShelterResponse | null>(null);
  const [reliefPoints, setReliefPoints] =
    useState<ReliefPointNearbyResponse | null>(null);
  const [roadClosures, setRoadClosures] = useState<RoadClosureRecord[]>([]);
  const [shelterState, setShelterState] = useState<ShelterState>({
    status: "loading",
    message: "Checking shelters",
  });

  async function fetchRoadClosures(lat: number, lng: number) {
    if (!navigator.onLine) {
      setRoadClosures([]);
      return;
    }
    try {
      const response = await fetch(
        `${ROAD_CLOSURE_API_BASE}/road-closures?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}&limit=6`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload = (await response.json()) as RoadClosureListResponse;
      setRoadClosures(payload.closures);
    } catch {
      setRoadClosures([]);
    }
  }

  async function fetchRisk(lat: number, lng: number) {
    if (!navigator.onLine) {
      setRisk(null);
      setRiskState({
        status: "error",
        message: "Responder lookup needs a connection. Reconnect and try again.",
      });
      return;
    }

    setRiskState({ status: "loading", message: "Loading responders" });

    try {
      const response = await fetch(
        `${RISK_API_BASE}/risk?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload = (await response.json()) as AreaRiskResponse;
      setRisk(payload);
      setRiskState({ status: "idle" });
    } catch {
      setRisk(null);
      setRiskState({
        status: "error",
        message: "Couldn't reach the risk service. Try refreshing.",
      });
    }
  }

  async function fetchReliefPoints(lat: number, lng: number) {
    if (!navigator.onLine) {
      setReliefPoints(null);
      return;
    }

    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/relief-points/nearby?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload = (await response.json()) as ReliefPointNearbyResponse;
      setReliefPoints(payload);
    } catch {
      setReliefPoints(null);
    }
  }

  async function fetchShelters(lat: number, lng: number) {
    if (!navigator.onLine) {
      setShelterSupport(null);
      setShelterState({
        status: "error",
        message: "Shelter lookup needs a connection. Reconnect and try again.",
      });
      return;
    }

    setShelterState({ status: "loading", message: "Checking shelters" });

    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/shelters/nearby?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as NearbyShelterResponse;
      setShelterSupport(payload);
      setShelterState({
        status: "idle",
        message: "Shelter and recovery support updated.",
      });
    } catch {
      setShelterSupport(null);
      setShelterState({
        status: "error",
        message: "Couldn't reach the shelter service. Try refreshing.",
      });
    }
  }

  useEffect(() => {
    const { lat, lng } = areaPresets[0];
    void fetchRisk(lat, lng);
    void fetchShelters(lat, lng);
    void fetchReliefPoints(lat, lng);
    void fetchRoadClosures(lat, lng);
    // Load live data for the default Accra preset once on mount.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const refreshShelterSupport = () => {
    const lat = Number(riskCoordinates.lat);
    const lng = Number(riskCoordinates.lng);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      setShelterState({
        status: "error",
        message: "Shelter refresh needs valid risk coordinates first.",
      });
      return;
    }

    void fetchRisk(lat, lng);
    void fetchShelters(lat, lng);
    void fetchReliefPoints(lat, lng);
  };

  return (
    <Reveal className="citizen-section">
      <section aria-label="Shelters and relief" id="resources">
        <PageHeader
          icon={LifeBuoy}
          title="Shelters, routes & relief"
          subtitle="Find a safe place, plan a route, and locate aid near you."
          tone="green"
          action={
            <Button
              type="button"
              variant="outlined"
              size="small"
              startIcon={
                shelterState.status === "loading" ? (
                  <Loader2 size={16} className="spin-icon" />
                ) : (
                  <RefreshCw size={16} />
                )
              }
              onClick={refreshShelterSupport}
              disabled={shelterState.status === "loading"}
            >
              Refresh
            </Button>
          }
        />

        <RoutePlanner />

        <Grid
          container
          spacing={2.5}
          className="citizen-subsection"
          sx={{ flexDirection: "column" }}
        >
          <Grid size={{ xs: 12 }}>
            <Paper className="surface">
              <PageHeader
                icon={Cross}
                title="Nearby shelters"
                subtitle="Capacity and facilities"
                tone="green"
                as="h3"
              />
              {shelterState.status === "error" ? (
                <Alert severity="error" className="warning-alert">
                  {shelterState.message}
                </Alert>
              ) : null}
              <Stack spacing={1.25}>
                {shelterState.status === "loading" ? (
                  <SkeletonRows />
                ) : shelterSupport && shelterSupport.shelters.length > 0 ? (
                  shelterSupport.shelters.map((shelter) => (
                    <Paper
                      variant="outlined"
                      className="shelter-row"
                      key={shelter.id}
                    >
                      <Stack
                        direction="row"
                        spacing={1}
                        sx={{
                          justifyContent: "space-between"
                        }}
                      >
                        <Box>
                          <Typography variant="subtitle2">
                            {shelter.name}
                          </Typography>
                          <Typography variant="body2" sx={{
                            color: "text.secondary"
                          }}>
                            {formatOccupancy(shelter)}
                            {shelter.distanceMeters
                              ? ` · ${formatDistance(shelter.distanceMeters)}`
                              : ""}
                          </Typography>
                          {shelter.facilities.length ? (
                            <Typography
                              variant="caption"
                              sx={{
                                color: "text.secondary"
                              }}
                            >
                              {formatListLabel(shelter.facilities)}
                            </Typography>
                          ) : null}
                        </Box>
                        {shelter.contact ? (
                          <Chip
                            size="small"
                            label={shelter.contact}
                            color={
                              shelter.status === "full" ? "warning" : "success"
                            }
                          />
                        ) : null}
                      </Stack>
                    </Paper>
                  ))
                ) : shelterState.status === "error" ? null : (
                  <Alert severity="info" className="warning-alert">
                    No nearby shelters were returned for this area.
                  </Alert>
                )}
              </Stack>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Paper className="surface">
              <PageHeader
                icon={LifeBuoy}
                title="Relief distribution"
                subtitle="Food, water, medical and supply points"
                tone="gold"
                as="h3"
              />
              <Stack spacing={1.25}>
                {shelterState.status === "loading" ? (
                  <SkeletonRows />
                ) : reliefPoints && reliefPoints.reliefPoints.length > 0 ? (
                  reliefPoints.reliefPoints.map((point) => (
                    <Paper
                      variant="outlined"
                      className="shelter-row"
                      key={point.id}
                    >
                      <Stack spacing={1}>
                        <Stack
                          direction="row"
                          spacing={1}
                          sx={{
                            justifyContent: "space-between"
                          }}
                        >
                          <Box>
                            <Typography variant="subtitle2">
                              {point.name}
                            </Typography>
                            <Typography variant="body2" sx={{
                              color: "text.secondary"
                            }}>
                              {formatSupportType(point.type)}
                              {point.distanceMeters
                                ? ` · ${formatDistance(point.distanceMeters)}`
                                : ""}
                            </Typography>
                          </Box>
                          <Chip
                            size="small"
                            label={point.status}
                            color={
                              point.status === "open"
                                ? "success"
                                : point.status === "limited"
                                  ? "warning"
                                  : "default"
                            }
                          />
                        </Stack>
                        <Typography variant="body2">
                          {formatReliefStock(point.stockCategories)}
                        </Typography>
                        <Typography variant="caption" sx={{
                          color: "text.secondary"
                        }}>
                          {point.operatingHours} · {point.schedule}
                        </Typography>
                        {point.eligibility ? (
                          <Alert severity="info" className="warning-alert">
                            {point.eligibility}
                          </Alert>
                        ) : null}
                      </Stack>
                    </Paper>
                  ))
                ) : shelterState.status === "error" ? null : (
                  <Alert severity="info" className="warning-alert">
                    No relief distribution points were returned for this area.
                  </Alert>
                )}
              </Stack>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Paper className="surface">
              <PageHeader
                icon={LifeBuoy}
                title="Recovery support"
                subtitle="Registration and case follow-up"
                tone="green"
                as="h3"
              />
              <Stack spacing={1.25}>
                {shelterState.status === "loading" ? (
                  <SkeletonRows />
                ) : shelterSupport &&
                  shelterSupport.recoverySupport.length > 0 ? (
                  shelterSupport.recoverySupport.map((support) => (
                    <Paper
                      variant="outlined"
                      className="shelter-row"
                      key={support.id}
                    >
                      <Stack
                        direction="row"
                        spacing={1}
                        sx={{
                          justifyContent: "space-between"
                        }}
                      >
                        <Box>
                          <Typography variant="subtitle2">
                            {support.name}
                          </Typography>
                          <Typography variant="body2" sx={{
                            color: "text.secondary"
                          }}>
                            {formatSupportType(support.type)}
                            {support.distanceMeters
                              ? ` · ${formatDistance(support.distanceMeters)}`
                              : ""}
                          </Typography>
                          <Typography variant="caption" sx={{
                            color: "text.secondary"
                          }}>
                            {support.hours} ·{" "}
                            {formatListLabel(support.services)}
                          </Typography>
                        </Box>
                        <Chip
                          size="small"
                          label={support.status}
                          color={
                            support.status === "open" ? "success" : "warning"
                          }
                        />
                      </Stack>
                    </Paper>
                  ))
                ) : shelterState.status === "error" ? null : (
                  <Alert severity="info" className="warning-alert">
                    No recovery support locations were returned for this area.
                  </Alert>
                )}
              </Stack>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Paper className="surface">
              <PageHeader
                icon={LifeBuoy}
                title="Nearby responders"
                subtitle="Agencies and facilities near you"
                tone="red"
                as="h3"
              />
              <Stack spacing={1.25}>
                {riskState.status === "loading" ? (
                  <SkeletonRows />
                ) : riskState.status === "error" ? (
                  <Alert severity="error" className="warning-alert">
                    {riskState.message}
                  </Alert>
                ) : risk && risk.nearbyFacilities.length > 0 ? (
                  risk.nearbyFacilities.map((facility) => (
                    <Paper
                      variant="outlined"
                      className="shelter-row"
                      key={facility.id}
                    >
                      <Stack
                        direction="row"
                        spacing={1}
                        sx={{
                          justifyContent: "space-between"
                        }}
                      >
                        <Box>
                          <Typography variant="subtitle2">
                            {facility.name}
                          </Typography>
                          <Typography variant="body2" sx={{
                            color: "text.secondary"
                          }}>
                            {facility.type.replace("_", " ")}
                            {facility.distanceMeters
                              ? ` · ${formatDistance(facility.distanceMeters)}`
                              : ""}
                          </Typography>
                        </Box>
                        {facility.contact ? (
                          <Chip
                            size="small"
                            label={facility.contact}
                            color="error"
                          />
                        ) : null}
                      </Stack>
                    </Paper>
                  ))
                ) : (
                  <Alert severity="info" className="warning-alert">
                    No nearby response facilities were returned for this area.
                  </Alert>
                )}
              </Stack>
            </Paper>
          </Grid>

          {roadClosures.length > 0 ? (
            <Grid size={{ xs: 12 }}>
              <Paper className="surface">
                <PageHeader
                  icon={TriangleAlert}
                  title="Road closures"
                  subtitle="Active closures near this area"
                  tone="gold"
                  as="h3"
                />
                <Grid container spacing={1.25}>
                  {roadClosures.map((closure) => (
                    <Grid size={{ xs: 12, md: 6 }} key={closure.id}>
                      <Paper variant="outlined" className="shelter-row">
                        <Stack
                          direction="row"
                          spacing={1}
                          sx={{
                            justifyContent: "space-between"
                          }}
                        >
                          <Box>
                            <Typography variant="subtitle2">
                              {closure.roadName}
                            </Typography>
                            <Typography variant="body2" sx={{
                              color: "text.secondary"
                            }}>
                              {closure.reason ?? "Road closure"} ·{" "}
                              {closure.severity}
                            </Typography>
                            {closure.detourNote ? (
                              <Typography
                                variant="caption"
                                sx={{
                                  color: "text.secondary"
                                }}
                              >
                                Detour: {closure.detourNote}
                              </Typography>
                            ) : null}
                          </Box>
                          <Chip
                            size="small"
                            label={closure.status}
                            color="warning"
                          />
                        </Stack>
                      </Paper>
                    </Grid>
                  ))}
                </Grid>
              </Paper>
            </Grid>
          ) : null}
        </Grid>
      </section>
    </Reveal>
  );
}

/** Shelters & relief (route `/shelters`). Migrated from the legacy `#resources` section. */
export function SheltersPage() {
  return (
    <>
      <PageBanner
        eyebrow="Shelters & routes"
        subtitle="Find the nearest shelters and relief points, plan a safe evacuation route, and check current road closures near you."
        title="Shelters, safe routes & relief"
      />
      <div className="citizen-shell">
        <ShelterResources />
      </div>
    </>
  );
}

export default SheltersPage;
