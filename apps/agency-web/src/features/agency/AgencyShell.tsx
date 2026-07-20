import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { BookOpen, Settings } from "lucide-react";
import {
  signOutAgency,
  useAgencySession,
  type AgencySession,
} from "@/app/session";
import { useAgencyData } from "./useAgencyData";
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
import { Topbar, type AgencyNotification } from "./components/Topbar";
import { AppTour } from "./components/AppTour";
import { AccountSettings, UserGuide, type SettingsTab } from "./account";
import { OverviewView } from "./components/views/OverviewView";
import { AssignedIncidentsView } from "./components/views/AssignedIncidentsView";
import { CapacityView } from "./components/views/CapacityView";
import { ReliefView } from "./components/views/ReliefView";
import { AidView } from "./components/views/AidView";

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
  description: "Step-by-step help for every agency workspace",
  icon: BookOpen,
};

type ShellView = ViewId | "settings" | "guide";

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
  const {
    preferences,
    updateProfile,
    updatePreferences,
    changePassword,
  } = useAgencySession();
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
    if (data.incidentLoadState === "error") {
      list.push({
        id: "feed-error",
        title: "Incident feed unavailable",
        detail: "Live incident API could not be reached. Retry from the incidents queue.",
        tone: "red",
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
    signOutAgency();
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
        return <AssignedIncidentsView data={data} />;
      case "capacity":
        return <CapacityView data={data} onNavigate={selectView} />;
      case "relief":
        return <ReliefView data={data} />;
      case "aid":
        return <AidView data={data} />;
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
        return <OverviewView data={data} onNavigate={selectView} />;
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
