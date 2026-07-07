import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  FormControl,
  Grid,
  InputLabel,
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  Ambulance,
  Bed,
  Building2,
  ClipboardList,
  HandHeart,
  MapPin,
  Users,
} from "lucide-react";
import type {
  AidPledgeRecord,
  AidRequestRecord,
  HospitalCapacityRecord,
  IncidentRecord,
  IncidentStatus,
  ReliefPointRecord,
  ReliefStockHistoryEntry,
  ShelterRecord,
} from "@nadaa/shared-types";
import {
  aidRequestCategoryOptions,
  aidRequestPriorityOptions,
  hazardLabel,
  hospitalCapacityOptions,
  hospitalUnitStatusOptions,
  reliefPointStatusOptions,
  reliefPointTypeOptions,
  severityLabel,
  shelterStatusOptions,
  statusLabel,
} from "./data";
import {
  aidLabel,
  aidPriorityColor,
  aidProgressPercent,
  aidStatusColor,
  allowedTransitions,
  hospitalBedPercent,
  hospitalCapacityColor,
  reliefLabel,
  reliefStatusColor,
  severityColor,
  stockSummary,
} from "./utils";
import type {
  AidRequestFormState,
  HospitalCapacityFormState,
  IncidentFilterState,
  ReliefPointFormState,
  ShelterOccupancyFormState,
  StatusFormState,
} from "./types";

export function LoadingState({ message }: { message?: string }) {
  return (
    <Box
      alignItems="center"
      display="flex"
      justifyContent="center"
      minHeight={200}
    >
      <Stack alignItems="center" spacing={2}>
        <CircularProgress />
        <Typography color="text.secondary">{message ?? "Loading"}</Typography>
      </Stack>
    </Box>
  );
}

export function EmptyState({ message }: { message: string }) {
  return (
    <Alert severity="info" sx={{ mt: 2 }}>
      {message}
    </Alert>
  );
}

export function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <Alert
      action={
        onRetry ? (
          <Button color="inherit" onClick={onRetry} size="small">
            Retry
          </Button>
        ) : null
      }
      severity="error"
      sx={{ mt: 2 }}
    >
      {message}
    </Alert>
  );
}

export function MetricCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Users;
  label: string;
  value: number | string;
}) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack alignItems="center" direction="row" spacing={2}>
          <Box color="primary.main">
            <Icon size={24} />
          </Box>
          <Box>
            <Typography color="text.secondary" variant="caption">
              {label}
            </Typography>
            <Typography fontWeight={800} variant="h5">
              {value}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

