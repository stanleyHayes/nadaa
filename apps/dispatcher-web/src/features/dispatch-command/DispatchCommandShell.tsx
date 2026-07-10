import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { signOutDispatcher, type DispatcherSession } from "@/app/session";
import { useDispatchData } from "./useDispatchData";
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
import { TriageView } from "./components/views/TriageView";
import { MLReviewView } from "./components/views/MLReviewView";
import { AlertsView } from "./components/views/AlertsView";
import { CapacityView } from "./components/views/CapacityView";

const VIEW_KEY = "nadaa.dispatcher.view";
const COLLAPSE_KEY = "nadaa.dispatcher.sidebar.collapsed";

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

export function DispatchCommandShell({
  session,
}: {
  session: DispatcherSession;
}) {
  const data = useDispatchData();
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
    const mlNeedsReview = data.mlPredictions.filter(
      (prediction) => prediction.reviewStatus === "needs_review",
    ).length;
    return { openIncidents, pendingAlerts, mlNeedsReview };
  }, [data.incidents, data.alerts, data.mlPredictions]);

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
    if (badges.mlNeedsReview > 0) {
      list.push({
        id: "ml-review",
        title: `${badges.mlNeedsReview} ML flood signal${badges.mlNeedsReview === 1 ? "" : "s"} to review`,
        detail: "Human review is required before any ML-driven draft.",
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
    signOutDispatcher();
  };

  const renderView = () => {
    switch (activeView) {
      case "incidents":
        return <IncidentsView data={data} />;
      case "triage":
        return <TriageView data={data} />;
      case "ml-review":
        return <MLReviewView data={data} />;
      case "alerts":
        return <AlertsView data={data} />;
      case "capacity":
        return <CapacityView data={data} />;
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
