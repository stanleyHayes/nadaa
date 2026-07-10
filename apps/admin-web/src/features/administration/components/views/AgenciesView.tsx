import { Stack } from "@mui/material";
import { Building2 } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { AgencyGovernancePanel } from "../AgencyGovernancePanel";
import { ViewIntro } from "../primitives";

export function AgenciesView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={Building2}
        title="Agencies"
        description="Registered agencies, their operating scope, user counts, and MFA coverage."
      />
      <AgencyGovernancePanel agencies={data.agencies} />
    </Stack>
  );
}
