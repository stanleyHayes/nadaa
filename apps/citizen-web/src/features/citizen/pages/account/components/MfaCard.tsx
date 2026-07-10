import { useId, useState, type ReactNode } from "react";
import {
  Alert,
  Box,
  Button,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { ShieldCheck, Smartphone } from "lucide-react";
import { useCitizenSession } from "../../../session";

/** Rounded status pill: green when MFA is on, gold ("attention") when it is off. */
function MfaStatusChip({ enabled }: { enabled: boolean }) {
  const token = enabled ? "green" : "gold";
  return (
    <Box
      component="span"
      sx={{
        display: "inline-flex",
        alignItems: "center",
        gap: 0.75,
        px: 1.25,
        py: 0.4,
        borderRadius: "999px",
        fontSize: "0.72rem",
        fontWeight: 800,
        letterSpacing: "0.04em",
        textTransform: "uppercase",
        whiteSpace: "nowrap",
        // Gold tint reads best with navy text; green tone keeps its own colour.
        color: enabled
          ? "var(--nadaa-green, #1f8a4c)"
          : "var(--nadaa-navy, #0d1b3d)",
        backgroundColor: `color-mix(in srgb, var(--nadaa-${token}) 16%, transparent)`,
        border: `1px solid color-mix(in srgb, var(--nadaa-${token}) 42%, transparent)`,
      }}
    >
      <Box
        aria-hidden
        component="span"
        sx={{
          width: 8,
          height: 8,
          borderRadius: "999px",
          backgroundColor: `var(--nadaa-${token})`,
        }}
      />
      {enabled ? "Enabled" : "Not enabled"}
    </Box>
  );
}

/** Muted supporting line, matching the account-settings secondary text. */
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

/** Inset panel used for the status row and the enrolment step. */
function FieldPanel({ children }: { children: ReactNode }) {
  return (
    <Box
      sx={{
        p: 2,
        borderRadius: "12px",
        border: "1px solid var(--nadaa-border, #dfeaf2)",
        backgroundColor: "var(--nadaa-mist, #f5f8fc)",
      }}
    >
      {children}
    </Box>
  );
}

/**
 * Multi-factor authentication card for the citizen Security settings. Shown
 * ABOVE the change-password form. `setMfaEnabled` in the session store is a
 * mock (no authenticator backend yet), so this focuses on a clear enable /
 * disable flow with a six-digit confirmation step and status feedback.
 */
export function MfaCard() {
  const { mfaEnabled, setMfaEnabled } = useCitizenSession();
  const [enrolling, setEnrolling] = useState(false);
  const [code, setCode] = useState("");
  const [notice, setNotice] = useState<string | null>(null);
  const codeFieldId = useId();

  const startEnrolment = () => {
    setEnrolling(true);
    setCode("");
    setNotice(null);
  };

  const cancelEnrolment = () => {
    setEnrolling(false);
    setCode("");
  };

  const verifyAndEnable = () => {
    if (code.length !== 6) {
      return;
    }
    setMfaEnabled(true);
    setEnrolling(false);
    setCode("");
    setNotice("Multi-factor authentication is on. We will ask for a code next time you sign in.");
  };

  const disableMfa = () => {
    setMfaEnabled(false);
    setEnrolling(false);
    setCode("");
    setNotice("Multi-factor authentication is off. Your account now signs in with a password only.");
  };

  return (
    <Paper className="surface" component="section" elevation={0}>
      <Stack direction="row" spacing={1.5} sx={{
        alignItems: "flex-start"
      }}>
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 40,
            height: 40,
            borderRadius: "10px",
            color: "var(--nadaa-navy, #0d1b3d)",
            backgroundColor:
              "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 8%, transparent)",
          }}
        >
          <ShieldCheck size={20} />
        </Box>
        <Box sx={{ minWidth: 0 }}>
          <Typography
            component="h3"
            sx={{
              fontSize: "1.02rem",
              fontWeight: 800,
              lineHeight: 1.2,
              color: "var(--nadaa-ink, #101828)",
            }}
          >
            Multi-factor authentication
          </Typography>
          <Typography
            sx={{
              mt: 0.25,
              fontSize: "0.85rem",
              color: "var(--nadaa-text-secondary, #555b66)",
            }}
          >
            Protect your account with a six-digit authenticator code.
          </Typography>
        </Box>
      </Stack>
      <Stack spacing={2} sx={{ mt: 2.5 }}>
        {notice ? (
          <Alert
            severity="success"
            className="warning-alert"
            onClose={() => setNotice(null)}
          >
            {notice}
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
            <Stack spacing={0.75} sx={{ minWidth: 0 }}>
              <MfaStatusChip enabled={mfaEnabled} />
              <MutedNote>
                {mfaEnabled
                  ? "A code from your authenticator app is required each time you sign in."
                  : "Add a six-digit code from an authenticator app for a second layer of protection."}
              </MutedNote>
            </Stack>
            {mfaEnabled ? (
              <Button
                type="button"
                onClick={disableMfa}
                variant="outlined"
                color="error"
                sx={{ flexShrink: 0 }}
              >
                Disable
              </Button>
            ) : !enrolling ? (
              <Button
                type="button"
                onClick={startEnrolment}
                variant="contained"
                color="warning"
                className="signin-submit"
                startIcon={<Smartphone size={18} />}
                sx={{ flexShrink: 0 }}
              >
                Enable MFA
              </Button>
            ) : null}
          </Stack>
        </FieldPanel>

        {!mfaEnabled && enrolling ? (
          <FieldPanel>
            <Stack spacing={1.75}>
              <Box>
                <Typography
                  component="h4"
                  sx={{
                    fontSize: "0.95rem",
                    fontWeight: 700,
                    color: "var(--nadaa-ink, #101828)",
                  }}
                >
                  Confirm your authenticator code
                </Typography>
                <MutedNote>
                  Add NADAA to an authenticator app (such as Google
                  Authenticator or Authy), then type the current six-digit code
                  below to finish turning MFA on.
                </MutedNote>
              </Box>
              <TextField
                id={codeFieldId}
                label="Six-digit code"
                value={code}
                onChange={(event) =>
                  setCode(event.target.value.replace(/\D/g, "").slice(0, 6))
                }
                autoFocus
                autoComplete="one-time-code"
                placeholder="000000"
                helperText="Enter the 6 digits shown in your authenticator app."
                sx={{ maxWidth: 260 }}
                slotProps={{
                  htmlInput: {
                    inputMode: "numeric",
                    pattern: "[0-9]*",
                    maxLength: 6,
                    "aria-label": "Six-digit authenticator code",
                    style: { letterSpacing: "0.35em", fontWeight: 700 },
                  }
                }}
              />
              <Stack direction="row" spacing={1.5} sx={{
                flexWrap: "wrap"
              }}>
                <Button
                  type="button"
                  onClick={verifyAndEnable}
                  variant="contained"
                  color="warning"
                  className="signin-submit"
                  disabled={code.length !== 6}
                  startIcon={<ShieldCheck size={18} />}
                >
                  Verify &amp; enable
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
      </Stack>
    </Paper>
  );
}

export default MfaCard;
