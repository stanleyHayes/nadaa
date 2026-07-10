import {
  Accessibility,
  AlertTriangle,
  BellRing,
  Building2,
  CheckCircle2,
  ChevronRight,
  FileWarning,
  HeartHandshake,
  Landmark,
  Languages,
  Lock,
  Mail,
  MapPinned,
  Menu,
  PhoneCall,
  Radar,
  Route,
  ShieldCheck,
  UsersRound,
  X,
} from "lucide-react";
import { useEffect, useState } from "react";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";
import {
  benefits,
  complianceItems,
  contactCards,
  coreFeatures,
  differentiators,
  heroMetrics,
  impactStats,
  legalLinks,
  navItems,
  platformPositioning,
  researchNotes,
  responseLoop,
  roleSurfaces,
  serviceLines,
  trustPoints,
} from "./data";

const serviceIcons = {
  bell: BellRing,
  building: Building2,
  fileWarning: FileWarning,
  radar: Radar,
  route: Route,
  shield: ShieldCheck,
} as const;

const roleIcons = {
  citizen: UsersRound,
  authority: Radar,
  dispatcher: Route,
  agency: Building2,
  admin: ShieldCheck,
} as const;

const complianceIcons = [Lock, Landmark, ShieldCheck, Accessibility] as const;

