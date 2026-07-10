import { Grid } from "@mui/material";
import type { CommandData } from "../../useCommandData";
import { ShelterCapacityPanel } from "../ShelterCapacityPanel";
import { ReliefDistributionPanel } from "../ReliefDistributionPanel";

export function SheltersView({ data }: { data: CommandData }) {
  return (
    <Grid container spacing={2.5} alignItems="flex-start">
      <Grid size={{ xs: 12, lg: 6 }}>
        <ShelterCapacityPanel
          shelters={data.shelters}
          shelterForm={data.shelterForm}
          loadState={data.shelterLoadState}
          feedback={data.shelterFeedback}
          busy={data.shelterBusy}
          canDelete={data.canDelete}
          onUpdateForm={data.updateShelterForm}
          onRefresh={() => void data.refreshShelters()}
          onEdit={data.editShelter}
          onSave={() => data.updateShelterCapacity()}
          onDelete={(shelter) => data.deleteShelter(shelter)}
        />
      </Grid>
      <Grid size={{ xs: 12, lg: 6 }}>
        <ReliefDistributionPanel
          reliefPoints={data.reliefPoints}
          reliefForm={data.reliefForm}
          reliefHistory={data.reliefHistory}
          loadState={data.reliefLoadState}
          feedback={data.reliefFeedback}
          busy={data.reliefBusy}
          canDelete={data.canDelete}
          onUpdateForm={data.updateReliefForm}
          onRefresh={() => void data.refreshReliefPoints()}
          onStartCreate={data.startReliefPointDraft}
          onStartEdit={data.editReliefPoint}
          onSave={() => data.saveReliefPoint()}
          onDelete={(point) => data.deleteReliefPoint(point)}
        />
      </Grid>
    </Grid>
  );
}
