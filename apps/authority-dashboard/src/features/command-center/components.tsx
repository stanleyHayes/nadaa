import { type ChangeEvent, type ReactNode, useEffect, useRef } from "react";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  Divider,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
  BellRing,
  CheckCheck,
  Crosshair,
  GitMerge,
  ShieldAlert,
  Truck,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AlertTarget,
  AuthorityAlertRecord,
  DuplicateReviewCandidate,
} from "@nadaa/shared-types";
import {
  alertSeverityOptions,
  alertTargetTypeOptions,
  assignmentAgencyOptions,
  incidentTransitionOptions,
  severityColors,
} from "./data";
import type {
  AbuseReviewFormState,
  AlertFormState,
  AlertLoadState,
  AssignmentFormState,
  CommandIncident,
  IncidentStatusFormState,
} from "./types";
import {
  abuseDecisionLabel,
  abuseScoreLabel,
  alertSeverityLabel,
  alertStatusColor,
  alertStatusLabel,
  alertTargetSummary,
  alertTargetTypeLabel,
  alertTargetWarnings,
  buildAlertTarget,
  canAssignIncident,
  formatShortTime,
  hazardLabel,
  requiresIncidentResolution,
  severityLabel,
  statusLabel,
} from "./utils";

export function CommandSelect({
  children,
  label,
  onChange,
  value,
}: {
  children: ReactNode;
  label: string;
  onChange: (event: SelectChangeEvent) => void;
  value: string;
}) {
  return (
    <FormControl fullWidth size="small">
      <InputLabel>{label}</InputLabel>
      <Select label={label} value={value} onChange={onChange}>
        {children}
      </Select>
    </FormControl>
  );
}

export function IncidentMap({
  incidents,
  onSelect,
  selectedIncidentId,
}: {
  incidents: CommandIncident[];
  onSelect: (incidentId: string) => void;
  selectedIncidentId?: string;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [5.586, -0.18],
      zoom: 11,
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
    if (!layer || !map) {
      return;
    }

    layer.clearLayers();
    if (!incidents.length) {
      return;
    }

    const bounds = L.latLngBounds([]);
    incidents.forEach((incident) => {
      const isSelected = incident.id === selectedIncidentId;
      const marker = L.circleMarker(
        [incident.location.lat, incident.location.lng],
        {
          radius: isSelected ? 13 : 9,
          color: "#FFFFFF",
          weight: isSelected ? 4 : 2,
          fillColor: severityColors[incident.severity],
          fillOpacity: isSelected ? 0.95 : 0.78,
        },
      );
      marker.bindPopup(
        `<strong>${incident.reference}</strong><br>${hazardLabel(incident.type)} · ${severityLabel(incident.severity)}<br>${incident.locality}`,
      );
      marker.on("click", () => onSelect(incident.id));
      marker.addTo(layer);
      bounds.extend([incident.location.lat, incident.location.lng]);
    });

    if (bounds.isValid()) {
      map.fitBounds(bounds.pad(0.18), { animate: true, maxZoom: 13 });
    }
  }, [incidents, onSelect, selectedIncidentId]);

  useEffect(() => {
    const map = mapRef.current;
    const selected = incidents.find(
      (incident) => incident.id === selectedIncidentId,
    );
    if (!map || !selected) {
      return;
    }
    map.flyTo(
      [selected.location.lat, selected.location.lng],
      Math.max(map.getZoom(), 12),
      {
        animate: true,
        duration: 0.45,
      },
    );
  }, [incidents, selectedIncidentId]);

  return (
    <Box className="map-frame">
      <Box ref={containerRef} className="leaflet-command-map" />
      {!incidents.length ? (
        <Box className="map-empty">
          <EmptyState
            title="No map markers"
            detail="No incidents match the current command filters."
          />
        </Box>
      ) : null}
    </Box>
  );
}

