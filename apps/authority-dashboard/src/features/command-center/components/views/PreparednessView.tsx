import { Stack } from "@mui/material";
import CampaignManagerPanel from "../CampaignManagerPanel";
import { SchoolPreparednessPanel } from "../SchoolPreparednessPanel";

export function PreparednessView() {
  return (
    <Stack spacing={2.5}>
      <CampaignManagerPanel />
      <SchoolPreparednessPanel />
    </Stack>
  );
}
