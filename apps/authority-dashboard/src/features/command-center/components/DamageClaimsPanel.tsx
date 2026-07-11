import { type ChangeEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogContent,
  DialogTitle,
  Divider,
  Grid,
  IconButton,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  CheckCheck,
  Eye,
  FileDown,
  FileText,
  Loader2,
  RefreshCw,
  ShieldAlert,
  X,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import { DAMAGE_CLAIM_API_BASE } from "@/app/config";
import { authorityHeaders } from "@/app/session";
import { DataTable } from "./DataTable";
import { CommandSelect, EmptyState, Fact, SkeletonRows } from "./shared";

type DamageClaimStatus = "draft" | "submitted" | "closed";
type DamageClaimVerificationStatus = "pending" | "verified" | "rejected";

interface DamageClaimReporter {
  name: string;
  phone?: string;
  email?: string;
  userId?: string;
}

interface DamageClaimLocation {
  lat: number;
  lng: number;
  address?: string;
}

interface DamageClaimRecord {
  id: string;
  reference: string;
  incidentId?: string;
  incidentReference?: string;
  incidentLocation?: string;
  reporter: DamageClaimReporter;
  damageType: string;
  damageDescription: string;
  estimatedLossAmount: string;
  damagePhotos: string[];
  location: DamageClaimLocation;
  verificationStatus: DamageClaimVerificationStatus;
  verifiedBy?: string;
  verifiedAt?: string;
  verificationNotes?: string;
  status: DamageClaimStatus;
  privacyConsent: boolean;
  createdAt: string;
  updatedAt: string;
}

interface DamageClaimListResponse {
  claims: DamageClaimRecord[];
  generatedAt: string;
}

type LoadState = "loading" | "ready" | "empty" | "error";

interface ClaimFilters {
  verificationStatus: "all" | DamageClaimVerificationStatus;
  status: "all" | DamageClaimStatus;
  incidentId: string;
  query: string;
}

const defaultFilters: ClaimFilters = {
  verificationStatus: "all",
  status: "all",
  incidentId: "",
  query: "",
};

export default function DamageClaimsPanel() {
  const [claims, setClaims] = useState<DamageClaimRecord[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading damage claims");
  const [filters, setFilters] = useState<ClaimFilters>(defaultFilters);
  const [selectedId, setSelectedId] = useState<string>("");
  const [busy, setBusy] = useState(false);
  const [exportBusy, setExportBusy] = useState<"csv" | "pdf" | null>(null);
  const [feedback, setFeedback] = useState("");
  const [verifyStatus, setVerifyStatus] =
    useState<DamageClaimVerificationStatus>("verified");
  const [verifyNotes, setVerifyNotes] = useState("");
  const [closeReason, setCloseReason] = useState("");

  const [detailOpen, setDetailOpen] = useState(false);
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));

  const selectedClaim = useMemo(
    () => claims.find((claim) => claim.id === selectedId),
    [claims, selectedId],
  );

  const openClaim = (claimId: string) => {
    setSelectedId(claimId);
    setDetailOpen(true);
  };

  const closeDetail = () => setDetailOpen(false);

  useEffect(() => {
    const controller = new AbortController();
    void refreshClaims(controller.signal);
    return () => controller.abort();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters]);

  useEffect(() => {
    if (claims.length && !claims.some((claim) => claim.id === selectedId)) {
      setSelectedId(claims[0]!.id);
    }
  }, [claims, selectedId]);

  useEffect(() => {
    setVerifyStatus("verified");
    setVerifyNotes("");
    setCloseReason("");
    setFeedback("");
  }, [selectedId]);

  const buildQueryString = () => {
    const params = new URLSearchParams();
    if (filters.verificationStatus !== "all") {
      params.set("verificationStatus", filters.verificationStatus);
    }
    if (filters.status !== "all") {
      params.set("status", filters.status);
    }
    if (filters.incidentId.trim()) {
      params.set("incidentId", filters.incidentId.trim());
    }
    if (filters.query.trim()) {
      params.set("q", filters.query.trim());
    }
    const query = params.toString();
    return query ? `?${query}` : "";
  };

  const refreshClaims = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading damage claims");

    try {
      const response = await fetch(
        `${DAMAGE_CLAIM_API_BASE}/claims${buildQueryString()}`,
        {
          headers: authorityHeaders(),
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`damage claim API returned ${response.status}`);
      }

      const payload = (await response.json()) as DamageClaimListResponse;
      setClaims(payload.claims);
      setLoadState(payload.claims.length ? "ready" : "empty");
      setLoadMessage(
        payload.claims.length
          ? "Damage claim API connected."
          : "No damage claims match the current filters.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setClaims([]);
      setLoadState("error");
      setLoadMessage(
        "Damage claim API unavailable. Check that damage-claim-service is running.",
      );
    }
  };

  const updateFilter =
    (key: keyof ClaimFilters) => (event: SelectChangeEvent) => {
      setFilters((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateTextFilter =
    (key: "incidentId" | "query") =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setFilters((current) => ({ ...current, [key]: event.target.value }));
    };

  const applyClaimUpdate = (updated: DamageClaimRecord) => {
    setClaims((current) =>
      current.map((claim) => (claim.id === updated.id ? updated : claim)),
    );
  };

  const verifySelectedClaim = async () => {
    if (!selectedClaim) {
      return;
    }

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DAMAGE_CLAIM_API_BASE}/claims/${selectedClaim.id}/verify`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify({
            verificationStatus: verifyStatus,
            notes: verifyNotes.trim(),
          }),
        },
      );
      if (!response.ok) {
        throw new Error(`damage claim API returned ${response.status}`);
      }
      const updated = (await response.json()) as DamageClaimRecord;
      applyClaimUpdate(updated);
      setFeedback(
        `Claim ${updated.reference} marked as ${updated.verificationStatus}.`,
      );
    } catch (error) {
      setFeedback(
        "Verification needs the damage-claim-service API running with a pending claim.",
      );
    } finally {
      setBusy(false);
    }
  };

  const closeSelectedClaim = async () => {
    if (!selectedClaim) {
      return;
    }

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DAMAGE_CLAIM_API_BASE}/claims/${selectedClaim.id}/close`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify({ reason: closeReason.trim() }),
        },
      );
      if (!response.ok) {
        throw new Error(`damage claim API returned ${response.status}`);
      }
      const updated = (await response.json()) as DamageClaimRecord;
      applyClaimUpdate(updated);
      setFeedback(`Claim ${updated.reference} closed.`);
    } catch (error) {
      setFeedback(
        "Close action needs the damage-claim-service API running with a valid claim.",
      );
    } finally {
      setBusy(false);
    }
  };

  const exportSelectedClaim = async (format: "csv" | "pdf") => {
    if (!selectedClaim) {
      return;
    }

    setExportBusy(format);
    setFeedback("");
    try {
      const response = await fetch(
        `${DAMAGE_CLAIM_API_BASE}/claims/${selectedClaim.id}/export?format=${format}`,
        {
          headers: authorityHeaders(),
        },
      );
      if (!response.ok) {
        throw new Error(`damage claim API returned ${response.status}`);
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);

      if (format === "pdf") {
        window.open(url, "_blank", "noopener,noreferrer");
        setTimeout(() => URL.revokeObjectURL(url), 30_000);
      } else {
        const anchor = document.createElement("a");
        anchor.href = url;
        const contentDisposition = response.headers.get("content-disposition");
        let filename = `claim_${selectedClaim.reference}.csv`;
        if (contentDisposition) {
          const match = /filename="?([^"]+)"?/.exec(contentDisposition);
          if (match?.[1]) {
            filename = match[1];
          }
        }
        anchor.download = filename;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        setTimeout(() => URL.revokeObjectURL(url), 5_000);
      }
    } catch (error) {
      setFeedback(
        `Export failed. ${error instanceof Error ? error.message : "Unknown error"}`,
      );
    } finally {
      setExportBusy(null);
    }
  };

  return (
    <Paper className="surface damage-claims-panel">
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
          <ShieldAlert size={21} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h6">Damage claim review</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Verify, close and export insurance damage claims.
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
          onClick={() => void refreshClaims()}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>
      <Grid container spacing={1.5}>
        <Grid size={{ xs: 12, sm: 6 }}>
          <CommandSelect
            label="Verification"
            value={filters.verificationStatus}
            onChange={updateFilter("verificationStatus")}
          >
            <MenuItem value="all">All verification statuses</MenuItem>
            <MenuItem value="pending">Pending</MenuItem>
            <MenuItem value="verified">Verified</MenuItem>
            <MenuItem value="rejected">Rejected</MenuItem>
          </CommandSelect>
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <CommandSelect
            label="Claim status"
            value={filters.status}
            onChange={updateFilter("status")}
          >
            <MenuItem value="all">All statuses</MenuItem>
            <MenuItem value="draft">Draft</MenuItem>
            <MenuItem value="submitted">Submitted</MenuItem>
            <MenuItem value="closed">Closed</MenuItem>
          </CommandSelect>
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label="Incident ID"
            size="small"
            fullWidth
            value={filters.incidentId}
            onChange={updateTextFilter("incidentId")}
            slotProps={{
              htmlInput: { "aria-label": "Filter by incident ID" }
            }}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label="Search"
            size="small"
            fullWidth
            value={filters.query}
            onChange={updateTextFilter("query")}
            slotProps={{
              htmlInput: {
                "aria-label":
                  "Search claims by reference, reporter, type or address",
              }
            }}
          />
        </Grid>
      </Grid>
      {loadMessage && loadState !== "loading" ? (
        <Alert
          severity={
            loadState === "error"
              ? "error"
              : loadState === "empty"
                ? "info"
                : "success"
          }
          className="feed-alert"
        >
          {loadMessage}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <SkeletonRows />
      ) : null}
      {claims.length ? (
        <>
          <DataTable
            rows={claims}
            getRowKey={(claim) => claim.id}
            filters={[
              {
                // Verification, Status and Search are already handled server-side
                // by the toolbar above; only Category (damage type) has no server
                // control, so keep just that one here to avoid duplicate/conflicting filters.
                key: "damageType",
                label: "Category",
                options: Array.from(
                  new Set(claims.map((claim) => claim.damageType)),
                ),
                valueOf: (claim) => claim.damageType,
              },
            ]}
            columns={[
              {
                key: "reference",
                label: "Reference",
                render: (claim) => (
                  <>
                    <Typography variant="subtitle2">
                      {claim.reference}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{ color: "text.secondary" }}
                    >
                      {claim.incidentReference || "No incident"}
                    </Typography>
                  </>
                ),
              },
              {
                key: "reporter",
                label: "Reporter",
                render: (claim) => claim.reporter.name,
              },
              {
                key: "damageType",
                label: "Type",
                render: (claim) => claim.damageType,
              },
              {
                key: "verificationStatus",
                label: "Verification",
                render: (claim) => (
                  <Chip
                    size="small"
                    label={claim.verificationStatus}
                    color={
                      claim.verificationStatus === "pending"
                        ? "warning"
                        : claim.verificationStatus === "verified"
                          ? "success"
                          : "error"
                    }
                  />
                ),
              },
              {
                key: "status",
                label: "Status",
                render: (claim) => (
                  <Chip
                    size="small"
                    label={claim.status}
                    variant="outlined"
                    color={claim.status === "closed" ? "default" : "primary"}
                  />
                ),
              },
            ]}
            rowActions={(claim) => (
              <Tooltip title="Review claim">
                <IconButton
                  size="small"
                  aria-label={`Review claim ${claim.reference}`}
                  onClick={() => openClaim(claim.id)}
                >
                  <Eye size={16} />
                </IconButton>
              </Tooltip>
            )}
            rowActionsLabel="Actions"
            pageSize={8}
            emptyState={
              <EmptyState
                title="No claims"
                detail="No damage claims match the search or filters."
              />
            }
          />

          <Dialog
            open={detailOpen}
            onClose={closeDetail}
            maxWidth="md"
            fullWidth
            scroll="paper"
            fullScreen={fullScreen}
            aria-labelledby="damage-claim-detail-title"
          >
            <DialogTitle
              id="damage-claim-detail-title"
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 1,
              }}
            >
              <span>{selectedClaim?.reference ?? "Damage claim"}</span>
              <IconButton
                aria-label="Close claim detail"
                onClick={closeDetail}
                size="small"
              >
                <X size={18} />
              </IconButton>
            </DialogTitle>
            <DialogContent dividers>
              {selectedClaim ? (
                <Stack spacing={2}>
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    sx={{
                      justifyContent: "space-between",
                      alignItems: { xs: "flex-start", sm: "center" },
                      gap: 1
                    }}>
                    <Box>
                      <Typography variant="overline" color="secondary">
                        Selected claim
                      </Typography>
                      <Typography variant="h6">
                        {selectedClaim.reference}
                      </Typography>
                    </Box>
                    <Stack direction="row" spacing={1} sx={{
                      flexWrap: "wrap"
                    }}>
                      <Chip
                        size="small"
                        label={selectedClaim.verificationStatus}
                        color={
                          selectedClaim.verificationStatus === "pending"
                            ? "warning"
                            : selectedClaim.verificationStatus === "verified"
                              ? "success"
                              : "error"
                        }
                      />
                      <Chip
                        size="small"
                        label={selectedClaim.status}
                        variant="outlined"
                        color={
                          selectedClaim.status === "closed"
                            ? "default"
                            : "primary"
                        }
                      />
                    </Stack>
                  </Stack>

                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    {selectedClaim.damageDescription}
                  </Typography>

                  <Grid container spacing={1.5}>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <Fact
                        label="Reporter"
                        value={selectedClaim.reporter.name}
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <Fact
                        label="Phone"
                        value={selectedClaim.reporter.phone || "Not provided"}
                      />
                    </Grid>
                    {selectedClaim.reporter.email ? (
                      <Grid size={{ xs: 12, sm: 6 }}>
                        <Fact
                          label="Email"
                          value={selectedClaim.reporter.email}
                        />
                      </Grid>
                    ) : null}
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <Fact
                        label="Estimated loss"
                        value={selectedClaim.estimatedLossAmount}
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <Fact
                        label="Damage type"
                        value={selectedClaim.damageType}
                      />
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                      <Fact
                        label="Location"
                        value={
                          selectedClaim.location.address
                            ? `${selectedClaim.location.address} (${selectedClaim.location.lat.toFixed(5)}, ${selectedClaim.location.lng.toFixed(5)})`
                            : `${selectedClaim.location.lat.toFixed(5)}, ${selectedClaim.location.lng.toFixed(5)}`
                        }
                      />
                    </Grid>
                    {selectedClaim.incidentReference ? (
                      <Grid size={{ xs: 12 }}>
                        <Fact
                          label="Incident"
                          value={selectedClaim.incidentReference}
                        />
                      </Grid>
                    ) : null}
                  </Grid>

                  {selectedClaim.damagePhotos.length ? (
                    <Box>
                      <Typography variant="subtitle2" gutterBottom>
                        Photos
                      </Typography>
                      <Stack direction="row" spacing={1} sx={{
                        flexWrap: "wrap"
                      }}>
                        {selectedClaim.damagePhotos.map((photo, index) => (
                          <Box
                            key={index}
                            component="img"
                            src={photo}
                            alt={`Damage photo ${index + 1} for ${selectedClaim.reference}`}
                            sx={{
                              width: 96,
                              height: 96,
                              objectFit: "cover",
                              borderRadius: 1,
                              border: "1px solid rgba(13, 27, 61, 0.08)",
                            }}
                          />
                        ))}
                      </Stack>
                    </Box>
                  ) : (
                    <Typography variant="body2" sx={{
                      color: "text.secondary"
                    }}>
                      No photos attached.
                    </Typography>
                  )}

                  <Divider />

                  {feedback ? (
                    <Alert
                      severity={
                        feedback.includes("failed") ||
                        feedback.includes("needs") ||
                        feedback.includes("Unavailable") ||
                        feedback.includes("valid")
                          ? "warning"
                          : "success"
                      }
                    >
                      {feedback}
                    </Alert>
                  ) : null}

                  {selectedClaim.verificationStatus === "pending" ? (
                    <Stack spacing={1.5}>
                      <Typography variant="subtitle2">Verify claim</Typography>
                      <CommandSelect
                        label="Decision"
                        value={verifyStatus}
                        onChange={(event) =>
                          setVerifyStatus(
                            event.target.value as DamageClaimVerificationStatus,
                          )
                        }
                      >
                        <MenuItem value="verified">Approve (verified)</MenuItem>
                        <MenuItem value="rejected">Reject</MenuItem>
                      </CommandSelect>
                      <TextField
                        label="Verification notes"
                        size="small"
                        fullWidth
                        multiline
                        minRows={2}
                        value={verifyNotes}
                        onChange={(event) => setVerifyNotes(event.target.value)}
                      />
                      <Button
                        type="button"
                        variant="contained"
                        color="success"
                        disabled={busy || !verifyNotes.trim()}
                        onClick={() => void verifySelectedClaim()}
                        startIcon={<CheckCheck size={17} />}
                        aria-label="Save verification decision"
                      >
                        {busy ? "Saving" : "Save verification"}
                      </Button>
                    </Stack>
                  ) : (
                    <Alert severity="info">
                      This claim has been {selectedClaim.verificationStatus}.
                    </Alert>
                  )}

                  {selectedClaim.status !== "closed" ? (
                    <Stack spacing={1.5}>
                      <Typography variant="subtitle2">Close claim</Typography>
                      <TextField
                        label="Close reason"
                        size="small"
                        fullWidth
                        multiline
                        minRows={2}
                        value={closeReason}
                        onChange={(event) => setCloseReason(event.target.value)}
                      />
                      <Button
                        type="button"
                        variant="outlined"
                        color="error"
                        disabled={busy || !closeReason.trim()}
                        onClick={() => void closeSelectedClaim()}
                        startIcon={<ShieldAlert size={17} />}
                        aria-label="Close claim with reason"
                      >
                        {busy ? "Closing" : "Close claim"}
                      </Button>
                    </Stack>
                  ) : (
                    <Alert severity="info">This claim is closed.</Alert>
                  )}

                  <Stack direction="row" spacing={1} sx={{
                    flexWrap: "wrap"
                  }}>
                    <Button
                      type="button"
                      variant="outlined"
                      size="small"
                      disabled={Boolean(exportBusy)}
                      onClick={() => void exportSelectedClaim("csv")}
                      startIcon={<FileDown size={17} />}
                      aria-label="Export claim as CSV"
                    >
                      {exportBusy === "csv" ? "Exporting…" : "Export CSV"}
                    </Button>
                    <Button
                      type="button"
                      variant="outlined"
                      size="small"
                      disabled={Boolean(exportBusy)}
                      onClick={() => void exportSelectedClaim("pdf")}
                      startIcon={<FileText size={17} />}
                      aria-label="Export claim as PDF"
                    >
                      {exportBusy === "pdf" ? "Exporting…" : "Export PDF"}
                    </Button>
                  </Stack>
                </Stack>
              ) : null}
            </DialogContent>
          </Dialog>
        </>
      ) : loadState !== "loading" ? (
        <EmptyState
          title={loadState === "error" ? "Claims unavailable" : "No claims"}
          detail={
            loadState === "error"
              ? "Damage claim service is unavailable. Refresh to retry."
              : "No damage claims match these filters."
          }
        />
      ) : null}
    </Paper>
  );
}
