import { FormEvent, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  FormControl,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  Grid,
  Paper,
  Radio,
  RadioGroup,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { Loader2, LocateFixed, MapPin, Navigation, Shield } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  RoutePlanRequest,
  RoutePlanResponse,
  RouteWaypointType,
} from "@nadaa/shared-types";
import { ROUTE_API_BASE } from "@/app/config";
import { extractAPIError, formatDateTime, formatDistance } from "../utils";

type PlannerStatus =
  | { status: "idle"; message?: string }
  | { status: "loading"; message: string }
  | { status: "success"; message?: string }
  | { status: "error"; message: string };

interface RouteForm {
  originLat: string;
  originLng: string;
  destLat: string;
  destLng: string;
  waypointType: RouteWaypointType;
}

const initialForm: RouteForm = {
  originLat: "5.560000",
  originLng: "-0.200000",
  destLat: "",
  destLng: "",
  waypointType: "shelter",
};

const waypointOptions: { value: RouteWaypointType; label: string }[] = [
  { value: "shelter", label: "Nearest safe shelter" },
  { value: "higher_ground", label: "Higher ground nearby" },
  { value: "manual", label: "Manual destination" },
];

function parseCoordinate(text: string): number | null {
  const value = Number(text.trim());
  return Number.isFinite(value) ? value : null;
}

function isValidCoordinate(lat: number, lng: number): boolean {
  return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
}

