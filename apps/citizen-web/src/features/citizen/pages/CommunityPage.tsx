import { PageBanner } from "../components/PageBanner";

/** Community & recovery (route `/community`). Content migrates from the legacy `#community` section. */
export function CommunityPage() {
  return (
    <>
      <PageBanner
        eyebrow="Community & recovery"
        subtitle="Donate, claim damage support, reunite families, and explore open disaster data."
        title="Community & recovery"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default CommunityPage;
