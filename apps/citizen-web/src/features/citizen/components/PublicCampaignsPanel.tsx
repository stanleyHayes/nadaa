import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
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
  Typography,
} from "@mui/material";
import { BookOpen, CheckCircle2, Megaphone, RefreshCw } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { Campaign, CampaignContentBlock } from "@nadaa/shared-types";
import { CAMPAIGN_API_BASE } from "@/app/config";
import { hazardOptions } from "../data";

type LoadState = "idle" | "loading" | "ready" | "error";

const regionOptions = [
  { label: "All regions", value: "all" },
  { label: "Greater Accra", value: "Greater Accra" },
  { label: "Ashanti", value: "Ashanti" },
  { label: "Western", value: "Western" },
  { label: "Northern", value: "Northern" },
  { label: "Upper East", value: "Upper East" },
  { label: "Upper West", value: "Upper West" },
];

function hazardLabel(value: string) {
  return hazardOptions.find((option) => option.value === value)?.label ?? value;
}

export default function PublicCampaignsPanel() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [message, setMessage] = useState("");
  const [regionFilter, setRegionFilter] = useState("all");
  const [hazardFilter, setHazardFilter] = useState("all");
  const [selectedCampaign, setSelectedCampaign] = useState<Campaign | null>(
    null,
  );

  const filteredCampaigns = useMemo(() => {
    return campaigns.filter((campaign) => {
      const regionMatch =
        regionFilter === "all" ||
        campaign.targetRegions.some(
          (region) => region.toLowerCase() === regionFilter.toLowerCase(),
        );
      const hazardMatch =
        hazardFilter === "all" || campaign.hazardType === hazardFilter;
      return regionMatch && hazardMatch;
    });
  }, [campaigns, regionFilter, hazardFilter]);

  const fetchCampaigns = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setMessage("Loading preparedness campaigns");
    try {
      const response = await fetch(`${CAMPAIGN_API_BASE}/campaigns`, {
        signal,
      });
      if (!response.ok)
        throw new Error(`campaign API returned ${response.status}`);
      const payload = (await response.json()) as { campaigns: Campaign[] };
      setCampaigns(payload.campaigns ?? []);
      setLoadState("ready");
      setMessage("Preparedness campaigns loaded.");
    } catch (error) {
      if (signal?.aborted) return;
      setCampaigns([]);
      setLoadState("error");
      setMessage(
        error instanceof Error ? error.message : "Could not load campaigns.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void fetchCampaigns(controller.signal);
    return () => controller.abort();
  }, []);

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
            <Typography variant="h6">Preparedness campaigns</Typography>
            <Typography variant="caption" color="text.secondary">
              Official guidance for your region and hazard
            </Typography>
          </Box>
        </Stack>
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
      </Stack>

      {loadState === "error" ? (
        <Alert severity="warning" className="warning-alert">
          {message}
        </Alert>
      ) : null}
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      <Grid
        container
        spacing={1.25}
        className="guide-filter-grid"
        sx={{ mb: 2 }}
      >
        <Grid size={{ xs: 12, sm: 6 }}>
          <FormControl fullWidth size="small">
            <InputLabel id="campaign-region-label">Region</InputLabel>
            <Select
              labelId="campaign-region-label"
              label="Region"
              value={regionFilter}
              onChange={(event) => setRegionFilter(event.target.value)}
            >
              {regionOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <FormControl fullWidth size="small">
            <InputLabel id="campaign-hazard-label">Hazard</InputLabel>
            <Select
              labelId="campaign-hazard-label"
              label="Hazard"
              value={hazardFilter}
              onChange={(event) => setHazardFilter(event.target.value)}
            >
              <MenuItem value="all">All hazards</MenuItem>
              {hazardOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
      </Grid>

      {filteredCampaigns.length ? (
        <Stack spacing={1.5}>
          {filteredCampaigns.map((campaign) => (
            <Paper
              key={campaign.id}
              variant="outlined"
              className="guide-list-row"
            >
              <Stack spacing={1}>
                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  justifyContent="space-between"
                  flexWrap="wrap"
                >
                  <Typography variant="subtitle1">{campaign.title}</Typography>
                  <Chip size="small" label={hazardLabel(campaign.hazardType)} />
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {campaign.targetRegions.join(" · ")} ·{" "}
                  {campaign.languages.join(", ")}
                </Typography>
                <Button
                  size="small"
                  variant="outlined"
                  onClick={() => setSelectedCampaign(campaign)}
                  startIcon={<BookOpen size={16} />}
                >
                  Read campaign
                </Button>
              </Stack>
            </Paper>
          ))}
        </Stack>
      ) : (
        <Alert severity="info" className="warning-alert">
          No active preparedness campaigns match your filters.
        </Alert>
      )}

      <Dialog
        open={selectedCampaign !== null}
        onClose={() => setSelectedCampaign(null)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>{selectedCampaign?.title}</DialogTitle>
        <DialogContent>
          {selectedCampaign ? (
            <Stack spacing={2}>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                <Chip
                  size="small"
                  label={hazardLabel(selectedCampaign.hazardType)}
                />
                {selectedCampaign.targetRegions.map((region) => (
                  <Chip
                    key={region}
                    size="small"
                    variant="outlined"
                    label={region}
                  />
                ))}
                {selectedCampaign.languages.map((language) => (
                  <Chip
                    key={language}
                    size="small"
                    variant="outlined"
                    label={language}
                  />
                ))}
              </Stack>

              {selectedCampaign.contentBlocks.map((block, index) => (
                <Box key={index}>
                  <Typography variant="subtitle2">{block.title}</Typography>
                  <ContentBlockView block={block} />
                </Box>
              ))}

              {selectedCampaign.linkedGuideIds?.length ||
              selectedCampaign.linkedAlertIds?.length ? (
                <Box>
                  <Typography variant="subtitle2">
                    Related guidance and alerts
                  </Typography>
                  <Stack
                    direction="row"
                    spacing={1}
                    flexWrap="wrap"
                    sx={{ mt: 1 }}
                  >
                    {selectedCampaign.linkedGuideIds?.map((id) => (
                      <Chip
                        key={id}
                        size="small"
                        icon={<BookOpen size={14} />}
                        label={`Guide ${id}`}
                      />
                    ))}
                    {selectedCampaign.linkedAlertIds?.map((id) => (
                      <Chip
                        key={id}
                        size="small"
                        icon={<Megaphone size={14} />}
                        label={`Alert ${id}`}
                      />
                    ))}
                  </Stack>
                </Box>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
      </Dialog>
    </Paper>
  );
}

function ContentBlockView({ block }: { block: CampaignContentBlock }) {
  if (block.type === "checklist") {
    return (
      <Stack spacing={0.5} sx={{ mt: 1 }}>
        {(block.items ?? []).map((item, index) => (
          <Stack
            direction="row"
            spacing={1}
            alignItems="flex-start"
            key={index}
          >
            <CheckCircle2 size={16} color={nadaaBrand.colors.green} />
            <Typography variant="body2">{item}</Typography>
          </Stack>
        ))}
      </Stack>
    );
  }
  return (
    <Typography variant="body2" className="guide-body" sx={{ mt: 0.5 }}>
      {block.body}
    </Typography>
  );
}
