import {
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
import { Ambulance, Building2, ClipboardList, Users } from "lucide-react";

import type { IncidentRecord, IncidentStatus } from "@nadaa/shared-types";
import { statusLabel } from "../data";
import { allowedTransitions } from "../utils";
import type { IncidentFilterState, StatusFormState } from "../types";

import { HazardChip, MetricCard, SeverityChip } from "./shared";

export function IncidentFilters({
  filters,
  onChange,
}: {
  filters: IncidentFilterState;
  onChange: (filters: IncidentFilterState) => void;
}) {
  return (
    <Stack
      direction="row"
      sx={{
        flexWrap: "wrap",
        gap: 2
      }}>
      <FormControl size="small" sx={{ minWidth: 140 }}>
        <InputLabel>Hazard</InputLabel>
        <Select
          label="Hazard"
          onChange={(event: SelectChangeEvent) =>
            onChange({
              ...filters,
              hazard: event.target.value as typeof filters.hazard,
            })
          }
          value={filters.hazard}
        >
          <MenuItem value="all">All hazards</MenuItem>
          <MenuItem value="flood">Flood</MenuItem>
          <MenuItem value="fire">Fire</MenuItem>
          <MenuItem value="road_crash">Road crash</MenuItem>
          <MenuItem value="building_collapse">Building collapse</MenuItem>
          <MenuItem value="medical_emergency">Medical emergency</MenuItem>
          <MenuItem value="security_incident">Security incident</MenuItem>
          <MenuItem value="disease_outbreak">Disease outbreak</MenuItem>
          <MenuItem value="electrical_hazard">Electrical hazard</MenuItem>
          <MenuItem value="blocked_drain">Blocked drain</MenuItem>
          <MenuItem value="landslide">Landslide</MenuItem>
          <MenuItem value="marine_accident">Marine accident</MenuItem>
          <MenuItem value="storm">Storm</MenuItem>
          <MenuItem value="tidal_wave">Tidal wave</MenuItem>
          <MenuItem value="other">Other</MenuItem>
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 140 }}>
        <InputLabel>Severity</InputLabel>
        <Select
          label="Severity"
          onChange={(event: SelectChangeEvent) =>
            onChange({
              ...filters,
              severity: event.target.value as typeof filters.severity,
            })
          }
          value={filters.severity}
        >
          <MenuItem value="all">All severities</MenuItem>
          <MenuItem value="emergency">Emergency</MenuItem>
          <MenuItem value="severe">Severe</MenuItem>
          <MenuItem value="high">High</MenuItem>
          <MenuItem value="moderate">Moderate</MenuItem>
          <MenuItem value="low">Low</MenuItem>
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 140 }}>
        <InputLabel>Status</InputLabel>
        <Select
          label="Status"
          onChange={(event: SelectChangeEvent) =>
            onChange({
              ...filters,
              status: event.target.value as typeof filters.status,
            })
          }
          value={filters.status}
        >
          <MenuItem value="all">All statuses</MenuItem>
          <MenuItem value="reported">Reported</MenuItem>
          <MenuItem value="under_review">Under review</MenuItem>
          <MenuItem value="verified">Verified</MenuItem>
          <MenuItem value="assigned">Assigned</MenuItem>
          <MenuItem value="response_en_route">Response en route</MenuItem>
          <MenuItem value="on_scene">On scene</MenuItem>
          <MenuItem value="contained">Contained</MenuItem>
          <MenuItem value="recovery_ongoing">Recovery ongoing</MenuItem>
          <MenuItem value="closed">Closed</MenuItem>
          <MenuItem value="false_report">False report</MenuItem>
        </Select>
      </FormControl>
    </Stack>
  );
}

export function IncidentListItem({
  incident,
  onClick,
  selected,
}: {
  incident: IncidentRecord;
  onClick: () => void;
  selected?: boolean;
}) {
  return (
    <Paper
      elevation={selected ? 3 : 0}
      onClick={onClick}
      sx={{
        border: (theme) =>
          selected
            ? `2px solid ${theme.palette.primary.main}`
            : "1px solid var(--nadaa-divider)",
        borderRadius: 2,
        cursor: "pointer",
        p: 2,
        transition: "box-shadow 0.2s",
        "&:hover": { boxShadow: 2 },
      }}
    >
      <Stack direction="row" spacing={2} sx={{
        justifyContent: "space-between"
      }}>
        <Box>
          <Typography variant="subtitle1" sx={{
            fontWeight: 700
          }}>
            {incident.reference}
          </Typography>
          <Typography variant="body2" sx={{
            color: "text.secondary"
          }}>
            {incident.description}
          </Typography>
        </Box>
        <SeverityChip severity={incident.severity} />
      </Stack>
      <Stack
        direction="row"
        sx={{
          flexWrap: "wrap",
          gap: 1,
          mt: 1
        }}>
        <HazardChip hazard={incident.type} />
        <Chip
          label={statusLabel(incident.status)}
          size="small"
          variant="outlined"
        />
        {incident.priorityReview ? (
          <Chip color="error" label="Priority" size="small" />
        ) : null}
      </Stack>
    </Paper>
  );
}

