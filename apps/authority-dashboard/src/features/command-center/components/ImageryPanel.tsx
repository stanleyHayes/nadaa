import {
  type ChangeEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  FormControlLabel,
  Grid,
  IconButton,
  MenuItem,
  Paper,
  Stack,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  CloudUpload,
  Eye,
  EyeOff,
  Loader2,
  RefreshCw,
  Satellite,
  Trash2,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CreateImageryRequest,
  ImageryGeoJSONFeatureCollection,
  ImageryListResponse,
  ImageryRecord,
  ImagerySource,
  ImageryStatus,
} from "@nadaa/shared-types";
import { IMAGERY_API_BASE } from "@/app/config";
import { authorityHeaders } from "@/app/session";
import { CommandSelect } from "./shared";

type LoadState = "loading" | "ready" | "fallback" | "error";

type ImageryFormState = CreateImageryRequest;

const sourceColors: Record<ImagerySource, string> = {
  drone: nadaaBrand.colors.gold,
  satellite: nadaaBrand.colors.navy,
  other: nadaaBrand.colors.slate,
};

const sourceLabels: Record<ImagerySource, string> = {
  drone: "Drone",
  satellite: "Satellite",
  other: "Other",
};

const defaultGeometry = JSON.stringify(
  {
    type: "Polygon",
    coordinates: [
      [
        [-0.22, 5.56],
        [-0.19, 5.56],
        [-0.19, 5.59],
        [-0.22, 5.59],
        [-0.22, 5.56],
      ],
    ],
  },
  null,
  2,
);

