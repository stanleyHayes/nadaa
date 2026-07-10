import { HeartHandshake, PhoneCall } from "lucide-react";
import { Link } from "react-router-dom";
import { nadaaBrand } from "@nadaa/brand";
import { marketingLinks } from "@/app/config";

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="footer-grid">
        <div className="footer-brand-col">
          <Link className="footer-brand" to="/">
            <img alt="" src="/brand/nadaa-logo.png" />
            <span>
              <strong>{nadaaBrand.name}</strong>
              <small>{nadaaBrand.slogan}</small>
            </span>
          </Link>
          <p className="footer-mission">
            Ghana's National Disaster Alert and Response Platform — early
            warnings, reporting, and coordinated response, in six languages.
          </p>
          <a className="footer-cta" href={marketingLinks.partnerMail}>
            <HeartHandshake aria-hidden="true" size={18} />
            Partnership briefing
          </a>
        </div>

        <nav aria-label="Platform" className="footer-col">
          <h3>Platform</h3>
          <Link to="/platforms">Platforms</Link>
          <Link to="/how-it-works">How it works</Link>
          <Link to="/trust">Trust &amp; compliance</Link>
          <Link to="/signup">Sign up</Link>
        </nav>

        <nav aria-label="Get help" className="footer-col">
          <h3>Get help</h3>
          <a href={marketingLinks.citizenWeb}>Check my risk</a>
          <a href={marketingLinks.citizenWeb}>Find a shelter</a>
          <a href={marketingLinks.citizenWeb}>Report an incident</a>
          <a href={marketingLinks.emergencyPhone}>Emergency — call 112</a>
        </nav>

        <nav aria-label="Organisation" className="footer-col">
          <h3>Organisation</h3>
          <Link to="/contact">Contact</Link>
          <a href={marketingLinks.partnerMail}>Partnerships &amp; demos</a>
          <a href="https://www.nadmo.gov.gh/" rel="noreferrer" target="_blank">
            NADMO
          </a>
        </nav>
      </div>

      <div className="footer-emergency" role="note">
        <span className="footer-emergency-badge" aria-hidden="true">
          <PhoneCall size={20} />
        </span>
        <div className="footer-emergency-copy">
          <strong>In a life-threatening emergency, call 112 first.</strong>
          <span>
            NADAA supports NADMO and Ghana's 112 service — it does not replace
            it.
          </span>
        </div>
        <a
          className="footer-emergency-call"
          href={marketingLinks.emergencyPhone}
        >
          <PhoneCall aria-hidden="true" size={18} />
          Call 112
        </a>
      </div>

      <div className="footer-legal">
        <small>
          © NADMO — {nadaaBrand.fullName}. Data handled under Ghana's Data
          Protection Act, 2012 (Act 843).
        </small>
        <small>Be Aware. Be Prepared. Be Safe.</small>
      </div>
    </footer>
  );
}