export function IncidentFilters({
  filters,
  onChange,
}: {
  filters: IncidentFilterState;
  onChange: (filters: IncidentFilterState) => void;
}) {
  return (
    <Stack direction="row" flexWrap="wrap" gap={2}>
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
            : "1px solid #e5e7eb",
        borderRadius: 2,
        cursor: "pointer",
        p: 2,
        transition: "box-shadow 0.2s",
        "&:hover": { boxShadow: 2 },
      }}
    >
      <Stack direction="row" justifyContent="space-between" spacing={2}>
        <Box>
          <Typography fontWeight={700} variant="subtitle1">
            {incident.reference}
          </Typography>
          <Typography color="text.secondary" variant="body2">
            {incident.description}
          </Typography>
        </Box>
        <Chip
          color={severityColor(incident.severity)}
          label={severityLabel(incident.severity)}
          size="small"
        />
      </Stack>
      <Stack direction="row" flexWrap="wrap" gap={1} mt={1}>
        <Chip
          icon={<MapPin size={14} />}
          label={hazardLabel(incident.type)}
          size="small"
          variant="outlined"
        />
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
      <Stack direction="row" justifyContent="space-between">
        <Typography fontWeight={800} variant="h5">
          {incident.reference}
        </Typography>
        <Chip
          color={severityColor(incident.severity)}
          label={severityLabel(incident.severity)}
        />
      </Stack>

      <Typography variant="body1">{incident.description}</Typography>

      <Stack direction="row" flexWrap="wrap" gap={1}>
        <Chip label={statusLabel(incident.status)} />
        <Chip icon={<MapPin size={14} />} label={hazardLabel(incident.type)} />
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

      <Typography fontWeight={700} variant="h6">
        Location
      </Typography>
      <Typography color="text.secondary" variant="body2">
        {incident.location.lat.toFixed(5)}, {incident.location.lng.toFixed(5)}
      </Typography>

      {incident.assignments.length > 0 ? (
        <>
          <Divider />
          <Typography fontWeight={700} variant="h6">
            Assignment
          </Typography>
          {incident.assignments.map((assignment) => (
            <Paper key={assignment.id} sx={{ p: 2 }} variant="outlined">
              <Typography fontWeight={600}>{assignment.agencyName}</Typography>
              <Typography color="text.secondary" variant="body2">
                Priority: {assignment.priority} · Lead:{" "}
                {assignment.responderLead ?? "Unassigned"}
              </Typography>
              <Typography variant="body2">{assignment.instructions}</Typography>
            </Paper>
          ))}
        </>
      ) : null}

      <Divider />

      <Typography fontWeight={700} variant="h6">
        Timeline
      </Typography>
      {incident.timeline.length === 0 ? (
        <Typography color="text.secondary">No timeline events yet.</Typography>
      ) : (
        <Stack spacing={1}>
          {incident.timeline.map((event) => (
            <Paper key={event.id} sx={{ p: 2 }} variant="outlined">
              <Typography variant="body2">{event.message}</Typography>
              <Typography color="text.secondary" variant="caption">
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
          label="Resolution notes (required)"
          multiline
          onChange={(event) =>
            onChange({ ...form, resolutionNotes: event.target.value })
          }
          placeholder="Explain how the incident was resolved"
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

export function ShelterCard({ shelter }: { shelter: ShelterRecord }) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack direction="row" justifyContent="space-between">
          <Typography fontWeight={700}>{shelter.name}</Typography>
          <Chip label={shelter.status} size="small" />
        </Stack>
        <Typography color="text.secondary" variant="body2">
          {shelter.address}
        </Typography>
        <Stack direction="row" flexWrap="wrap" gap={1} mt={1}>
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
}: {
  form: ShelterOccupancyFormState;
  onChange: (form: ShelterOccupancyFormState) => void;
  onSubmit: () => void;
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
      <Button onClick={onSubmit} variant="contained">
        Update occupancy
      </Button>
    </Stack>
  );
}

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
        <Stack direction="row" justifyContent="space-between" spacing={2}>
          <Box>
            <Typography fontWeight={700}>{point.name}</Typography>
            <Typography color="text.secondary" variant="body2">
              {reliefLabel(point.type)} · {point.district}
            </Typography>
          </Box>
          <Chip
            color={reliefStatusColor(point.status)}
            label={reliefLabel(point.status)}
            size="small"
          />
        </Stack>
        <Typography color="text.secondary" mt={1} variant="body2">
          {point.address}
        </Typography>
        <Typography mt={1} variant="body2">
          {stockSummary(point.stockCategories)}
        </Typography>
        <Stack direction="row" flexWrap="wrap" gap={1} mt={1}>
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
        label="Name"
        onChange={(event) => onChange({ ...form, name: event.target.value })}
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
          fullWidth
          label="Latitude"
          onChange={(event) => onChange({ ...form, lat: event.target.value })}
          size="small"
          type="number"
          value={form.lat}
        />
        <TextField
          fullWidth
          label="Longitude"
          onChange={(event) => onChange({ ...form, lng: event.target.value })}
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
          <Typography fontWeight={700}>
            {new Date(entry.changedAt).toLocaleString("en-GH")}
          </Typography>
          <Typography color="text.secondary" variant="body2">
            {entry.changedBy}
          </Typography>
          <Typography mt={1} variant="body2">
            {stockSummary(entry.stockCategories)}
          </Typography>
        </Paper>
      ))}
    </Stack>
  );
}

export function AidRequestCard({
  onSelect,
  request,
  selected,
}: {
  onSelect: () => void;
  request: AidRequestRecord;
  selected?: boolean;
}) {
  const progress = aidProgressPercent(request);

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
        <Stack direction="row" justifyContent="space-between" spacing={2}>
          <Box>
            <Typography fontWeight={700}>{request.title}</Typography>
            <Typography color="text.secondary" variant="body2">
              {aidLabel(request.category)} · {request.district}
            </Typography>
          </Box>
          <Stack alignItems="flex-end" spacing={1}>
            <Chip
              color={aidStatusColor(request.status)}
              label={aidLabel(request.status)}
              size="small"
            />
            <Chip
              color={aidPriorityColor(request.priority)}
              label={aidLabel(request.priority)}
              size="small"
              variant="outlined"
            />
          </Stack>
        </Stack>

        <Typography mt={1} variant="body2">
          {request.quantityPledged.toLocaleString("en-GH")} /{" "}
          {request.quantityNeeded.toLocaleString("en-GH")}{" "}
          {request.quantityUnit} pledged
        </Typography>
        <LinearProgress sx={{ mt: 1 }} value={progress} variant="determinate" />
        <Stack direction="row" flexWrap="wrap" gap={1} mt={1.5}>
          <Chip
            icon={<HandHeart size={14} />}
            label={`${request.pledges.length} pledges`}
            size="small"
            variant="outlined"
          />
          <Chip
            label={`Needed by ${new Date(request.neededBy).toLocaleDateString("en-GH")}`}
            size="small"
            variant="outlined"
          />
        </Stack>
      </CardContent>
    </Card>
  );
}

