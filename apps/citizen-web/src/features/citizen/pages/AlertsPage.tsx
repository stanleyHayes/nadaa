import { PageBanner } from "../components/PageBanner";

/** Alerts feed (route `/alerts`). Content migrates from the legacy `#alerts` section. */
export function AlertsPage() {
  return (
    <>
      <PageBanner
        eyebrow="Warnings & alerts"
        subtitle="Approved flood and hazard warnings for your area — current and past."
        title="Live flood & hazard alerts"
      />
      <div className="citizen-shell">
        <p className="page-placeholder">This page's content is being migrated.</p>
      </div>
    </>
  );
}

export default AlertsPage;
