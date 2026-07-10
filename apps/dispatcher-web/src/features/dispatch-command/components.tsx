import { type ChangeEvent, type ReactNode, useEffect, useRef } from "react";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  Divider,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
  AlertCircle,
  AlertOctagon,
  AlertTriangle,
  BellRing,
  BrainCircuit,
  Bug,
  Car,
  CheckCheck,
  CheckCircle2,
  CloudLightning,
  Crosshair,
  FileText,
  Flame,
  GitMerge,
  HeartPulse,
  Hospital,
  Info,
  ListChecks,
  MapPin,
  Mountain,
  RefreshCw,
  ShieldAlert,
  Truck,
} from "lucide-react";
import {
  nadaaBrand,
  hazardRoles,
  severityRoles,
  type Hazard,
  type Severity,
} from "@nadaa/brand";
import type {
  AlertTarget,
  AuthorityAlertRecord,
  DuplicateReviewCandidate,
  HazardType,
  HospitalCapacityRecord,
  IncidentTriageSuggestion,
  ReliefPointRecord,
  RiskLevel,
  RoadClosureRecord,
} from "@nadaa/shared-types";
import type {
  TriageLoadState,
  TriageSuggestionFormState,
  TriageSuggestionReview,
} from "./types";
import {
  alertSeverityOptions,
  alertTargetTypeOptions,
  assignmentAgencyOptions,
  incidentTransitionOptions,
  triageSeverityOptions,
} from "./data";
import type {
  AbuseReviewFormState,
  AlertFormState,
  AlertLoadState,
  AssignmentFormState,
  CapacityLoadState,
  CommandIncident,
  HospitalCapacityFilterState,
  IncidentStatusFormState,
  MLPredictionReview,
  MLReviewLoadState,
} from "./types";
import {
  abuseDecisionLabel,
  abuseScoreLabel,
  agencyTypeLabel,
  alertSeverityLabel,
  alertStatusColor,
  alertStatusLabel,
  alertTargetSummary,
  alertTargetTypeLabel,
  alertTargetWarnings,
  buildAlertTarget,
  canAssignIncident,
  confidenceLabel,
  contributionLabel,
  contributionProgress,
  expectedOnsetLabel,
  formatShortTime,
  hazardLabel,
  hospitalBedPercent,
  hospitalCapacityColor,
  hospitalCapacityLabel,
  hospitalUnitStatusLabel,
  hospitalUpdatedLabel,
  metersLabel,
  parseTargetGeometry,
  probabilityLabel,
  requiresIncidentResolution,
  severityLabel,
  statusLabel,
  triageConfidenceLabel,
} from "./utils";

export function CommandSelect({
  children,
  label,
  onChange,
  value,
}: {
  children: ReactNode;
  label: string;
  onChange: (event: SelectChangeEvent) => void;
  value: string;
}) {
  return (
    <FormControl fullWidth size="small">
      <InputLabel>{label}</InputLabel>
      <Select label={label} value={value} onChange={onChange}>
        {children}
      </Select>
    </FormControl>
  );
}

const severityIconMap = {
  CheckCircle2,
  AlertTriangle,
  AlertOctagon,
  Info,
};

const hazardIconMap: Record<Hazard, React.ComponentType<{ size?: number }>> = {
  flood: MapPin,
  fire: Flame,
  medical: HeartPulse,
  geological: Mountain,
  road: Car,
  storm: CloudLightning,
  disease: Bug,
  default: AlertCircle,
};

function severityRoleKey(severity: string): Severity {
  if (severity === "emergency") return "severe";
  if (severity === "moderate") return "medium";
  if (severity in severityRoles) return severity as Severity;
  return "info";
}

function hazardRoleKey(type: string): Hazard {
  switch (type) {
    case "flood":
    case "flash_flood":
      return "flood";
    case "fire":
    case "electrical_hazard":
    case "chemical_hazard":
      return "fire";
    case "medical_emergency":
      return "medical";
    case "landslide":
    case "earthquake":
    case "sinkhole":
      return "geological";
    case "road_crash":
    case "traffic_incident":
      return "road";
    case "storm":
    case "wind_damage":
    case "tornado":
      return "storm";
    case "disease_outbreak":
      return "disease";
    default:
      return "default";
  }
}

export function SeverityChip({
  severity,
  size = "small",
}: {
  severity: string;
  size?: "small" | "medium";
}) {
  const role = severityRoles[severityRoleKey(severity)];
  const Icon = severityIconMap[role.icon];
  return (
    <Chip
      size={size}
      icon={<Icon size={14} aria-hidden="true" />}
      label={severityLabel(severity as RiskLevel)}
      className="severity-chip"
      sx={{
        backgroundColor: role.background,
        color: role.foreground,
        border: `1px solid ${role.border}`,
        fontWeight: 800,
        "& .MuiChip-icon": {
          color: "inherit",
        },
      }}
    />
  );
}

export function HazardChip({
  hazard,
  size = "small",
}: {
  hazard: string;
  size?: "small" | "medium";
}) {
  const roleKey = hazardRoleKey(hazard);
  const role = hazardRoles[roleKey];
  const Icon = hazardIconMap[roleKey];
  return (
    <Chip
      size={size}
      icon={<Icon size={14} aria-hidden="true" />}
      label={hazardLabel(hazard as HazardType)}
      sx={{
        backgroundColor: role.background,
        color: role.foreground,
        border: `1px solid ${role.border}`,
        fontWeight: 600,
        "& .MuiChip-icon": {
          color: "inherit",
        },
      }}
    />
  );
}

const closureSeverityColors: Record<string, string> = {
  emergency: "#7f1d1d",
  severe: "#dc2626",
  high: "#f97316",
  moderate: "#eab308",
  low: "#64748b",
};

export function IncidentMap({
  incidents,
  onSelect,
  selectedIncidentId,
  closures,
  reliefPoints,
}: {
  incidents: CommandIncident[];
  onSelect: (incidentId: string) => void;
  selectedIncidentId?: string;
  closures?: RoadClosureRecord[];
  reliefPoints?: ReliefPointRecord[];
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [5.586, -0.18],
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

    const bounds = L.latLngBounds([]);
    incidents.forEach((incident) => {
      const isSelected = incident.id === selectedIncidentId;
      const marker = L.circleMarker(
        [incident.location.lat, incident.location.lng],
        {
          radius: isSelected ? 13 : 9,
          color: "#FFFFFF",
          weight: isSelected ? 4 : 2,
          fillColor:
            severityRoles[severityRoleKey(incident.severity)].foreground,
          fillOpacity: isSelected ? 0.95 : 0.78,
        },
      );
      marker.bindPopup(
        `<strong>${incident.reference}</strong><br>${hazardLabel(incident.type)} · ${severityLabel(incident.severity)}<br>${incident.locality}`,
      );
      marker.on("click", () => onSelect(incident.id));
      marker.addTo(layer);
      bounds.extend([incident.location.lat, incident.location.lng]);
    });

    (closures ?? []).forEach((closure) => {
      if (closure.status === "lifted" || closure.status === "cancelled") {
        return;
      }
      const latlngs = closure.geometry.coordinates.map((point) => [
        point[1],
        point[0],
      ]) as L.LatLngExpression[];
      const polyline = L.polyline(latlngs, {
        color: closureSeverityColors[closure.severity] ?? "#64748b",
        weight: 5,
        opacity: 0.85,
        dashArray: closure.status === "scheduled" ? "8,8" : undefined,
      });
      polyline.bindPopup(
        `<strong>${closure.roadName}</strong><br>${closure.reason ?? "Road closure"} · ${closure.severity}<br>${closure.detourNote ?? "No detour noted"}`,
      );
      polyline.addTo(layer);
      latlngs.forEach((latlng) => bounds.extend(latlng));
    });

    (reliefPoints ?? []).forEach((point) => {
      const marker = L.circleMarker([point.location.lat, point.location.lng], {
        radius: 8,
        color: "#FFFFFF",
        weight: 2,
        fillColor: "#0B6FB8",
        fillOpacity: 0.85,
      });
      const stock = point.stockCategories
        .map((item) => `${item.category}: ${item.quantity} ${item.unit}`)
        .join("<br>");
      marker.bindPopup(
        `<strong>${point.name}</strong><br>${point.type} · ${point.status}<br>${point.address}<br>${stock || "No stock recorded"}`,
      );
      marker.addTo(layer);
      bounds.extend([point.location.lat, point.location.lng]);
    });

    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 13 });
    }
  }, [incidents, onSelect, selectedIncidentId, closures, reliefPoints]);

  useEffect(() => {
    const map = mapRef.current;
    const selected = incidents.find(
      (incident) => incident.id === selectedIncidentId,
    );
    if (!map || !selected) {
      return;
    }
    map.flyTo(
      [selected.location.lat, selected.location.lng],
      Math.max(map.getZoom(), 12),
      {
        animate: true,
        duration: 0.45,
      },
    );
  }, [incidents, selectedIncidentId]);

  return (
    <Box className="map-frame">
      <Box ref={containerRef} className="leaflet-command-map" />
      {!incidents.length ? (
        <Box className="map-empty">
          <EmptyState
            title="No map markers"
            detail="No incidents match the current command filters."
          />
        </Box>
      ) : null}
    </Box>
  );
}

