import { ChangeEvent, FormEvent, useMemo, useState } from "react";
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
  createTheme
} from "@mui/material";
import {
  Bell,
  CheckCircle2,
  Cross,
  ImagePlus,
  LifeBuoy,
  Loader2,
  LocateFixed,
  MapPin,
  Megaphone,
  Phone,
  ShieldCheck,
  Siren,
  Waves
} from "lucide-react";
import { featurePillars, nadaaBrand } from "@nadaa/brand";
import type {
  AreaRiskResponse,
  CreateIncidentRequest,
  CreateIncidentResponse,
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
  InitiateMediaUploadRequest,
  MediaUploadResponse,
  RiskLevel
} from "@nadaa/shared-types";

const INCIDENT_API_BASE = import.meta.env.VITE_INCIDENT_API_URL ?? "http://localhost:8084/api/v1";

const riskTone: Record<RiskLevel, "success" | "warning" | "error" | "info"> = {
  low: "success",
  moderate: "info",
  high: "warning",
  severe: "error",
  emergency: "error"
};

const sampleRisk: AreaRiskResponse = {
  location: "Accra Central",
  overallRisk: "high",
  risks: [
    {
      type: "flood",
      level: "severe",
      probability: 0.82,
      reason: "Heavy rainfall forecast, low elevation, and historical flood reports nearby."
    },
    {
      type: "fire",
      level: "moderate",
      probability: 0.34,
      reason: "Dense market activity and recent dry periods increase localized risk."
    }
  ],
  nearestShelters: [
    {
      id: "shelter-ama-001",
      name: "Accra Metro Assembly Shelter",
      location: { lat: 5.56, lng: -0.2 },
      capacity: 450,
      currentOccupancy: 116,
      contact: "112"
    },
    {
      id: "shelter-osu-002",
      name: "Osu Community Hall",
      location: { lat: 5.55, lng: -0.18 },
      capacity: 220,
      currentOccupancy: 34,
      contact: "112"
    }
  ],
  recommendedActions: [
    "Avoid low-lying roads and open drains.",
    "Move valuables above ground level.",
    "Prepare an evacuation route to the nearest safe shelter."
  ]
};

const alerts = [
  {
    title: "Severe Flood Watch",
    area: "Accra Metro, Tema",
    severity: "Severe Warning",
    expires: "18:00",
    body: "Heavy rainfall may cause flooding in low-lying communities."
  },
  {
    title: "Road Hazard Notice",
    area: "Kaneshie Market Road",
    severity: "Watch",
    expires: "16:30",
    body: "Slow movement expected. Emergency vehicles have priority."
  }
];

const hazardOptions: { label: string; value: HazardType }[] = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Blocked drain", value: "blocked_drain" },
  { label: "Other", value: "other" }
];

const urgencyOptions: { label: string; value: IncidentUrgency }[] = [
  { label: "Moderate", value: "moderate" },
  { label: "High", value: "high" },
  { label: "Life threatening", value: "life_threatening" },
  { label: "Low", value: "low" }
];

const supportedMediaTypes: IncidentMediaContentType[] = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "video/mp4",
  "video/quicktime",
  "audio/mpeg",
  "audio/mp4",
  "audio/wav"
];

const mediaSizeLimits: Record<IncidentMediaContentType, number> = {
  "image/jpeg": 10 * 1024 * 1024,
  "image/png": 10 * 1024 * 1024,
  "image/webp": 10 * 1024 * 1024,
  "video/mp4": 100 * 1024 * 1024,
  "video/quicktime": 100 * 1024 * 1024,
  "audio/mpeg": 25 * 1024 * 1024,
  "audio/mp4": 25 * 1024 * 1024,
  "audio/wav": 25 * 1024 * 1024
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
  files: []
};

const theme = createTheme({
  palette: {
    primary: { main: nadaaBrand.colors.navy },
    secondary: { main: nadaaBrand.colors.green },
    error: { main: nadaaBrand.colors.red },
    warning: { main: nadaaBrand.colors.gold },
    background: { default: "#F4F7FB", paper: nadaaBrand.colors.white },
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
    button: { fontWeight: 800, textTransform: "none" }
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: "none"
        }
      }
    },
    MuiButton: {
      styleOverrides: {
        root: {
          minHeight: 42
        }
      }
    }
  }
});