export function IncidentDetail({ incident }: { incident: IncidentRecord }) {
  return (
    <Stack spacing={2}>
      <Stack direction="row" sx={{
        justifyContent: "space-between"
      }}>
        <Typography variant="h5" sx={{
          fontWeight: 800
        }}>
          {incident.reference}
        </Typography>
        <SeverityChip severity={incident.severity} size="medium" />
      </Stack>
      <Typography variant="body1">{incident.description}</Typography>
      <Stack
        direction="row"
        sx={{
          flexWrap: "wrap",
          gap: 1
        }}>
        <Chip label={statusLabel(incident.status)} />
        <HazardChip hazard={incident.type} size="medium" />
        {incident.priorityReview ? (
          <Chip color="error" label="Priority review" />
        ) : null}
        {incident.anonymous ? <Chip label="Anonymous report" /> : null}
      </Stack>
      <Divider />
      <Grid container spacing={2}>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricCard
            icon={Users}
            label="People affected"
            value={incident.peopleAffected}
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricCard
            icon={Ambulance}
            label="Injuries"
            value={incident.injuriesReported ? "Yes" : "No"}
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricCard
            icon={ClipboardList}
            label="Urgency"
            value={incident.urgency}
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricCard
            icon={Building2}
            label="Contact permission"
            value={incident.contactPermission ? "Yes" : "No"}
          />
        </Grid>
      </Grid>
      <Divider />
      <Typography variant="h6" sx={{
        fontWeight: 700
      }}>
        Location
      </Typography>
      <Typography variant="body2" sx={{
        color: "text.secondary"
      }}>
        {incident.location.lat.toFixed(5)}, {incident.location.lng.toFixed(5)}
      </Typography>
      {incident.assignments.length > 0 ? (
        <>
          <Divider />
          <Typography variant="h6" sx={{
            fontWeight: 700
          }}>
            Assignment
          </Typography>
          {incident.assignments.map((assignment) => (
            <Paper key={assignment.id} sx={{ p: 2 }} variant="outlined">
              <Typography sx={{
                fontWeight: 600
              }}>{assignment.agencyName}</Typography>
              <Typography variant="body2" sx={{
                color: "text.secondary"
              }}>
                Priority: {assignment.priority} · Lead:{" "}
                {assignment.responderLead ?? "Unassigned"}
              </Typography>
              <Typography variant="body2">{assignment.instructions}</Typography>
            </Paper>
          ))}
        </>
      ) : null}
      <Divider />
      <Typography variant="h6" sx={{
        fontWeight: 700
      }}>
        Timeline
      </Typography>
      {incident.timeline.length === 0 ? (
        <Typography sx={{
          color: "text.secondary"
        }}>No timeline events yet.</Typography>
      ) : (
        <Stack spacing={1}>
          {incident.timeline.map((event) => (
            <Paper key={event.id} sx={{ p: 2 }} variant="outlined">
              <Typography variant="body2">{event.message}</Typography>
              <Typography variant="caption" sx={{
                color: "text.secondary"
              }}>
                {event.actorRole ? `${event.actorRole} · ` : ""}
                {new Date(event.createdAt).toLocaleString("en-GH")}
              </Typography>
            </Paper>
          ))}
        </Stack>
      )}
    </Stack>
  );
}

export function StatusUpdateForm({
  currentStatus,
  form,
  onChange,
  onSubmit,
  submitLabel,
}: {
  currentStatus: IncidentStatus;
  form: StatusFormState;
  onChange: (form: StatusFormState) => void;
  onSubmit: () => void;
  submitLabel: string;
}) {
  const transitions = allowedTransitions(currentStatus);
  const requiresNotes =
    form.status === "closed" || form.status === "false_report";
  const disabled =
    form.status === currentStatus ||
    (requiresNotes && !form.resolutionNotes.trim());

  return (
    <Stack spacing={2}>
      <FormControl fullWidth size="small">
        <InputLabel>New status</InputLabel>
        <Select
          label="New status"
          onChange={(event: SelectChangeEvent) =>
            onChange({ ...form, status: event.target.value as IncidentStatus })
          }
          value={form.status}
        >
          {transitions.map((status) => (
            <MenuItem key={status} value={status}>
              {statusLabel(status)}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      <TextField
        label="Note"
        multiline
        onChange={(event) => onChange({ ...form, note: event.target.value })}
        placeholder="Reason for status change or timeline note"
        rows={2}
        size="small"
        value={form.note}
      />

      {requiresNotes ? (
        <TextField
          error={requiresNotes && !form.resolutionNotes.trim()}
          helperText={
            requiresNotes && !form.resolutionNotes.trim()
              ? "Resolution notes are required to close or mark as false report"
              : ""
          }
          label="Resolution notes (required)"
          multiline
          onChange={(event) =>
            onChange({ ...form, resolutionNotes: event.target.value })
          }
          placeholder="Explain how the incident was resolved"
          required
          rows={3}
          size="small"
          value={form.resolutionNotes}
        />
      ) : null}

      <Button disabled={disabled} onClick={onSubmit} variant="contained">
        {submitLabel}
      </Button>
    </Stack>
  );
}
