import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { signOutAuthority, type AuthoritySession } from "@/app/session";
import { useCommandData } from "./useCommandData";
import {
  DEFAULT_VIEW,
  groupLabelForView,
  isViewId,
  navItemById,
  type BadgeKey,
  type ViewId,
} from "./navigation";
import { Sidebar } from "./components/Sidebar";
import { Topbar, type CommandNotification } from "./components/Topbar";
import { OverviewView } from "./components/views/OverviewView";
import { IncidentsView } from "./components/views/IncidentsView";
import { AlertsView } from "./components/views/AlertsView";
import { SheltersView } from "./components/views/SheltersView";
import { ForecastingView } from "./components/views/ForecastingView";
import { EvidenceView } from "./components/views/EvidenceView";
import { RecoveryView } from "./components/views/RecoveryView";
import { PreparednessView } from "./components/views/PreparednessView";

const VIEW_KEY = "nadaa.authority.view";
const COLLAPSE_KEY = "nadaa.authority.sidebar.collapsed";

function readInitialView(): ViewId {
  if (typeof window === "undefined") {
    return DEFAULT_VIEW;
  }
  const stored = window.localStorage.getItem(VIEW_KEY);
  return isViewId(stored) ? stored : DEFAULT_VIEW;
}

function readInitialCollapsed(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  return window.localStorage.getItem(COLLAPSE_KEY) === "true";
}

export function CommandCenterShell({ session }: { session: AuthoritySession }) {
  const data = useCommandData();
  const [activeView, setActiveView] = useState<ViewId>(readInitialView);
  const [collapsed, setCollapsed] = useState<boolean>(readInitialCollapsed);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  useEffect(() => {
    try {
      window.localStorage.setItem(VIEW_KEY, activeView);
    } catch {
      /* storage unavailable */
    }
  }, [activeView]);

  useEffect(() => {
    try {
      window.localStorage.setItem(COLLAPSE_KEY, collapsed ? "true" : "false");
    } catch {
      /* storage unavailable */
    }
  }, [collapsed]);

  const badges: Record<BadgeKey, number> = useMemo(() => {
    const openIncidents = data.incidents.filter(
      (incident) =>
        incident.status !== "closed" && incident.status !== "false_report",
    ).length;
    const pendingAlerts = data.alerts.filter(
      (alert) => alert.status === "draft" || alert.status === "submitted",
    ).length;
    const sheltersCritical = data.shelters.filter(
      (shelter) =>
        shelter.status === "full" ||
        (shelter.capacity > 0 &&
          shelter.currentOccupancy / shelter.capacity >= 0.9),
    ).length;
    return { openIncidents, pendingAlerts, sheltersCritical };
  }, [data.incidents, data.alerts, data.shelters]);

  const notifications: CommandNotification[] = useMemo(() => {
    const list: CommandNotification[] = [];
    const critical = data.incidents.filter(
      (incident) =>
        (incident.severity === "emergency" || incident.severity === "severe") &&
        incident.status !== "closed" &&
        incident.status !== "false_report",
    ).length;
    if (critical > 0) {
      list.push({
        id: "critical-incidents",
        title: `${critical} high-severity incident${critical === 1 ? "" : "s"} active`,
        detail: "Emergency and severe reports need eyes on the queue.",
        tone: "red",
      });
    }
    if (badges.pendingAlerts > 0) {
      list.push({
        id: "pending-alerts",
        title: `${badges.pendingAlerts} alert${badges.pendingAlerts === 1 ? "" : "s"} awaiting approval`,
        detail: "Review the broadcast queue before publishing.",
        tone: "gold",
      });
    }
    if (badges.sheltersCritical > 0) {
      list.push({
        id: "shelters-critical",
        title: `${badges.sheltersCritical} shelter${badges.sheltersCritical === 1 ? "" : "s"} near capacity`,
        detail: "Confirm occupancy and open overflow if needed.",
        tone: "gold",
      });
    }
    if (data.loadState === "fallback") {
      list.push({
        id: "feed-fixture",
        title: "Incident feed on fixtures",
        detail: "Live incident API is unavailable; showing command fixtures.",
        tone: "gold",
      });
    }
    return list;
  }, [data.incidents, data.loadState, badges]);

  const selectView = (view: ViewId) => {
    setActiveView(view);
    setMobileNavOpen(false);
  };

  const handleSignOut = () => {
    setMobileNavOpen(false);
    signOutAuthority();
  };

  const renderView = () => {
    switch (activeView) {
      case "incidents":
        return <IncidentsView data={data} />;
      case "alerts":
        return <AlertsView data={data} />;
      case "shelters":
        return <SheltersView data={data} />;
      case "forecasting":
        return <ForecastingView />;
      case "evidence":
        return <EvidenceView data={data} />;
      case "recovery":
        return <RecoveryView />;
      case "preparedness":
        return <PreparednessView />;
      case "overview":
      default:
        return (
          <OverviewView data={data} session={session} onNavigate={selectView} />
        );
    }
  };

  return (
    <div className={`cc-shell${collapsed ? " is-collapsed" : ""}`}>
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>

      <aside className="cc-shell__rail">
        <Sidebar
          activeView={activeView}
          onSelect={setActiveView}
          badges={badges}
          variant="rail"
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed((value) => !value)}
        />
      </aside>

      <Drawer
        open={mobileNavOpen}
        onClose={() => setMobileNavOpen(false)}
        className="cc-shell__drawer"
        slotProps={{ paper: { className: "cc-shell__drawer-paper" } }}
      >
        <Sidebar
          activeView={activeView}
          onSelect={selectView}
          badges={badges}
          variant="drawer"
        />
      </Drawer>

      <div className="cc-main">
        <Topbar
          view={navItemById(activeView)}
          groupLabel={groupLabelForView(activeView)}
          session={session}
          notifications={notifications}
          onSignOut={handleSignOut}
          onOpenMobileNav={() => setMobileNavOpen(true)}
        />
        <main id="main-content" className="cc-content" tabIndex={-1}>
          {renderView()}
        </main>
      </div>
    </div>
  );
}
