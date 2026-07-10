import { type ChangeEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import {
  AlertTriangle,
  CheckCircle2,
  Eye,
  Image,
  ThumbsDown,
  ThumbsUp,
  XCircle,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CVAnalysisRequest,
  CVAnalysisResult,
  CVImageItem,
  CVReviewStatus,
} from "@nadaa/shared-types";
import { CV_API_BASE } from "@/app/config";
import { EmptyState } from "./shared";

const cvConfidenceThreshold = 0.7;

function confidenceColor(confidence: number): string {
  if (confidence >= cvConfidenceThreshold) {
    return nadaaBrand.colors.green;
  }
  if (confidence >= 0.5) {
    return nadaaBrand.colors.gold;
  }
  return nadaaBrand.colors.red;
}

function labelSeverityRole(
  label: string,
): "success" | "warning" | "error" | "info" {
  if (label === "flood_evidence" || label === "fire_evidence") {
    return "error";
  }
  if (label === "smoke_evidence") {
    return "warning";
  }
  if (label === "sensitive" || label === "person_in_distress") {
    return "error";
  }
  if (label === "no_evidence" || label === "unclear") {
    return "info";
  }
  return "info";
}

export function CVEvidencePanel() {
  const [images, setImages] = useState<CVImageItem[]>([]);
  const [loadState, setLoadState] = useState<"loading" | "ready" | "fallback">(
    "loading",
  );
  const [feedback, setFeedback] = useState("Loading CV results");
  const [selectedImageId, setSelectedImageId] = useState("");
  const [reviewNote, setReviewNote] = useState("");
  const [reviewBusy, setReviewBusy] = useState(false);

  const refreshCVResults = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading CV results");

    try {
      const response = await fetch(`${CV_API_BASE}/cv/results`, {
        signal,
      });
      if (!response.ok) {
        throw new Error(`CV API returned ${response.status}`);
      }

      const payload = (await response.json()) as {
        results: CVAnalysisResult[];
      };
      const nextImages = payload.results.map(
        (result) =>
          ({
            id: result.imageId,
            url: "/brand/nadaa-logo.png",
            name: result.imageId,
            incidentId: undefined,
            uploadedAt: result.createdAt,
            status:
              (result.reviewStatus as CVImageItem["status"]) ?? "analyzed",
            cvResult: result,
          }) satisfies CVImageItem,
      );
      setImages(nextImages);
      setSelectedImageId(nextImages[0]?.id ?? "");
      setLoadState("ready");
      setFeedback(
        nextImages.length
          ? "CV API connected."
          : "CV API connected. No analyzed evidence yet.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setImages([]);
      setSelectedImageId("");
      setLoadState("fallback");
      setFeedback(
        "CV / ML service unavailable. Start the ml-service on the configured URL to load evidence analysis.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshCVResults(controller.signal);
    return () => controller.abort();
  }, []);

  const selectedImage = useMemo(
    () => images.find((img) => img.id === selectedImageId) ?? images[0],
    [images, selectedImageId],
  );

  const imagesNeedingReview = useMemo(
    () =>
      images.filter(
        (img) =>
          img.cvResult?.humanReviewRequired &&
          img.cvResult?.reviewStatus === "pending",
      ),
    [images],
  );

  const runCVAnalysis = async (image: CVImageItem) => {
    setReviewBusy(true);
    try {
      const body: CVAnalysisRequest = {
        imageId: image.id,
        imageName: image.name,
      };
      const response = await fetch(`${CV_API_BASE}/cv/analyze`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!response.ok) {
        throw new Error(`CV API returned ${response.status}`);
      }
      const payload = (await response.json()) as {
        result: CVAnalysisResult;
      };
      setImages((current) =>
        current.map((img) =>
          img.id === image.id
            ? { ...img, cvResult: payload.result, status: "analyzed" }
            : img,
        ),
      );
      setFeedback(`CV analysis completed for ${image.name}.`);
    } catch (error) {
      setFeedback(
        `CV analysis needs the ml-service running on ${CV_API_BASE}.`,
      );
    } finally {
      setReviewBusy(false);
    }
  };

  const updateReviewStatus = (image: CVImageItem, status: CVReviewStatus) => {
    if (!image.cvResult) {
      return;
    }
    setReviewBusy(true);
    setTimeout(() => {
      setImages((current) =>
        current.map((img) =>
          img.id === image.id
            ? {
                ...img,
                status,
                cvResult: {
                  ...img.cvResult!,
                  reviewStatus: status,
                  reviewedBy: "authority-dashboard",
                  reviewNote: reviewNote.trim() || undefined,
                },
              }
            : img,
        ),
      );
      setReviewNote("");
      setReviewBusy(false);
      setFeedback(`${image.name} marked as ${status}.`);
    }, 400);
  };

  return (
    <Paper className="surface cv-panel">
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        className="section-heading"
      >
        <Image size={21} color="var(--nadaa-navy)" />
        <Box>
          <Typography variant="h6">CV Evidence Panel</Typography>
          <Typography variant="caption" color="text.secondary">
            Decision-support image analysis; human review required for
            low-confidence or sensitive results
          </Typography>
        </Box>
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
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}

      {imagesNeedingReview.length > 0 && (
        <Alert severity="warning" icon={<AlertTriangle size={18} />}>
          {imagesNeedingReview.length} image
          {imagesNeedingReview.length === 1 ? "" : "s"} require human review
        </Alert>
      )}

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Analyzed images</Typography>
        <Stack direction="row" spacing={1} flexWrap="wrap">
          {images.map((image) => (
            <Button
              key={image.id}
              size="small"
              variant={image.id === selectedImageId ? "contained" : "outlined"}
              onClick={() => setSelectedImageId(image.id)}
              startIcon={
                image.cvResult?.humanReviewRequired ? (
                  <AlertTriangle size={14} />
                ) : (
                  <Eye size={14} />
                )
              }
              color={
                image.cvResult?.humanReviewRequired ? "warning" : "primary"
              }
            >
              {image.name}
            </Button>
          ))}
        </Stack>

        {selectedImage ? (
          <CVImageDetail
            image={selectedImage}
            reviewBusy={reviewBusy}
            reviewNote={reviewNote}
            onReviewNoteChange={(event: ChangeEvent<HTMLInputElement>) =>
              setReviewNote(event.target.value)
            }
            onApprove={() => updateReviewStatus(selectedImage, "approved")}
            onReject={() => updateReviewStatus(selectedImage, "rejected")}
            onReanalyze={() => runCVAnalysis(selectedImage)}
          />
        ) : (
          <EmptyState
            title="No images"
            detail="No images have been analyzed yet."
          />
        )}
      </Stack>
    </Paper>
  );
}

function CVImageDetail({
  image,
  reviewBusy,
  reviewNote,
  onReviewNoteChange,
  onApprove,
  onReject,
  onReanalyze,
}: {
  image: CVImageItem;
  reviewBusy: boolean;
  reviewNote: string;
  onReviewNoteChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onApprove: () => void;
  onReject: () => void;
  onReanalyze: () => void;
}) {
  const result = image.cvResult;

  return (
    <Stack spacing={1.5}>
      <Box
        sx={{
          width: "100%",
          height: 180,
          backgroundColor: nadaaBrand.colors.slate + "20",
          borderRadius: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          border: `1px solid ${nadaaBrand.colors.slate}30`,
        }}
      >
        <Image size={48} color="var(--nadaa-slate)" />
      </Box>

      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Box>
          <Typography variant="subtitle2">{image.name}</Typography>
          <Typography variant="caption" color="text.secondary">
            {image.incidentId
              ? `Incident: ${image.incidentId}`
              : "No incident linked"}
            {" · "}
            {new Date(image.uploadedAt).toLocaleString()}
          </Typography>
        </Box>
        <Chip
          size="small"
          label={image.status}
          color={
            image.status === "approved"
              ? "success"
              : image.status === "rejected"
                ? "error"
                : "default"
          }
        />
      </Stack>

      {result ? (
        <>
          <Divider />

          <Stack spacing={1}>
            <Stack
              direction="row"
              justifyContent="space-between"
              alignItems="center"
            >
              <Typography variant="subtitle2">
                CV Labels ({result.labels.length})
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Model: {result.modelVersion}
              </Typography>
            </Stack>

            {result.labels.map((label) => (
              <Box key={label.label}>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                  spacing={1}
                >
                  <Chip
                    size="small"
                    label={label.label}
                    color={labelSeverityRole(label.label)}
                  />
                  <Typography variant="caption" fontWeight={600}>
                    {(label.confidence * 100).toFixed(0)}%
                  </Typography>
                </Stack>
                <LinearProgress
                  variant="determinate"
                  value={label.confidence * 100}
                  sx={{
                    height: 6,
                    borderRadius: 1,
                    backgroundColor: nadaaBrand.colors.slate + "20",
                    "& .MuiLinearProgress-bar": {
                      backgroundColor: confidenceColor(label.confidence),
                    },
                  }}
                />
              </Box>
            ))}
          </Stack>

          {result.humanReviewRequired && (
            <Alert severity="warning" icon={<AlertTriangle size={18} />}>
              Human review required: low confidence or sensitive content
              detected.
            </Alert>
          )}

          {!result.humanReviewRequired && result.reviewStatus === "pending" && (
            <Alert severity="info" icon={<CheckCircle2 size={18} />}>
              High-confidence result. Review still recommended before
              operational use.
            </Alert>
          )}

          <Alert severity="info" sx={{ fontSize: "0.75rem" }}>
            {result.limitations}
          </Alert>

          {result.reviewStatus === "pending" ? (
            <>
              <Divider />
              <Stack spacing={1}>
                <Typography variant="subtitle2">Human review</Typography>
                <Stack direction="row" spacing={1}>
                  <Button
                    size="small"
                    variant="contained"
                    color="success"
                    startIcon={<ThumbsUp size={16} />}
                    disabled={reviewBusy}
                    onClick={onApprove}
                  >
                    Approve
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    color="error"
                    startIcon={<ThumbsDown size={16} />}
                    disabled={reviewBusy}
                    onClick={onReject}
                  >
                    Reject
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    startIcon={<XCircle size={16} />}
                    disabled={reviewBusy}
                    onClick={onReanalyze}
                  >
                    Re-analyze
                  </Button>
                </Stack>
              </Stack>
            </>
          ) : (
            <Alert severity="success">
              Reviewed by {result.reviewedBy} as {result.reviewStatus}
              {result.reviewNote ? `: "${result.reviewNote}"` : ""}
            </Alert>
          )}
        </>
      ) : (
        <EmptyState
          title="No CV result"
          detail="Run CV analysis to see evidence labels."
        />
      )}
    </Stack>
  );
}
