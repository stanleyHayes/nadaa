import { PageBanner } from "../components/PageBanner";

/** Shelters & relief (route `/shelters`). Content migrates from the legacy `#resources` section. */
export function SheltersPage() {
  return (
    <>
      <PageBanner
        eyebrow="Shelters & relief"
        subtitle="Find the nearest shelters and relief points, safe evacuation routes, and current road closures."
        title="Find shelter & relief nearby"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default SheltersPage;
