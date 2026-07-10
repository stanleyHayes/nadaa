import { type ChangeEvent, useEffect, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  Grid,
  Paper,
  Slider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { Play, Waves } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CreateFloodSimulationRequest,
  FloodSimulationFrame,
  FloodSimulationRun,
} from "@nadaa/shared-types";
import { SIMULATION_API_BASE } from "@/app/config";
import { EmptyState } from "./shared";
import { severityColors } from "../data";

function isPositiveNumber(value: string) {
  const n = Number(value);
  return value !== "" && Number.isFinite(n) && n > 0;
}

function frameSeveritySummary(frame: FloodSimulationFrame) {
  const counts: Record<string, number> = {};
  for (const cell of frame.cells) {
    counts[cell.severity] = (counts[cell.severity] ?? 0) + 1;
  }
  return counts;
}

export function FloodSimulationPanel() {
  const [name, setName] = useState("Accra +50 mm rainfall scenario");
  const [rainfallOverride, setRainfallOverride] = useState("50");
  const [waterLevelOverride, setWaterLevelOverride] = useState("10");
  const [duration, setDuration] = useState("6");
  const [timeStep, setTimeStep] = useState("1");

  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState("");
  const [run, setRun] = useState<FloodSimulationRun | null>(null);
  const [selectedFrameIndex, setSelectedFrameIndex] = useState(0);

  const canRun =
    name.trim().length > 0 &&
    isPositiveNumber(duration) &&
    isPositiveNumber(timeStep);

  const runSimulation = async () => {
    if (!canRun) {
      setFeedback(
        "Enter a name and positive duration/time step before running.",
      );
      return;
    }

    setBusy(true);
    setFeedback("");
    setRun(null);
    setSelectedFrameIndex(0);

    const body: CreateFloodSimulationRequest = {
      name: name.trim(),
      rainfallMmOverride: Number(rainfallOverride) || undefined,
      waterLevelTrendCmOverride: Number(waterLevelOverride) || undefined,
      durationHours: Number(duration),
      timeStepHours: Number(timeStep),
    };

    try {
      const response = await fetch(
        `${SIMULATION_API_BASE}/ml/flood/simulations`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        },
      );
      if (!response.ok) {
        const payload = (await response.json()) as {
          error?: { message?: string };
        };
        throw new Error(
          payload.error?.message ?? `ml-service returned ${response.status}`,
        );
      }
      const payload = (await response.json()) as {
        simulation: FloodSimulationRun;
      };
      setRun(payload.simulation);
      setFeedback(
        `Simulation ${payload.simulation.reference} completed with ${payload.simulation.frames.length} frame(s).`,
      );
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Flood simulation needs ml-service running on the configured URL.",
      );
    } finally {
      setBusy(false);
    }
  };

  const selectedFrame = run?.frames[selectedFrameIndex];

  return (
    <Paper className="surface simulation-panel">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Waves size={21} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h6">Real-time flood simulation</Typography>
            <Typography variant="caption" color="text.secondary">
              Decision-support scenario runner; cannot publish alerts
            </Typography>
          </Box>
        </Stack>
      </Stack>

      {feedback ? (
        <Alert severity={run ? "success" : "warning"} className="feed-alert">
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <TextField
          label="Scenario name"
          size="small"
          fullWidth
          value={name}
          onChange={(event: ChangeEvent<HTMLInputElement>) =>
            setName(event.target.value)
          }
        />
        <Grid container spacing={1}>
          <Grid size={6}>
            <TextField
              label="Rainfall override (mm)"
              size="small"
              fullWidth
              value={rainfallOverride}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setRainfallOverride(event.target.value)
              }
              inputProps={{ inputMode: "decimal" }}
              helperText="Added to 24h forecast"
            />
          </Grid>
          <Grid size={6}>
            <TextField
              label="Water-level override (cm)"
              size="small"
              fullWidth
              value={waterLevelOverride}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setWaterLevelOverride(event.target.value)
              }
              inputProps={{ inputMode: "decimal" }}
              helperText="Added to level trend"
            />
          </Grid>
        </Grid>
        <Grid container spacing={1}>
          <Grid size={6}>
            <TextField
              label="Duration (hours)"
              size="small"
              fullWidth
              value={duration}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setDuration(event.target.value)
              }
              inputProps={{ inputMode: "numeric" }}
              error={Boolean(duration) && !isPositiveNumber(duration)}
            />
          </Grid>
          <Grid size={6}>
            <TextField
              label="Time step (hours)"
              size="small"
              fullWidth
              value={timeStep}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setTimeStep(event.target.value)
              }
              inputProps={{ inputMode: "numeric" }}
              error={Boolean(timeStep) && !isPositiveNumber(timeStep)}
            />
          </Grid>
        </Grid>
        <Button
          type="button"
          variant="contained"
          startIcon={<Play size={17} />}
          disabled={busy || !canRun}
          onClick={() => void runSimulation()}
        >
          {busy ? "Running" : "Run simulation"}
        </Button>
      </Stack>

      {run ? (
        <>
          <Divider className="detail-divider" />
          <Stack spacing={1.5}>
            <Stack
              direction="row"
              spacing={1}
              alignItems="center"
              justifyContent="space-between"
            >
              <Typography variant="subtitle2">
                Frames ({run.frames.length})
              </Typography>
              <Chip
                size="small"
                label={run.status}
                color={run.status === "completed" ? "success" : "default"}
              />
            </Stack>
            {run.frames.length > 1 ? (
              <Slider
                value={selectedFrameIndex}
                onChange={(_event, value) =>
                  setSelectedFrameIndex(value as number)
                }
                step={1}
                min={0}
                max={run.frames.length - 1}
                marks={run.frames.map((frame, index) => ({
                  value: index,
                  label: `+${(index + 1) * run.scenario.timeStepHours}h`,
                }))}
                valueLabelDisplay="auto"
              />
            ) : null}
            {selectedFrame ? (
              <SimulationFrameSummary frame={selectedFrame} />
            ) : null}
            <Box className="simulation-map-frame">
              <SimulationMap
                run={run}
                selectedFrameIndex={selectedFrameIndex}
              />
            </Box>
            <Alert severity="info">
              {run.safety.message} Model {run.modelVersion} ·{" "}
              {run.featureSetVersion}
            </Alert>
            <Stack spacing={0.5}>
              <Typography variant="caption" color="text.secondary">
                Limitations
              </Typography>
              <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
                {run.limitations.map((limitation) => (
                  <li key={limitation}>
                    <Typography variant="caption" color="text.secondary">
                      {limitation}
                    </Typography>
                  </li>
                ))}
              </ul>
            </Stack>
          </Stack>
        </>
      ) : (
        <EmptyState
          title="No simulation run"
          detail="Configure scenario inputs and run a flood simulation."
        />
      )}
    </Paper>
  );
}

