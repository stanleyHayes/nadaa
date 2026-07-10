import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  ButtonGroup,
  Chip,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import {
  AlertOctagon,
  Clock3,
  Loader2,
  Megaphone,
  RefreshCw,
} from "lucide-react";
import type {
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
} from "@nadaa/shared-types";
import { NOTIFICATION_API_BASE } from "@/app/config";
import { AnimatedCounter, PageHeader, Reveal } from "../components";
import { PageBanner } from "../components/PageBanner";
import { buildFallbackAlerts } from "../data";
import type { AlertFeedState, AlertFeedView } from "../types";
import {
  alertSeverityLabel,
  alertSeverityTone,
  alertStatusLabel,
  extractAPIError,
  formatDateTime,
  hazardLabel,
  hazardRoleFor,
  hazardRoles,
} from "../utils";

/**
 * Self-contained approved-warnings feed migrated from the legacy `#alerts`
 * section: loads the notification alert feed (with an offline fallback),
 * filters current/expired/all, and renders each warning as a severity-toned
 * card. Owns its own state, effect and refresh handler.
 */
function AlertsFeed() {
  const [alertFeed, setAlertFeed] = useState<CitizenAlertFeedItem[]>(() =>
    buildFallbackAlerts(),
  );
  const [alertFeedView, setAlertFeedView] = useState<AlertFeedView>("current");
  const [alertFeedState, setAlertFeedState] = useState<AlertFeedState>({
    status: "idle",
    message: "Showing saved warnings until the feed refreshes.",
  });

  const visibleAlerts = useMemo(
    () =>
      alertFeed.filter((alert) => {
        if (alertFeedView === "all") {
          return true;
        }
        return alert.status === alertFeedView;
      }),
    [alertFeed, alertFeedView],
  );
  const currentAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "current").length,
    [alertFeed],
  );
  const expiredAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "expired").length,
    [alertFeed],
  );

  async function fetchAlertFeed() {
    if (!navigator.onLine) {
      setAlertFeedState({
        status: "error",
        message: "Alert feed needs a connection. Showing saved warnings.",
      });
      return;
    }

    setAlertFeedState({ status: "loading", message: "Refreshing alerts" });

    try {
      const response = await fetch(
        `${NOTIFICATION_API_BASE}/notifications/alerts?includeExpired=true`,
      );
      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const payload = (await response.json()) as CitizenAlertFeedResponse;
      setAlertFeed(
        payload.alerts.length > 0 ? payload.alerts : buildFallbackAlerts(),
      );
      setAlertFeedState({
        status: "idle",
        message: `Alert feed updated ${formatDateTime(payload.generatedAt)}.`,
      });
    } catch (error) {
      setAlertFeed(buildFallbackAlerts());
      setAlertFeedState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Live alert feed unavailable. Showing saved warnings.",
      });
    }
  }

  useEffect(() => {
    void fetchAlertFeed();
  }, []);

  return (
    <Reveal className="citizen-section">
      <Paper className="surface" id="alerts" component="section">
        <PageHeader
          icon={Megaphone}
          title="Live warnings"
          subtitle={
            <>
              <AnimatedCounter value={currentAlertCount} /> current ·{" "}
              <AnimatedCounter value={expiredAlertCount} /> expired
            </>
          }
          tone="red"
          action={
            <Button
              type="button"
              variant="outlined"
              size="small"
              startIcon={
                alertFeedState.status === "loading" ? (
                  <Loader2 size={16} className="spin-icon" />
                ) : (
                  <RefreshCw size={16} />
                )
              }
              onClick={() => void fetchAlertFeed()}
              disabled={alertFeedState.status === "loading"}
            >
              Refresh
            </Button>
          }
        />
        <ButtonGroup
          variant="outlined"
          size="small"
          className="alert-filter-group"
          aria-label="alert feed filter"
        >
          <Button
            variant={alertFeedView === "current" ? "contained" : "outlined"}
            onClick={() => setAlertFeedView("current")}
          >
            Current
          </Button>
          <Button
            variant={alertFeedView === "expired" ? "contained" : "outlined"}
            onClick={() => setAlertFeedView("expired")}
          >
            Expired
          </Button>
          <Button
            variant={alertFeedView === "all" ? "contained" : "outlined"}
            onClick={() => setAlertFeedView("all")}
          >
            All
          </Button>
        </ButtonGroup>

        {alertFeedState.status === "error" ? (
          <Alert severity="warning" className="warning-alert">
            {alertFeedState.message}
          </Alert>
        ) : null}
        {alertFeedState.status === "idle" && alertFeedState.message ? (
          <Typography
            variant="caption"
            color="text.secondary"
            className="alert-feed-note"
          >
            {alertFeedState.message}
          </Typography>
        ) : null}

        <Stack spacing={1.5}>
          {visibleAlerts.length > 0 ? (
            visibleAlerts.map((alert) => (
              <Alert
                key={alert.id}
                severity={alertSeverityTone(alert.severity, alert.status)}
                className="warning-alert citizen-alert-card"
                icon={
                  alert.status === "expired" ? <Clock3 size={20} /> : undefined
                }
              >
                <Stack spacing={0.75}>
                  <Stack
                    direction="row"
                    spacing={1}
                    justifyContent="space-between"
                    alignItems="flex-start"
                  >
                    <Box>
                      <Typography variant="subtitle2">{alert.title}</Typography>
                      <Typography variant="body2">
                        {alert.targetLabel} ·{" "}
                        {alertSeverityLabel(alert.severity)}
                      </Typography>
                    </Box>
                    <Chip
                      size="small"
                      icon={
                        alert.status === "current" ? (
                          <AlertOctagon size={16} />
                        ) : (
                          <Clock3 size={16} />
                        )
                      }
                      label={alertStatusLabel(alert.status)}
                      color={alert.status === "current" ? "error" : "default"}
                    />
                  </Stack>
                  <Typography variant="body2">{alert.message}</Typography>
                  <Typography variant="body2">
                    {alert.recommendedAction}
                  </Typography>
                  <Stack direction="row" spacing={0.75} flexWrap="wrap">
                    {(() => {
                      const hRole = hazardRoleFor(alert.hazardType);
                      return (
                        <Chip
                          size="small"
                          variant="outlined"
                          label={hazardLabel(alert.hazardType)}
                          sx={{
                            borderColor: hazardRoles[hRole].border,
                            color: hazardRoles[hRole].foreground,
                            backgroundColor: hazardRoles[hRole].background,
                          }}
                        />
                      );
                    })()}
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Until ${formatDateTime(alert.expiresAt)}`}
                    />
                    {alert.evacuationRequired ? (
                      <Chip
                        size="small"
                        color="error"
                        label="Evacuation possible"
                      />
                    ) : null}
                  </Stack>
                </Stack>
              </Alert>
            ))
          ) : (
            <Alert severity="info" className="warning-alert">
              No {alertFeedView === "all" ? "" : alertFeedView} alerts are
              available.
            </Alert>
          )}
        </Stack>
      </Paper>
    </Reveal>
  );
}

/** Alerts feed (route `/alerts`). Migrated from the legacy `#alerts` section. */
export function AlertsPage() {
  return (
    <>
      <PageBanner
        eyebrow="Alerts"
        subtitle="Approved flood and hazard warnings for your area — current and past."
        title="Live flood & hazard alerts"
      />
      <div className="citizen-shell">
        <AlertsFeed />
      </div>
    </>
  );
}

export default AlertsPage;