export function AlertTargetPreview({ target }: { target: AlertTarget }) {
  const warnings = alertTargetWarnings(target);
  return (
    <Box className="target-preview">
      <Stack direction="row" justifyContent="space-between" gap={1}>
        <Box>
          <Typography variant="subtitle2">Affected area preview</Typography>
          <Typography variant="caption" color="text.secondary">
            {alertTargetSummary(target)}
          </Typography>
        </Box>
        <Chip
          size="small"
          label={alertTargetTypeLabel(target.type)}
          color={target.type === "national" ? "error" : "warning"}
        />
      </Stack>

      <TargetPreviewMap target={target} />

      <Grid container spacing={1}>
        <Grid size={4}>
          <Fact
            label="Area"
            value={`${Math.round((target.areaSqKm ?? 0) * 10) / 10} sq km`}
          />
        </Grid>
        <Grid size={4}>
          <Fact
            label="Population"
            value={`${target.estimatedPopulation ?? 0}`}
          />
        </Grid>
        <Grid size={4}>
          <Fact
            label="Radius"
            value={
              target.radiusMeters
                ? `${Math.round(target.radiusMeters / 100) / 10} km`
                : "Geometry"
            }
          />
        </Grid>
      </Grid>

      {warnings.map((warning) => (
        <Alert severity="warning" key={warning}>
          {warning}
        </Alert>
      ))}
    </Box>
  );
}

export function TargetPreviewMap({ target }: { target: AlertTarget }) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [target.center?.lat ?? 5.586, target.center?.lng ?? -0.18],
      zoom: 11,
      zoomControl: false,
      scrollWheelZoom: false,
      dragging: false,
      doubleClickZoom: false,
      boxZoom: false,
      keyboard: false,
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
    const color =
      target.type === "national" ? nadaaBrand.colors.red : "#0B6FB8";
    if (target.geometry?.coordinates?.[0]?.length) {
      const polygonPoints = target.geometry.coordinates[0].map(
        ([lng, lat]) => [lat, lng] as [number, number],
      );
      const polygon = L.polygon(polygonPoints, {
        color,
        fillColor: color,
        fillOpacity: 0.18,
        weight: 2,
      }).addTo(layer);
      map.fitBounds(polygon.getBounds().pad(0.12), { animate: false });
      return;
    }

    if (target.center) {
      const radius = target.radiusMeters || 2000;
      const circle = L.circle([target.center.lat, target.center.lng], {
        radius,
        color,
        fillColor: color,
        fillOpacity: 0.18,
        weight: 2,
      }).addTo(layer);
      map.fitBounds(circle.getBounds().pad(0.12), { animate: false });
    }
  }, [target]);

  return <Box ref={containerRef} className="target-preview-map" />;
}

