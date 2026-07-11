import { useEffect, useMemo, useRef, useState } from "react";
import { useCitizenSession } from "../session";
import {
  playAlertTone,
  quietHoursActive,
  severityRank,
  shouldPlayAlertSound,
} from "../alertSound";
import {
  Alert,
  Button,
  Chip,
  LinearProgress,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { Loader2, Megaphone, RefreshCw, ShieldCheck } from "lucide-react";
import type {
  CitizenAlertFeedItem,
  CitizenAlertFeedResponse,
  CitizenAlertFeedStatus,
} from "@nadaa/shared-types";
import { NOTIFICATION_API_BASE } from "@/app/config";
import {
  AnimatedCounter,
  DataTable,
  DetailDialog,
  EmptyState,
  PageHeader,
  Reveal,
  type DataTableColumn,
  type DataTableFilter,
  type DetailField,
} from "../components";
import { PageBanner } from "../components/PageBanner";
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
        <Typography variant="caption" sx={{
          color: "text.secondary"
        }}>
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
        <Typography variant="caption" sx={{
          color: "text.secondary"
        }}>
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
 * Fuller detail for a single warning, shown in the row-opened dialog (the
 * light-detail half of the list/detail split). The long guidance/message body
 * spans the full dialog width; the rest read as a compact definition list.
 */
function alertDetailFields(alert: CitizenAlertFeedItem): DetailField[] {
  return [
    { label: "Hazard", value: hazardLabel(alert.hazardType) },
    { label: "Area", value: alert.targetLabel },
    { label: "Severity", value: alertSeverityLabel(alert.severity) },
    { label: "Status", value: ALERT_STATUS_DISPLAY[alert.status] },
    { label: "Effective", value: formatDateTime(alert.startsAt) },
    { label: "Expires", value: formatDateTime(alert.expiresAt) },
    {
      label: "Recommended action",
      value: alert.recommendedAction,
      full: true,
    },
    { label: "Guidance", value: alert.message, full: true },
  ];
}

/**
 * Self-contained approved-warnings feed migrated from the legacy `#alerts`
 * section: loads the notification alert feed (with an offline fallback) and
 * presents current, upcoming and expired warnings in one public, searchable and
 * filterable `DataTable`. Owns its own state, effect and refresh handler.
 */
function AlertsFeed() {
  const [alertFeed, setAlertFeed] = useState<CitizenAlertFeedItem[]>([]);
  const [alertFeedState, setAlertFeedState] = useState<AlertFeedState>({
    status: "loading",
    message: "Loading alerts",
  });
  const [detailAlert, setDetailAlert] = useState<CitizenAlertFeedItem | null>(
    null,
  );
  const { preferences } = useCitizenSession();
  // Track which alerts we've already seen so a chime only fires for genuinely
  // new warnings, never on the first load.
  const seenAlertIds = useRef<Set<string>>(new Set());
  const seededRef = useRef(false);

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
      setAlertFeed([]);
      setAlertFeedState({
        status: "error",
        message: "Alert feed needs a connection. Reconnect and try again.",
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
      setAlertFeed(payload.alerts);

      // Sound a warning tone for newly-arrived alerts. Quiet Hours (DND) silence
      // everything below level 5; an emergency (level 5) alert always sounds.
      const fresh = payload.alerts.filter(
        (alert) => !seenAlertIds.current.has(alert.id),
      );
      fresh.forEach((alert) => seenAlertIds.current.add(alert.id));
      if (seededRef.current && fresh.length > 0) {
        const loudest = fresh.reduce((max, alert) =>
          severityRank(alert.severity) > severityRank(max.severity) ? alert : max,
        );
        if (
          shouldPlayAlertSound(loudest.severity, {
            soundEnabled: preferences.soundAlerts ?? true,
            quietHoursActive: quietHoursActive(preferences.quietHours),
          })
        ) {
          playAlertTone(loudest.severity);
        }
      }
      seededRef.current = true;

      setAlertFeedState({
        status: "idle",
        message: `Alert feed updated ${formatDateTime(payload.generatedAt)}.`,
      });
    } catch {
      setAlertFeed([]);
      setAlertFeedState({
        status: "error",
        message: "Couldn't reach the alerts service. Try refreshing.",
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

          {alertFeedState.status === "loading" ? (
            <LinearProgress className="feed-progress" />
          ) : null}
          {alertFeedState.status === "error" ? (
            <Alert severity="error" className="warning-alert">
              {alertFeedState.message}
            </Alert>
          ) : null}
          {alertFeedState.status === "idle" && alertFeedState.message ? (
            <Typography
              variant="caption"
              className="alert-feed-note"
              sx={{
                color: "text.secondary"
              }}
            >
              {alertFeedState.message}
            </Typography>
          ) : null}
        </Paper>

        {alertFeedState.status === "loading" && alertFeed.length === 0 ? null : (
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
            emptyState={
              <EmptyState
                icon={ShieldCheck}
                tone="green"
                title="No active warnings"
                description="No approved alerts match your search — that's good news."
              />
            }
            onRowClick={setDetailAlert}
          />
        )}

        <DetailDialog
          open={detailAlert !== null}
          onClose={() => setDetailAlert(null)}
          title={detailAlert?.title}
          subtitle={
            detailAlert
              ? `${hazardLabel(detailAlert.hazardType)} · ${detailAlert.targetLabel}`
              : undefined
          }
          fields={detailAlert ? alertDetailFields(detailAlert) : []}
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
