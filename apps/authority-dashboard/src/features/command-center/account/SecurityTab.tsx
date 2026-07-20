import { useEffect, useState, type FormEvent, type ReactNode } from "react";
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
import {
  startAuthorityMfaEnrollment,
  syncAuthorityProfile,
  verifyAuthorityMfaEnrollment,
  type AuthoritySession,
  type PasswordChangeResult,
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
  onChangePassword,
}: {
  user: AuthoritySession;
  onChangePassword: (
    current: string,
    next: string,
  ) => Promise<PasswordChangeResult>;
}) {
  const mfaEnabled = Boolean(user.mfaEnabled);
  const [enrolling, setEnrolling] = useState(false);
  const [setup, setSetup] = useState<{
    secret: string;
    otpauthUrl?: string;
  } | null>(null);
  const [enrollPassword, setEnrollPassword] = useState("");
  const [code, setCode] = useState("");
  const [mfaBusy, setMfaBusy] = useState(false);
  const [mfaError, setMfaError] = useState<string | null>(null);
  const [mfaToast, setMfaToast] = useState<string | null>(null);

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [passwordBusy, setPasswordBusy] = useState(false);
  const [passwordResult, setPasswordResult] =
    useState<PasswordChangeResult | null>(null);

  const confirmMismatch = confirm.length > 0 && confirm !== next;

  // The displayed MFA state comes from the directory profile, not a local
  // toggle — refresh it from the auth service when the tab mounts.
  useEffect(() => {
    void syncAuthorityProfile();
  }, []);

  const startEnrolment = async () => {
    if (!enrollPassword) {
      setMfaError("Enter your current password to start enrollment.");
      return;
    }
    setMfaBusy(true);
    setMfaError(null);
    const result = await startAuthorityMfaEnrollment(enrollPassword);
    setMfaBusy(false);
    if (!result.ok || !result.secret) {
      setMfaError(result.message);
      return;
    }
    setSetup({ secret: result.secret, otpauthUrl: result.otpauthUrl });
    setCode("");
  };

  const cancelEnrolment = () => {
    setEnrolling(false);
    setSetup(null);
    setEnrollPassword("");
    setCode("");
    setMfaError(null);
  };

  const verifyMfa = async () => {
    if (code.length !== 6) {
      return;
    }
    setMfaBusy(true);
    setMfaError(null);
    const result = await verifyAuthorityMfaEnrollment(code, enrollPassword);
    setMfaBusy(false);
    if (!result.ok) {
      setMfaError(result.message);
      return;
    }
    setEnrolling(false);
    setSetup(null);
    setEnrollPassword("");
    setCode("");
    setMfaToast(result.message);
  };

  const submitPassword = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (confirmMismatch) {
      setPasswordResult({
        ok: false,
        message: "New password and confirmation do not match.",
      });
      return;
    }
    setPasswordBusy(true);
    const result = await onChangePassword(current, next);
    setPasswordBusy(false);
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
          {mfaError ? (
            <Alert
              severity="error"
              variant="outlined"
              onClose={() => setMfaError(null)}
            >
              {mfaError}
            </Alert>
          ) : null}
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
              {mfaEnabled ? null : !enrolling ? (
                <Button
                  type="button"
                  onClick={() => {
                    setEnrolling(true);
                    setMfaError(null);
                  }}
                  variant="contained"
                  color="primary"
                  size="small"
                  startIcon={<Smartphone size={16} />}
                >
                  Enable MFA
                </Button>
              ) : null}
            </Stack>
            {mfaEnabled ? (
              <MutedNote>
                MFA is required for authority accounts. Your agency
                administrator can reset enrollment if you lose your
                authenticator.
              </MutedNote>
            ) : null}
          </FieldPanel>

          {!mfaEnabled && enrolling && !setup ? (
            <FieldPanel>
              <Stack spacing={1.5}>
                <Box>
                  <Typography
                    sx={{
                      fontWeight: 700,
                      color: "var(--nadaa-ink, #101828)",
                    }}
                  >
                    Confirm it is you
                  </Typography>
                  <MutedNote>
                    Enter your current password to generate an authenticator
                    secret for this account.
                  </MutedNote>
                </Box>
                <TextField
                  label="Current password"
                  type="password"
                  value={enrollPassword}
                  onChange={(event) => setEnrollPassword(event.target.value)}
                  required
                  fullWidth
                  autoComplete="current-password"
                />
                <Stack direction="row" spacing={1.5}>
                  <Button
                    type="button"
                    onClick={() => void startEnrolment()}
                    variant="contained"
                    color="primary"
                    disabled={!enrollPassword || mfaBusy}
                    startIcon={<Smartphone size={16} />}
                  >
                    {mfaBusy ? "Starting" : "Start enrollment"}
                  </Button>
                  <Button
                    type="button"
                    onClick={cancelEnrolment}
                    variant="text"
                    color="inherit"
                    disabled={mfaBusy}
                  >
                    Cancel
                  </Button>
                </Stack>
              </Stack>
            </FieldPanel>
          ) : null}

          {!mfaEnabled && enrolling && setup ? (
            <FieldPanel>
              <Stack spacing={1.5}>
                <Box>
                  <Typography
                    sx={{
                      fontWeight: 700,
                      color: "var(--nadaa-ink, #101828)",
                    }}
                  >
                    Add this secret to your authenticator
                  </Typography>
                  <MutedNote>
                    Scan or paste these details into your authenticator app,
                    then type the current six-digit code to finish enabling
                    MFA.
                  </MutedNote>
                </Box>
                <Box
                  sx={{
                    p: 1.5,
                    borderRadius: "10px",
                    border: "1px dashed var(--nadaa-border, #dfeaf2)",
                    backgroundColor: "var(--nadaa-white, #ffffff)",
                    fontFamily: "monospace",
                    fontSize: "0.95rem",
                    fontWeight: 700,
                    letterSpacing: "0.08em",
                    wordBreak: "break-all",
                    userSelect: "all",
                  }}
                >
                  {setup.secret}
                </Box>
                {setup.otpauthUrl ? (
                  <Box
                    sx={{
                      fontFamily: "monospace",
                      fontSize: "0.75rem",
                      wordBreak: "break-all",
                      userSelect: "all",
                      color: "var(--nadaa-text-secondary, #555b66)",
                    }}
                  >
                    {setup.otpauthUrl}
                  </Box>
                ) : null}
                <OtpInput
                  value={code}
                  onChange={setCode}
                  autoFocus
                  onComplete={() => undefined}
                />
                <Stack direction="row" spacing={1.5}>
                  <Button
                    type="button"
                    onClick={() => void verifyMfa()}
                    variant="contained"
                    color="primary"
                    disabled={code.length !== 6 || mfaBusy}
                    startIcon={<ShieldCheck size={16} />}
                  >
                    {mfaBusy ? "Verifying" : "Verify and enable"}
                  </Button>
                  <Button
                    type="button"
                    onClick={cancelEnrolment}
                    variant="text"
                    color="inherit"
                    disabled={mfaBusy}
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
              disabled={passwordBusy}
              startIcon={<Save size={17} />}
              sx={{ px: 2.5 }}
            >
              {passwordBusy ? "Updating" : "Update password"}
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
