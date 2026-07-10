import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";

import type {
  ReliefPointRecord,
  ReliefStockHistoryEntry,
} from "@nadaa/shared-types";
import { reliefPointStatusOptions, reliefPointTypeOptions } from "../data";
import { reliefLabel, reliefStatusColor, stockSummary } from "../utils";
import type { ReliefPointFormState } from "../types";

import { EmptyState } from "./shared";

export function ReliefPointCard({
  onSelect,
  point,
  selected,
}: {
  onSelect: () => void;
  point: ReliefPointRecord;
  selected?: boolean;
}) {
  return (
    <Card
      onClick={onSelect}
      sx={{
        borderColor: selected ? "primary.main" : undefined,
        cursor: "pointer",
      }}
      variant="outlined"
    >
      <CardContent>
        <Stack direction="row" spacing={2} sx={{
          justifyContent: "space-between"
        }}>
          <Box>
            <Typography sx={{
              fontWeight: 700
            }}>{point.name}</Typography>
            <Typography variant="body2" sx={{
              color: "text.secondary"
            }}>
              {reliefLabel(point.type)} · {point.district}
            </Typography>
          </Box>
          <Chip
            color={reliefStatusColor(point.status)}
            label={reliefLabel(point.status)}
            size="small"
          />
        </Stack>
        <Typography
          variant="body2"
          sx={{
            color: "text.secondary",
            mt: 1
          }}>
          {point.address}
        </Typography>
        <Typography variant="body2" sx={{
          mt: 1
        }}>
          {stockSummary(point.stockCategories)}
        </Typography>
        <Stack
          direction="row"
          sx={{
            flexWrap: "wrap",
            gap: 1,
            mt: 1
          }}>
          <Chip label={point.operatingHours} size="small" variant="outlined" />
          <Chip label={point.schedule} size="small" variant="outlined" />
        </Stack>
        {point.eligibility ? (
          <Alert severity="info" sx={{ mt: 1 }}>
            {point.eligibility}
          </Alert>
        ) : null}
      </CardContent>
    </Card>
  );
}

export function ReliefPointForm({
  form,
  onChange,
  onSubmit,
  submitLabel,
}: {
  form: ReliefPointFormState;
  onChange: (form: ReliefPointFormState) => void;
  onSubmit: () => void;
  submitLabel: string;
}) {
  return (
    <Stack spacing={2}>
      <TextField
        error={!form.name.trim()}
        helperText={!form.name.trim() ? "Name is required" : ""}
        label="Name"
        onChange={(event) => onChange({ ...form, name: event.target.value })}
        required
        size="small"
        value={form.name}
      />
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <FormControl fullWidth size="small">
          <InputLabel>Type</InputLabel>
          <Select
            label="Type"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                type: event.target.value as typeof form.type,
              })
            }
            value={form.type}
          >
            {reliefPointTypeOptions.map((type) => (
              <MenuItem key={type} value={type}>
                {reliefLabel(type)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl fullWidth size="small">
          <InputLabel>Status</InputLabel>
          <Select
            label="Status"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                status: event.target.value as typeof form.status,
              })
            }
            value={form.status}
          >
            {reliefPointStatusOptions.map((status) => (
              <MenuItem key={status} value={status}>
                {reliefLabel(status)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Region"
          onChange={(event) =>
            onChange({ ...form, region: event.target.value })
          }
          size="small"
          value={form.region}
        />
        <TextField
          fullWidth
          label="District"
          onChange={(event) =>
            onChange({ ...form, district: event.target.value })
          }
          size="small"
          value={form.district}
        />
      </Stack>
      <TextField
        label="Address"
        onChange={(event) => onChange({ ...form, address: event.target.value })}
        size="small"
        value={form.address}
      />
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          error={!form.lat.trim()}
          fullWidth
          helperText={!form.lat.trim() ? "Latitude is required" : ""}
          label="Latitude"
          onChange={(event) => onChange({ ...form, lat: event.target.value })}
          required
          size="small"
          type="number"
          value={form.lat}
        />
        <TextField
          error={!form.lng.trim()}
          fullWidth
          helperText={!form.lng.trim() ? "Longitude is required" : ""}
          label="Longitude"
          onChange={(event) => onChange({ ...form, lng: event.target.value })}
          required
          size="small"
          type="number"
          value={form.lng}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Contact"
          onChange={(event) =>
            onChange({ ...form, contact: event.target.value })
          }
          size="small"
          value={form.contact}
        />
        <TextField
          fullWidth
          label="Operating hours"
          onChange={(event) =>
            onChange({ ...form, operatingHours: event.target.value })
          }
          size="small"
          value={form.operatingHours}
        />
      </Stack>
      <TextField
        label="Schedule"
        onChange={(event) =>
          onChange({ ...form, schedule: event.target.value })
        }
        size="small"
        value={form.schedule}
      />
      <TextField
        label="Eligibility"
        multiline
        onChange={(event) =>
          onChange({ ...form, eligibility: event.target.value })
        }
        rows={2}
        size="small"
        value={form.eligibility}
      />
      <TextField
        helperText="Use category:quantity:unit, separated by commas"
        label="Stock categories"
        multiline
        onChange={(event) =>
          onChange({ ...form, stockCategories: event.target.value })
        }
        rows={2}
        size="small"
        value={form.stockCategories}
      />
      <Button
        disabled={!form.name.trim() || !form.lat.trim() || !form.lng.trim()}
        onClick={onSubmit}
        variant="contained"
      >
        {submitLabel}
      </Button>
    </Stack>
  );
}

export function ReliefStockHistoryList({
  history,
}: {
  history: ReliefStockHistoryEntry[];
}) {
  if (!history.length) {
    return <EmptyState message="No stock history recorded yet." />;
  }

  return (
    <Stack spacing={1.25}>
      {history.map((entry) => (
        <Paper key={entry.id} sx={{ p: 2 }} variant="outlined">
          <Typography sx={{
            fontWeight: 700
          }}>
            {new Date(entry.changedAt).toLocaleString("en-GH")}
          </Typography>
          <Typography variant="body2" sx={{
            color: "text.secondary"
          }}>
            {entry.changedBy}
          </Typography>
          <Typography variant="body2" sx={{
            mt: 1
          }}>
            {stockSummary(entry.stockCategories)}
          </Typography>
        </Paper>
      ))}
    </Stack>
  );
}
