import { Paper, Stack } from "@mui/material";
import { BellRing } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { AlertRulePanel } from "../AlertRulePanel";
import { EmptyState, ErrorState } from "../index";
import { SkeletonRows, ViewIntro } from "../primitives";

export function AlertRulesView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={BellRing}
        title="Alert rules"
        description="Approval, emergency override, targeting, and audit controls that keep every public broadcast human-approved."
      />
      {data.loadState === "loading" ? (
        <Paper className="surface">
          <SkeletonRows rows={4} />
        </Paper>
      ) : data.alertRules.length ? (
        <AlertRulePanel rules={data.alertRules} />
      ) : data.loadState === "error" ? (
        <ErrorState message={data.loadMessage} onRetry={data.refresh} />
      ) : (
        <EmptyState
          title="No alert rules"
          detail="No alert governance rules are available from the alert service yet."
        />
      )}
    </Stack>
  );
}
