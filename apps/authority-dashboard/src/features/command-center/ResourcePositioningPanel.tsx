import { type ChangeEvent, useEffect, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  Grid,
  LinearProgress,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { Ambulance, MapPin } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CompareScenarioRequest,
  CompareScenarioResponse,
  DemandForecast,
  ForecastListResponse,
  ScenarioResult,
  StagingSuggestion,
  StagingSuggestionListResponse,
} from "@nadaa/shared-types";
import { FORECAST_API_BASE } from "../../app/config";
import { EmptyState } from "./components";
import { severityColors } from "./data";

type ForecastLoadState = "loading" | "ready" | "fallback";

const agencyColors: Record<string, string> = {
  fire: nadaaBrand.colors.red,
  ambulance: nadaaBrand.colors.navy,
  nadmo: nadaaBrand.colors.gold,
};

const riskLevelOptions = [
  "any",
  "low",
  "moderate",
  "high",
  "severe",
  "emergency",
] as const;

// Fixtures mirror the ml-service resource-forecast output so the panel renders
// meaningful decision support even when the service is unavailable.
const fallbackForecasts: DemandForecast[] = [
  {
    id: "forecast_accra_metropolitan",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    timeWindowStart: "2026-07-10T00:00:00Z",
    timeWindowEnd: "2026-07-11T00:00:00Z",
    predictedIncidentCount: 15,
    hazardType: "flood",
    confidence: "high",
    confidenceScore: 0.95,
    factors: [
      {
        name: "historical_incidents",
        label: "Historical flood reports (30d)",
        value: 10,
        weight: 0.35,
        direction: "increases_demand",
      },
      {
        name: "rainfall_forecast",
        label: "Rainfall forecast 24h (mm)",
        value: 61.17,
        weight: 0.25,
        direction: "increases_demand",
      },
      {
        name: "risk_score",
        label: "Composite flood-risk score",
        value: 0.8026,
        weight: 0.25,
        direction: "increases_demand",
      },
      {
        name: "population_exposure",
        label: "Vulnerable population (%)",
        value: 19.33,
        weight: 0.15,
        direction: "increases_demand",
      },
    ],
    riskLevel: "severe",
    generatedAt: "2026-07-10T00:00:00Z",
  },
  {
    id: "forecast_tema_metropolitan",
    region: "Greater Accra",
    district: "Tema Metropolitan",
    timeWindowStart: "2026-07-10T00:00:00Z",
    timeWindowEnd: "2026-07-11T00:00:00Z",
    predictedIncidentCount: 3,
    hazardType: "flood",
    confidence: "medium",
    confidenceScore: 0.75,
    factors: [
      {
        name: "historical_incidents",
        label: "Historical flood reports (30d)",
        value: 1,
        weight: 0.35,
        direction: "increases_demand",
      },
      {
        name: "risk_score",
        label: "Composite flood-risk score",
        value: 0.4544,
        weight: 0.25,
        direction: "increases_demand",
      },
    ],
    riskLevel: "moderate",
    generatedAt: "2026-07-10T00:00:00Z",
  },
];

const fallbackStaging: StagingSuggestion[] = [
  {
    id: "staging_accra_central_fire",
    location: { lat: 5.545, lng: -0.205 },
    locationLabel: "Accra Central Fire Station",
    agencyType: "fire",
    reason:
      "Elevated predicted flood demand in Accra Metropolitan (~15 incidents in 24h, severe risk)",
    confidence: "high",
    confidenceScore: 0.95,
    operationalConstraints: [
      "Road congestion can extend response times during peak hours.",
      "Confirm water tanker and hydrant availability before repositioning.",
      "Flooded access routes may require alternate approaches; verify with route planning (NADAA-130).",
    ],
    recommendedUnits: 2,
    radiusMeters: 5000,
    generatedAt: "2026-07-10T00:00:00Z",
  },
  {
    id: "staging_ridge_ambulance",
    location: { lat: 5.563, lng: -0.19 },
    locationLabel: "Ridge Ambulance Base",
    agencyType: "ambulance",
    reason:
      "Elevated predicted flood demand in Accra Metropolitan (~15 incidents in 24h, severe risk)",
    confidence: "high",
    confidenceScore: 0.95,
    operationalConstraints: [
      "Road congestion can extend response times during peak hours.",
      "Coordinate with hospital emergency capacity before staging (NADAA-121).",
    ],
    recommendedUnits: 3,
    radiusMeters: 5000,
    generatedAt: "2026-07-10T00:00:00Z",
  },
];

function confidenceColor(score: number): string {
  if (score >= 0.8) {
    return nadaaBrand.colors.green;
  }
  if (score >= 0.6) {
    return nadaaBrand.colors.gold;
  }
  return nadaaBrand.colors.red;
}

function riskColor(riskLevel: string): string {
  return (
    severityColors[riskLevel as keyof typeof severityColors] ??
    nadaaBrand.colors.slate
  );
}

