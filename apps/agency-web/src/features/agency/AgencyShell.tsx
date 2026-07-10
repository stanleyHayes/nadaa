import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { signOutAgency, type AgencySession } from "@/app/session";
import { useAgencyData } from "./useAgencyData";
import {
  DEFAULT_VIEW,
  groupLabelForView,
  isViewId,
  navItemById,
  type BadgeKey,
  type ViewId,
} from "./navigation";
import { Sidebar } from "./components/Sidebar";
import { Topbar, type AgencyNotification } from "./components/Topbar";
import { OverviewView } from "./components/views/OverviewView";
import { AssignedIncidentsView } from "./components/views/AssignedIncidentsView";
import { CapacityView } from "./components/views/CapacityView";
import { ReliefView } from "./components/views/ReliefView";
import { AidView } from "./components/views/AidView";

const VIEW_KEY = "nadaa.agency.view";
const COLLAPSE_KEY = "nadaa.agency.sidebar.collapsed";

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

export function AgencyShell({ session }: { session: AgencySession }) {
  const data = useAgencyData(session);
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

  const badges: Record<BadgeKey, number> = useMemo(
    () => ({
      openIncidents: data.metrics.open,
      sheltersCritical: data.sheltersCritical,
      reliefOpen: data.metrics.reliefOpen,
      aidOpen: data.metrics.aidOpen,
    }),
    [data.metrics, data.sheltersCritical],
  );

  const notifications: AgencyNotification[] = useMemo(() => {
    const list: AgencyNotification[] = [];
    if (data.metrics.priority > 0) {
      list.push({
        id: "priority-incidents",
        title: `${data.metrics.priority} incident${data.metrics.priority === 1 ? "" : "s"} flagged for priority review`,
        detail: "Confirm response posture on the flagged reports.",
        tone: "red",
      });
    }
    if (data.metrics.assigned > 0) {
      list.push({
        id: "assigned-incidents",
        title: `${data.metrics.assigned} incident${data.metrics.assigned === 1 ? "" : "s"} awaiting response`,
        detail: "Move assigned incidents to en route once teams roll.",
        tone: "gold",
      });
    }
    if (data.metrics.aidPending > 0) {
      list.push({
        id: "aid-pending",
        title: `${data.metrics.aidPending} aid need${data.metrics.aidPending === 1 ? "" : "s"} awaiting review`,
        detail: "Verify receiving organisations before listing.",
        tone: "gold",
      });
    }
    if (data.sheltersCritical > 0) {
      list.push({
        id: "shelters-critical",
        title: `${data.sheltersCritical} shelter${data.sheltersCritical === 1 ? "" : "s"} near capacity`,
        detail: "Confirm occupancy and open overflow if needed.",
        tone: "gold",
      });
    }
    if (data.incidentLoadState === "fallback") {
      list.push({
        id: "feed-fixture",
        title: "Incident feed on fixtures",
        detail: "Live incident API is unavailable; showing agency fixtures.",
        tone: "gold",
      });
    }
    return list;
  }, [
    data.metrics.priority,
    data.metrics.assigned,
    data.metrics.aidPending,
    data.sheltersCritical,
    data.incidentLoadState,
  ]);

  const selectView = (view: ViewId) => {
    setActiveView(view);
    setMobileNavOpen(false);
  };

  const handleSignOut = () => {
    setMobileNavOpen(false);
    signOutAgency();
  };

  const renderView = () => {
    switch (activeView) {
      case "incidents":
        return <AssignedIncidentsView data={data} />;
      case "capacity":
        return <CapacityView data={data} onNavigate={selectView} />;
      case "relief":
        return <ReliefView data={data} />;
      case "aid":
        return <AidView data={data} />;
      case "overview":
      default:
        return <OverviewView data={data} onNavigate={selectView} />;
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
