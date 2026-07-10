import { Stack } from "@mui/material";
import { DatabaseZap } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { DataSourcePanel, EmptyState } from "../index";
import { ViewIntro } from "../primitives";

export function IntegrationsView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={DatabaseZap}
        title="Data sources"
        description="Integration contracts, refresh cadence, PII posture, and safe secret scopes for partner feeds."
      />
      {data.dataSources.length ? (
        <DataSourcePanel dataSources={data.dataSources} />
      ) : (
        <EmptyState
          title="No data sources"
          detail="No integration contracts are currently visible to the admin console."
        />
      )}
    </Stack>
  );
}
