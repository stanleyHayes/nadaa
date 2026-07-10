import { Languages, MapPinned, PhoneCall, Radio } from "lucide-react";
import { impactStats } from "../data";
import { AnimatedCounter } from "./AnimatedCounter";
import { Reveal } from "./Reveal";

const statIcons = {
  regions: MapPinned,
  languages: Languages,
  channels: Radio,
  emergency: PhoneCall,
} as const;

type StatBandProps = {
  plain?: boolean;
};

/** Gilded stat band with icon chips and count-up values. */
export function StatBand({ plain = false }: StatBandProps) {
  return (
    <section
      aria-label="Platform at a glance"
      className={plain ? "stats-band stats-band--plain" : "stats-band"}
    >
      <div className="stats-inner">
        {impactStats.map((stat, index) => {
          const Icon = statIcons[stat.icon];
          return (
            <Reveal
              className="stat-tile"
              delay={index * 90}
              key={stat.label}
              variant="up"
            >
              <span className="stat-chip" aria-hidden="true">
                <Icon size={18} />
              </span>
              <strong>
                <AnimatedCounter value={Number(stat.value)} />
              </strong>
              <span className="stat-label">{stat.label}</span>
              <span className="stat-detail">{stat.detail}</span>
            </Reveal>
          );
        })}
      </div>
    </section>
  );
}
