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
  createTheme,
} from "@mui/material";
import {
  Bell,
  BookOpen,
  CheckCircle2,
  Clock3,
  Cross,
  ImagePlus,
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
  Waves,
  WifiOff,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  AlertSeverity,
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
  CreateIncidentRequest,
  CreateIncidentResponse,
  EmergencyGuideRecord,
  GuideListResponse,
  GuideStage,
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
  InitiateMediaUploadRequest,
  MediaUploadResponse,
  RiskLevel,
} from "@nadaa/shared-types";

const INCIDENT_API_BASE =
  import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";
const RISK_API_BASE =
  import.meta.env.VITE_RISK_API_URL ?? "http://localhost:8081/api/v1";
const NOTIFICATION_API_BASE =
  import.meta.env.VITE_NOTIFICATION_API_URL ?? "http://localhost:8090/api/v1";
const GUIDE_API_BASE =
  import.meta.env.VITE_GUIDE_API_URL ?? "http://localhost:8086/api/v1";
const GUIDE_CACHE_KEY = "nadaa.citizen.guides.v1";

const riskTone: Record<RiskLevel, "success" | "warning" | "error" | "info"> = {
  low: "success",
  moderate: "info",
  high: "warning",
  severe: "error",
  emergency: "error",
};

const sampleRisk: AreaRiskResponse = {
  location: "Accra Central",
  overallRisk: "high",
  risks: [
    {
      type: "flood",
      level: "severe",
      probability: 0.82,
      reason:
        "Heavy rainfall forecast, low elevation, and historical flood reports nearby.",
    },
    {
      type: "fire",
      level: "moderate",
      probability: 0.34,
      reason:
        "Dense market activity and recent dry periods increase localized risk.",
    },
  ],
  nearestShelters: [
    {
      id: "shelter-ama-001",
      name: "Accra Metro Assembly Shelter",
      location: { lat: 5.56, lng: -0.2 },
      capacity: 450,
      currentOccupancy: 116,
      contact: "112",
    },
    {
      id: "shelter-osu-002",
      name: "Osu Community Hall",
      location: { lat: 5.55, lng: -0.18 },
      capacity: 220,
      currentOccupancy: 34,
      contact: "112",
    },
  ],
  nearbyFacilities: [
    {
      id: "agency-nadmo-ama",
      name: "NADMO Accra Metro",
      type: "nadmo",
      location: { lat: 5.56, lng: -0.2 },
      contact: "112",
    },
  ],
  recommendedActions: [
    "Avoid low-lying roads and open drains.",
    "Move valuables above ground level.",
    "Prepare an evacuation route to the nearest safe shelter.",
  ],
};

const areaPresets = [
  { label: "Accra Metropolitan", lat: 5.6037, lng: -0.187 },
  { label: "Accra flood zone", lat: 5.56, lng: -0.2 },
  { label: "Kumasi area", lat: 6.6885, lng: -1.6244 },
];

function buildFallbackAlerts(): CitizenAlertFeedItem[] {
  const now = new Date();
  return [
    {
      id: "alert_feed_current_flood",
      title: "Severe flood warning",
      hazardType: "flood",
      severity: "severe_warning",
      message:
        "Heavy rainfall and rising drains may flood low-lying parts of Accra Metro and Tema.",
      target: {
        type: "district",
        ids: ["accra-metropolitan", "tema-metropolitan"],
        label: "Accra Metro and Tema",
      },
      targetLabel: "Accra Metro and Tema",
      startsAt: new Date(now.getTime() - 30 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() + 5 * 60 * 60 * 1000).toISOString(),
      status: "current",
      recommendedAction:
        "Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed.",
      evacuationRequired: true,
      shelterIds: ["shelter-ama-001", "shelter-osu-002"],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 20 * 60 * 1000).toISOString(),
    },
    {
      id: "alert_feed_current_fire",
      title: "Market fire watch",
      hazardType: "fire",
      severity: "watch",
      message:
        "Responders are monitoring dense market areas after smoke reports near electrical kiosks.",
      target: {
        type: "community",
        ids: ["accra-central"],
        label: "Accra Central",
      },
      targetLabel: "Accra Central",
      startsAt: new Date(now.getTime() - 20 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() + 3 * 60 * 60 * 1000).toISOString(),
      status: "current",
      recommendedAction:
        "Keep access lanes open, avoid overloaded sockets, and call 112 if you see flames or heavy smoke.",
      evacuationRequired: false,
      shelterIds: [],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 15 * 60 * 1000).toISOString(),
    },
    {
      id: "alert_feed_expired_road",
      title: "Road hazard resolved",
      hazardType: "road_crash",
      severity: "advisory",
      message:
        "Earlier congestion near Kaneshie Market Road has cleared after responders reopened the lane.",
      target: {
        type: "radius",
        ids: ["kaneshie-market-road"],
        label: "Kaneshie Market Road",
        center: { lat: 5.566, lng: -0.242 },
        radiusMeters: 1500,
      },
      targetLabel: "Kaneshie Market Road",
      startsAt: new Date(now.getTime() - 8 * 60 * 60 * 1000).toISOString(),
      expiresAt: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
      status: "expired",
      recommendedAction:
        "Continue to drive carefully and give way to emergency vehicles.",
      evacuationRequired: false,
      shelterIds: [],
      source: "fixture",
      updatedAt: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
    },
  ];
}

