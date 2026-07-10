import { Box, Button, Paper, Typography } from "@mui/material";
import { LogIn, ShieldCheck } from "lucide-react";
import type { ReactNode } from "react";
import { nadaaBrand } from "@nadaa/brand";
import { useCitizenSession } from "../session";

type SignInGateProps = {
  title?: string;
  message?: string;
  children: ReactNode;
};

/**
 * Wraps a SUBMISSION surface. Viewing the platform is public, but submitting
 * requires a signed-in citizen — no anonymous submissions — so responders can
 * verify and follow up. When signed out, shows a sign-in prompt instead of the
 * form and opens the shared sign-in dialog.
 */
export function SignInGate({
  title = "Sign in to submit",
  message = "Browsing is open to everyone. To submit, sign in with your name and phone — anonymous submissions aren't accepted, so responders can verify your report and follow up.",
  children,
}: SignInGateProps) {
  const { session, requestSignIn } = useCitizenSession();

  if (session) {
    return <>{children}</>;
  }

  return (
    <Paper
      elevation={0}
      sx={{
        display: "flex",
        flexDirection: "column",
        alignItems: "flex-start",
        gap: 1.5,
        p: { xs: 2.5, sm: 3.5 },
        borderRadius: 2,
        border: `1px solid ${nadaaBrand.colors.navy}22`,
        borderLeft: `4px solid ${nadaaBrand.colors.gold}`,
        background: `${nadaaBrand.colors.gold}0F`,
      }}
    >
      <Box
        aria-hidden
        sx={{
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          width: 46,
          height: 46,
          borderRadius: 2,
          color: nadaaBrand.colors.navy,
          background: `${nadaaBrand.colors.gold}2E`,
        }}
      >
        <ShieldCheck size={24} />
      </Box>
      <Typography
        sx={{ fontWeight: 800, color: nadaaBrand.colors.navy }}
        variant="h6"
      >
        {title}
      </Typography>
      <Typography
        sx={{ color: "text.secondary", maxWidth: 560, lineHeight: 1.6 }}
      >
        {message}
      </Typography>
      <Button
        onClick={requestSignIn}
        size="large"
        startIcon={<LogIn size={18} />}
        sx={{ mt: 0.5, fontWeight: 800 }}
        variant="contained"
      >
        Sign in to continue
      </Button>
    </Paper>
  );
}
