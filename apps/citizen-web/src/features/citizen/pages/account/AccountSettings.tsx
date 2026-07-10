import { useState } from "react";
import { Box, Paper, Stack, Tab, Tabs } from "@mui/material";
import { KeyRound, SlidersHorizontal, UserCog } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { PageHeader } from "../../components/PageHeader";
import { useCitizenSession } from "../../session";
import {
  AppearanceCard,
  PasswordForm,
  PreferencesForm,
  ProfileForm,
} from "./components";

type SettingsTab = {
  key: string;
  label: string;
  icon: LucideIcon;
};

const SETTINGS_TABS: SettingsTab[] = [
  { key: "profile", label: "Profile", icon: UserCog },
  { key: "password", label: "Password", icon: KeyRound },
  { key: "preferences", label: "Preferences", icon: SlidersHorizontal },
];

/**
 * Settings (route `/account/settings`) with three in-page sections switched by
 * MUI tabs: edit profile, change password, and alert preferences.
 */
export function AccountSettings() {
  const { session, preferences } = useCitizenSession();
  const [active, setActive] = useState(0);

  if (!session) {
    return null;
  }

  return (
    <div className="account-section">
      <Paper className="surface" component="section">
        <PageHeader
          icon={UserCog}
          title="Settings"
          subtitle="Keep your details, password and alert preferences up to date."
          tone="navy"
          as="h2"
        />

        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 3 }}>
          <Tabs
            allowScrollButtonsMobile
            aria-label="Account settings sections"
            onChange={(_event, value: number) => setActive(value)}
            scrollButtons="auto"
            value={active}
            variant="scrollable"
          >
            {SETTINGS_TABS.map(({ key, label, icon: Icon }, index) => (
              <Tab
                aria-controls={`settings-panel-${key}`}
                icon={<Icon aria-hidden="true" size={18} />}
                iconPosition="start"
                id={`settings-tab-${key}`}
                key={key}
                label={label}
                sx={{ minHeight: 52, textTransform: "none" }}
                value={index}
              />
            ))}
          </Tabs>
        </Box>

        <div
          aria-labelledby="settings-tab-profile"
          hidden={active !== 0}
          id="settings-panel-profile"
          role="tabpanel"
        >
          {active === 0 ? <ProfileForm session={session} /> : null}
        </div>
        <div
          aria-labelledby="settings-tab-password"
          hidden={active !== 1}
          id="settings-panel-password"
          role="tabpanel"
        >
          {active === 1 ? <PasswordForm /> : null}
        </div>
        <div
          aria-labelledby="settings-tab-preferences"
          hidden={active !== 2}
          id="settings-panel-preferences"
          role="tabpanel"
        >
          {active === 2 ? (
            <Stack spacing={3}>
              <AppearanceCard />
              <PreferencesForm preferences={preferences} />
            </Stack>
          ) : null}
        </div>
      </Paper>
    </div>
  );
}

export default AccountSettings;
