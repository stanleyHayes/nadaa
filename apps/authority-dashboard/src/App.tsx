import {
  AppBar,
  Box,
  Button,
  Chip,
  Container,
  CssBaseline,
  Grid,
  LinearProgress,
  Paper,
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
import {
  AlertTriangle,
  BellRing,
  CheckCheck,
  Clock3,
  MapPinned,
  RadioTower,
  Route,
  ShieldAlert,
  Truck
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { IncidentStatus } from "@nadaa/shared-types";

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

const incidents: Array<{
  id: string;
  hazard: string;
  district: string;
  severity: string;
  status: IncidentStatus;
  assigned: string;
  age: string;
}> = [
  {
    id: "INC-0241",
    hazard: "Flood",
    district: "Accra Metro",
    severity: "Severe",
    status: "under_review",
    assigned: "NADMO AMA",
    age: "8 min"
  },
  {
    id: "INC-0239",
    hazard: "Road crash",
    district: "Tema",
    severity: "High",
    status: "assigned",
    assigned: "Ambulance + Police",
    age: "21 min"
  },
  {
    id: "INC-0236",
    hazard: "Blocked drain",
    district: "Ablekuma",
    severity: "Moderate",
    status: "verified",
    assigned: "District Assembly",
    age: "43 min"
  }
];

const queues = [
  { label: "New reports", value: 38, icon: ShieldAlert, color: nadaaBrand.colors.red },
  { label: "Verified", value: 17, icon: CheckCheck, color: nadaaBrand.colors.green },
  { label: "Teams en route", value: 9, icon: Truck, color: "#0B6FB8" },
  { label: "Alerts pending", value: 4, icon: BellRing, color: nadaaBrand.colors.gold }
];

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="sticky" elevation={0} className="topbar">
        <Toolbar className="toolbar">
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Box component="img" src="/brand/nadaa-logo.png" alt="NADAA shield" className="brand-logo" />
            <Box>
              <Typography variant="h6">NADAA Command</Typography>
              <Typography variant="caption">National Disaster Alert & Response Platform</Typography>
            </Box>
          </Stack>
          <Stack direction="row" spacing={1}>
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
        <Grid container spacing={2.5}>
          {queues.map((item) => {
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

        <Grid container spacing={2.5} className="main-grid">
          <Grid size={{ xs: 12, lg: 8 }}>
            <Paper className="surface map-surface">
              <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={2}>
                <Box>
                  <Typography variant="overline" color="secondary">
                    Incident command
                  </Typography>
                  <Typography variant="h5">Live Greater Accra operations map</Typography>
                </Box>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Chip label="Flood" color="info" />
                  <Chip label="Fire" color="error" />
                  <Chip label="Medical" color="success" />
                  <Chip label="Road" />
                </Stack>
              </Stack>

              <Box className="map-panel">
                <Box className="map-grid" />
                <Box className="map-marker marker-one">
                  <AlertTriangle size={18} />
                  Flood
                </Box>
                <Box className="map-marker marker-two">
                  <Truck size={18} />
                  Ambulance
                </Box>
                <Box className="map-marker marker-three">
                  <Route size={18} />
                  Closure
                </Box>
              </Box>
            </Paper>

            <Paper className="surface incident-table">
              <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                <MapPinned size={21} color={nadaaBrand.colors.navy} />
                <Typography variant="h6">Incident queue</Typography>
              </Stack>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>ID</TableCell>
                    <TableCell>Hazard</TableCell>
                    <TableCell>District</TableCell>
                    <TableCell>Severity</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Assigned</TableCell>
                    <TableCell>Age</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {incidents.map((incident) => (
                    <TableRow key={incident.id} hover>
                      <TableCell>{incident.id}</TableCell>
                      <TableCell>{incident.hazard}</TableCell>
                      <TableCell>{incident.district}</TableCell>
                      <TableCell>
                        <Chip
                          size="small"
                          label={incident.severity}
                          color={incident.severity === "Severe" ? "error" : "warning"}
                        />
                      </TableCell>
                      <TableCell>{incident.status.replaceAll("_", " ")}</TableCell>
                      <TableCell>{incident.assigned}</TableCell>
                      <TableCell>{incident.age}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, lg: 4 }}>
            <Stack spacing={2.5}>
              <Paper className="surface alert-panel">
                <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                  <BellRing size={21} color={nadaaBrand.colors.red} />
                  <Typography variant="h6">Alert approval</Typography>
                </Stack>
                <Stack spacing={1.5}>
                  <Box>
                    <Stack direction="row" justifyContent="space-between">
                      <Typography variant="subtitle2">Severe Flood Warning</Typography>
                      <Chip size="small" label="Draft" color="warning" />
                    </Stack>
                    <Typography variant="body2" color="text.secondary">
                      Accra Metro and Tema · expires 18:00
                    </Typography>
                  </Box>
                  <LinearProgress variant="determinate" value={68} />
                  <Button variant="contained" color="error" startIcon={<BellRing size={17} />}>
                    Review alert
                  </Button>
                </Stack>
              </Paper>

              <Paper className="surface">
                <Stack direction="row" spacing={1} alignItems="center" className="section-heading">
                  <Clock3 size={21} color={nadaaBrand.colors.gold} />
                  <Typography variant="h6">Response timeline</Typography>
                </Stack>
                <Stack spacing={1.5}>
                  {[
                    "INC-0241 reported by citizen with photo",
                    "Duplicate reports grouped for Accra Central",
                    "NADMO AMA dispatcher reviewing severity",
                    "Shelter capacity checked at Accra Metro Assembly"
                  ].map((event) => (
                    <Box className="timeline-row" key={event}>
                      <Typography variant="body2">{event}</Typography>
                    </Box>
                  ))}
                </Stack>
              </Paper>

              <Paper className="surface">
                <Typography variant="h6" className="section-heading">
                  Operating posture
                </Typography>
                <Stack spacing={1.25}>
                  <Stack direction="row" justifyContent="space-between">
                    <Typography variant="body2">Rainfall data</Typography>
                    <Chip size="small" label="Fresh" color="success" />
                  </Stack>
                  <Stack direction="row" justifyContent="space-between">
                    <Typography variant="body2">SMS provider</Typography>
                    <Chip size="small" label="Mock" />
                  </Stack>
                  <Stack direction="row" justifyContent="space-between">
                    <Typography variant="body2">ML alerts</Typography>
                    <Chip size="small" label="Human review" color="warning" />
                  </Stack>
                </Stack>
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      </Container>
    </ThemeProvider>
  );
}

export default App;

