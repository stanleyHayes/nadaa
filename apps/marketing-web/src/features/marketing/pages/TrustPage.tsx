import {
  Accessibility,
  Landmark,
  Lock,
  ShieldCheck,
} from "lucide-react";
import { PageBanner } from "../components/PageBanner";
import { Reveal } from "../components/Reveal";
import { SourcesDock } from "../components/SourcesDock";
import { StatBand } from "../components/StatBand";
import { complianceItems, trustPoints } from "../data";

const complianceIcons = [Lock, Landmark, ShieldCheck, Accessibility] as const;

export function TrustPage() {
  return (
    <>
      <PageBanner
        eyebrow="Trust & compliance"
        subtitle="NADAA supports NADMO and Ghana's 112 service. Public-safety decisions stay in human hands, personal data is minimised, and every sensitive action is auditable."
        title="Accountable, private, and built for Ghana."
      />

      <StatBand plain />

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
        <SourcesDock />
      </section>
    </>
  );
}
