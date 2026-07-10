import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControl,
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
import { HandHeart } from "lucide-react";

import type { AidPledgeRecord, AidRequestRecord } from "@nadaa/shared-types";
import { aidRequestCategoryOptions, aidRequestPriorityOptions } from "../data";
import {
  aidLabel,
  aidPriorityColor,
  aidProgressPercent,
  aidStatusColor,
} from "../utils";
import type { AidRequestFormState } from "../types";

import { EmptyState } from "./shared";

export function AidRequestCard({
  onSelect,
  request,
  selected,
}: {
  onSelect: () => void;
  request: AidRequestRecord;
  selected?: boolean;
}) {
  const progress = aidProgressPercent(request);

  return (
    <Card
      onClick={onSelect}
      sx={{
        borderColor: selected ? "primary.main" : undefined,
        cursor: "pointer",
      }}
      variant="outlined"
    >
      <CardContent>
        <Stack direction="row" spacing={2} sx={{
          justifyContent: "space-between"
        }}>
          <Box>
            <Typography sx={{
              fontWeight: 700
            }}>{request.title}</Typography>
            <Typography variant="body2" sx={{
              color: "text.secondary"
            }}>
              {aidLabel(request.category)} · {request.district}
            </Typography>
          </Box>
          <Stack spacing={1} sx={{
            alignItems: "flex-end"
          }}>
            <Chip
              color={aidStatusColor(request.status)}
              label={aidLabel(request.status)}
              size="small"
            />
            <Chip
              color={aidPriorityColor(request.priority)}
              label={aidLabel(request.priority)}
              size="small"
              variant="outlined"
            />
          </Stack>
        </Stack>

        <Typography variant="body2" sx={{
          mt: 1
        }}>
          {request.quantityPledged.toLocaleString("en-GH")} /{" "}
          {request.quantityNeeded.toLocaleString("en-GH")}{" "}
          {request.quantityUnit} pledged
        </Typography>
        <LinearProgress sx={{ mt: 1 }} value={progress} variant="determinate" />
        <Stack
          direction="row"
          sx={{
            flexWrap: "wrap",
            gap: 1,
            mt: 1.5
          }}>
          <Chip
            icon={<HandHeart size={14} />}
            label={`${request.pledges.length} pledges`}
            size="small"
            variant="outlined"
          />
          <Chip
            label={`Needed by ${new Date(request.neededBy).toLocaleDateString("en-GH")}`}
            size="small"
            variant="outlined"
          />
        </Stack>
      </CardContent>
    </Card>
  );
}

