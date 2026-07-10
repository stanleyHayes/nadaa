import { Button, Chip, IconButton, Paper, Stack, Tooltip, Typography } from "@mui/material";
import { Bell, Check, CheckCheck } from "lucide-react";
import { PageHeader } from "../../components/PageHeader";
import { useCitizenSession } from "../../session";
import { formatDateTime } from "../../utils";
import { notificationCategoryLabel, notificationTone } from "./data";

/** Notifications (route `/account/notifications`) with read/unread + mark-all. */
export function AccountNotifications() {
  const { session, notifications, markNotificationRead, markAllRead } =
    useCitizenSession();

  if (!session) {
    return null;
  }

  const unread = notifications.filter((item) => !item.read).length;

  return (
    <Stack spacing={2.5} className="account-section">
      <Paper className="surface" component="section">
        <PageHeader
          icon={Bell}
          title="Notifications"
          subtitle={
            unread > 0
              ? `${unread} unread of ${notifications.length}`
              : "You're all caught up."
          }
          tone="red"
          as="h2"
          action={
            <Button
              type="button"
              variant="outlined"
              size="small"
              startIcon={<CheckCheck size={16} />}
              onClick={markAllRead}
              disabled={unread === 0}
            >
              Mark all read
            </Button>
          }
        />

        {notifications.length === 0 ? (
          <Typography sx={{ py: 3 }} align="center" color="text.secondary">
            No notifications yet. Alerts and report updates will appear here.
          </Typography>
        ) : (
          <Stack spacing={1.25}>
            {notifications.map((item) => (
              <Paper
                variant="outlined"
                key={item.id}
                className={
                  item.read
                    ? "account-notice"
                    : "account-notice account-notice--unread"
                }
              >
                <span
                  className="account-notice__dot"
                  aria-hidden="true"
                  data-read={item.read}
                />
                <div className="account-notice__body">
                  <div className="account-notice__head">
                    <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                      {item.title}
                    </Typography>
                    <Chip
                      size="small"
                      label={notificationCategoryLabel[item.category]}
                      color={notificationTone[item.category]}
                      variant="outlined"
                    />
                  </div>
                  <Typography variant="body2" color="text.secondary">
                    {item.body}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {formatDateTime(item.at)}
                  </Typography>
                </div>
                {item.read ? null : (
                  <Tooltip title="Mark as read">
                    <IconButton
                      aria-label={`Mark "${item.title}" as read`}
                      onClick={() => markNotificationRead(item.id)}
                      size="small"
                    >
                      <Check size={18} />
                    </IconButton>
                  </Tooltip>
                )}
              </Paper>
            ))}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}

export default AccountNotifications;