export function AidRequestForm({
  form,
  onChange,
  onSubmit,
  submitLabel,
}: {
  form: AidRequestFormState;
  onChange: (form: AidRequestFormState) => void;
  onSubmit: () => void;
  submitLabel: string;
}) {
  return (
    <Stack spacing={2}>
      <TextField
        label="Title"
        onChange={(event) => onChange({ ...form, title: event.target.value })}
        size="small"
        value={form.title}
      />
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <FormControl fullWidth size="small">
          <InputLabel>Category</InputLabel>
          <Select
            label="Category"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                category: event.target.value as typeof form.category,
              })
            }
            value={form.category}
          >
            {aidRequestCategoryOptions.map((category) => (
              <MenuItem key={category} value={category}>
                {aidLabel(category)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl fullWidth size="small">
          <InputLabel>Priority</InputLabel>
          <Select
            label="Priority"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                priority: event.target.value as typeof form.priority,
              })
            }
            value={form.priority}
          >
            {aidRequestPriorityOptions.map((priority) => (
              <MenuItem key={priority} value={priority}>
                {aidLabel(priority)}
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
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Latitude"
          onChange={(event) => onChange({ ...form, lat: event.target.value })}
          size="small"
          type="number"
          value={form.lat}
        />
        <TextField
          fullWidth
          label="Longitude"
          onChange={(event) => onChange({ ...form, lng: event.target.value })}
          size="small"
          type="number"
          value={form.lng}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Receiving organization"
          onChange={(event) =>
            onChange({ ...form, receivingOrganization: event.target.value })
          }
          size="small"
          value={form.receivingOrganization}
        />
        <TextField
          fullWidth
          label="Contact"
          onChange={(event) =>
            onChange({ ...form, contact: event.target.value })
          }
          size="small"
          value={form.contact}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Quantity needed"
          onChange={(event) =>
            onChange({ ...form, quantityNeeded: event.target.value })
          }
          size="small"
          type="number"
          value={form.quantityNeeded}
        />
        <TextField
          fullWidth
          label="Unit"
          onChange={(event) =>
            onChange({ ...form, quantityUnit: event.target.value })
          }
          size="small"
          value={form.quantityUnit}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          InputLabelProps={{ shrink: true }}
          label="Needed by"
          onChange={(event) =>
            onChange({ ...form, neededBy: event.target.value })
          }
          size="small"
          type="datetime-local"
          value={form.neededBy}
        />
        <FormControl fullWidth size="small">
          <InputLabel>Visibility</InputLabel>
          <Select
            label="Visibility"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                visibility: event.target.value as typeof form.visibility,
              })
            }
            value={form.visibility}
          >
            <MenuItem value="public">Public</MenuItem>
            <MenuItem value="partners_only">Partners only</MenuItem>
          </Select>
        </FormControl>
      </Stack>
      <TextField
        label="Source relief point ID"
        onChange={(event) =>
          onChange({ ...form, sourceReliefPointId: event.target.value })
        }
        size="small"
        value={form.sourceReliefPointId}
      />
      <TextField
        label="Description"
        multiline
        onChange={(event) =>
          onChange({ ...form, description: event.target.value })
        }
        rows={3}
        size="small"
        value={form.description}
      />
      <Button
        disabled={
          !form.title.trim() ||
          !form.receivingOrganization.trim() ||
          !form.quantityNeeded.trim() ||
          !form.description.trim()
        }
        onClick={onSubmit}
        variant="contained"
      >
        {submitLabel}
      </Button>
    </Stack>
  );
}

export function AidPledgeList({ pledges }: { pledges: AidPledgeRecord[] }) {
  if (!pledges.length) {
    return <EmptyState message="No partner pledges recorded yet." />;
  }

  return (
    <Stack spacing={1.25}>
      {pledges.map((pledge) => (
        <Paper key={pledge.id} sx={{ p: 2 }} variant="outlined">
          <Stack direction="row" justifyContent="space-between" spacing={2}>
            <Box>
              <Typography fontWeight={700}>{pledge.donorName}</Typography>
              <Typography color="text.secondary" variant="body2">
                {aidLabel(pledge.donorType)} ·{" "}
                {new Date(pledge.pledgedAt).toLocaleString("en-GH")}
              </Typography>
            </Box>
            <Stack alignItems="flex-end" spacing={1}>
              <Chip label={aidLabel(pledge.status)} size="small" />
              <Chip
                color={
                  pledge.reviewStatus === "flagged"
                    ? "error"
                    : pledge.reviewStatus === "cleared"
                      ? "success"
                      : "warning"
                }
                label={aidLabel(pledge.reviewStatus)}
                size="small"
                variant="outlined"
              />
            </Stack>
          </Stack>
          <Typography mt={1} variant="body2">
            {pledge.quantity.toLocaleString("en-GH")} {pledge.unit}
          </Typography>
          {pledge.note ? (
            <Typography color="text.secondary" mt={0.5} variant="body2">
              {pledge.note}
            </Typography>
          ) : null}
          {pledge.fraudReviewNotes ? (
            <Alert severity="info" sx={{ mt: 1 }}>
              {pledge.fraudReviewNotes}
            </Alert>
          ) : null}
        </Paper>
      ))}
    </Stack>
  );
}
