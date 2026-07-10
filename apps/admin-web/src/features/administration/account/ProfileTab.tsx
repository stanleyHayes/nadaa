import { useState, type FormEvent } from "react";
import {
  Alert,
  Box,
  Button,
  MenuItem,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  Building2,
  CalendarClock,
  Mail,
  Save,
  ShieldCheck,
  UserRound,
} from "lucide-react";
import type { AdminProfilePatch, AdminSession } from "@/app/session";
import { roleLabel } from "../utils";
import {
  formatDateTime,
  InfoRow,
  SettingCard,
  StatusChip,
} from "./primitives";

const DEPARTMENT_OPTIONS = [
  "Platform Administration",
  "Agency Administration",
  "Flood Operations",
  "District Coordination",
  "Dispatch Operations",
  "Field Response",
  "Emergency Communications",
  "Logistics & Relief",
  "Partner Liaison",
];

export function ProfileTab({
  user,
  onUpdateProfile,
  onGoToSecurity,
}: {
  user: AdminSession;
  onUpdateProfile: (patch: AdminProfilePatch) => void;
  onGoToSecurity: () => void;
}) {
  const [name, setName] = useState(user.name);
  const [department, setDepartment] = useState(
    user.department ?? DEPARTMENT_OPTIONS[0],
  );
  const [saved, setSaved] = useState(false);

  const departmentOptions = DEPARTMENT_OPTIONS.includes(department)
    ? DEPARTMENT_OPTIONS
    : [department, ...DEPARTMENT_OPTIONS];

  const nameError = name.trim().length === 0;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (nameError) {
      return;
    }
    onUpdateProfile({ name: name.trim(), department });
    setSaved(true);
  };

  const mfaEnabled = Boolean(user.mfaEnabled);

  return (
    <Box
      sx={{
        display: "grid",
        gap: 2.5,
        alignItems: "start",
        gridTemplateColumns: { xs: "1fr", lg: "1.05fr 0.95fr" },
      }}
    >
      <SettingCard
        icon={UserRound}
        title="Edit profile"
        description="Keep your display name and department current."
      >
        <Box component="form" onSubmit={handleSubmit} sx={{ display: "grid", gap: 2.25 }}>
          <TextField
            label="Full name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            required
            fullWidth
            error={nameError}
            helperText={nameError ? "Enter your full name." : " "}
            autoComplete="name"
          />
          <TextField
            label="Email"
            value={user.email ?? ""}
            fullWidth
            disabled
            helperText="Managed by your administrator."
            autoComplete="email"
          />
          <TextField
            label="Department"
            value={department}
            onChange={(event) => setDepartment(event.target.value)}
            select
            fullWidth
            helperText=" "
          >
            {departmentOptions.map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </TextField>
          <Box>
            <Button
              type="submit"
              variant="contained"
              color="primary"
              disabled={nameError}
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
            >
              Save profile
            </Button>
          </Box>
        </Box>
      </SettingCard>

      <Stack spacing={2.5}>
        <SettingCard
          icon={ShieldCheck}
          title={user.name}
          description={roleLabel(user.role)}
        >
          <Stack spacing={1.5}>
            <InfoRow icon={Mail} label="Email">
              {user.email}
            </InfoRow>
            <InfoRow icon={ShieldCheck} label="Account status">
              <StatusChip
                label={user.status === "suspended" ? "Suspended" : "Active"}
                tone={user.status === "suspended" ? "red" : "green"}
              />
            </InfoRow>
            <InfoRow icon={ShieldCheck} label="Multi-factor auth">
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                flexWrap="wrap"
                useFlexGap
              >
                <StatusChip
                  label={mfaEnabled ? "Enabled" : "Not enabled"}
                  tone={mfaEnabled ? "green" : "gold"}
                />
                {!mfaEnabled ? (
                  <Button
                    type="button"
                    onClick={onGoToSecurity}
                    size="small"
                    variant="text"
                    sx={{
                      minHeight: "auto",
                      px: 0.5,
                      color: "var(--nadaa-navy, #0d1b3d)",
                      fontWeight: 700,
                    }}
                  >
                    Enable in settings
                  </Button>
                ) : null}
              </Stack>
            </InfoRow>
            <InfoRow icon={Building2} label="Agency">
              {user.agency}
            </InfoRow>
            <InfoRow icon={CalendarClock} label="Last sign in">
              {formatDateTime(user.lastLoginAt)}
            </InfoRow>
          </Stack>
          <Typography
            sx={{
              mt: 2,
              fontSize: "0.78rem",
              color: "var(--nadaa-text-secondary, #555b66)",
            }}
          >
            Access is assigned from your role. Contact a system administrator to
            change your agency.
          </Typography>
        </SettingCard>
      </Stack>

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
          Profile saved.
        </Alert>
      </Snackbar>
    </Box>
  );
}
