import { type ChangeEvent, useCallback, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  LinearProgress,
  Paper,
  Stack,
  TextField,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";

import "leaflet/dist/leaflet.css";
import {
  BrainCircuit,
  FileText,
  ListChecks,
  ShieldAlert,
  X,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";

import type { MLPredictionReview, MLReviewLoadState } from "../types";
import {
  confidenceLabel,
  contributionLabel,
  contributionProgress,
  expectedOnsetLabel,
  probabilityLabel,
  severityLabel,
} from "../utils";

import { Fact, PredictionReviewMap, SeverityChip, SkeletonRows } from "./shared";

export function MLPredictionReviewPanel({
  busy,
  feedback,
  loadMessage,
  loadState,
  onCreateDraft,
  onRefresh,
  onSelectPrediction,
  onUpdateReviewNote,
  predictions,
  reviewNote,
  selectedPrediction,
  selectedPredictionId,
}: {
  busy: boolean;
  feedback: string;
  loadMessage: string;
  loadState: MLReviewLoadState;
  onCreateDraft: () => void;
  onRefresh: () => void;
  onSelectPrediction: (predictionId: string) => void;
  onUpdateReviewNote: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  predictions: MLPredictionReview[];
  reviewNote: string;
  selectedPrediction?: MLPredictionReview;
  selectedPredictionId: string;
}) {
  const live = loadState === "ready";

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));
  const [detailOpen, setDetailOpen] = useState(false);

  const openPrediction = useCallback(
    (predictionId: string) => {
      onSelectPrediction(predictionId);
      setDetailOpen(true);
    },
    [onSelectPrediction],
  );

  const closeDetail = () => setDetailOpen(false);

  return (
    <Paper className="surface ml-review-panel">
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
            <Typography variant="h5">ML flood review</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Review probability, severity, confidence, and explanation before
              drafting an alert.
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
                ? "Live ML"
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
            disabled={loadState === "loading"}
            onClick={onRefresh}
          >
            Refresh ML
          </Button>
        </Stack>
      </Stack>
      {loadState === "fallback" || loadState === "error" ? (
        <Alert severity="warning" className="ml-review-alert">
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? <SkeletonRows /> : null}
      <PredictionReviewMap
        predictions={predictions}
        selectedPredictionId={selectedPredictionId}
        onSelect={openPrediction}
      />
      <Stack className="prediction-list" spacing={1}>
        {predictions.map((prediction) => (
          <Box
            key={prediction.id}
            className={`prediction-row${
              prediction.id === selectedPredictionId ? " selected" : ""
            }`}
            onClick={() => openPrediction(prediction.id)}
          >
            <Stack
              direction="row"
              sx={{
                justifyContent: "space-between",
                gap: 1,
                alignItems: "flex-start"
              }}>
              <Box>
                <Typography variant="subtitle2">
                  {prediction.community}
                </Typography>
                <Typography variant="caption" sx={{
                  color: "text.secondary"
                }}>
                  {prediction.district} · {prediction.cellId}
                </Typography>
              </Box>
              <SeverityChip severity={prediction.severity} />
            </Stack>
            <Stack direction="row" spacing={1} sx={{
              flexWrap: "wrap"
            }}>
              <Chip
                size="small"
                label={probabilityLabel(prediction.probability)}
              />
              <Chip
                size="small"
                label={confidenceLabel(prediction.confidence)}
              />
              <Chip
                size="small"
                label={expectedOnsetLabel(prediction.expectedOnset)}
              />
              {prediction.reviewStatus === "draft_created" ? (
                <Chip size="small" color="success" label="Draft created" />
              ) : null}
            </Stack>
          </Box>
        ))}
      </Stack>
      <Dialog
        open={detailOpen && Boolean(selectedPrediction)}
        onClose={closeDetail}
        maxWidth="md"
        fullWidth
        scroll="paper"
        fullScreen={fullScreen}
      >
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "flex-start",
            justifyContent: "space-between",
            gap: 2,
          }}
        >
          <Box>
            <Typography
              component="span"
              sx={{ display: "block", fontWeight: 800 }}
            >
              {selectedPrediction?.community ?? "Prediction"}
            </Typography>
            {selectedPrediction ? (
              <Typography
                component="span"
                sx={{
                  display: "block",
                  color: "text.secondary",
                  fontSize: "0.85rem",
                  fontWeight: 600,
                }}
              >
                {selectedPrediction.district} · {selectedPrediction.cellId}
              </Typography>
            ) : null}
          </Box>
          <IconButton aria-label="Close" onClick={closeDetail} size="small">
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {selectedPrediction ? (
            <Stack spacing={1.25} className="prediction-detail">
              <Stack
                direction="row"
                sx={{
                  justifyContent: "space-between",
                  gap: 1
                }}>
                <Box>
                  <Typography variant="overline" color="secondary">
                    Selected prediction
                  </Typography>
                  <Typography variant="h6">
                    {selectedPrediction.community}
                  </Typography>
                </Box>
                <Chip
                  size="small"
                  label={probabilityLabel(selectedPrediction.probability)}
                  color={
                    selectedPrediction.severity === "severe" ||
                    selectedPrediction.severity === "emergency"
                      ? "error"
                      : selectedPrediction.severity === "low"
                        ? "success"
                        : "warning"
                  }
                />
              </Stack>

              <Grid container spacing={1}>
                <Grid size={6}>
                  <Fact
                    label="Severity"
                    value={severityLabel(selectedPrediction.severity)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact
                    label="Confidence"
                    value={confidenceLabel(selectedPrediction.confidence)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact
                    label="Expected onset"
                    value={expectedOnsetLabel(selectedPrediction.expectedOnset)}
                  />
                </Grid>
                <Grid size={6}>
                  <Fact label="Model" value={selectedPrediction.modelVersion} />
                </Grid>
              </Grid>

              <Alert severity="info" icon={<ShieldAlert size={18} />}>
                Human review is required and this prediction cannot auto-publish
                a public alert.
              </Alert>

              <Stack spacing={1}>
                <Stack direction="row" spacing={1} sx={{
                  alignItems: "center"
                }}>
                  <ListChecks size={18} color={nadaaBrand.colors.green} />
                  <Typography variant="subtitle2">
                    Explanation factors
                  </Typography>
                </Stack>
                {selectedPrediction.explanationFactors.map((factor) => (
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

              <TextField
                size="small"
                label="Review note"
                value={reviewNote}
                onChange={onUpdateReviewNote}
                multiline
                minRows={2}
              />

              {feedback ? (
                <Alert
                  severity={
                    feedback.includes("unavailable") ? "warning" : "success"
                  }
                >
                  {feedback}
                </Alert>
              ) : null}

              <Button
                variant="contained"
                color="error"
                startIcon={<FileText size={17} />}
                disabled={busy || loadState !== "ready"}
                onClick={onCreateDraft}
              >
                Create reviewed draft
              </Button>
              {!live ? (
                <Typography variant="caption" sx={{
                  color: "text.secondary"
                }}>
                  Drafts require live ML predictions; fixture data can never
                  drive a real alert draft.
                </Typography>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
      </Dialog>
    </Paper>
  );
}
