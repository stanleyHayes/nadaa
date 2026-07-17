import { useState, type FormEvent } from "react";
import {
  Alert,
  Box,
  Button,
  IconButton,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  ArrowLeft,
  ArrowRight,
  AtSign,
  Eye,
  EyeOff,
  KeyRound,
  Loader2,
  Lock,
  ScrollText,
  ShieldCheck,
  Smartphone,
} from "lucide-react";
import { OtpInput } from "./OtpInput";
import type {
  AgencyMFASetupResponse,
  LoginAgencyResponse,
} from "@nadaa/shared-types";
import {
  AuthApiError,
  AuthUnavailableError,
  loginAgency,
  setupAgencyMfa,
  verifyAgencyMfa,
} from "@/app/auth";
import { signInAdmin, type AdminSession } from "@/app/session";

type Step = "credentials" | "mfa" | "mfaSetup";

function cap(word: string) {
  return word ? word[0].toUpperCase() + word.slice(1) : word;
}

function displayName(identifier: string) {
  const trimmed = identifier.trim();
  if (!trimmed) {
    return "Administrator";
  }
  if (trimmed.includes("@")) {
    return (
      trimmed
        .split("@")[0]
        .split(/[._-]+/)
        .filter(Boolean)
        .map(cap)
        .join(" ") || "Administrator"
    );
  }
  return trimmed;
}

/** Human-facing message for a failed credentials or verification attempt. */
function signInErrorMessage(error: unknown): string {
  if (error instanceof AuthUnavailableError) {
    return error.message;
  }
  if (error instanceof AuthApiError) {
    if (error.code === "too_many_attempts") {
      return "Too many failed attempts. This account is temporarily locked — wait before trying again.";
    }
    if (error.code === "invalid_credentials") {
      return "Email or password is incorrect.";
    }
    return error.message;
  }
  return "Sign-in failed. Try again.";
}

const assurances = [
  {
    icon: ShieldCheck,
    title: "MFA-enforced console",
    detail: "Two-step verification before any governance surface loads.",
  },
  {
    icon: KeyRound,
    title: "Scoped to your role",
    detail: "Only system and agency admins reach platform settings.",
  },
  {
    icon: ScrollText,
    title: "Every action audited",
    detail: "User, alert, and integration changes are traced for review.",
  },
];

