import {
  Box,
  Chip,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { UsersRound } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { ManagedAgency } from "../types";
import {
  agencyTypeLabel,
  formatDateTime,
  formatPercent,
  statusColor,
} from "../utils";
import { SectionHeader } from "./shared";

type ChipColor =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

export function AgencyGovernancePanel({
  agencies,
}: {
  agencies: ManagedAgency[];
}) {
  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Agency governance"
        title="Registered agencies and operating scope"
        icon={<UsersRound size={22} color={nadaaBrand.colors.navy} />}
      />
      <Box
        className="admin-table"
        tabIndex={0}
        aria-label="Agency governance table, scroll horizontally on small screens"
      >
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Agency</TableCell>
              <TableCell>Scope</TableCell>
              <TableCell>Users</TableCell>
              <TableCell>MFA</TableCell>
              <TableCell>Last audit</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {agencies.map((agency) => (
              <TableRow className="agency-row" key={agency.id}>
                <TableCell>
                  <Stack spacing={0.75}>
                    <Typography fontWeight={800}>{agency.name}</Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap">
                      <Chip
                        className="status-chip"
                        size="small"
                        color={statusColor(agency.status) as ChipColor}
                        label={agency.status}
                      />
                      <Chip size="small" label={agencyTypeLabel(agency.type)} />
                    </Stack>
                    <Typography variant="caption" color="text.secondary">
                      {agency.region} / {agency.district}
                    </Typography>
                  </Stack>
                </TableCell>
                <TableCell>{agency.dataScope}</TableCell>
                <TableCell>
                  <Typography fontWeight={800}>{agency.users}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {agency.openAssignments} open assignments
                  </Typography>
                </TableCell>
                <TableCell>
                  <Stack spacing={0.75}>
                    <LinearProgress
                      className="mfa-meter"
                      variant="determinate"
                      color={agency.mfaCoverage >= 90 ? "success" : "warning"}
                      value={agency.mfaCoverage}
                      aria-label={`${agency.name} MFA coverage`}
                    />
                    <Typography variant="caption">
                      {formatPercent(agency.mfaCoverage)}
                    </Typography>
                  </Stack>
                </TableCell>
                <TableCell>{formatDateTime(agency.lastAuditAt)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Paper>
  );
}
