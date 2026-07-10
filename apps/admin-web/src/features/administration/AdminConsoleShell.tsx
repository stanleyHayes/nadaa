import { useEffect, useMemo, useState } from "react";
import { Drawer } from "@mui/material";
import { BookOpen, Settings } from "lucide-react";
import {
  signOutAdmin,
  useAdminSession,
  type AdminSession,
} from "@/app/session";
import { useAdminData } from "./useAdminData";
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
import { Topbar, type AdminNotification } from "./components/Topbar";
import { AppTour } from "./components/AppTour";
import { AccountSettings, UserGuide, type SettingsTab } from "./account";
import { OverviewView } from "./components/views/OverviewView";
import { AgenciesView } from "./components/views/AgenciesView";
import { UsersView } from "./components/views/UsersView";
import { RolesView } from "./components/views/RolesView";
import { MfaView } from "./components/views/MfaView";
import { AuditView } from "./components/views/AuditView";
import { IntegrationsView } from "./components/views/IntegrationsView";
import { AlertRulesView } from "./components/views/AlertRulesView";

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
  description: "Step-by-step help for every admin-console workspace",
  icon: BookOpen,
};

type ShellView = ViewId | "settings" | "guide";

const VIEW_KEY = "nadaa.admin.view";
const COLLAPSE_KEY = "nadaa.admin.sidebar.collapsed";

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

export function AdminConsoleShell({ session }: { session: AdminSession }) {
  const data = useAdminData();
  const {
    preferences,
    updateProfile,
    updatePreferences,
    setMfaEnabled,
    changePassword,
  } = useAdminSession();
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
    const usersAwaitingMfa = data.users.filter(
      (user) => !user.mfaEnabled,
    ).length;
    return { agencies: data.agencies.length, usersAwaitingMfa };
  }, [data.agencies, data.users]);

  const notifications: AdminNotification[] = useMemo(() => {
    const list: AdminNotification[] = [];
    if (badges.usersAwaitingMfa > 0) {
      list.push({
        id: "awaiting-mfa",
        title: `${badges.usersAwaitingMfa} user${badges.usersAwaitingMfa === 1 ? "" : "s"} awaiting MFA`,
        detail: "Setup must complete before these users can sign in.",
        tone: "gold",
      });
    }
    const pendingRules = data.alertRules.filter(
      (rule) => rule.status === "draft" || rule.status === "submitted",
    ).length;
    if (pendingRules > 0) {
      list.push({
        id: "pending-rules",
        title: `${pendingRules} alert rule${pendingRules === 1 ? "" : "s"} in review`,
        detail: "Governance rules are still draft or submitted.",
        tone: "gold",
      });
    }
    if (data.loadState === "error") {
      list.push({
        id: "governance-error",
        title: "Governance APIs unavailable",
        detail: "One or more admin APIs could not be reached. Retry from the console.",
        tone: "red",
      });
    }
    return list;
  }, [badges.usersAwaitingMfa, data.alertRules, data.loadState]);

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
    signOutAdmin();
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
      case "agencies":
        return <AgenciesView data={data} />;
      case "users":
        return <UsersView data={data} />;
      case "roles":
        return <RolesView />;
      case "mfa":
        return <MfaView data={data} />;
      case "audit":
        return <AuditView data={data} />;
      case "integrations":
        return <IntegrationsView data={data} />;
      case "alertRules":
        return <AlertRulesView data={data} />;
      case "settings":
        return (
          <AccountSettings
            tab={settingsTab}
            onTabChange={setSettingsTab}
            user={session}
            preferences={preferences}
            onUpdateProfile={updateProfile}
            onUpdatePreferences={updatePreferences}
            onSetMfaEnabled={setMfaEnabled}
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
