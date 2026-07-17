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
  Loader2,
  Lock,
  Radio,
  ShieldCheck,
  Truck,
} from "lucide-react";
import { OtpInput } from "./OtpInput";
import type { LoginAgencyResponse } from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";
import {
  agencyRoles,
  signInAgency,
  type AgencySession,
} from "@/app/session";

const assurances = [
  {
    icon: ShieldCheck,
    title: "MFA-enforced access",
    detail: "Two-step verification before any response surface loads.",
  },
  {
    icon: Truck,
    title: "Scoped to your desk",
    detail: "Every action is signed with your agency, role, and district.",
  },
  {
    icon: Radio,
    title: "Coordinated response",
    detail: "Assignments, capacity, and relief stay in sync with command.",
  },
];

/** Read the structured `{error: {code, message}}` body the services return. */
async function readAuthError(
  response: Response,
): Promise<{ code: string; message: string }> {
  try {
    const body = (await response.json()) as {
      error?: { code?: string; message?: string };
    };
    return {
      code: body.error?.code ?? "",
      message:
        body.error?.message ?? `Sign-in failed (${response.status}).`,
    };
  } catch {
    return { code: "", message: `Sign-in failed (${response.status}).` };
  }
}

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
    if (!agencyRoles.includes(user.role)) {
      setError("Your account role is not permitted on the agency console.");
      return;
    }
    if (!user.mfaEnabled) {
      // A token without the MFA claim is rejected by every authority
      // endpoint, so there is no usable session to enter with.
      setError(
        "MFA is not enabled on this account. Contact your agency administrator to complete enrollment.",
      );
      return;
    }
    const session: AgencySession = {
      id: user.id,
      name: user.name,
      role: user.role,
      agencyId: user.agency.id,
      agency: user.agency.name,
      district: user.agency.district,
      token: payload.accessToken,
      mfaCompleted: true,
      email: user.email,
      mfaEnabled: user.mfaEnabled,
      lastLoginAt: new Date().toISOString(),
    };
    signInAgency(session);
  };

  const handleAuthFailure = (status: number, code: string, message: string) => {
    if (status === 401 && code === "mfa_required") {
      setError("");
      setStep("mfa");
      return;
    }
    if (status === 403 && code === "mfa_setup_required") {
      setError(
        "MFA setup is required for this account. Contact your agency administrator to complete enrollment, then sign in.",
      );
      return;
    }
    if (status === 429) {
      setError("Too many failed attempts. Wait a few minutes and try again.");
      return;
    }
    if (status === 401) {
      setError("Invalid email, password, or authenticator code.");
      return;
    }
    setError(message);
  };

  const submitLogin = async (mfaCode: string) => {
    setBusy(true);
    setError("");
    try {
      const response = await fetch(`${AUTH_API_BASE}/auth/agency/login`, {
        body: JSON.stringify({
          email: email.trim(),
          password,
          mfaCode,
        }),
        headers: { "Content-Type": "application/json" },
        method: "POST",
      });
      if (response.ok) {
        completeSignIn((await response.json()) as LoginAgencyResponse);
        return;
      }
      const failure = await readAuthError(response);
      handleAuthFailure(response.status, failure.code, failure.message);
    } catch {
      setError("Could not reach the sign-in service. Check your connection.");
    } finally {
      setBusy(false);
    }
  };

  const submitCredentials = (event: FormEvent) => {
    event.preventDefault();
    if (!email.trim()) {
      setError("Enter your agency email.");
      return;
    }
    if (!password) {
      setError("Enter your password to continue.");
      return;
    }
    void submitLogin("");
  };

  const submitMfa = (event: FormEvent) => {
    event.preventDefault();
    if (!/^\d{6}$/.test(code)) {
      setError("Enter the 6-digit code from your authenticator.");
      return;
    }
    void submitLogin(code);
  };

  return (
    <main className="cc-auth" id="main-content">
      <section className="cc-auth__brand" aria-hidden={false}>
        <div className="cc-auth__brand-top">
          <span className="cc-chip cc-auth__brand-chip" aria-hidden>
            <Box
              component="img"
              src="/brand/nadaa-logo.png"
              alt=""
              className="cc-auth__logo"
            />
          </span>
          <div>
            <p className="cc-auth__wordmark">NADAA Agency</p>
            <p className="cc-auth__org">
              National Disaster Alert &amp; Response Platform
            </p>
          </div>
        </div>

        <div className="cc-auth__pitch">
          <p className="cc-eyebrow cc-eyebrow--inverse">Greater Accra desk</p>
          <h1 className="cc-auth__headline">
            Sign in to run field response operations.
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
                Agency sign-in
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Use your agency credentials to reach response operations.
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
                {busy ? "Signing in" : "Continue"}
              </Button>

              <p className="cc-auth__hint">
                Access is limited to enrolled agency accounts. Your agency
                administrator provisions credentials and MFA.
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
                {busy ? "Verifying" : "Verify and enter operations"}
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