function App() {
  const [area, setArea] = useState("Accra Central");
  const [reportForm, setReportForm] = useState<ReportForm>(initialReportForm);
  const [reportState, setReportState] = useState<ReportState>({ status: "idle" });
  const risk = useMemo(() => ({ ...sampleRisk, location: area || "Selected area" }), [area]);

  const updateReportForm = <Key extends keyof ReportForm>(key: Key, value: ReportForm[Key]) => {
    setReportForm((current) => ({ ...current, [key]: value }));
  };

  const useCurrentLocation = () => {
    if (!navigator.geolocation) {
      setReportState({ status: "error", message: "Location is not available on this device." });
      return;
    }

    setReportState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setReportForm((current) => ({
          ...current,
          lat: position.coords.latitude.toFixed(6),
          lng: position.coords.longitude.toFixed(6)
        }));
        setReportState({ status: "idle" });
      },
      () => {
        setReportState({ status: "error", message: "Location permission was not granted." });
      },
      { enableHighAccuracy: true, timeout: 10000 }
    );
  };

  const handleFileSelection = (event: ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(event.target.files ?? []);
    event.currentTarget.value = "";

    if (selectedFiles.length > 10) {
      setReportState({ status: "error", message: "Attach at most 10 media files to one report." });
      return;
    }

    const invalidFile = selectedFiles.find((file) => {
      if (!supportedMediaTypes.includes(file.type as IncidentMediaContentType)) {
        return true;
      }

      return file.size <= 0 || file.size > mediaSizeLimits[file.type as IncidentMediaContentType];
    });

    if (invalidFile) {
      setReportState({
        status: "error",
        message: `${invalidFile.name} is not supported or is too large for this report.`
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

    if (!Number.isFinite(lat) || lat < -90 || lat > 90 || !Number.isFinite(lng) || lng < -180 || lng > 180) {
      setReportState({ status: "error", message: "Enter valid coordinates before sending the report." });
      return;
    }

    if (reportForm.description.trim().length < 5) {
      setReportState({ status: "error", message: "Add a short description of what happened." });
      return;
    }

    if (!Number.isInteger(peopleAffected) || peopleAffected < 0) {
      setReportState({ status: "error", message: "People affected must be zero or more." });
      return;
    }

    if (!navigator.onLine) {
      setReportState({
        status: "error",
        message: "You appear to be offline. Keep this report open and try again when the connection returns."
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
        contactPermission: reportForm.anonymous ? false : reportForm.contactPermission,
        accessibilityNeeds: reportForm.accessibilityNeeds.trim() || undefined,
        media: mediaIds,
        reporter: reportForm.anonymous
          ? undefined
          : {
              userId: "usr_demo_citizen",
              phone: reportForm.contactPermission ? "+233200000000" : undefined
            }
      };

      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const incident = (await response.json()) as CreateIncidentResponse;
      setReportState({
        status: "success",
        reference: incident.reference,
        priorityReview: incident.priorityReview
      });
      setReportForm(initialReportForm);
    } catch (error) {
      setReportState({
        status: "error",
        message: error instanceof Error ? error.message : "Could not send report."
      });
    }
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="sticky" elevation={0} className="topbar">
        <Toolbar className="toolbar">
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Box component="img" src="/brand/nadaa-logo.png" alt="NADAA shield" className="brand-logo" />
            <Box>
              <Typography variant="h6" component="div" className="brand-wordmark">
                {nadaaBrand.name}
              </Typography>
              <Typography variant="caption" className="brand-subtitle">
                {nadaaBrand.slogan}
              </Typography>
            </Box>
          </Stack>
          <Button color="inherit" variant="outlined" startIcon={<Phone size={18} />} className="call-button">
            Call 112
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="xl" className="app-shell">
        <Grid container spacing={2.5}>
          <Grid size={{ xs: 12, lg: 8 }}>
            <Paper className="surface risk-surface">
              <Stack direction={{ xs: "column", md: "row" }} spacing={2} justifyContent="space-between">
                <Box>
                  <Typography variant="overline" color="secondary">
                    Citizen operations
                  </Typography>
                  <Typography variant="h4">Know your risk before conditions change</Typography>
                </Box>
                <ButtonGroup variant="contained" aria-label="risk view selector" className="mode-group">
                  <Button startIcon={<Waves size={17} />}>Risk</Button>
                  <Button startIcon={<Bell size={17} />}>Alerts</Button>
                  <Button startIcon={<Siren size={17} />}>Report</Button>
                </ButtonGroup>
              </Stack>

              <Box className="risk-lookup">
                <TextField
                  label="Area"
                  value={area}
                  onChange={(event) => setArea(event.target.value)}
                  fullWidth
                />
                <Button variant="contained" startIcon={<MapPin size={18} />}>
                  Use location
                </Button>
              </Box>

              <Grid container spacing={2}>
                <Grid size={{ xs: 12, md: 5 }}>
                  <Box className="risk-score">
                    <Typography variant="overline">Overall risk</Typography>
                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                      <Typography variant="h2">{risk.overallRisk}</Typography>
                      <Chip label="Rainfall rising" color="warning" />
                    </Stack>
                    <Typography color="text.secondary">
                      {risk.location} is currently being watched for flood conditions and blocked-drain reports.
                    </Typography>
                  </Box>
                </Grid>
                <Grid size={{ xs: 12, md: 7 }}>
                  <Stack spacing={1.5}>
                    {risk.risks.map((item) => (
                      <Paper variant="outlined" className="risk-row" key={item.type}>
                        <Stack direction="row" spacing={1.5} alignItems="flex-start">
                          <ShieldCheck size={22} color={item.type === "flood" ? "#0B6FB8" : nadaaBrand.colors.red} />
                          <Box>
                            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                              <Typography variant="subtitle1">{item.type.replace("_", " ")}</Typography>
                              <Chip size="small" label={item.level} color={riskTone[item.level]} />
                              {item.probability ? (
                                <Chip size="small" variant="outlined" label={`${Math.round(item.probability * 100)}%`} />
                              ) : null}
                            </Stack>
                            <Typography variant="body2" color="text.secondary">
                              {item.reason}
                            </Typography>
                          </Box>
                        </Stack>
                      </Paper>
                    ))}
                  </Stack>
                </Grid>
              </Grid>
            </Paper>

            <Grid container spacing={2.5} className="section-grid">
              <Grid size={{ xs: 12, md: 6 }}>
                <Paper className="surface">
                  <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                    <Megaphone size={21} color={nadaaBrand.colors.red} />
                    <Typography variant="h6">Live warnings</Typography>
                  </Stack>
                  <Stack spacing={1.5}>
                    {alerts.map((alert) => (
                      <Alert
                        key={alert.title}
                        severity={alert.severity.includes("Severe") ? "error" : "warning"}
                        className="warning-alert"
                      >
                        <Typography variant="subtitle2">{alert.title}</Typography>
                        <Typography variant="body2">{alert.area} · Expires {alert.expires}</Typography>
                        <Typography variant="body2">{alert.body}</Typography>
                      </Alert>
                    ))}
                  </Stack>
                </Paper>
              </Grid>

              <Grid size={{ xs: 12, md: 6 }}>
                <Paper className="surface report-surface">
                  <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                    <Siren size={21} color={nadaaBrand.colors.gold} />
                    <Typography variant="h6">Report incident</Typography>
                  </Stack>
                  <Stack component="form" spacing={1.5} onSubmit={submitReport} noValidate>
                    <FormControl fullWidth>
                      <InputLabel>Hazard type</InputLabel>
                      <Select
                        value={reportForm.hazard}
                        label="Hazard type"
                        onChange={(event) => updateReportForm("hazard", event.target.value as HazardType)}
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
                          onChange={(event) => updateReportForm("lat", event.target.value)}
                          fullWidth
                          inputMode="decimal"
                        />
                      </Grid>
                      <Grid size={{ xs: 6 }}>
                        <TextField
                          label="Longitude"
                          value={reportForm.lng}
                          onChange={(event) => updateReportForm("lng", event.target.value)}
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
                      onChange={(event) => updateReportForm("description", event.target.value)}
                      multiline
                      minRows={3}
                      inputProps={{ maxLength: 2000 }}
                    />
                    <Grid container spacing={1.25}>
                      <Grid size={{ xs: 6 }}>
                        <TextField
                          label="People affected"
                          value={reportForm.peopleAffected}
                          onChange={(event) => updateReportForm("peopleAffected", event.target.value)}
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
                            onChange={(event) => updateReportForm("urgency", event.target.value as IncidentUrgency)}
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
                        <Typography variant="body2">Call 112 immediately after sending this report.</Typography>
                      </Alert>
                    ) : null}
                    <TextField
                      label="Accessibility needs"
                      value={reportForm.accessibilityNeeds}
                      onChange={(event) => updateReportForm("accessibilityNeeds", event.target.value)}
                      inputProps={{ maxLength: 500 }}
                    />
                    <Button component="label" variant="outlined" startIcon={<ImagePlus size={18} />}>
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
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Injuries reported</Typography>
                      <Switch
                        checked={reportForm.injuriesReported}
                        onChange={(event) => updateReportForm("injuriesReported", event.target.checked)}
                      />
                    </Stack>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Report anonymously</Typography>
                      <Switch
                        checked={reportForm.anonymous}
                        onChange={(event) =>
                          setReportForm((current) => ({
                            ...current,
                            anonymous: event.target.checked,
                            contactPermission: event.target.checked ? false : current.contactPermission
                          }))
                        }
                      />
                    </Stack>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Allow contact</Typography>
                      <Switch
                        checked={!reportForm.anonymous && reportForm.contactPermission}
                        onChange={(event) => updateReportForm("contactPermission", event.target.checked)}
                        disabled={reportForm.anonymous}
                      />
                    </Stack>
                    {reportState.status === "error" ? (
                      <Alert severity="error" className="warning-alert">
                        {reportState.message}
                      </Alert>
                    ) : null}
                    {reportState.status === "success" ? (
                      <Alert severity={reportState.priorityReview ? "warning" : "success"} className="warning-alert">
                        <Typography variant="subtitle2">Report {reportState.reference} received</Typography>
                        <Typography variant="body2">Call 112 if anyone is in immediate danger.</Typography>
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
                      {reportState.status === "loading" ? reportState.message : "Send report"}
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
                    <Typography variant="body2">Police, fire, ambulance, NADMO and relief agencies.</Typography>
                  </Box>
                </Stack>
                <Button fullWidth variant="contained" color="error" startIcon={<Phone size={18} />}>
                  Call 112 now
                </Button>
              </Paper>

              <Paper className="surface">
                <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                  <Cross size={21} color={nadaaBrand.colors.green} />
                  <Typography variant="h6">Nearby shelters</Typography>
                </Stack>
                <Stack spacing={1.25}>
                  {risk.nearestShelters.map((shelter) => (
                    <Paper variant="outlined" className="shelter-row" key={shelter.id}>
                      <Stack direction="row" justifyContent="space-between" spacing={1}>
                        <Box>
                          <Typography variant="subtitle2">{shelter.name}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {shelter.currentOccupancy}/{shelter.capacity} occupied
                          </Typography>
                        </Box>
                        <Chip size="small" label={shelter.contact} color="success" />
                      </Stack>
                    </Paper>
                  ))}
                </Stack>
              </Paper>

              <Paper className="surface">
                <Typography variant="h6" className="section-heading">
                  Preparedness guides
                </Typography>
                <Stack spacing={1.25}>
                  {risk.recommendedActions.map((action) => (
                    <Stack direction="row" spacing={1.25} key={action}>
                      <CheckCircle2 size={19} color={nadaaBrand.colors.green} />
                      <Typography variant="body2">{action}</Typography>
                    </Stack>
                  ))}
                </Stack>
                <Divider className="guide-divider" />
                <Grid container spacing={1}>
                  {featurePillars.slice(0, 4).map((pillar) => (
                    <Grid size={{ xs: 6 }} key={pillar.title}>
                      <Box className="pillar-tile" style={{ borderColor: pillar.accent }}>
                        <Typography variant="subtitle2">{pillar.title}</Typography>
                        <Typography variant="caption">{pillar.description}</Typography>
                      </Box>
                    </Grid>
                  ))}
                </Grid>
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      </Container>
    </ThemeProvider>
  );
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
      uploadedBy: "usr_demo_citizen"
    };

    const response = await fetch(`${INCIDENT_API_BASE}/media/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
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
    return payload.error?.message ?? `Request failed with status ${response.status}`;
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
