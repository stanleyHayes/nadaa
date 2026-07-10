import { useState, type FormEvent, type ReactNode } from "react";
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
import type {
  AuthoritySession,
  PasswordChangeResult,
} from "@/app/session";
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
  onChangePassword,
}: {
  user: AuthoritySession;
  onSetMfaEnabled: (enabled: boolean) => void;
  onChangePassword: (current: string, next: string) => PasswordChangeResult;
}) {
  const mfaEnabled = Boolean(user.mfaEnabled);
  const [enrolling, setEnrolling] = useState(false);
  const [code, setCode] = useState("");
  const [mfaToast, setMfaToast] = useState<string | null>(null);

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [passwordResult, setPasswordResult] =
    useState<PasswordChangeResult | null>(null);

  const confirmMismatch = confirm.length > 0 && confirm !== next;

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

  const submitPassword = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (confirmMismatch) {
      setPasswordResult({
        ok: false,
        message: "New password and confirmation do not match.",
      });
      return;
    }
    const result = onChangePassword(current, next);
    setPasswordResult(result);
    if (result.ok) {
      setCurrent("");
      setNext("");
      setConfirm("");
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
        description="Change your password using your current one."
      >
        <Box
          component="form"
          onSubmit={submitPassword}
          sx={{ display: "grid", gap: 2.25 }}
        >
          {passwordResult ? (
            <Alert
              severity={passwordResult.ok ? "success" : "error"}
              variant="outlined"
              onClose={() => setPasswordResult(null)}
            >
              {passwordResult.message}
            </Alert>
          ) : null}
          <TextField
            label="Current password"
            type="password"
            value={current}
            onChange={(event) => setCurrent(event.target.value)}
            required
            fullWidth
            autoComplete="current-password"
          />
          <TextField
            label="New password"
            type="password"
            value={next}
            onChange={(event) => setNext(event.target.value)}
            required
            fullWidth
            autoComplete="new-password"
            helperText="Use at least 8 characters."
          />
          <TextField
            label="Confirm new password"
            type="password"
            value={confirm}
            onChange={(event) => setConfirm(event.target.value)}
            required
            fullWidth
            autoComplete="new-password"
            error={confirmMismatch}
            helperText={confirmMismatch ? "Passwords do not match." : " "}
          />
          <Box>
            <Button
              type="submit"
              variant="contained"
              color="primary"
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
            >
              Update password
            </Button>
          </Box>
        </Box>
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
