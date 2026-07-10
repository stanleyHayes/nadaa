import { useEffect, useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { hasCommandAccess, useDispatcherSession } from "@/app/session";
import { createDispatcherTheme } from "@/app/theme";
import { useThemeMode } from "@/app/theme-mode";
import { DispatchCommandShell } from "./DispatchCommandShell";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Provides the theme and gates the entire dispatch console behind
 * an authenticated dispatcher session. Until a controller signs in and completes
 * MFA, only the sign-in screen renders; the shell (sidebar + views) mounts and
 * loads its data once access is granted.
 */
function DispatcherCommandApp() {
  const { session, preferences } = useDispatcherSession();
  const authorized = hasCommandAccess(session);
  const mode = useThemeMode();

  // Rebuild the MUI theme whenever the appearance mode or reduced-motion
  // preference changes, so dialogs/menus/inputs match the CSS token flip.
  const theme = useMemo(
    () =>
      createDispatcherTheme({ mode, reducedMotion: preferences.reducedMotion }),
    [mode, preferences.reducedMotion],
  );

  // Apply the controller's reduced-motion preference across the whole app.
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
      {authorized && session ? (
        <DispatchCommandShell session={session} />
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

export default DispatcherCommandApp;
