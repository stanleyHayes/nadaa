import { useEffect, useState } from "react";
import { Menu, MenuItem, ListItemIcon, Divider } from "@mui/material";
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
import { nadaaBrand } from "@nadaa/brand";
import type { CitizenSession } from "../session";

const EMERGENCY_TEL = "tel:112";

const sections = [
  { id: "risk", label: "Risk" },
  { id: "alerts", label: "Alerts" },
  { id: "report", label: "Report" },
  { id: "resources", label: "Shelters" },
  { id: "guides", label: "Guides" },
  { id: "community", label: "Community" },
] as const;

type CitizenHeaderProps = {
  session: CitizenSession | null;
  onOpenSignIn: () => void;
  onSignOut: () => void;
};

/** Track which section is in view to light up the matching nav pill. */
function useActiveSection(): string {
  const [active, setActive] = useState<string>(sections[0].id);

  useEffect(() => {
    if (typeof IntersectionObserver === "undefined") {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
        if (visible) {
          setActive(visible.target.id);
        }
      },
      { rootMargin: "-45% 0px -50% 0px", threshold: [0, 0.2, 0.5] },
    );
    sections.forEach(({ id }) => {
      const element = document.getElementById(id);
      if (element) {
        observer.observe(element);
      }
    });
    return () => observer.disconnect();
  }, []);

  return active;
}

export function CitizenHeader({
  session,
  onOpenSignIn,
  onSignOut,
}: CitizenHeaderProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [userAnchor, setUserAnchor] = useState<null | HTMLElement>(null);
  const active = useActiveSection();

  const handleNavClick = () => setMenuOpen(false);

  const focusRisk = () => {
    setMenuOpen(false);
    window.setTimeout(() => {
      document.getElementById("risk-area")?.focus({ preventScroll: true });
    }, 480);
  };

  return (
    <>
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
        <a className="citizen-brand" href="#risk" aria-label="NADAA home">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>
              <ShieldCheck aria-hidden="true" size={12} />
              Be Aware. Be Prepared. Be Safe.
            </small>
          </span>
        </a>

        <nav
          aria-label="Citizen sections"
          className={menuOpen ? "citizen-nav is-open" : "citizen-nav"}
          id="citizen-primary-nav"
        >
          {sections.map((section) => (
            <a
              key={section.id}
              href={`#${section.id}`}
              className={active === section.id ? "is-active" : undefined}
              aria-current={active === section.id ? "true" : undefined}
              onClick={handleNavClick}
            >
              {section.label}
            </a>
          ))}
        </nav>

        <div className="citizen-header__actions">
          <a className="citizen-cta" href="#risk" onClick={focusRisk}>
            <MapPinned aria-hidden="true" size={16} />
            Check my risk
          </a>
          <a className="citizen-emergency-link" href={EMERGENCY_TEL}>
            <PhoneCall aria-hidden="true" size={17} />
            112
          </a>

          {session ? (
            <>
              <button
                type="button"
                className="citizen-user"
                aria-haspopup="true"
                aria-expanded={Boolean(userAnchor)}
                onClick={(event) => setUserAnchor(event.currentTarget)}
              >
                <UserRound aria-hidden="true" size={16} />
                <span>{session.name.split(" ")[0]}</span>
              </button>
              <Menu
                anchorEl={userAnchor}
                open={Boolean(userAnchor)}
                onClose={() => setUserAnchor(null)}
                anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
                transformOrigin={{ vertical: "top", horizontal: "right" }}
              >
                <MenuItem disabled>
                  {session.name} · {session.region}
                </MenuItem>
                <Divider />
                <MenuItem
                  onClick={() => {
                    setUserAnchor(null);
                    onSignOut();
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
              type="button"
              className="citizen-signin"
              onClick={onOpenSignIn}
            >
              <UserPlus aria-hidden="true" size={16} />
              <span>Sign in</span>
            </button>
          )}

          <button
            type="button"
            className="citizen-menu-toggle"
            aria-controls="citizen-primary-nav"
            aria-expanded={menuOpen}
            aria-label={menuOpen ? "Close menu" : "Open menu"}
            onClick={() => setMenuOpen((current) => !current)}
          >
            {menuOpen ? (
              <X aria-hidden="true" size={22} />
            ) : (
              <MenuIcon aria-hidden="true" size={22} />
            )}
          </button>
        </div>
      </header>
    </>
  );
}

export default CitizenHeader;
