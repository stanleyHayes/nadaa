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
import type {
  ReliefPointRecord,
  ReliefStockHistoryEntry,
} from "@nadaa/shared-types";
import type { IncidentLoadState, ReliefPointFormState } from "../types";
import { formatShortDate } from "../utils";
import { CommandSelect } from "./shared";
import { SectionCard } from "./primitives";

type FieldChange = (
  key: keyof ReliefPointFormState,
) => (
  event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
) => void;

export function ReliefDistributionPanel({
  reliefPoints,
  reliefForm,
  selectedReliefPoint,
  reliefHistory,
  loadState,
  feedback,
  busy,
  onUpdateForm,
  onRefresh,
  onSave,
}: {
  reliefPoints: ReliefPointRecord[];
  reliefForm: ReliefPointFormState;
  selectedReliefPoint?: ReliefPointRecord;
  reliefHistory: ReliefStockHistoryEntry[];
  loadState: IncidentLoadState;
  feedback: string;
  busy: boolean;
  onUpdateForm: FieldChange;
  onRefresh: () => void;
  onSave: () => void;
}) {
  return (
    <SectionCard
      title="Relief distribution"
      eyebrow="Points, stock & eligibility"
      icon={LifeBuoy}
      accent="gold"
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
          label="Relief point"
          value={reliefForm.reliefPointId}
          onChange={onUpdateForm("reliefPointId")}
        >
          <MenuItem value="__new__">New relief point</MenuItem>
          {reliefPoints.map((point) => (
            <MenuItem value={point.id} key={point.id}>
              {point.name}
            </MenuItem>
          ))}
        </CommandSelect>

        {selectedReliefPoint && reliefForm.reliefPointId !== "__new__" ? (
          <Box className="shelter-capacity-summary">
            <Stack direction="row" spacing={1} flexWrap="wrap">
              <Chip
                size="small"
                label={selectedReliefPoint.status}
                color={
                  selectedReliefPoint.status === "open" ? "success" : "warning"
                }
              />
              <Chip
                size="small"
                variant="outlined"
                label={`${selectedReliefPoint.stockCategories.length} stock lines`}
              />
            </Stack>
            <Typography variant="caption" color="text.secondary">
              {selectedReliefPoint.district} · {selectedReliefPoint.address}
            </Typography>
          </Box>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 7 }}>
            <TextField
              label="Name"
              size="small"
              fullWidth
              required
              value={reliefForm.name}
              onChange={onUpdateForm("name")}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 5 }}>
            <CommandSelect
              label="Type"
              value={reliefForm.type}
              onChange={onUpdateForm("type")}
            >
              <MenuItem value="food">Food</MenuItem>
              <MenuItem value="water">Water</MenuItem>
              <MenuItem value="medical">Medical</MenuItem>
              <MenuItem value="hygiene">Hygiene</MenuItem>
              <MenuItem value="blankets">Blankets</MenuItem>
              <MenuItem value="cash">Cash</MenuItem>
              <MenuItem value="mixed">Mixed</MenuItem>
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 6, sm: 4 }}>
            <CommandSelect
              label="Status"
              value={reliefForm.status}
              onChange={onUpdateForm("status")}
            >
              <MenuItem value="open">Open</MenuItem>
              <MenuItem value="limited">Limited</MenuItem>
              <MenuItem value="paused">Paused</MenuItem>
              <MenuItem value="closed">Closed</MenuItem>
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 8 }}>
            <TextField
              label="Address"
              size="small"
              fullWidth
              value={reliefForm.address}
              onChange={onUpdateForm("address")}
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="Latitude"
              size="small"
              fullWidth
              required
              value={reliefForm.latitude}
              onChange={onUpdateForm("latitude")}
              inputProps={{ inputMode: "decimal" }}
              error={
                Boolean(reliefForm.latitude) &&
                !Number.isFinite(Number(reliefForm.latitude))
              }
              helperText={
                Boolean(reliefForm.latitude) &&
                !Number.isFinite(Number(reliefForm.latitude))
                  ? "Latitude must be a number"
                  : ""
              }
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="Longitude"
              size="small"
              fullWidth
              required
              value={reliefForm.longitude}
              onChange={onUpdateForm("longitude")}
              inputProps={{ inputMode: "decimal" }}
              error={
                Boolean(reliefForm.longitude) &&
                !Number.isFinite(Number(reliefForm.longitude))
              }
              helperText={
                Boolean(reliefForm.longitude) &&
                !Number.isFinite(Number(reliefForm.longitude))
                  ? "Longitude must be a number"
                  : ""
              }
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="District"
              size="small"
              fullWidth
              value={reliefForm.district}
              onChange={onUpdateForm("district")}
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="Contact"
              size="small"
              fullWidth
              value={reliefForm.contact}
              onChange={onUpdateForm("contact")}
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="Hours"
              size="small"
              fullWidth
              value={reliefForm.operatingHours}
              onChange={onUpdateForm("operatingHours")}
            />
          </Grid>
          <Grid size={{ xs: 6 }}>
            <TextField
              label="Schedule"
              size="small"
              fullWidth
              value={reliefForm.schedule}
              onChange={onUpdateForm("schedule")}
            />
          </Grid>
          <Grid size={12}>
            <TextField
              label="Eligibility"
              size="small"
              fullWidth
              multiline
              minRows={2}
              value={reliefForm.eligibility}
              onChange={onUpdateForm("eligibility")}
            />
          </Grid>
          <Grid size={12}>
            <TextField
              label="Stock lines"
              size="small"
              fullWidth
              multiline
              minRows={3}
              value={reliefForm.stockCategories}
              onChange={onUpdateForm("stockCategories")}
            />
          </Grid>
        </Grid>

        {reliefHistory.length ? (
          <Box className="relief-history">
            <Typography variant="caption" color="text.secondary">
              Recent stock history
            </Typography>
            {reliefHistory.slice(0, 2).map((entry) => (
              <Typography variant="caption" color="text.secondary" key={entry.id}>
                {formatShortDate(entry.changedAt)} ·{" "}
                {entry.stockCategories.length} stock lines · {entry.changedBy}
              </Typography>
            ))}
          </Box>
        ) : null}

        <Button
          type="button"
          variant="contained"
          startIcon={<LifeBuoy size={17} />}
          onClick={onSave}
          disabled={busy}
        >
          {busy
            ? "Saving"
            : reliefForm.reliefPointId === "__new__"
              ? "Publish relief point"
              : "Update relief point"}
        </Button>
      </Stack>
    </SectionCard>
  );
}
