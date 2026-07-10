import type { ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import { LifeBuoy, Loader2, RefreshCw } from "lucide-react";
import type { ShelterRecord } from "@nadaa/shared-types";
import type { IncidentLoadState, ShelterFormState } from "../types";
import { CommandSelect } from "./shared";
import { SectionCard } from "./primitives";

type FieldChange = (
  key: keyof ShelterFormState,
) => (
  event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
) => void;

export function ShelterCapacityPanel({
  shelters,
  shelterForm,
  selectedShelter,
  loadState,
  feedback,
  busy,
  onUpdateForm,
  onRefresh,
  onUpdateCapacity,
}: {
  shelters: ShelterRecord[];
  shelterForm: ShelterFormState;
  selectedShelter?: ShelterRecord;
  loadState: IncidentLoadState;
  feedback: string;
  busy: boolean;
  onUpdateForm: FieldChange;
  onRefresh: () => void;
  onUpdateCapacity: () => void;
}) {
  return (
    <SectionCard
      title="Shelter capacity"
      eyebrow="Occupancy & operating status"
      icon={LifeBuoy}
      accent="green"
      action={
        <Button
          type="button"
          variant="outlined"
          size="small"
          startIcon={
            loadState === "loading" ? (
              <Loader2 size={16} className="spin-icon" />
            ) : (
              <RefreshCw size={16} />
            )
          }
          onClick={onRefresh}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      }
    >
      {feedback ? (
        <Alert
          severity={loadState === "ready" ? "success" : "warning"}
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <CommandSelect
          label="Shelter"
          value={shelterForm.shelterId}
          onChange={onUpdateForm("shelterId")}
        >
          {shelters.map((shelter) => (
            <MenuItem value={shelter.id} key={shelter.id}>
              {shelter.name}
            </MenuItem>
          ))}
        </CommandSelect>

        {selectedShelter ? (
          <Box className="shelter-capacity-summary">
            <Stack direction="row" spacing={1} flexWrap="wrap">
              <Chip
                size="small"
                label={selectedShelter.status}
                color={
                  selectedShelter.status === "open" ? "success" : "warning"
                }
              />
              <Chip
                size="small"
                variant="outlined"
                label={`${selectedShelter.currentOccupancy}/${selectedShelter.capacity} occupied`}
              />
            </Stack>
            <Typography variant="caption" color="text.secondary">
              {selectedShelter.district} · {selectedShelter.address}
            </Typography>
          </Box>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={6}>
            <TextField
              label="Capacity"
              size="small"
              fullWidth
              required
              value={shelterForm.capacity}
              onChange={onUpdateForm("capacity")}
              inputProps={{ inputMode: "numeric" }}
              error={
                Boolean(shelterForm.capacity) &&
                !Number.isFinite(Number(shelterForm.capacity))
              }
              helperText={
                Boolean(shelterForm.capacity) &&
                !Number.isFinite(Number(shelterForm.capacity))
                  ? "Capacity must be a number"
                  : ""
              }
            />
          </Grid>
          <Grid size={6}>
            <TextField
              label="Occupancy"
              size="small"
              fullWidth
              required
              value={shelterForm.currentOccupancy}
              onChange={onUpdateForm("currentOccupancy")}
              inputProps={{ inputMode: "numeric" }}
              error={
                Boolean(shelterForm.currentOccupancy) &&
                !Number.isFinite(Number(shelterForm.currentOccupancy))
              }
              helperText={
                Boolean(shelterForm.currentOccupancy) &&
                !Number.isFinite(Number(shelterForm.currentOccupancy))
                  ? "Occupancy must be a number"
                  : ""
              }
            />
          </Grid>
          <Grid size={12}>
            <CommandSelect
              label="Status"
              value={shelterForm.status}
              onChange={onUpdateForm("status")}
            >
              <MenuItem value="open">Open</MenuItem>
              <MenuItem value="full">Full</MenuItem>
              <MenuItem value="closed">Closed</MenuItem>
              <MenuItem value="unknown">Unknown</MenuItem>
            </CommandSelect>
          </Grid>
          <Grid size={12}>
            <TextField
              label="Operational note"
              size="small"
              fullWidth
              multiline
              minRows={2}
              value={shelterForm.notes}
              onChange={onUpdateForm("notes")}
            />
          </Grid>
        </Grid>

        <Button
          type="button"
          variant="contained"
          startIcon={<LifeBuoy size={17} />}
          onClick={onUpdateCapacity}
          disabled={busy}
        >
          {busy ? "Updating" : "Update capacity"}
        </Button>
      </Stack>
    </SectionCard>
  );
}
