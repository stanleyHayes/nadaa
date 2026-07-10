import { useEffect, useMemo, useState } from "react";
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
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { Megaphone, Plus, RefreshCw } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  Campaign,
  CampaignContentBlock,
  CampaignMetric,
  CampaignStatus,
  CampaignTemplate,
  CreateCampaignRequest,
} from "@nadaa/shared-types";
import { CAMPAIGN_API_BASE } from "@/app/config";
import { authorityHeaders } from "@/app/session";

type LoadState = "idle" | "loading" | "ready" | "error";

const hazardOptions: { label: string; value: string }[] = [
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Disease outbreak", value: "disease_outbreak" },
  { label: "Other", value: "other" },
];

const statusOptions: { label: string; value: CampaignStatus | "all" }[] = [
  { label: "All", value: "all" },
  { label: "Draft", value: "draft" },
  { label: "Published", value: "published" },
  { label: "Archived", value: "archived" },
];

function formatDateTimeLocal(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export default function CampaignManagerPanel() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [templates, setTemplates] = useState<CampaignTemplate[]>([]);
  const [metricsMap, setMetricsMap] = useState<
    Record<string, CampaignMetric[]>
  >({});
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [message, setMessage] = useState("");
  const [statusFilter, setStatusFilter] = useState<CampaignStatus | "all">(
    "all",
  );
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCampaign, setEditingCampaign] = useState<Campaign | null>(null);
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState<CreateCampaignRequest>(buildDefaultForm());

  const filteredCampaigns = useMemo(() => {
    if (statusFilter === "all") return campaigns;
    return campaigns.filter((c) => c.status === statusFilter);
  }, [campaigns, statusFilter]);

  const fetchCampaigns = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setMessage("Loading campaigns");
    try {
      const response = await fetch(`${CAMPAIGN_API_BASE}/campaigns`, {
        headers: authorityHeaders(),
        signal,
      });
      if (!response.ok)
        throw new Error(`campaign API returned ${response.status}`);
      const payload = (await response.json()) as { campaigns: Campaign[] };
      setCampaigns(payload.campaigns ?? []);
      setLoadState("ready");
      setMessage("Campaign API connected.");
    } catch (error) {
      if (signal?.aborted) return;
      setCampaigns([]);
      setLoadState("error");
      setMessage(
        error instanceof Error ? error.message : "Could not load campaigns.",
      );
    }
  };

  const fetchTemplates = async (signal?: AbortSignal) => {
    try {
      const response = await fetch(`${CAMPAIGN_API_BASE}/campaign-templates`, {
        signal,
      });
      if (!response.ok)
        throw new Error(`template API returned ${response.status}`);
      const payload = (await response.json()) as {
        templates: CampaignTemplate[];
      };
      setTemplates(payload.templates ?? []);
    } catch {
      setTemplates([]);
    }
  };

  const fetchMetrics = async (campaignId: string) => {
    try {
      const response = await fetch(
        `${CAMPAIGN_API_BASE}/campaigns/${campaignId}/metrics`,
      );
      if (!response.ok) return;
      const payload = (await response.json()) as { metrics: CampaignMetric[] };
      setMetricsMap((current) => ({
        ...current,
        [campaignId]: payload.metrics ?? [],
      }));
    } catch {
      // ignored
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void fetchCampaigns(controller.signal);
    void fetchTemplates(controller.signal);
    return () => controller.abort();
  }, []);

  useEffect(() => {
    for (const campaign of campaigns.slice(0, 10)) {
      void fetchMetrics(campaign.id);
    }
  }, [campaigns]);

  const openNew = () => {
    setEditingCampaign(null);
    setForm(buildDefaultForm());
    setDialogOpen(true);
  };

  const openEdit = (campaign: Campaign) => {
    setEditingCampaign(campaign);
    setForm({
      title: campaign.title,
      hazardType: campaign.hazardType,
      targetRegions: campaign.targetRegions,
      languages: campaign.languages,
      contentBlocks: campaign.contentBlocks,
      publishWindow: campaign.publishWindow,
      status: campaign.status,
      linkedGuideIds: campaign.linkedGuideIds ?? [],
      linkedAlertIds: campaign.linkedAlertIds ?? [],
    });
    setDialogOpen(true);
  };

  const applyTemplate = (templateId: string) => {
    const template = templates.find((t) => t.id === templateId);
    if (!template) return;
    setForm((current) => ({
      ...current,
      hazardType: template.hazardType as CreateCampaignRequest["hazardType"],
      contentBlocks: template.defaultContent,
      title: current.title || template.name,
    }));
  };

  const updateForm = <Key extends keyof CreateCampaignRequest>(
    key: Key,
    value: CreateCampaignRequest[Key],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const updateBlock = (index: number, patch: Partial<CampaignContentBlock>) => {
    setForm((current) => {
      const next = [...current.contentBlocks];
      next[index] = { ...next[index], ...patch };
      return { ...current, contentBlocks: next };
    });
  };

  const addBlock = () => {
    setForm((current) => ({
      ...current,
      contentBlocks: [
        ...current.contentBlocks,
        { type: "article", title: "", body: "" },
      ],
    }));
  };

  const removeBlock = (index: number) => {
    setForm((current) => ({
      ...current,
      contentBlocks: current.contentBlocks.filter((_, i) => i !== index),
    }));
  };

  const saveCampaign = async () => {
    setBusy(true);
    try {
      const url = editingCampaign
        ? `${CAMPAIGN_API_BASE}/campaigns/${editingCampaign.id}`
        : `${CAMPAIGN_API_BASE}/campaigns`;
      const method = editingCampaign ? "PUT" : "POST";
      const response = await fetch(url, {
        method,
        headers: authorityHeaders(),
        body: JSON.stringify(form),
      });
      if (!response.ok) {
        const body = await response.text();
        throw new Error(`save failed: ${response.status} ${body}`);
      }
      setDialogOpen(false);
      await fetchCampaigns();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Save failed.");
    } finally {
      setBusy(false);
    }
  };

  const totals = useMemo(() => {
    let reach = 0;
    let engagement = 0;
    for (const metrics of Object.values(metricsMap)) {
      for (const metric of metrics) {
        reach += metric.reach;
        engagement += metric.engagement;
      }
    }
    return { reach, engagement };
  }, [metricsMap]);

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
          <Megaphone size={21} color={nadaaBrand.colors.gold} />
          <Box>
            <Typography variant="h6">Campaign management</Typography>
            <Typography variant="caption" color="text.secondary">
              Publish preparedness campaigns by region, hazard, and language
            </Typography>
          </Box>
        </Stack>
        <Stack direction="row" spacing={1}>
          <Button
            type="button"
            variant="outlined"
            size="small"
            startIcon={
              loadState === "loading" ? (
                <RefreshCw size={16} className="spin-icon" />
              ) : (
                <RefreshCw size={16} />
              )
            }
            onClick={() => void fetchCampaigns()}
            disabled={loadState === "loading"}
          >
            Refresh
          </Button>
          <Button
            type="button"
            variant="contained"
            size="small"
            startIcon={<Plus size={16} />}
            onClick={openNew}
          >
            New campaign
          </Button>
        </Stack>
      </Stack>

      {loadState === "error" ? (
        <Alert severity="warning" className="warning-alert">
          {message}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid size={{ xs: 12, sm: 4 }}>
          <Paper variant="outlined" className="metric-card">
            <Typography variant="body2" color="text.secondary">
              Total reach
            </Typography>
            <Typography variant="h5">
              {totals.reach.toLocaleString()}
            </Typography>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, sm: 4 }}>
          <Paper variant="outlined" className="metric-card">
            <Typography variant="body2" color="text.secondary">
              Total engagement
            </Typography>
            <Typography variant="h5">
              {totals.engagement.toLocaleString()}
            </Typography>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, sm: 4 }}>
          <Paper variant="outlined" className="metric-card">
            <Typography variant="body2" color="text.secondary">
              Active campaigns
            </Typography>
            <Typography variant="h5">
              {campaigns.filter((c) => c.status === "published").length}
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1.5 }}>
        <Typography variant="body2" color="text.secondary">
          Filter by status:
        </Typography>
        <FormControl size="small" sx={{ minWidth: 140 }}>
          <InputLabel id="campaign-status-filter-label">Status</InputLabel>
          <Select
            labelId="campaign-status-filter-label"
            label="Status"
            value={statusFilter}
            onChange={(event) =>
              setStatusFilter(event.target.value as CampaignStatus | "all")
            }
          >
            {statusOptions.map((option) => (
              <MenuItem key={option.value} value={option.value}>
                {option.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Stack>

      {filteredCampaigns.length ? (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Title</TableCell>
              <TableCell>Hazard</TableCell>
              <TableCell>Regions</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Reach</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredCampaigns.map((campaign) => {
              const metrics = metricsMap[campaign.id] ?? [];
              const reach = metrics.reduce((sum, m) => sum + m.reach, 0);
              return (
                <TableRow key={campaign.id} hover>
                  <TableCell>
                    <Typography variant="subtitle2">
                      {campaign.title}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {campaign.languages.join(", ")}
                    </Typography>
                  </TableCell>
                  <TableCell>{campaign.hazardType}</TableCell>
                  <TableCell>{campaign.targetRegions.join(", ")}</TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      label={campaign.status}
                      color={
                        campaign.status === "published"
                          ? "success"
                          : campaign.status === "draft"
                            ? "default"
                            : "warning"
                      }
                    />
                  </TableCell>
                  <TableCell>{reach.toLocaleString()}</TableCell>
                  <TableCell>
                    <Button size="small" onClick={() => openEdit(campaign)}>
                      Edit
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      ) : (
        <Alert severity="info" className="warning-alert">
          No campaigns match this filter.
        </Alert>
      )}

      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          {editingCampaign ? "Edit campaign" : "Create campaign"}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 0.5 }}>
            <FormControl fullWidth size="small">
              <InputLabel id="campaign-template-label">Template</InputLabel>
              <Select
                labelId="campaign-template-label"
                label="Template"
                value=""
                onChange={(event) => applyTemplate(event.target.value)}
              >
                <MenuItem value="">None</MenuItem>
                {templates.map((template) => (
                  <MenuItem key={template.id} value={template.id}>
                    {template.name} ({template.season})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Title"
              value={form.title}
              onChange={(event) => updateForm("title", event.target.value)}
              fullWidth
            />

            <FormControl fullWidth size="small">
              <InputLabel id="campaign-hazard-label">Hazard type</InputLabel>
              <Select
                labelId="campaign-hazard-label"
                label="Hazard type"
                value={form.hazardType}
                onChange={(event) =>
                  updateForm(
                    "hazardType",
                    event.target.value as CreateCampaignRequest["hazardType"],
                  )
                }
              >
                {hazardOptions.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Target regions (comma separated)"
              value={form.targetRegions.join(", ")}
              onChange={(event) =>
                updateForm(
                  "targetRegions",
                  event.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean),
                )
              }
              fullWidth
            />

            <TextField
              label="Languages (comma separated)"
              value={form.languages.join(", ")}
              onChange={(event) =>
                updateForm(
                  "languages",
                  event.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean),
                )
              }
              fullWidth
            />

            <Stack direction="row" spacing={2}>
              <TextField
                label="Publish from"
                type="datetime-local"
                value={formatDateTimeLocal(form.publishWindow.startsAt)}
                onChange={(event) =>
                  updateForm("publishWindow", {
                    ...form.publishWindow,
                    startsAt: new Date(event.target.value).toISOString(),
                  })
                }
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
              <TextField
                label="Publish until"
                type="datetime-local"
                value={formatDateTimeLocal(form.publishWindow.endsAt)}
                onChange={(event) =>
                  updateForm("publishWindow", {
                    ...form.publishWindow,
                    endsAt: new Date(event.target.value).toISOString(),
                  })
                }
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            </Stack>

            <FormControl fullWidth size="small">
              <InputLabel id="campaign-status-label">Status</InputLabel>
              <Select
                labelId="campaign-status-label"
                label="Status"
                value={form.status}
                onChange={(event) =>
                  updateForm("status", event.target.value as CampaignStatus)
                }
              >
                {statusOptions
                  .filter((option) => option.value !== "all")
                  .map((option) => (
                    <MenuItem key={option.value} value={option.value}>
                      {option.label}
                    </MenuItem>
                  ))}
              </Select>
            </FormControl>

            <TextField
              label="Linked guide IDs (comma separated)"
              value={(form.linkedGuideIds ?? []).join(", ")}
              onChange={(event) =>
                updateForm(
                  "linkedGuideIds",
                  event.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean),
                )
              }
              fullWidth
            />

            <TextField
              label="Linked alert IDs (comma separated)"
              value={(form.linkedAlertIds ?? []).join(", ")}
              onChange={(event) =>
                updateForm(
                  "linkedAlertIds",
                  event.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean),
                )
              }
              fullWidth
            />

            <Typography variant="subtitle2">Content blocks</Typography>
            {form.contentBlocks.map((block, index) => (
              <Paper key={index} variant="outlined" sx={{ p: 1.5 }}>
                <Stack spacing={1.5}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <FormControl size="small" sx={{ minWidth: 120 }}>
                      <InputLabel id={`block-type-${index}-label`}>
                        Type
                      </InputLabel>
                      <Select
                        labelId={`block-type-${index}-label`}
                        label="Type"
                        value={block.type}
                        onChange={(event) =>
                          updateBlock(index, {
                            type: event.target
                              .value as CampaignContentBlock["type"],
                          })
                        }
                      >
                        <MenuItem value="article">Article</MenuItem>
                        <MenuItem value="checklist">Checklist</MenuItem>
                        <MenuItem value="media">Media</MenuItem>
                      </Select>
                    </FormControl>
                    <TextField
                      label="Block title"
                      value={block.title}
                      onChange={(event) =>
                        updateBlock(index, { title: event.target.value })
                      }
                      fullWidth
                    />
                    <Button color="error" onClick={() => removeBlock(index)}>
                      Remove
                    </Button>
                  </Stack>
                  {block.type === "checklist" ? (
                    <TextField
                      label="Checklist items (one per line)"
                      value={(block.items ?? []).join("\n")}
                      onChange={(event) =>
                        updateBlock(index, {
                          items: event.target.value
                            .split("\n")
                            .map((s) => s.trim())
                            .filter(Boolean),
                        })
                      }
                      multiline
                      minRows={3}
                      fullWidth
                    />
                  ) : (
                    <TextField
                      label={
                        block.type === "media"
                          ? "Media URL / description"
                          : "Body"
                      }
                      value={block.body ?? ""}
                      onChange={(event) =>
                        updateBlock(index, { body: event.target.value })
                      }
                      multiline
                      minRows={3}
                      fullWidth
                    />
                  )}
                </Stack>
              </Paper>
            ))}
            <Button
              variant="outlined"
              onClick={addBlock}
              startIcon={<Plus size={16} />}
            >
              Add content block
            </Button>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={() => void saveCampaign()}
            disabled={busy}
          >
            {busy ? "Saving" : editingCampaign ? "Update" : "Create"}
          </Button>
        </DialogActions>
      </Dialog>
    </Paper>
  );
}

function buildDefaultForm(): CreateCampaignRequest {
  const now = new Date();
  const startsAt = new Date(now.getTime() + 60 * 60 * 1000).toISOString();
  const endsAt = new Date(
    now.getTime() + 30 * 24 * 60 * 60 * 1000,
  ).toISOString();
  return {
    title: "",
    hazardType: "flood",
    targetRegions: [],
    languages: ["en"],
    contentBlocks: [{ type: "article", title: "", body: "" }],
    publishWindow: { startsAt, endsAt },
    status: "draft",
    linkedGuideIds: [],
    linkedAlertIds: [],
  };
}
