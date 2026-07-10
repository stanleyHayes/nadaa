import { useState } from "react";
import { Alert, Box, Button, Snackbar, Stack } from "@mui/material";
import { Rows3, Save, SlidersHorizontal, Zap } from "lucide-react";
import type { DispatcherAccountPreferences } from "@/app/session";
import { PreferenceRow, SettingCard } from "./primitives";

export function PreferencesTab({
  preferences,
  onUpdatePreferences,
}: {
  preferences: DispatcherAccountPreferences;
  onUpdatePreferences: (patch: Partial<DispatcherAccountPreferences>) => void;
}) {
  const [draft, setDraft] = useState({
    compactTables: preferences.compactTables,
    reducedMotion: preferences.reducedMotion,
  });
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    onUpdatePreferences(draft);
    setSaved(true);
  };

  return (
    <Box sx={{ maxWidth: 720 }}>
      <SettingCard
        icon={SlidersHorizontal}
        title="Preferences"
        description="Tune how this browser presents the dispatch console."
      >
        <Stack spacing={1.5}>
          <PreferenceRow
            icon={Rows3}
            label="Compact tables"
            description="Use tighter row spacing for dense incident and alert tables."
            checked={draft.compactTables}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, compactTables: checked }))
            }
          />
          <PreferenceRow
            icon={Zap}
            label="Reduced motion"
            description="Minimise transitions and animated effects across the desk."
            checked={draft.reducedMotion}
            onChange={(checked) =>
              setDraft((current) => ({ ...current, reducedMotion: checked }))
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
              Save preferences
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
          Preferences saved.
        </Alert>
      </Snackbar>
    </Box>
  );
}
