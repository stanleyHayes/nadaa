import { CssBaseline, ThemeProvider } from "@mui/material";
import { useAdminSession } from "@/app/session";
import { adminTheme } from "@/app/theme";
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
  const session = useAdminSession();
  const authorized = Boolean(
    session && hasAdminAccess(session.role, session.mfaCompleted),
  );

  return (
    <ThemeProvider theme={adminTheme}>
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
