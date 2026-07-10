import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { LogOut, ShieldAlert } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import { signOutAdmin, type AdminSession } from "@/app/session";
import { roleLabel } from "../utils";

/**
 * Shown when a session exists but the account is not permitted to reach the
 * governance console (non-admin role, or MFA not completed). Offers a way back
 * to sign-in so the screen is never a dead end.
 */
export function AccessDenied({ session }: { session: AdminSession }) {
  return (
    <Box component="main" id="main-content" className="access-shell">
      <Paper className="access-panel">
        <ShieldAlert size={38} color={nadaaBrand.colors.red} />
        <Typography variant="h5">Admin access denied</Typography>
        <Typography color="text.secondary">
          The governance console requires a permitted admin role (system or
          agency administrator) and completed MFA before any platform settings
          are visible.
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Signed in as <strong>{session.name}</strong> ({roleLabel(session.role)})
          {session.mfaCompleted ? "" : " · MFA incomplete"}.
        </Typography>
        <Stack direction="row" justifyContent="center">
          <Button
            variant="outlined"
            startIcon={<LogOut size={18} />}
            onClick={() => signOutAdmin()}
          >
            Sign out
          </Button>
        </Stack>
      </Paper>
    </Box>
  );
}
