import { PhoneCall, Siren } from "lucide-react";

const EMERGENCY_TEL = "tel:112";

/**
 * Persistent, unmissable 112 band. Kept high-contrast and one tap from a call
 * so the core life-safety action is always obvious to an anonymous visitor.
 */
export function EmergencyBand() {
  return (
    <aside className="emergency-band" aria-label="Emergency contact">
      <span className="emergency-band__mark" aria-hidden="true">
        <Siren size={22} strokeWidth={2.2} />
      </span>
      <div className="emergency-band__text">
        <strong>In danger now? Call 112.</strong>
        <span>Police, fire, ambulance, NADMO and relief agencies — 24/7.</span>
      </div>
      <a className="emergency-band__call" href={EMERGENCY_TEL}>
        <PhoneCall aria-hidden="true" size={18} />
        Call 112
      </a>
    </aside>
  );
}

export default EmergencyBand;
