import {
  Alert,
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

import type { HospitalCapacityRecord } from "@nadaa/shared-types";
import { hospitalCapacityOptions } from "../data";
import { hospitalCapacityColor } from "../utils";
import type { HospitalCapacityFormState } from "../types";

export function HospitalCapacityCard({
  facility,
}: {
  facility: HospitalCapacityRecord;
}) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack direction="row" justifyContent="space-between">
          <Typography fontWeight={700}>{facility.name}</Typography>
          <Chip
            color={hospitalCapacityColor(facility.emergencyCapacity)}
            label={facility.emergencyCapacity}
            size="small"
          />
        </Stack>
        <Typography color="text.secondary" variant="body2">
          {facility.address}
        </Typography>
        <Stack direction="row" flexWrap="wrap" gap={1} mt={1}>
          <Chip
            icon={<Bed size={14} />}
            label={`Beds ${facility.availableBeds}/${facility.totalBeds}`}
            size="small"
          />
          <Chip
            label={`ICU ${facility.icuBedsAvailable}`}
            size="small"
            variant="outlined"
          />
          <Chip
            label={`Ambulance ${facility.ambulancesAvailable}`}
            size="small"
            variant="outlined"
          />
        </Stack>
        {facility.stale ? (
          <Alert severity="warning" sx={{ mt: 1 }}>
            Stale data · confirm before transfer
          </Alert>
        ) : null}
      </CardContent>
    </Card>
  );
}

export function HospitalCapacityUpdateForm({
  form,
  onChange,
  onSubmit,
}: {
  form: HospitalCapacityFormState;
  onChange: (form: HospitalCapacityFormState) => void;
  onSubmit: () => void;
}) {
  return (
    <Stack spacing={2}>
      <TextField
        label="Total beds"
        onChange={(event) =>
          onChange({ ...form, totalBeds: event.target.value })
        }
        size="small"
        type="number"
        value={form.totalBeds}
      />
      <TextField
        label="Available beds"
        onChange={(event) =>
          onChange({ ...form, availableBeds: event.target.value })
        }
        size="small"
        type="number"
        value={form.availableBeds}
      />
      <TextField
        label="ICU beds available"
        onChange={(event) =>
          onChange({ ...form, icuBedsAvailable: event.target.value })
        }
        size="small"
        type="number"
        value={form.icuBedsAvailable}
      />
      <FormControl fullWidth size="small">
        <InputLabel>Emergency capacity</InputLabel>
        <Select
          label="Emergency capacity"
          onChange={(event: SelectChangeEvent) =>
            onChange({
              ...form,
              emergencyCapacity: event.target
                .value as typeof form.emergencyCapacity,
            })
          }
          value={form.emergencyCapacity}
        >
          {hospitalCapacityOptions.map((option) => (
            <MenuItem key={option} value={option}>
              {option}
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
      <Button onClick={onSubmit} variant="contained">
        Update capacity
      </Button>
    </Stack>
  );
}
