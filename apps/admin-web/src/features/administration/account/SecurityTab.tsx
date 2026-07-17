import { useState, type ReactNode } from "react";
import {
  Alert,
  Box,
  Button,
  Snackbar,
  Stack,
  Typography,
} from "@mui/material";
import { KeyRound, Save, ShieldCheck, Smartphone } from "lucide-react";
import type { AdminSession } from "@/app/session";
import { OtpInput } from "../components/OtpInput";
import { formatDateTime, SettingCard, StatusChip } from "./primitives";

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
  const [mfaToast, setMfaToast] = useState<string | null>(null);

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
    setMfaToast("Multi-factor authentication enabled.");
  };

  const disableMfa = () => {
    onSetMfaEnabled(false);
    setEnrolling(false);
    setCode("");
    setMfaToast("Multi-factor authentication disabled.");
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
          <Alert severity="info" variant="outlined">
            Password change is not available in this release: the auth service
            does not expose a password-change endpoint yet. If you need a new
            password, ask a system administrator to re-provision your account.
          </Alert>
          <Box>
            <Button
              type="button"
              variant="contained"
              color="primary"
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
              disabled
              title="Unavailable in this release"
            >
              Update password
            </Button>
          </Box>
        </Stack>
      </SettingCard>
      <Snackbar
        open={Boolean(mfaToast)}
        autoHideDuration={3000}
        onClose={() => setMfaToast(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          onClose={() => setMfaToast(null)}
          severity="success"
          variant="filled"
          sx={{ width: "100%" }}
        >
          {mfaToast}
        </Alert>
      </Snackbar>
    </Stack>
  );
}
