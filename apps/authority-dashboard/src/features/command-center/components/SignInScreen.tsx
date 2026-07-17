import { useState, type FormEvent } from "react";
import {
  Alert,
  Box,
  Button,
  IconButton,
  InputAdornment,
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
  ShieldCheck,
} from "lucide-react";
import { OtpInput } from "./OtpInput";
import type {
  LoginAgencyRequest,
  LoginAgencyResponse,
} from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";
import { signInAuthority, type AuthoritySession } from "@/app/session";

type LoginErrorBody = { error?: { code?: string; message?: string } };

type LoginResult = {
  payload?: LoginAgencyResponse;
  code?: string;
  message?: string;
};

/** POST the auth-service agency login contract; resolves the payload on 200. */
async function agencyLogin(request: LoginAgencyRequest): Promise<LoginResult> {
  const response = await fetch(`${AUTH_API_BASE}/auth/agency/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  if (response.ok) {
    return { payload: (await response.json()) as LoginAgencyResponse };
  }
  try {
    const body = (await response.json()) as LoginErrorBody;
    return {
      code: body.error?.code ?? `http_${response.status}`,
      message: body.error?.message ?? `Sign-in failed (${response.status}).`,
    };
  } catch {
    return {
      code: `http_${response.status}`,
      message: `Sign-in failed (${response.status}).`,
    };
  }
}

/** Map the login error codes operators can act on to clear copy. */
function loginErrorMessage(code: string | undefined, fallback: string) {
  switch (code) {
    case "invalid_credentials":
      return "Email, password, or authenticator code is invalid.";
    case "too_many_attempts":
      return "Too many failed attempts. Wait for the lockout to clear, then try again.";
    case "mfa_setup_required":
      return "MFA enrollment is required before your first sign-in. Ask your agency administrator to complete MFA setup for your account.";
    default:
      return fallback;
  }
}

const assurances = [
  {
    icon: ShieldCheck,
    title: "MFA-enforced access",
    detail: "Two-step verification before any command surface loads.",
  },
  {
    icon: KeyRound,
    title: "Scoped to your desk",
    detail: "Every action is signed with your agency and role.",
  },
  {
    icon: Lock,
    title: "Human-approved alerts",
    detail: "No public broadcast leaves the queue without sign-off.",
  },
];

export function SignInScreen() {
  const [step, setStep] = useState<"credentials" | "mfa">("credentials");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const completeSignIn = (payload: LoginAgencyResponse) => {
    const { user } = payload;
    // Identity (role, agency, district) comes from the directory profile the
    // token was issued for — never from operator-picked form values.
    const session: AuthoritySession = {
      id: user.id,
      name: user.name,
      role: user.role,
      agencyId: user.agency.id,
      agency: user.agency.name,
      district: user.agency.district,
      mfaCompleted: user.mfaEnabled,
      accessToken: payload.accessToken,
      tokenExpiresAt: payload.expiresAt,
      email: user.email,
      mfaEnabled: user.mfaEnabled,
      lastLoginAt: new Date().toISOString(),
    };
    signInAuthority(session);
  };

  const submitCredentials = async (event: FormEvent) => {
    event.preventDefault();
    if (!email.trim()) {
      setError("Enter your agency email.");
      return;
    }
    if (!password) {
      setError("Enter your password to continue.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      const result = await agencyLogin({ email: email.trim(), password });
      if (result.payload) {
        completeSignIn(result.payload);
        return;
      }
      if (result.code === "mfa_required") {
        setStep("mfa");
        return;
      }
      setError(
        loginErrorMessage(
          result.code,
          result.message ?? "Sign-in failed. Try again.",
        ),
      );
    } catch {
      setError("Sign-in needs the auth-service API running.");
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
      const result = await agencyLogin({
        email: email.trim(),
        password,
        mfaCode: code,
      });
      if (result.payload) {
        completeSignIn(result.payload);
        return;
      }
      setError(
        loginErrorMessage(
          result.code,
          result.message ?? "Verification failed. Try again.",
        ),
      );
    } catch {
      setError("Verification needs the auth-service API running.");
    } finally {
      setBusy(false);
    }
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
            <p className="cc-auth__wordmark">NADAA Command</p>
            <p className="cc-auth__org">
              National Disaster Alert &amp; Response Platform
            </p>
          </div>
        </div>

        <div className="cc-auth__pitch">
          <p className="cc-eyebrow cc-eyebrow--inverse">Greater Accra desk</p>
          <h1 className="cc-auth__headline">
            Sign in to run flood and fire response.
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
              className={`cc-auth__step${step === "mfa" ? " is-active" : ""}`}
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
                Authority sign-in
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Use your agency credentials to reach the command center.
              </Typography>

              <TextField
                label="Agency email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                fullWidth
                type="email"
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
                          {showPassword ? (
                            <EyeOff size={17} />
                          ) : (
                            <Eye size={17} />
                          )}
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
                {busy ? "Checking credentials" : "Continue"}
              </Button>

              <p className="cc-auth__hint">
                Access is limited to issued agency accounts. Repeated failed
                attempts trigger a temporary lockout.
              </p>
            </form>
          ) : (
            <form className="cc-auth__form" onSubmit={submitMfa}>
              <Typography variant="h5" className="cc-auth__title">
                Verify it is you
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Enter the 6-digit authenticator code for{" "}
                <strong>{email.trim()}</strong>.
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
                {busy ? "Verifying" : "Verify and enter command"}
              </Button>

              <Button
                type="button"
                variant="text"
                startIcon={<ArrowLeft size={16} />}
                onClick={() => {
                  setStep("credentials");
                  setError("");
                  setCode("");
                }}
                className="cc-auth__back"
                disabled={busy}
              >
                Back to credentials
              </Button>
            </form>
          )}
        </div>
      </section>
    </main>
  );
}