export function ResourcePositioningPanel() {
  const [forecasts, setForecasts] =
    useState<DemandForecast[]>(fallbackForecasts);
  const [staging, setStaging] = useState<StagingSuggestion[]>(fallbackStaging);
  const [loadState, setLoadState] = useState<ForecastLoadState>("loading");
  const [feedback, setFeedback] = useState("Loading resource forecasts");

  const [historicalWeight, setHistoricalWeight] = useState("1.5");
  const [timeWindowHours, setTimeWindowHours] = useState("24");
  const [riskLevel, setRiskLevel] =
    useState<(typeof riskLevelOptions)[number]>("any");
  const [scenarios, setScenarios] = useState<ScenarioResult[] | null>(null);
  const [compareBusy, setCompareBusy] = useState(false);
  const [compareFeedback, setCompareFeedback] = useState("");

  useEffect(() => {
    const controller = new AbortController();

    const refresh = async () => {
      setLoadState("loading");
      setFeedback("Loading resource forecasts");
      try {
        const [forecastResponse, stagingResponse] = await Promise.all([
          fetch(`${FORECAST_API_BASE}/forecasts`, {
            signal: controller.signal,
          }),
          fetch(`${FORECAST_API_BASE}/staging-suggestions`, {
            signal: controller.signal,
          }),
        ]);
        if (!forecastResponse.ok || !stagingResponse.ok) {
          throw new Error("forecast API unavailable");
        }
        const forecastPayload =
          (await forecastResponse.json()) as ForecastListResponse;
        const stagingPayload =
          (await stagingResponse.json()) as StagingSuggestionListResponse;
        if (controller.signal.aborted) {
          return;
        }
        if (forecastPayload.forecasts.length) {
          setForecasts(forecastPayload.forecasts);
        }
        if (stagingPayload.suggestions.length) {
          setStaging(stagingPayload.suggestions);
        }
        setLoadState("ready");
        setFeedback("Resource forecast API connected.");
      } catch (error) {
        if (controller.signal.aborted) {
          return;
        }
        setForecasts(fallbackForecasts);
        setStaging(fallbackStaging);
        setLoadState("fallback");
        setFeedback(
          "Resource forecast API unavailable. Showing rule-based fixture forecasts.",
        );
      }
    };

    void refresh();
    return () => controller.abort();
  }, []);

  const runComparison = async () => {
    setCompareBusy(true);
    setCompareFeedback("");
    setScenarios(null);

    const body: CompareScenarioRequest = {
      historicalWeight: Number(historicalWeight) || undefined,
      timeWindowHours: Number(timeWindowHours) || undefined,
      riskLevel: riskLevel === "any" ? undefined : riskLevel,
    };

    try {
      const response = await fetch(`${FORECAST_API_BASE}/forecasts/compare`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as {
          error?: { message?: string };
        } | null;
        throw new Error(
          payload?.error?.message ?? `forecast API returned ${response.status}`,
        );
      }
      const payload = (await response.json()) as CompareScenarioResponse;
      setScenarios(payload.scenarios);
      setCompareFeedback("Scenario comparison ready.");
    } catch (error) {
      setCompareFeedback(
        error instanceof Error
          ? error.message
          : "Scenario comparison needs ml-service running on the configured URL.",
      );
    } finally {
      setCompareBusy(false);
    }
  };

  const live = loadState === "ready";

  return (
    <Paper className="surface forecast-panel">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Ambulance size={21} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h6">
              Predictive resource positioning
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Decision-support staging suggestions; agency leadership retains
              deployment authority
            </Typography>
          </Box>
        </Stack>
        <Chip
          size="small"
          label={
            live ? "Live" : loadState === "loading" ? "Loading" : "Fixture"
          }
          color={live ? "success" : "warning"}
        />
      </Stack>

      {feedback ? (
        <Alert
          severity={
            live ? "success" : loadState === "loading" ? "info" : "warning"
          }
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Demand forecast by district</Typography>
        {forecasts.length ? (
          <Stack spacing={1}>
            {forecasts.map((forecast) => (
              <Box key={forecast.id} className="forecast-row">
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                  spacing={1}
                >
                  <Box>
                    <Typography variant="body2">{forecast.district}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {forecast.region} · {forecast.predictedIncidentCount}{" "}
                      predicted {forecast.hazardType} incident(s)
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    label={forecast.riskLevel}
                    sx={{
                      backgroundColor: riskColor(forecast.riskLevel),
                      color: "#fff",
                      fontWeight: 600,
                    }}
                  />
                </Stack>
                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  sx={{ mt: 0.5 }}
                >
                  <Typography variant="caption" color="text.secondary">
                    Confidence {forecast.confidence}
                  </Typography>
                  <Box sx={{ flexGrow: 1 }}>
                    <LinearProgress
                      variant="determinate"
                      value={forecast.confidenceScore * 100}
                      sx={{
                        height: 6,
                        borderRadius: 1,
                        backgroundColor: nadaaBrand.colors.slate + "20",
                        "& .MuiLinearProgress-bar": {
                          backgroundColor: confidenceColor(
                            forecast.confidenceScore,
                          ),
                        },
                      }}
                    />
                  </Box>
                </Stack>
              </Box>
            ))}
          </Stack>
        ) : (
          <EmptyState
            title="No forecasts"
            detail="No demand forecasts are available for the current inputs."
          />
        )}

        <Divider className="detail-divider" />

        <Typography variant="subtitle2">Suggested staging positions</Typography>
        <Box className="forecast-map-frame">
          <StagingMap staging={staging} />
        </Box>
        <Stack spacing={1}>
          {staging.map((suggestion) => (
            <Box key={suggestion.id} className="forecast-row">
              <Stack
                direction="row"
                justifyContent="space-between"
                alignItems="center"
                spacing={1}
              >
                <Stack direction="row" spacing={1} alignItems="center">
                  <MapPin
                    size={16}
                    color={
                      agencyColors[suggestion.agencyType] ??
                      nadaaBrand.colors.slate
                    }
                  />
                  <Typography variant="body2">
                    {suggestion.locationLabel}
                  </Typography>
                </Stack>
                <Chip
                  size="small"
                  label={`${suggestion.agencyType} · ${suggestion.recommendedUnits} unit(s)`}
                  variant="outlined"
                />
              </Stack>
              <Typography variant="caption" color="text.secondary">
                {suggestion.reason}
              </Typography>
              <ul style={{ margin: "0.25rem 0 0", paddingLeft: "1.25rem" }}>
                {suggestion.operationalConstraints.map((constraint) => (
                  <li key={constraint}>
                    <Typography variant="caption" color="text.secondary">
                      {constraint}
                    </Typography>
                  </li>
                ))}
              </ul>
            </Box>
          ))}
        </Stack>

        <Divider className="detail-divider" />

        <Typography variant="subtitle2">Scenario comparison</Typography>
        <Grid container spacing={1}>
          <Grid size={4}>
            <TextField
              label="History weight"
              size="small"
              fullWidth
              value={historicalWeight}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setHistoricalWeight(event.target.value)
              }
              inputProps={{ inputMode: "decimal" }}
            />
          </Grid>
          <Grid size={4}>
            <TextField
              label="Window (h)"
              size="small"
              fullWidth
              value={timeWindowHours}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setTimeWindowHours(event.target.value)
              }
              inputProps={{ inputMode: "numeric" }}
            />
          </Grid>
          <Grid size={4}>
            <TextField
              label="Risk"
              size="small"
              fullWidth
              select
              value={riskLevel}
              onChange={(event) =>
                setRiskLevel(
                  event.target.value as (typeof riskLevelOptions)[number],
                )
              }
            >
              {riskLevelOptions.map((option) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
        </Grid>
        <Button
          type="button"
          variant="outlined"
          disabled={compareBusy}
          onClick={() => void runComparison()}
        >
          {compareBusy ? "Comparing" : "Compare scenarios"}
        </Button>
        {compareFeedback ? (
          <Alert
            severity={scenarios ? "success" : "warning"}
            className="feed-alert"
          >
            {compareFeedback}
          </Alert>
        ) : null}
        {scenarios ? (
          <Grid container spacing={1}>
            {scenarios.map((scenario) => (
              <Grid size={6} key={scenario.name}>
                <Box className="forecast-row">
                  <Typography variant="caption" color="text.secondary">
                    {scenario.name}
                  </Typography>
                  <Typography variant="body2">
                    {scenario.summary.totalPredictedIncidents} predicted
                    incident(s)
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    Avg confidence{" "}
                    {(scenario.summary.averageConfidenceScore * 100).toFixed(0)}
                    % · {scenario.forecasts.length} district(s)
                  </Typography>
                </Box>
              </Grid>
            ))}
          </Grid>
        ) : null}

        <Alert severity="info">
          Predictions are decision-support only; no automatic deployment orders
          are generated and agency leadership retains final deployment
          authority. Model {"resource-forecast-rules-0.1.0"}.
        </Alert>
      </Stack>
    </Paper>
  );
}

function StagingMap({ staging }: { staging: StagingSuggestion[] }) {
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

    const group = L.featureGroup();
    for (const suggestion of staging) {
      const color =
        agencyColors[suggestion.agencyType] ?? nadaaBrand.colors.slate;
      L.circle([suggestion.location.lat, suggestion.location.lng], {
        radius: suggestion.radiusMeters,
        color,
        weight: 1,
        fillColor: color,
        fillOpacity: 0.1,
      }).addTo(group);
      const marker = L.circleMarker(
        [suggestion.location.lat, suggestion.location.lng],
        {
          radius: 8,
          color: "#FFFFFF",
          weight: 2,
          fillColor: color,
          fillOpacity: 0.9,
        },
      );
      marker.bindPopup(
        `<strong>${suggestion.locationLabel}</strong><br/>${suggestion.agencyType} · ${suggestion.recommendedUnits} unit(s)<br/>Confidence: ${suggestion.confidence}`,
      );
      marker.addTo(group);
    }

    if (staging.length === 0) {
      return;
    }
    group.addTo(layer);
    const bounds = group.getBounds();
    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.2), { animate: true, maxZoom: 12 });
    }
  }, [staging]);

  return <Box ref={containerRef} className="forecast-map" />;
}
