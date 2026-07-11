import { Link } from "react-router-dom";
import { PageBanner } from "../components/PageBanner";

/**
 * Branded 404 rendered inside the site chrome (header + footer stay), replacing
 * the old silent redirect to home. Offers the main ways back plus the standing
 * emergency line, since this is a public-safety site.
 */
export function NotFoundPage() {
  return (
    <>
      <PageBanner
        eyebrow="404"
        title="This page could not be found."
        subtitle="The page may have moved or the link may be incomplete. Every NADAA response still starts from the home page — let's get you back."
      />

      <section className="content-section" aria-label="Page not found">
        <div className="section-heading">
          <p className="eyebrow">Get back on track</p>
          <h2>Where would you like to go?</h2>
        </div>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "12px",
            alignItems: "center",
          }}
        >
          <Link className="cta-button" to="/">
            Back to home
          </Link>
          <Link className="link-button" to="/platforms">
            Explore platforms
          </Link>
          <Link className="link-button" to="/contact">
            Contact NADAA
          </Link>
        </div>
        <p style={{ marginTop: "20px", color: "var(--nadaa-text-secondary)" }}>
          In an emergency, call <a href="tel:112">112</a>.
        </p>
      </section>
    </>
  );
}
