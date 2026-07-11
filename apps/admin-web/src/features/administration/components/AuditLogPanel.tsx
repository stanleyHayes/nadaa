import { Paper, Typography } from "@mui/material";
import { ShieldCheck } from "lucide-react";
import type { AuditLogRecord } from "@nadaa/shared-types";
import {
  auditSnapshotSummary,
  auditTargetSummary,
  formatDateTime,
} from "../utils";
import { DataTable } from "./DataTable";
import { EmptyState, SectionHeader } from "./shared";

export function AuditLogPanel({ logs }: { logs: AuditLogRecord[] }) {
  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Audit trail"
        title="Sensitive action trace"
        icon={<ShieldCheck size={22} color="var(--nadaa-navy)" />}
      />
      {logs.length ? (
        <DataTable
          rows={logs}
          getRowKey={(log) => log.id}
          searchOf={(log) =>
            `${log.action} ${log.actorRole ?? "system"} ${
              log.actorUserId ?? "anonymous"
            } ${auditTargetSummary(log)}`
          }
          searchPlaceholder="Search action, actor, or target"
          filters={[
            {
              key: "action",
              label: "Action",
              options: Array.from(new Set(logs.map((log) => log.action))),
              valueOf: (log) => log.action,
            },
          ]}
          columns={[
            {
              key: "action",
              label: "Action",
              render: (log) => (
                <>
                  <Typography sx={{ fontWeight: 800 }}>{log.action}</Typography>
                  <Typography
                    variant="caption"
                    sx={{ color: "text.secondary" }}
                  >
                    {log.requestId ?? "No request id"}
                  </Typography>
                </>
              ),
            },
            {
              key: "actor",
              label: "Actor",
              render: (log) => (
                <>
                  <Typography variant="body2">
                    {log.actorRole ?? "system"}
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{ color: "text.secondary" }}
                  >
                    {log.actorUserId ?? "anonymous"}
                  </Typography>
                </>
              ),
            },
            {
              key: "target",
              label: "Target",
              render: (log) => auditTargetSummary(log),
            },
            {
              key: "snapshot",
              label: "Snapshot",
              render: (log) => auditSnapshotSummary(log),
            },
            {
              key: "time",
              label: "Time",
              render: (log) => formatDateTime(log.createdAt),
            },
          ]}
          emptyState={
            <EmptyState
              title="No audit logs"
              detail="No sensitive action records match the current filters."
            />
          }
        />
      ) : (
        <EmptyState
          title="No audit logs"
          detail="No sensitive action records are available for this view."
        />
      )}
    </Paper>
  );
}
