import {
  ArrowRight,
  ChevronRight,
  PhoneCall,
  Radar,
  Radio,
  ShieldCheck,
  Truck,
  UsersRound,
} from "lucide-react";
import type { CSSProperties } from "react";
import { Link } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";
import { Reveal } from "../components/Reveal";
import { StatBand } from "../components/StatBand";
import {
  heroMetrics,
  platformPositioning,
  responseLoop,
  roleSurfaces,
} from "../data";
import { useParallax } from "../hooks";

const roleIcons = {
  citizen: UsersRound,
  authority: Radar,
  dispatcher: Radio,
  agency: Truck,
  admin: ShieldCheck,
} as const;

export function HomePage() {
  const parallaxRef = useParallax<HTMLDivElement>(0.12);

  return (
    <>
      <section aria-labelledby="hero-title" className="hero-section">
        <div className="hero-media" ref={parallaxRef} />
        <div className="hero-shade" />
        <div className="hero-content">
          <p className="eyebrow">One nation. One platform. One response.</p>
          <h1 id="hero-title">{nadaaBrand.name}</h1>
          <p className="hero-subtitle">
            Ghana's National Disaster Alert and Response Platform — early
            warnings, risk checks, incident reporting, command coordination, and
            recovery, in six Ghanaian languages.
          </p>
          <div className="hero-actions">
            <Link className="primary-action" to="/signup">
              Sign up as a citizen
              <ChevronRight aria-hidden="true" size={18} />
            </Link>
            <Link className="secondary-action" to="/platforms">
              Explore platforms
            </Link>
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
        <p>{platformPositioning}</p>
      </section>

      <StatBand />

      <section aria-labelledby="home-loop-title" className="content-section">
        <div className="section-heading">
          <p className="eyebrow">How it works</p>
          <h2 id="home-loop-title">From risk signal to recovery.</h2>
          <p>
            NADAA carries a flood or fire through one accountable loop — every
            public-safety decision stays in human hands.
          </p>
        </div>
        <ol className="loop-grid">
          {responseLoop.slice(0, 3).map((stage, index) => (
            <Reveal
              className="loop-card"
              delay={index * 110}
              key={stage.step}
              variant="up"
            >
              <span aria-hidden="true" className="loop-step">
                {stage.step}
              </span>
              <h3>{stage.title}</h3>
              <p>{stage.description}</p>
            </Reveal>
          ))}
        </ol>
        <Link className="section-link" to="/how-it-works">
          See the full response loop
          <ArrowRight aria-hidden="true" size={16} />
        </Link>
      </section>

      <section aria-labelledby="home-roles-title" className="content-section">
        <div className="section-heading">
          <p className="eyebrow">Platforms</p>
          <h2 id="home-roles-title">Built for each role in a response.</h2>
          <p>
            Citizen, command center, dispatcher, response agency, and admin —
            each with its own job, on one shared platform.
          </p>
        </div>
        <div className="role-lane-grid">
          {roleSurfaces.map((surface, index) => {
            const Icon = roleIcons[surface.icon];
            return (
              <Reveal
                className="role-lane"
                delay={index * 70}
                key={surface.role}
                style={{ "--role-accent": surface.accent } as CSSProperties}
                variant="3d"
              >
                <Icon
                  aria-hidden="true"
                  className="role-lane__watermark"
                  size={132}
                  strokeWidth={1.25}
                />
                <span aria-hidden="true" className="role-lane__chip">
                  <Icon size={20} strokeWidth={2} />
                </span>
                <p className="role-lane__kicker">{surface.audience}</p>
                <h3 className="role-lane__title">{surface.role}</h3>
                <p className="role-lane__copy">{surface.oneLiner}</p>
                <span aria-hidden="true" className="role-lane__more">
                  Explore
                  <ArrowRight size={14} />
                </span>
              </Reveal>
            );
          })}
        </div>
        <Link className="section-link" to="/platforms">
          Explore every platform
          <ArrowRight aria-hidden="true" size={16} />
        </Link>
      </section>

      <section className="cta-band" aria-labelledby="home-cta-title">
        <Reveal variant="scale">
          <div className="cta-card">
            <div>
              <p className="eyebrow">Get warnings where you are</p>
              <h2 id="home-cta-title">Sign up and stay ahead of the water.</h2>
              <p>
                Create a free citizen account to check your area's risk, get
                urgent warnings, and report incidents — online or offline.
              </p>
            </div>
            <div className="cta-actions">
              <Link className="primary-action" to="/signup">
                Sign up as a citizen
                <ChevronRight aria-hidden="true" size={18} />
              </Link>
              <a className="ghost-action" href={marketingLinks.emergencyPhone}>
                <PhoneCall aria-hidden="true" size={17} />
                Emergency? Call 112
              </a>
            </div>
          </div>
        </Reveal>
      </section>
    </>
  );
}