export default function RoutePlanner() {
  const [form, setForm] = useState<RouteForm>(initialForm);
  const [errors, setErrors] = useState<
    Partial<Record<keyof RouteForm, string>>
  >({});
  const [plannerState, setPlannerState] = useState<PlannerStatus>({
    status: "idle",
  });
  const [result, setResult] = useState<RoutePlanResponse | null>(null);

  const clearError = (key: keyof RouteForm) => {
    setErrors((current) => {
      if (!current[key]) return current;
      const next = { ...current };
      delete next[key];
      return next;
    });
  };

  const updateForm = <Key extends keyof RouteForm>(
    key: Key,
    value: RouteForm[Key],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
    clearError(key);
  };

  const useOriginLocation = () => {
    if (!navigator.geolocation) {
      setPlannerState({
        status: "error",
        message: "Location is not available on this device.",
      });
      return;
    }

    setPlannerState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setForm((current) => ({
          ...current,
          originLat: position.coords.latitude.toFixed(6),
          originLng: position.coords.longitude.toFixed(6),
        }));
        setPlannerState({ status: "idle" });
      },
      () => {
        setPlannerState({
          status: "error",
          message:
            "Location permission was not granted. Enter coordinates manually.",
        });
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  };

  const validateForm = (): RoutePlanRequest | null => {
    const nextErrors: Partial<Record<keyof RouteForm, string>> = {};

    const originLat = parseCoordinate(form.originLat);
    const originLng = parseCoordinate(form.originLng);
    if (
      originLat === null ||
      originLng === null ||
      !isValidCoordinate(originLat, originLng)
    ) {
      nextErrors.originLat = "Enter a valid latitude.";
      nextErrors.originLng = "Enter a valid longitude.";
    }

    let destination = undefined;
    if (form.waypointType === "manual") {
      const destLat = parseCoordinate(form.destLat);
      const destLng = parseCoordinate(form.destLng);
      if (
        destLat === null ||
        destLng === null ||
        !isValidCoordinate(destLat, destLng)
      ) {
        nextErrors.destLat = "Enter a valid destination latitude.";
        nextErrors.destLng = "Enter a valid destination longitude.";
      } else {
        destination = { lat: destLat, lng: destLng };
      }
    }

    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return null;
    }

    return {
      origin: { lat: originLat!, lng: originLng! },
      destination,
      waypointType: form.waypointType,
      avoidRiskLevels: ["severe", "emergency"],
    };
  };

  const submitPlan = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setResult(null);

    if (!navigator.onLine) {
      setPlannerState({
        status: "error",
        message:
          "Route planning needs a connection. Try again when you are online.",
      });
      return;
    }

    const request = validateForm();
    if (!request) {
      setPlannerState({
        status: "error",
        message: "Please correct the highlighted fields.",
      });
      return;
    }

    setPlannerState({ status: "loading", message: "Planning route" });

    try {
      const response = await fetch(`${ROUTE_API_BASE}/routes/plan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
      });
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as RoutePlanResponse;
      setResult(payload);
      setPlannerState({
        status: "success",
        message: `Route planned at ${formatDateTime(payload.generatedAt)}.`,
      });
      setErrors({});
    } catch (error) {
      setPlannerState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not plan route.",
      });
    }
  };

  return (
    <Paper className="surface">
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        className="section-heading"
      >
        <Navigation size={21} color={nadaaBrand.colors.red} />
        <Box>
          <Typography variant="h6">Plan evacuation route</Typography>
          <Typography variant="caption" color="text.secondary">
            Find a safer path to shelter or higher ground
          </Typography>
        </Box>
      </Stack>

      <Stack component="form" spacing={1.5} onSubmit={submitPlan} noValidate>
        <FormControl>
          <FormLabel id="route-waypoint-type-label">Destination type</FormLabel>
          <RadioGroup
            row
            aria-labelledby="route-waypoint-type-label"
            value={form.waypointType}
            onChange={(event) =>
              updateForm(
                "waypointType",
                event.target.value as RouteWaypointType,
              )
            }
          >
            {waypointOptions.map((option) => (
              <FormControlLabel
                key={option.value}
                value={option.value}
                control={<Radio size="small" />}
                label={option.label}
              />
            ))}
          </RadioGroup>
        </FormControl>

        <Grid container spacing={1.25}>
          <Grid size={{ xs: 6 }}>
            <TextField
              id="route-origin-lat"
              label="Origin latitude"
              value={form.originLat}
              onChange={(event) => updateForm("originLat", event.target.value)}
              fullWidth
              inputMode="decimal"
              error={Boolean(errors.originLat)}
              helperText={errors.originLat}
              FormHelperTextProps={{ id: "route-origin-lat-error" }}
              inputProps={{ "aria-describedby": "route-origin-lat-error" }}
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              id="route-origin-lng"
              label="Origin longitude"
              value={form.originLng}
              onChange={(event) => updateForm("originLng", event.target.value)}
              fullWidth
              inputMode="decimal"
              error={Boolean(errors.originLng)}
              helperText={errors.originLng}
              FormHelperTextProps={{ id: "route-origin-lng-error" }}
              inputProps={{ "aria-describedby": "route-origin-lng-error" }}
            />
          </Grid>
        </Grid>

        <Button
          type="button"
          variant="outlined"
          size="small"
          startIcon={<LocateFixed size={18} />}
          onClick={useOriginLocation}
          disabled={plannerState.status === "loading"}
        >
          Use my location
        </Button>

        {form.waypointType === "manual" ? (
          <Grid container spacing={1.25}>
            <Grid size={{ xs: 6 }}>
              <TextField
                id="route-dest-lat"
                label="Destination latitude"
                value={form.destLat}
                onChange={(event) => updateForm("destLat", event.target.value)}
                fullWidth
                inputMode="decimal"
                error={Boolean(errors.destLat)}
                helperText={errors.destLat}
                FormHelperTextProps={{ id: "route-dest-lat-error" }}
                inputProps={{ "aria-describedby": "route-dest-lat-error" }}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                id="route-dest-lng"
                label="Destination longitude"
                value={form.destLng}
                onChange={(event) => updateForm("destLng", event.target.value)}
                fullWidth
                inputMode="decimal"
                error={Boolean(errors.destLng)}
                helperText={errors.destLng}
                FormHelperTextProps={{ id: "route-dest-lng-error" }}
                inputProps={{ "aria-describedby": "route-dest-lng-error" }}
              />
            </Grid>
          </Grid>
        ) : null}

        {plannerState.status === "error" ? (
          <Alert severity="error" className="warning-alert">
            {plannerState.message}
          </Alert>
        ) : null}

        <Button
          type="submit"
          variant="contained"
          color="error"
          disabled={plannerState.status === "loading"}
          startIcon={
            plannerState.status === "loading" ? (
              <Loader2 size={18} className="spin-icon" />
            ) : (
              <MapPin size={18} />
            )
          }
        >
          {plannerState.status === "loading"
            ? plannerState.message
            : "Plan route"}
        </Button>
      </Stack>

      {result ? (
        <Stack spacing={1.5} sx={{ mt: 2 }}>
          <Alert
            severity={result.decisionSupport ? "warning" : "info"}
            icon={<Shield size={20} />}
            className="warning-alert"
          >
            <Typography variant="subtitle2">{result.disclaimer}</Typography>
          </Alert>

          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1}
            flexWrap="wrap"
          >
            <Chip
              icon={<MapPin size={16} />}
              label={formatDistance(result.distanceMeters)}
              color="primary"
            />
            <Chip
              label={`${result.estimatedDurationMinutes} min walking`}
              variant="outlined"
            />
            {result.targetShelter ? (
              <Chip
                label={`Shelter: ${result.targetShelter.name}`}
                color="success"
              />
            ) : null}
          </Stack>

          <Stack
            direction="row"
            spacing={1}
            flexWrap="wrap"
            sx={{ color: "text.secondary" }}
          >
            <Typography variant="caption">
              Avoided {result.avoidedClosures.length} closure
              {result.avoidedClosures.length === 1 ? "" : "s"} and{" "}
              {result.avoidedRiskZones.length} risk zone
              {result.avoidedRiskZones.length === 1 ? "" : "s"}
            </Typography>
          </Stack>

          {result.segments.length > 0 ? (
            <Paper variant="outlined" className="shelter-row">
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                Route steps
              </Typography>
              <Stack spacing={0.75}>
                {result.segments.map((segment, index) => (
                  <Stack
                    key={index}
                    direction="row"
                    spacing={1}
                    alignItems="center"
                  >
                    <Typography
                      variant="caption"
                      sx={{ minWidth: 24, fontWeight: 600 }}
                    >
                      {index + 1}.
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {segment.start.lat.toFixed(5)},{" "}
                      {segment.start.lng.toFixed(5)} →{" "}
                      {segment.end.lat.toFixed(5)}, {segment.end.lng.toFixed(5)}
                    </Typography>
                    <Typography variant="caption" sx={{ ml: "auto" }}>
                      {formatDistance(segment.distanceMeters)}
                    </Typography>
                  </Stack>
                ))}
              </Stack>
            </Paper>
          ) : null}
        </Stack>
      ) : null}

      {plannerState.status === "success" && plannerState.message ? (
        <FormHelperText sx={{ mt: 1 }}>{plannerState.message}</FormHelperText>
      ) : null}
    </Paper>
  );
}
