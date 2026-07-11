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
  AlertOctagon,
  AlertTriangle,
  CheckCircle2,
  Crosshair,
  Flame,
  Info,
  ShieldAlert,
  Waves,
} from "lucide-react";
import { nadaaBrand, severityRoles, hazardRoles } from "@nadaa/brand";
import type {
  AlertSeverity,
  AlertTarget,
  HazardType,
  ImageryGeoJSONFeatureCollection,
  ImageryGeoJSONFeatureProperties,
  RiskLevel,
} from "@nadaa/shared-types";
import { severityColors } from "../data";
import type { CommandIncident } from "../types";
import {
  alertSeverityLabel,
  alertTargetSummary,
  alertTargetTypeLabel,
  alertTargetWarnings,
  hazardLabel,
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

const severityIconComponents = {
  CheckCircle2,
  AlertTriangle,
  AlertOctagon,
  Info,
} as const;

const alertSeverityRole: Record<AlertSeverity, keyof typeof severityRoles> = {
  advisory: "info",
  watch: "low",
  warning: "medium",
  severe_warning: "high",
  emergency: "severe",
};

const incidentSeverityRole: Record<RiskLevel, keyof typeof severityRoles> = {
  low: "low",
  moderate: "medium",
  high: "high",
  severe: "severe",
  emergency: "severe",
};

function chipLabelForSeverity(severity: RiskLevel | AlertSeverity) {
  if (
    ["advisory", "watch", "warning", "severe_warning", "emergency"].includes(
      severity,
    )
  ) {
    return alertSeverityLabel(severity as AlertSeverity);
  }
  return severityLabel(severity as RiskLevel);
}

export function SeverityChip({
  label,
  severity,
  size = "small",
}: {
  label?: string;
  severity: RiskLevel | AlertSeverity;
  size?: "small" | "medium";
}) {
  const directRole = severityRoles[severity as keyof typeof severityRoles];
  const roleKey = directRole
    ? (severity as keyof typeof severityRoles)
    : (incidentSeverityRole[severity as RiskLevel] ??
      alertSeverityRole[severity as AlertSeverity] ??
      "info");
  const role = severityRoles[roleKey];
  const Icon = severityIconComponents[role.icon];
  return (
    <Chip
      size={size}
      label={label ?? chipLabelForSeverity(severity)}
      icon={<Icon size={14} />}
      className="severity-chip"
      style={{
        backgroundColor: role.background,
        color: role.foreground,
        borderColor: role.border,
      }}
    />
  );
}

const hazardIconComponents: Record<
  string,
  React.ComponentType<{ size?: number }>
> = {
  flood: Waves,
  fire: Flame,
  storm: Waves,
  medical: ShieldAlert,
};

const hazardRoleMap: Record<string, keyof typeof hazardRoles> = {
  flood: "flood",
  fire: "fire",
  road_crash: "road",
  building_collapse: "geological",
  medical_emergency: "medical",
  disease_outbreak: "disease",
  electrical_hazard: "fire",
  blocked_drain: "default",
  landslide: "geological",
  marine_accident: "default",
  storm: "storm",
  tidal_wave: "default",
  security_incident: "default",
  other: "default",
};

export function HazardChip({
  hazard,
  label,
  size = "small",
}: {
  hazard: HazardType;
  label?: string;
  size?: "small" | "medium";
}) {
  const roleKey = hazardRoleMap[hazard] ?? "default";
  const role = hazardRoles[roleKey];
  const Icon = hazardIconComponents[hazard] ?? ShieldAlert;
  return (
    <Chip
      size={size}
      label={label ?? hazardLabel(hazard)}
      icon={<Icon size={14} />}
      className="hazard-chip"
      style={{
        backgroundColor: role.background,
        color: role.foreground,
        borderColor: role.border,
      }}
    />
  );
}

export function ScrollableTable({
  children,
  label,
}: {
  children: React.ReactNode;
  label: string;
}) {
  return (
    <Box
      className="incident-table"
      tabIndex={0}
      role="region"
      aria-label={label}
    >
      {children}
    </Box>
  );
}

export function IncidentMap({
  incidents,
  imageryFeatures,
  onSelect,
  selectedIncidentId,
}: {
  incidents: CommandIncident[];
  imageryFeatures?: ImageryGeoJSONFeatureCollection;
  onSelect: (incidentId: string) => void;
  selectedIncidentId?: string;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);
  const imageryLayerRef = useRef<L.LayerGroup | null>(null);

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
    imageryLayerRef.current = L.layerGroup().addTo(map);

    return () => {
      map.remove();
      mapRef.current = null;
      layerRef.current = null;
      imageryLayerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const layer = layerRef.current;
    const map = mapRef.current;
    if (!layer || !map) {
      return;
    }

    layer.clearLayers();
    if (!incidents.length) {
      return;
    }

    const bounds = L.latLngBounds([]);
    incidents.forEach((incident) => {
      const isSelected = incident.id === selectedIncidentId;
      const marker = L.circleMarker(
        [incident.location.lat, incident.location.lng],
        {
          radius: isSelected ? 13 : 9,
          color: "#FFFFFF",
          weight: isSelected ? 4 : 2,
          fillColor: severityColors[incident.severity],
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

    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 13 });
    }
  }, [incidents, onSelect, selectedIncidentId]);

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

  useEffect(() => {
    const layer = imageryLayerRef.current;
    const map = mapRef.current;
    if (!layer || !map) {
      return;
    }

    layer.clearLayers();
    if (!imageryFeatures?.features?.length) {
      return;
    }

    const colors: Record<string, string> = {
      drone: nadaaBrand.colors.gold,
      satellite: nadaaBrand.colors.navy,
      other: nadaaBrand.colors.slate,
    };

    const geoJson = L.geoJSON(
      imageryFeatures as unknown as GeoJSON.GeoJsonObject,
      {
        style: (feature) => {
          const source =
            (feature?.properties?.source as string | undefined) ?? "other";
          const color = colors[source] ?? nadaaBrand.colors.slate;
          return {
            color,
            fillColor: color,
            fillOpacity: 0.12,
            weight: 2,
          };
        },
        onEachFeature: (feature, leafletFeature) => {
          const props = feature.properties as
            ImageryGeoJSONFeatureProperties | undefined;
          if (props) {
            leafletFeature.bindPopup(
              `<strong>${props.reference}</strong><br>Source: ${props.source}<br>Resolution: ${props.resolutionMeters} m<br>Captured: ${new Date(
                props.captureTime,
              ).toLocaleString()}`,
            );
          }
        },
      },
    ).addTo(layer);

    const bounds = geoJson.getBounds();
    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.08), { animate: true, maxZoom: 13 });
    }
  }, [imageryFeatures]);

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
      <Stack
        direction="row"
        sx={{
          justifyContent: "space-between",
          gap: 1
        }}>
        <Box>
          <Typography variant="subtitle2">Affected area preview</Typography>
          <Typography variant="caption" sx={{
            color: "text.secondary"
          }}>
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

