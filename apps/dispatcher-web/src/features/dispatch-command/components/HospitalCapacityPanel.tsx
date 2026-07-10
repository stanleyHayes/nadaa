import { type ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  FormControlLabel,
  Grid,
  LinearProgress,
  MenuItem,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";

import "leaflet/dist/leaflet.css";
import { Hospital } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { HospitalCapacityRecord } from "@nadaa/shared-types";

import type { CapacityLoadState, HospitalCapacityFilterState } from "../types";
import {
  hospitalBedPercent,
  hospitalCapacityColor,
  hospitalCapacityLabel,
  hospitalUnitStatusLabel,
  hospitalUpdatedLabel,
  metersLabel,
} from "../utils";

import { CommandSelect, EmptyState, Fact } from "./shared";

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
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          gap: 1.5
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <Hospital size={22} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h5">Hospital capacity</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Beds, emergency unit status, ambulances, oxygen, and stale update
              warnings for dispatcher routing.
            </Typography>
          </Box>
        </Stack>
        <Stack
          direction="row"
          spacing={1}
          sx={{
            alignItems: "center",
            flexWrap: "wrap"
          }}>
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live capacity"
                : loadState === "loading"
                  ? "Loading"
                  : loadState === "empty"
                    ? "No matches"
                    : "Unavailable"
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
                <Stack
                  direction="row"
                  sx={{
                    justifyContent: "space-between",
                    gap: 1
                  }}>
                  <Box>
                    <Typography variant="subtitle2">{facility.name}</Typography>
                    <Typography variant="caption" sx={{
                      color: "text.secondary"
                    }}>
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
                    sx={{
                      justifyContent: "space-between",
                      alignItems: "center"
                    }}>
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

                <Typography variant="caption" sx={{
                  color: "text.secondary"
                }}>
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