const guideHazardOptions: { label: string; value: GuideHazardFilter }[] = [
  { label: "All hazards", value: "all" },
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Electrical", value: "electrical_hazard" },
  { label: "Disease", value: "disease_outbreak" },
  { label: "General", value: "other" },
];

const guideStageOptions: { label: string; value: GuideStageFilter }[] = [
  { label: "All stages", value: "all" },
  { label: "Before", value: "before" },
  { label: "During", value: "during" },
  { label: "After", value: "after" },
  { label: "Recovery", value: "recovery" },
];

const guideLanguageOptions = [
  { label: "English", value: "en" },
  { label: "Twi", value: "tw" },
  { label: "Ga", value: "ga" },
];

function buildFallbackGuides(): EmergencyGuideRecord[] {
  const now = new Date().toISOString();
  return [
    {
      id: "guide_flood_before_en",
      hazardType: "flood",
      stage: "before",
      title: "Prepare before flooding",
      body: "Know your nearest shelter, keep documents dry, clear drains safely, prepare drinking water, and agree on a family meeting point.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 10,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_flood_during_en",
      hazardType: "flood",
      stage: "during",
      title: "Stay safe during flooding",
      body: "Move to higher ground, avoid walking or driving through floodwater, turn off electricity only if safe, and call 112 for life-threatening danger.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 20,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_fire_during_en",
      hazardType: "fire",
      stage: "during",
      title: "Fire safety response",
      body: "Leave immediately, warn people nearby, stay low under smoke, never use lifts, and call 112 for Ghana National Fire Service support.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 40,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_evacuation_during_en",
      hazardType: "other",
      stage: "during",
      title: "Safe evacuation",
      body: "Take only essentials, follow official routes, help children and elderly people first, avoid floodwater or smoke, and tell relatives where you are going.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 80,
      createdAt: now,
      updatedAt: now,
    },
    {
      id: "guide_112_during_en",
      hazardType: "other",
      stage: "during",
      title: "Calling 112",
      body: "Call 112 for life-threatening emergencies. Share the hazard, exact location, people affected, injuries, and a safe callback number if available.",
      language: "en",
      offlineAvailable: true,
      sortOrder: 110,
      createdAt: now,
      updatedAt: now,
    },
  ];
}

const hazardOptions: { label: string; value: HazardType }[] = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Blocked drain", value: "blocked_drain" },
  { label: "Other", value: "other" },
];

const urgencyOptions: { label: string; value: IncidentUrgency }[] = [
  { label: "Moderate", value: "moderate" },
  { label: "High", value: "high" },
  { label: "Life threatening", value: "life_threatening" },
  { label: "Low", value: "low" },
];

const supportedMediaTypes: IncidentMediaContentType[] = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "video/mp4",
  "video/quicktime",
  "audio/mpeg",
  "audio/mp4",
  "audio/wav",
];

const mediaSizeLimits: Record<IncidentMediaContentType, number> = {
  "image/jpeg": 10 * 1024 * 1024,
  "image/png": 10 * 1024 * 1024,
  "image/webp": 10 * 1024 * 1024,
  "video/mp4": 100 * 1024 * 1024,
  "video/quicktime": 100 * 1024 * 1024,
  "audio/mpeg": 25 * 1024 * 1024,
  "audio/mp4": 25 * 1024 * 1024,
  "audio/wav": 25 * 1024 * 1024,
};