export function AidRequestForm({
  form,
  onChange,
  onSubmit,
  submitLabel,
}: {
  form: AidRequestFormState;
  onChange: (form: AidRequestFormState) => void;
  onSubmit: () => void;
  submitLabel: string;
}) {
  return (
    <Stack spacing={2}>
      <TextField
        error={!form.title.trim()}
        helperText={!form.title.trim() ? "Title is required" : ""}
        label="Title"
        onChange={(event) => onChange({ ...form, title: event.target.value })}
        required
        size="small"
        value={form.title}
      />
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <FormControl fullWidth size="small">
          <InputLabel>Category</InputLabel>
          <Select
            label="Category"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                category: event.target.value as typeof form.category,
              })
            }
            value={form.category}
          >
            {aidRequestCategoryOptions.map((category) => (
              <MenuItem key={category} value={category}>
                {aidLabel(category)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl fullWidth size="small">
          <InputLabel>Priority</InputLabel>
          <Select
            label="Priority"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                priority: event.target.value as typeof form.priority,
              })
            }
            value={form.priority}
          >
            {aidRequestPriorityOptions.map((priority) => (
              <MenuItem key={priority} value={priority}>
                {aidLabel(priority)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Region"
          onChange={(event) =>
            onChange({ ...form, region: event.target.value })
          }
          size="small"
          value={form.region}
        />
        <TextField
          fullWidth
          label="District"
          onChange={(event) =>
            onChange({ ...form, district: event.target.value })
          }
          size="small"
          value={form.district}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Latitude"
          onChange={(event) => onChange({ ...form, lat: event.target.value })}
          size="small"
          type="number"
          value={form.lat}
        />
        <TextField
          fullWidth
          label="Longitude"
          onChange={(event) => onChange({ ...form, lng: event.target.value })}
          size="small"
          type="number"
          value={form.lng}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          error={!form.receivingOrganization.trim()}
          fullWidth
          helperText={
            !form.receivingOrganization.trim()
              ? "Receiving organization is required"
              : ""
          }
          label="Receiving organization"
          onChange={(event) =>
            onChange({ ...form, receivingOrganization: event.target.value })
          }
          required
          size="small"
          value={form.receivingOrganization}
        />
        <TextField
          fullWidth
          label="Contact"
          onChange={(event) =>
            onChange({ ...form, contact: event.target.value })
          }
          size="small"
          value={form.contact}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          error={!form.quantityNeeded.trim()}
          fullWidth
          helperText={
            !form.quantityNeeded.trim() ? "Quantity needed is required" : ""
          }
          label="Quantity needed"
          onChange={(event) =>
            onChange({ ...form, quantityNeeded: event.target.value })
          }
          required
          size="small"
          type="number"
          value={form.quantityNeeded}
        />
        <TextField
          fullWidth
          label="Unit"
          onChange={(event) =>
            onChange({ ...form, quantityUnit: event.target.value })
          }
          size="small"
          value={form.quantityUnit}
        />
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField
          fullWidth
          label="Needed by"
          onChange={(event) =>
            onChange({ ...form, neededBy: event.target.value })
          }
          size="small"
          type="datetime-local"
          value={form.neededBy}
          slotProps={{
            inputLabel: { shrink: true }
          }}
        />
        <FormControl fullWidth size="small">
          <InputLabel>Visibility</InputLabel>
          <Select
            label="Visibility"
            onChange={(event: SelectChangeEvent) =>
              onChange({
                ...form,
                visibility: event.target.value as typeof form.visibility,
              })
            }
            value={form.visibility}
          >
            <MenuItem value="public">Public</MenuItem>
            <MenuItem value="partners_only">Partners only</MenuItem>
          </Select>
        </FormControl>
      </Stack>
      <TextField
        label="Source relief point ID"
        onChange={(event) =>
          onChange({ ...form, sourceReliefPointId: event.target.value })
        }
        size="small"
        value={form.sourceReliefPointId}
      />
      <TextField
        error={!form.description.trim()}
        helperText={!form.description.trim() ? "Description is required" : ""}
        label="Description"
        multiline
        onChange={(event) =>
          onChange({ ...form, description: event.target.value })
        }
        required
        rows={3}
        size="small"
        value={form.description}
      />
      <Button
        disabled={
          !form.title.trim() ||
          !form.receivingOrganization.trim() ||
          !form.quantityNeeded.trim() ||
          !form.description.trim()
        }
        onClick={onSubmit}
        variant="contained"
      >
        {submitLabel}
      </Button>
    </Stack>
  );
}

export function AidPledgeList({ pledges }: { pledges: AidPledgeRecord[] }) {
  if (!pledges.length) {
    return <EmptyState message="No partner pledges recorded yet." />;
  }

  return (
    <Stack spacing={1.25}>
      {pledges.map((pledge) => (
        <Paper key={pledge.id} sx={{ p: 2 }} variant="outlined">
          <Stack direction="row" spacing={2} sx={{
            justifyContent: "space-between"
          }}>
            <Box>
              <Typography sx={{
                fontWeight: 700
              }}>{pledge.donorName}</Typography>
              <Typography variant="body2" sx={{
                color: "text.secondary"
              }}>
                {aidLabel(pledge.donorType)} ·{" "}
                {new Date(pledge.pledgedAt).toLocaleString("en-GH")}
              </Typography>
            </Box>
            <Stack spacing={1} sx={{
              alignItems: "flex-end"
            }}>
              <Chip label={aidLabel(pledge.status)} size="small" />
              <Chip
                color={
                  pledge.reviewStatus === "flagged"
                    ? "error"
                    : pledge.reviewStatus === "cleared"
                      ? "success"
                      : "warning"
                }
                label={aidLabel(pledge.reviewStatus)}
                size="small"
                variant="outlined"
              />
            </Stack>
          </Stack>
          <Typography variant="body2" sx={{
            mt: 1
          }}>
            {pledge.quantity.toLocaleString("en-GH")} {pledge.unit}
          </Typography>
          {pledge.note ? (
            <Typography
              variant="body2"
              sx={{
                color: "text.secondary",
                mt: 0.5
              }}>
              {pledge.note}
            </Typography>
          ) : null}
          {pledge.fraudReviewNotes ? (
            <Alert severity="info" sx={{ mt: 1 }}>
              {pledge.fraudReviewNotes}
            </Alert>
          ) : null}
        </Paper>
      ))}
    </Stack>
  );
}