export function MLPredictionReviewPanel({
  busy,
  feedback,
  loadMessage,
  loadState,
  onCreateDraft,
  onRefresh,
  onSelectPrediction,
  onUpdateReviewNote,
  predictions,
  reviewNote,
  selectedPrediction,
  selectedPredictionId,
}: {
  busy: boolean;
  feedback: string;
  loadMessage: string;
  loadState: MLReviewLoadState;
  onCreateDraft: () => void;
  onRefresh: () => void;
  onSelectPrediction: (predictionId: string) => void;
  onUpdateReviewNote: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  predictions: MLPredictionReview[];
  reviewNote: string;
  selectedPrediction?: MLPredictionReview;
  selectedPredictionId: string;
}) {
  const live = loadState === "ready";

  return (
    <Paper className="surface ml-review-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        gap={1.5}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <BrainCircuit size={22} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h5">ML flood review</Typography>
            <Typography variant="caption" color="text.secondary">
              Review probability, severity, confidence, and explanation before
              drafting an alert.
            </Typography>
          </Box>
        </Stack>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Chip
            size="small"
            label={
              live ? "Live ML" : loadState === "loading" ? "Loading" : "Fixture"
            }
            color={live ? "success" : "warning"}
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<BrainCircuit size={16} />}
            disabled={loadState === "loading"}
            onClick={onRefresh}
          >
            Refresh ML
          </Button>
        </Stack>
      </Stack>

      {loadState === "fallback" || loadState === "error" ? (
        <Alert severity="warning" className="ml-review-alert">
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <PredictionReviewMap
            predictions={predictions}
            selectedPredictionId={selectedPredictionId}
            onSelect={onSelectPrediction}
          />

          <Stack className="prediction-list" spacing={1}>
            {predictions.map((prediction) => (
              <Box
                key={prediction.id}
                className={`prediction-row${
                  prediction.id === selectedPredictionId ? " selected" : ""
                }`}
                onClick={() => onSelectPrediction(prediction.id)}
              >
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  gap={1}
                  alignItems="flex-start"
                >
                  <Box>
                    <Typography variant="subtitle2">
                      {prediction.community}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {prediction.district} · {prediction.cellId}
                    </Typography>
                  </Box>
                  <SeverityChip severity={prediction.severity} />
                </Stack>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Chip
                    size="small"
                    label={probabilityLabel(prediction.probability)}
                  />
                  <Chip
                    size="small"
                    label={confidenceLabel(prediction.confidence)}
                  />
                  <Chip
                    size="small"
                    label={expectedOnsetLabel(prediction.expectedOnset)}
                  />
                  {prediction.reviewStatus === "draft_created" ? (
                    <Chip size="small" color="success" label="Draft created" />
                  ) : null}
                </Stack>
              </Box>
            ))}
          </Stack>
        </Grid>

        <Grid size={{ xs: 12, lg: 5 }}>
          {selectedPrediction ? (
            <Stack spacing={1.25} className="prediction-detail">
              <Stack direction="row" justifyContent="space-between" gap={1}>
                <Box>
                  <Typography variant="overline" color="secondary">
                    Selected prediction
                  </Typography>
                  <Typography variant="h6">
                    {selectedPrediction.community}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={probabilityLabel(selectedPrediction.probability)}
                  color={
                    selectedPrediction.severity === "severe" ||
                    selectedPrediction.severity === "emergency"
                      ? "error"
                      : selectedPrediction.severity === "low"
                        ? "success"
                        : "warning"
                  }
                />
              </Stack>

              <Grid container spacing={1}>
                <Grid size={6}>
                  <Fact
                    label="Severity"
                    value={severityLabel(selectedPrediction.severity)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact
                    label="Confidence"
                    value={confidenceLabel(selectedPrediction.confidence)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact
                    label="Expected onset"
                    value={expectedOnsetLabel(selectedPrediction.expectedOnset)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact label="Model" value={selectedPrediction.modelVersion} />
                </Grid>
              </Grid>

              <Alert severity="info" icon={<ShieldAlert size={18} />}>
                Human review is required and this prediction cannot auto-publish
                a public alert.
              </Alert>

              <Stack spacing={1}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <ListChecks size={18} color={nadaaBrand.colors.green} />
                  <Typography variant="subtitle2">
                    Explanation factors
                  </Typography>
                </Stack>
                {selectedPrediction.explanationFactors.map((factor) => (
                  <Box className="factor-row" key={factor.feature}>
                    <Stack
                      direction="row"
                      justifyContent="space-between"
                      gap={1}
                    >
                      <Box>
                        <Typography variant="body2">{factor.label}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {String(factor.value)} ·{" "}
                          {factor.direction === "increases_risk"
                            ? "Increases risk"
                            : "Reduces risk"}
                        </Typography>
                      </Box>
                      <Chip
                        size="small"
                        color={
                          factor.direction === "increases_risk"
                            ? "warning"
                            : "success"
                        }
                        label={contributionLabel(factor.contribution)}
                      />
                    </Stack>
                    <LinearProgress
                      variant="determinate"
                      value={contributionProgress(factor.contribution)}
                      color={
                        factor.direction === "increases_risk"
                          ? "warning"
                          : "success"
                      }
                    />
                  </Box>
                ))}
              </Stack>

              <TextField
                size="small"
                label="Review note"
                value={reviewNote}
                onChange={onUpdateReviewNote}
                multiline
                minRows={2}
              />

              {feedback ? (
                <Alert
                  severity={
                    feedback.includes("unavailable") ? "warning" : "success"
                  }
                >
                  {feedback}
                </Alert>
              ) : null}

              <Button
                variant="contained"
                color="error"
                startIcon={<FileText size={17} />}
                disabled={busy}
                onClick={onCreateDraft}
              >
                Create reviewed draft
              </Button>
            </Stack>
          ) : (
            <EmptyState
              title="No prediction selected"
              detail="Choose a prediction cell from the map or list."
            />
          )}
        </Grid>
      </Grid>
    </Paper>
  );
}

export function AITriageSuggestionPanel({
  busy,
  canAccept,
  canOverride,
  feedback,
  form,
  incident,
  loadMessage,
  loadState,
  onAccept,
  onOverride,
  onRefresh,
  onUpdateForm,
  populationError,
  reasonError,
  suggestion,
}: {
  busy: boolean;
  canAccept: boolean;
  canOverride: boolean;
  feedback: string;
  form: TriageSuggestionFormState;
  incident?: TriageSuggestionReview;
  loadMessage: string;
  loadState: TriageLoadState;
  onAccept: () => void;
  onOverride: () => void;
  onRefresh: () => void;
  onUpdateForm: (
    key: keyof TriageSuggestionFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  populationError: string;
  reasonError: string;
  suggestion?: IncidentTriageSuggestion;
}) {
  const live = loadState === "ready";

  return (
    <Paper className="surface triage-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        gap={1.5}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <BrainCircuit size={22} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h5">AI incident triage</Typography>
            <Typography variant="caption" color="text.secondary">
              Severity, duplicate likelihood, affected population, and agency
              routing suggestion.
            </Typography>
          </Box>
        </Stack>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Chip
            size="small"
            label={
              live ? "Live" : loadState === "loading" ? "Loading" : "Fixture"
            }
            color={live ? "success" : "warning"}
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<BrainCircuit size={16} />}
            disabled={loadState === "loading" || !incident}
            onClick={onRefresh}
          >
            Refresh triage
          </Button>
        </Stack>
      </Stack>

      {loadState === "fallback" ||
      loadState === "error" ||
      loadState === "empty" ? (
        <Alert
          severity={loadState === "empty" ? "info" : "warning"}
          className="ml-review-alert"
        >
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      {suggestion ? (
        <Stack spacing={1.5}>
          <Alert severity="info" icon={<ShieldAlert size={18} />}>
            Human review is required. This suggestion cannot auto-verify,
            auto-assign, or auto-publish an alert.
          </Alert>

          <Grid container spacing={1}>
            <Grid size={6}>
              <Fact
                label="Severity"
                value={severityLabel(suggestion.severity)}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Confidence"
                value={triageConfidenceLabel(suggestion.confidence)}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Duplicate likelihood"
                value={`${Math.round(suggestion.duplicateLikelihood * 100)}%`}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Affected population"
                value={`${suggestion.affectedPopulation.toLocaleString()} (estimate)`}
              />
            </Grid>
            <Grid size={12}>
              <Fact
                label="Suggested agency"
                value={`${suggestion.suggestedAgency.name} (${agencyTypeLabel(suggestion.suggestedAgency.agencyType)})`}
              />
            </Grid>
          </Grid>

          <Stack spacing={1}>
            <Stack direction="row" spacing={1} alignItems="center">
              <ListChecks size={18} color={nadaaBrand.colors.green} />
              <Typography variant="subtitle2">Explanation factors</Typography>
            </Stack>
            {suggestion.explanationFactors.map((factor) => (
              <Box className="factor-row" key={factor.feature}>
                <Stack direction="row" justifyContent="space-between" gap={1}>
                  <Box>
                    <Typography variant="body2">{factor.label}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {String(factor.value)} ·{" "}
                      {factor.direction === "increases_risk"
                        ? "Increases risk"
                        : "Reduces risk"}
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    color={
                      factor.direction === "increases_risk"
                        ? "warning"
                        : "success"
                    }
                    label={contributionLabel(factor.contribution)}
                  />
                </Stack>
                <LinearProgress
                  variant="determinate"
                  value={contributionProgress(factor.contribution)}
                  color={
                    factor.direction === "increases_risk"
                      ? "warning"
                      : "success"
                  }
                />
              </Box>
            ))}
          </Stack>

          <Divider className="detail-divider" />

          <Typography variant="subtitle2">Dispatcher override</Typography>

          <Grid container spacing={1.25}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <FormControl fullWidth size="small">
                <InputLabel>Severity</InputLabel>
                <Select
                  label="Severity"
                  value={form.severity}
                  onChange={onUpdateForm("severity")}
                >
                  {triageSeverityOptions.map((severity) => (
                    <MenuItem key={severity} value={severity}>
                      {severityLabel(severity)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                size="small"
                label="Affected population"
                value={form.affectedPopulation}
                onChange={onUpdateForm("affectedPopulation")}
                error={!!populationError}
                helperText={populationError || undefined}
                fullWidth
              />
            </Grid>
            <Grid size={12}>
              <FormControl fullWidth size="small">
                <InputLabel>Agency</InputLabel>
                <Select
                  label="Agency"
                  value={form.agencyId}
                  onChange={onUpdateForm("agencyId")}
                >
                  {assignmentAgencyOptions.map((agency) => (
                    <MenuItem key={agency.id} value={agency.id}>
                      {agency.name} ({agencyTypeLabel(agency.type)})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
          </Grid>

          <TextField
            size="small"
            label="Override reason"
            value={form.reason}
            onChange={onUpdateForm("reason")}
            helperText={
              reasonError || "Required when overriding. Logged for review."
            }
            multiline
            minRows={2}
          />

          {feedback ? (
            <Alert
              severity={
                feedback.includes("not logged") ||
                feedback.includes("not recorded") ||
                feedback.includes("unavailable") ||
                feedback.includes("needs") ||
                feedback.includes("only be logged")
                  ? "warning"
                  : "success"
              }
            >
              {feedback}
            </Alert>
          ) : null}

          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              startIcon={<CheckCircle2 size={17} />}
              disabled={busy || !canAccept}
              onClick={onAccept}
            >
              Accept suggestion
            </Button>
            <Button
              variant="contained"
              color="error"
              startIcon={<ShieldAlert size={17} />}
              disabled={
                busy || !canOverride || !!reasonError || !!populationError
              }
              onClick={onOverride}
            >
              Override
            </Button>
          </Stack>

          <Typography variant="caption" color="text.secondary">
            Model {suggestion.modelVersion}
            {suggestion.suggestionId
              ? ` · Suggestion ${suggestion.suggestionId}`
              : " · Fixture suggestion (not logged)"}
          </Typography>
        </Stack>
      ) : (
        <EmptyState
          title="No triage suggestion"
          detail={
            incident
              ? "Refresh to generate a suggestion."
              : "Select an incident first."
          }
        />
      )}
    </Paper>
  );
}

export function HospitalCapacityPanel({
  facilities,
  filters,
  loadMessage,
  loadState,
  onRefresh,
  onUpdateCapacity,
  onUpdateIncludeStale,
  onUpdateMinBeds,
  onUpdateService,
}: {
  facilities: HospitalCapacityRecord[];
  filters: HospitalCapacityFilterState;
  loadMessage: string;
  loadState: CapacityLoadState;
  onRefresh: () => void;
  onUpdateCapacity: (event: SelectChangeEvent) => void;
  onUpdateIncludeStale: (checked: boolean) => void;
  onUpdateMinBeds: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onUpdateService: (event: SelectChangeEvent) => void;
}) {
  const staleCount = facilities.filter((facility) => facility.stale).length;
  const minBedsInvalid =
    filters.minAvailableBeds.trim() !== "" &&
    (!Number.isFinite(Number(filters.minAvailableBeds)) ||
      Number(filters.minAvailableBeds) < 0);

  return (
    <Paper className="surface capacity-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        gap={1.5}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Hospital size={22} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h5">Hospital capacity</Typography>
            <Typography variant="caption" color="text.secondary">
              Beds, emergency unit status, ambulances, oxygen, and stale update
              warnings for dispatcher routing.
            </Typography>
          </Box>
        </Stack>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live capacity"
                : loadState === "loading"
                  ? "Loading"
                  : loadState === "empty"
                    ? "No matches"
                    : "Fixture capacity"
            }
            color={
              loadState === "ready"
                ? "success"
                : loadState === "empty"
                  ? "default"
                  : "warning"
            }
          />
          {staleCount ? (
            <Chip size="small" color="warning" label={`${staleCount} stale`} />
          ) : null}
          <Button
            variant="outlined"
            size="small"
            startIcon={<Hospital size={16} />}
            disabled={loadState === "loading"}
            onClick={onRefresh}
          >
            Refresh capacity
          </Button>
        </Stack>
      </Stack>

      {loadState === "fallback" ? (
        <Alert severity="warning" className="feed-alert">
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      <Grid container spacing={1.5} className="capacity-filters">
        <Grid size={{ xs: 12, md: 3 }}>
          <CommandSelect
            label="Service"
            value={filters.service}
            onChange={onUpdateService}
          >
            <MenuItem value="all">All services</MenuItem>
            <MenuItem value="emergency">Emergency</MenuItem>
            <MenuItem value="trauma">Trauma</MenuItem>
            <MenuItem value="icu">ICU</MenuItem>
            <MenuItem value="maternity">Maternity</MenuItem>
            <MenuItem value="pediatric">Pediatric</MenuItem>
            <MenuItem value="ambulance">Ambulance</MenuItem>
            <MenuItem value="oxygen">Oxygen</MenuItem>
          </CommandSelect>
        </Grid>
        <Grid size={{ xs: 12, md: 3 }}>
          <CommandSelect
            label="Capacity"
            value={filters.emergencyCapacity}
            onChange={onUpdateCapacity}
          >
            <MenuItem value="all">All capacity</MenuItem>
            <MenuItem value="available">Available</MenuItem>
            <MenuItem value="limited">Limited</MenuItem>
            <MenuItem value="full">Full</MenuItem>
            <MenuItem value="offline">Offline</MenuItem>
            <MenuItem value="unknown">Unknown</MenuItem>
          </CommandSelect>
        </Grid>
        <Grid size={{ xs: 12, md: 3 }}>
          <TextField
            fullWidth
            label="Min beds"
            size="small"
            type="number"
            value={filters.minAvailableBeds}
            onChange={onUpdateMinBeds}
            error={minBedsInvalid}
            helperText={minBedsInvalid ? "Enter a non-negative number" : ""}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 3 }}>
          <FormControlLabel
            control={
              <Switch
                checked={filters.includeStale}
                onChange={(event) => onUpdateIncludeStale(event.target.checked)}
              />
            }
            label="Show stale"
          />
        </Grid>
      </Grid>

      {facilities.length ? (
        <Grid container spacing={1.5} className="capacity-list">
          {facilities.map((facility) => (
            <Grid size={{ xs: 12, md: 4 }} key={facility.id}>
              <Box className="hospital-card">
                <Stack direction="row" justifyContent="space-between" gap={1}>
                  <Box>
                    <Typography variant="subtitle2">{facility.name}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {facility.district}
                      {facility.distanceMeters
                        ? ` · ${metersLabel(facility.distanceMeters)}`
                        : ""}
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    color={hospitalCapacityColor(facility.emergencyCapacity)}
                    label={hospitalCapacityLabel(facility.emergencyCapacity)}
                  />
                </Stack>

                <Stack spacing={0.75}>
                  <Stack
                    direction="row"
                    justifyContent="space-between"
                    alignItems="center"
                  >
                    <Typography variant="body2">Available beds</Typography>
                    <Typography variant="subtitle2">
                      {facility.availableBeds}/{facility.totalBeds}
                    </Typography>
                  </Stack>
                  <LinearProgress
                    variant="determinate"
                    value={hospitalBedPercent(facility)}
                    color={
                      facility.emergencyCapacity === "available"
                        ? "success"
                        : facility.emergencyCapacity === "limited"
                          ? "warning"
                          : "error"
                    }
                  />
                </Stack>

                <Grid container spacing={1}>
                  <Grid size={6}>
                    <Fact
                      label="Emergency"
                      value={hospitalUnitStatusLabel(
                        facility.emergencyUnitStatus,
                      )}
                    />
                  </Grid>
                  <Grid size={6}>
                    <Fact
                      label="Ambulances"
                      value={`${facility.ambulancesAvailable}`}
                    />
                  </Grid>
                  <Grid size={6}>
                    <Fact label="ICU" value={`${facility.icuBedsAvailable}`} />
                  </Grid>
                  <Grid size={6}>
                    <Fact
                      label="Oxygen"
                      value={facility.oxygenAvailable ? "Available" : "No"}
                    />
                  </Grid>
                </Grid>

                {facility.stale ? (
                  <Alert severity="warning">
                    {facility.staleReason ?? "Capacity update is stale."}
                  </Alert>
                ) : null}

                <Typography variant="caption" color="text.secondary">
                  Updated {hospitalUpdatedLabel(facility.updatedAt)} via{" "}
                  {facility.source}
                </Typography>
              </Box>
            </Grid>
          ))}
        </Grid>
      ) : (
        <EmptyState
          title="No hospital capacity matches"
          detail="Adjust service, bed, capacity, or stale-data filters."
        />
      )}
    </Paper>
  );
}

export function ReliefPointPanel({
  loadMessage,
  loadState,
  onRefresh,
  reliefPoints,
}: {
  loadMessage: string;
  loadState: CapacityLoadState;
  onRefresh: () => void;
  reliefPoints: ReliefPointRecord[];
}) {
  return (
    <Paper className="surface capacity-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        gap={1.5}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Truck size={22} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h5">Relief distribution points</Typography>
            <Typography variant="caption" color="text.secondary">
              Food, water, medical, hygiene, and blanket distribution locations
              for affected communities.
            </Typography>
          </Box>
        </Stack>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live relief points"
                : loadState === "loading"
                  ? "Loading"
                  : loadState === "empty"
                    ? "No matches"
                    : "Fixture relief points"
            }
            color={
              loadState === "ready"
                ? "success"
                : loadState === "empty"
                  ? "default"
                  : "warning"
            }
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshCw size={16} />}
            onClick={onRefresh}
          >
            Refresh
          </Button>
        </Stack>
      </Stack>

      {loadState === "loading" ? (
        <LinearProgress />
      ) : loadState === "fallback" || loadState === "empty" ? (
        <Alert severity={loadState === "empty" ? "info" : "warning"}>
          {loadMessage}
        </Alert>
      ) : null}

      {reliefPoints.length > 0 ? (
        <Grid container spacing={2}>
          {reliefPoints.map((point) => (
            <Grid size={{ xs: 12, md: 6, lg: 4 }} key={point.id}>
              <Box className="capacity-card">
                <Stack spacing={1}>
                  <Stack
                    direction="row"
                    justifyContent="space-between"
                    alignItems="center"
                  >
                    <Typography variant="subtitle1" fontWeight={700}>
                      {point.name}
                    </Typography>
                    <Chip label={point.status} size="small" />
                  </Stack>
                  <Typography variant="body2" color="text.secondary">
                    {point.type} · {point.district}
                  </Typography>
                  <Typography variant="body2">{point.address}</Typography>
                  <Stack direction="row" flexWrap="wrap" gap={0.5}>
                    {point.stockCategories.map((stock) => (
                      <Chip
                        key={stock.category}
                        label={`${stock.category}: ${stock.quantity} ${stock.unit}`}
                        size="small"
                        variant="outlined"
                      />
                    ))}
                  </Stack>
                  {point.operatingHours ? (
                    <Typography variant="caption" color="text.secondary">
                      Hours: {point.operatingHours}
                    </Typography>
                  ) : null}
                  {point.eligibility ? (
                    <Alert severity="info" sx={{ py: 0.5 }}>
                      {point.eligibility}
                    </Alert>
                  ) : null}
                </Stack>
              </Box>
            </Grid>
          ))}
        </Grid>
      ) : (
        <EmptyState
          title="No relief points"
          detail="No relief distribution points are currently available."
        />
      )}
    </Paper>
  );
}

export function PredictionReviewMap({
  onSelect,
  predictions,
  selectedPredictionId,
}: {
  onSelect: (predictionId: string) => void;
  predictions: MLPredictionReview[];
  selectedPredictionId: string;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [5.586, -0.18],
      zoom: 8,
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
    if (!predictions.length) {
      return;
    }

    const bounds = L.latLngBounds([]);
    predictions.forEach((prediction) => {
      const color =
        severityRoles[severityRoleKey(prediction.severity)].foreground;
      const selected = prediction.id === selectedPredictionId;
      if (prediction.geometry?.coordinates?.[0]?.length) {
        const polygonPoints = prediction.geometry.coordinates[0].map(
          ([lng, lat]) => [lat, lng] as [number, number],
        );
        const polygon = L.polygon(polygonPoints, {
          color: selected ? "#0D1B3D" : color,
          fillColor: color,
          fillOpacity: selected ? 0.28 : 0.18,
          weight: selected ? 4 : 2,
        });
        polygon.bindPopup(
          `<strong>${prediction.community}</strong><br>${probabilityLabel(
            prediction.probability,
          )} · ${severityLabel(prediction.severity)}`,
        );
        polygon.on("click", () => onSelect(prediction.id));
        polygon.addTo(layer);
        polygonPoints.forEach((point) => bounds.extend(point));
        return;
      }

      const marker = L.circleMarker(
        [prediction.location.lat, prediction.location.lng],
        {
          radius: selected ? 12 : 8,
          color: "#FFFFFF",
          fillColor: color,
          fillOpacity: selected ? 0.95 : 0.75,
          weight: selected ? 4 : 2,
        },
      );
      marker.bindPopup(
        `<strong>${prediction.community}</strong><br>${probabilityLabel(
          prediction.probability,
        )} · ${severityLabel(prediction.severity)}`,
      );
      marker.on("click", () => onSelect(prediction.id));
      marker.addTo(layer);
      bounds.extend([prediction.location.lat, prediction.location.lng]);
    });

    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 12 });
    }
  }, [onSelect, predictions, selectedPredictionId]);

  return <Box ref={containerRef} className="prediction-review-map" />;
}

export function AlertWorkflowPanel({
  alerts,
  busy,
  feedback,
  form,
  loadState,
  onCreateDraft,
  onRunAction,
  onUpdateForm,
  selectedIncident,
}: {
  alerts: AuthorityAlertRecord[];
  busy: boolean;
  feedback: string;
  form: AlertFormState;
  loadState: AlertLoadState;
  onCreateDraft: () => void;
  onRunAction: (
    alert: AuthorityAlertRecord,
    action: "submit" | "approve" | "reject" | "emergency-override",
  ) => void;
  onUpdateForm: (
    key: keyof AlertFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  selectedIncident?: CommandIncident;
}) {
  const queueAlerts = alerts.filter(
    (alert) => alert.status !== "published" && alert.status !== "expired",
  );
  const radiusFieldsInvalid =
    form.targetType === "radius" &&
    (!form.targetLatitude.trim() ||
      !form.targetLongitude.trim() ||
      !form.targetRadiusMeters.trim());
  const customGeometryInvalid =
    form.targetType === "custom" &&
    form.targetGeometry.trim() !== "" &&
    !parseTargetGeometry(form.targetGeometry);

  return (
    <Paper className="surface alert-panel">
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        className="section-heading"
      >
        <BellRing size={21} color={nadaaBrand.colors.red} />
        <Box>
          <Typography variant="h6">Alert workflow</Typography>
          <Typography variant="caption" color="text.secondary">
            Draft, submit, approve, reject, or override with audit.
          </Typography>
        </Box>
      </Stack>

      {selectedIncident ? (
        <Stack spacing={1.5}>
          <Box>
            <Stack direction="row" justifyContent="space-between" gap={1}>
              <Typography variant="subtitle2">
                Draft from {selectedIncident.reference}
              </Typography>
              <Chip
                size="small"
                label={alertSeverityLabel(form.severity)}
                color={form.severity === "emergency" ? "error" : "warning"}
              />
            </Stack>
            <Typography variant="body2" color="text.secondary">
              {hazardLabel(selectedIncident.type)} · {selectedIncident.district}
            </Typography>
          </Box>

          <TextField
            size="small"
            label="Title"
            value={form.title}
            onChange={onUpdateForm("title")}
          />
          <TextField
            size="small"
            label="Message"
            value={form.message}
            onChange={onUpdateForm("message")}
            multiline
            minRows={3}
          />

          <Grid container spacing={1.25}>
            <Grid size={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Severity</InputLabel>
                <Select
                  label="Severity"
                  value={form.severity}
                  onChange={onUpdateForm("severity")}
                >
                  {alertSeverityOptions.map((severity) => (
                    <MenuItem key={severity} value={severity}>
                      {alertSeverityLabel(severity)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Target type</InputLabel>
                <Select
                  label="Target type"
                  value={form.targetType}
                  onChange={onUpdateForm("targetType")}
                >
                  {alertTargetTypeOptions.map((type) => (
                    <MenuItem key={type} value={type}>
                      {alertTargetTypeLabel(type)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={{ xs: 12, sm: 5 }}>
              <TextField
                size="small"
                label="Target IDs"
                value={form.targetIds}
                onChange={onUpdateForm("targetIds")}
                fullWidth
                disabled={form.targetType === "national"}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 7 }}>
              <TextField
                size="small"
                label="Target label"
                value={form.targetLabel}
                onChange={onUpdateForm("targetLabel")}
                fullWidth
              />
            </Grid>
            {form.targetType === "radius" ? (
              <>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Latitude"
                    value={form.targetLatitude}
                    onChange={onUpdateForm("targetLatitude")}
                    fullWidth
                    error={radiusFieldsInvalid && !form.targetLatitude.trim()}
                    helperText={
                      radiusFieldsInvalid && !form.targetLatitude.trim()
                        ? "Latitude is required"
                        : ""
                    }
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Longitude"
                    value={form.targetLongitude}
                    onChange={onUpdateForm("targetLongitude")}
                    fullWidth
                    error={radiusFieldsInvalid && !form.targetLongitude.trim()}
                    helperText={
                      radiusFieldsInvalid && !form.targetLongitude.trim()
                        ? "Longitude is required"
                        : ""
                    }
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Radius meters"
                    value={form.targetRadiusMeters}
                    onChange={onUpdateForm("targetRadiusMeters")}
                    fullWidth
                    error={
                      radiusFieldsInvalid && !form.targetRadiusMeters.trim()
                    }
                    helperText={
                      radiusFieldsInvalid && !form.targetRadiusMeters.trim()
                        ? "Radius is required"
                        : ""
                    }
                  />
                </Grid>
              </>
            ) : null}
            {form.targetType === "custom" ? (
              <Grid size={12}>
                <TextField
                  size="small"
                  label="Custom polygon JSON"
                  value={form.targetGeometry}
                  onChange={onUpdateForm("targetGeometry")}
                  multiline
                  minRows={3}
                  fullWidth
                  error={customGeometryInvalid}
                  helperText={
                    customGeometryInvalid ? "Enter valid GeoJSON polygon" : ""
                  }
                />
              </Grid>
            ) : null}
            <Grid size={6}>
              <TextField
                size="small"
                label="Starts"
                value={form.startsAt}
                onChange={onUpdateForm("startsAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
            <Grid size={6}>
              <TextField
                size="small"
                label="Expires"
                value={form.expiresAt}
                onChange={onUpdateForm("expiresAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
          </Grid>

          <AlertTargetPreview target={buildAlertTarget(form)} />

          <TextField
            size="small"
            label="Recommended action"
            value={form.recommendedAction}
            onChange={onUpdateForm("recommendedAction")}
          />
          <TextField
            size="small"
            label="Shelter IDs"
            value={form.shelterIds}
            onChange={onUpdateForm("shelterIds")}
          />
          <FormControlLabel
            control={
              <Switch
                checked={form.evacuationRequired}
                onChange={onUpdateForm("evacuationRequired")}
              />
            }
            label="Evacuation required"
          />
          <Button
            variant="contained"
            color="error"
            startIcon={<BellRing size={17} />}
            disabled={busy}
            onClick={onCreateDraft}
          >
            Create draft
          </Button>
        </Stack>
      ) : (
        <EmptyState
          title="No incident selected"
          detail="Choose an incident before drafting an alert."
        />
      )}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          gap={1}
        >
          <Typography variant="subtitle2">Approval queue</Typography>
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live"
                : loadState === "loading"
                  ? "Loading"
                  : "Fixture"
            }
            color={loadState === "ready" ? "success" : "warning"}
          />
        </Stack>
        {feedback ? (
          <Alert
            severity={
              loadState === "ready"
                ? "success"
                : loadState === "loading"
                  ? "info"
                  : "warning"
            }
          >
            {feedback}
          </Alert>
        ) : null}
        {queueAlerts.length ? (
          queueAlerts.slice(0, 4).map((alert) => (
            <Box key={alert.id} className="alert-queue-row">
              <Stack direction="row" justifyContent="space-between" gap={1}>
                <Box>
                  <Typography variant="subtitle2">{alert.title}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {alert.target.label} · {alertSeverityLabel(alert.severity)}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={alertStatusLabel(alert.status)}
                  color={alertStatusColor(alert.status)}
                />
              </Stack>
              <Stack
                direction="row"
                spacing={1}
                flexWrap="wrap"
                className="alert-actions"
              >
                {alert.status === "draft" ? (
                  <Button
                    size="small"
                    variant="outlined"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "submit")}
                  >
                    Submit
                  </Button>
                ) : null}
                {alert.status === "submitted" ? (
                  <>
                    <Button
                      size="small"
                      variant="contained"
                      color="success"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "approve")}
                    >
                      Approve
                    </Button>
                    <Button
                      size="small"
                      variant="outlined"
                      color="error"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "reject")}
                    >
                      Reject
                    </Button>
                  </>
                ) : null}
                {alert.status === "draft" ||
                alert.status === "submitted" ||
                alert.status === "rejected" ? (
                  <Button
                    size="small"
                    color="error"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "emergency-override")}
                  >
                    Override
                  </Button>
                ) : null}
              </Stack>
            </Box>
          ))
        ) : (
          <EmptyState
            title="No alerts in queue"
            detail="Create a draft to begin the approval workflow."
          />
        )}
      </Stack>
    </Paper>
  );
}

export function IncidentDetailPanel({
  abuseBusy,
  abuseFeedback,
  abuseForm,
  assignmentBusy,
  assignmentFeedback,
  assignmentForm,
  busy,
  duplicateCandidates,
  feedback,
  form,
  incident,
  mergeBusy,
  mergeFeedback,
  onAssign,
  onMergeDuplicates,
  onReviewAbuse,
  onToggleDuplicate,
  onUpdateAbuseForm,
  onUpdateAssignmentForm,
  onUpdateForm,
  onUpdateStatus,
  onVerify,
  selectedDuplicateIds,
}: {
  abuseBusy: boolean;
  abuseFeedback: string;
  abuseForm: AbuseReviewFormState;
  assignmentBusy: boolean;
  assignmentFeedback: string;
  assignmentForm: AssignmentFormState;
  busy: boolean;
  duplicateCandidates: DuplicateReviewCandidate[];
  feedback: string;
  form: IncidentStatusFormState;
  incident?: CommandIncident;
  mergeBusy: boolean;
  mergeFeedback: string;
  onAssign: () => void;
  onMergeDuplicates: () => void;
  onReviewAbuse: () => void;
  onToggleDuplicate: (incidentId: string) => void;
  onUpdateAbuseForm: (
    key: keyof AbuseReviewFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateAssignmentForm: (
    key: keyof AssignmentFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateForm: (
    key: keyof IncidentStatusFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateStatus: () => void;
  onVerify: () => void;
  selectedDuplicateIds: string[];
}) {
  if (!incident) {
    return (
      <Paper className="surface">
        <EmptyState
          title="No incident selected"
          detail="Choose a map marker or queue row to inspect the incident."
        />
      </Paper>
    );
  }

  const nextStatuses = incidentTransitionOptions[incident.status];
  const terminal = nextStatuses.length === 0;
  const resolutionRequired = requiresIncidentResolution(form.status);
  const canVerify = nextStatuses.includes("verified");
  const canAssign = canAssignIncident(incident.status);
  const activeAssignments = incident.assignments.filter(
    (assignment) => assignment.status === "active",
  );
  const canReviewAbuse =
    incident.source === "api" &&
    incident.status !== "closed" &&
    incident.status !== "false_report";
  const abuseResolutionRequired = abuseForm.decision === "false_report";
  const canMerge =
    incident.source === "api" && selectedDuplicateIds.length > 0 && !mergeBusy;

  return (
    <Paper className="surface detail-panel">
      <Stack
        direction="row"
        justifyContent="space-between"
        gap={1}
        className="section-heading"
      >
        <Box>
          <Typography variant="overline" color="secondary">
            Selected incident
          </Typography>
          <Typography variant="h6">{incident.reference}</Typography>
        </Box>
        <SeverityChip severity={incident.severity} />
      </Stack>

      <Typography variant="body2" color="text.secondary">
        {incident.description}
      </Typography>

      <Divider className="detail-divider" />

      <Grid container spacing={1.5}>
        <Grid size={6}>
          <Fact label="Hazard" value={hazardLabel(incident.type)} />
        </Grid>
        <Grid size={6}>
          <Fact label="Status" value={statusLabel(incident.status)} />
        </Grid>
        <Grid size={6}>
          <Fact label="People" value={`${incident.peopleAffected}`} />
        </Grid>
        <Grid size={6}>
          <Fact label="Responder ETA" value={incident.responderEta} />
        </Grid>
        <Grid size={12}>
          <Fact label="Assigned agency" value={incident.assignedAgency} />
        </Grid>
      </Grid>

      <Alert
        severity={
          incident.anonymous || !incident.privacy?.reporterContactVisible
            ? "info"
            : "success"
        }
        icon={<ShieldAlert size={18} />}
        className="privacy-alert"
      >
        <Stack spacing={0.75}>
          <Stack direction="row" spacing={1} flexWrap="wrap">
            <Chip size="small" label={privacyReporterLabel(incident)} />
            <Chip size="small" label={privacyContactLabel(incident)} />
            <Chip
              size="small"
              label={`${incident.privacy?.locationPrecision ?? "exact"} location`}
            />
          </Stack>
          <Typography variant="body2">
            {incident.privacy?.disclosure ??
              "Location is used for emergency response coordination."}
          </Typography>
          {incident.privacy?.notes?.length ? (
            <Typography variant="caption" color="text.secondary">
              {incident.privacy.notes[0]}
            </Typography>
          ) : null}
        </Stack>
      </Alert>

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Report safety review</Typography>
            <Typography variant="caption" color="text.secondary">
              {incident.abuseReviewRequired
                ? "Dispatcher review required"
                : "No active safety hold"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={abuseScoreLabel(incident.abuseScore)}
            color={incident.abuseReviewRequired ? "warning" : "default"}
          />
        </Stack>

        {incident.abuseReviewRequired ? (
          <Alert
            severity={incident.priorityReview ? "error" : "warning"}
            icon={<ShieldAlert size={18} />}
          >
            {incident.abuseReviewReason ||
              "Suspicious report signals need dispatcher review."}
          </Alert>
        ) : null}

        {incident.abuseSignals.length ? (
          <Stack spacing={1}>
            {incident.abuseSignals.map((signal) => (
              <Box className="abuse-signal-row" key={signal.code}>
                <Box>
                  <Typography variant="subtitle2">{signal.label}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {signal.detail}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={`${Math.round(signal.weight * 100)}%`}
                  color="warning"
                />
              </Box>
            ))}
          </Stack>
        ) : (
          <Alert severity="info">No suspicious report signals recorded.</Alert>
        )}

        {incident.abuseReviewDecision ? (
          <Alert severity="success">
            Last review: {abuseDecisionLabel(incident.abuseReviewDecision)}
            {incident.abuseReviewedAt
              ? ` at ${formatShortTime(incident.abuseReviewedAt)}`
              : ""}
          </Alert>
        ) : null}

        {abuseFeedback ? (
          <Alert
            severity={
              abuseFeedback.includes("needs") || abuseFeedback.includes("valid")
                ? "warning"
                : "success"
            }
          >
            {abuseFeedback}
          </Alert>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 5 }}>
            <FormControl fullWidth size="small" disabled={!canReviewAbuse}>
              <InputLabel>Decision</InputLabel>
              <Select
                label="Decision"
                value={abuseForm.decision}
                onChange={onUpdateAbuseForm("decision")}
              >
                {(["clear", "monitor", "false_report"] as const).map(
                  (decision) => (
                    <MenuItem value={decision} key={decision}>
                      {abuseDecisionLabel(decision)}
                    </MenuItem>
                  ),
                )}
              </Select>
            </FormControl>
          </Grid>
          <Grid size={{ xs: 12, sm: 7 }}>
            <TextField
              size="small"
              label="Review note"
              value={abuseForm.note}
              onChange={onUpdateAbuseForm("note")}
              disabled={!canReviewAbuse}
              fullWidth
            />
          </Grid>
        </Grid>

        {abuseResolutionRequired ? (
          <TextField
            size="small"
            label="False report resolution"
            value={abuseForm.resolutionNotes}
            onChange={onUpdateAbuseForm("resolutionNotes")}
            disabled={!canReviewAbuse}
            multiline
            minRows={3}
          />
        ) : null}

        <Button
          variant="outlined"
          disabled={
            abuseBusy ||
            !canReviewAbuse ||
            !abuseForm.note.trim() ||
            (abuseResolutionRequired && !abuseForm.resolutionNotes.trim())
          }
          onClick={onReviewAbuse}
          startIcon={<ShieldAlert size={17} />}
        >
          Save safety review
        </Button>

        {incident.source !== "api" ? (
          <Alert severity="info">
            Start incident-service to save fixture safety reviews.
          </Alert>
        ) : null}
      </Stack>

      <Divider className="detail-divider" />

      <Stack spacing={1}>
        <Typography variant="subtitle2">Response timeline</Typography>
        {incident.timelineEntries.map((event) => (
          <Box className="timeline-row" key={event}>
            <Typography variant="body2">{event}</Typography>
          </Box>
        ))}
      </Stack>

      {incident.duplicateCandidates.length ||
      duplicateCandidates.length ||
      incident.mergedIncidentIds.length ||
      incident.mergedIntoId ? (
        <>
          <Divider className="detail-divider" />
          <Stack spacing={1.25}>
            <Stack direction="row" justifyContent="space-between" gap={1}>
              <Box>
                <Typography variant="subtitle2">Duplicate review</Typography>
                <Typography variant="caption" color="text.secondary">
                  {duplicateCandidates.length
                    ? "Side-by-side candidate check"
                    : "No open candidates"}
                </Typography>
              </Box>
              <Chip
                size="small"
                label={`${duplicateCandidates.length} candidate${
                  duplicateCandidates.length === 1 ? "" : "s"
                }`}
                color={duplicateCandidates.length ? "warning" : "default"}
              />
            </Stack>

            {incident.mergedIntoId ? (
              <Alert severity="info">
                This report was merged into another incident and remains
                traceable in audit and timeline history.
              </Alert>
            ) : null}

            {incident.mergedIncidentIds.length ? (
              <Alert severity="success">
                {incident.mergedIncidentIds.length} duplicate report
                {incident.mergedIncidentIds.length === 1 ? "" : "s"} already
                merged into this incident.
              </Alert>
            ) : null}

            {mergeFeedback ? (
              <Alert
                severity={
                  mergeFeedback.includes("needs") ? "warning" : "success"
                }
              >
                {mergeFeedback}
              </Alert>
            ) : null}

            {duplicateCandidates.map((item) => (
              <Box className="duplicate-review-row" key={item.incident.id}>
                <FormControlLabel
                  control={
                    <Checkbox
                      size="small"
                      checked={selectedDuplicateIds.includes(item.incident.id)}
                      onChange={() => onToggleDuplicate(item.incident.id)}
                      disabled={incident.source !== "api" || mergeBusy}
                    />
                  }
                  label={item.incident.reference}
                />
                <Box className="duplicate-comparison">
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Selected
                    </Typography>
                    <Typography variant="body2">
                      {incident.description}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Candidate
                    </Typography>
                    <Typography variant="body2">
                      {item.incident.description}
                    </Typography>
                  </Box>
                </Box>
                <Stack direction="row" spacing={0.75} flexWrap="wrap">
                  <Chip
                    size="small"
                    label={`${Math.round(item.candidate.score * 100)}%`}
                  />
                  <Chip
                    size="small"
                    label={`${Math.round(item.candidate.distanceMeters)}m`}
                  />
                  <Chip
                    size="small"
                    label={`${item.candidate.minutesApart}m`}
                  />
                </Stack>
              </Box>
            ))}

            {incident.source !== "api" && duplicateCandidates.length ? (
              <Alert severity="info">
                Start incident-service to merge fixture duplicate reports.
              </Alert>
            ) : null}

            {duplicateCandidates.length ? (
              <Button
                variant="outlined"
                disabled={!canMerge}
                onClick={onMergeDuplicates}
                startIcon={<GitMerge size={17} />}
              >
                Merge selected
              </Button>
            ) : null}
          </Stack>
        </>
      ) : null}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Agency assignment</Typography>
            <Typography variant="caption" color="text.secondary">
              {canAssign ? "Dispatch coordination" : "Verification required"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={activeAssignments.length ? "Assigned" : "Unassigned"}
            color={activeAssignments.length ? "success" : "default"}
          />
        </Stack>

        {activeAssignments.length ? (
          <Stack spacing={1}>
            {activeAssignments.map((assignment) => (
              <Box className="assignment-row" key={assignment.id}>
                <Box>
                  <Typography variant="subtitle2">
                    {assignment.agencyName}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {assignment.responderLead || "Response lead pending"}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={assignment.priority}
                  color={assignment.priority === "urgent" ? "error" : "warning"}
                />
              </Box>
            ))}
          </Stack>
        ) : null}

        {assignmentFeedback ? (
          <Alert
            severity={
              assignmentFeedback.includes("needs") ? "warning" : "success"
            }
          >
            {assignmentFeedback}
          </Alert>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 7 }}>
            <FormControl fullWidth size="small" disabled={!canAssign}>
              <InputLabel>Agency</InputLabel>
              <Select
                label="Agency"
                value={assignmentForm.agencyId}
                onChange={onUpdateAssignmentForm("agencyId")}
              >
                {assignmentAgencyOptions.map((agency) => (
                  <MenuItem value={agency.id} key={agency.id}>
                    {agency.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid size={{ xs: 12, sm: 5 }}>
            <FormControl fullWidth size="small" disabled={!canAssign}>
              <InputLabel>Priority</InputLabel>
              <Select
                label="Priority"
                value={assignmentForm.priority}
                onChange={onUpdateAssignmentForm("priority")}
              >
                {(["low", "normal", "high", "urgent"] as const).map(
                  (priority) => (
                    <MenuItem value={priority} key={priority}>
                      {priority}
                    </MenuItem>
                  ),
                )}
              </Select>
            </FormControl>
          </Grid>
        </Grid>

        <TextField
          size="small"
          label="Instructions"
          value={assignmentForm.instructions}
          onChange={onUpdateAssignmentForm("instructions")}
          disabled={!canAssign}
          multiline
          minRows={2}
        />

        <TextField
          size="small"
          label="Responder lead"
          value={assignmentForm.responderLead}
          onChange={onUpdateAssignmentForm("responderLead")}
          disabled={!canAssign}
        />

        <Button
          variant="outlined"
          disabled={
            assignmentBusy || !canAssign || !assignmentForm.instructions.trim()
          }
          onClick={onAssign}
          startIcon={<Truck size={17} />}
        >
          Assign agency
        </Button>
      </Stack>

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Status workflow</Typography>
            <Typography variant="caption" color="text.secondary">
              {terminal
                ? "Terminal incident state"
                : "Audited dispatcher action"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={incident.source === "api" ? "Live" : "Fixture"}
            color={incident.source === "api" ? "success" : "warning"}
          />
        </Stack>

        {feedback ? (
          <Alert
            severity={
              feedback.includes("needs") || feedback.includes("valid")
                ? "warning"
                : "success"
            }
          >
            {feedback}
          </Alert>
        ) : null}

        <FormControl fullWidth size="small" disabled={terminal}>
          <InputLabel>Next status</InputLabel>
          <Select
            label="Next status"
            value={form.status}
            onChange={onUpdateForm("status")}
          >
            {(nextStatuses.length ? nextStatuses : [incident.status]).map(
              (status) => (
                <MenuItem value={status} key={status}>
                  {statusLabel(status)}
                </MenuItem>
              ),
            )}
          </Select>
        </FormControl>

        <TextField
          size="small"
          label="Status note"
          value={form.note}
          onChange={onUpdateForm("note")}
          multiline
          minRows={2}
        />

        {resolutionRequired ? (
          <TextField
            size="small"
            label="Resolution notes"
            value={form.resolutionNotes}
            onChange={onUpdateForm("resolutionNotes")}
            multiline
            minRows={3}
          />
        ) : null}
      </Stack>

      <Stack direction="row" spacing={1} className="detail-actions">
        <Button
          variant="contained"
          disabled={busy || !canVerify}
          onClick={onVerify}
          startIcon={<CheckCheck size={17} />}
        >
          Verify
        </Button>
        <Button
          variant="outlined"
          disabled={
            busy ||
            terminal ||
            (resolutionRequired && !form.resolutionNotes.trim())
          }
          onClick={onUpdateStatus}
          startIcon={<Truck size={17} />}
        >
          Update status
        </Button>
      </Stack>
    </Paper>
  );
}

export function Fact({ label, value }: { label: string; value: string }) {
  return (
    <Box className="fact">
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="subtitle2">{value}</Typography>
    </Box>
  );
}

export function PrivacyChip({ incident }: { incident: CommandIncident }) {
  return (
    <Chip
      size="small"
      label={privacyReporterLabel(incident)}
      color={incident.anonymous ? "warning" : "default"}
    />
  );
}

function privacyReporterLabel(incident: CommandIncident) {
  if (incident.anonymous) {
    return "Anonymous";
  }
  if (incident.privacy?.reporterIdentityVisible) {
    return "Identity visible";
  }
  return "Identity hidden";
}

function privacyContactLabel(incident: CommandIncident) {
  if (incident.privacy?.reporterContactVisible) {
    return "Contact visible";
  }
  if (incident.contactPermission) {
    return "Contact restricted";
  }
  return "Contact denied";
}

export function StatusLine({
  color,
  label,
  value,
}: {
  color: "success" | "warning" | "default";
  label: string;
  value: string;
}) {
  return (
    <Stack
      direction="row"
      justifyContent="space-between"
      alignItems="center"
      gap={1}
    >
      <Typography variant="body2">{label}</Typography>
      <Chip size="small" label={value} color={color} />
    </Stack>
  );
}

export function EmptyState({
  detail,
  title,
}: {
  detail: string;
  title: string;
}) {
  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      spacing={1}
      className="empty-state"
    >
      <Crosshair size={28} color={nadaaBrand.colors.slate} />
      <Typography variant="subtitle2">{title}</Typography>
      <Typography variant="body2" color="text.secondary" textAlign="center">
        {detail}
      </Typography>
    </Stack>
  );
}
