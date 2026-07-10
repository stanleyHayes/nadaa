import { hasAgencyAccess, useAgencySession } from "@/app/session";
import { AgencyShell } from "./AgencyShell";
import { SignInScreen } from "./components/SignInScreen";

/**
 * Feature root. Gates every agency-operations surface behind an authenticated
 * agency session. Until an operator signs in and completes MFA, only the
 * sign-in screen renders; the shell (sidebar + views) mounts and loads its data
 * once access is granted. The MUI ThemeProvider is applied in main.tsx.
 */
export function AgencyApp() {
  const session = useAgencySession();
  const authorized = hasAgencyAccess(session);

  if (authorized && session) {
    return <AgencyShell session={session} />;
  }

  return (
    <>
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <SignInScreen />
    </>
  );
}
