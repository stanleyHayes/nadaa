import { useState } from "react";
import { Alert, Box, Button, Snackbar, Stack } from "@mui/material";
import { Bell, Mail, MessageSquareWarning, Save, Siren } from "lucide-react";
import type { AuthorityAccountPreferences } from "@/app/session";
import { PreferenceRow, SettingCard } from "./primitives";

export function NotificationsTab({
  preferences,
  onUpdatePreferences,
}: {
  preferences: AuthorityAccountPreferences;
  onUpdatePreferences: (patch: Partial<AuthorityAccountPreferences>) => void;
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
        description="Choose how the command desk reaches you in this browser."
      >
        <Stack spacing={1.5}>
          <PreferenceRow
            icon={Siren}
            label="In-app alerts"
            description="Show incident and approval notices in the top navigation bell."
            checked={draft.inAppAlerts}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, inAppAlerts: checked }))
            }
          />
          <PreferenceRow
            icon={MessageSquareWarning}
            label="Critical incident SMS"
            description="Text your duty phone when an emergency or severe incident opens."
            checked={draft.criticalSms}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, criticalSms: checked }))
            }
          />
          <PreferenceRow
            icon={Mail}
            label="Alert-approval email"
            description="Email you when a public alert is waiting for your approval."
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