export function SignInScreen() {
  const [step, setStep] = useState<Step>("credentials");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [setupUserId, setSetupUserId] = useState("");
  const [setupChallenge, setSetupChallenge] =
    useState<AgencyMFASetupResponse | null>(null);

  /** Open the console with the identity and token from the login response. */
  const establishSession = (payload: LoginAgencyResponse) => {
    const session: AdminSession = {
      id: payload.user.id,
      name: payload.user.name,
      role: payload.user.role,
      agencyId: payload.user.agency.id,
      agency: payload.user.agency.name,
      mfaCompleted: payload.user.mfaEnabled,
      accessToken: payload.accessToken,
      tokenExpiresAt: payload.expiresAt,
      email: payload.user.email,
      mfaEnabled: payload.user.mfaEnabled,
      lastLoginAt: new Date().toISOString(),
    };
    signInAdmin(session);
  };

  /**
   * Route a failed login: MFA-enrolled accounts move to code entry, accounts
   * that must enrol move to the setup walkthrough, everything else surfaces
   * an error message on the current step.
   */
  const handleLoginFailure = (cause: unknown) => {
    if (cause instanceof AuthApiError && cause.code === "mfa_required") {
      setError("");
      setCode("");
      setStep("mfa");
      return;
    }
    if (cause instanceof AuthApiError && cause.code === "mfa_setup_required") {
      setError("");
      setCode("");
      setSetupChallenge(null);
      setStep("mfaSetup");
      return;
    }
    setError(signInErrorMessage(cause));
  };

  const submitCredentials = async (event: FormEvent) => {
    event.preventDefault();
    const trimmed = email.trim();
    if (!trimmed.includes("@")) {
      setError("Enter your agency email address.");
      return;
    }
    if (!password) {
      setError("Enter your password to continue.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      establishSession(await loginAgency(trimmed, password));
    } catch (cause) {
      handleLoginFailure(cause);
    } finally {
      setBusy(false);
    }
  };

  const submitMfa = async (event: FormEvent) => {
    event.preventDefault();
    if (!/^\d{6}$/.test(code)) {
      setError("Enter the 6-digit code from your authenticator.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      establishSession(await loginAgency(email.trim(), password, code));
    } catch (cause) {
      if (cause instanceof AuthApiError && cause.code === "mfa_setup_required") {
        handleLoginFailure(cause);
      } else if (
        cause instanceof AuthApiError &&
        cause.code === "invalid_credentials"
      ) {
        setError("That authenticator code was not accepted. Try again.");
      } else {
        setError(signInErrorMessage(cause));
      }
    } finally {
      setBusy(false);
    }
  };

  const startMfaSetup = async (event: FormEvent) => {
    event.preventDefault();
    if (!setupUserId.trim()) {
      setError("Enter the account user ID from your provisioning administrator.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      const challenge = await setupAgencyMfa(
        setupUserId.trim(),
        email.trim(),
        password,
      );
      setSetupChallenge(challenge);
      setCode("");
    } catch (cause) {
      if (cause instanceof AuthApiError && cause.code === "mfa_already_enabled") {
        setError("");
        setCode("");
        setStep("mfa");
        return;
      }
      if (cause instanceof AuthApiError && cause.code === "invalid_credentials") {
        setError(
          "Account ID, email, or temporary password did not match our records.",
        );
      } else {
        setError(signInErrorMessage(cause));
      }
    } finally {
      setBusy(false);
    }
  };

  const completeMfaSetup = async (event: FormEvent) => {
    event.preventDefault();
    if (!/^\d{6}$/.test(code)) {
      setError("Enter the 6-digit setup code.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      await verifyAgencyMfa(setupUserId.trim(), email.trim(), password, code);
      // MFA is now enabled; the verified code signs the admin straight in.
      establishSession(await loginAgency(email.trim(), password, code));
    } catch (cause) {
      if (cause instanceof AuthApiError && cause.code === "invalid_credentials") {
        setError("That setup code was not accepted. Request a new one or retry.");
      } else {
        setError(signInErrorMessage(cause));
      }
    } finally {
      setBusy(false);
    }
  };

  const backToCredentials = () => {
    setStep("credentials");
    setError("");
    setCode("");
    setSetupChallenge(null);
  };

  return (
    <main className="cc-auth" id="main-content">
      <section className="cc-auth__brand" aria-hidden={false}>
        <div className="cc-auth__brand-top">
          <Box
            component="img"
            src="/brand/nadaa-logo.png"
            alt=""
            className="cc-auth__logo"
          />
          <div>
            <p className="cc-auth__wordmark">NADAA Governance</p>
            <p className="cc-auth__org">
              National Disaster Alert &amp; Response Platform
            </p>
          </div>
        </div>

        <div className="cc-auth__pitch">
          <p className="cc-eyebrow cc-eyebrow--inverse">Admin console</p>
          <h1 className="cc-auth__headline">
            Sign in to govern agencies, access, and alerts.
          </h1>
          <p className="cc-auth__slogan">Be aware. Be prepared. Be safe.</p>
        </div>

        <ul className="cc-auth__assurances">
          {assurances.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.title}>
                <span className="cc-auth__assurance-icon" aria-hidden>
                  <Icon size={18} />
                </span>
                <span>
                  <span className="cc-auth__assurance-title">{item.title}</span>
                  <span className="cc-auth__assurance-detail">
                    {item.detail}
                  </span>
                </span>
              </li>
            );
          })}
        </ul>
      </section>
      <section className="cc-auth__panel">
        <div className="cc-auth__card">
          <div className="cc-auth__steps" aria-hidden>
            <span
              className={`cc-auth__step${step === "credentials" ? " is-active" : " is-done"}`}
            >
              1 · Credentials
            </span>
            <span className="cc-auth__step-rule" />
            <span
              className={`cc-auth__step${step !== "credentials" ? " is-active" : ""}`}
            >
              2 · Verify
            </span>
          </div>

          {error ? (
            <Alert severity="error" className="cc-auth__error">
              {error}
            </Alert>
          ) : null}

          {step === "credentials" ? (
            <form className="cc-auth__form" onSubmit={submitCredentials}>
              <Typography variant="h5" className="cc-auth__title">
                Admin sign-in
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Use your agency account to reach the governance console. Your
                role loads with your profile.
              </Typography>

              <TextField
                label="Agency email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                fullWidth
                autoComplete="username"
                autoFocus
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <AtSign size={17} />
                      </InputAdornment>
                    ),
                  }
                }}
              />

              <TextField
                label="Password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                type={showPassword ? "text" : "password"}
                fullWidth
                autoComplete="current-password"
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <Lock size={17} />
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowPassword((value) => !value)}
                          edge="end"
                          aria-label={
                            showPassword ? "Hide password" : "Show password"
                          }
                          size="small"
                        >
                          {showPassword ? <EyeOff size={17} /> : <Eye size={17} />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }
                }}
              />

              <Button
                type="submit"
                variant="contained"
                size="large"
                disabled={busy}
                endIcon={
                  busy ? (
                    <Loader2 size={18} className="spin-icon" />
                  ) : (
                    <ArrowRight size={18} />
                  )
                }
                className="cc-auth__submit"
              >
                {busy ? "Signing in" : "Continue"}
              </Button>

              <p className="cc-auth__hint">
                Newly provisioned account? Sign in with your temporary password
                to enrol an authenticator.
              </p>
            </form>
          ) : null}

          {step === "mfa" ? (
            <form className="cc-auth__form" onSubmit={submitMfa}>
              <Typography variant="h5" className="cc-auth__title">
                Verify it is you
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Enter the 6-digit authenticator code for{" "}
                <strong>{displayName(email)}</strong>.
              </Typography>

              <Box>
                <Typography
                  component="label"
                  variant="subtitle2"
                  sx={{ display: "block", mb: 1, fontWeight: 700 }}
                >
                  Authenticator code
                </Typography>
                <OtpInput
                  autoFocus
                  ariaDescribedBy="cc-auth-mfa-hint"
                  onChange={setCode}
                  value={code}
                />
              </Box>
              <p id="cc-auth-mfa-hint" className="cc-auth__hint">
                Codes rotate every 30 seconds in your authenticator app.
              </p>

              <Button
                type="submit"
                variant="contained"
                size="large"
                disabled={busy}
                startIcon={
                  busy ? (
                    <Loader2 size={18} className="spin-icon" />
                  ) : (
                    <ShieldCheck size={18} />
                  )
                }
                className="cc-auth__submit"
              >
                {busy ? "Verifying" : "Verify and enter console"}
              </Button>

              <Button
                type="button"
                variant="text"
                startIcon={<ArrowLeft size={16} />}
                onClick={backToCredentials}
                className="cc-auth__back"
                disabled={busy}
              >
                Back to credentials
              </Button>
            </form>
          ) : null}

          {step === "mfaSetup" ? (
            <form
              className="cc-auth__form"
              onSubmit={setupChallenge ? completeMfaSetup : startMfaSetup}
            >
              <Typography variant="h5" className="cc-auth__title">
                Set up two-step verification
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                This account must enrol an authenticator before its first
                sign-in. The password you entered is your temporary password.
              </Typography>

              <TextField
                label="Account user ID"
                value={setupUserId}
                onChange={(event) => setSetupUserId(event.target.value)}
                fullWidth
                autoFocus={!setupChallenge}
                disabled={Boolean(setupChallenge)}
                helperText="Provided by the administrator who created your account (usr_…)."
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <Smartphone size={17} />
                      </InputAdornment>
                    ),
                  }
                }}
              />

              {setupChallenge ? (
                <>
                  <Alert severity="info">
                    Add this setup key to your authenticator app, then enter the
                    6-digit code it shows:{" "}
                    <strong>{setupChallenge.secret}</strong>
                  </Alert>
                  {setupChallenge.devCode ? (
                    <Alert severity="warning">
                      Development build — setup code:{" "}
                      <strong>{setupChallenge.devCode}</strong>
                    </Alert>
                  ) : null}
                  <Box>
                    <Typography
                      component="label"
                      variant="subtitle2"
                      sx={{ display: "block", mb: 1, fontWeight: 700 }}
                    >
                      Setup code
                    </Typography>
                    <OtpInput
                      autoFocus
                      ariaDescribedBy="cc-auth-setup-hint"
                      onChange={setCode}
                      value={code}
                    />
                  </Box>
                  <p id="cc-auth-setup-hint" className="cc-auth__hint">
                    The setup challenge expires 10 minutes after it was issued.
                  </p>
                </>
              ) : null}

              <Button
                type="submit"
                variant="contained"
                size="large"
                disabled={busy}
                startIcon={
                  busy ? (
                    <Loader2 size={18} className="spin-icon" />
                  ) : (
                    <ShieldCheck size={18} />
                  )
                }
                className="cc-auth__submit"
              >
                {busy
                  ? "Working"
                  : setupChallenge
                    ? "Verify and sign in"
                    : "Start MFA setup"}
              </Button>

              {setupChallenge ? (
                <Button
                  type="button"
                  variant="text"
                  onClick={() => {
                    setSetupChallenge(null);
                    setCode("");
                    setError("");
                  }}
                  className="cc-auth__back"
                  disabled={busy}
                >
                  Request a new setup code
                </Button>
              ) : null}

              <Button
                type="button"
                variant="text"
                startIcon={<ArrowLeft size={16} />}
                onClick={backToCredentials}
                className="cc-auth__back"
                disabled={busy}
              >
                Back to credentials
              </Button>
            </form>
          ) : null}
        </div>
      </section>
    </main>
  );
}
