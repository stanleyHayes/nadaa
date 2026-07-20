import { Alert, Paper, Stack } from "@mui/material";
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
        description="Approval, emergency override, targeting, and audit posture derived from live alerts — every public broadcast stays human-approved."
      />
      <Alert severity="info" variant="outlined">
        Read-only view: these cards are derived from live alert records so you
        can review the governance posture each broadcast carried. Alert rules
        cannot be created or edited from this console.
      </Alert>
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
