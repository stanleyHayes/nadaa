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
          selectedShelter={data.selectedShelter}
          loadState={data.shelterLoadState}
          feedback={data.shelterFeedback}
          busy={data.shelterBusy}
          onUpdateForm={data.updateShelterForm}
          onRefresh={() => void data.refreshShelters()}
          onUpdateCapacity={() => void data.updateShelterCapacity()}
        />
      </Grid>
      <Grid size={{ xs: 12, lg: 6 }}>
        <ReliefDistributionPanel
          reliefPoints={data.reliefPoints}
          reliefForm={data.reliefForm}
          selectedReliefPoint={data.selectedReliefPoint}
          reliefHistory={data.reliefHistory}
          loadState={data.reliefLoadState}
          feedback={data.reliefFeedback}
          busy={data.reliefBusy}
          onUpdateForm={data.updateReliefForm}
          onRefresh={() => void data.refreshReliefPoints()}
          onSave={() => void data.saveReliefPoint()}
        />
      </Grid>
    </Grid>
  );
}
