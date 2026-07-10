import {
  BellRing,
  Building2,
  FileWarning,
  Radar,
  Route,
  ShieldCheck,
  UsersRound,
} from "lucide-react";
import { PageBanner } from "../components/PageBanner";
import { Reveal } from "../components/Reveal";
import { coreFeatures, roleSurfaces, serviceLines } from "../data";

const roleIcons = {
  citizen: UsersRound,
  authority: Radar,
  dispatcher: Route,
  agency: Building2,
  admin: ShieldCheck,
} as const;

const serviceIcons = {
  bell: BellRing,
  building: Building2,
  fileWarning: FileWarning,
  radar: Radar,
  route: Route,
  shield: ShieldCheck,
} as const;

export function PlatformsPage() {
  return (
    <>
      <PageBanner
        eyebrow="Platforms"
        subtitle="Five surfaces — citizen, command center, dispatch, response agency, and admin — each with its own job, access model, and channels, on one shared platform."
        title="Built for each role in a response."
      />

      <section className="content-section" aria-label="Role platforms">
        <div className="role-grid">
          {roleSurfaces.map((surface, index) => {
            const Icon = roleIcons[surface.icon];
            return (
              <Reveal
                className="role-card"
                delay={(index % 2) * 90}
                key={surface.role}
                variant="3d"
              >
                <header className="role-head">
                  <span className="role-icon" style={{ color: surface.accent }}>
                    <Icon aria-hidden="true" size={22} />
                  </span>
                  <div className="role-title">
                    <h2>{surface.role}</h2>
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
              </Reveal>
            );
          })}
        </div>
      </section>

      <section className="content-section" aria-labelledby="platforms-services">
        <div className="section-heading">
          <p className="eyebrow">Services</p>
          <h2 id="platforms-services">
            From public warning to accountable operations.
          </h2>
        </div>
        <div className="service-grid">
          {serviceLines.map((service, index) => {
            const Icon = serviceIcons[service.icon];
            return (
              <Reveal
                className="service-card"
                delay={(index % 3) * 80}
                key={service.title}
                variant="up"
              >
                <Icon size={24} />
                <h3>{service.title}</h3>
                <p>{service.description}</p>
              </Reveal>
            );
          })}
        </div>
      </section>

      <section className="content-section" aria-labelledby="platforms-features">
        <div className="section-heading">
          <p className="eyebrow">One response loop</p>
          <h2 id="platforms-features">Every feature serves the same job.</h2>
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
    </>
  );
}
