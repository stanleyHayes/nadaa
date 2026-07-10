import { useEffect, useState } from "react";
import {
  Divider,
  Drawer,
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
} from "@mui/material";
import {
  Languages,
  LogOut,
  MapPinned,
  Menu as MenuIcon,
  PhoneCall,
  ShieldCheck,
  UserPlus,
  UserRound,
  X,
} from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { useCitizenSession } from "../session";
import { EmergencyBand } from "./EmergencyBand";
import { SignInDialog } from "./SignInDialog";

const EMERGENCY_TEL = "tel:112";

/** The six citizen routes, in header/drawer order. */
const navItems = [
  { to: "/", label: "Risk", end: true },
  { to: "/alerts", label: "Alerts", end: false },
  { to: "/report", label: "Report", end: false },
  { to: "/shelters", label: "Shelters", end: false },
  { to: "/guides", label: "Guides", end: false },
  { to: "/community", label: "Community", end: false },
] as const;

const navClass = ({ isActive }: { isActive: boolean }) =>
  isActive ? "is-active" : undefined;

/**
 * Shared citizen chrome: a sticky glass header with pill NavLinks (desktop) or a
 * hamburger-driven MUI Drawer (mobile, below ~760px), the optional light
 * sign-in, the routed page via <Outlet/>, a persistent 112 emergency band, and
 * scroll-to-top on every route change.
 */
export function CitizenLayout() {
  const { pathname } = useLocation();
  const { session, signIn, signOut } = useCitizenSession();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [signInOpen, setSignInOpen] = useState(false);
  const [userAnchor, setUserAnchor] = useState<null | HTMLElement>(null);

  // Reset scroll and close the mobile drawer whenever the route changes.
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "auto" });
    setDrawerOpen(false);
  }, [pathname]);

  const closeDrawer = () => setDrawerOpen(false);

  return (
    <div className="citizen-app-shell">
      <a className="skip-link" href="#main-content">
        Skip to main content
      </a>

      <div className="citizen-utility">
        <span className="citizen-utility__org">
          <span className="citizen-status-dot" aria-hidden="true" />
          {nadaaBrand.fullName}
        </span>
        <span className="citizen-utility__actions">
          <span className="citizen-utility__lang">
            <Languages aria-hidden="true" size={14} />
            EN
          </span>
          <a className="citizen-utility__emergency" href={EMERGENCY_TEL}>
            <PhoneCall aria-hidden="true" size={14} />
            Emergency? Call 112
          </a>
        </span>
      </div>

      <header className="citizen-header">
        <NavLink className="citizen-brand" to="/" aria-label="NADAA home">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>
              <ShieldCheck aria-hidden="true" size={12} />
              Be Aware. Be Prepared. Be Safe.
            </small>
          </span>
        </NavLink>

        <nav aria-label="Citizen sections" className="citizen-router-nav">
          {navItems.map((item) => (
            <NavLink
              className={navClass}
              end={item.end}
              key={item.to}
              to={item.to}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="citizen-header__actions">
          <Link className="citizen-cta" to="/">
            <MapPinned aria-hidden="true" size={16} />
            <span>Check my risk</span>
          </Link>
          <a className="citizen-emergency-link" href={EMERGENCY_TEL}>
            <PhoneCall aria-hidden="true" size={17} />
            112
          </a>

          {session ? (
            <>
              <button
                aria-expanded={Boolean(userAnchor)}
                aria-haspopup="true"
                className="citizen-user"
                onClick={(event) => setUserAnchor(event.currentTarget)}
                type="button"
              >
                <UserRound aria-hidden="true" size={16} />
                <span>{session.name.split(" ")[0]}</span>
              </button>
              <Menu
                anchorEl={userAnchor}
                anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
                onClose={() => setUserAnchor(null)}
                open={Boolean(userAnchor)}
                transformOrigin={{ vertical: "top", horizontal: "right" }}
              >
                <MenuItem disabled>
                  {session.name} · {session.region}
                </MenuItem>
                <Divider />
                <MenuItem
                  onClick={() => {
                    setUserAnchor(null);
                    signOut();
                  }}
                >
                  <ListItemIcon>
                    <LogOut size={18} />
                  </ListItemIcon>
                  Sign out
                </MenuItem>
              </Menu>
            </>
          ) : (
            <button
              className="citizen-signin"
              onClick={() => setSignInOpen(true)}
              type="button"
            >
              <UserPlus aria-hidden="true" size={16} />
              <span>Sign in</span>
            </button>
          )}

          <button
            aria-controls="citizen-drawer"
            aria-expanded={drawerOpen}
            aria-label="Open menu"
            className="citizen-router-burger"
            onClick={() => setDrawerOpen(true)}
            type="button"
          >
            <MenuIcon aria-hidden="true" size={22} />
          </button>
        </div>
      </header>

      <SignInDialog
        onClose={() => setSignInOpen(false)}
        onSignIn={(details) => {
          signIn(details);
          setSignInOpen(false);
        }}
        open={signInOpen}
      />

      <Drawer
        anchor="right"
        id="citizen-drawer"
        onClose={closeDrawer}
        open={drawerOpen}
        PaperProps={{ sx: { width: "min(320px, 82vw)" } }}
      >
        <div className="citizen-drawer">
          <div className="citizen-drawer__head">
            <strong>Menu</strong>
            <IconButton aria-label="Close menu" onClick={closeDrawer}>
              <X size={20} />
            </IconButton>
          </div>

          <nav aria-label="Citizen sections" className="citizen-drawer__nav">
            {navItems.map((item) => (
              <NavLink
                className={navClass}
                end={item.end}
                key={item.to}
                onClick={closeDrawer}
                to={item.to}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>

          <Divider />

          <Link className="citizen-cta" onClick={closeDrawer} to="/">
            <MapPinned aria-hidden="true" size={16} />
            Check my risk
          </Link>
          <a
            className="citizen-emergency-link"
            href={EMERGENCY_TEL}
            onClick={closeDrawer}
          >
            <PhoneCall aria-hidden="true" size={17} />
            Call 112
          </a>
        </div>
      </Drawer>

      <main className="citizen-main" id="main-content">
        <Outlet />
        <EmergencyBand />
      </main>
    </div>
  );
}

export default CitizenLayout;
