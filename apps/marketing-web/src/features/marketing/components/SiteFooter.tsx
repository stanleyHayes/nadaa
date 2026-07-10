import { HeartHandshake, PhoneCall } from "lucide-react";
import { Link } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";

const footerNav = [
  { to: "/platforms", label: "Platforms" },
  { to: "/how-it-works", label: "How it works" },
  { to: "/trust", label: "Trust & compliance" },
  { to: "/signup", label: "Sign up" },
  { to: "/contact", label: "Contact" },
] as const;

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="footer-top">
        <Link className="footer-brand" to="/">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>{nadaaBrand.slogan}</small>
          </span>
        </Link>
        <nav aria-label="Footer" className="footer-nav">
          {footerNav.map((link) => (
            <Link key={link.to} to={link.to}>
              {link.label}
            </Link>
          ))}
        </nav>
        <a className="footer-cta" href={marketingLinks.partnerMail}>
          <HeartHandshake aria-hidden="true" size={18} />
          Partnership briefing
        </a>
      </div>
      <p className="footer-emergency">
        <PhoneCall aria-hidden="true" size={15} />
        In a life-threatening emergency, always call 112 first. NADAA supports
        NADMO and Ghana's 112 service — it does not replace it.
      </p>
      <div className="footer-legal">
        <small>
          {nadaaBrand.fullName}. Data handled under Ghana's Data Protection Act,
          2012 (Act 843).
        </small>
        <small>Be Aware. Be Prepared. Be Safe.</small>
      </div>
    </footer>
  );
}
