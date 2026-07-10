import { Paper, Stack } from "@mui/material";
import { DatabaseZap } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { DataSourcePanel, EmptyState, ErrorState } from "../index";
import { SkeletonRows, ViewIntro } from "../primitives";

export function IntegrationsView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={DatabaseZap}
        title="Data sources"
        description="Integration contracts, refresh cadence, PII posture, and safe secret scopes for partner feeds."
      />
      {data.loadState === "loading" ? (
        <Paper className="surface">
          <SkeletonRows rows={4} />
        </Paper>
      ) : data.dataSources.length ? (
        <DataSourcePanel dataSources={data.dataSources} />
      ) : data.loadState === "error" ? (
        <ErrorState message={data.loadMessage} />
      ) : (
        <EmptyState
          title="No data sources"
          detail="No integration contracts are currently visible to the admin console."
        />
      )}
    </Stack>
  );
}
