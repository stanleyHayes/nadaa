import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { BookOpen, Settings } from "lucide-react";
import {
  signOutAuthority,
  useAuthoritySession,
  type AuthoritySession,
} from "@/app/session";
import { useCommandData } from "./useCommandData";
import {
  DEFAULT_VIEW,
  groupLabelForView,
  isViewId,
  navItemById,
  type BadgeKey,
  type NavItem,
  type ViewId,
} from "./navigation";
import { getPageGuide, type GuideKey } from "./pageGuides";
import { Sidebar } from "./components/Sidebar";
import { Topbar, type CommandNotification } from "./components/Topbar";
import { AppTour } from "./components/AppTour";
import { AccountSettings, UserGuide, type SettingsTab } from "./account";
import { OverviewView } from "./components/views/OverviewView";
import { IncidentsView } from "./components/views/IncidentsView";
import { AlertsView } from "./components/views/AlertsView";
import { SheltersView } from "./components/views/SheltersView";
import { ForecastingView } from "./components/views/ForecastingView";
import { EvidenceView } from "./components/views/EvidenceView";
import { RecoveryView } from "./components/views/RecoveryView";
import { PreparednessView } from "./components/views/PreparednessView";

/** Synthetic nav item so the topbar can title the settings view. */
const SETTINGS_NAV_ITEM: NavItem = {
  id: DEFAULT_VIEW,
  label: "Settings",
  description: "Manage your profile, security, notifications and preferences",
  icon: Settings,
};

/** Synthetic nav item so the topbar can title the user-guide view. */
const GUIDE_NAV_ITEM: NavItem = {
  id: DEFAULT_VIEW,
  label: "User guide",
  description: "Step-by-step help for every command-center workspace",
  icon: BookOpen,
};

type ShellView = ViewId | "settings" | "guide";

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
  const {
    preferences,
    updateProfile,
    updatePreferences,
    changePassword,
  } = useAuthoritySession();
  const [activeView, setActiveView] = useState<ShellView>(readInitialView);
  const [settingsTab, setSettingsTab] = useState<SettingsTab>("profile");
  const [collapsed, setCollapsed] = useState<boolean>(readInitialCollapsed);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  useEffect(() => {
    if (activeView === "settings" || activeView === "guide") {
      return;
    }
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
    if (data.loadState === "error") {
      list.push({
        id: "feed-offline",
        title: "Incident feed unavailable",
        detail: "Live incident API is unreachable; reconnect the incident-service.",
        tone: "gold",
      });
    }
    return list;
  }, [data.incidents, data.loadState, badges]);

  const selectView = (view: ViewId) => {
    setActiveView(view);
    setMobileNavOpen(false);
  };

  const openSettings = (tab: SettingsTab) => {
    setSettingsTab(tab);
    setActiveView("settings");
    setMobileNavOpen(false);
  };

  const openGuide = () => {
    setActiveView("guide");
    setMobileNavOpen(false);
  };

  /** Jump from a user-guide card to the workspace it documents. */
  const openGuideTarget = (key: GuideKey) => {
    if (key === "guide") {
      return;
    }
    if (key === "settings") {
      openSettings("profile");
      return;
    }
    selectView(key);
  };

  const handleSignOut = () => {
    setMobileNavOpen(false);
    signOutAuthority();
  };

  const isSettings = activeView === "settings";
  const isGuide = activeView === "guide";
  const topbarView = isSettings
    ? SETTINGS_NAV_ITEM
    : isGuide
      ? GUIDE_NAV_ITEM
      : navItemById(activeView);
  const topbarGroup = isSettings
    ? "Account"
    : isGuide
      ? "Help"
      : groupLabelForView(activeView);
  const activeGuide = getPageGuide(activeView);

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
      case "settings":
        return (
          <AccountSettings
            tab={settingsTab}
            onTabChange={setSettingsTab}
            user={session}
            preferences={preferences}
            onUpdateProfile={updateProfile}
            onUpdatePreferences={updatePreferences}
            onChangePassword={changePassword}
          />
        );
      case "guide":
        return <UserGuide onOpen={openGuideTarget} />;
      case "overview":
      default:
        return (
          <OverviewView data={data} session={session} onNavigate={selectView} />
        );
    }
  };

  return (
    <div className={`cc-shell${collapsed ? " is-collapsed" : ""}`}>

      <aside className="cc-shell__rail" data-tour="sidebar">
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
          view={topbarView}
          groupLabel={topbarGroup}
          guide={activeGuide}
          session={session}
          notifications={notifications}
          onSignOut={handleSignOut}
          onOpenSettings={openSettings}
          onOpenGuide={openGuide}
          onOpenMobileNav={() => setMobileNavOpen(true)}
        />
        <main id="main-content" className="cc-content" tabIndex={-1}>
          {renderView()}
        </main>
      </div>

      <AppTour />
    </div>
  );
}
