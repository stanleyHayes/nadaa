import { useState, type ReactNode } from "react";
import {
  Alert,
  Box,
  Button,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { KeyRound, Save, ShieldCheck, Smartphone } from "lucide-react";
import type { AdminSession } from "@/app/session";
import {
  AuthApiError,
  AuthUnavailableError,
  changeAdminPassword,
} from "@/app/auth";
import { OtpInput } from "../components/OtpInput";
import { formatDateTime, SettingCard, StatusChip } from "./primitives";

/** Human-facing message for a failed password change. */
function passwordChangeErrorMessage(error: unknown): string {
  if (error instanceof AuthUnavailableError) {
    return error.message;
  }
  if (error instanceof AuthApiError) {
    if (error.code === "invalid_credentials") {
      return "The current password you entered is incorrect.";
    }
    if (error.code === "weak_password") {
      return error.message;
    }
    if (error.code === "locked" || error.code === "too_many_attempts") {
      return "This account is temporarily locked — wait before trying again.";
    }
    return error.message;
  }
  return "Password change failed. Try again.";
}

function MutedNote({ children }: { children: ReactNode }) {
  return (
    <Typography
      sx={{
        fontSize: "0.85rem",
        lineHeight: 1.5,
        color: "var(--nadaa-text-secondary, #555b66)",
      }}
    >
      {children}
    </Typography>
  );
}

function FieldPanel({ children }: { children: ReactNode }) {
  return (
    <Box
      sx={{
        p: 2,
        border: "1px solid var(--nadaa-border, #dfeaf2)",
        borderRadius: "12px",
        backgroundColor: "var(--nadaa-mist, #f5f8fc)",
      }}
    >
      {children}
    </Box>
  );
}

export function SecurityTab({
  user,
  onSetMfaEnabled,
}: {
  user: AdminSession;
  onSetMfaEnabled: (enabled: boolean) => void;
}) {
  const mfaEnabled = Boolean(user.mfaEnabled);
  const [enrolling, setEnrolling] = useState(false);
  const [code, setCode] = useState("");
  const [toast, setToast] = useState<string | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordBusy, setPasswordBusy] = useState(false);
  const [passwordError, setPasswordError] = useState("");

  const startEnrolment = () => {
    setEnrolling(true);
    setCode("");
  };

  const cancelEnrolment = () => {
    setEnrolling(false);
    setCode("");
  };

  const verifyMfa = () => {
    if (code.length !== 6) {
      return;
    }
    onSetMfaEnabled(true);
    setEnrolling(false);
    setCode("");
    setToast("Multi-factor authentication enabled.");
  };

  const disableMfa = () => {
    onSetMfaEnabled(false);
    setEnrolling(false);
    setCode("");
    setToast("Multi-factor authentication disabled.");
  };

  const submitPasswordChange = async () => {
    if (!currentPassword) {
      setPasswordError("Enter your current password.");
      return;
    }
    if (newPassword.length < 12) {
      setPasswordError("Choose a new password of at least 12 characters.");
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError("The new passwords do not match.");
      return;
    }
    setPasswordError("");
    setPasswordBusy(true);
    try {
      await changeAdminPassword(currentPassword, newPassword);
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setToast("Password updated.");
    } catch (cause) {
      setPasswordError(passwordChangeErrorMessage(cause));
    } finally {
      setPasswordBusy(false);
    }
  };

  return (
    <Stack spacing={2.5}>
      <SettingCard
        icon={ShieldCheck}
        title="Multi-factor authentication"
        description="Protect sign-in with a six-digit authenticator code."
      >
        <Stack spacing={2}>
          <FieldPanel>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              spacing={1.5}
              sx={{
                alignItems: { xs: "flex-start", sm: "center" },
                justifyContent: "space-between"
              }}>
              <Stack direction="row" spacing={1.25} sx={{
                alignItems: "center"
              }}>
                <StatusChip
                  label={mfaEnabled ? "Enabled" : "Not enabled"}
                  tone={mfaEnabled ? "green" : "gold"}
                />
                <MutedNote>
                  {mfaEnabled
                    ? "Authenticator codes are required when you sign in."
                    : "Add a six-digit authenticator code to this account."}
                </MutedNote>
              </Stack>
              {mfaEnabled ? (
                <Button
                  type="button"
                  onClick={disableMfa}
                  variant="outlined"
                  color="error"
                  size="small"
                >
                  Disable
                </Button>
              ) : !enrolling ? (
                <Button
                  type="button"
                  onClick={startEnrolment}
                  variant="contained"
                  color="primary"
                  size="small"
                  startIcon={<Smartphone size={16} />}
                >
                  Enable MFA
                </Button>
              ) : null}
            </Stack>
          </FieldPanel>

          {!mfaEnabled && enrolling ? (
            <FieldPanel>
              <Stack spacing={1.5}>
                <Box>
                  <Typography
                    sx={{
                      fontWeight: 700,
                      color: "var(--nadaa-ink, #101828)",
                    }}
                  >
                    Enter your authenticator code
                  </Typography>
                  <MutedNote>
                    Open your authenticator app and type the current six-digit
                    code to finish enabling MFA.
                  </MutedNote>
                </Box>
                <OtpInput
                  value={code}
                  onChange={setCode}
                  autoFocus
                  onComplete={() => undefined}
                />
                <Stack direction="row" spacing={1.5}>
                  <Button
                    type="button"
                    onClick={verifyMfa}
                    variant="contained"
                    color="primary"
                    disabled={code.length !== 6}
                    startIcon={<ShieldCheck size={16} />}
                  >
                    Verify and enable
                  </Button>
                  <Button
                    type="button"
                    onClick={cancelEnrolment}
                    variant="text"
                    color="inherit"
                  >
                    Cancel
                  </Button>
                </Stack>
              </Stack>
            </FieldPanel>
          ) : null}

          <FieldPanel>
            <Typography
              sx={{
                fontSize: "0.68rem",
                fontWeight: 700,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                color: "var(--nadaa-text-secondary, #555b66)",
              }}
            >
              Last sign in
            </Typography>
            <Typography
              sx={{
                mt: 0.5,
                fontWeight: 700,
                color: "var(--nadaa-ink, #101828)",
              }}
            >
              {formatDateTime(user.lastLoginAt)}
            </Typography>
          </FieldPanel>
        </Stack>
      </SettingCard>
      <SettingCard
        icon={KeyRound}
        title="Password"
        description="Password changes are managed by the auth service."
      >
        <Stack spacing={2}>
          {passwordError ? (
            <Alert severity="error" variant="outlined">
              {passwordError}
            </Alert>
          ) : null}
          <TextField
            label="Current password"
            type="password"
            size="small"
            fullWidth
            autoComplete="current-password"
            value={currentPassword}
            onChange={(event) => setCurrentPassword(event.target.value)}
          />
          <TextField
            label="New password"
            type="password"
            size="small"
            fullWidth
            autoComplete="new-password"
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
          />
          <TextField
            label="Confirm new password"
            type="password"
            size="small"
            fullWidth
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
          />
          <Box>
            <Button
              type="button"
              variant="contained"
              color="primary"
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
              disabled={passwordBusy}
              onClick={() => void submitPasswordChange()}
            >
              {passwordBusy ? "Updating" : "Update password"}
            </Button>
          </Box>
        </Stack>
      </SettingCard>
      <Snackbar
        open={Boolean(toast)}
        autoHideDuration={3000}
        onClose={() => setToast(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          onClose={() => setToast(null)}
          severity="success"
          variant="filled"
          sx={{ width: "100%" }}
        >
          {toast}
        </Alert>
      </Snackbar>
    </Stack>
  );
}
