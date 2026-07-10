import { Box } from "@mui/material";
import type { DispatchData } from "../../useDispatchData";
import { AlertWorkflowPanel } from "../AlertWorkflowPanel";

export function AlertsView({ data }: { data: DispatchData }) {
  return (
    <Box className="cc-view-single">
      <AlertWorkflowPanel
        alerts={data.alerts}
        busy={data.alertBusy}
        feedback={data.alertFeedback || data.alertMessage}
        form={data.alertForm}
        loadState={data.alertLoadState}
        onCreateDraft={data.createAlertDraft}
        onRunAction={data.runAlertAction}
        onUpdateForm={data.updateAlertForm}
        selectedIncident={data.selectedIncident}
      />
    </Box>
  );
}
