import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  AlertOctagon,
  AlertTriangle,
  CheckCircle2,
  Info,
  Loader2,
  LocateFixed,
  MapPin,
  ShieldCheck,
  Waves,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  NearbyShelterResponse,
} from "@nadaa/shared-types";
import { RISK_API_BASE, SHELTER_API_BASE } from "@/app/config";
import {
  areaPresets,
  buildFallbackAlerts,
  buildFallbackGuides,
  sampleRisk,
  sampleShelterResponse,
} from "../data";
import type { RiskState } from "../types";
import {
  extractAPIError,
  hazardRoleFor,
  hazardRoles,
  readGuideCache,
  resolveAreaLookup,
  severityRoleFor,
  severityRoles,
} from "../utils";
import { useParallax } from "../hooks";
import {
  AnimatedCounter,
  PageHeader,
  Reveal,
  RiskMap,
  type RiskMapMarker,
} from "../components";
import { PageBanner } from "../components/PageBanner";

/**
 * Risk checker landing (route `/`). Migrated from the legacy `#risk` section:
 * the check-your-risk hero, the Leaflet `RiskMap`, and the risk breakdown.
 *
 * Self-contained: it owns the area lookup, risk fetch and the shelter markers
 * shown on its own map. Alerts, guides, relief points and road closures live on
 * their dedicated pages, so this page only derives the hero summary counts from
 * the shared offline fixtures/cache (matching the legacy initial render) instead
 * of duplicating those pages' data flows. The persistent `EmergencyBand` (112)
 * is rendered by `CitizenLayout`, so the legacy inline band is dropped here.
 */