export function Fact({ label, value }: { label: string; value: string }) {
  return (
    <Box className="fact">
      <Typography variant="caption" sx={{
        color: "text.secondary"
      }}>
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
      sx={{
        justifyContent: "space-between",
        alignItems: "center",
        gap: 1
      }}>
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
      spacing={1}
      className="empty-state"
      sx={{
        alignItems: "center",
        justifyContent: "center"
      }}>
      <span aria-hidden="true" className="empty-state__icon">
        <Crosshair size={28} strokeWidth={1.75} />
      </span>
      <Typography variant="subtitle2">{title}</Typography>
      <Typography
        variant="body2"
        sx={{
          color: "text.secondary",
          textAlign: "center"
        }}>
        {detail}
      </Typography>
    </Stack>
  );
}

/**
 * Loading skeleton — shimmering placeholder rows shown while content loads,
 * in place of a progress bar or spinner.
 */
export function SkeletonRows({
  rows = 3,
  height = 46,
}: {
  rows?: number;
  height?: number;
}) {
  return (
    <Stack aria-hidden spacing={1.25} sx={{ my: 1 }}>
      {Array.from({ length: rows }).map((_, index) => (
        <div className="nadaa-skeleton" key={index} style={{ height }} />
      ))}
    </Stack>
  );
}
