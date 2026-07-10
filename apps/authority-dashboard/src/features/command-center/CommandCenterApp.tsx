import { CssBaseline, ThemeProvider } from "@mui/material";
import { hasCommandAccess, useAuthoritySession } from "@/app/session";
import { authorityTheme } from "@/app/theme";
import { CommandCenterShell } from "./CommandCenterShell";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Provides the theme and gates the entire command center behind
 * an authenticated authority session. Until an operator signs in and completes
 * MFA, only the sign-in screen renders; the shell (sidebar + views) mounts and
 * loads its data once access is granted.
 */
function CommandCenterApp() {
  const session = useAuthoritySession();
  const authorized = hasCommandAccess(session);

  return (
    <ThemeProvider theme={authorityTheme}>
      <CssBaseline />
      {authorized && session ? (
        <CommandCenterShell session={session} />
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

export default CommandCenterApp;