export function HomePage() {
  const heroAuraRef = useParallax<HTMLDivElement>(0.12);
  const [area, setArea] = useState("Accra Central");
  const [risk, setRisk] = useState<AreaRiskResponse>(sampleRisk);
  const [riskCoordinates, setRiskCoordinates] = useState({
    lat: "5.603700",
    lng: "-0.187000",
  });
  const [riskState, setRiskState] = useState<RiskState>({ status: "idle" });
  const [shelterSupport, setShelterSupport] = useState<NearbyShelterResponse>(
    sampleShelterResponse,
  );

  const severityIcons = {
    low: CheckCircle2,
    medium: AlertTriangle,
    high: AlertTriangle,
    severe: AlertOctagon,
    info: Info,
  };

  const currentAlertCount = useMemo(
    () =>
      buildFallbackAlerts().filter((alert) => alert.status === "current")
        .length,
    [],
  );
  const offlineGuideCount = useMemo(() => {
    const cached = readGuideCache();
    const guides = cached?.guides ?? buildFallbackGuides();
    return guides.filter((guide) => guide.offlineAvailable).length;
  }, []);

  const floodRisk = useMemo(
    () => risk.risks.find((item) => item.type === "flood"),
    [risk.risks],
  );

  const mapLat = Number(riskCoordinates.lat);
  const mapLng = Number(riskCoordinates.lng);
  const hasMapCoords = Number.isFinite(mapLat) && Number.isFinite(mapLng);
  const overallSeverityRole = severityRoleFor(risk.overallRisk);
  const overallSeverityColor = severityRoles[overallSeverityRole].foreground;
  const shelterMarkers = useMemo<RiskMapMarker[]>(
    () =>
      shelterSupport.shelters.slice(0, 4).map((shelter, index) => ({
        lat: shelter.location.lat,
        lng: shelter.location.lng,
        title: shelter.name,
        color: nadaaBrand.colors.green,
        glyph: String(index + 1),
        kind: "shelter",
      })),
    [shelterSupport.shelters],
  );

  useEffect(() => {
    void fetchRisk(
      areaPresets[0].lat,
      areaPresets[0].lng,
      areaPresets[0].label,
    );
  }, []);

  async function fetchRisk(lat: number, lng: number, label: string) {
    if (!navigator.onLine) {
      setRiskState({
        status: "error",
        message:
          "Risk lookup needs a connection. Try again when you are online.",
      });
      return;
    }

    setRiskState({ status: "loading", message: "Checking area risk" });

    try {
      const response = await fetch(
        `${RISK_API_BASE}/risk?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as AreaRiskResponse;
      setRisk(payload);
      setArea(label);
      setRiskCoordinates({ lat: lat.toFixed(6), lng: lng.toFixed(6) });
      setRiskState({
        status: "idle",
        message: `Updated for ${payload.location}`,
      });
      void fetchShelters(lat, lng, payload);
    } catch (error) {
      setRiskState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not load area risk.",
      });
    }
  }

  async function fetchShelters(
    lat: number,
    lng: number,
    riskPayload: AreaRiskResponse = risk,
  ) {
    if (!navigator.onLine) {
      setShelterSupport(shelterPayloadFromRisk(riskPayload));
      return;
    }

    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/shelters/nearby?lat=${encodeURIComponent(lat)}&lng=${encodeURIComponent(lng)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as NearbyShelterResponse;
      setShelterSupport(payload);
    } catch {
      setShelterSupport(shelterPayloadFromRisk(riskPayload));
    }
  }

  const submitRiskLookup = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const lookup = resolveAreaLookup(area);
    if (!lookup) {
      setRiskState({
        status: "error",
        message:
          "Choose Accra Metropolitan, Accra flood zone, Kumasi area, or enter coordinates as lat,lng.",
      });
      return;
    }

    void fetchRisk(lookup.lat, lookup.lng, lookup.label);
  };

  const useRiskLocation = () => {
    if (!navigator.geolocation) {
      setRiskState({
        status: "error",
        message: "Location is not available on this device.",
      });
      return;
    }

    setRiskState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        void fetchRisk(
          position.coords.latitude,
          position.coords.longitude,
          "Current location",
        );
      },
      () => {
        setRiskState({
          status: "permission-denied",
          message:
            "Location permission was not granted. Choose an area or enter coordinates instead.",
        });
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  };

  const riskFormInvalid =
    riskState.status === "error" || riskState.status === "permission-denied";

  return (
    <>
      <PageBanner
        eyebrow="Check your risk"
        subtitle="Live risk scoring and the nearest shelters on the map, plus what's driving the danger right now."
        title="Your area's flood & hazard risk"
      />

      {/* ------------------------------------------------------------- *
       * Hero: check-your-risk front and centre (persistent 112 band is
       * rendered once by CitizenLayout, so the legacy inline band is gone).
       * ------------------------------------------------------------- */}
      <section className="risk-hero" id="risk">
        <div className="risk-hero__aura" aria-hidden="true" ref={heroAuraRef} />
        <div className="risk-hero__inner">
          <div className="risk-hero__copy">
            <p className="citizen-eyebrow">
              <Waves aria-hidden="true" size={16} />
              Flood risk checker
            </p>
            <h1>Check your area&rsquo;s risk</h1>
            <p className="risk-hero__sub">
              See live flood and hazard risk for anywhere in Ghana, find the
              nearest shelter, and get emergency guidance &mdash; no sign-in
              needed.
            </p>

            <dl className="risk-hero__stats">
              <div>
                <dt>Current warnings</dt>
                <dd>
                  <AnimatedCounter value={currentAlertCount} />
                </dd>
              </div>
              <div>
                <dt>Offline guides</dt>
                <dd>
                  <AnimatedCounter value={offlineGuideCount} />
                </dd>
              </div>
              <div>
                <dt>Nearby shelters</dt>
                <dd>
                  <AnimatedCounter value={shelterSupport.shelters.length} />
                </dd>
              </div>
            </dl>
          </div>

          <Paper className="surface risk-checker" elevation={0}>
            <Box
              component="form"
              className="risk-lookup"
              onSubmit={submitRiskLookup}
            >
              <TextField
                id="risk-area"
                label="Area or coordinates"
                value={area}
                onChange={(event) => setArea(event.target.value)}
                fullWidth
                error={riskFormInvalid}
                helperText={riskFormInvalid ? riskState.message : ""}
                FormHelperTextProps={{ id: "risk-area-error" }}
                inputProps={{ "aria-describedby": "risk-area-error" }}
              />
              <Button
                type="submit"
                variant="contained"
                startIcon={
                  riskState.status === "loading" ? (
                    <Loader2 size={18} className="spin-icon" />
                  ) : (
                    <MapPin size={18} />
                  )
                }
                disabled={riskState.status === "loading"}
              >
                {riskState.status === "loading"
                  ? riskState.message
                  : "Check risk"}
              </Button>
              <Button
                type="button"
                variant="outlined"
                startIcon={<LocateFixed size={18} />}
                onClick={useRiskLocation}
                disabled={riskState.status === "loading"}
              >
                Use location
              </Button>
            </Box>
            <Stack
              direction="row"
              spacing={1}
              flexWrap="wrap"
              className="risk-presets"
            >
              {areaPresets.map((preset) => (
                <Chip
                  key={preset.label}
                  label={preset.label}
                  onClick={() =>
                    void fetchRisk(preset.lat, preset.lng, preset.label)
                  }
                  variant={area === preset.label ? "filled" : "outlined"}
                  color={area === preset.label ? "warning" : "default"}
                />
              ))}
            </Stack>
            {riskFormInvalid ? (
              <Alert
                id="risk-form-error"
                severity={
                  riskState.status === "permission-denied" ? "warning" : "error"
                }
                className="warning-alert"
              >
                {riskState.message}
              </Alert>
            ) : null}
            {riskState.status === "idle" && riskState.message ? (
              <Alert severity="success" className="warning-alert">
                {riskState.message}
              </Alert>
            ) : null}

            <Box className="risk-score">
              <Typography variant="overline">Overall risk</Typography>
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                flexWrap="wrap"
              >
                <Typography variant="h2">{risk.overallRisk}</Typography>
                <Chip
                  label={`${overallSeverityRole} severity`}
                  sx={{
                    backgroundColor:
                      severityRoles[overallSeverityRole].background,
                    color: overallSeverityColor,
                    borderColor: severityRoles[overallSeverityRole].border,
                    borderWidth: 1,
                    borderStyle: "solid",
                    fontWeight: 700,
                    textTransform: "capitalize",
                  }}
                />
              </Stack>
              <Typography color="text.secondary">
                {risk.location} is currently reporting{" "}
                {floodRisk?.level ?? risk.overallRisk} flood risk.
              </Typography>
            </Box>

            {hasMapCoords ? (
              <div className="risk-map-frame">
                <RiskMap
                  lat={mapLat}
                  lng={mapLng}
                  riskColor={overallSeverityColor}
                  markers={shelterMarkers}
                  ariaLabel={`Map of ${risk.location} showing the selected area and ${shelterMarkers.length} nearby shelters`}
                />
                <span className="risk-map-frame__coords">
                  {riskCoordinates.lat}, {riskCoordinates.lng}
                </span>
              </div>
            ) : null}
          </Paper>
        </div>
      </section>

      <div className="citizen-shell">
        {/* Risk breakdown */}
        <Reveal className="citizen-section">
          <section aria-label="Risk breakdown">
            <PageHeader
              icon={ShieldCheck}
              title="What's driving the risk"
              subtitle={`Hazard signals for ${risk.location}.`}
              tone="gold"
            />
            <Grid container spacing={2} className="stagger">
              {risk.risks.length > 0 ? (
                risk.risks.map((item) => (
                  <Grid size={{ xs: 12, md: 6 }} key={item.type}>
                    <Paper variant="outlined" className="risk-row">
                      <Stack
                        direction="row"
                        spacing={1.5}
                        alignItems="flex-start"
                      >
                        <ShieldCheck
                          size={22}
                          color={
                            hazardRoles[hazardRoleFor(item.type)].foreground
                          }
                        />
                        <Box>
                          <Stack
                            direction="row"
                            spacing={1}
                            alignItems="center"
                            flexWrap="wrap"
                          >
                            <Typography variant="subtitle1">
                              {item.type.replace("_", " ")}
                            </Typography>
                            {(() => {
                              const role = severityRoleFor(item.level);
                              const Icon = severityIcons[role];
                              return (
                                <Chip
                                  size="small"
                                  icon={
                                    <Icon
                                      size={16}
                                      color={severityRoles[role].foreground}
                                    />
                                  }
                                  label={item.level}
                                  sx={{
                                    backgroundColor:
                                      severityRoles[role].background,
                                    color: severityRoles[role].foreground,
                                    borderColor: severityRoles[role].border,
                                    borderWidth: 1,
                                    borderStyle: "solid",
                                    fontWeight: 600,
                                    ".MuiChip-icon": {
                                      color: severityRoles[role].foreground,
                                    },
                                  }}
                                />
                              );
                            })()}
                            {item.probability ? (
                              <Chip
                                size="small"
                                variant="outlined"
                                label={`${Math.round(item.probability * 100)}%`}
                              />
                            ) : null}
                          </Stack>
                          <Typography variant="body2" color="text.secondary">
                            {item.reason}
                          </Typography>
                        </Box>
                      </Stack>
                    </Paper>
                  </Grid>
                ))
              ) : (
                <Grid size={{ xs: 12 }}>
                  <Alert severity="info" className="warning-alert">
                    No active risk records were returned for this area.
                  </Alert>
                </Grid>
              )}
            </Grid>
          </section>
        </Reveal>
      </div>
    </>
  );
}

function shelterPayloadFromRisk(
  riskPayload: AreaRiskResponse,
): NearbyShelterResponse {
  const generatedAt = new Date().toISOString();

  return {
    generatedAt,
    shelters: riskPayload.nearestShelters.map((shelter) => ({
      id: shelter.id,
      name: shelter.name,
      type: "temporary_shelter",
      region: "Greater Accra",
      district: riskPayload.location,
      address: riskPayload.location,
      location: shelter.location,
      capacity: shelter.capacity ?? 0,
      currentOccupancy: shelter.currentOccupancy ?? 0,
      status: shelter.status ?? "unknown",
      contact: shelter.contact ?? "112",
      facilities: shelter.facilities ?? [],
      distanceMeters: shelter.distanceMeters,
      updatedAt: generatedAt,
    })),
    recoverySupport: sampleShelterResponse.recoverySupport,
  };
}

export default HomePage;
