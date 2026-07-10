import { Paper, Stack } from "@mui/material";
import { ScrollText } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { AuditLogPanel } from "../AuditLogPanel";
import { SkeletonRows, ViewIntro } from "../primitives";

export function AuditView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={ScrollText}
        title="Audit trail"
        description="Sensitive-action trace across the platform: who changed what, when, and from where."
      />
      {data.loadState === "loading" ? (
        <Paper className="surface">
          <SkeletonRows rows={6} />
        </Paper>
      ) : (
        <AuditLogPanel logs={data.auditLogs} />
      )}
    </Stack>
  );
}
