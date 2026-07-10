import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  AppBar,
  Box,
  Button,
  ButtonGroup,
  Chip,
  Container,
  CssBaseline,
  Divider,
  FormControl,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  ThemeProvider,
  Toolbar,
  Typography,
} from "@mui/material";
import {
  AlertOctagon,
  AlertTriangle,
  Bell,
  BookOpen,
  CheckCircle2,
  Clock3,
  Cross,
  ImagePlus,
  Info,
  Languages,
  LifeBuoy,
  Loader2,
  LocateFixed,
  MapPin,
  Megaphone,
  Phone,
  RefreshCw,
  ShieldCheck,
  Siren,
  TriangleAlert,
  Waves,
  WifiOff,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
  CreateIncidentRequest,
  CreateIncidentResponse,
  EmergencyGuideRecord,
  GuideListResponse,
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
  NearbyShelterResponse,
  ReliefPointNearbyResponse,
  RoadClosureListResponse,
  RoadClosureRecord,
} from "@nadaa/shared-types";
import {
  GUIDE_API_BASE,
  INCIDENT_API_BASE,
  NOTIFICATION_API_BASE,
  RISK_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "../../app/config";
import { citizenTheme } from "../../app/theme";
import {
  areaPresets,
  buildFallbackAlerts,
  buildFallbackGuides,
  guideHazardOptions,
  guideLanguageOptions,
  guideStageOptions,
  hazardOptions,
  initialReportForm,
  mediaSizeLimits,
  sampleReliefPointResponse,
  sampleRisk,
  sampleShelterResponse,
  supportedMediaTypes,
  urgencyOptions,
} from "./data";

import type {
  AlertFeedState,
  AlertFeedView,
  GuideCacheInfo,
  GuideFilters,
  GuideHazardFilter,
  GuideStageFilter,
  GuideState,
  ReportForm,
  ReportState,
  RiskState,
  ShelterState,
} from "./types";
import {
  alertSeverityLabel,
  alertSeverityTone,
  alertStatusLabel,
  extractAPIError,
  filterGuides,
  formatDateTime,
  formatDistance,
  formatFileSize,
  formatListLabel,
  formatOccupancy,
  formatReliefStock,
  formatSupportType,
  guideLanguageLabel,
  guideStageLabel,
  hazardLabel,
  hazardRoleFor,
  hazardRoles,
  initiateMediaUploads,
  readGuideCache,
  registerCitizenServiceWorker,
  resolveAreaLookup,
  severityRoleFor,
  severityRoles,
  writeGuideCache,
} from "./utils";
import RoutePlanner from "./RoutePlanner";
import DonorPortal from "./DonorPortal";
import MissingPersonsPanel from "./MissingPersonsPanel";
import DamageClaim from "./DamageClaim";
import PublicCampaignsPanel from "./PublicCampaignsPanel";
import { OpenDataPortal } from "./OpenDataPortal";

function CitizenApp() {
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
  const [reliefPoints, setReliefPoints] = useState<ReliefPointNearbyResponse>(
    sampleReliefPointResponse,
  );
  const [roadClosures, setRoadClosures] = useState<RoadClosureRecord[]>([]);
  const [shelterState, setShelterState] = useState<ShelterState>({
    status: "idle",
    message: "Shelter and recovery support fixtures are ready.",
  });
  const [alertFeed, setAlertFeed] = useState<CitizenAlertFeedItem[]>(() =>
    buildFallbackAlerts(),
  );
  const [alertFeedView, setAlertFeedView] = useState<AlertFeedView>("current");
  const [alertFeedState, setAlertFeedState] = useState<AlertFeedState>({
    status: "idle",
    message: "Showing saved warnings until the feed refreshes.",
  });
  const [guideFilters, setGuideFilters] = useState<GuideFilters>({
    hazard: "all",
    stage: "during",
    language: "en",
  });
  const [guides, setGuides] = useState<EmergencyGuideRecord[]>(() => {
    const cached = readGuideCache();
    return cached?.guides ?? buildFallbackGuides();
  });
  const [guideCacheInfo, setGuideCacheInfo] = useState<GuideCacheInfo>(() => {
    const cached = readGuideCache();
    return cached
      ? {
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        }
      : {
          cachedAt: new Date().toISOString(),
          source: "fixture",
          language: "en",
        };
  });
  const [guideState, setGuideState] = useState<GuideState>({
    status: "idle",
    message: "Offline guides are ready.",
  });
  const [reportForm, setReportForm] = useState<ReportForm>(initialReportForm);
  const [reportState, setReportState] = useState<ReportState>({
    status: "idle",
  });
  const [reportErrors, setReportErrors] = useState<
    Partial<Record<keyof ReportForm, string>>
  >({});

  const severityIcons = {
    low: CheckCircle2,
    medium: AlertTriangle,
    high: AlertTriangle,
    severe: AlertOctagon,
    info: Info,
  };

  const clearReportError = (key: keyof ReportForm) => {
    setReportErrors((current) => {
      if (!current[key]) return current;
      const next = { ...current };
      delete next[key];
      return next;
    });
  };
  const floodRisk = useMemo(
    () => risk.risks.find((item) => item.type === "flood"),
    [risk.risks],
  );
  const visibleAlerts = useMemo(
    () =>
      alertFeed.filter((alert) => {
        if (alertFeedView === "all") {
          return true;
        }
        return alert.status === alertFeedView;
      }),
    [alertFeed, alertFeedView],
  );
  const currentAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "current").length,
    [alertFeed],
  );
  const expiredAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "expired").length,
    [alertFeed],
  );
  const visibleGuides = useMemo(
    () => filterGuides(guides, guideFilters),
    [guides, guideFilters],
  );
  const featuredGuide = visibleGuides[0];
  const offlineGuideCount = useMemo(
    () => guides.filter((guide) => guide.offlineAvailable).length,
    [guides],
  );

  useEffect(() => {
    void fetchRisk(
      areaPresets[0].lat,
      areaPresets[0].lng,
      areaPresets[0].label,
    );
    void fetchAlertFeed();
    void fetchGuides(guideFilters.language);
    registerCitizenServiceWorker();
  }, []);

  const updateReportForm = <Key extends keyof ReportForm>(
    key: Key,
    value: ReportForm[Key],
  ) => {
    setReportForm((current) => ({ ...current, [key]: value }));
  };

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
      void fetchReliefPoints(lat, lng);
      void fetchRoadClosures(lat, lng);
    } catch (error) {
      setRiskState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not load area risk.",
      });
    }
  }

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

  async function fetchReliefPoints(lat: number, lng: number) {
    if (!navigator.onLine) {
      setReliefPoints(sampleReliefPointResponse);
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
      setReliefPoints(
        payload.reliefPoints.length ? payload : sampleReliefPointResponse,
      );
    } catch {
      setReliefPoints(sampleReliefPointResponse);
    }
  }

  async function fetchShelters(
    lat: number,
    lng: number,
    riskPayload: AreaRiskResponse = risk,
  ) {
    if (!navigator.onLine) {
      setShelterSupport(shelterPayloadFromRisk(riskPayload));
      setShelterState({
        status: "fallback",
        message: "Shelter lookup needs a connection. Showing saved resources.",
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
    } catch (error) {
      setShelterSupport(shelterPayloadFromRisk(riskPayload));
      setShelterState({
        status: "fallback",
        message:
          error instanceof Error
            ? `Shelter service unavailable. ${error.message}`
            : "Shelter service unavailable. Showing saved resources.",
      });
    }
  }

  async function fetchAlertFeed() {
    if (!navigator.onLine) {
      setAlertFeedState({
        status: "error",
        message: "Alert feed needs a connection. Showing saved warnings.",
      });
      return;
    }

    setAlertFeedState({ status: "loading", message: "Refreshing alerts" });

    try {
      const response = await fetch(
        `${NOTIFICATION_API_BASE}/notifications/alerts?includeExpired=true`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as CitizenAlertFeedResponse;
      setAlertFeed(
        payload.alerts.length > 0 ? payload.alerts : buildFallbackAlerts(),
      );
      setAlertFeedState({
        status: "idle",
        message: `Alert feed updated ${formatDateTime(payload.generatedAt)}.`,
      });
    } catch (error) {
      setAlertFeed(buildFallbackAlerts());
      setAlertFeedState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Live alert feed unavailable. Showing saved warnings.",
      });
    }
  }

  async function fetchGuides(language = guideFilters.language) {
    if (!navigator.onLine) {
      const cached = readGuideCache();
      if (cached?.guides.length) {
        setGuides(cached.guides);
        setGuideCacheInfo({
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        });
        setGuideState({
          status: "offline",
          message: `Offline guide cache ready from ${formatDateTime(cached.cachedAt)}.`,
        });
        return;
      }

      setGuides(buildFallbackGuides());
      setGuideCacheInfo({
        cachedAt: new Date().toISOString(),
        source: "fixture",
        language: "en",
      });
      setGuideState({
        status: "offline",
        message: "Showing starter emergency guides until connection returns.",
      });
      return;
    }

    setGuideState({ status: "loading", message: "Refreshing guides" });

    try {
      const params = new URLSearchParams({ language, offline: "true" });

      const response = await fetch(`${GUIDE_API_BASE}/guides?${params}`);
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as GuideListResponse;
      const nextGuides = payload.guides.length
        ? payload.guides
        : buildFallbackGuides();
      const cachedAt = new Date().toISOString();
      setGuides(nextGuides);
      setGuideCacheInfo({ cachedAt, source: "network", language });
      writeGuideCache(nextGuides, language, cachedAt);
      setGuideState({
        status: "idle",
        message: `Saved ${nextGuides.length} offline guides for ${guideLanguageLabel(language)}.`,
      });
    } catch (error) {
      const cached = readGuideCache();
      if (cached?.guides.length) {
        setGuides(cached.guides);
        setGuideCacheInfo({
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        });
        setGuideState({
          status: "offline",
          message: `Live guide service unavailable. Using offline cache from ${formatDateTime(cached.cachedAt)}.`,
        });
        return;
      }

      setGuides(buildFallbackGuides());
      setGuideCacheInfo({
        cachedAt: new Date().toISOString(),
        source: "fixture",
        language: "en",
      });
      setGuideState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Could not load emergency guides.",
      });
    }
  }

  function updateGuideLanguage(language: string) {
    setGuideFilters((current) => ({ ...current, language }));
    void fetchGuides(language);
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

    void fetchShelters(lat, lng, risk);
    void fetchReliefPoints(lat, lng);
  };

  const useCurrentLocation = () => {
    if (!navigator.geolocation) {
      setReportState({
        status: "error",
        message: "Location is not available on this device.",
      });
      return;
    }

    setReportState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setReportForm((current) => ({
          ...current,
          lat: position.coords.latitude.toFixed(6),
          lng: position.coords.longitude.toFixed(6),
        }));
        setReportState({ status: "idle" });
      },
      () => {
        setReportState({
          status: "error",
          message: "Location permission was not granted.",
        });
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  };

  const handleFileSelection = (event: ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(event.target.files ?? []);
    event.currentTarget.value = "";

    if (selectedFiles.length > 10) {
      setReportState({
        status: "error",
        message: "Attach at most 10 media files to one report.",
      });
      return;
    }

    const invalidFile = selectedFiles.find((file) => {
      if (
        !supportedMediaTypes.includes(file.type as IncidentMediaContentType)
      ) {
        return true;
      }

      return (
        file.size <= 0 ||
        file.size > mediaSizeLimits[file.type as IncidentMediaContentType]
      );
    });

    if (invalidFile) {
      setReportState({
        status: "error",
        message: `${invalidFile.name} is not supported or is too large for this report.`,
      });
      return;
    }

    updateReportForm("files", selectedFiles);
    setReportState({ status: "idle" });
  };

  const submitReport = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const lat = Number(reportForm.lat);
    const lng = Number(reportForm.lng);
    const peopleAffected = Number(reportForm.peopleAffected || 0);
    const nextErrors: Partial<Record<keyof ReportForm, string>> = {};

    if (
      !Number.isFinite(lat) ||
      lat < -90 ||
      lat > 90 ||
      !Number.isFinite(lng) ||
      lng < -180 ||
      lng > 180
    ) {
      nextErrors.lat = "Enter a valid latitude.";
      nextErrors.lng = "Enter a valid longitude.";
    }

    if (reportForm.description.trim().length < 5) {
      nextErrors.description = "Add a short description of what happened.";
    }

    if (!Number.isInteger(peopleAffected) || peopleAffected < 0) {
      nextErrors.peopleAffected = "People affected must be zero or more.";
    }

    if (Object.keys(nextErrors).length > 0) {
      setReportErrors(nextErrors);
      setReportState({
        status: "error",
        message: "Please correct the highlighted fields before sending.",
      });
      return;
    }

    if (!navigator.onLine) {
      setReportState({
        status: "error",
        message:
          "You appear to be offline. Keep this report open and try again when the connection returns.",
      });
      return;
    }

    setReportState({ status: "loading", message: "Sending report" });

    try {
      const mediaIds = await initiateMediaUploads(reportForm.files);
      const payload: CreateIncidentRequest = {
        type: reportForm.hazard,
        description: reportForm.description.trim(),
        location: { lat, lng },
        peopleAffected,
        injuriesReported: reportForm.injuriesReported,
        urgency: reportForm.urgency,
        anonymous: reportForm.anonymous,
        contactPermission: reportForm.anonymous
          ? false
          : reportForm.contactPermission,
        accessibilityNeeds: reportForm.accessibilityNeeds.trim() || undefined,
        media: mediaIds,
        reporter: reportForm.anonymous
          ? undefined
          : {
              userId: "usr_demo_citizen",
              phone: reportForm.contactPermission ? "+233200000000" : undefined,
            },
      };

      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const incident = (await response.json()) as CreateIncidentResponse;
      setReportState({
        status: "success",
        reference: incident.reference,
        priorityReview: incident.priorityReview,
      });
      setReportErrors({});
      setReportForm(initialReportForm);
    } catch (error) {
      setReportState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not send report.",
      });
    }
  };

  return (
    <ThemeProvider theme={citizenTheme}>
      <CssBaseline />
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <AppBar position="sticky" elevation={0} className="topbar">
        <Toolbar className="toolbar">
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Box
              component="img"
              src="/brand/nadaa-logo.png"
              alt="NADAA shield"
              className="brand-logo"
            />
            <Box>
              <Typography
                variant="h6"
                component="div"
                className="brand-wordmark"
              >
                {nadaaBrand.name}
              </Typography>
              <Typography variant="caption" className="brand-subtitle">
                {nadaaBrand.slogan}
              </Typography>
            </Box>
          </Stack>
          <Button
            color="inherit"
            variant="outlined"
            startIcon={<Phone size={18} />}
            className="call-button"
          >
            Call 112
          </Button>
        </Toolbar>
      </AppBar>

      <Box component="main" id="main-content">
        <Container maxWidth="xl" className="app-shell">
          <Grid container spacing={2.5}>
            <Grid size={{ xs: 12, lg: 8 }}>
              <Paper className="surface risk-surface">
                <Stack
                  direction={{ xs: "column", md: "row" }}
                  spacing={2}
                  justifyContent="space-between"
                >
                  <Box>
                    <Typography variant="overline" color="secondary">
                      Citizen operations
                    </Typography>
                    <Typography variant="h4">
                      Know your risk before conditions change
                    </Typography>
                  </Box>
                  <ButtonGroup
                    variant="contained"
                    aria-label="risk view selector"
                    className="mode-group"
                  >
                    <Button startIcon={<Waves size={17} />}>Risk</Button>
                    <Button startIcon={<Bell size={17} />}>Alerts</Button>
                    <Button startIcon={<Siren size={17} />}>Report</Button>
                  </ButtonGroup>
                </Stack>

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
                    error={
                      riskState.status === "error" ||
                      riskState.status === "permission-denied"
                    }
                    helperText={
                      riskState.status === "error" ||
                      riskState.status === "permission-denied"
                        ? riskState.message
                        : ""
                    }
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
                      color={area === preset.label ? "secondary" : "default"}
                    />
                  ))}
                </Stack>
                {riskState.status === "error" ||
                riskState.status === "permission-denied" ? (
                  <Alert
                    id="risk-form-error"
                    severity={
                      riskState.status === "permission-denied"
                        ? "warning"
                        : "error"
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

                <Grid container spacing={2}>
                  <Grid size={{ xs: 12, md: 5 }}>
                    <Box className="risk-score">
                      <Typography variant="overline">Overall risk</Typography>
                      <Stack
                        direction="row"
                        spacing={1}
                        alignItems="center"
                        flexWrap="wrap"
                      >
                        <Typography variant="h2">{risk.overallRisk}</Typography>
                        <Chip label="Rainfall rising" color="warning" />
                      </Stack>
                      <Typography color="text.secondary">
                        {risk.location} is currently reporting{" "}
                        {floodRisk?.level ?? risk.overallRisk} flood risk.
                      </Typography>
                      <Box
                        className="risk-map-preview"
                        aria-label="Selected risk coordinates"
                      >
                        <Typography variant="caption" color="text.secondary">
                          {riskCoordinates.lat}, {riskCoordinates.lng}
                        </Typography>
                        <Box className="risk-map-point" />
                      </Box>
                    </Box>
                  </Grid>
                  <Grid size={{ xs: 12, md: 7 }}>
                    <Stack spacing={1.5}>
                      {risk.risks.length > 0 ? (
                        risk.risks.map((item) => (
                          <Paper
                            variant="outlined"
                            className="risk-row"
                            key={item.type}
                          >
                            <Stack
                              direction="row"
                              spacing={1.5}
                              alignItems="flex-start"
                            >
                              <ShieldCheck
                                size={22}
                                color={
                                  hazardRoles[hazardRoleFor(item.type)]
                                    .foreground
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
                                            color={
                                              severityRoles[role].foreground
                                            }
                                          />
                                        }
                                        label={item.level}
                                        sx={{
                                          backgroundColor:
                                            severityRoles[role].background,
                                          color: severityRoles[role].foreground,
                                          borderColor:
                                            severityRoles[role].border,
                                          borderWidth: 1,
                                          borderStyle: "solid",
                                          fontWeight: 600,
                                          ".MuiChip-icon": {
                                            color:
                                              severityRoles[role].foreground,
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
                                <Typography
                                  variant="body2"
                                  color="text.secondary"
                                >
                                  {item.reason}
                                </Typography>
                              </Box>
                            </Stack>
                          </Paper>
                        ))
                      ) : (
                        <Alert severity="info" className="warning-alert">
                          No active risk records were returned for this area.
                        </Alert>
                      )}
                    </Stack>
                  </Grid>
                </Grid>
              </Paper>

              <Grid container spacing={2.5} className="section-grid">
                <Grid size={{ xs: 12, md: 6 }}>
                  <Paper className="surface">
                    <Stack
                      direction={{ xs: "column", sm: "row" }}
                      spacing={1}
                      justifyContent="space-between"
                      alignItems={{ xs: "stretch", sm: "center" }}
                      className="section-heading"
                    >
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Megaphone size={21} color={nadaaBrand.colors.red} />
                        <Box>
                          <Typography variant="h6">Live warnings</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {currentAlertCount} current · {expiredAlertCount}{" "}
                            expired
                          </Typography>
                        </Box>
                      </Stack>
                      <Button
                        type="button"
                        variant="outlined"
                        size="small"
                        startIcon={
                          alertFeedState.status === "loading" ? (
                            <Loader2 size={16} className="spin-icon" />
                          ) : (
                            <RefreshCw size={16} />
                          )
                        }
                        onClick={() => void fetchAlertFeed()}
                        disabled={alertFeedState.status === "loading"}
                      >
                        Refresh
                      </Button>
                    </Stack>
                    <ButtonGroup
                      variant="outlined"
                      size="small"
                      className="alert-filter-group"
                      aria-label="alert feed filter"
                    >
                      <Button
                        variant={
                          alertFeedView === "current" ? "contained" : "outlined"
                        }
                        onClick={() => setAlertFeedView("current")}
                      >
                        Current
                      </Button>
                      <Button
                        variant={
                          alertFeedView === "expired" ? "contained" : "outlined"
                        }
                        onClick={() => setAlertFeedView("expired")}
                      >
                        Expired
                      </Button>
                      <Button
                        variant={
                          alertFeedView === "all" ? "contained" : "outlined"
                        }
                        onClick={() => setAlertFeedView("all")}
                      >
                        All
                      </Button>
                    </ButtonGroup>

                    {alertFeedState.status === "error" ? (
                      <Alert severity="warning" className="warning-alert">
                        {alertFeedState.message}
                      </Alert>
                    ) : null}
                    {alertFeedState.status === "idle" &&
                    alertFeedState.message ? (
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        className="alert-feed-note"
                      >
                        {alertFeedState.message}
                      </Typography>
                    ) : null}

                    <Stack spacing={1.5}>
                      {visibleAlerts.length > 0 ? (
                        visibleAlerts.map((alert) => (
                          <Alert
                            key={alert.id}
                            severity={alertSeverityTone(
                              alert.severity,
                              alert.status,
                            )}
                            className="warning-alert citizen-alert-card"
                            icon={
                              alert.status === "expired" ? (
                                <Clock3 size={20} />
                              ) : undefined
                            }
                          >
                            <Stack spacing={0.75}>
                              <Stack
                                direction="row"
                                spacing={1}
                                justifyContent="space-between"
                                alignItems="flex-start"
                              >
                                <Box>
                                  <Typography variant="subtitle2">
                                    {alert.title}
                                  </Typography>
                                  <Typography variant="body2">
                                    {alert.targetLabel} ·{" "}
                                    {alertSeverityLabel(alert.severity)}
                                  </Typography>
                                </Box>
                                <Chip
                                  size="small"
                                  icon={
                                    alert.status === "current" ? (
                                      <AlertOctagon size={16} />
                                    ) : (
                                      <Clock3 size={16} />
                                    )
                                  }
                                  label={alertStatusLabel(alert.status)}
                                  color={
                                    alert.status === "current"
                                      ? "error"
                                      : "default"
                                  }
                                />
                              </Stack>
                              <Typography variant="body2">
                                {alert.message}
                              </Typography>
                              <Typography variant="body2">
                                {alert.recommendedAction}
                              </Typography>
                              <Stack
                                direction="row"
                                spacing={0.75}
                                flexWrap="wrap"
                              >
                                {(() => {
                                  const hRole = hazardRoleFor(alert.hazardType);
                                  return (
                                    <Chip
                                      size="small"
                                      variant="outlined"
                                      label={hazardLabel(alert.hazardType)}
                                      sx={{
                                        borderColor: hazardRoles[hRole].border,
                                        color: hazardRoles[hRole].foreground,
                                        backgroundColor:
                                          hazardRoles[hRole].background,
                                      }}
                                    />
                                  );
                                })()}
                                <Chip
                                  size="small"
                                  variant="outlined"
                                  label={`Until ${formatDateTime(alert.expiresAt)}`}
                                />
                                {alert.evacuationRequired ? (
                                  <Chip
                                    size="small"
                                    color="error"
                                    label="Evacuation possible"
                                  />
                                ) : null}
                              </Stack>
                            </Stack>
                          </Alert>
                        ))
                      ) : (
                        <Alert severity="info" className="warning-alert">
                          No {alertFeedView === "all" ? "" : alertFeedView}{" "}
                          alerts are available.
                        </Alert>
                      )}
                    </Stack>
                  </Paper>
                </Grid>

                <Grid size={{ xs: 12, md: 6 }}>
                  <RoutePlanner />
                </Grid>

                <Grid size={{ xs: 12, md: 6 }}>
                  <DonorPortal />
                  <DamageClaim />
                </Grid>

                <Grid size={{ xs: 12 }}>
                  <MissingPersonsPanel />
                </Grid>

                <Grid size={{ xs: 12 }}>
                  <OpenDataPortal />
                </Grid>

                <Grid size={{ xs: 12, md: 6 }}>
                  <Paper className="surface report-surface">
                    <Stack
                      direction="row"
                      spacing={1}
                      alignItems="center"
                      className="section-heading"
                    >
                      <Siren size={21} color={nadaaBrand.colors.gold} />
                      <Typography variant="h6">Report incident</Typography>
                    </Stack>
                    <Stack
                      component="form"
                      spacing={1.5}
                      onSubmit={submitReport}
                      noValidate
                    >
                      <FormControl
                        fullWidth
                        error={Boolean(reportErrors.hazard)}
                      >
                        <InputLabel id="report-hazard-label">
                          Hazard type
                        </InputLabel>
                        <Select
                          id="report-hazard"
                          labelId="report-hazard-label"
                          value={reportForm.hazard}
                          label="Hazard type"
                          onChange={(event) => {
                            clearReportError("hazard");
                            updateReportForm(
                              "hazard",
                              event.target.value as HazardType,
                            );
                          }}
                          aria-describedby="report-hazard-error"
                        >
                          {hazardOptions.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                              {option.label}
                            </MenuItem>
                          ))}
                        </Select>
                        {reportErrors.hazard ? (
                          <FormHelperText id="report-hazard-error">
                            {reportErrors.hazard}
                          </FormHelperText>
                        ) : null}
                      </FormControl>
                      <Grid container spacing={1.25}>
                        <Grid size={{ xs: 6 }}>
                          <TextField
                            id="report-lat"
                            label="Latitude"
                            value={reportForm.lat}
                            onChange={(event) => {
                              clearReportError("lat");
                              updateReportForm("lat", event.target.value);
                            }}
                            fullWidth
                            inputMode="decimal"
                            error={Boolean(reportErrors.lat)}
                            helperText={reportErrors.lat}
                            FormHelperTextProps={{ id: "report-lat-error" }}
                            inputProps={{
                              "aria-describedby": "report-lat-error",
                            }}
                          />
                        </Grid>
                        <Grid size={{ xs: 6 }}>
                          <TextField
                            id="report-lng"
                            label="Longitude"
                            value={reportForm.lng}
                            onChange={(event) => {
                              clearReportError("lng");
                              updateReportForm("lng", event.target.value);
                            }}
                            fullWidth
                            inputMode="decimal"
                            error={Boolean(reportErrors.lng)}
                            helperText={reportErrors.lng}
                            FormHelperTextProps={{ id: "report-lng-error" }}
                            inputProps={{
                              "aria-describedby": "report-lng-error",
                            }}
                          />
                        </Grid>
                      </Grid>
                      <Button
                        type="button"
                        variant="outlined"
                        startIcon={<LocateFixed size={18} />}
                        onClick={useCurrentLocation}
                        disabled={reportState.status === "loading"}
                      >
                        Use GPS
                      </Button>
                      <TextField
                        id="report-description"
                        label="What happened?"
                        value={reportForm.description}
                        onChange={(event) => {
                          clearReportError("description");
                          updateReportForm("description", event.target.value);
                        }}
                        multiline
                        minRows={3}
                        error={Boolean(reportErrors.description)}
                        helperText={reportErrors.description}
                        FormHelperTextProps={{ id: "report-description-error" }}
                        inputProps={{
                          maxLength: 2000,
                          "aria-describedby": "report-description-error",
                        }}
                      />
                      <Grid container spacing={1.25}>
                        <Grid size={{ xs: 6 }}>
                          <TextField
                            id="report-people-affected"
                            label="People affected"
                            value={reportForm.peopleAffected}
                            onChange={(event) => {
                              clearReportError("peopleAffected");
                              updateReportForm(
                                "peopleAffected",
                                event.target.value,
                              );
                            }}
                            fullWidth
                            inputMode="numeric"
                            error={Boolean(reportErrors.peopleAffected)}
                            helperText={reportErrors.peopleAffected}
                            FormHelperTextProps={{
                              id: "report-people-affected-error",
                            }}
                            inputProps={{
                              "aria-describedby":
                                "report-people-affected-error",
                            }}
                          />
                        </Grid>
                        <Grid size={{ xs: 6 }}>
                          <FormControl
                            fullWidth
                            error={Boolean(reportErrors.urgency)}
                          >
                            <InputLabel id="report-urgency-label">
                              Urgency
                            </InputLabel>
                            <Select
                              id="report-urgency"
                              labelId="report-urgency-label"
                              value={reportForm.urgency}
                              label="Urgency"
                              onChange={(event) => {
                                clearReportError("urgency");
                                updateReportForm(
                                  "urgency",
                                  event.target.value as IncidentUrgency,
                                );
                              }}
                              aria-describedby="report-urgency-error"
                            >
                              {urgencyOptions.map((option) => (
                                <MenuItem
                                  key={option.value}
                                  value={option.value}
                                >
                                  {option.label}
                                </MenuItem>
                              ))}
                            </Select>
                            {reportErrors.urgency ? (
                              <FormHelperText id="report-urgency-error">
                                {reportErrors.urgency}
                              </FormHelperText>
                            ) : null}
                          </FormControl>
                        </Grid>
                      </Grid>
                      {reportForm.urgency === "life_threatening" ? (
                        <Alert severity="error" className="warning-alert">
                          <Typography variant="body2">
                            Call 112 immediately after sending this report.
                          </Typography>
                        </Alert>
                      ) : null}
                      <TextField
                        id="report-accessibility-needs"
                        label="Accessibility needs"
                        value={reportForm.accessibilityNeeds}
                        onChange={(event) =>
                          updateReportForm(
                            "accessibilityNeeds",
                            event.target.value,
                          )
                        }
                        inputProps={{ maxLength: 500 }}
                      />
                      <Button
                        component="label"
                        variant="outlined"
                        startIcon={<ImagePlus size={18} />}
                      >
                        Add media
                        <input
                          type="file"
                          hidden
                          multiple
                          accept={supportedMediaTypes.join(",")}
                          onChange={handleFileSelection}
                        />
                      </Button>
                      {reportForm.files.length > 0 ? (
                        <Stack spacing={0.75}>
                          {reportForm.files.map((file) => (
                            <Chip
                              key={`${file.name}-${file.size}`}
                              label={`${file.name} · ${formatFileSize(file.size)}`}
                              className="media-chip"
                            />
                          ))}
                        </Stack>
                      ) : null}
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        alignItems="center"
                      >
                        <Typography>Injuries reported</Typography>
                        <Switch
                          checked={reportForm.injuriesReported}
                          onChange={(event) =>
                            updateReportForm(
                              "injuriesReported",
                              event.target.checked,
                            )
                          }
                        />
                      </Stack>
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        alignItems="center"
                      >
                        <Typography>Report anonymously</Typography>
                        <Switch
                          checked={reportForm.anonymous}
                          onChange={(event) =>
                            setReportForm((current) => ({
                              ...current,
                              anonymous: event.target.checked,
                              contactPermission: event.target.checked
                                ? false
                                : current.contactPermission,
                            }))
                          }
                        />
                      </Stack>
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        alignItems="center"
                      >
                        <Typography>Allow contact</Typography>
                        <Switch
                          checked={
                            !reportForm.anonymous &&
                            reportForm.contactPermission
                          }
                          onChange={(event) =>
                            updateReportForm(
                              "contactPermission",
                              event.target.checked,
                            )
                          }
                          disabled={reportForm.anonymous}
                        />
                      </Stack>
                      <Alert severity="info" className="warning-alert">
                        NADAA uses report location to route emergency response,
                        detect duplicates, and coordinate verified authority
                        actions. Anonymous reports hide your identity; disabling
                        contact means responders cannot call you back through
                        this report.
                      </Alert>
                      {reportState.status === "error" ? (
                        <Alert severity="error" className="warning-alert">
                          {reportState.message}
                        </Alert>
                      ) : null}
                      {reportState.status === "success" ? (
                        <Alert
                          severity={
                            reportState.priorityReview ? "warning" : "success"
                          }
                          className="warning-alert"
                        >
                          <Typography variant="subtitle2">
                            Report {reportState.reference} received
                          </Typography>
                          <Typography variant="body2">
                            Call 112 if anyone is in immediate danger.
                          </Typography>
                        </Alert>
                      ) : null}
                      <Button
                        type="submit"
                        variant="contained"
                        color="error"
                        disabled={reportState.status === "loading"}
                        startIcon={
                          reportState.status === "loading" ? (
                            <Loader2 size={18} className="spin-icon" />
                          ) : (
                            <Siren size={18} />
                          )
                        }
                      >
                        {reportState.status === "loading"
                          ? reportState.message
                          : "Send report"}
                      </Button>
                    </Stack>
                  </Paper>
                </Grid>
              </Grid>
            </Grid>

            <Grid size={{ xs: 12, lg: 4 }}>
              <Stack spacing={2.5}>
                <Paper className="surface emergency-card">
                  <Stack direction="row" spacing={1.5} alignItems="center">
                    <LifeBuoy size={26} />
                    <Box>
                      <Typography variant="h6">Emergency help</Typography>
                      <Typography variant="body2">
                        Police, fire, ambulance, NADMO and relief agencies.
                      </Typography>
                    </Box>
                  </Stack>
                  <Button
                    fullWidth
                    variant="contained"
                    color="error"
                    startIcon={<Phone size={18} />}
                  >
                    Call 112 now
                  </Button>
                </Paper>

                <Paper className="surface">
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={1}
                    justifyContent="space-between"
                    alignItems={{ xs: "stretch", sm: "center" }}
                    className="section-heading"
                  >
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Cross size={21} color={nadaaBrand.colors.green} />
                      <Box>
                        <Typography variant="h6">Nearby shelters</Typography>
                        <Typography variant="caption" color="text.secondary">
                          Capacity and facilities
                        </Typography>
                      </Box>
                    </Stack>
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
                  </Stack>
                  {shelterState.status === "fallback" ||
                  shelterState.status === "error" ? (
                    <Alert
                      severity={
                        shelterState.status === "fallback" ? "warning" : "error"
                      }
                      className="warning-alert"
                    >
                      {shelterState.message}
                    </Alert>
                  ) : null}
                  <Box
                    className="shelter-map-preview"
                    aria-label="Nearby shelter map preview"
                  >
                    <Typography variant="caption" color="text.secondary">
                      {riskCoordinates.lat}, {riskCoordinates.lng}
                    </Typography>
                    {shelterSupport.shelters
                      .slice(0, 3)
                      .map((shelter, index) => (
                        <Box
                          className={`shelter-map-dot shelter-map-dot-${index}`}
                          key={shelter.id}
                          title={shelter.name}
                        >
                          {index + 1}
                        </Box>
                      ))}
                  </Box>
                  <Stack spacing={1.25}>
                    {shelterSupport.shelters.length > 0 ? (
                      shelterSupport.shelters.map((shelter) => (
                        <Paper
                          variant="outlined"
                          className="shelter-row"
                          key={shelter.id}
                        >
                          <Stack
                            direction="row"
                            justifyContent="space-between"
                            spacing={1}
                          >
                            <Box>
                              <Typography variant="subtitle2">
                                {shelter.name}
                              </Typography>
                              <Typography
                                variant="body2"
                                color="text.secondary"
                              >
                                {formatOccupancy(shelter)}
                                {shelter.distanceMeters
                                  ? ` · ${formatDistance(shelter.distanceMeters)}`
                                  : ""}
                              </Typography>
                              {shelter.facilities.length ? (
                                <Typography
                                  variant="caption"
                                  color="text.secondary"
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
                                  shelter.status === "full"
                                    ? "warning"
                                    : "success"
                                }
                              />
                            ) : null}
                          </Stack>
                        </Paper>
                      ))
                    ) : (
                      <Alert severity="info" className="warning-alert">
                        No nearby shelters were returned for this area.
                      </Alert>
                    )}
                  </Stack>
                </Paper>

                {roadClosures.length > 0 && (
                  <Paper className="surface">
                    <Stack
                      direction="row"
                      spacing={1}
                      alignItems="center"
                      className="section-heading"
                    >
                      <TriangleAlert size={21} color={nadaaBrand.colors.gold} />
                      <Box>
                        <Typography variant="h6">Road closures</Typography>
                        <Typography variant="caption" color="text.secondary">
                          Active closures near this area
                        </Typography>
                      </Box>
                    </Stack>
                    <Stack spacing={1.25}>
                      {roadClosures.map((closure) => (
                        <Paper
                          variant="outlined"
                          className="shelter-row"
                          key={closure.id}
                        >
                          <Stack
                            direction="row"
                            justifyContent="space-between"
                            spacing={1}
                          >
                            <Box>
                              <Typography variant="subtitle2">
                                {closure.roadName}
                              </Typography>
                              <Typography
                                variant="body2"
                                color="text.secondary"
                              >
                                {closure.reason ?? "Road closure"} ·{" "}
                                {closure.severity}
                              </Typography>
                              {closure.detourNote ? (
                                <Typography
                                  variant="caption"
                                  color="text.secondary"
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
                      ))}
                    </Stack>
                  </Paper>
                )}

                <Paper className="surface">
                  <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                    className="section-heading"
                  >
                    <LifeBuoy size={21} color={nadaaBrand.colors.gold} />
                    <Box>
                      <Typography variant="h6">Relief distribution</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Food, water, medical and supply points
                      </Typography>
                    </Box>
                  </Stack>
                  <Stack spacing={1.25}>
                    {reliefPoints.reliefPoints.length > 0 ? (
                      reliefPoints.reliefPoints.map((point) => (
                        <Paper
                          variant="outlined"
                          className="shelter-row"
                          key={point.id}
                        >
                          <Stack spacing={1}>
                            <Stack
                              direction="row"
                              justifyContent="space-between"
                              spacing={1}
                            >
                              <Box>
                                <Typography variant="subtitle2">
                                  {point.name}
                                </Typography>
                                <Typography
                                  variant="body2"
                                  color="text.secondary"
                                >
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
                            <Typography
                              variant="caption"
                              color="text.secondary"
                            >
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
                    ) : (
                      <Alert severity="info" className="warning-alert">
                        No relief distribution points were returned for this
                        area.
                      </Alert>
                    )}
                  </Stack>
                </Paper>

                <Paper className="surface">
                  <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                    className="section-heading"
                  >
                    <LifeBuoy size={21} color={nadaaBrand.colors.green} />
                    <Typography variant="h6">Recovery support</Typography>
                  </Stack>
                  <Stack spacing={1.25}>
                    {shelterSupport.recoverySupport.length > 0 ? (
                      shelterSupport.recoverySupport.map((support) => (
                        <Paper
                          variant="outlined"
                          className="shelter-row"
                          key={support.id}
                        >
                          <Stack
                            direction="row"
                            justifyContent="space-between"
                            spacing={1}
                          >
                            <Box>
                              <Typography variant="subtitle2">
                                {support.name}
                              </Typography>
                              <Typography
                                variant="body2"
                                color="text.secondary"
                              >
                                {formatSupportType(support.type)}
                                {support.distanceMeters
                                  ? ` · ${formatDistance(support.distanceMeters)}`
                                  : ""}
                              </Typography>
                              <Typography
                                variant="caption"
                                color="text.secondary"
                              >
                                {support.hours} ·{" "}
                                {formatListLabel(support.services)}
                              </Typography>
                            </Box>
                            <Chip
                              size="small"
                              label={support.status}
                              color={
                                support.status === "open"
                                  ? "success"
                                  : "warning"
                              }
                            />
                          </Stack>
                        </Paper>
                      ))
                    ) : (
                      <Alert severity="info" className="warning-alert">
                        No recovery support locations were returned for this
                        area.
                      </Alert>
                    )}
                  </Stack>
                </Paper>

                <Paper className="surface">
                  <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                    className="section-heading"
                  >
                    <LifeBuoy size={21} color={nadaaBrand.colors.red} />
                    <Typography variant="h6">Nearby responders</Typography>
                  </Stack>
                  <Stack spacing={1.25}>
                    {risk.nearbyFacilities.length > 0 ? (
                      risk.nearbyFacilities.map((facility) => (
                        <Paper
                          variant="outlined"
                          className="shelter-row"
                          key={facility.id}
                        >
                          <Stack
                            direction="row"
                            justifyContent="space-between"
                            spacing={1}
                          >
                            <Box>
                              <Typography variant="subtitle2">
                                {facility.name}
                              </Typography>
                              <Typography
                                variant="body2"
                                color="text.secondary"
                              >
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
                        No nearby response facilities were returned for this
                        area.
                      </Alert>
                    )}
                  </Stack>
                </Paper>

                <Paper className="surface">
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={1}
                    justifyContent="space-between"
                    alignItems={{ xs: "stretch", sm: "center" }}
                    className="section-heading"
                  >
                    <Stack direction="row" spacing={1} alignItems="center">
                      <BookOpen size={21} color={nadaaBrand.colors.green} />
                      <Box>
                        <Typography variant="h6">Emergency guides</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {offlineGuideCount} saved for offline use
                        </Typography>
                      </Box>
                    </Stack>
                    <Button
                      type="button"
                      variant="outlined"
                      size="small"
                      startIcon={
                        guideState.status === "loading" ? (
                          <Loader2 size={16} className="spin-icon" />
                        ) : (
                          <RefreshCw size={16} />
                        )
                      }
                      onClick={() => void fetchGuides()}
                      disabled={guideState.status === "loading"}
                    >
                      Refresh
                    </Button>
                  </Stack>

                  <Grid container spacing={1.25} className="guide-filter-grid">
                    <Grid size={{ xs: 12, sm: 4 }}>
                      <FormControl fullWidth size="small">
                        <InputLabel>Hazard</InputLabel>
                        <Select
                          value={guideFilters.hazard}
                          label="Hazard"
                          onChange={(event) =>
                            setGuideFilters((current) => ({
                              ...current,
                              hazard: event.target.value as GuideHazardFilter,
                            }))
                          }
                        >
                          {guideHazardOptions.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                              {option.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    </Grid>
                    <Grid size={{ xs: 12, sm: 4 }}>
                      <FormControl fullWidth size="small">
                        <InputLabel>Stage</InputLabel>
                        <Select
                          value={guideFilters.stage}
                          label="Stage"
                          onChange={(event) =>
                            setGuideFilters((current) => ({
                              ...current,
                              stage: event.target.value as GuideStageFilter,
                            }))
                          }
                        >
                          {guideStageOptions.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                              {option.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    </Grid>
                    <Grid size={{ xs: 12, sm: 4 }}>
                      <FormControl fullWidth size="small">
                        <InputLabel>Language</InputLabel>
                        <Select
                          value={guideFilters.language}
                          label="Language"
                          onChange={(event) =>
                            updateGuideLanguage(event.target.value)
                          }
                          startAdornment={
                            <Languages
                              size={16}
                              className="select-leading-icon"
                            />
                          }
                        >
                          {guideLanguageOptions.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                              {option.label}
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>
                    </Grid>
                  </Grid>

                  <Alert
                    severity={
                      guideState.status === "error"
                        ? "error"
                        : guideState.status === "offline"
                          ? "warning"
                          : "info"
                    }
                    icon={
                      guideState.status === "offline" ? <WifiOff /> : undefined
                    }
                    className="warning-alert guide-cache-alert"
                  >
                    {guideState.message ??
                      `Cached ${guideLanguageLabel(guideCacheInfo.language)} guides ${formatDateTime(guideCacheInfo.cachedAt)}.`}
                  </Alert>

                  {featuredGuide ? (
                    <Box className="guide-feature">
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Chip
                          size="small"
                          color="success"
                          label={guideStageLabel(featuredGuide.stage)}
                        />
                        <Chip
                          size="small"
                          variant="outlined"
                          label={hazardLabel(featuredGuide.hazardType)}
                        />
                      </Stack>
                      <Typography variant="subtitle1">
                        {featuredGuide.title}
                      </Typography>
                      <Typography variant="body2" className="guide-body">
                        {featuredGuide.body}
                      </Typography>
                      <Button
                        fullWidth
                        variant="contained"
                        color="error"
                        startIcon={<Phone size={18} />}
                      >
                        Call 112
                      </Button>
                    </Box>
                  ) : (
                    <Alert severity="info" className="warning-alert">
                      No guide matches this filter yet. Try all hazards or all
                      stages.
                    </Alert>
                  )}

                  <Divider className="guide-divider" />

                  <Stack spacing={1}>
                    {visibleGuides.slice(0, 5).map((guide) => (
                      <Paper
                        variant="outlined"
                        className="guide-list-row"
                        key={guide.id}
                      >
                        <Stack direction="row" spacing={1.25}>
                          <CheckCircle2
                            size={18}
                            color={nadaaBrand.colors.green}
                          />
                          <Box>
                            <Stack
                              direction="row"
                              spacing={0.75}
                              alignItems="center"
                              flexWrap="wrap"
                            >
                              <Typography variant="subtitle2">
                                {guide.title}
                              </Typography>
                              {guide.offlineAvailable ? (
                                <Chip size="small" label="Offline" />
                              ) : null}
                            </Stack>
                            <Typography variant="body2" color="text.secondary">
                              {guideStageLabel(guide.stage)} ·{" "}
                              {hazardLabel(guide.hazardType)}
                            </Typography>
                          </Box>
                        </Stack>
                      </Paper>
                    ))}
                  </Stack>
                </Paper>

                <PublicCampaignsPanel />
              </Stack>
            </Grid>
          </Grid>
        </Container>
      </Box>
    </ThemeProvider>
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

export default CitizenApp;
