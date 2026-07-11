import {
  ChevronRight,
  House,
  Languages,
  LayoutGrid,
  Mail,
  Menu,
  Moon,
  PhoneCall,
  ShieldCheck,
  Sun,
  UserPlus,
  Workflow,
  X,
} from "lucide-react";
import { type CSSProperties, useLayoutEffect, useRef, useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";
import { toggleThemeMode, useThemeMode } from "@/app/theme-mode";

const pages = [
  { to: "/", label: "Home", end: true, icon: House },
  { to: "/platforms", label: "Platforms", end: false, icon: LayoutGrid },
  { to: "/how-it-works", label: "How it works", end: false, icon: Workflow },
  { to: "/trust", label: "Trust", end: false, icon: ShieldCheck },
  { to: "/contact", label: "Contact", end: false, icon: Mail },
] as const;

export function SiteHeader() {
  const [menuOpen, setMenuOpen] = useState(false);
  const mode = useThemeMode();
  const isDark = mode === "dark";
  const navClass = ({ isActive }: { isActive: boolean }) =>
    isActive ? "is-active" : undefined;

  // Sliding pill indicator: measure the active desktop pill and move a single
  // highlight to it, so switching routes slides smoothly instead of jumping.
  const { pathname } = useLocation();
  const navRef = useRef<HTMLElement>(null);
  const [pillStyle, setPillStyle] = useState<CSSProperties>({ opacity: 0 });

  useLayoutEffect(() => {
    const nav = navRef.current;
    if (!nav) {
      return;
    }
    const measure = () => {
      const active = nav.querySelector<HTMLElement>("a.is-active");
      if (!active) {
        setPillStyle((prev) => ({ ...prev, opacity: 0 }));
        return;
      }
      setPillStyle({
        opacity: 1,
        width: active.offsetWidth,
        height: active.offsetHeight,
        top: active.offsetTop,
        transform: `translateX(${active.offsetLeft}px)`,
      });
    };
    measure();
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [pathname]);

  return (
    <>
      <div className="utility-bar">
        <span className="utility-org">
          <span className="status-dot" aria-hidden="true" />
          {nadaaBrand.fullName}
        </span>
        <span className="utility-actions">
          <span className="utility-lang">
            <Languages aria-hidden="true" size={14} />
            EN
          </span>
          <a className="utility-emergency" href={marketingLinks.emergencyPhone}>
            <PhoneCall aria-hidden="true" size={14} />
            Emergency? Call 112
          </a>
        </span>
      </div>

      <header className="site-header">
        <NavLink className="brand-mark" to="/" aria-label="NADAA home">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>
              <ShieldCheck aria-hidden="true" size={12} />
              Official NADMO platform
            </small>
          </span>
        </NavLink>

        {menuOpen ? (
          <button
            aria-label="Close menu"
            className="nav-scrim"
            onClick={() => setMenuOpen(false)}
            type="button"
          />
        ) : null}

        <nav
          aria-label="Primary"
          className={menuOpen ? "site-nav is-open" : "site-nav"}
          id="primary-nav"
          ref={navRef}
        >
          <span
            aria-hidden="true"
            className="nav-pill-indicator"
            style={pillStyle}
          />
          {pages.map((page) => {
            const Icon = page.icon;
            return (
              <NavLink
                className={navClass}
                end={page.end}
                key={page.to}
                onClick={() => setMenuOpen(false)}
                to={page.to}
              >
                <span aria-hidden="true" className="site-nav__icon">
                  <Icon size={18} />
                </span>
                <span className="site-nav__label">{page.label}</span>
                <ChevronRight
                  aria-hidden="true"
                  className="site-nav__chevron"
                  size={16}
                />
              </NavLink>
            );
          })}
        </nav>

        <div className="header-actions">
          <button
            type="button"
            className="theme-toggle"
            onClick={(event) => {
              const rect = event.currentTarget.getBoundingClientRect();
              toggleThemeMode({
                x: rect.left + rect.width / 2,
                y: rect.top + rect.height / 2,
              });
            }}
            aria-label={
              isDark ? "Switch to light theme" : "Switch to dark theme"
            }
            aria-pressed={isDark}
          >
            {isDark ? (
              <Sun aria-hidden="true" size={18} />
            ) : (
              <Moon aria-hidden="true" size={18} />
            )}
          </button>
          <NavLink className="cta-button" to="/signup">
            <UserPlus aria-hidden="true" size={16} />
            Sign up
          </NavLink>
          <a
            className="link-button emergency"
            href={marketingLinks.emergencyPhone}
          >
            <PhoneCall aria-hidden="true" size={17} />
            112
          </a>
          <button
            aria-controls="primary-nav"
            aria-expanded={menuOpen}
            aria-label={menuOpen ? "Close menu" : "Open menu"}
            className="icon-button"
            onClick={() => setMenuOpen((current) => !current)}
            type="button"
          >
            {menuOpen ? (
              <X aria-hidden="true" size={22} />
            ) : (
              <Menu aria-hidden="true" size={22} />
            )}
          </button>
        </div>
      </header>
    </>
  );
}
