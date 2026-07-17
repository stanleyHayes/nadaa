import { type ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
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

import "leaflet/dist/leaflet.css";
import {
  BrainCircuit,
  CheckCircle2,
  ListChecks,
  ShieldAlert,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { IncidentTriageSuggestion } from "@nadaa/shared-types";
import type {
  TriageLoadState,
  TriageSuggestionFormState,
  TriageSuggestionReview,
} from "../types";
import { assignmentAgencyOptions, triageSeverityOptions } from "../data";

import {
  agencyTypeLabel,
  contributionLabel,
  contributionProgress,
  severityLabel,
  triageConfidenceLabel,
} from "../utils";

import { EmptyState, Fact, SkeletonRows } from "./shared";

export function AITriageSuggestionPanel({
  busy,
  canAccept,
  canOverride,
  feedback,
  form,
  incident,
  loadMessage,
  loadState,
  onAccept,
  onOverride,
  onRefresh,
  onUpdateForm,
  populationError,
  reasonError,
  suggestion,
}: {
  busy: boolean;
  canAccept: boolean;
  canOverride: boolean;
  feedback: string;
  form: TriageSuggestionFormState;
  incident?: TriageSuggestionReview;
  loadMessage: string;
  loadState: TriageLoadState;
  onAccept: () => void;
  onOverride: () => void;
  onRefresh: () => void;
  onUpdateForm: (
    key: keyof TriageSuggestionFormState,
  ) => (
    event:
      ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
  ) => void;
  populationError: string;
  reasonError: string;
  suggestion?: IncidentTriageSuggestion;
}) {
  const live = loadState === "ready";

  return (
    <Paper className="surface triage-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          gap: 1.5
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <BrainCircuit size={22} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h5">AI incident triage</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Severity, duplicate likelihood, affected population, and agency
              routing suggestion.
            </Typography>
          </Box>
        </Stack>
        <Stack
          direction="row"
          spacing={1}
          sx={{
            alignItems: "center",
            flexWrap: "wrap"
          }}>
          <Chip
            size="small"
            label={
              live
                ? "Live"
                : loadState === "loading"
                  ? "Loading"
                  : loadState === "error"
                    ? "Unavailable"
                    : "Fixture"
            }
            color={live ? "success" : "warning"}
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<BrainCircuit size={16} />}
            disabled={loadState === "loading" || !incident}
            onClick={onRefresh}
          >
            Refresh triage
          </Button>
        </Stack>
      </Stack>
      {loadState === "fallback" ||
      loadState === "error" ||
      loadState === "empty" ? (
        <Alert
          severity={loadState === "empty" ? "info" : "warning"}
          className="ml-review-alert"
        >
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? <SkeletonRows /> : null}
      {suggestion ? (
        <Stack spacing={1.5}>
          <Alert severity="info" icon={<ShieldAlert size={18} />}>
            Human review is required. This suggestion cannot auto-verify,
            auto-assign, or auto-publish an alert.
          </Alert>

          <Grid container spacing={1}>
            <Grid size={6}>
              <Fact
                label="Severity"
                value={severityLabel(suggestion.severity)}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Confidence"
                value={triageConfidenceLabel(suggestion.confidence)}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Duplicate likelihood"
                value={`${Math.round(suggestion.duplicateLikelihood * 100)}%`}
              />
            </Grid>
            <Grid size={6}>
              <Fact
                label="Affected population"
                value={`${suggestion.affectedPopulation.toLocaleString()} (estimate)`}
              />
            </Grid>
            <Grid size={12}>
              <Fact
                label="Suggested agency"
                value={`${suggestion.suggestedAgency.name} (${agencyTypeLabel(suggestion.suggestedAgency.agencyType)})`}
              />
            </Grid>
          </Grid>

          <Stack spacing={1}>
            <Stack direction="row" spacing={1} sx={{
              alignItems: "center"
            }}>
              <ListChecks size={18} color={nadaaBrand.colors.green} />
              <Typography variant="subtitle2">Explanation factors</Typography>
            </Stack>
            {suggestion.explanationFactors.map((factor) => (
              <Box className="factor-row" key={factor.feature}>
                <Stack
                  direction="row"
                  sx={{
                    justifyContent: "space-between",
                    gap: 1
                  }}>
                  <Box>
                    <Typography variant="body2">{factor.label}</Typography>
                    <Typography variant="caption" sx={{
                      color: "text.secondary"
                    }}>
                      {String(factor.value)} ·{" "}
                      {factor.direction === "increases_risk"
                        ? "Increases risk"
                        : "Reduces risk"}
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    color={
                      factor.direction === "increases_risk"
                        ? "warning"
                        : "success"
                    }
                    label={contributionLabel(factor.contribution)}
                  />
                </Stack>
                <LinearProgress
                  variant="determinate"
                  value={contributionProgress(factor.contribution)}
                  color={
                    factor.direction === "increases_risk"
                      ? "warning"
                      : "success"
                  }
                />
              </Box>
            ))}
          </Stack>

          <Divider className="detail-divider" />

          <Typography variant="subtitle2">Dispatcher override</Typography>

          <Grid container spacing={1.25}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <FormControl fullWidth size="small">
                <InputLabel>Severity</InputLabel>
                <Select
                  label="Severity"
                  value={form.severity}
                  onChange={onUpdateForm("severity")}
                >
                  {triageSeverityOptions.map((severity) => (
                    <MenuItem key={severity} value={severity}>
                      {severityLabel(severity)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                size="small"
                label="Affected population"
                value={form.affectedPopulation}
                onChange={onUpdateForm("affectedPopulation")}
                error={!!populationError}
                helperText={populationError || undefined}
                fullWidth
              />
            </Grid>
            <Grid size={12}>
              <FormControl fullWidth size="small">
                <InputLabel>Agency</InputLabel>
                <Select
                  label="Agency"
                  value={form.agencyId}
                  onChange={onUpdateForm("agencyId")}
                >
                  {assignmentAgencyOptions.map((agency) => (
                    <MenuItem key={agency.id} value={agency.id}>
                      {agency.name} ({agencyTypeLabel(agency.type)})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
          </Grid>

          <TextField
            size="small"
            label="Override reason"
            value={form.reason}
            onChange={onUpdateForm("reason")}
            helperText={
              reasonError || "Required when overriding. Logged for review."
            }
            multiline
            minRows={2}
          />

          {feedback ? (
            <Alert
              severity={
                feedback.includes("not logged") ||
                feedback.includes("not recorded") ||
                feedback.includes("unavailable") ||
                feedback.includes("needs") ||
                feedback.includes("only be logged")
                  ? "warning"
                  : "success"
              }
            >
              {feedback}
            </Alert>
          ) : null}

          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              startIcon={<CheckCircle2 size={17} />}
              disabled={busy || !canAccept}
              onClick={onAccept}
            >
              Accept suggestion
            </Button>
            <Button
              variant="contained"
              color="error"
              startIcon={<ShieldAlert size={17} />}
              disabled={
                busy || !canOverride || !!reasonError || !!populationError
              }
              onClick={onOverride}
            >
              Override
            </Button>
          </Stack>

          <Typography variant="caption" sx={{
            color: "text.secondary"
          }}>
            Model {suggestion.modelVersion}
            {suggestion.suggestionId
              ? ` · Suggestion ${suggestion.suggestionId}`
              : " · Fixture suggestion (not logged)"}
          </Typography>
        </Stack>
      ) : (
        <EmptyState
          title="No triage suggestion"
          detail={
            incident
              ? "Refresh to generate a suggestion."
              : "Select an incident first."
          }
        />
      )}
    </Paper>
  );
}
