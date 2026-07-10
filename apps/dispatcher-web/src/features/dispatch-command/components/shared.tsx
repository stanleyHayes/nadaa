import { type ReactNode, useEffect, useRef } from "react";
import {
  Alert,
  Box,
  Chip,
  FormControl,
  Grid,
  InputLabel,
  Select,
  Stack,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
  AlertCircle,
  AlertOctagon,
  AlertTriangle,
  Bug,
  Car,
  CheckCircle2,
  CloudLightning,
  Crosshair,
  Flame,
  HeartPulse,
  Info,
  MapPin,
  Mountain,
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
  HazardType,
  ReliefPointRecord,
  RiskLevel,
  RoadClosureRecord,
} from "@nadaa/shared-types";

import type { CommandIncident, MLPredictionReview } from "../types";
import {
  alertTargetSummary,
  alertTargetTypeLabel,
  alertTargetWarnings,
  hazardLabel,
  probabilityLabel,
  severityLabel,
} from "../utils";

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

export const severityIconMap = {
  CheckCircle2,
  AlertTriangle,
  AlertOctagon,
  Info,
};

export const hazardIconMap: Record<
  Hazard,
  React.ComponentType<{ size?: number }>
> = {
  flood: MapPin,
  fire: Flame,
  medical: HeartPulse,
  geological: Mountain,
  road: Car,
  storm: CloudLightning,
  disease: Bug,
  default: AlertCircle,
};

export function severityRoleKey(severity: string): Severity {
  if (severity === "emergency") return "severe";
  if (severity === "moderate") return "medium";
  if (severity in severityRoles) return severity as Severity;
  return "info";
}

export function hazardRoleKey(type: string): Hazard {
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

export const closureSeverityColors: Record<string, string> = {
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

export function privacyReporterLabel(incident: CommandIncident) {
  if (incident.anonymous) {
    return "Anonymous";
  }
  if (incident.privacy?.reporterIdentityVisible) {
    return "Identity visible";
  }
  return "Identity hidden";
}

export function privacyContactLabel(incident: CommandIncident) {
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
      <Crosshair size={28} color="var(--nadaa-slate)" />
      <Typography variant="subtitle2">{title}</Typography>
      <Typography variant="body2" color="text.secondary" textAlign="center">
        {detail}
      </Typography>
    </Stack>
  );
}
