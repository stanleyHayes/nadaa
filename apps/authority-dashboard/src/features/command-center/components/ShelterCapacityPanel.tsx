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
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
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
  RefreshCw,
  Trash2,
  X,
} from "lucide-react";
import type { ShelterRecord, ShelterStatus } from "@nadaa/shared-types";
import type { IncidentLoadState, ShelterFormState } from "../types";
import { CommandSelect, EmptyState, Fact } from "./shared";
import { CapacityMeter, SectionCard } from "./primitives";

type FieldChange = (
  key: keyof ShelterFormState,
) => (
  event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
) => void;

function toTitle(value: string) {
  return value
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function shelterStatusColor(
  status: ShelterStatus,
): "success" | "warning" | "default" {
  if (status === "open") {
    return "success";
  }
  if (status === "full") {
    return "warning";
  }
  return "default";
}

export function ShelterCapacityPanel({
  shelters,
  shelterForm,
  loadState,
  feedback,
  busy,
  canDelete,
  onUpdateForm,
  onRefresh,
  onEdit,
  onSave,
  onDelete,
}: {
  shelters: ShelterRecord[];
  shelterForm: ShelterFormState;
  loadState: IncidentLoadState;
  feedback: string;
  busy: boolean;
  canDelete: boolean;
  onUpdateForm: FieldChange;
  onRefresh: () => void;
  onEdit: (shelter: ShelterRecord) => void;
  onSave: () => Promise<boolean>;
  onDelete: (shelter: ShelterRecord) => Promise<boolean>;
}) {
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));

  const [viewShelter, setViewShelter] = useState<ShelterRecord | undefined>();
  const [editShelter, setEditShelter] = useState<ShelterRecord | undefined>();
  const [formOpen, setFormOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ShelterRecord | undefined>();
  const [deleteError, setDeleteError] = useState("");

  const openEdit = (shelter: ShelterRecord) => {
    onEdit(shelter);
    setEditShelter(shelter);
    setFormOpen(true);
  };

  const openDelete = (shelter: ShelterRecord) => {
    setDeleteError("");
    setDeleteTarget(shelter);
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
      title="Shelter capacity"
      eyebrow="Occupancy & operating status"
      icon={LifeBuoy}
      accent="green"
      action={
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

      {shelters.length ? (
        <TableContainer component={Box} sx={{ overflowX: "auto" }}>
          <Table size="small" aria-label="Shelter capacity register">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Area</TableCell>
                <TableCell>Occupancy</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {shelters.map((shelter) => (
                <TableRow key={shelter.id} hover>
                  <TableCell>{shelter.name}</TableCell>
                  <TableCell>{shelter.district || shelter.region}</TableCell>
                  <TableCell>
                    {shelter.currentOccupancy}/{shelter.capacity}
                  </TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      label={toTitle(shelter.status)}
                      color={shelterStatusColor(shelter.status)}
                    />
                  </TableCell>
                  <TableCell align="right">
                    <Stack
                      direction="row"
                      spacing={0.5}
                      justifyContent="flex-end"
                    >
                      <Tooltip title="View detail">
                        <IconButton
                          size="small"
                          aria-label={`View ${shelter.name}`}
                          onClick={() => setViewShelter(shelter)}
                        >
                          <Eye size={16} />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit capacity">
                        <IconButton
                          size="small"
                          aria-label={`Edit ${shelter.name}`}
                          onClick={() => openEdit(shelter)}
                        >
                          <Pencil size={16} />
                        </IconButton>
                      </Tooltip>
                      {canDelete ? (
                        <Tooltip title="Delete">
                          <IconButton
                            size="small"
                            color="error"
                            aria-label={`Delete ${shelter.name}`}
                            onClick={() => openDelete(shelter)}
                          >
                            <Trash2 size={16} />
                          </IconButton>
                        </Tooltip>
                      ) : null}
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : loadState === "loading" ? null : (
        <EmptyState
          title="No shelters"
          detail="No shelters are currently registered for this area."
        />
      )}

      {/* View detail dialog (read-only) */}
      <Dialog
        open={Boolean(viewShelter)}
        onClose={() => setViewShelter(undefined)}
        maxWidth="sm"
        fullWidth
        fullScreen={fullScreen}
        aria-labelledby="shelter-view-title"
      >
        <DialogTitle
          id="shelter-view-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>{viewShelter?.name ?? "Shelter"}</span>
          <IconButton
            aria-label="Close shelter detail"
            onClick={() => setViewShelter(undefined)}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {viewShelter ? (
            <Stack spacing={1.75}>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                <Chip
                  size="small"
                  label={toTitle(viewShelter.status)}
                  color={shelterStatusColor(viewShelter.status)}
                />
                <Chip
                  size="small"
                  variant="outlined"
                  label={`${viewShelter.currentOccupancy}/${viewShelter.capacity} occupied`}
                />
                <Chip
                  size="small"
                  variant="outlined"
                  label={toTitle(viewShelter.type)}
                />
              </Stack>

              <CapacityMeter
                value={viewShelter.currentOccupancy}
                max={viewShelter.capacity}
                tone={viewShelter.status === "full" ? "red" : "green"}
              />

              <Grid container spacing={1.5}>
                <Grid size={{ xs: 6 }}>
                  <Fact label="Region" value={viewShelter.region || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact label="District" value={viewShelter.district || "—"} />
                </Grid>
                <Grid size={{ xs: 12 }}>
                  <Fact label="Address" value={viewShelter.address || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact label="Contact" value={viewShelter.contact || "—"} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Fact
                    label="Coordinates"
                    value={`${viewShelter.location.lat}, ${viewShelter.location.lng}`}
                  />
                </Grid>
              </Grid>

              {viewShelter.facilities.length ? (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Facilities
                  </Typography>
                  <Stack direction="row" spacing={0.75} flexWrap="wrap">
                    {viewShelter.facilities.map((facility) => (
                      <Chip
                        size="small"
                        variant="outlined"
                        label={facility}
                        key={facility}
                      />
                    ))}
                  </Stack>
                </Box>
              ) : null}

              {viewShelter.notes ? (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    Operational note
                  </Typography>
                  <Typography variant="body2">{viewShelter.notes}</Typography>
                </Box>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setViewShelter(undefined)}>Close</Button>
          <Button
            variant="contained"
            startIcon={<Pencil size={16} />}
            onClick={() => {
              if (viewShelter) {
                openEdit(viewShelter);
                setViewShelter(undefined);
              }
            }}
          >
            Edit capacity
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit capacity dialog */}
      <Dialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        maxWidth="sm"
        fullWidth
        fullScreen={fullScreen}
        aria-labelledby="shelter-form-title"
      >
        <DialogTitle
          id="shelter-form-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>Edit {editShelter?.name ?? "shelter"} capacity</span>
          <IconButton
            aria-label="Close shelter form"
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
            <Grid size={6}>
              <TextField
                label="Capacity"
                size="small"
                fullWidth
                required
                value={shelterForm.capacity}
                onChange={onUpdateForm("capacity")}
                inputProps={{ inputMode: "numeric" }}
                error={
                  Boolean(shelterForm.capacity) &&
                  !Number.isFinite(Number(shelterForm.capacity))
                }
                helperText={
                  Boolean(shelterForm.capacity) &&
                  !Number.isFinite(Number(shelterForm.capacity))
                    ? "Capacity must be a number"
                    : ""
                }
              />
            </Grid>
            <Grid size={6}>
              <TextField
                label="Occupancy"
                size="small"
                fullWidth
                required
                value={shelterForm.currentOccupancy}
                onChange={onUpdateForm("currentOccupancy")}
                inputProps={{ inputMode: "numeric" }}
                error={
                  Boolean(shelterForm.currentOccupancy) &&
                  !Number.isFinite(Number(shelterForm.currentOccupancy))
                }
                helperText={
                  Boolean(shelterForm.currentOccupancy) &&
                  !Number.isFinite(Number(shelterForm.currentOccupancy))
                    ? "Occupancy must be a number"
                    : ""
                }
              />
            </Grid>
            <Grid size={12}>
              <CommandSelect
                label="Status"
                value={shelterForm.status}
                onChange={onUpdateForm("status")}
              >
                <MenuItem value="open">Open</MenuItem>
                <MenuItem value="full">Full</MenuItem>
                <MenuItem value="closed">Closed</MenuItem>
                <MenuItem value="unknown">Unknown</MenuItem>
              </CommandSelect>
            </Grid>
            <Grid size={12}>
              <TextField
                label="Operational note"
                size="small"
                fullWidth
                multiline
                minRows={2}
                value={shelterForm.notes}
                onChange={onUpdateForm("notes")}
              />
            </Grid>
          </Grid>
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
            Save changes
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete confirm dialog */}
      <Dialog
        open={Boolean(deleteTarget)}
        onClose={() => setDeleteTarget(undefined)}
        maxWidth="xs"
        fullWidth
        aria-labelledby="shelter-delete-title"
      >
        <DialogTitle id="shelter-delete-title">Delete shelter</DialogTitle>
        <DialogContent>
          <Typography variant="body2">
            Remove <strong>{deleteTarget?.name}</strong> from the shelter
            register? Citizens will no longer see this shelter. This cannot be
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