export function AlertTargetPreview({ target }: { target: AlertTarget }) {
  const warnings = alertTargetWarnings(target);
  return (
    <Box className="target-preview">
      <Stack direction="row" justifyContent="space-between" gap={1}>
        <Box>
          <Typography variant="subtitle2">Affected area preview</Typography>
          <Typography variant="caption" color="text.secondary">
            {alertTargetSummary(target)}
          </Typography>
        </Box>
        <Chip
          size="small"
          label={alertTargetTypeLabel(target.type)}
          color={target.type === "national" ? "error" : "warning"}
        />
      </Stack>

      <TargetPreviewMap target={target} />

      <Grid container spacing={1}>
        <Grid size={4}>
          <Fact
            label="Area"
            value={`${Math.round((target.areaSqKm ?? 0) * 10) / 10} sq km`}
          />
        </Grid>
        <Grid size={4}>
          <Fact
            label="Population"
            value={`${target.estimatedPopulation ?? 0}`}
          />
        </Grid>
        <Grid size={4}>
          <Fact
            label="Radius"
            value={
              target.radiusMeters
                ? `${Math.round(target.radiusMeters / 100) / 10} km`
                : "Geometry"
            }
          />
        </Grid>
      </Grid>

      {warnings.map((warning) => (
        <Alert severity="warning" key={warning}>
          {warning}
        </Alert>
      ))}
    </Box>
  );
}

export function TargetPreviewMap({ target }: { target: AlertTarget }) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(containerRef.current, {
      center: [target.center?.lat ?? 5.586, target.center?.lng ?? -0.18],
      zoom: 11,
      zoomControl: false,
      scrollWheelZoom: false,
      dragging: false,
      doubleClickZoom: false,
      boxZoom: false,
      keyboard: false,
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
    if (!layer || !map) {
      return;
    }

    layer.clearLayers();
    const color =
      target.type === "national" ? nadaaBrand.colors.red : "#0B6FB8";
    if (target.geometry?.coordinates?.[0]?.length) {
      const polygonPoints = target.geometry.coordinates[0].map(
        ([lng, lat]) => [lat, lng] as [number, number],
      );
      const polygon = L.polygon(polygonPoints, {
        color,
        fillColor: color,
        fillOpacity: 0.18,
        weight: 2,
      }).addTo(layer);
      map.fitBounds(polygon.getBounds().pad(0.12), { animate: false });
      return;
    }

    if (target.center) {
      const radius = target.radiusMeters || 2000;
      const circle = L.circle([target.center.lat, target.center.lng], {
        radius,
        color,
        fillColor: color,
        fillOpacity: 0.18,
        weight: 2,
      }).addTo(layer);
      map.fitBounds(circle.getBounds().pad(0.12), { animate: false });
    }
  }, [target]);

  return <Box ref={containerRef} className="target-preview-map" />;
}

