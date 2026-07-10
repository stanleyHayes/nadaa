import { Box } from "@mui/material";
import type { DispatchData } from "../../useDispatchData";
import { MLPredictionReviewPanel } from "../MLPredictionReviewPanel";

export function MLReviewView({ data }: { data: DispatchData }) {
  return (
    <Box className="cc-view-single">
      <MLPredictionReviewPanel
        busy={data.mlDraftBusy}
        feedback={data.mlDraftFeedback}
        loadMessage={data.mlReviewMessage}
        loadState={data.mlReviewLoadState}
        onCreateDraft={() => void data.createAlertDraftFromPrediction()}
        onRefresh={() => void data.refreshMLPredictions()}
        onSelectPrediction={data.setSelectedPredictionId}
        onUpdateReviewNote={data.updatePredictionReviewNote}
        predictions={data.mlPredictions}
        reviewNote={data.selectedPredictionReviewNote}
        selectedPrediction={data.selectedPrediction}
        selectedPredictionId={data.selectedPredictionId}
      />
    </Box>
  );
}
