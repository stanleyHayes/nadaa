import { Stack } from "@mui/material";
import { Building2 } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { AgencyGovernancePanel } from "../AgencyGovernancePanel";
import { EmptyState } from "../index";
import { ViewIntro } from "../primitives";

export function AgenciesView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={Building2}
        title="Agencies"
        description="Registered agencies, their operating scope, user counts, and MFA coverage."
      />
      {data.agencies.length ? (
        <AgencyGovernancePanel agencies={data.agencies} />
      ) : (
        <EmptyState
          title="No agencies"
          detail="The agency directory is not yet connected to this console."
        />
      )}
    </Stack>
  );
}
