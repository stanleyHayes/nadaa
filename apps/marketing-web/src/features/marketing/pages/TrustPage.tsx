import {
  Accessibility,
  AlertTriangle,
  Landmark,
  Lock,
  ShieldCheck,
} from "lucide-react";
import { AnimatedCounter } from "../components/AnimatedCounter";
import { Reveal } from "../components/Reveal";
import {
  complianceItems,
  impactStats,
  researchNotes,
  trustPoints,
} from "../data";

const complianceIcons = [Lock, Landmark, ShieldCheck, Accessibility] as const;

export function TrustPage() {
  return (
    <>
      <section className="page-head">
        <p className="eyebrow">Trust &amp; compliance</p>
        <h1>Accountable, private, and built for Ghana.</h1>
        <p>
          NADAA supports NADMO and Ghana's 112 service. Public-safety decisions
          stay in human hands, personal data is minimised, and every sensitive
          action is auditable.
        </p>
      </section>

      <section
        className="stats-band stats-band--plain"
        aria-label="At a glance"
      >
        <div className="stats-inner">
          {impactStats.map((stat, index) => (
            <Reveal
              className="stat-tile"
              delay={index * 80}
              key={stat.label}
              variant="up"
            >
              <strong>
                <AnimatedCounter value={Number(stat.value)} />
              </strong>
              <span>{stat.label}</span>
            </Reveal>
          ))}
        </div>
      </section>

      <section className="content-section" aria-label="Compliance">
        <div className="compliance-grid compliance-grid--light">
          {complianceItems.map((item, index) => {
            const Icon = complianceIcons[index] ?? ShieldCheck;
            return (
              <Reveal
                className="compliance-card compliance-card--light"
                delay={(index % 2) * 80}
                key={item.title}
                variant="up"
              >
                <Icon aria-hidden="true" size={22} />
                <h3>{item.title}</h3>
                <p>{item.description}</p>
              </Reveal>
            );
          })}
        </div>
        <ul className="trust-points trust-points--light">
          {trustPoints.map((point) => (
            <li key={point}>
              <ShieldCheck aria-hidden="true" size={16} />
              <span>{point}</span>
            </li>
          ))}
        </ul>
      </section>

      <section aria-labelledby="trust-research" className="content-section">
        <div className="section-heading compact">
          <p className="eyebrow">Research context</p>
          <h2 id="trust-research">Grounded in Ghana's emergency reality.</h2>
        </div>
        <div className="research-grid research-grid--light">
          {researchNotes.map((note) => (
            <article
              className="research-card research-card--light"
              key={note.title}
            >
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
    </>
  );
}
