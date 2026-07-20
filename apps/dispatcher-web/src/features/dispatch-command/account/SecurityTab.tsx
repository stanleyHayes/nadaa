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
  DispatcherSession,
  PasswordChangeResult,
} from "@/app/session";
import { AUTH_API_BASE } from "@/app/config";
import { OtpInput } from "../components/OtpInput";
import { formatDateTime, SettingCard, StatusChip } from "./primitives";

type AuthErrorBody = { error?: { code?: string; message?: string } };

/** auth-service error carrying the machine-readable code from the body. */
class MfaApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

async function postAgencyMfa<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`${AUTH_API_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    const parsed = (await response
      .json()
      .catch(() => null)) as AuthErrorBody | null;
    throw new MfaApiError(
      response.status,
      parsed?.error?.code ?? "",
      parsed?.error?.message ?? "",
    );
  }
  return (await response.json()) as T;
}

/** TOTP enrolment material returned by the MFA setup endpoint. */
type MfaSetupState = {
  secret: string;
  otpauthUrl: string;
};

function mfaErrorMessage(error: unknown): string {
  if (error instanceof MfaApiError) {
    if (error.code === "too_many_attempts") {
      return "Too many failed attempts. This account is temporarily locked — try again later.";
    }
    if (error.code === "mfa_already_enabled") {
      return "MFA is already enabled on this account.";
    }
    if (error.status === 401 || error.code === "invalid_credentials") {
      return "Password or authenticator code is incorrect.";
    }
    return error.message || "MFA enrolment failed. Try again.";
  }
  return "Auth service unavailable. Check your connection and try again.";
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
  onChangePassword,
}: {
  user: DispatcherSession;
  onSetMfaEnabled: (enabled: boolean) => void;
  onChangePassword: (
    current: string,
    next: string,
  ) => Promise<PasswordChangeResult>;
}) {
  // Displayed state derives from the server profile captured at sign-in and is
  // only updated here after the MFA verify endpoint confirms enrolment.
  const mfaEnabled = Boolean(user.mfaEnabled);
  const [enrolling, setEnrolling] = useState(false);
  const [password, setPassword] = useState("");
  const [setup, setSetup] = useState<MfaSetupState | null>(null);
  const [code, setCode] = useState("");
  const [mfaBusy, setMfaBusy] = useState(false);
  const [mfaError, setMfaError] = useState("");
  const [mfaToast, setMfaToast] = useState<string | null>(null);

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [passwordBusy, setPasswordBusy] = useState(false);
  const [passwordResult, setPasswordResult] =
    useState<PasswordChangeResult | null>(null);

  const confirmMismatch = confirm.length > 0 && confirm !== next;

  const startEnrolment = () => {
    setEnrolling(true);
    setPassword("");
    setSetup(null);
    setCode("");
    setMfaError("");
  };

  const cancelEnrolment = () => {
    setEnrolling(false);
    setPassword("");
    setSetup(null);
    setCode("");
    setMfaError("");
  };

  // Step 1: re-authenticate with the current password; the service answers
  // with the TOTP secret and otpauth URL to add to an authenticator app.
  const beginSetup = async () => {
    if (!password) {
      setMfaError("Enter your current password to start enrolment.");
      return;
    }
    setMfaBusy(true);
    setMfaError("");
    try {
      const response = await postAgencyMfa<{
        secret: string;
        otpauthUrl?: string;
      }>(`/auth/agency-users/${encodeURIComponent(user.id)}/mfa/setup`, {
        email: user.email ?? "",
        temporaryPassword: password,
      });
      setSetup({
        secret: response.secret,
        otpauthUrl: response.otpauthUrl ?? "",
      });
      setCode("");
    } catch (error) {
      setMfaError(mfaErrorMessage(error));
    } finally {
      setMfaBusy(false);
    }
  };

  // Step 2: the authenticator's current code proves enrolment to the service.
  const verifyMfa = async () => {
    if (code.length !== 6 || !setup) {
      return;
    }
    setMfaBusy(true);
    setMfaError("");
    try {
      const response = await postAgencyMfa<{
        user: { mfaEnabled: boolean };
      }>(`/auth/agency-users/${encodeURIComponent(user.id)}/mfa/verify`, {
        email: user.email ?? "",
        temporaryPassword: password,
        code,
      });
      onSetMfaEnabled(response.user.mfaEnabled);
      cancelEnrolment();
      setMfaToast("Multi-factor authentication enabled.");
    } catch (error) {
      setMfaError(mfaErrorMessage(error));
    } finally {
      setMfaBusy(false);
    }
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
                    ? "Authenticator codes are required when you sign in. Contact your agency administrator to reset your authenticator."
                    : "Add a six-digit authenticator code to this account."}
                </MutedNote>
              </Stack>
              {!mfaEnabled && !enrolling ? (
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
                {mfaError ? (
                  <Alert severity="error" variant="outlined">
                    {mfaError}
                  </Alert>
                ) : null}
                {!setup ? (
                  <>
                    <Box>
                      <Typography
                        sx={{
                          fontWeight: 700,
                          color: "var(--nadaa-ink, #101828)",
                        }}
                      >
                        Confirm your password
                      </Typography>
                      <MutedNote>
                        Re-authenticate to generate an authenticator secret for
                        this account.
                      </MutedNote>
                    </Box>
                    <TextField
                      label="Current password"
                      type="password"
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                      required
                      fullWidth
                      size="small"
                      autoComplete="current-password"
                    />
                    <Stack direction="row" spacing={1.5}>
                      <Button
                        type="button"
                        onClick={beginSetup}
                        variant="contained"
                        color="primary"
                        disabled={mfaBusy || !password}
                        startIcon={<Smartphone size={16} />}
                      >
                        Continue
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
                  </>
                ) : (
                  <>
                    <Box>
                      <Typography
                        sx={{
                          fontWeight: 700,
                          color: "var(--nadaa-ink, #101828)",
                        }}
                      >
                        Add this key to your authenticator app
                      </Typography>
                      <MutedNote>
                        Enter the setup key below in your authenticator app (or
                        open the otpauth link on this device), then type the
                        current six-digit code to finish enabling MFA.
                      </MutedNote>
                    </Box>
                    <Box
                      sx={{
                        p: 1.5,
                        border: "1px solid var(--nadaa-border, #dfeaf2)",
                        borderRadius: "10px",
                        backgroundColor: "var(--nadaa-white, #ffffff)",
                        fontFamily: "monospace",
                        fontSize: "0.95rem",
                        letterSpacing: "0.04em",
                        wordBreak: "break-all",
                      }}
                    >
                      {setup.secret}
                    </Box>
                    {setup.otpauthUrl ? (
                      <Typography
                        sx={{
                          fontSize: "0.78rem",
                          color: "var(--nadaa-text-secondary, #555b66)",
                          wordBreak: "break-all",
                        }}
                      >
                        {setup.otpauthUrl}
                      </Typography>
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
                        onClick={verifyMfa}
                        variant="contained"
                        color="primary"
                        disabled={mfaBusy || code.length !== 6}
                        startIcon={<ShieldCheck size={16} />}
                      >
                        Verify and enable
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
                  </>
                )}
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
