import {
  Button,
  Card,
  CardContent,
  Chip,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import { Bed } from "lucide-react";

import type { ShelterRecord } from "@nadaa/shared-types";
import { shelterStatusOptions } from "../data";

import type { ShelterOccupancyFormState } from "../types";

export function ShelterCard({ shelter }: { shelter: ShelterRecord }) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack direction="row" sx={{
          justifyContent: "space-between"
        }}>
          <Typography sx={{
            fontWeight: 700
          }}>{shelter.name}</Typography>
          <Chip label={shelter.status} size="small" />
        </Stack>
        <Typography variant="body2" sx={{
          color: "text.secondary"
        }}>
          {shelter.address}
        </Typography>
        <Stack
          direction="row"
          sx={{
            flexWrap: "wrap",
            gap: 1,
            mt: 1
          }}>
          <Chip
            icon={<Bed size={14} />}
            label={`${shelter.currentOccupancy} / ${shelter.capacity}`}
            size="small"
          />
          {shelter.facilities.map((facility) => (
            <Chip
              key={facility}
              label={facility}
              size="small"
              variant="outlined"
            />
          ))}
        </Stack>
      </CardContent>
    </Card>
  );
}

export function ShelterOccupancyForm({
  form,
  onChange,
  onSubmit,
  submitting = false,
}: {
  form: ShelterOccupancyFormState;
  onChange: (form: ShelterOccupancyFormState) => void;
  onSubmit: () => void;
  submitting?: boolean;
}) {
  return (
    <Stack spacing={2}>
      <TextField
        label="Capacity"
        onChange={(event) =>
          onChange({ ...form, capacity: event.target.value })
        }
        size="small"
        type="number"
        value={form.capacity}
      />
      <TextField
        label="Current occupancy"
        onChange={(event) =>
          onChange({ ...form, currentOccupancy: event.target.value })
        }
        size="small"
        type="number"
        value={form.currentOccupancy}
      />
      <FormControl fullWidth size="small">
        <InputLabel>Status</InputLabel>
        <Select
          label="Status"
          onChange={(event: SelectChangeEvent) =>
            onChange({ ...form, status: event.target.value })
          }
          value={form.status}
        >
          {shelterStatusOptions.map((status) => (
            <MenuItem key={status} value={status}>
              {status}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        label="Notes"
        multiline
        onChange={(event) => onChange({ ...form, notes: event.target.value })}
        rows={2}
        size="small"
        value={form.notes}
      />
      <Button disabled={submitting} onClick={onSubmit} variant="contained">
        {submitting ? "Updating..." : "Update occupancy"}
      </Button>
    </Stack>
  );
}
