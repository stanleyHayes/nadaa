import { useEffect, useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { useAdminSession } from "@/app/session";
import { createAdminTheme } from "@/app/theme";
import { useThemeMode } from "@/app/theme-mode";
import { AdminConsoleShell } from "./AdminConsoleShell";
import { hasAdminAccess } from "./rbac";
import { AccessDenied } from "./components/AccessDenied";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Provides the theme and gates the entire governance console
 * behind an authenticated admin session. Until an admin signs in and completes
 * MFA, only the sign-in screen renders. A signed-in account that is not a
 * permitted admin role (or has not completed MFA) sees an access-denied screen
 * with a way back to sign-in; the shell (sidebar + views) only mounts once the
 * RBAC + MFA gate passes.
 */
function AdminConsoleApp() {
  const { session, preferences } = useAdminSession();
  const authorized = Boolean(
    session && hasAdminAccess(session.role, session.mfaCompleted),
  );
  const mode = useThemeMode();

  // Rebuild the MUI theme whenever the appearance mode or reduced-motion
  // preference changes, so dialogs/menus/inputs match the CSS token flip.
  const theme = useMemo(
    () => createAdminTheme({ mode, reducedMotion: preferences.reducedMotion }),
    [mode, preferences.reducedMotion],
  );

  // Apply the admin's reduced-motion preference across the whole app.
  useEffect(() => {
    const root = document.documentElement;
    if (preferences.reducedMotion) {
      root.setAttribute("data-nadaa-reduced-motion", "reduce");
    } else {
      root.removeAttribute("data-nadaa-reduced-motion");
    }
  }, [preferences.reducedMotion]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {session ? (
        authorized ? (
          <AdminConsoleShell session={session} />
        ) : (
          <>
            <a href="#main-content" className="skip-link">
              Skip to main content
            </a>
            <AccessDenied session={session} />
          </>
        )
      ) : (
        <>
          <a href="#main-content" className="skip-link">
            Skip to main content
          </a>
          <SignInScreen />
        </>
      )}
    </ThemeProvider>
  );
}

export default AdminConsoleApp;
