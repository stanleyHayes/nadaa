import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { SiteFooter } from "./SiteFooter";
import { SiteHeader } from "./SiteHeader";

/** Shared marketing chrome: skip link, header, routed page, footer. */
export function SiteLayout() {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "auto" });
  }, [pathname]);

  return (
    <div className="site-shell">
      <a className="skip-link" href="#main-content">
        Skip to main content
      </a>
      <SiteHeader />
      <main id="main-content">
        {/* Keyed on the path so each navigation remounts and replays the
            fade-and-lift enter animation — pages ease in instead of hard-cutting. */}
        <div className="route-view" key={pathname}>
          <Outlet />
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
