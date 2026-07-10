import { Stack } from "@mui/material";
import type { DispatchData } from "../../useDispatchData";
import { HospitalCapacityPanel } from "../HospitalCapacityPanel";
import { ReliefPointPanel } from "../ReliefPointPanel";

export function CapacityView({ data }: { data: DispatchData }) {
  return (
    <Stack spacing={2.5}>
      <HospitalCapacityPanel
        facilities={data.hospitalFacilities}
        filters={data.hospitalFilters}
        loadMessage={data.hospitalMessage}
        loadState={data.hospitalLoadState}
        onRefresh={() => void data.refreshHospitalCapacity()}
        onUpdateCapacity={data.updateHospitalCapacityFilter}
        onUpdateIncludeStale={data.updateHospitalIncludeStale}
        onUpdateMinBeds={data.updateHospitalMinBeds}
        onUpdateService={data.updateHospitalServiceFilter}
      />
      <ReliefPointPanel
        loadMessage={data.reliefPointMessage}
        loadState={data.reliefPointLoadState}
        onRefresh={() => void data.refreshReliefPoints()}
        reliefPoints={data.reliefPoints}
      />
    </Stack>
  );
}
