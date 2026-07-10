import { useState } from "react";
import { Box, Tab, Tabs } from "@mui/material";
import { LifeBuoy, PackageOpen } from "lucide-react";
import type { CommandData } from "../../useCommandData";
import { ShelterCapacityPanel } from "../ShelterCapacityPanel";
import { ReliefDistributionPanel } from "../ReliefDistributionPanel";

/**
 * Shelters + relief distribution as two tabs. Each tab owns its own searchable,
 * filterable, paginated register (see DataTable) plus the View / Edit / Delete
 * actions, so operators focus on one entity at a time instead of two crowded
 * side-by-side cards.
 */
export function SheltersView({ data }: { data: CommandData }) {
  const [tab, setTab] = useState(0);

  return (
    <Box>
      <Tabs
        value={tab}
        onChange={(_event, next) => setTab(next)}
        aria-label="Shelter and relief registers"
        sx={{ mb: 2 }}
      >
        <Tab
          icon={<LifeBuoy size={16} />}
          iconPosition="start"
          label="Shelters"
          id="shelters-tab-shelters"
          aria-controls="shelters-panel-shelters"
        />
        <Tab
          icon={<PackageOpen size={16} />}
          iconPosition="start"
          label="Relief distribution"
          id="shelters-tab-relief"
          aria-controls="shelters-panel-relief"
        />
      </Tabs>

      <Box
        role="tabpanel"
        hidden={tab !== 0}
        id="shelters-panel-shelters"
        aria-labelledby="shelters-tab-shelters"
      >
        {tab === 0 ? (
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
        ) : null}
      </Box>

      <Box
        role="tabpanel"
        hidden={tab !== 1}
        id="shelters-panel-relief"
        aria-labelledby="shelters-tab-relief"
      >
        {tab === 1 ? (
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
        ) : null}
      </Box>
    </Box>
  );
}
