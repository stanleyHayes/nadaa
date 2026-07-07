import {
  AlertTriangle,
  BellRing,
  Building2,
  CheckCircle2,
  ChevronRight,
  FileWarning,
  HeartHandshake,
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
import { useState } from "react";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "../../app/config";
import {
  benefits,
  contactCards,
  coreFeatures,
  heroMetrics,
  navItems,
  platformLanes,
  researchNotes,
  serviceLines,
} from "./data";

const serviceIcons = {
  bell: BellRing,
  building: Building2,
  fileWarning: FileWarning,
  radar: Radar,
  route: Route,
  shield: ShieldCheck,
} as const;

export default function MarketingApp() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div className="site-shell">
      <header className="site-header">
        <a className="brand-mark" href="#top" aria-label="NADAA home">
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>{nadaaBrand.country}</small>
          </span>
        </a>

        <nav className={menuOpen ? "site-nav is-open" : "site-nav"}>
          {navItems.map((item) => (
            <a
              href={item.href}
              key={item.href}
              onClick={() => setMenuOpen(false)}
            >
              {item.label}
            </a>
          ))}
        </nav>

        <div className="header-actions">
          <a
            className="link-button emergency"
            href={marketingLinks.emergencyPhone}
          >
            <PhoneCall size={17} />
            112
          </a>
          <button
            aria-label={menuOpen ? "Close menu" : "Open menu"}
            className="icon-button"
            onClick={() => setMenuOpen((current) => !current)}
            type="button"
          >
            {menuOpen ? <X size={22} /> : <Menu size={22} />}
          </button>
        </div>
      </header>

      <main id="top">
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
                <ChevronRight size={18} />
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

        <section className="content-section platform-section" id="platforms">
          <div className="section-heading">
            <p className="eyebrow">Platforms</p>
            <h2>Separate surfaces for separate responsibilities.</h2>
            <p>
              Each product has its own job, access model, and release path.
              Shared brand and contracts keep the ecosystem consistent.
            </p>
          </div>
          <div className="platform-grid">
            {platformLanes.map((platform) => (
              <article className="platform-card" key={platform.title}>
                <div>
                  <span>{platform.channels}</span>
                  <h3>{platform.title}</h3>
                </div>
                <p>{platform.summary}</p>
                <strong>{platform.status}</strong>
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
                  <PhoneCall size={24} />
                ) : card.title === "Partnerships and demos" ? (
                  <Mail size={24} />
                ) : (
                  <MapPinned size={24} />
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
        <div>
          <img alt="" src="/brand/nadaa-logo.png" />
          <span>
            <strong>{nadaaBrand.name}</strong>
            <small>{nadaaBrand.slogan}</small>
          </span>
        </div>
        <p>
          Public marketing surface for the Ghana National Disaster Alert and
          Response Platform.
        </p>
        <a href={marketingLinks.partnerMail}>
          <HeartHandshake size={18} />
          Partnership briefing
        </a>
      </footer>
    </div>
  );
}
