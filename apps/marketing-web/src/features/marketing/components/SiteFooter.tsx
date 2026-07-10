import {
  FileWarning,
  HeartHandshake,
  House,
  Landmark,
  LayoutGrid,
  type LucideIcon,
  Mail,
  PhoneCall,
  Radar,
  ShieldCheck,
  UserPlus,
  Workflow,
} from "lucide-react";
import { Link } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";

type FooterLink = {
  label: string;
  icon: LucideIcon;
  to?: string;
  href?: string;
  external?: boolean;
};

const footerColumns: { heading: string; links: FooterLink[] }[] = [
  {
    heading: "Platform",
    links: [
      { label: "Platforms", to: "/platforms", icon: LayoutGrid },
      { label: "How it works", to: "/how-it-works", icon: Workflow },
      { label: "Trust & compliance", to: "/trust", icon: ShieldCheck },
      { label: "Sign up", to: "/signup", icon: UserPlus },
    ],
  },
  {
    heading: "Get help",
    links: [
      { label: "Check my risk", href: marketingLinks.citizenWeb, icon: Radar },
      { label: "Find a shelter", href: marketingLinks.citizenWeb, icon: House },
      {
        label: "Report an incident",
        href: marketingLinks.citizenWeb,
        icon: FileWarning,
      },
      {
        label: "Emergency — call 112",
        href: marketingLinks.emergencyPhone,
        icon: PhoneCall,
      },
    ],
  },
  {
    heading: "Organisation",
    links: [
      { label: "Contact", to: "/contact", icon: Mail },
      {
        label: "Partnerships & demos",
        href: marketingLinks.partnerMail,
        icon: HeartHandshake,
      },
      {
        label: "NADMO",
        href: "https://www.nadmo.gov.gh/",
        external: true,
        icon: Landmark,
      },
    ],
  },
];

function FooterItem({ link }: { link: FooterLink }) {
  const Icon = link.icon;
  const content = (
    <>
      <Icon aria-hidden="true" size={16} />
      {link.label}
    </>
  );
  if (link.to) {
    return <Link to={link.to}>{content}</Link>;
  }
  return (
    <a
      href={link.href}
      rel={link.external ? "noreferrer" : undefined}
      target={link.external ? "_blank" : undefined}
    >
      {content}
    </a>
  );
}

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="footer-grid">
        <div className="footer-brand-col">
          <Link className="footer-brand" to="/">
            <img alt="" src="/brand/nadaa-logo.png" />
            <span>
              <strong>{nadaaBrand.name}</strong>
              <small>{nadaaBrand.slogan}</small>
            </span>
          </Link>
          <p className="footer-mission">
            Ghana's National Disaster Alert and Response Platform — early
            warnings, reporting, and coordinated response, in six languages.
          </p>
          <a className="footer-cta" href={marketingLinks.partnerMail}>
            <HeartHandshake aria-hidden="true" size={18} />
            Partnership briefing
          </a>
        </div>

        {footerColumns.map((column) => (
          <nav
            aria-label={column.heading}
            className="footer-col"
            key={column.heading}
          >
            <h3>{column.heading}</h3>
            {column.links.map((link) => (
              <FooterItem key={link.label} link={link} />
            ))}
          </nav>
        ))}
      </div>

      <div className="footer-emergency" role="note">
        <span className="footer-emergency-badge" aria-hidden="true">
          <PhoneCall size={20} />
        </span>
        <div className="footer-emergency-copy">
          <strong>In a life-threatening emergency, call 112 first.</strong>
          <span>
            NADAA supports NADMO and Ghana's 112 service — it does not replace
            it.
          </span>
        </div>
        <a
          className="footer-emergency-call"
          href={marketingLinks.emergencyPhone}
        >
          <PhoneCall aria-hidden="true" size={18} />
          Call 112
        </a>
      </div>

      <div className="footer-legal">
        <small>
          © NADMO — {nadaaBrand.fullName}. Data handled under Ghana's Data
          Protection Act, 2012 (Act 843).
        </small>
        <small>Be Aware. Be Prepared. Be Safe.</small>
      </div>
    </footer>
  );
}
