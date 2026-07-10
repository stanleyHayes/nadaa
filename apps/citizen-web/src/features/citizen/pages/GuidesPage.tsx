import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Typography,
} from "@mui/material";
import {
  BookOpen,
  CheckCircle2,
  Languages,
  Loader2,
  Phone,
  RefreshCw,
  WifiOff,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  EmergencyGuideRecord,
  GuideListResponse,
} from "@nadaa/shared-types";
import { GUIDE_API_BASE } from "@/app/config";
import {
  buildFallbackGuides,
  guideHazardOptions,
  guideLanguageOptions,
  guideStageOptions,
} from "../data";
import type {
  GuideCacheInfo,
  GuideFilters,
  GuideHazardFilter,
  GuideStageFilter,
  GuideState,
} from "../types";
import {
  extractAPIError,
  filterGuides,
  formatDateTime,
  guideLanguageLabel,
  guideStageLabel,
  hazardLabel,
  readGuideCache,
  writeGuideCache,
} from "../utils";
import { AnimatedCounter, PageHeader, Reveal } from "../components";
import { PageBanner } from "../components/PageBanner";

/**
 * Emergency guides (route `/guides`). Self-contained migration of the legacy
 * `#guides` section: hazard/stage/language filters over the offline guide
 * cache, a featured guide, a compact guide list, and network/offline/error
 * states backed by the guide service with a local-cache fallback.
 */
