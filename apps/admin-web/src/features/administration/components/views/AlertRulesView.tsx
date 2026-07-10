import { Stack } from "@mui/material";
import { BellRing } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { AlertRulePanel } from "../AlertRulePanel";
import { ViewIntro } from "../primitives";

export function AlertRulesView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={BellRing}
        title="Alert rules"
        description="Approval, emergency override, targeting, and audit controls that keep every public broadcast human-approved."
      />
      <AlertRulePanel rules={data.alertRules} />
    </Stack>
  );
}