type ReportForm = {
  hazard: HazardType;
  lat: string;
  lng: string;
  description: string;
  peopleAffected: string;
  injuriesReported: boolean;
  urgency: IncidentUrgency;
  anonymous: boolean;
  contactPermission: boolean;
  accessibilityNeeds: string;
  files: File[];
};

type ReportState =
  | { status: "idle" }
  | { status: "loading"; message: string }
  | { status: "success"; reference: string; priorityReview: boolean }
  | { status: "error"; message: string };

type RiskState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "error"; message: string }
  | { status: "permission-denied"; message: string };

type AlertFeedView = "current" | "expired" | "all";

type AlertFeedState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "error"; message: string };

type GuideHazardFilter = "all" | HazardType;
type GuideStageFilter = "all" | GuideStage;

type GuideFilters = {
  hazard: GuideHazardFilter;
  stage: GuideStageFilter;
  language: string;
};

type GuideState =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "offline"; message: string }
  | { status: "error"; message: string };

type GuideCachePayload = {
  guides: EmergencyGuideRecord[];
  cachedAt: string;
  language: string;
};

type GuideCacheInfo = {
  cachedAt: string;
  source: "cache" | "fixture" | "network";
  language: string;
};

const initialReportForm: ReportForm = {
  hazard: "flood",
  lat: "5.579",
  lng: "-0.212",
  description: "",
  peopleAffected: "0",
  injuriesReported: false,
  urgency: "moderate",
  anonymous: false,
  contactPermission: true,
  accessibilityNeeds: "",
  files: [],
};

const theme = createTheme({
  palette: {
    primary: { main: nadaaBrand.colors.navy },
    secondary: { main: nadaaBrand.colors.green },
    error: { main: nadaaBrand.colors.red },
    warning: { main: nadaaBrand.colors.gold },
    background: { default: "#F4F7FB", paper: nadaaBrand.colors.white },
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
    button: { fontWeight: 800, textTransform: "none" },
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none",
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          minHeight: 42,
        },
      },
    },
  },
});

