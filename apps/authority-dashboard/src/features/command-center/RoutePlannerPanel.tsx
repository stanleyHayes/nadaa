import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { MapPinned, Navigation, Route } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { RoutePlanResponse, RouteWaypointType } from "@nadaa/shared-types";
import { ROUTE_API_BASE } from "../../app/config";
import { CommandSelect, EmptyState } from "./components";
import type { RoutePlanFormWaypointType } from "./types";

const waypointTypeOptions: RoutePlanFormWaypointType[] = [
  "shelter",
  "higher_ground",
  "manual",
];

function waypointTypeLabel(type: RoutePlanFormWaypointType) {
  return type
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(" ");
}

function isValidCoordinate(lat: string, lng: string) {
  const latitude = Number(lat);
  const longitude = Number(lng);
  return (
    Number.isFinite(latitude) &&
    Number.isFinite(longitude) &&
    latitude >= -90 &&
    latitude <= 90 &&
    longitude >= -180 &&
    longitude <= 180
  );
}

function formatDistance(meters: number) {
  if (meters >= 1000) {
    return `${Math.round((meters / 1000) * 10) / 10} km`;
  }
  return `${meters} m`;
}

function formatDuration(minutes: number) {
  if (minutes < 60) {
    return `${minutes} min`;
  }
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  return remainingMinutes
    ? `${hours} hr ${remainingMinutes} min`
    : `${hours} hr`;
}

interface RoutePlannerPanelProps {
  selectedIncident?: {
    id: string;
    reference: string;
    location: { lat: number; lng: number };
  };
}

