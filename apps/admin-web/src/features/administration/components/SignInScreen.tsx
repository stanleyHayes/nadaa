import { useState, type FormEvent } from "react";
import {
  Alert,
  Box,
  Button,
  IconButton,
  InputAdornment,
  MenuItem,
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
} from "lucide-react";
import { OtpInput } from "./OtpInput";
import type { AgencyUserRole } from "@nadaa/shared-types";
import {
  adminSignInRoles,
  agencyByRole,
  signInAdmin,
  type AdminSession,
} from "@/app/session";
import { roleLabel } from "../utils";

const DEFAULT_AGENCY_ID = "00000000-0000-0000-0000-000000000101";

function cap(word: string) {
  return word ? word[0].toUpperCase() + word.slice(1) : word;
}

function displayName(identifier: string, role: AgencyUserRole) {
  const trimmed = identifier.trim();
  if (!trimmed) {
    return roleLabel(role);
  }
  if (trimmed.includes("@")) {
    return (trimmed
      .split("@")[0]
      .split(/[._-]+/)
      .filter(Boolean)
      .map(cap)
      .join(" ") || roleLabel(role));
  }
  return trimmed;
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
  const [step, setStep] = useState<"credentials" | "mfa">("credentials");
  const [identifier, setIdentifier] = useState("system.admin@nadaa.gov.gh");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [role, setRole] = useState<AgencyUserRole>("system_admin");
  const [agency, setAgency] = useState(agencyByRole.system_admin);
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const onRoleChange = (nextRole: AgencyUserRole) => {
    setRole(nextRole);
    setAgency(agencyByRole[nextRole]);
  };

  const submitCredentials = (event: FormEvent) => {
    event.preventDefault();
    if (!identifier.trim()) {
      setError("Enter your admin ID or agency email.");
      return;
    }
    if (password.length < 6) {
      setError("Enter your password to continue.");
      return;
    }
    setError("");
    setStep("mfa");
  };

  const submitMfa = (event: FormEvent) => {
    event.preventDefault();
    if (!/^\d{6}$/.test(code)) {
      setError("Enter the 6-digit code from your authenticator.");
      return;
    }
    setError("");
    setBusy(true);
    const trimmed = identifier.trim();
    const session: AdminSession = {
      id: trimmed || `usr_${role}`,
      name: displayName(trimmed, role),
      role,
      agencyId: DEFAULT_AGENCY_ID,
      agency: agency.trim() || agencyByRole[role],
      mfaCompleted: true,
    };
    window.setTimeout(() => signInAdmin(session), 500);
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
                Admin sign-in
              </Typography>
              <Typography className="cc-auth__lede" sx={{
                color: "text.secondary"
              }}>
                Use your admin credentials to reach the governance console.
              </Typography>

              <TextField
                label="Admin ID or agency email"
                value={identifier}
                onChange={(event) => setIdentifier(event.target.value)}
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

              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <TextField
                  select
                  label="Admin role"
                  value={role}
                  onChange={(event) =>
                    onRoleChange(event.target.value as AgencyUserRole)
                  }
                  fullWidth
                >
                  {adminSignInRoles.map((option) => (
                    <MenuItem value={option} key={option}>
                      {roleLabel(option)}
                    </MenuItem>
                  ))}
                </TextField>
                <TextField
                  label="Agency"
                  value={agency}
                  onChange={(event) => setAgency(event.target.value)}
                  fullWidth
                />
              </Stack>

              <Button
                type="submit"
                variant="contained"
                size="large"
                endIcon={<ArrowRight size={18} />}
                className="cc-auth__submit"
              >
                Continue
              </Button>

              <p className="cc-auth__hint">
                Demo console: any password (6+ characters) and any 6-digit code.
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
                Enter the 6-digit code for{" "}
                <strong>{displayName(identifier, role)}</strong> at{" "}
                {agency || agencyByRole[role]}.
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
