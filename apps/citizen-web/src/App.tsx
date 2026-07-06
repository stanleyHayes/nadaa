import { useMemo, useState } from "react";
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
  LifeBuoy,
  MapPin,
  Megaphone,
  Phone,
  ShieldCheck,
  Siren,
  Waves
} from "lucide-react";
import { featurePillars, nadaaBrand } from "@nadaa/brand";
import type { AreaRiskResponse, HazardType, RiskLevel } from "@nadaa/shared-types";

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
  const [hazard, setHazard] = useState<HazardType>("flood");
  const [anonymous, setAnonymous] = useState(false);
  const risk = useMemo(() => ({ ...sampleRisk, location: area || "Selected area" }), [area]);

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
                <Paper className="surface">
                  <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                    <Siren size={21} color={nadaaBrand.colors.gold} />
                    <Typography variant="h6">Report incident</Typography>
                  </Stack>
                  <Stack spacing={1.5}>
                    <FormControl fullWidth>
                      <InputLabel>Hazard type</InputLabel>
                      <Select value={hazard} label="Hazard type" onChange={(event) => setHazard(event.target.value as HazardType)}>
                        {hazardOptions.map((option) => (
                          <MenuItem key={option.value} value={option.value}>
                            {option.label}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <TextField label="Location" defaultValue="Use GPS or enter nearby landmark" />
                    <TextField label="What happened?" multiline minRows={3} />
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Report anonymously</Typography>
                      <Switch checked={anonymous} onChange={(event) => setAnonymous(event.target.checked)} />
                    </Stack>
                    <Button variant="contained" color="error" startIcon={<Siren size={18} />}>
                      Send report
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

export default App;

