import { useMemo } from "react";
import { Outlet } from "react-router-dom";
import { PageBanner } from "../../components/PageBanner";
import { useCitizenSession } from "../../session";
import { AccountSidenav, SignInGate, initialsOf } from "./components";

/**
 * Account area shell (route `/account`). Wraps every account view in a shared
 * navy banner and a two-column layout: a sticky sub-navigation rail (which
 * collapses to a scrollable row on mobile) beside the routed view via
 * <Outlet/>. The whole area is gated — signed-out visitors see the sign-in
 * prompt in place instead of being redirected away.
 */
export function AccountLayout() {
  const { session, notifications, requestSignIn } = useCitizenSession();

  const unread = useMemo(
    () => notifications.filter((item) => !item.read).length,
    [notifications],
  );

  const firstName = session ? session.name.split(" ")[0] : "";

  return (
    <>
      <PageBanner
        eyebrow="Your account"
        title={session ? `Welcome back, ${firstName}` : "Your NADAA account"}
        subtitle={
          session
            ? "Your dashboard, reports, notifications and settings — all in one place."
            : "Sign in to reach your dashboard, report history, notifications and settings."
        }
      >
        {session ? (
          <span className="account-banner-badge">
            <span className="account-banner-badge__avatar" aria-hidden="true">
              {initialsOf(session.name)}
            </span>
            {session.name} · {session.region}
          </span>
        ) : null}
      </PageBanner>

      <div className="citizen-shell">
        {session ? (
          <div className="account-layout">
            <AccountSidenav unread={unread} />
            <div className="account-content">
              <Outlet />
            </div>
          </div>
        ) : (
          <div className="citizen-section">
            <SignInGate onSignIn={requestSignIn} />
          </div>
        )}
      </div>
    </>
  );
}

export default AccountLayout;
