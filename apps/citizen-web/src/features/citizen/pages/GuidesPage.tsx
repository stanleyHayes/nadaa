import { PageBanner } from "../components/PageBanner";

/** Emergency guides (route `/guides`). Content migrates from the legacy `#guides` section. */
export function GuidesPage() {
  return (
    <>
      <PageBanner
        eyebrow="Emergency guidance"
        subtitle="Step-by-step flood and hazard guidance in six Ghanaian languages — available offline."
        title="Emergency preparedness guides"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default GuidesPage;