function buildDefaultForm(selectedIncidentId?: string): ImageryFormState {
  const now = new Date();
  now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
  return {
    source: "drone",
    captureTime: now.toISOString().slice(0, 16),
    geometry: defaultGeometry,
    coverageAreaKm2: "",
    resolutionMeters: "",
    license: "",
    relatedIncidentId: selectedIncidentId ?? "",
    relatedRiskZoneId: "",
    mlWorkflowId: "",
  };
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`;
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en-GH", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function ImageryPanel({
  selectedIncidentId,
  showOverlay,
  onToggleOverlay,
  onFeaturesChange,
}: {
  selectedIncidentId?: string;
  showOverlay: boolean;
  onToggleOverlay: (show: boolean) => void;
  onFeaturesChange: (
    features: ImageryGeoJSONFeatureCollection | undefined,
  ) => void;
}) {
  const [records, setRecords] = useState<ImageryRecord[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("");
  const [form, setForm] = useState(() => buildDefaultForm(selectedIncidentId));
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const refreshList = useCallback(async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("");
    try {
      const response = await fetch(`${IMAGERY_API_BASE}/imagery`, {
        headers: authorityHeaders(),
        signal,
      });
      if (!response.ok) {
        throw new Error(`imagery API returned ${response.status}`);
      }
      const payload = (await response.json()) as ImageryListResponse;
      setRecords(payload.imagery);
      setLoadState("ready");
      setFeedback("Imagery API connected.");
    } catch (error) {
      if (signal?.aborted) return;
      setRecords([]);
      setLoadState("fallback");
      setFeedback(
        "Imagery API unavailable. Upload and list actions need the service running.",
      );
    }
  }, []);

  const refreshGeoJSON = useCallback(
    async (signal?: AbortSignal) => {
      try {
        const response = await fetch(`${IMAGERY_API_BASE}/imagery/geojson`, {
          signal,
        });
        if (!response.ok) {
          throw new Error(`geojson API returned ${response.status}`);
        }
        const payload =
          (await response.json()) as ImageryGeoJSONFeatureCollection;
        onFeaturesChange(payload);
      } catch (error) {
        if (!signal?.aborted) {
          onFeaturesChange(undefined);
          setFeedback("Could not load imagery overlay.");
        }
      }
    },
    [onFeaturesChange],
  );

  useEffect(() => {
    const controller = new AbortController();
    void refreshList(controller.signal);
    return () => controller.abort();
  }, [refreshList]);

  const handleToggleOverlay = useCallback(
    async (_event: ChangeEvent<HTMLInputElement>, checked: boolean) => {
      onToggleOverlay(checked);
      if (checked) {
        await refreshGeoJSON();
      } else {
        onFeaturesChange(undefined);
      }
    },
    [onToggleOverlay, onFeaturesChange, refreshGeoJSON],
  );

  const updateForm =
    (key: keyof ImageryFormState) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const handleSourceChange = (event: SelectChangeEvent) => {
    setForm((current) => ({
      ...current,
      source: event.target.value as ImagerySource,
    }));
  };

  const handleUpload = async () => {
    if (!file) {
      setFeedback("Please select an image file to upload.");
      return;
    }
    if (file.size > 20 * 1024 * 1024) {
      setFeedback("File exceeds 20 MB upload limit.");
      return;
    }

    const body = new FormData();
    body.append("file", file);
    body.append("source", form.source);
    body.append("captureTime", new Date(form.captureTime).toISOString());
    body.append("geometry", form.geometry);
    body.append("coverageAreaKm2", form.coverageAreaKm2);
    body.append("resolutionMeters", form.resolutionMeters);
    if (form.license) body.append("license", form.license);
    if (form.relatedIncidentId)
      body.append("relatedIncidentId", form.relatedIncidentId);
    if (form.relatedRiskZoneId)
      body.append("relatedRiskZoneId", form.relatedRiskZoneId);
    if (form.mlWorkflowId) body.append("mlWorkflowId", form.mlWorkflowId);

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(`${IMAGERY_API_BASE}/imagery`, {
        method: "POST",
        headers: {
          "X-NADAA-Actor-ID": authorityHeaders()["X-NADAA-Actor-ID"],
          "X-NADAA-Actor-Role": authorityHeaders()["X-NADAA-Actor-Role"],
          "X-NADAA-Agency-ID": authorityHeaders()["X-NADAA-Agency-ID"],
          "X-NADAA-MFA-Completed": authorityHeaders()["X-NADAA-MFA-Completed"],
          "X-NADAA-Request-ID": authorityHeaders()["X-NADAA-Request-ID"],
        },
        body,
      });
      if (!response.ok) {
        throw new Error(`upload returned ${response.status}`);
      }
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
      setFeedback("Imagery uploaded successfully.");
      await refreshList();
      if (showOverlay) {
        await refreshGeoJSON();
      }
    } catch (error) {
      setFeedback(
        "Upload failed. Ensure the imagery-service is running and the file is an image under 20 MB.",
      );
    } finally {
      setBusy(false);
    }
  };

  const handleDownload = (record: ImageryRecord) => {
    window.open(`${IMAGERY_API_BASE}/imagery/${record.id}/download`, "_blank");
  };

  const handleExpire = async (record: ImageryRecord) => {
    try {
      const response = await fetch(
        `${IMAGERY_API_BASE}/imagery/${record.id}/expire`,
        {
          method: "POST",
          headers: authorityHeaders(),
        },
      );
      if (!response.ok) {
        throw new Error(`expire returned ${response.status}`);
      }
      setFeedback(`${record.reference} marked expired.`);
      await refreshList();
      if (showOverlay) await refreshGeoJSON();
    } catch (error) {
      setFeedback("Expire action failed.");
    }
  };

  const handleDelete = async (record: ImageryRecord) => {
    if (!window.confirm(`Delete ${record.reference} permanently?`)) return;
    try {
      const response = await fetch(`${IMAGERY_API_BASE}/imagery/${record.id}`, {
        method: "DELETE",
        headers: authorityHeaders(),
      });
      if (!response.ok) {
        throw new Error(`delete returned ${response.status}`);
      }
      setFeedback(`${record.reference} deleted.`);
      await refreshList();
      if (showOverlay) await refreshGeoJSON();
    } catch (error) {
      setFeedback("Delete action failed.");
    }
  };

  const handleRunLifecycle = async () => {
    try {
      const response = await fetch(
        `${IMAGERY_API_BASE}/imagery/lifecycle/run`,
        {
          method: "POST",
          headers: authorityHeaders(),
        },
      );
      if (!response.ok) {
        throw new Error(`lifecycle returned ${response.status}`);
      }
      const payload = (await response.json()) as { expiredCount: number };
      setFeedback(`Lifecycle run: ${payload.expiredCount} record(s) expired.`);
      await refreshList();
      if (showOverlay) await refreshGeoJSON();
    } catch (error) {
      setFeedback("Lifecycle run failed.");
    }
  };

  const activeRecords = records.filter((record) => record.status === "active");
  const expiredRecords = records.filter(
    (record) => record.status === "expired",
  );

  return (
    <Paper className="surface imagery-panel">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          alignItems: { xs: "stretch", sm: "center" }
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <Satellite size={21} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h6">Imagery ingestion</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Drone and satellite image footprints and lifecycle
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
          onClick={() => void refreshList()}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>
      {feedback ? (
        <Alert
          severity={
            feedback.includes("successfully") ||
            feedback.includes("connected") ||
            feedback.includes("Lifecycle run")
              ? "success"
              : "warning"
          }
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}
      <Stack spacing={1.5}>
        <FormControlLabel
          control={
            <Switch
              checked={showOverlay}
              onChange={handleToggleOverlay}
              slotProps={{ input: { "aria-label": "Toggle imagery overlay" } }}
            />
          }
          label={
            <Stack direction="row" spacing={0.5} sx={{
              alignItems: "center"
            }}>
              {showOverlay ? <Eye size={16} /> : <EyeOff size={16} />}
              <Typography variant="body2">
                {showOverlay ? "Overlay on map" : "Overlay hidden"}
              </Typography>
            </Stack>
          }
        />

        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <CommandSelect
              label="Source"
              value={form.source}
              onChange={handleSourceChange}
            >
              <MenuItem value="drone">Drone</MenuItem>
              <MenuItem value="satellite">Satellite</MenuItem>
              <MenuItem value="other">Other</MenuItem>
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Capture time"
              type="datetime-local"
              size="small"
              fullWidth
              value={form.captureTime}
              onChange={updateForm("captureTime")}
              slotProps={{
                inputLabel: { shrink: true }
              }}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4 }}>
            <TextField
              label="Area km²"
              size="small"
              fullWidth
              value={form.coverageAreaKm2}
              onChange={updateForm("coverageAreaKm2")}
              slotProps={{
                htmlInput: { inputMode: "decimal" }
              }}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4 }}>
            <TextField
              label="Resolution m"
              size="small"
              fullWidth
              value={form.resolutionMeters}
              onChange={updateForm("resolutionMeters")}
              slotProps={{
                htmlInput: { inputMode: "decimal" }
              }}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="License"
              size="small"
              fullWidth
              value={form.license}
              onChange={updateForm("license")}
              placeholder="e.g. CC-BY-4.0"
            />
          </Grid>
          <Grid size={{ xs: 12 }}>
            <TextField
              label="Footprint GeoJSON polygon"
              size="small"
              fullWidth
              multiline
              minRows={3}
              value={form.geometry}
              onChange={updateForm("geometry")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="Related incident"
              size="small"
              fullWidth
              value={form.relatedIncidentId}
              onChange={updateForm("relatedIncidentId")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="Related risk zone"
              size="small"
              fullWidth
              value={form.relatedRiskZoneId}
              onChange={updateForm("relatedRiskZoneId")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="ML workflow"
              size="small"
              fullWidth
              value={form.mlWorkflowId}
              onChange={updateForm("mlWorkflowId")}
            />
          </Grid>
        </Grid>

        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            style={{ display: "none" }}
            onChange={(event) => setFile(event.target.files?.[0] ?? null)}
          />
          <Button
            type="button"
            variant="outlined"
            size="small"
            startIcon={<CloudUpload size={16} />}
            onClick={() => fileInputRef.current?.click()}
          >
            {file ? file.name : "Select image"}
          </Button>
          {file ? (
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              {formatBytes(file.size)}
            </Typography>
          ) : null}
        </Stack>

        <Button
          type="button"
          variant="contained"
          startIcon={
            busy ? (
              <Loader2 size={17} className="spin-icon" />
            ) : (
              <CloudUpload size={17} />
            )
          }
          onClick={() => void handleUpload()}
          disabled={busy || !file}
        >
          {busy ? "Uploading" : "Upload imagery"}
        </Button>

        <Divider />

        <Stack
          direction="row"
          spacing={1}
          sx={{
            justifyContent: "space-between",
            alignItems: "center"
          }}>
          <Typography variant="subtitle2">
            {activeRecords.length} active · {expiredRecords.length} expired
          </Typography>
          <Button
            type="button"
            variant="outlined"
            size="small"
            onClick={() => void handleRunLifecycle()}
          >
            Run lifecycle
          </Button>
        </Stack>

        {records.length ? (
          <Box className="responsive-table-wrapper">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Ref</TableCell>
                  <TableCell>Source</TableCell>
                  <TableCell>Captured</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {records.map((record) => (
                  <TableRow key={record.id} hover>
                    <TableCell>
                      <Typography variant="subtitle2">
                        {record.reference}
                      </Typography>
                      <Typography variant="caption" sx={{
                        color: "text.secondary"
                      }}>
                        {formatBytes(record.sizeBytes)} ·{" "}
                        {record.resolutionMeters} m
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={sourceLabels[record.source]}
                        style={{
                          backgroundColor: sourceColors[record.source],
                          color: "#FFFFFF",
                        }}
                      />
                    </TableCell>
                    <TableCell>{formatDateTime(record.captureTime)}</TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={record.status}
                        color={
                          record.status === "active" ? "success" : "default"
                        }
                      />
                    </TableCell>
                    <TableCell align="right">
                      <Stack
                        direction="row"
                        spacing={0.5}
                        sx={{
                          justifyContent: "flex-end"
                        }}
                      >
                        <IconButton
                          size="small"
                          aria-label="Download"
                          onClick={() => handleDownload(record)}
                        >
                          <CloudUpload size={16} />
                        </IconButton>
                        {record.status === "active" ? (
                          <IconButton
                            size="small"
                            aria-label="Expire"
                            onClick={() => void handleExpire(record)}
                          >
                            <EyeOff size={16} />
                          </IconButton>
                        ) : null}
                        <IconButton
                          size="small"
                          aria-label="Delete"
                          onClick={() => void handleDelete(record)}
                        >
                          <Trash2 size={16} />
                        </IconButton>
                      </Stack>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        ) : (
          <Typography variant="body2" sx={{
            color: "text.secondary"
          }}>
            No imagery records available. Upload a drone or satellite image to
            begin.
          </Typography>
        )}
      </Stack>
    </Paper>
  );
}
