import {
  Languages,
  Menu,
  PhoneCall,
  ShieldCheck,
  UserPlus,
  X,
} from "lucide-react";
import { useState } from "react";
import { NavLink } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";

const pages = [
  { to: "/", label: "Home", end: true },
  { to: "/platforms", label: "Platforms", end: false },
  { to: "/how-it-works", label: "How it works", end: false },
  { to: "/trust", label: "Trust", end: false },
  { to: "/contact", label: "Contact", end: false },
] as const;

export function SiteHeader() {
  const [menuOpen, setMenuOpen] = useState(false);
  const navClass = ({ isActive }: { isActive: boolean }) =>
    isActive ? "is-active" : undefined;

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

        <nav
          aria-label="Primary"
          className={menuOpen ? "site-nav is-open" : "site-nav"}
          id="primary-nav"
        >
          {pages.map((page) => (
            <NavLink
              className={navClass}
              end={page.end}
              key={page.to}
              onClick={() => setMenuOpen(false)}
              to={page.to}
            >
              {page.label}
            </NavLink>
          ))}
        </nav>

        <div className="header-actions">
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