export default function MarketingApp() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [activeSection, setActiveSection] = useState<string>("");

  useEffect(() => {
    const sections = navItems
      .map((item) => document.getElementById(item.href.replace("#", "")))
      .filter((element): element is HTMLElement => element !== null);
    if (sections.length === 0) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
        if (visible) {
          setActiveSection(visible.target.id);
        }
      },
      { rootMargin: "-45% 0px -50% 0px", threshold: [0, 0.2, 0.5, 1] },
    );
    sections.forEach((section) => observer.observe(section));
    return () => observer.disconnect();
  }, []);

  return (
    <div className="site-shell">
      <a className="skip-link" href="#main-content">
        Skip to main content
      </a>

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
        <a className="brand-mark" href="#main-content" aria-label="NADAA home">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>
              <ShieldCheck aria-hidden="true" size={12} />
              Official NADMO platform
            </small>
          </span>
        </a>

        <nav
          aria-label="Primary"
          className={menuOpen ? "site-nav is-open" : "site-nav"}
          id="primary-nav"
        >
          {navItems.map((item) => {
            const isActive = activeSection === item.href.replace("#", "");
            return (
              <a
                aria-current={isActive ? "true" : undefined}
                className={isActive ? "is-active" : undefined}
                href={item.href}
                key={item.href}
                onClick={() => setMenuOpen(false)}
              >
                {item.label}
              </a>
            );
          })}
        </nav>

        <div className="header-actions">
          <a className="cta-button" href={marketingLinks.citizenWeb}>
            <BellRing aria-hidden="true" size={16} />
            Get alerts
          </a>
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

      <main id="main-content">
        <section className="hero-section" aria-labelledby="hero-title">
          <div className="hero-shade" />
          <div className="hero-content">
            <p className="eyebrow">One nation. One platform. One response.</p>
            <h1 id="hero-title">{nadaaBrand.name}</h1>
            <p className="hero-subtitle">
              {nadaaBrand.fullName} for early warnings, risk intelligence,
              incident reporting, command coordination, and recovery support
              across Ghana.
            </p>
            <div className="hero-actions">
              <a className="primary-action" href={marketingLinks.citizenWeb}>
                Open citizen app
                <ChevronRight aria-hidden="true" size={18} />
              </a>
              <a className="secondary-action" href="#platforms">
                Explore platforms
              </a>
            </div>
            <div className="hero-metrics" aria-label="Platform summary">
              {heroMetrics.map((metric) => (
                <span key={metric.label}>
                  <strong>{metric.value}</strong>
                  <small>{metric.label}</small>
                </span>
              ))}
            </div>
          </div>
        </section>

        <section className="statement-band" aria-label="Platform mission">
          <div>
            <p className="eyebrow">Be aware. Be prepared. Be safe.</p>
            <h2>Care, protection, and safety for communities.</h2>
          </div>
          <p>
            NADAA connects public alerts, citizen reports, operational
            dashboards, responder tools, relief logistics, and governance
            controls into a coordinated national disaster intelligence platform.
          </p>
        </section>

        <section className="section-grid about-section" id="about">
          <div className="section-copy">
            <p className="eyebrow">About NADAA</p>
            <h2>
              Built for preparedness before, during, and after emergencies.
            </h2>
            <p>
              NADAA is designed as Ghana's shared disaster alert and response
              layer. Citizens get practical warnings and ways to report. Public
              agencies get verified incident context and role-specific tools.
              Admin teams get governance controls that keep sensitive workflows
              accountable.
            </p>
          </div>
          <div className="about-panel">
            <CheckCircle2 size={28} />
            <h3>What makes it different</h3>
            <p>
              The platform is not one overloaded dashboard. It is a family of
              web and mobile products for the people who need to act: citizens,
              dispatchers, response agencies, and administrators.
            </p>
          </div>
        </section>

        <section
          aria-labelledby="loop-title"
          className="content-section loop-section"
          id="how"
        >
          <div className="section-heading">
            <p className="eyebrow">How it works</p>
            <h2 id="loop-title">One response loop, end to end.</h2>
            <p>{platformPositioning}</p>
          </div>
          <ol className="loop-grid">
            {responseLoop.map((stage) => (
              <li className="loop-card" key={stage.step}>
                <span aria-hidden="true" className="loop-step">
                  {stage.step}
                </span>
                <h3>{stage.title}</h3>
                <p>{stage.description}</p>
              </li>
            ))}
          </ol>
        </section>

        <section className="content-section" id="features">
          <div className="section-heading">
            <p className="eyebrow">Features</p>
            <h2>Every feature supports one response loop.</h2>
            <p>
              Detect risk, warn people, receive reports, coordinate teams,
              support recovery, and improve national readiness.
            </p>
          </div>
          <div className="feature-grid">
            {coreFeatures.map((feature) => (
              <article className="feature-card" key={feature.title}>
                <span style={{ backgroundColor: feature.accent }} />
                <h3>{feature.title}</h3>
                <p>{feature.description}</p>
              </article>
            ))}
          </div>
        </section>

        <section
          aria-labelledby="platforms-title"
          className="content-section platform-section"
          id="platforms"
        >
          <div className="section-heading">
            <p className="eyebrow">Platforms</p>
            <h2 id="platforms-title">Built for each role in a response.</h2>
            <p>
              Five surfaces — citizen, command center, dispatch, response
              agency, and admin — each with its own job, access model, and
              channels, on one shared platform.
            </p>
          </div>
          <div className="role-grid">
            {roleSurfaces.map((surface) => {
              const Icon = roleIcons[surface.icon];
              return (
                <article className="role-card" key={surface.role}>
                  <header className="role-head">
                    <span
                      className="role-icon"
                      style={{ color: surface.accent }}
                    >
                      <Icon aria-hidden="true" size={22} />
                    </span>
                    <div className="role-title">
                      <h3>{surface.role}</h3>
                      <p className="role-audience">{surface.audience}</p>
                    </div>
                  </header>
                  <p className="role-tagline" style={{ color: surface.accent }}>
                    {surface.tagline}
                  </p>
                  <p className="role-oneliner">{surface.oneLiner}</p>
                  <ul className="role-caps">
                    {surface.capabilities.map((cap) => (
                      <li key={cap.title}>
                        <strong>{cap.title}.</strong> {cap.description}
                      </li>
                    ))}
                  </ul>
                  <div className="role-channels" aria-label="Channels">
                    {surface.channels.map((channel) => (
                      <span key={channel}>{channel}</span>
                    ))}
                  </div>
                </article>
              );
            })}
          </div>
        </section>

        <section
          aria-labelledby="why-title"
          className="content-section why-section"
          id="why"
        >
          <div className="section-heading">
            <p className="eyebrow">Why NADAA</p>
            <h2 id="why-title">Designed for real emergencies, not demos.</h2>
          </div>
          <div className="why-grid">
            {differentiators.map((item) => (
              <article className="why-card" key={item.title}>
                <CheckCircle2 aria-hidden="true" size={22} />
                <h3>{item.title}</h3>
                <p>{item.description}</p>
              </article>
            ))}
          </div>
        </section>

        <section className="content-section" id="services">
          <div className="section-heading">
            <p className="eyebrow">Services</p>
            <h2>From public warning to accountable operations.</h2>
          </div>
          <div className="service-grid">
            {serviceLines.map((service) => {
              const Icon = serviceIcons[service.icon];
              return (
                <article className="service-card" key={service.title}>
                  <Icon size={24} />
                  <h3>{service.title}</h3>
                  <p>{service.description}</p>
                </article>
              );
            })}
          </div>
        </section>

        <section className="research-section" aria-labelledby="research-title">
          <div className="section-heading compact">
            <p className="eyebrow">Research context</p>
            <h2 id="research-title">Grounded in Ghana's emergency reality.</h2>
          </div>
          <div className="research-grid">
            {researchNotes.map((note) => (
              <article className="research-card" key={note.title}>
                <AlertTriangle size={22} />
                <h3>{note.title}</h3>
                <p>{note.body}</p>
                <a href={note.href} rel="noreferrer" target="_blank">
                  {note.source}
                </a>
              </article>
            ))}
          </div>
        </section>

        <section className="content-section benefits-section" id="benefits">
          <div className="section-heading">
            <p className="eyebrow">Benefits</p>
            <h2>Safer decisions for every role.</h2>
          </div>
          <div className="benefit-grid">
            {benefits.map((benefit) => (
              <article className="benefit-card" key={benefit.audience}>
                <UsersRound size={24} />
                <h3>{benefit.audience}</h3>
                <ul>
                  {benefit.points.map((point) => (
                    <li key={point}>{point}</li>
                  ))}
                </ul>
              </article>
            ))}
          </div>
        </section>

        <section
          aria-labelledby="trust-title"
          className="trust-section"
          id="trust"
        >
          <div className="trust-inner">
            <div className="section-heading compact">
              <p className="eyebrow">Trust &amp; compliance</p>
              <h2 id="trust-title">
                Accountable, private, and built for Ghana.
              </h2>
            </div>
            <div className="impact-strip" aria-label="Platform at a glance">
              {impactStats.map((stat) => (
                <div className="impact-item" key={stat.label}>
                  <strong>{stat.value}</strong>
                  <span>{stat.label}</span>
                </div>
              ))}
            </div>
            <div className="compliance-grid">
              {complianceItems.map((item, index) => {
                const Icon = complianceIcons[index] ?? ShieldCheck;
                return (
                  <article className="compliance-card" key={item.title}>
                    <Icon aria-hidden="true" size={22} />
                    <h3>{item.title}</h3>
                    <p>{item.description}</p>
                  </article>
                );
              })}
            </div>
            <ul className="trust-points">
              {trustPoints.map((point) => (
                <li key={point}>
                  <ShieldCheck aria-hidden="true" size={16} />
                  <span>{point}</span>
                </li>
              ))}
            </ul>
          </div>
        </section>

        <section className="contact-section" id="contact">
          <div className="contact-copy">
            <p className="eyebrow">Contact</p>
            <h2>Start with the right path.</h2>
            <p>
              Immediate emergencies belong on 112. Platform partnerships,
              deployments, and demos belong in the pilot onboarding lane.
            </p>
          </div>
          <div className="contact-grid">
            {contactCards.map((card) => (
              <a className="contact-card" href={card.href} key={card.title}>
                {card.title === "Emergency help" ? (
                  <PhoneCall aria-hidden="true" size={24} />
                ) : card.title === "Partnerships and demos" ? (
                  <Mail aria-hidden="true" size={24} />
                ) : (
                  <MapPinned aria-hidden="true" size={24} />
                )}
                <span>{card.title}</span>
                <strong>{card.primary}</strong>
                <p>{card.detail}</p>
              </a>
            ))}
          </div>
        </section>
      </main>

      <footer className="site-footer">
        <div className="footer-top">
          <div className="footer-brand">
            <img alt="" src="/brand/nadaa-logo.png" />
            <span>
              <strong>{nadaaBrand.name}</strong>
              <small>{nadaaBrand.slogan}</small>
            </span>
          </div>
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
          <nav aria-label="Legal">
            {legalLinks.map((link) => (
              <a href={link.href} key={link.label}>
                {link.label}
              </a>
            ))}
          </nav>
          <small>
            {nadaaBrand.fullName}. Data handled under Ghana's Data Protection
            Act, 2012 (Act 843).
          </small>
        </div>
      </footer>
    </div>
  );
}