function App() {
  const [area, setArea] = useState("Accra Central");
  const [risk, setRisk] = useState<AreaRiskResponse>(sampleRisk);
  const [riskCoordinates, setRiskCoordinates] = useState({
    lat: "5.603700",
    lng: "-0.187000",
  });
  const [riskState, setRiskState] = useState<RiskState>({ status: "idle" });
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
    } catch (error) {
      setRiskState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not load area risk.",
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

    if (
      !Number.isFinite(lat) ||
      lat < -90 ||
      lat > 90 ||
      !Number.isFinite(lng) ||
      lng < -180 ||
      lng > 180
    ) {
      setReportState({
        status: "error",
        message: "Enter valid coordinates before sending the report.",
      });
      return;
    }

    if (reportForm.description.trim().length < 5) {
      setReportState({
        status: "error",
        message: "Add a short description of what happened.",
      });
      return;
    }

    if (!Number.isInteger(peopleAffected) || peopleAffected < 0) {
      setReportState({
        status: "error",
        message: "People affected must be zero or more.",
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
    <ThemeProvider theme={theme}>
      <CssBaseline />
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
                  label="Area or coordinates"
                  value={area}
                  onChange={(event) => setArea(event.target.value)}
                  fullWidth
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
                                item.type === "flood"
                                  ? "#0B6FB8"
                                  : nadaaBrand.colors.red
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
                                <Chip
                                  size="small"
                                  label={item.level}
                                  color={riskTone[item.level]}
                                />
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
                              <Chip
                                size="small"
                                variant="outlined"
                                label={hazardLabel(alert.hazardType)}
                              />
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
                        No {alertFeedView === "all" ? "" : alertFeedView} alerts
                        are available.
                      </Alert>
                    )}
                  </Stack>
                </Paper>
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
                    <FormControl fullWidth>
                      <InputLabel>Hazard type</InputLabel>
                      <Select
                        value={reportForm.hazard}
                        label="Hazard type"
                        onChange={(event) =>
                          updateReportForm(
                            "hazard",
                            event.target.value as HazardType,
                          )
                        }
                      >
                        {hazardOptions.map((option) => (
                          <MenuItem key={option.value} value={option.value}>
                            {option.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <Grid container spacing={1.25}>
                      <Grid size={{ xs: 6 }}>
                        <TextField
                          label="Latitude"
                          value={reportForm.lat}
                          onChange={(event) =>
                            updateReportForm("lat", event.target.value)
                          }
                          fullWidth
                          inputMode="decimal"
                        />
                      </Grid>
                      <Grid size={{ xs: 6 }}>
                        <TextField
                          label="Longitude"
                          value={reportForm.lng}
                          onChange={(event) =>
                            updateReportForm("lng", event.target.value)
                          }
                          fullWidth
                          inputMode="decimal"
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
                      label="What happened?"
                      value={reportForm.description}
                      onChange={(event) =>
                        updateReportForm("description", event.target.value)
                      }
                      multiline
                      minRows={3}
                      inputProps={{ maxLength: 2000 }}
                    />
                    <Grid container spacing={1.25}>
                      <Grid size={{ xs: 6 }}>
                        <TextField
                          label="People affected"
                          value={reportForm.peopleAffected}
                          onChange={(event) =>
                            updateReportForm(
                              "peopleAffected",
                              event.target.value,
                            )
                          }
                          fullWidth
                          inputMode="numeric"
                        />
                      </Grid>
                      <Grid size={{ xs: 6 }}>
                        <FormControl fullWidth>
                          <InputLabel>Urgency</InputLabel>
                          <Select
                            value={reportForm.urgency}
                            label="Urgency"
                            onChange={(event) =>
                              updateReportForm(
                                "urgency",
                                event.target.value as IncidentUrgency,
                              )
                            }
                          >
                            {urgencyOptions.map((option) => (
                              <MenuItem key={option.value} value={option.value}>
                                {option.label}
                              </MenuItem>
                            ))}
                          </Select>
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
                          !reportForm.anonymous && reportForm.contactPermission
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
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  className="section-heading"
                >
                  <Cross size={21} color={nadaaBrand.colors.green} />
                  <Typography variant="h6">Nearby shelters</Typography>
                </Stack>
                <Stack spacing={1.25}>
                  {risk.nearestShelters.length > 0 ? (
                    risk.nearestShelters.map((shelter) => (
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
                            <Typography variant="body2" color="text.secondary">
                              {formatOccupancy(shelter)}
                              {shelter.distanceMeters
                                ? ` · ${formatDistance(shelter.distanceMeters)}`
                                : ""}
                            </Typography>
                          </Box>
                          {shelter.contact ? (
                            <Chip
                              size="small"
                              label={shelter.contact}
                              color="success"
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
                            <Typography variant="body2" color="text.secondary">
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
            </Stack>
          </Grid>
        </Grid>
      </Container>
    </ThemeProvider>
  );
}

function resolveAreaLookup(value: string) {
  const normalized = value.trim().toLowerCase();
  const preset = areaPresets.find(
    (item) => item.label.toLowerCase() === normalized,
  );
  if (preset) {
    return preset;
  }

  const partialPreset = areaPresets.find((item) =>
    item.label.toLowerCase().includes(normalized),
  );
  if (partialPreset && normalized.length >= 4) {
    return partialPreset;
  }

  const [latText, lngText] = value.split(",").map((part) => part.trim());
  const lat = Number(latText);
  const lng = Number(lngText);
  if (
    Number.isFinite(lat) &&
    Number.isFinite(lng) &&
    lat >= -90 &&
    lat <= 90 &&
    lng >= -180 &&
    lng <= 180
  ) {
    return { label: `${lat.toFixed(4)}, ${lng.toFixed(4)}`, lat, lng };
  }

  return null;
}

function alertSeverityTone(
  severity: AlertSeverity,
  status: CitizenAlertFeedItem["status"],
): "success" | "warning" | "error" | "info" {
  if (status === "expired") {
    return "info";
  }
  if (severity === "emergency" || severity === "severe_warning") {
    return "error";
  }
  if (severity === "warning" || severity === "watch") {
    return "warning";
  }
  return "info";
}

function alertSeverityLabel(severity: AlertSeverity): string {
  return severity
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function alertStatusLabel(status: CitizenAlertFeedItem["status"]): string {
  return status.charAt(0).toUpperCase() + status.slice(1);
}

function hazardLabel(hazard: HazardType): string {
  return hazard
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function formatDateTime(value: string): string {
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

function filterGuides(
  guides: EmergencyGuideRecord[],
  filters: GuideFilters,
): EmergencyGuideRecord[] {
  const stageAndHazardMatches = guides.filter((guide) => {
    if (filters.hazard !== "all" && guide.hazardType !== filters.hazard) {
      return false;
    }
    if (filters.stage !== "all" && guide.stage !== filters.stage) {
      return false;
    }
    return true;
  });

  const languageMatches = stageAndHazardMatches.filter(
    (guide) => guide.language === filters.language,
  );
  const fallbackMatches =
    languageMatches.length > 0 || filters.language === "en"
      ? languageMatches
      : stageAndHazardMatches.filter((guide) => guide.language === "en");

  return [...fallbackMatches].sort((a, b) => {
    if (a.sortOrder === b.sortOrder) {
      return a.title.localeCompare(b.title);
    }
    return a.sortOrder - b.sortOrder;
  });
}

function readGuideCache(): GuideCachePayload | null {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(GUIDE_CACHE_KEY);
    if (!raw) {
      return null;
    }
    const payload = JSON.parse(raw) as GuideCachePayload;
    if (!isGuideCachePayload(payload)) {
      return null;
    }
    return payload;
  } catch {
    return null;
  }
}

function writeGuideCache(
  guides: EmergencyGuideRecord[],
  language: string,
  cachedAt: string,
) {
  if (typeof window === "undefined") {
    return;
  }

  const offlineGuides = guides.filter((guide) => guide.offlineAvailable);
  if (!offlineGuides.length) {
    return;
  }

  try {
    window.localStorage.setItem(
      GUIDE_CACHE_KEY,
      JSON.stringify({ guides: offlineGuides, language, cachedAt }),
    );
  } catch {
    // Local storage can be unavailable in private or restricted browser modes.
  }
}

function registerCitizenServiceWorker() {
  if (
    typeof window === "undefined" ||
    !("serviceWorker" in navigator) ||
    !import.meta.env.PROD
  ) {
    return;
  }

  navigator.serviceWorker.register("/sw.js").catch(() => undefined);
}

function isGuideCachePayload(value: unknown): value is GuideCachePayload {
  return (
    typeof value === "object" &&
    value !== null &&
    Array.isArray((value as GuideCachePayload).guides) &&
    typeof (value as GuideCachePayload).cachedAt === "string" &&
    typeof (value as GuideCachePayload).language === "string" &&
    (value as GuideCachePayload).guides.every(
      (guide) =>
        typeof guide.id === "string" &&
        typeof guide.title === "string" &&
        typeof guide.body === "string" &&
        typeof guide.language === "string",
    )
  );
}

function guideStageLabel(stage: GuideStage): string {
  return stage.charAt(0).toUpperCase() + stage.slice(1);
}

function guideLanguageLabel(language: string): string {
  return (
    guideLanguageOptions.find((option) => option.value === language)?.label ??
    language.toUpperCase()
  );
}

function formatOccupancy(
  shelter: AreaRiskResponse["nearestShelters"][number],
): string {
  if (
    typeof shelter.currentOccupancy === "number" &&
    typeof shelter.capacity === "number"
  ) {
    return `${shelter.currentOccupancy}/${shelter.capacity} occupied`;
  }

  return shelter.status ? shelter.status : "Shelter status unavailable";
}

function formatDistance(meters: number): string {
  if (meters < 1000) {
    return `${Math.max(1, Math.round(meters))} m`;
  }

  return `${(meters / 1000).toFixed(1)} km`;
}

async function initiateMediaUploads(files: File[]): Promise<string[]> {
  const mediaIds: string[] = [];

  for (const file of files) {
    if (!supportedMediaTypes.includes(file.type as IncidentMediaContentType)) {
      throw new Error(`${file.name} is not a supported media type.`);
    }

    const payload: InitiateMediaUploadRequest = {
      purpose: "incident_media",
      fileName: file.name,
      contentType: file.type as IncidentMediaContentType,
      sizeBytes: file.size,
      uploadedBy: "usr_demo_citizen",
    };

    const response = await fetch(`${INCIDENT_API_BASE}/media/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(await extractAPIError(response));
    }

    const upload = (await response.json()) as MediaUploadResponse;
    mediaIds.push(upload.mediaId);
  }

  return mediaIds;
}

async function extractAPIError(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: { message?: string } };
    return (
      payload.error?.message ?? `Request failed with status ${response.status}`
    );
  } catch {
    return `Request failed with status ${response.status}`;
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024 * 1024) {
    return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default App;