export function RoutePlannerPanel({
  selectedIncident,
}: RoutePlannerPanelProps) {
  const [originLat, setOriginLat] = useState("");
  const [originLng, setOriginLng] = useState("");
  const [destinationLat, setDestinationLat] = useState("");
  const [destinationLng, setDestinationLng] = useState("");
  const [waypointType, setWaypointType] =
    useState<RoutePlanFormWaypointType>("shelter");
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState("");
  const [result, setResult] = useState<RoutePlanResponse | null>(null);

  useEffect(() => {
    if (selectedIncident) {
      setOriginLat(selectedIncident.location.lat.toFixed(5));
      setOriginLng(selectedIncident.location.lng.toFixed(5));
      setFeedback(`Origin pre-filled from ${selectedIncident.reference}.`);
    }
  }, [selectedIncident?.id]);

  const originValid = isValidCoordinate(originLat, originLng);
  const destinationValid = isValidCoordinate(destinationLat, destinationLng);
  const manualRequiresDestination = waypointType === "manual";
  const canPlan =
    originValid && (!manualRequiresDestination || destinationValid);

  const updateWaypointType = (event: SelectChangeEvent) => {
    setWaypointType(event.target.value as RouteWaypointType);
  };

  const planRoute = async () => {
    if (!canPlan) {
      setFeedback("Enter valid origin coordinates before planning a route.");
      return;
    }

    setBusy(true);
    setFeedback("");
    setResult(null);

    const body: Record<string, unknown> = {
      origin: { lat: Number(originLat), lng: Number(originLng) },
      waypointType: waypointType as RouteWaypointType,
    };

    if (destinationValid) {
      body.destination = {
        lat: Number(destinationLat),
        lng: Number(destinationLng),
      };
    }

    try {
      const response = await fetch(`${ROUTE_API_BASE}/routes/plan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!response.ok) {
        const payload = (await response.json()) as {
          error?: { message?: string };
        };
        throw new Error(
          payload.error?.message ?? `route-service returned ${response.status}`,
        );
      }
      const plan = (await response.json()) as RoutePlanResponse;
      setResult(plan);
      setFeedback(
        `Route planned: ${formatDistance(plan.distanceMeters)} · ${formatDuration(plan.estimatedDurationMinutes)}`,
      );
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Route planning needs route-service running on the configured URL.",
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <Paper className="surface route-panel">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Route size={21} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h6">Evacuation route planner</Typography>
            <Typography variant="caption" color="text.secondary">
              Decision-support routes to shelter, higher ground, or a manual
              waypoint
            </Typography>
          </Box>
        </Stack>
      </Stack>

      {feedback ? (
        <Alert severity={result ? "success" : "warning"} className="feed-alert">
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Origin</Typography>
        <Grid container spacing={1}>
          <Grid size={6}>
            <TextField
              label="Origin latitude"
              size="small"
              fullWidth
              required
              value={originLat}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setOriginLat(event.target.value)
              }
              inputProps={{ inputMode: "decimal" }}
              error={Boolean(originLat) && !originValid}
              helperText={
                Boolean(originLat) && !originValid
                  ? "Enter a valid latitude"
                  : ""
              }
            />
          </Grid>
          <Grid size={6}>
            <TextField
              label="Origin longitude"
              size="small"
              fullWidth
              required
              value={originLng}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                setOriginLng(event.target.value)
              }
              inputProps={{ inputMode: "decimal" }}
              error={Boolean(originLng) && !originValid}
              helperText={
                Boolean(originLng) && !originValid
                  ? "Enter a valid longitude"
                  : ""
              }
            />
          </Grid>
        </Grid>

        <FormControl fullWidth size="small">
          <InputLabel>Waypoint type</InputLabel>
          <Select
            label="Waypoint type"
            value={waypointType}
            onChange={updateWaypointType}
          >
            {waypointTypeOptions.map((type) => (
              <MenuItem value={type} key={type}>
                {waypointTypeLabel(type)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {waypointType === "manual" ? (
          <>
            <Typography variant="subtitle2">Destination</Typography>
            <Grid container spacing={1}>
              <Grid size={6}>
                <TextField
                  label="Destination latitude"
                  size="small"
                  fullWidth
                  required={manualRequiresDestination}
                  value={destinationLat}
                  onChange={(event: ChangeEvent<HTMLInputElement>) =>
                    setDestinationLat(event.target.value)
                  }
                  inputProps={{ inputMode: "decimal" }}
                  error={
                    Boolean(destinationLat) &&
                    !isValidCoordinate(destinationLat, destinationLng)
                  }
                  helperText={
                    Boolean(destinationLat) &&
                    !isValidCoordinate(destinationLat, destinationLng)
                      ? "Enter a valid latitude"
                      : ""
                  }
                />
              </Grid>
              <Grid size={6}>
                <TextField
                  label="Destination longitude"
                  size="small"
                  fullWidth
                  required={manualRequiresDestination}
                  value={destinationLng}
                  onChange={(event: ChangeEvent<HTMLInputElement>) =>
                    setDestinationLng(event.target.value)
                  }
                  inputProps={{ inputMode: "decimal" }}
                  error={
                    Boolean(destinationLng) &&
                    !isValidCoordinate(destinationLat, destinationLng)
                  }
                  helperText={
                    Boolean(destinationLng) &&
                    !isValidCoordinate(destinationLat, destinationLng)
                      ? "Enter a valid longitude"
                      : ""
                  }
                />
              </Grid>
            </Grid>
          </>
        ) : null}

        <Button
          type="button"
          variant="contained"
          startIcon={<Navigation size={17} />}
          disabled={busy || !canPlan}
          onClick={() => void planRoute()}
        >
          {busy ? "Planning" : "Plan route"}
        </Button>
      </Stack>

      {result ? (
        <>
          <Divider className="detail-divider" />
          <RouteResultSummary plan={result} />
          <Box className="route-map-frame">
            <RouteMap plan={result} />
          </Box>
          <RouteSegmentList segments={result.segments} />
          <Alert severity="info" className="route-disclaimer">
            {result.disclaimer}
          </Alert>
        </>
      ) : (
        <EmptyState
          title="No route planned"
          detail="Enter coordinates and choose a waypoint type, then plan a route."
        />
      )}
    </Paper>
  );
}

function RouteResultSummary({ plan }: { plan: RoutePlanResponse }) {
  return (
    <Grid container spacing={1.5}>
      <Grid size={6}>
        <Box className="route-stat">
          <Typography variant="caption" color="text.secondary">
            Distance
          </Typography>
          <Typography variant="subtitle2">
            {formatDistance(plan.distanceMeters)}
          </Typography>
        </Box>
      </Grid>
      <Grid size={6}>
        <Box className="route-stat">
          <Typography variant="caption" color="text.secondary">
            Est. walking time
          </Typography>
          <Typography variant="subtitle2">
            {formatDuration(plan.estimatedDurationMinutes)}
          </Typography>
        </Box>
      </Grid>
      {plan.targetShelter ? (
        <Grid size={12}>
          <Box className="route-stat">
            <Typography variant="caption" color="text.secondary">
              Target shelter
            </Typography>
            <Stack direction="row" spacing={1} alignItems="center">
              <MapPinned size={16} />
              <Typography variant="subtitle2">
                {plan.targetShelter.name}
              </Typography>
              <Chip
                size="small"
                label={plan.targetShelter.status}
                color="success"
              />
            </Stack>
          </Box>
        </Grid>
      ) : null}
      <Grid size={6}>
        <Box className="route-stat">
          <Typography variant="caption" color="text.secondary">
            Avoided closures
          </Typography>
          <Typography variant="subtitle2">
            {plan.avoidedClosures.length}
          </Typography>
        </Box>
      </Grid>
      <Grid size={6}>
        <Box className="route-stat">
          <Typography variant="caption" color="text.secondary">
            Avoided risk zones
          </Typography>
          <Typography variant="subtitle2">
            {plan.avoidedRiskZones.length}
          </Typography>
        </Box>
      </Grid>
    </Grid>
  );
}

function RouteSegmentList({
  segments,
}: {
  segments: RoutePlanResponse["segments"];
}) {
  if (!segments.length) {
    return null;
  }

  return (
    <Stack spacing={1}>
      <Typography variant="subtitle2">Route segments</Typography>
      <Stack spacing={1}>
        {segments.map((segment, index) => (
          <Box
            className="route-segment"
            key={`${segment.start.lat}-${segment.start.lng}-${index}`}
          >
            <Typography variant="body2">
              {index + 1}. {formatDistance(segment.distanceMeters)}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {segment.start.lat.toFixed(5)}, {segment.start.lng.toFixed(5)} →{" "}
              {segment.end.lat.toFixed(5)}, {segment.end.lng.toFixed(5)}
            </Typography>
          </Box>
        ))}
      </Stack>
    </Stack>
  );
}

function RouteMap({ plan }: { plan: RoutePlanResponse }) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  const routePoints = useMemo(
    () => plan.route.map((point) => [point.lat, point.lng] as [number, number]),
    [plan.route],
  );

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const center = routePoints[0] ?? [5.56, -0.2];
    const map = L.map(containerRef.current, {
      center,
      zoom: 13,
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
    if (!layer || !map || routePoints.length < 2) {
      return;
    }

    layer.clearLayers();

    const polyline = L.polyline(routePoints, {
      color: nadaaBrand.colors.navy,
      weight: 5,
      opacity: 0.9,
    }).addTo(layer);

    routePoints.forEach((point, index) => {
      const marker = L.circleMarker(point, {
        radius: index === 0 || index === routePoints.length - 1 ? 9 : 6,
        color: "#FFFFFF",
        weight: 2,
        fillColor:
          index === 0
            ? nadaaBrand.colors.green
            : index === routePoints.length - 1
              ? nadaaBrand.colors.red
              : nadaaBrand.colors.navy,
        fillOpacity: 1,
      });
      marker.bindPopup(
        index === 0
          ? "Origin"
          : index === routePoints.length - 1
            ? "Destination"
            : `Waypoint ${index + 1}`,
      );
      marker.addTo(layer);
    });

    map.fitBounds(polyline.getBounds().pad(0.18), { animate: true });
  }, [routePoints]);

  return <Box ref={containerRef} className="route-map" />;
}