export function AlertWorkflowPanel({
  alerts,
  busy,
  feedback,
  form,
  loadState,
  onCreateDraft,
  onRunAction,
  onUpdateForm,
  selectedIncident,
}: {
  alerts: AuthorityAlertRecord[];
  busy: boolean;
  feedback: string;
  form: AlertFormState;
  loadState: AlertLoadState;
  onCreateDraft: () => void;
  onRunAction: (
    alert: AuthorityAlertRecord,
    action: "submit" | "approve" | "reject" | "emergency-override",
  ) => void;
  onUpdateForm: (
    key: keyof AlertFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  selectedIncident?: CommandIncident;
}) {
  const queueAlerts = alerts.filter(
    (alert) => alert.status !== "published" && alert.status !== "expired",
  );

  return (
    <Paper className="surface alert-panel">
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        className="section-heading"
      >
        <BellRing size={21} color={nadaaBrand.colors.red} />
        <Box>
          <Typography variant="h6">Alert workflow</Typography>
          <Typography variant="caption" color="text.secondary">
            Draft, submit, approve, reject, or override with audit.
          </Typography>
        </Box>
      </Stack>

      {selectedIncident ? (
        <Stack spacing={1.5}>
          <Box>
            <Stack direction="row" justifyContent="space-between" gap={1}>
              <Typography variant="subtitle2">
                Draft from {selectedIncident.reference}
              </Typography>
              <Chip
                size="small"
                label={alertSeverityLabel(form.severity)}
                color={form.severity === "emergency" ? "error" : "warning"}
              />
            </Stack>
            <Typography variant="body2" color="text.secondary">
              {hazardLabel(selectedIncident.type)} · {selectedIncident.district}
            </Typography>
          </Box>

          <TextField
            size="small"
            label="Title"
            value={form.title}
            onChange={onUpdateForm("title")}
          />
          <TextField
            size="small"
            label="Message"
            value={form.message}
            onChange={onUpdateForm("message")}
            multiline
            minRows={3}
          />

          <Grid container spacing={1.25}>
            <Grid size={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Severity</InputLabel>
                <Select
                  label="Severity"
                  value={form.severity}
                  onChange={onUpdateForm("severity")}
                >
                  {alertSeverityOptions.map((severity) => (
                    <MenuItem key={severity} value={severity}>
                      {alertSeverityLabel(severity)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={6}>
              <FormControl fullWidth size="small">
                <InputLabel>Target type</InputLabel>
                <Select
                  label="Target type"
                  value={form.targetType}
                  onChange={onUpdateForm("targetType")}
                >
                  {alertTargetTypeOptions.map((type) => (
                    <MenuItem key={type} value={type}>
                      {alertTargetTypeLabel(type)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={{ xs: 12, sm: 5 }}>
              <TextField
                size="small"
                label="Target IDs"
                value={form.targetIds}
                onChange={onUpdateForm("targetIds")}
                fullWidth
                disabled={form.targetType === "national"}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 7 }}>
              <TextField
                size="small"
                label="Target label"
                value={form.targetLabel}
                onChange={onUpdateForm("targetLabel")}
                fullWidth
              />
            </Grid>
            {form.targetType === "radius" ? (
              <>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Latitude"
                    value={form.targetLatitude}
                    onChange={onUpdateForm("targetLatitude")}
                    fullWidth
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Longitude"
                    value={form.targetLongitude}
                    onChange={onUpdateForm("targetLongitude")}
                    fullWidth
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Radius meters"
                    value={form.targetRadiusMeters}
                    onChange={onUpdateForm("targetRadiusMeters")}
                    fullWidth
                  />
                </Grid>
              </>
            ) : null}
            {form.targetType === "custom" ? (
              <Grid size={12}>
                <TextField
                  size="small"
                  label="Custom polygon JSON"
                  value={form.targetGeometry}
                  onChange={onUpdateForm("targetGeometry")}
                  multiline
                  minRows={3}
                  fullWidth
                />
              </Grid>
            ) : null}
            <Grid size={6}>
              <TextField
                size="small"
                label="Starts"
                value={form.startsAt}
                onChange={onUpdateForm("startsAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
            <Grid size={6}>
              <TextField
                size="small"
                label="Expires"
                value={form.expiresAt}
                onChange={onUpdateForm("expiresAt")}
                type="datetime-local"
                fullWidth
                slotProps={{ inputLabel: { shrink: true } }}
              />
            </Grid>
          </Grid>

          <AlertTargetPreview target={buildAlertTarget(form)} />

          <TextField
            size="small"
            label="Recommended action"
            value={form.recommendedAction}
            onChange={onUpdateForm("recommendedAction")}
          />
          <TextField
            size="small"
            label="Shelter IDs"
            value={form.shelterIds}
            onChange={onUpdateForm("shelterIds")}
          />
          <FormControlLabel
            control={
              <Switch
                checked={form.evacuationRequired}
                onChange={onUpdateForm("evacuationRequired")}
              />
            }
            label="Evacuation required"
          />
          <Button
            variant="contained"
            color="error"
            startIcon={<BellRing size={17} />}
            disabled={busy}
            onClick={onCreateDraft}
          >
            Create draft
          </Button>
        </Stack>
      ) : (
        <EmptyState
          title="No incident selected"
          detail="Choose an incident before drafting an alert."
        />
      )}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          gap={1}
        >
          <Typography variant="subtitle2">Approval queue</Typography>
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live"
                : loadState === "loading"
                  ? "Loading"
                  : "Fixture"
            }
            color={loadState === "ready" ? "success" : "warning"}
          />
        </Stack>
        {feedback ? (
          <Alert
            severity={
              loadState === "ready"
                ? "success"
                : loadState === "loading"
                  ? "info"
                  : "warning"
            }
          >
            {feedback}
          </Alert>
        ) : null}
        {queueAlerts.length ? (
          queueAlerts.slice(0, 4).map((alert) => (
            <Box key={alert.id} className="alert-queue-row">
              <Stack direction="row" justifyContent="space-between" gap={1}>
                <Box>
                  <Typography variant="subtitle2">{alert.title}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {alert.target.label} · {alertSeverityLabel(alert.severity)}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={alertStatusLabel(alert.status)}
                  color={alertStatusColor(alert.status)}
                />
              </Stack>
              <Stack
                direction="row"
                spacing={1}
                flexWrap="wrap"
                className="alert-actions"
              >
                {alert.status === "draft" ? (
                  <Button
                    size="small"
                    variant="outlined"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "submit")}
                  >
                    Submit
                  </Button>
                ) : null}
                {alert.status === "submitted" ? (
                  <>
                    <Button
                      size="small"
                      variant="contained"
                      color="success"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "approve")}
                    >
                      Approve
                    </Button>
                    <Button
                      size="small"
                      variant="outlined"
                      color="error"
                      disabled={busy}
                      onClick={() => onRunAction(alert, "reject")}
                    >
                      Reject
                    </Button>
                  </>
                ) : null}
                {alert.status === "draft" ||
                alert.status === "submitted" ||
                alert.status === "rejected" ? (
                  <Button
                    size="small"
                    color="error"
                    disabled={busy}
                    onClick={() => onRunAction(alert, "emergency-override")}
                  >
                    Override
                  </Button>
                ) : null}
              </Stack>
            </Box>
          ))
        ) : (
          <EmptyState
            title="No alerts in queue"
            detail="Create a draft to begin the approval workflow."
          />
        )}
      </Stack>
    </Paper>
  );
}

export function IncidentDetailPanel({
  abuseBusy,
  abuseFeedback,
  abuseForm,
  assignmentBusy,
  assignmentFeedback,
  assignmentForm,
  busy,
  duplicateCandidates,
  feedback,
  form,
  incident,
  mergeBusy,
  mergeFeedback,
  onAssign,
  onMergeDuplicates,
  onReviewAbuse,
  onToggleDuplicate,
  onUpdateAbuseForm,
  onUpdateAssignmentForm,
  onUpdateForm,
  onUpdateStatus,
  onVerify,
  selectedDuplicateIds,
}: {
  abuseBusy: boolean;
  abuseFeedback: string;
  abuseForm: AbuseReviewFormState;
  assignmentBusy: boolean;
  assignmentFeedback: string;
  assignmentForm: AssignmentFormState;
  busy: boolean;
  duplicateCandidates: DuplicateReviewCandidate[];
  feedback: string;
  form: IncidentStatusFormState;
  incident?: CommandIncident;
  mergeBusy: boolean;
  mergeFeedback: string;
  onAssign: () => void;
  onMergeDuplicates: () => void;
  onReviewAbuse: () => void;
  onToggleDuplicate: (incidentId: string) => void;
  onUpdateAbuseForm: (
    key: keyof AbuseReviewFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateAssignmentForm: (
    key: keyof AssignmentFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateForm: (
    key: keyof IncidentStatusFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  onUpdateStatus: () => void;
  onVerify: () => void;
  selectedDuplicateIds: string[];
}) {
  if (!incident) {
    return (
      <Paper className="surface">
        <EmptyState
          title="No incident selected"
          detail="Choose a map marker or queue row to inspect the incident."
        />
      </Paper>
    );
  }

  const nextStatuses = incidentTransitionOptions[incident.status];
  const terminal = nextStatuses.length === 0;
  const resolutionRequired = requiresIncidentResolution(form.status);
  const canVerify = nextStatuses.includes("verified");
  const canAssign = canAssignIncident(incident.status);
  const activeAssignments = incident.assignments.filter(
    (assignment) => assignment.status === "active",
  );
  const canReviewAbuse =
    incident.source === "api" &&
    incident.status !== "closed" &&
    incident.status !== "false_report";
  const abuseResolutionRequired = abuseForm.decision === "false_report";
  const canMerge =
    incident.source === "api" && selectedDuplicateIds.length > 0 && !mergeBusy;

  return (
    <Paper className="surface detail-panel">
      <Stack
        direction="row"
        justifyContent="space-between"
        gap={1}
        className="section-heading"
      >
        <Box>
          <Typography variant="overline" color="secondary">
            Selected incident
          </Typography>
          <Typography variant="h6">{incident.reference}</Typography>
        </Box>
        <Chip
          size="small"
          label={severityLabel(incident.severity)}
          style={{
            backgroundColor: severityColors[incident.severity],
            color: "#FFFFFF",
          }}
        />
      </Stack>

      <Typography variant="body2" color="text.secondary">
        {incident.description}
      </Typography>

      <Divider className="detail-divider" />

      <Grid container spacing={1.5}>
        <Grid size={6}>
          <Fact label="Hazard" value={hazardLabel(incident.type)} />
        </Grid>
        <Grid size={6}>
          <Fact label="Status" value={statusLabel(incident.status)} />
        </Grid>
        <Grid size={6}>
          <Fact label="People" value={`${incident.peopleAffected}`} />
        </Grid>
        <Grid size={6}>
          <Fact label="Responder ETA" value={incident.responderEta} />
        </Grid>
        <Grid size={12}>
          <Fact label="Assigned agency" value={incident.assignedAgency} />
        </Grid>
      </Grid>

      <Alert
        severity={
          incident.anonymous || !incident.privacy?.reporterContactVisible
            ? "info"
            : "success"
        }
        icon={<ShieldAlert size={18} />}
        className="privacy-alert"
      >
        <Stack spacing={0.75}>
          <Stack direction="row" spacing={1} flexWrap="wrap">
            <Chip size="small" label={privacyReporterLabel(incident)} />
            <Chip size="small" label={privacyContactLabel(incident)} />
            <Chip
              size="small"
              label={`${incident.privacy?.locationPrecision ?? "exact"} location`}
            />
          </Stack>
          <Typography variant="body2">
            {incident.privacy?.disclosure ??
              "Location is used for emergency response coordination."}
          </Typography>
          {incident.privacy?.notes?.length ? (
            <Typography variant="caption" color="text.secondary">
              {incident.privacy.notes[0]}
            </Typography>
          ) : null}
        </Stack>
      </Alert>

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Report safety review</Typography>
            <Typography variant="caption" color="text.secondary">
              {incident.abuseReviewRequired
                ? "Dispatcher review required"
                : "No active safety hold"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={abuseScoreLabel(incident.abuseScore)}
            color={incident.abuseReviewRequired ? "warning" : "default"}
          />
        </Stack>

        {incident.abuseReviewRequired ? (
          <Alert
            severity={incident.priorityReview ? "error" : "warning"}
            icon={<ShieldAlert size={18} />}
          >
            {incident.abuseReviewReason ||
              "Suspicious report signals need dispatcher review."}
          </Alert>
        ) : null}

        {incident.abuseSignals.length ? (
          <Stack spacing={1}>
            {incident.abuseSignals.map((signal) => (
              <Box className="abuse-signal-row" key={signal.code}>
                <Box>
                  <Typography variant="subtitle2">{signal.label}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {signal.detail}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={`${Math.round(signal.weight * 100)}%`}
                  color="warning"
                />
              </Box>
            ))}
          </Stack>
        ) : (
          <Alert severity="info">No suspicious report signals recorded.</Alert>
        )}

        {incident.abuseReviewDecision ? (
          <Alert severity="success">
            Last review: {abuseDecisionLabel(incident.abuseReviewDecision)}
            {incident.abuseReviewedAt
              ? ` at ${formatShortTime(incident.abuseReviewedAt)}`
              : ""}
          </Alert>
        ) : null}

        {abuseFeedback ? (
          <Alert
            severity={
              abuseFeedback.includes("needs") || abuseFeedback.includes("valid")
                ? "warning"
                : "success"
            }
          >
            {abuseFeedback}
          </Alert>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 5 }}>
            <FormControl fullWidth size="small" disabled={!canReviewAbuse}>
              <InputLabel>Decision</InputLabel>
              <Select
                label="Decision"
                value={abuseForm.decision}
                onChange={onUpdateAbuseForm("decision")}
              >
                {(["clear", "monitor", "false_report"] as const).map(
                  (decision) => (
                    <MenuItem value={decision} key={decision}>
                      {abuseDecisionLabel(decision)}
                    </MenuItem>
                  ),
                )}
              </Select>
            </FormControl>
          </Grid>
          <Grid size={{ xs: 12, sm: 7 }}>
            <TextField
              size="small"
              label="Review note"
              value={abuseForm.note}
              onChange={onUpdateAbuseForm("note")}
              disabled={!canReviewAbuse}
              fullWidth
            />
          </Grid>
        </Grid>

        {abuseResolutionRequired ? (
          <TextField
            size="small"
            label="False report resolution"
            value={abuseForm.resolutionNotes}
            onChange={onUpdateAbuseForm("resolutionNotes")}
            disabled={!canReviewAbuse}
            multiline
            minRows={3}
          />
        ) : null}

        <Button
          variant="outlined"
          disabled={
            abuseBusy ||
            !canReviewAbuse ||
            !abuseForm.note.trim() ||
            (abuseResolutionRequired && !abuseForm.resolutionNotes.trim())
          }
          onClick={onReviewAbuse}
          startIcon={<ShieldAlert size={17} />}
        >
          Save safety review
        </Button>

        {incident.source !== "api" ? (
          <Alert severity="info">
            Start incident-service to save fixture safety reviews.
          </Alert>
        ) : null}
      </Stack>

      <Divider className="detail-divider" />

      <Stack spacing={1}>
        <Typography variant="subtitle2">Response timeline</Typography>
        {incident.timelineEntries.map((event) => (
          <Box className="timeline-row" key={event}>
            <Typography variant="body2">{event}</Typography>
          </Box>
        ))}
      </Stack>

      {incident.duplicateCandidates.length ||
      duplicateCandidates.length ||
      incident.mergedIncidentIds.length ||
      incident.mergedIntoId ? (
        <>
          <Divider className="detail-divider" />
          <Stack spacing={1.25}>
            <Stack direction="row" justifyContent="space-between" gap={1}>
              <Box>
                <Typography variant="subtitle2">Duplicate review</Typography>
                <Typography variant="caption" color="text.secondary">
                  {duplicateCandidates.length
                    ? "Side-by-side candidate check"
                    : "No open candidates"}
                </Typography>
              </Box>
              <Chip
                size="small"
                label={`${duplicateCandidates.length} candidate${
                  duplicateCandidates.length === 1 ? "" : "s"
                }`}
                color={duplicateCandidates.length ? "warning" : "default"}
              />
            </Stack>

            {incident.mergedIntoId ? (
              <Alert severity="info">
                This report was merged into another incident and remains
                traceable in audit and timeline history.
              </Alert>
            ) : null}

            {incident.mergedIncidentIds.length ? (
              <Alert severity="success">
                {incident.mergedIncidentIds.length} duplicate report
                {incident.mergedIncidentIds.length === 1 ? "" : "s"} already
                merged into this incident.
              </Alert>
            ) : null}

            {mergeFeedback ? (
              <Alert
                severity={
                  mergeFeedback.includes("needs") ? "warning" : "success"
                }
              >
                {mergeFeedback}
              </Alert>
            ) : null}

            {duplicateCandidates.map((item) => (
              <Box className="duplicate-review-row" key={item.incident.id}>
                <FormControlLabel
                  control={
                    <Checkbox
                      size="small"
                      checked={selectedDuplicateIds.includes(item.incident.id)}
                      onChange={() => onToggleDuplicate(item.incident.id)}
                      disabled={incident.source !== "api" || mergeBusy}
                    />
                  }
                  label={item.incident.reference}
                />
                <Box className="duplicate-comparison">
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Selected
                    </Typography>
                    <Typography variant="body2">
                      {incident.description}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Candidate
                    </Typography>
                    <Typography variant="body2">
                      {item.incident.description}
                    </Typography>
                  </Box>
                </Box>
                <Stack direction="row" spacing={0.75} flexWrap="wrap">
                  <Chip
                    size="small"
                    label={`${Math.round(item.candidate.score * 100)}%`}
                  />
                  <Chip
                    size="small"
                    label={`${Math.round(item.candidate.distanceMeters)}m`}
                  />
                  <Chip
                    size="small"
                    label={`${item.candidate.minutesApart}m`}
                  />
                </Stack>
              </Box>
            ))}

            {incident.source !== "api" && duplicateCandidates.length ? (
              <Alert severity="info">
                Start incident-service to merge fixture duplicate reports.
              </Alert>
            ) : null}

            {duplicateCandidates.length ? (
              <Button
                variant="outlined"
                disabled={!canMerge}
                onClick={onMergeDuplicates}
                startIcon={<GitMerge size={17} />}
              >
                Merge selected
              </Button>
            ) : null}
          </Stack>
        </>
      ) : null}

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Agency assignment</Typography>
            <Typography variant="caption" color="text.secondary">
              {canAssign ? "Dispatch coordination" : "Verification required"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={activeAssignments.length ? "Assigned" : "Unassigned"}
            color={activeAssignments.length ? "success" : "default"}
          />
        </Stack>

        {activeAssignments.length ? (
          <Stack spacing={1}>
            {activeAssignments.map((assignment) => (
              <Box className="assignment-row" key={assignment.id}>
                <Box>
                  <Typography variant="subtitle2">
                    {assignment.agencyName}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {assignment.responderLead || "Response lead pending"}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={assignment.priority}
                  color={assignment.priority === "urgent" ? "error" : "warning"}
                />
              </Box>
            ))}
          </Stack>
        ) : null}

        {assignmentFeedback ? (
          <Alert
            severity={
              assignmentFeedback.includes("needs") ? "warning" : "success"
            }
          >
            {assignmentFeedback}
          </Alert>
        ) : null}

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 7 }}>
            <FormControl fullWidth size="small" disabled={!canAssign}>
              <InputLabel>Agency</InputLabel>
              <Select
                label="Agency"
                value={assignmentForm.agencyId}
                onChange={onUpdateAssignmentForm("agencyId")}
              >
                {assignmentAgencyOptions.map((agency) => (
                  <MenuItem value={agency.id} key={agency.id}>
                    {agency.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid size={{ xs: 12, sm: 5 }}>
            <FormControl fullWidth size="small" disabled={!canAssign}>
              <InputLabel>Priority</InputLabel>
              <Select
                label="Priority"
                value={assignmentForm.priority}
                onChange={onUpdateAssignmentForm("priority")}
              >
                {(["low", "normal", "high", "urgent"] as const).map(
                  (priority) => (
                    <MenuItem value={priority} key={priority}>
                      {priority}
                    </MenuItem>
                  ),
                )}
              </Select>
            </FormControl>
          </Grid>
        </Grid>

        <TextField
          size="small"
          label="Instructions"
          value={assignmentForm.instructions}
          onChange={onUpdateAssignmentForm("instructions")}
          disabled={!canAssign}
          multiline
          minRows={2}
        />

        <TextField
          size="small"
          label="Responder lead"
          value={assignmentForm.responderLead}
          onChange={onUpdateAssignmentForm("responderLead")}
          disabled={!canAssign}
        />

        <Button
          variant="outlined"
          disabled={
            assignmentBusy || !canAssign || !assignmentForm.instructions.trim()
          }
          onClick={onAssign}
          startIcon={<Truck size={17} />}
        >
          Assign agency
        </Button>
      </Stack>

      <Divider className="detail-divider" />

      <Stack spacing={1.25}>
        <Stack direction="row" justifyContent="space-between" gap={1}>
          <Box>
            <Typography variant="subtitle2">Status workflow</Typography>
            <Typography variant="caption" color="text.secondary">
              {terminal
                ? "Terminal incident state"
                : "Audited dispatcher action"}
            </Typography>
          </Box>
          <Chip
            size="small"
            label={incident.source === "api" ? "Live" : "Fixture"}
            color={incident.source === "api" ? "success" : "warning"}
          />
        </Stack>

        {feedback ? (
          <Alert
            severity={
              feedback.includes("needs") || feedback.includes("valid")
                ? "warning"
                : "success"
            }
          >
            {feedback}
          </Alert>
        ) : null}

        <FormControl fullWidth size="small" disabled={terminal}>
          <InputLabel>Next status</InputLabel>
          <Select
            label="Next status"
            value={form.status}
            onChange={onUpdateForm("status")}
          >
            {(nextStatuses.length ? nextStatuses : [incident.status]).map(
              (status) => (
                <MenuItem value={status} key={status}>
                  {statusLabel(status)}
                </MenuItem>
              ),
            )}
          </Select>
        </FormControl>

        <TextField
          size="small"
          label="Status note"
          value={form.note}
          onChange={onUpdateForm("note")}
          multiline
          minRows={2}
        />

        {resolutionRequired ? (
          <TextField
            size="small"
            label="Resolution notes"
            value={form.resolutionNotes}
            onChange={onUpdateForm("resolutionNotes")}
            multiline
            minRows={3}
          />
        ) : null}
      </Stack>

      <Stack direction="row" spacing={1} className="detail-actions">
        <Button
          variant="contained"
          disabled={busy || !canVerify}
          onClick={onVerify}
          startIcon={<CheckCheck size={17} />}
        >
          Verify
        </Button>
        <Button
          variant="outlined"
          disabled={
            busy ||
            terminal ||
            (resolutionRequired && !form.resolutionNotes.trim())
          }
          onClick={onUpdateStatus}
          startIcon={<Truck size={17} />}
        >
          Update status
        </Button>
      </Stack>
    </Paper>
  );
}

export function Fact({ label, value }: { label: string; value: string }) {
  return (
    <Box className="fact">
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="subtitle2">{value}</Typography>
    </Box>
  );
}

export function PrivacyChip({ incident }: { incident: CommandIncident }) {
  return (
    <Chip
      size="small"
      label={privacyReporterLabel(incident)}
      color={incident.anonymous ? "warning" : "default"}
    />
  );
}

function privacyReporterLabel(incident: CommandIncident) {
  if (incident.anonymous) {
    return "Anonymous";
  }
  if (incident.privacy?.reporterIdentityVisible) {
    return "Identity visible";
  }
  return "Identity hidden";
}

function privacyContactLabel(incident: CommandIncident) {
  if (incident.privacy?.reporterContactVisible) {
    return "Contact visible";
  }
  if (incident.contactPermission) {
    return "Contact restricted";
  }
  return "Contact denied";
}

export function StatusLine({
  color,
  label,
  value,
}: {
  color: "success" | "warning" | "default";
  label: string;
  value: string;
}) {
  return (
    <Stack
      direction="row"
      justifyContent="space-between"
      alignItems="center"
      gap={1}
    >
      <Typography variant="body2">{label}</Typography>
      <Chip size="small" label={value} color={color} />
    </Stack>
  );
}

export function EmptyState({
  detail,
  title,
}: {
  detail: string;
  title: string;
}) {
  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      spacing={1}
      className="empty-state"
    >
      <Crosshair size={28} color={nadaaBrand.colors.slate} />
      <Typography variant="subtitle2">{title}</Typography>
      <Typography variant="body2" color="text.secondary" textAlign="center">
        {detail}
      </Typography>
    </Stack>
  );
}
