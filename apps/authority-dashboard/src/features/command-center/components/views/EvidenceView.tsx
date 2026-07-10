import { Stack } from "@mui/material";
import type { CommandData } from "../../useCommandData";
import { ImageryPanel } from "../ImageryPanel";
import { CVEvidencePanel } from "../CVEvidencePanel";

export function EvidenceView({ data }: { data: CommandData }) {
  return (
    <Stack spacing={2.5}>
      <ImageryPanel
        selectedIncidentId={data.selectedIncident?.id}
        showOverlay={data.showImageryOverlay}
        onToggleOverlay={data.setShowImageryOverlay}
        onFeaturesChange={data.setImageryFeatures}
      />
      <CVEvidencePanel />
    </Stack>
  );
}
