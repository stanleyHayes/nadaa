import { PageBanner } from "../components/PageBanner";

/** Incident reporting (route `/report`). Content migrates from the legacy `#report` section. */
export function ReportPage() {
  return (
    <>
      <PageBanner
        eyebrow="Report an incident"
        subtitle="Tell NADMO what you're seeing — with photos and location, online or offline."
        title="Report a flood or hazard"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default ReportPage;
