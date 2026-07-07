import {
  Box,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { ShieldCheck } from "lucide-react";
import type { AuditLogRecord } from "@nadaa/shared-types";
import {
  auditSnapshotSummary,
  auditTargetSummary,
  formatDateTime,
} from "../../lib/utils";
import { EmptyState, SectionHeader } from "../shared";

export function AuditLogPanel({ logs }: { logs: AuditLogRecord[] }) {
  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Audit trail"
        title="Sensitive action trace"
        icon={<ShieldCheck size={22} color="#0D1B3D" />}
      />
      {logs.length ? (
        <Box className="admin-table">
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Action</TableCell>
                <TableCell>Actor</TableCell>
                <TableCell>Target</TableCell>
                <TableCell>Snapshot</TableCell>
                <TableCell>Time</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {logs.map((log) => (
                <TableRow className="audit-row" key={log.id}>
                  <TableCell>
                    <Typography fontWeight={800}>{log.action}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {log.requestId ?? "No request id"}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2">
                      {log.actorRole ?? "system"}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {log.actorUserId ?? "anonymous"}
                    </Typography>
                  </TableCell>
                  <TableCell>{auditTargetSummary(log)}</TableCell>
                  <TableCell>{auditSnapshotSummary(log)}</TableCell>
                  <TableCell>{formatDateTime(log.createdAt)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Box>
      ) : (
        <EmptyState
          title="No audit logs"
          detail="No sensitive action records are available for this view."
        />
      )}
    </Paper>
  );
}
