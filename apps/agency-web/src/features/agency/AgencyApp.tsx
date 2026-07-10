import { useEffect } from "react";
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
  const { session, preferences } = useAgencySession();
  const authorized = hasAgencyAccess(session);

  // Apply the operator's reduced-motion preference across the whole app.
  useEffect(() => {
    const root = document.documentElement;
    if (preferences.reducedMotion) {
      root.setAttribute("data-nadaa-reduced-motion", "reduce");
    } else {
      root.removeAttribute("data-nadaa-reduced-motion");
    }
  }, [preferences.reducedMotion]);

  if (authorized && session) {
    return <AgencyShell session={session} />;
  }

  return (
    <>
      <SignInScreen />
    </>
  );
}
