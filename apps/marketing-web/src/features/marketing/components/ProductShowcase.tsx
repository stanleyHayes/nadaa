import {
  BookOpen,
  LifeBuoy,
  Megaphone,
  Radar,
  ShieldCheck,
  Siren,
} from "lucide-react";
import type { ComponentType } from "react";
import { Reveal } from "./Reveal";

type Capability = {
  img: string;
  icon: ComponentType<{ size?: number | string; "aria-hidden"?: boolean }>;
  title: string;
  desc: string;
};

const capabilities: Capability[] = [
  {
    img: "/screens/cap-risk.webp",
    icon: Radar,
    title: "Check your risk",
    desc: "Live flood and hazard scoring for any area in Ghana, with the nearest shelters on the map.",
  },
  {
    img: "/screens/cap-alerts.webp",
    icon: Siren,
    title: "Get live alerts",
    desc: "Approved flood and hazard warnings for your area — current and past, filtered by severity.",
  },
  {
    img: "/screens/cap-report.webp",
    icon: Megaphone,
    title: "Report an incident",
    desc: "Tell NADMO what you're seeing, with photos and location — online or offline.",
  },
  {
    img: "/screens/cap-shelters.webp",
    icon: LifeBuoy,
    title: "Find shelters & routes",
    desc: "Locate the nearest safe shelter, plan an evacuation route, and check road closures.",
  },
  {
    img: "/screens/cap-guides.webp",
    icon: BookOpen,
    title: "Prepare with guides",
    desc: "Step-by-step flood and hazard guidance in six Ghanaian languages — available offline.",
  },
  {
    img: "/screens/cap-command.webp",
    icon: ShieldCheck,
    title: "Coordinated command",
    desc: "Behind every alert: an MFA-secured command center where officers verify and dispatch.",
  },
];

export function ProductShowcase() {
  return (
    <section aria-labelledby="showcase-title" className="content-section showcase">
      <div className="section-heading">
        <p className="eyebrow">See it in action</p>
        <h2 id="showcase-title">One app, every step to safety.</h2>
        <p>
          From checking your risk to getting the alert and finding a safe route
          — the citizen experience, plus the command center coordinating behind
          it.
        </p>
      </div>
      <div className="showcase-grid">
        {capabilities.map((cap, index) => {
          const Icon = cap.icon;
          return (
            <Reveal
              className="showcase-card"
              delay={index * 80}
              key={cap.title}
              variant="up"
            >
              <div className="showcase-frame">
                <span aria-hidden="true" className="showcase-dots">
                  <i />
                  <i />
                  <i />
                </span>
                <img
                  alt={`NADAA ${cap.title} screen`}
                  height={296}
                  loading="lazy"
                  src={cap.img}
                  width={1280}
                />
              </div>
              <div className="showcase-body">
                <span aria-hidden="true" className="showcase-badge">
                  <Icon size={20} />
                </span>
                <div>
                  <h3>{cap.title}</h3>
                  <p>{cap.desc}</p>
                </div>
              </div>
            </Reveal>
          );
        })}
      </div>
    </section>
  );
}
