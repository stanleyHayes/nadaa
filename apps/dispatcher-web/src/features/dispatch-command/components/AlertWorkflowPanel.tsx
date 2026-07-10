import { type ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
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

import "leaflet/dist/leaflet.css";
import { BellRing } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { AuthorityAlertRecord } from "@nadaa/shared-types";

import { alertSeverityOptions, alertTargetTypeOptions } from "../data";
import type { AlertFormState, AlertLoadState, CommandIncident } from "../types";
import {
  alertSeverityLabel,
  alertStatusColor,
  alertStatusLabel,
  alertTargetTypeLabel,
  buildAlertTarget,
  hazardLabel,
  parseTargetGeometry,
} from "../utils";

import { AlertTargetPreview, EmptyState } from "./shared";

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
  const radiusFieldsInvalid =
    form.targetType === "radius" &&
    (!form.targetLatitude.trim() ||
      !form.targetLongitude.trim() ||
      !form.targetRadiusMeters.trim());
  const customGeometryInvalid =
    form.targetType === "custom" &&
    form.targetGeometry.trim() !== "" &&
    !parseTargetGeometry(form.targetGeometry);

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
                    error={radiusFieldsInvalid && !form.targetLatitude.trim()}
                    helperText={
                      radiusFieldsInvalid && !form.targetLatitude.trim()
                        ? "Latitude is required"
                        : ""
                    }
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Longitude"
                    value={form.targetLongitude}
                    onChange={onUpdateForm("targetLongitude")}
                    fullWidth
                    error={radiusFieldsInvalid && !form.targetLongitude.trim()}
                    helperText={
                      radiusFieldsInvalid && !form.targetLongitude.trim()
                        ? "Longitude is required"
                        : ""
                    }
                  />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                  <TextField
                    size="small"
                    label="Radius meters"
                    value={form.targetRadiusMeters}
                    onChange={onUpdateForm("targetRadiusMeters")}
                    fullWidth
                    error={
                      radiusFieldsInvalid && !form.targetRadiusMeters.trim()
                    }
                    helperText={
                      radiusFieldsInvalid && !form.targetRadiusMeters.trim()
                        ? "Radius is required"
                        : ""
                    }
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
                  error={customGeometryInvalid}
                  helperText={
                    customGeometryInvalid ? "Enter valid GeoJSON polygon" : ""
                  }
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
