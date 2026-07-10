import { Paper, Stack, Typography } from "@mui/material";
import {
  Bell,
  BellRing,
  FileText,
  LayoutDashboard,
  MapPinned,
  Radio,
  Settings,
  Siren,
} from "lucide-react";
import { PageHeader } from "../../components/PageHeader";
import { useCitizenSession } from "../../session";
import { formatDateTime } from "../../utils";
import { ShortcutCard, StatCard } from "./components";

/** Account overview / dashboard (index route of `/account`). */
export function AccountOverview() {
  const { session, savedReports, notifications, preferences } =
    useCitizenSession();

  if (!session) {
    return null;
  }

  const unread = notifications.filter((item) => !item.read).length;
  const activeChannels = [
    preferences.alertChannels.sms && "SMS",
    preferences.alertChannels.email && "Email",
    preferences.alertChannels.push && "Push",
  ].filter(Boolean) as string[];

  const latestReport = savedReports[0];
  const latestNotification = notifications[0];

  return (
    <Stack spacing={2.5} className="account-section">
      <Paper className="surface" component="section">
        <PageHeader
          icon={LayoutDashboard}
          title="Your dashboard"
          subtitle="A quick read on your reports, alerts and how we reach you."
          tone="navy"
          as="h2"
        />
        <div className="account-stat-grid">
          <StatCard
            icon={FileText}
            value={savedReports.length}
            label={savedReports.length === 1 ? "Report filed" : "Reports filed"}
            hint="Tracked on this device"
            tone="green"
          />
          <StatCard
            icon={BellRing}
            value={unread}
            label={unread === 1 ? "Unread alert" : "Unread alerts"}
            hint={unread > 0 ? "Tap Notifications to catch up" : "You're all caught up"}
            tone="red"
          />
          <StatCard
            icon={MapPinned}
            value={preferences.regionOfInterest}
            label="Region watched"
            hint="Alerts focused here"
            tone="gold"
          />
          <StatCard
            icon={Radio}
            value={activeChannels.length > 0 ? activeChannels.join(" · ") : "None"}
            label="Alert channels"
            hint="Manage under Settings"
            tone="navy"
          />
        </div>
      </Paper>
      <Paper className="surface" component="section">
        <PageHeader
          icon={Siren}
          title="Quick actions"
          subtitle="Jump straight to what matters most."
          tone="gold"
          as="h3"
        />
        <div className="account-shortcut-grid">
          <ShortcutCard
            icon={Siren}
            title="File a report"
            description="Tell NADMO what you're seeing on the ground."
            to="/report"
          />
          <ShortcutCard
            icon={MapPinned}
            title="Check my risk"
            description="See the flood risk for your area right now."
            to="/"
          />
          <ShortcutCard
            icon={Bell}
            title="View notifications"
            description="Read the latest alerts and updates for you."
            to="/account/notifications"
          />
          <ShortcutCard
            icon={Settings}
            title="Update preferences"
            description="Change your language, channels and quiet hours."
            to="/account/settings"
          />
        </div>
      </Paper>
      <Paper className="surface" component="section">
        <PageHeader
          icon={FileText}
          title="Recent activity"
          subtitle="Your latest report and notification."
          tone="green"
          as="h3"
        />
        <Stack spacing={1.5}>
          <Paper variant="outlined" className="account-activity-row">
            <span className="account-activity-row__icon" aria-hidden="true">
              <FileText size={18} />
            </span>
            <div className="account-activity-row__text">
              {latestReport ? (
                <>
                  <Typography variant="subtitle2">
                    {latestReport.reference}
                  </Typography>
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    Filed {formatDateTime(latestReport.at)} ·{" "}
                    {latestReport.priorityReview
                      ? "Priority review"
                      : "Submitted"}
                  </Typography>
                </>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  You haven't filed a report yet. Filing one keeps a copy here.
                </Typography>
              )}
            </div>
          </Paper>
          <Paper variant="outlined" className="account-activity-row">
            <span className="account-activity-row__icon" aria-hidden="true">
              <Bell size={18} />
            </span>
            <div className="account-activity-row__text">
              {latestNotification ? (
                <>
                  <Typography variant="subtitle2">
                    {latestNotification.title}
                  </Typography>
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    {formatDateTime(latestNotification.at)}
                  </Typography>
                </>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  No notifications yet.
                </Typography>
              )}
            </div>
          </Paper>
        </Stack>
      </Paper>
    </Stack>
  );
}

export default AccountOverview;