export function GuidesPage() {
  const [guideFilters, setGuideFilters] = useState<GuideFilters>({
    hazard: "all",
    stage: "during",
    language: "en",
  });
  const [guides, setGuides] = useState<EmergencyGuideRecord[]>(() => {
    const cached = readGuideCache();
    return cached?.guides ?? buildFallbackGuides();
  });
  const [guideCacheInfo, setGuideCacheInfo] = useState<GuideCacheInfo>(() => {
    const cached = readGuideCache();
    return cached
      ? {
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        }
      : {
          cachedAt: new Date().toISOString(),
          source: "fixture",
          language: "en",
        };
  });
  const [guideState, setGuideState] = useState<GuideState>({
    status: "idle",
    message: "Offline guides are ready.",
  });

  const visibleGuides = useMemo(
    () => filterGuides(guides, guideFilters),
    [guides, guideFilters],
  );
  const featuredGuide = visibleGuides[0];
  const offlineGuideCount = useMemo(
    () => guides.filter((guide) => guide.offlineAvailable).length,
    [guides],
  );

  useEffect(() => {
    void fetchGuides(guideFilters.language);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function fetchGuides(language = guideFilters.language) {
    if (!navigator.onLine) {
      const cached = readGuideCache();
      if (cached?.guides.length) {
        setGuides(cached.guides);
        setGuideCacheInfo({
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        });
        setGuideState({
          status: "offline",
          message: `Offline guide cache ready from ${formatDateTime(cached.cachedAt)}.`,
        });
        return;
      }

      setGuides(buildFallbackGuides());
      setGuideCacheInfo({
        cachedAt: new Date().toISOString(),
        source: "fixture",
        language: "en",
      });
      setGuideState({
        status: "offline",
        message: "Showing starter emergency guides until connection returns.",
      });
      return;
    }

    setGuideState({ status: "loading", message: "Refreshing guides" });

    try {
      const params = new URLSearchParams({ language, offline: "true" });

      const response = await fetch(`${GUIDE_API_BASE}/guides?${params}`);
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as GuideListResponse;
      const nextGuides = payload.guides.length
        ? payload.guides
        : buildFallbackGuides();
      const cachedAt = new Date().toISOString();
      setGuides(nextGuides);
      setGuideCacheInfo({ cachedAt, source: "network", language });
      writeGuideCache(nextGuides, language, cachedAt);
      setGuideState({
        status: "idle",
        message: `Saved ${nextGuides.length} offline guides for ${guideLanguageLabel(language)}.`,
      });
    } catch (error) {
      const cached = readGuideCache();
      if (cached?.guides.length) {
        setGuides(cached.guides);
        setGuideCacheInfo({
          cachedAt: cached.cachedAt,
          source: "cache",
          language: cached.language,
        });
        setGuideState({
          status: "offline",
          message: `Live guide service unavailable. Using offline cache from ${formatDateTime(cached.cachedAt)}.`,
        });
        return;
      }

      setGuides(buildFallbackGuides());
      setGuideCacheInfo({
        cachedAt: new Date().toISOString(),
        source: "fixture",
        language: "en",
      });
      setGuideState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Could not load emergency guides.",
      });
    }
  }

  function updateGuideLanguage(language: string) {
    setGuideFilters((current) => ({ ...current, language }));
    void fetchGuides(language);
  }

  return (
    <>
      <PageBanner
        eyebrow="Emergency guidance"
        subtitle="Step-by-step flood and hazard guidance in six Ghanaian languages — available offline."
        title="Emergency preparedness guides"
      />
      <div className="citizen-shell">
        <Reveal className="citizen-section">
          <Paper className="surface" id="guides" component="section">
            <PageHeader
              icon={BookOpen}
              title="Emergency guides"
              subtitle={
                <>
                  <AnimatedCounter value={offlineGuideCount} /> saved for
                  offline use
                </>
              }
              tone="green"
              action={
                <Button
                  type="button"
                  variant="outlined"
                  size="small"
                  startIcon={
                    guideState.status === "loading" ? (
                      <Loader2 size={16} className="spin-icon" />
                    ) : (
                      <RefreshCw size={16} />
                    )
                  }
                  onClick={() => void fetchGuides()}
                  disabled={guideState.status === "loading"}
                >
                  Refresh
                </Button>
              }
            />

            <Grid container spacing={1.25} className="guide-filter-grid">
              <Grid size={{ xs: 12, sm: 4 }}>
                <FormControl fullWidth size="small">
                  <InputLabel>Hazard</InputLabel>
                  <Select
                    value={guideFilters.hazard}
                    label="Hazard"
                    onChange={(event) =>
                      setGuideFilters((current) => ({
                        ...current,
                        hazard: event.target.value as GuideHazardFilter,
                      }))
                    }
                  >
                    {guideHazardOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid size={{ xs: 12, sm: 4 }}>
                <FormControl fullWidth size="small">
                  <InputLabel>Stage</InputLabel>
                  <Select
                    value={guideFilters.stage}
                    label="Stage"
                    onChange={(event) =>
                      setGuideFilters((current) => ({
                        ...current,
                        stage: event.target.value as GuideStageFilter,
                      }))
                    }
                  >
                    {guideStageOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid size={{ xs: 12, sm: 4 }}>
                <FormControl fullWidth size="small">
                  <InputLabel>Language</InputLabel>
                  <Select
                    value={guideFilters.language}
                    label="Language"
                    onChange={(event) => updateGuideLanguage(event.target.value)}
                    startAdornment={
                      <Languages size={16} className="select-leading-icon" />
                    }
                  >
                    {guideLanguageOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <Alert
              severity={
                guideState.status === "error"
                  ? "error"
                  : guideState.status === "offline"
                    ? "warning"
                    : "info"
              }
              icon={guideState.status === "offline" ? <WifiOff /> : undefined}
              className="warning-alert guide-cache-alert"
            >
              {guideState.message ??
                `Cached ${guideLanguageLabel(guideCacheInfo.language)} guides ${formatDateTime(guideCacheInfo.cachedAt)}.`}
            </Alert>

            {featuredGuide ? (
              <Box className="guide-feature">
                <Stack direction="row" spacing={1} alignItems="center">
                  <Chip
                    size="small"
                    color="success"
                    label={guideStageLabel(featuredGuide.stage)}
                  />
                  <Chip
                    size="small"
                    variant="outlined"
                    label={hazardLabel(featuredGuide.hazardType)}
                  />
                </Stack>
                <Typography variant="subtitle1">
                  {featuredGuide.title}
                </Typography>
                <Typography variant="body2" className="guide-body">
                  {featuredGuide.body}
                </Typography>
                <Button
                  fullWidth
                  variant="contained"
                  color="error"
                  startIcon={<Phone size={18} />}
                  href="tel:112"
                >
                  Call 112
                </Button>
              </Box>
            ) : (
              <Alert severity="info" className="warning-alert">
                No guide matches this filter yet. Try all hazards or all stages.
              </Alert>
            )}

            <Divider className="guide-divider" />

            <Stack spacing={1}>
              {visibleGuides.slice(0, 5).map((guide) => (
                <Paper
                  variant="outlined"
                  className="guide-list-row"
                  key={guide.id}
                >
                  <Stack direction="row" spacing={1.25}>
                    <CheckCircle2 size={18} color={nadaaBrand.colors.green} />
                    <Box>
                      <Stack
                        direction="row"
                        spacing={0.75}
                        alignItems="center"
                        flexWrap="wrap"
                      >
                        <Typography variant="subtitle2">
                          {guide.title}
                        </Typography>
                        {guide.offlineAvailable ? (
                          <Chip size="small" label="Offline" />
                        ) : null}
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {guideStageLabel(guide.stage)} ·{" "}
                        {hazardLabel(guide.hazardType)}
                      </Typography>
                    </Box>
                  </Stack>
                </Paper>
              ))}
            </Stack>
          </Paper>
        </Reveal>
      </div>
    </>
  );
}

export default GuidesPage;
