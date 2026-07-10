import { Box, Paper, Stack, Tab, Tabs, Typography } from "@mui/material";
import {
  Bell,
  Settings,
  ShieldCheck,
  SlidersHorizontal,
  UserRound,
} from "lucide-react";
import type {
  AgencyAccountPreferences,
  AgencyProfilePatch,
  AgencySession,
  PasswordChangeResult,
} from "@/app/session";
import { ProfileTab } from "./ProfileTab";
import { SecurityTab } from "./SecurityTab";
import { NotificationsTab } from "./NotificationsTab";
import { PreferencesTab } from "./PreferencesTab";

export type SettingsTab =
  | "profile"
  | "security"
  | "notifications"
  | "preferences";

const TABS: Array<{
  value: SettingsTab;
  label: string;
  icon: typeof UserRound;
}> = [
  { value: "profile", label: "Profile", icon: UserRound },
  { value: "security", label: "Security", icon: ShieldCheck },
  { value: "notifications", label: "Notifications", icon: Bell },
  { value: "preferences", label: "Preferences", icon: SlidersHorizontal },
];

export type AccountSettingsProps = {
  tab: SettingsTab;
  onTabChange: (tab: SettingsTab) => void;
  user: AgencySession;
  preferences: AgencyAccountPreferences;
  onUpdateProfile: (patch: AgencyProfilePatch) => void;
  onUpdatePreferences: (patch: Partial<AgencyAccountPreferences>) => void;
  onSetMfaEnabled: (enabled: boolean) => void;
  onChangePassword: (current: string, next: string) => PasswordChangeResult;
};

export function AccountSettings({
  tab,
  onTabChange,
  user,
  preferences,
  onUpdateProfile,
  onUpdatePreferences,
  onSetMfaEnabled,
  onChangePassword,
}: AccountSettingsProps) {
  return (
    <Box sx={{ maxWidth: 1120, mx: "auto" }}>
      <Paper
        elevation={0}
        sx={{
          display: "flex",
          gap: 2,
          alignItems: "center",
          p: { xs: 2.5, md: 3.5 },
          borderRadius: "16px",
          color: "var(--nadaa-white, #ffffff)",
          background:
            "linear-gradient(150deg, var(--nadaa-navy, #0d1b3d) 0%, #0a1531 100%)",
          boxShadow: "var(--nadaa-shadow-md)",
        }}
      >
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 52,
            height: 52,
            borderRadius: "13px",
            color: "var(--nadaa-gold, #f4c20d)",
            backgroundColor: "rgba(255, 255, 255, 0.08)",
          }}
        >
          <Settings size={24} />
        </Box>
        <Box sx={{ minWidth: 0 }}>
          <Typography
            sx={{
              fontSize: "0.68rem",
              fontWeight: 700,
              letterSpacing: "0.18em",
              textTransform: "uppercase",
              color: "var(--nadaa-gold, #f4c20d)",
            }}
          >
            Account
          </Typography>
          <Typography
            component="h1"
            sx={{
              mt: 0.25,
              fontSize: { xs: "1.6rem", md: "1.9rem" },
              fontWeight: 800,
              lineHeight: 1.1,
            }}
          >
            Settings
          </Typography>
          <Typography
            sx={{
              mt: 0.5,
              fontSize: "0.9rem",
              color: "rgba(255, 255, 255, 0.72)",
            }}
          >
            Manage your profile, security, notifications and preferences.
          </Typography>
        </Box>
      </Paper>

      <Tabs
        value={tab}
        onChange={(_event, value: SettingsTab) => onTabChange(value)}
        variant="scrollable"
        scrollButtons="auto"
        allowScrollButtonsMobile
        aria-label="Account settings sections"
        sx={{
          mt: 3,
          minHeight: 52,
          borderBottom: "1px solid var(--nadaa-border, #dfeaf2)",
          "& .MuiTabs-indicator": {
            backgroundColor: "var(--nadaa-navy, #0d1b3d)",
            height: 3,
            borderRadius: "3px 3px 0 0",
          },
          "& .MuiTab-root": {
            minHeight: 52,
            px: 2,
            gap: 0.75,
            textTransform: "none",
            fontSize: "0.9rem",
            fontWeight: 600,
            color: "var(--nadaa-text-secondary, #555b66)",
          },
          "& .MuiTab-root.Mui-selected": {
            color: "var(--nadaa-navy, #0d1b3d)",
            fontWeight: 700,
          },
          "& .MuiTab-root.Mui-focusVisible": {
            outline: "2px solid var(--nadaa-gold, #f4c20d)",
            outlineOffset: "-2px",
            borderRadius: "8px",
          },
        }}
      >
        {TABS.map(({ value, label, icon: Icon }) => (
          <Tab
            key={value}
            value={value}
            label={label}
            icon={<Icon size={17} aria-hidden />}
            iconPosition="start"
            id={`account-tab-${value}`}
            aria-controls={`account-panel-${value}`}
          />
        ))}
      </Tabs>

      <Box
        role="tabpanel"
        id={`account-panel-${tab}`}
        aria-labelledby={`account-tab-${tab}`}
        sx={{ mt: 3 }}
      >
        {tab === "profile" ? (
          <ProfileTab
            user={user}
            onUpdateProfile={onUpdateProfile}
            onGoToSecurity={() => onTabChange("security")}
          />
        ) : null}
        {tab === "security" ? (
          <SecurityTab
            user={user}
            onSetMfaEnabled={onSetMfaEnabled}
            onChangePassword={onChangePassword}
          />
        ) : null}
        {tab === "notifications" ? (
          <NotificationsTab
            preferences={preferences}
            onUpdatePreferences={onUpdatePreferences}
          />
        ) : null}
        {tab === "preferences" ? (
          <PreferencesTab
            preferences={preferences}
            onUpdatePreferences={onUpdatePreferences}
          />
        ) : null}
      </Box>
    </Box>
  );
}
