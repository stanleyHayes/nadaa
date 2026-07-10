import { Mail, MapPinned, PhoneCall } from "lucide-react";
import { PageBanner } from "../components/PageBanner";
import { Reveal } from "../components/Reveal";
import { contactCards } from "../data";

export function ContactPage() {
  return (
    <>
      <PageBanner
        eyebrow="Contact"
        subtitle="Immediate emergencies belong on 112. Platform partnerships, deployments, and demos belong in the onboarding lane."
        title="Start with the right path."
      />

      <section className="content-section" aria-label="Contact options">
        <div className="contact-grid">
          {contactCards.map((card, index) => (
            <Reveal delay={(index % 3) * 80} key={card.title} variant="up">
              <a className="contact-card" href={card.href}>
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
            </Reveal>
          ))}
        </div>
      </section>
    </>
  );
}