function SimulationFrameSummary({ frame }: { frame: FloodSimulationFrame }) {
  const counts = frameSeveritySummary(frame);
  return (
    <Stack spacing={0.5}>
      <Typography variant="caption" color="text.secondary">
        Target time: {new Date(frame.targetTime).toLocaleString()}
      </Typography>
      <Stack direction="row" spacing={1} flexWrap="wrap">
        {Object.entries(counts).map(([severity, count]) => (
          <Chip
            key={severity}
            size="small"
            label={`${severity}: ${count}`}
            sx={{
              backgroundColor:
                severityColors[severity as keyof typeof severityColors] ??
                nadaaBrand.colors.slate,
              color: "#fff",
              fontWeight: 600,
            }}
          />
        ))}
      </Stack>
    </Stack>
  );
}

function SimulationMap({
  run,
  selectedFrameIndex,
}: {
  run: FloodSimulationRun;
  selectedFrameIndex: number;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [5.56, -0.2],
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

    const frame = run.frames[selectedFrameIndex];
    if (!frame) {
      return;
    }

    const group = L.featureGroup();
    for (const cell of frame.cells) {
      if (!cell.geometry || cell.geometry.type !== "Polygon") {
        continue;
      }
      const rings = cell.geometry.coordinates.map((ring) =>
        ring.map(([lng, lat]) => [lat, lng] as [number, number]),
      );
      const color =
        severityColors[cell.severity as keyof typeof severityColors] ??
        nadaaBrand.colors.slate;
      const polygon = L.polygon(rings, {
        color,
        weight: 1,
        fillColor: color,
        fillOpacity: 0.45,
      });
      polygon.bindPopup(
        `${cell.community}<br/>Severity: ${cell.severity}<br/>Probability: ${(cell.probability * 100).toFixed(1)}%<br/>Depth: ${cell.depthBand}`,
      );
      polygon.addTo(group);
    }

    group.addTo(layer);
    const bounds = group.getBounds();
    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.1), { animate: true });
    }
  }, [run, selectedFrameIndex]);

  return <Box ref={containerRef} className="simulation-map" />;
}
