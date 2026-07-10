import { type ChangeEvent, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  MenuItem,
  Stack,
  TextField,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import {
  Eye,
  LifeBuoy,
  Loader2,
  Pencil,
  Plus,
  RefreshCw,
  Trash2,
  X,
} from "lucide-react";
import type {
  ReliefPointRecord,
  ReliefPointStatus,
  ReliefStockHistoryEntry,
} from "@nadaa/shared-types";
import type { IncidentLoadState, ReliefPointFormState } from "../types";
import { formatShortDate } from "../utils";
import { CommandSelect, EmptyState, Fact } from "./shared";
import { SectionCard } from "./primitives";
import { DataTable } from "./DataTable";

type FieldChange = (
  key: keyof ReliefPointFormState,
) => (
  event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
) => void;

function toTitle(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function reliefStatusColor(
  status: ReliefPointStatus,
): "success" | "warning" | "default" {
  if (status === "open") {
    return "success";
  }
  if (status === "closed") {
    return "default";
  }
  return "warning";
}

export function ReliefDistributionPanel({
  reliefPoints,
  reliefForm,
  reliefHistory,
  loadState,
  feedback,
  busy,
  canDelete,
  onUpdateForm,
  onRefresh,
  onStartCreate,
  onStartEdit,
  onSave,
  onDelete,
}: {
  reliefPoints: ReliefPointRecord[];
  reliefForm: ReliefPointFormState;
  reliefHistory: ReliefStockHistoryEntry[];
  loadState: IncidentLoadState;
  feedback: string;
  busy: boolean;
  canDelete: boolean;
  onUpdateForm: FieldChange;
  onRefresh: () => void;
  onStartCreate: () => void;
  onStartEdit: (point: ReliefPointRecord) => void;
  onSave: () => Promise<boolean>;
  onDelete: (point: ReliefPointRecord) => Promise<boolean>;
}) {
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));

  const [viewPoint, setViewPoint] = useState<ReliefPointRecord | undefined>();
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [deleteTarget, setDeleteTarget] = useState<
    ReliefPointRecord | undefined
  >();
  const [deleteError, setDeleteError] = useState("");

  const openCreate = () => {
    onStartCreate();
    setFormMode("create");
    setFormOpen(true);
  };

  const openEdit = (point: ReliefPointRecord) => {
    onStartEdit(point);
    setFormMode("edit");
    setFormOpen(true);
  };

  const openDelete = (point: ReliefPointRecord) => {
    setDeleteError("");
    setDeleteTarget(point);
  };

  const submitForm = async () => {
    const ok = await onSave();
    if (ok) {
      setFormOpen(false);
    }
  };

  const confirmDelete = async () => {
    if (!deleteTarget) {
      return;
    }
    const ok = await onDelete(deleteTarget);
    if (ok) {
      setDeleteTarget(undefined);
    } else {
      setDeleteError(
        "Delete failed. The shelter-service must be reachable and you need an admin session.",
      );
    }
  };

  return (
    <SectionCard
      title="Relief distribution"
      eyebrow="Points, stock & eligibility"
      icon={LifeBuoy}
      accent="gold"
      action={
        <Stack direction="row" spacing={1}>
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
            onClick={onRefresh}
            disabled={loadState === "loading"}
          >
            Refresh
          </Button>
          <Button
            type="button"
            variant="contained"
            size="small"
            startIcon={<Plus size={16} />}
            onClick={openCreate}
          >
            Add relief point
          </Button>
        </Stack>
      }
    >
      {feedback ? (
        <Alert
          severity={
            loadState === "ready"
              ? "success"
              : loadState === "error"
                ? "error"
                : "warning"
          }
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}
      <DataTable
        rows={reliefPoints}
        getRowKey={(point) => point.id}
        searchOf={(point) =>
          `${point.name} ${point.district} ${point.region} ${point.type}`
        }
        searchPlaceholder="Search relief points"
        filters={[
          {
            key: "status",
            label: "Status",
            options: Array.from(
              new Set(reliefPoints.map((point) => toTitle(point.status))),
            ),
            valueOf: (point) => toTitle(point.status),
          },
          {
            key: "type",
            label: "Type",
            options: Array.from(
              new Set(reliefPoints.map((point) => toTitle(point.type))),
            ),
            valueOf: (point) => toTitle(point.type),
          },
        ]}
        columns={[
          { key: "name", label: "Name", render: (point) => point.name },
          {
            key: "type",
            label: "Type",
            render: (point) => toTitle(point.type),
          },
          {
            key: "area",
            label: "Area",
            render: (point) => point.district || point.region,
          },
          {
            key: "status",
            label: "Status",
            render: (point) => (
              <Chip
                size="small"
                label={toTitle(point.status)}
                color={reliefStatusColor(point.status)}
              />
            ),
          },
          {
            key: "stock",
            label: "Stock lines",
            align: "right",
            render: (point) => point.stockCategories.length,
          },
        ]}
        rowActions={(point) => (
          <Stack direction="row" spacing={0.5} sx={{
            justifyContent: "flex-end"
          }}>
            <Tooltip title="View detail">
              <IconButton
                size="small"
                aria-label={`View ${point.name}`}
                onClick={() => setViewPoint(point)}
              >
                <Eye size={16} />
              </IconButton>
            </Tooltip>
            <Tooltip title="Edit">
              <IconButton
                size="small"
                aria-label={`Edit ${point.name}`}
                onClick={() => openEdit(point)}
              >
                <Pencil size={16} />
              </IconButton>
            </Tooltip>
            {canDelete ? (
              <Tooltip title="Delete">
                <IconButton
                  size="small"
                  color="error"
                  aria-label={`Delete ${point.name}`}
                  onClick={() => openDelete(point)}
                >
                  <Trash2 size={16} />
                </IconButton>
              </Tooltip>
            ) : null}
          </Stack>
        )}
        emptyState={
          loadState === "loading" ? (
            <Box sx={{ py: 3 }} />
          ) : (
            <EmptyState
              title="No relief points"
              detail="Publish a relief distribution point to make it visible to citizens."
            />
          )
        }
      />
      {/* View detail dialog (read-only) */}
      <Dialog
        open={Boolean(viewPoint)}
        onClose={() => setViewPoint(undefined)}
        maxWidth="sm"
        fullWidth
        fullScreen={fullScreen}
        aria-labelledby="relief-point-view-title"
      >
        <DialogTitle
          id="relief-point-view-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>{viewPoint?.name ?? "Relief point"}</span>
          <IconButton
            aria-label="Close relief point detail"
            onClick={() => setViewPoint(undefined)}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {viewPoint ? (
            <Stack spacing={1.75}>
              <Stack direction="row" spacing={1} sx={{
                flexWrap: "wrap"
              }}>
                <Chip
                  size="small"
                  label={toTitle(viewPoint.status)}
                  color={reliefStatusColor(viewPoint.status)}
                />
                <Chip
                  size="small"
                  variant="outlined"
                  label={toTitle(viewPoint.type)}
                />
                <Chip
                  size="small"
                  variant="outlined"
                  label={`${viewPoint.stockCategories.length} stock lines`}
                />
              </Stack>

              <Grid container spacing={1.5}>
                <Grid size={{ xs: 6 }}>
                  <Fact label="Region" value={viewPoint.region || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact label="District" value={viewPoint.district || "—"} />
                </Grid>
                <Grid size={{ xs: 12 }}>
                  <Fact label="Address" value={viewPoint.address || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact label="Contact" value={viewPoint.contact || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact
                    label="Hours"
                    value={viewPoint.operatingHours || "—"}
                  />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact label="Schedule" value={viewPoint.schedule || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact
                    label="Coordinates"
                    value={`${viewPoint.location.lat}, ${viewPoint.location.lng}`}
                  />
                </Grid>
              </Grid>

              {viewPoint.eligibility ? (
                <Box>
                  <Typography variant="caption" sx={{
                    color: "text.secondary"
                  }}>
                    Eligibility
                  </Typography>
                  <Typography variant="body2">
                    {viewPoint.eligibility}
                  </Typography>
                </Box>
              ) : null}

              {viewPoint.stockCategories.length ? (
                <Box>
                  <Typography variant="caption" sx={{
                    color: "text.secondary"
                  }}>
                    Stock lines
                  </Typography>
                  <Stack spacing={0.25}>
                    {viewPoint.stockCategories.map((category) => (
                      <Typography
                        variant="body2"
                        key={`${category.category}-${category.unit}`}
                      >
                        {category.category} · {category.quantity}{" "}
                        {category.unit}
                      </Typography>
                    ))}
                  </Stack>
                </Box>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setViewPoint(undefined)}>Close</Button>
          <Button
            variant="contained"
            startIcon={<Pencil size={16} />}
            onClick={() => {
              if (viewPoint) {
                openEdit(viewPoint);
                setViewPoint(undefined);
              }
            }}
          >
            Edit
          </Button>
        </DialogActions>
      </Dialog>
      {/* Add / Edit form dialog */}
      <Dialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        maxWidth="sm"
        fullWidth
        scroll="paper"
        fullScreen={fullScreen}
        aria-labelledby="relief-point-form-title"
      >
        <DialogTitle
          id="relief-point-form-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>
            {formMode === "create"
              ? "Add relief point"
              : `Edit ${reliefForm.name || "relief point"}`}
          </span>
          <IconButton
            aria-label="Close relief point form"
            onClick={() => setFormOpen(false)}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {feedback ? (
            <Alert severity="error" sx={{ mb: 1.5 }}>
              {feedback}
            </Alert>
          ) : null}
          <Grid container spacing={1.5} sx={{ pt: 0.5 }}>
            <Grid size={{ xs: 12, sm: 7 }}>
              <TextField
                label="Name"
                size="small"
                fullWidth
                required
                value={reliefForm.name}
                onChange={onUpdateForm("name")}
              />
            </Grid>
            <Grid size={{ xs: 6, sm: 5 }}>
              <CommandSelect
                label="Type"
                value={reliefForm.type}
                onChange={onUpdateForm("type")}
              >
                <MenuItem value="food">Food</MenuItem>
                <MenuItem value="water">Water</MenuItem>
                <MenuItem value="medical">Medical</MenuItem>
                <MenuItem value="hygiene">Hygiene</MenuItem>
                <MenuItem value="blankets">Blankets</MenuItem>
                <MenuItem value="cash">Cash</MenuItem>
                <MenuItem value="mixed">Mixed</MenuItem>
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 6, sm: 4 }}>
              <CommandSelect
                label="Status"
                value={reliefForm.status}
                onChange={onUpdateForm("status")}
              >
                <MenuItem value="open">Open</MenuItem>
                <MenuItem value="limited">Limited</MenuItem>
                <MenuItem value="paused">Paused</MenuItem>
                <MenuItem value="closed">Closed</MenuItem>
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 6, sm: 4 }}>
              <TextField
                label="Region"
                size="small"
                fullWidth
                value={reliefForm.region}
                onChange={onUpdateForm("region")}
              />
            </Grid>
            <Grid size={{ xs: 6, sm: 4 }}>
              <TextField
                label="District"
                size="small"
                fullWidth
                value={reliefForm.district}
                onChange={onUpdateForm("district")}
              />
            </Grid>
            <Grid size={{ xs: 12 }}>
              <TextField
                label="Address"
                size="small"
                fullWidth
                value={reliefForm.address}
                onChange={onUpdateForm("address")}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Latitude"
                size="small"
                fullWidth
                required
                value={reliefForm.latitude}
                onChange={onUpdateForm("latitude")}
                error={
                  Boolean(reliefForm.latitude) &&
                  !Number.isFinite(Number(reliefForm.latitude))
                }
                helperText={
                  Boolean(reliefForm.latitude) &&
                  !Number.isFinite(Number(reliefForm.latitude))
                    ? "Latitude must be a number"
                    : ""
                }
                slotProps={{
                  htmlInput: { inputMode: "decimal" }
                }}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Longitude"
                size="small"
                fullWidth
                required
                value={reliefForm.longitude}
                onChange={onUpdateForm("longitude")}
                error={
                  Boolean(reliefForm.longitude) &&
                  !Number.isFinite(Number(reliefForm.longitude))
                }
                helperText={
                  Boolean(reliefForm.longitude) &&
                  !Number.isFinite(Number(reliefForm.longitude))
                    ? "Longitude must be a number"
                    : ""
                }
                slotProps={{
                  htmlInput: { inputMode: "decimal" }
                }}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Contact"
                size="small"
                fullWidth
                value={reliefForm.contact}
                onChange={onUpdateForm("contact")}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Hours"
                size="small"
                fullWidth
                value={reliefForm.operatingHours}
                onChange={onUpdateForm("operatingHours")}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Schedule"
                size="small"
                fullWidth
                value={reliefForm.schedule}
                onChange={onUpdateForm("schedule")}
              />
            </Grid>
            <Grid size={12}>
              <TextField
                label="Eligibility"
                size="small"
                fullWidth
                multiline
                minRows={2}
                value={reliefForm.eligibility}
                onChange={onUpdateForm("eligibility")}
              />
            </Grid>
            <Grid size={12}>
              <TextField
                label="Stock lines"
                size="small"
                fullWidth
                multiline
                minRows={3}
                helperText="One per line: category, quantity, unit"
                value={reliefForm.stockCategories}
                onChange={onUpdateForm("stockCategories")}
              />
            </Grid>
          </Grid>

          {formMode === "edit" && reliefHistory.length ? (
            <Box className="relief-history" sx={{ mt: 1.5 }}>
              <Typography variant="caption" sx={{
                color: "text.secondary"
              }}>
                Recent stock history
              </Typography>
              {reliefHistory.slice(0, 2).map((entry) => (
                <Typography
                  variant="caption"
                  key={entry.id}
                  sx={{
                    color: "text.secondary"
                  }}
                >
                  {formatShortDate(entry.changedAt)} ·{" "}
                  {entry.stockCategories.length} stock lines · {entry.changedBy}
                </Typography>
              ))}
            </Box>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setFormOpen(false)} disabled={busy}>
            Cancel
          </Button>
          <Button
            variant="contained"
            startIcon={
              busy ? (
                <Loader2 size={16} className="spin-icon" />
              ) : (
                <LifeBuoy size={16} />
              )
            }
            onClick={() => void submitForm()}
            disabled={busy}
          >
            {formMode === "create" ? "Publish relief point" : "Save changes"}
          </Button>
        </DialogActions>
      </Dialog>
      {/* Delete confirm dialog */}
      <Dialog
        open={Boolean(deleteTarget)}
        onClose={() => setDeleteTarget(undefined)}
        maxWidth="xs"
        fullWidth
        aria-labelledby="relief-point-delete-title"
      >
        <DialogTitle id="relief-point-delete-title">
          Delete relief point
        </DialogTitle>
        <DialogContent>
          <Typography variant="body2">
            Remove <strong>{deleteTarget?.name}</strong> from relief
            distribution? Citizens will no longer see this point. This cannot be
            undone.
          </Typography>
          {deleteError ? (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {deleteError}
            </Alert>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(undefined)} disabled={busy}>
            Cancel
          </Button>
          <Button
            color="error"
            variant="contained"
            startIcon={
              busy ? (
                <Loader2 size={16} className="spin-icon" />
              ) : (
                <Trash2 size={16} />
              )
            }
            onClick={() => void confirmDelete()}
            disabled={busy}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </SectionCard>
  );
}
