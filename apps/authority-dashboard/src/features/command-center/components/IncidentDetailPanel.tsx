import { type ChangeEvent } from "react";
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
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";

import "leaflet/dist/leaflet.css";
import { CheckCheck, GitMerge, ShieldAlert, Truck } from "lucide-react";

import type { DuplicateReviewCandidate } from "@nadaa/shared-types";
import { assignmentAgencyOptions, incidentTransitionOptions } from "../data";
import type {
  AbuseReviewFormState,
  AssignmentFormState,
  CommandIncident,
  IncidentStatusFormState,
} from "../types";
import {
  abuseDecisionLabel,
  abuseScoreLabel,
  canAssignIncident,
  formatShortTime,
  hazardLabel,
  requiresIncidentResolution,
  statusLabel,
} from "../utils";

import {
  EmptyState,
  Fact,
  SeverityChip,
  privacyContactLabel,
  privacyReporterLabel,
} from "./shared";

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
        <SeverityChip severity={incident.severity} />
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
