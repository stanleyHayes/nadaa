import { useState } from "react";
import { Alert, Box, Button, Snackbar, Stack } from "@mui/material";
import { Bell, Mail, MessageSquareWarning, Save, Siren } from "lucide-react";
import type { AgencyAccountPreferences } from "@/app/session";
import { PreferenceRow, SettingCard } from "./primitives";

export function NotificationsTab({
  preferences,
  onUpdatePreferences,
}: {
  preferences: AgencyAccountPreferences;
  onUpdatePreferences: (patch: Partial<AgencyAccountPreferences>) => void;
}) {
  const [draft, setDraft] = useState({
    inAppAlerts: preferences.inAppAlerts,
    criticalSms: preferences.criticalSms,
    approvalEmail: preferences.approvalEmail,
  });
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    onUpdatePreferences(draft);
    setSaved(true);
  };

  return (
    <Box sx={{ maxWidth: 720 }}>
      <SettingCard
        icon={Bell}
        title="Notifications"
        description="Choose how the agency desk reaches you in this browser."
      >
        <Stack spacing={1.5}>
          <PreferenceRow
            icon={Siren}
            label="In-app alerts"
            description="Show incident and response notices in the top navigation bell."
            checked={draft.inAppAlerts}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, inAppAlerts: checked }))
            }
          />
          <PreferenceRow
            icon={MessageSquareWarning}
            label="Critical incident SMS"
            description="Text your duty phone when an emergency or severe incident opens. Not connected yet — SMS delivery is coming soon; this toggle only saves your choice in this browser."
            checked={draft.criticalSms}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, criticalSms: checked }))
            }
          />
          <PreferenceRow
            icon={Mail}
            label="Aid-review email"
            description="Email you when an aid need is waiting for your review. Not connected yet — email delivery is coming soon; this toggle only saves your choice in this browser."
            checked={draft.approvalEmail}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, approvalEmail: checked }))
            }
          />
          <Box>
            <Button
              type="button"
              onClick={handleSave}
              variant="contained"
              color="primary"
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
            >
              Save notifications
            </Button>
          </Box>
        </Stack>
      </SettingCard>

      <Snackbar
        open={saved}
        autoHideDuration={3000}
        onClose={() => setSaved(false)}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          onClose={() => setSaved(false)}
          severity="success"
          variant="filled"
          sx={{ width: "100%" }}
        >
          Notification preferences saved.
        </Alert>
      </Snackbar>
    </Box>
  );
}
