import { CssBaseline, ThemeProvider } from "@mui/material";
import { hasCommandAccess, useDispatcherSession } from "@/app/session";
import { dispatcherTheme } from "@/app/theme";
import { DispatchCommandShell } from "./DispatchCommandShell";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Provides the theme and gates the entire dispatch console behind
 * an authenticated dispatcher session. Until a controller signs in and completes
 * MFA, only the sign-in screen renders; the shell (sidebar + views) mounts and
 * loads its data once access is granted.
 */
function DispatcherCommandApp() {
  const session = useDispatcherSession();
  const authorized = hasCommandAccess(session);

  return (
    <ThemeProvider theme={dispatcherTheme}>
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
