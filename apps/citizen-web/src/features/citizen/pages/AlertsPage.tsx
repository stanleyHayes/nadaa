import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Chip, Paper, Stack, Typography } from "@mui/material";
import { Loader2, Megaphone, RefreshCw } from "lucide-react";
import type {
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
  CitizenAlertFeedStatus,
} from "@nadaa/shared-types";
import { NOTIFICATION_API_BASE } from "@/app/config";
import {
  AnimatedCounter,
  DataTable,
  PageHeader,
  Reveal,
  type DataTableColumn,
  type DataTableFilter,
} from "../components";
import { PageBanner } from "../components/PageBanner";
import { buildFallbackAlerts } from "../data";
import type { AlertFeedState } from "../types";
import {
  alertSeverityLabel,
  extractAPIError,
  formatDateTime,
  hazardLabel,
  hazardRoleFor,
  hazardRoles,
  severityRoleFor,
  severityRoles,
} from "../utils";

/** Public-facing status label for each feed status (current reads as "Active"). */
const ALERT_STATUS_DISPLAY: Record<CitizenAlertFeedStatus, string> = {
  current: "Active",
  expired: "Expired",
  upcoming: "Upcoming",
};

/** Chip tone per feed status: active warnings are urgent, past ones muted. */
const ALERT_STATUS_COLOR: Record<
  CitizenAlertFeedStatus,
  "error" | "default" | "warning"
> = {
  current: "error",
  expired: "default",
  upcoming: "warning",
};

/**
 * Column definitions for the public alert table. Kept at module scope since they
 * only depend on shared helpers, not component state. Severity and hazard reuse
 * the brand colour roles so the chips match the rest of the "Navy Command" set.
 */
const alertColumns: DataTableColumn<CitizenAlertFeedItem>[] = [
  {
    key: "title",
    label: "Warning",
    render: (alert) => (
      <Stack spacing={0.25}>
        <Typography variant="body2" sx={{ fontWeight: 600 }}>
          {alert.title}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {alert.recommendedAction}
        </Typography>
      </Stack>
    ),
  },
  {
    key: "hazardType",
    label: "Hazard",
    render: (alert) => {
      const role = hazardRoleFor(alert.hazardType);
      return (
        <Chip
          size="small"
          variant="outlined"
          label={hazardLabel(alert.hazardType)}
          sx={{
            borderColor: hazardRoles[role].border,
            color: hazardRoles[role].foreground,
            backgroundColor: hazardRoles[role].background,
          }}
        />
      );
    },
  },
  {
    key: "targetLabel",
    label: "Area",
    render: (alert) => alert.targetLabel,
  },
  {
    key: "severity",
    label: "Severity",
    render: (alert) => {
      const role = severityRoleFor(alert.severity);
      return (
        <Chip
          size="small"
          variant="outlined"
          label={alertSeverityLabel(alert.severity)}
          sx={{
            borderColor: severityRoles[role].border,
            color: severityRoles[role].foreground,
            backgroundColor: severityRoles[role].background,
          }}
        />
      );
    },
  },
  {
    key: "startsAt",
    label: "Effective",
    render: (alert) => (
      <Stack spacing={0.25}>
        <Typography variant="body2">
          {formatDateTime(alert.startsAt)}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          Until {formatDateTime(alert.expiresAt)}
        </Typography>
      </Stack>
    ),
  },
  {
    key: "status",
    label: "Status",
    render: (alert) => (
      <Chip
        size="small"
        label={ALERT_STATUS_DISPLAY[alert.status]}
        color={ALERT_STATUS_COLOR[alert.status]}
      />
    ),
  },
];

/**
 * Self-contained approved-warnings feed migrated from the legacy `#alerts`
 * section: loads the notification alert feed (with an offline fallback) and
 * presents current, upcoming and expired warnings in one public, searchable and
 * filterable `DataTable`. Owns its own state, effect and refresh handler.
 */
function AlertsFeed() {
  const [alertFeed, setAlertFeed] = useState<CitizenAlertFeedItem[]>(() =>
    buildFallbackAlerts(),
  );
  const [alertFeedState, setAlertFeedState] = useState<AlertFeedState>({
    status: "idle",
    message: "Showing saved warnings until the feed refreshes.",
  });

  const currentAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "current").length,
    [alertFeed],
  );
  const expiredAlertCount = useMemo(
    () => alertFeed.filter((alert) => alert.status === "expired").length,
    [alertFeed],
  );

  // Filter options are derived from the distinct values actually in the feed.
  const alertFilters = useMemo<DataTableFilter<CitizenAlertFeedItem>[]>(
    () => [
      {
        key: "severity",
        label: "Severity",
        options: [
          ...new Set(alertFeed.map((alert) => alertSeverityLabel(alert.severity))),
        ].sort(),
        valueOf: (alert) => alertSeverityLabel(alert.severity),
      },
      {
        key: "hazardType",
        label: "Hazard",
        options: [
          ...new Set(alertFeed.map((alert) => hazardLabel(alert.hazardType))),
        ].sort(),
        valueOf: (alert) => hazardLabel(alert.hazardType),
      },
      {
        key: "status",
        label: "Status",
        options: [
          ...new Set(alertFeed.map((alert) => ALERT_STATUS_DISPLAY[alert.status])),
        ].sort(),
        valueOf: (alert) => ALERT_STATUS_DISPLAY[alert.status],
      },
    ],
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
      <Stack spacing={2.5}>
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
        </Paper>

        <DataTable
          rows={alertFeed}
          columns={alertColumns}
          getRowKey={(alert) => alert.id}
          searchOf={(alert) =>
            `${alert.title} ${alert.targetLabel} ${hazardLabel(alert.hazardType)}`
          }
          searchPlaceholder="Search warnings, area, or hazard"
          filters={alertFilters}
          emptyMessage="No approved alerts match your search."
        />
      </Stack>
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
