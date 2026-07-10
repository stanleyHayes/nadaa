import { type FormEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import {
  Database,
  Download,
  FileSpreadsheet,
  Loader2,
  RefreshCw,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  OpenDataCategory,
  OpenDataDataset,
  OpenDataDatasetDetailResponse,
  OpenDataDatasetDownloadResponse,
  OpenDataDatasetListResponse,
  OpenDataPrivacyReviewStatus,
  OpenDataRequestResponse,
} from "@nadaa/shared-types";
import { OPEN_DATA_API_BASE } from "@/app/config";
import { extractAPIError } from "../utils";

const categoryOptions: { label: string; value: OpenDataCategory | "all" }[] = [
  { label: "All categories", value: "all" },
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road closure", value: "road_closure" },
  { label: "Weather", value: "weather" },
  { label: "Shelter", value: "shelter" },
  { label: "Incident", value: "incident" },
  { label: "Relief", value: "relief" },
  { label: "Risk", value: "risk" },
  { label: "Other", value: "other" },
];

const privacyTone: Record<
  OpenDataPrivacyReviewStatus,
  "warning" | "success" | "error"
> = {
  pending_review: "warning",
  approved: "success",
  rejected: "error",
};

const anonymizationTone: Record<
  OpenDataDataset["anonymizationLevel"],
  "success" | "info" | "warning" | "default"
> = {
  none: "warning",
  aggregated: "success",
  anonymized: "success",
  synthetic: "info",
};

type LoadState = "loading" | "ready" | "error";

interface RequestForm {
  name: string;
  organization: string;
  email: string;
  useCase: string;
  purpose: string;
}

const buildDefaultRequestForm = (): RequestForm => ({
  name: "",
  organization: "",
  email: "",
  useCase: "",
  purpose: "",
});

export function OpenDataPortal() {
  const [datasets, setDatasets] = useState<OpenDataDataset[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("Loading open data catalog");
  const [category, setCategory] = useState<OpenDataCategory | "all">("all");
  const [selectedDataset, setSelectedDataset] =
    useState<OpenDataDataset | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [requestOpen, setRequestOpen] = useState(false);
  const [requestForm, setRequestForm] = useState<RequestForm>(
    buildDefaultRequestForm(),
  );
  const [requestBusy, setRequestBusy] = useState(false);
  const [requestResult, setRequestResult] =
    useState<OpenDataRequestResponse | null>(null);
  const [downloadBusy, setDownloadBusy] = useState(false);
  const [downloadResult, setDownloadResult] =
    useState<OpenDataDatasetDownloadResponse | null>(null);

  const filteredDatasets = useMemo(() => {
    if (category === "all") return datasets;
    return datasets.filter((dataset) => dataset.category === category);
  }, [datasets, category]);

  const fetchDatasets = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading open data catalog");
    try {
      const response = await fetch(`${OPEN_DATA_API_BASE}/open-data/datasets`, {
        signal,
      });
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload = (await response.json()) as OpenDataDatasetListResponse;
      setDatasets(payload.datasets);
      setLoadState("ready");
    } catch (error) {
      if (signal?.aborted) return;
      setLoadState("error");
      setFeedback(
        error instanceof Error
          ? error.message
          : "Could not load open data catalog.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void fetchDatasets(controller.signal);
    return () => controller.abort();
  }, []);

  const openDetail = async (dataset: OpenDataDataset) => {
    setSelectedDataset(dataset);
    setDownloadResult(null);
    setRequestResult(null);
    setDetailOpen(true);
    try {
      const response = await fetch(
        `${OPEN_DATA_API_BASE}/open-data/datasets/${dataset.id}`,
      );
      if (response.ok) {
        const payload =
          (await response.json()) as OpenDataDatasetDetailResponse;
        setSelectedDataset(payload.dataset);
      }
    } catch {
      // keep the list version if detail fails
    }
  };

  const downloadDataset = async (dataset: OpenDataDataset, format: string) => {
    if (dataset.privacyReviewStatus !== "approved") {
      return;
    }
    setDownloadBusy(true);
    try {
      const response = await fetch(
        `${OPEN_DATA_API_BASE}/open-data/datasets/${dataset.id}/download?format=${encodeURIComponent(format)}`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload =
        (await response.json()) as OpenDataDatasetDownloadResponse;
      setDownloadResult(payload);
    } catch (error) {
      // Surface the real failure instead of fabricating a "ready" result.
      setDownloadResult(null);
      setFeedback(
        error instanceof Error && error.message
          ? `Download failed: ${error.message}`
          : "Download failed. Please try again.",
      );
    } finally {
      setDownloadBusy(false);
    }
  };

  const openRequest = (dataset: OpenDataDataset) => {
    setSelectedDataset(dataset);
    setRequestResult(null);
    setRequestForm(buildDefaultRequestForm());
    setRequestOpen(true);
  };

  const submitRequest = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selectedDataset) return;

    if (requestForm.name.length < 2) {
      setFeedback("Please enter your full name.");
      return;
    }
    if (requestForm.email.length < 5 || !requestForm.email.includes("@")) {
      setFeedback("Please enter a valid email address.");
      return;
    }
    if (requestForm.purpose.length < 10) {
      setFeedback("Please describe your purpose in at least 10 characters.");
      return;
    }

    setRequestBusy(true);
    try {
      const response = await fetch(`${OPEN_DATA_API_BASE}/open-data/requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          datasetId: selectedDataset.id,
          requesterInfo: {
            name: requestForm.name,
            organization: requestForm.organization,
            email: requestForm.email,
            useCase: requestForm.useCase,
          },
          purpose: requestForm.purpose,
        }),
      });
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }
      const payload = (await response.json()) as OpenDataRequestResponse;
      setRequestResult(payload);
      setRequestForm(buildDefaultRequestForm());
    } catch (error) {
      setFeedback(
        error instanceof Error
          ? error.message
          : "Could not submit access request.",
      );
    } finally {
      setRequestBusy(false);
    }
  };

  return (
    <Paper className="surface">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Database size={21} color={nadaaBrand.colors.navy} />
          <Box>
            <Typography variant="h6">Open Data Portal</Typography>
            <Typography variant="caption" color="text.secondary">
              Approved, anonymized disaster datasets
            </Typography>
          </Box>
        </Stack>
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
          onClick={() => void fetchDatasets()}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>

      <FormControl fullWidth size="small" sx={{ mb: 2 }}>
        <InputLabel id="open-data-category-label">Category</InputLabel>
        <Select
          labelId="open-data-category-label"
          value={category}
          label="Category"
          onChange={(event) =>
            setCategory(event.target.value as OpenDataCategory | "all")
          }
        >
          {categoryOptions.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {loadState === "error" ? (
        <Alert severity="warning" className="warning-alert">
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        {filteredDatasets.length > 0 ? (
          filteredDatasets.map((dataset) => (
            <Paper
              variant="outlined"
              className="shelter-row"
              key={dataset.id}
              sx={{ cursor: "pointer" }}
              onClick={() => openDetail(dataset)}
            >
              <Stack spacing={1}>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="flex-start"
                  spacing={1}
                >
                  <Box>
                    <Typography variant="subtitle2">{dataset.title}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {dataset.category.replace("_", " ")} ·{" "}
                      {dataset.updateFrequency.replace("_", " ")}
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    label={dataset.privacyReviewStatus.replace("_", " ")}
                    color={privacyTone[dataset.privacyReviewStatus]}
                  />
                </Stack>
                <Typography variant="body2">{dataset.description}</Typography>
                <Stack
                  direction="row"
                  spacing={0.75}
                  flexWrap="wrap"
                  alignItems="center"
                >
                  <Chip
                    size="small"
                    variant="outlined"
                    label={dataset.license}
                  />
                  <Chip
                    size="small"
                    variant="outlined"
                    label={dataset.anonymizationLevel}
                    color={anonymizationTone[dataset.anonymizationLevel]}
                  />
                  {dataset.privacyReviewStatus === "approved" ? (
                    <Button
                      size="small"
                      variant="outlined"
                      startIcon={<Download size={14} />}
                      onClick={(event) => {
                        event.stopPropagation();
                        void downloadDataset(dataset, "csv");
                      }}
                      disabled={downloadBusy}
                    >
                      Download
                    </Button>
                  ) : (
                    <Button
                      size="small"
                      variant="outlined"
                      startIcon={<FileSpreadsheet size={14} />}
                      onClick={(event) => {
                        event.stopPropagation();
                        openRequest(dataset);
                      }}
                    >
                      Request access
                    </Button>
                  )}
                </Stack>
              </Stack>
            </Paper>
          ))
        ) : (
          <Alert severity="info" className="warning-alert">
            {loadState === "loading"
              ? "Loading datasets..."
              : "No datasets match the selected category."}
          </Alert>
        )}
      </Stack>

      <Dialog
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>{selectedDataset?.title}</DialogTitle>
        <DialogContent>
          {selectedDataset ? (
            <Stack spacing={2}>
              <Typography variant="body1">
                {selectedDataset.description}
              </Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                <Chip
                  size="small"
                  label={selectedDataset.category.replace("_", " ")}
                />
                <Chip size="small" label={selectedDataset.license} />
                <Chip
                  size="small"
                  label={selectedDataset.privacyReviewStatus.replace("_", " ")}
                  color={privacyTone[selectedDataset.privacyReviewStatus]}
                />
                <Chip
                  size="small"
                  label={selectedDataset.anonymizationLevel}
                  color={anonymizationTone[selectedDataset.anonymizationLevel]}
                />
                <Chip
                  size="small"
                  label={selectedDataset.updateFrequency.replace("_", " ")}
                />
              </Stack>

              {selectedDataset.accessRestriction ? (
                <Alert severity="warning" className="warning-alert">
                  {selectedDataset.accessRestriction}
                </Alert>
              ) : null}

              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Metadata
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Publisher: {selectedDataset.metadata.publisher}
                </Typography>
                {selectedDataset.metadata.contactEmail ? (
                  <Typography variant="body2" color="text.secondary">
                    Contact: {selectedDataset.metadata.contactEmail}
                  </Typography>
                ) : null}
                {selectedDataset.metadata.regionCoverage?.length ? (
                  <Typography variant="body2" color="text.secondary">
                    Regions:{" "}
                    {selectedDataset.metadata.regionCoverage.join(", ")}
                  </Typography>
                ) : null}
                {selectedDataset.metadata.temporalCoverage ? (
                  <Typography variant="body2" color="text.secondary">
                    Temporal coverage:{" "}
                    {selectedDataset.metadata.temporalCoverage}
                  </Typography>
                ) : null}
                {selectedDataset.metadata.anonymizationNotes ? (
                  <Typography variant="body2" color="text.secondary">
                    Anonymization: {selectedDataset.metadata.anonymizationNotes}
                  </Typography>
                ) : null}
              </Box>

              {selectedDataset.columns && selectedDataset.columns.length > 0 ? (
                <Box>
                  <Typography variant="subtitle2" gutterBottom>
                    Columns
                  </Typography>
                  <TableContainer
                    component={Paper}
                    variant="outlined"
                    className="responsive-table"
                  >
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          <TableCell>Name</TableCell>
                          <TableCell>Type</TableCell>
                          <TableCell>Description</TableCell>
                          <TableCell>Nullable</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {selectedDataset.columns.map((column) => (
                          <TableRow key={column.name}>
                            <TableCell>{column.name}</TableCell>
                            <TableCell>{column.type}</TableCell>
                            <TableCell>{column.description ?? "—"}</TableCell>
                            <TableCell>
                              {column.nullable ? "Yes" : "No"}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Box>
              ) : null}

              {selectedDataset.sampleRows &&
              selectedDataset.sampleRows.length > 0 ? (
                <Box>
                  <Typography variant="subtitle2" gutterBottom>
                    Sample data
                  </Typography>
                  <TableContainer
                    component={Paper}
                    variant="outlined"
                    className="responsive-table"
                  >
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          {Object.keys(selectedDataset.sampleRows[0]).map(
                            (key) => (
                              <TableCell key={key}>{key}</TableCell>
                            ),
                          )}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {selectedDataset.sampleRows.map((row, index) => (
                          <TableRow key={index}>
                            {Object.values(row).map((value, vIndex) => (
                              <TableCell key={vIndex}>
                                {String(value ?? "—")}
                              </TableCell>
                            ))}
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Box>
              ) : null}

              {downloadResult ? (
                <Alert
                  severity={downloadResult.auditLogged ? "success" : "warning"}
                  className="warning-alert"
                >
                  Download {downloadResult.download.format.toUpperCase()} ready.
                  Rate limit remaining: {downloadResult.rateLimit.remaining} /{" "}
                  {downloadResult.rateLimit.limit}.
                </Alert>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          {selectedDataset?.privacyReviewStatus === "approved" ? (
            <Button
              startIcon={<Download size={16} />}
              onClick={() =>
                selectedDataset && void downloadDataset(selectedDataset, "csv")
              }
              disabled={downloadBusy}
            >
              Download CSV
            </Button>
          ) : (
            <Button
              startIcon={<FileSpreadsheet size={16} />}
              onClick={() => selectedDataset && openRequest(selectedDataset)}
            >
              Request access
            </Button>
          )}
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={requestOpen}
        onClose={() => setRequestOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Request dataset access</DialogTitle>
        <DialogContent>
          {requestResult ? (
            <Alert severity="success" className="warning-alert">
              Access request {requestResult.request.id} submitted. Status:{" "}
              {requestResult.request.status}.
            </Alert>
          ) : (
            <Stack
              component="form"
              id="open-data-request-form"
              spacing={2}
              onSubmit={submitRequest}
              noValidate
            >
              <Typography variant="body2" color="text.secondary">
                Requesting access to <strong>{selectedDataset?.title}</strong>.
              </Typography>
              <TextField
                label="Full name"
                value={requestForm.name}
                onChange={(event) =>
                  setRequestForm((current) => ({
                    ...current,
                    name: event.target.value,
                  }))
                }
                fullWidth
                required
              />
              <TextField
                label="Organization"
                value={requestForm.organization}
                onChange={(event) =>
                  setRequestForm((current) => ({
                    ...current,
                    organization: event.target.value,
                  }))
                }
                fullWidth
              />
              <TextField
                label="Email"
                type="email"
                value={requestForm.email}
                onChange={(event) =>
                  setRequestForm((current) => ({
                    ...current,
                    email: event.target.value,
                  }))
                }
                fullWidth
                required
              />
              <TextField
                label="Use case"
                value={requestForm.useCase}
                onChange={(event) =>
                  setRequestForm((current) => ({
                    ...current,
                    useCase: event.target.value,
                  }))
                }
                fullWidth
              />
              <TextField
                label="Purpose"
                value={requestForm.purpose}
                onChange={(event) =>
                  setRequestForm((current) => ({
                    ...current,
                    purpose: event.target.value,
                  }))
                }
                multiline
                minRows={3}
                fullWidth
                required
                helperText="Describe why you need this dataset (at least 10 characters)."
              />
            </Stack>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRequestOpen(false)}>Close</Button>
          {!requestResult ? (
            <Button
              type="submit"
              form="open-data-request-form"
              variant="contained"
              disabled={requestBusy}
            >
              {requestBusy ? "Submitting..." : "Submit request"}
            </Button>
          ) : null}
        </DialogActions>
      </Dialog>
    </Paper>
  );
}
