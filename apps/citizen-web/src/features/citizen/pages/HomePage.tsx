import { PageBanner } from "../components/PageBanner";

/** Risk checker landing (route `/`). Content migrates from the legacy `#risk` section. */
export function HomePage() {
  return (
    <>
      <PageBanner
        eyebrow="Flood risk checker"
        subtitle="See live flood and hazard risk for anywhere in Ghana, find the nearest shelter, and get emergency guidance — no sign-in needed."
        title="Check your area's risk"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default HomePage;
