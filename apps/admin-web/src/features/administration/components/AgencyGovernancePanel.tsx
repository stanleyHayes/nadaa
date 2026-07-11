import {
  Chip,
  LinearProgress,
  Paper,
  Stack,
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
import { DataTable } from "./DataTable";
import { EmptyState, SectionHeader } from "./shared";

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
        icon={<UsersRound size={22} color="var(--nadaa-navy)" />}
      />
      <DataTable
        rows={agencies}
        getRowKey={(agency) => agency.id}
        searchOf={(agency) =>
          `${agency.name} ${agency.region} ${agency.district}`
        }
        searchPlaceholder="Search agencies"
        filters={[
          {
            key: "status",
            label: "Status",
            options: Array.from(
              new Set(agencies.map((agency) => agency.status)),
            ),
            valueOf: (agency) => agency.status,
          },
          {
            key: "type",
            label: "Type",
            options: Array.from(
              new Set(agencies.map((agency) => agencyTypeLabel(agency.type))),
            ),
            valueOf: (agency) => agencyTypeLabel(agency.type),
          },
        ]}
        columns={[
          {
            key: "agency",
            label: "Agency",
            render: (agency) => (
              <Stack spacing={0.75}>
                <Typography sx={{
                  fontWeight: 800
                }}>{agency.name}</Typography>
                <Stack direction="row" spacing={1} sx={{
                  flexWrap: "wrap"
                }}>
                  <Chip
                    className="status-chip"
                    size="small"
                    color={statusColor(agency.status) as ChipColor}
                    label={agency.status}
                  />
                  <Chip size="small" label={agencyTypeLabel(agency.type)} />
                </Stack>
                <Typography variant="caption" sx={{
                  color: "text.secondary"
                }}>
                  {agency.region} / {agency.district}
                </Typography>
              </Stack>
            ),
          },
          {
            key: "scope",
            label: "Scope",
            render: (agency) => agency.dataScope,
          },
          {
            key: "users",
            label: "Users",
            render: (agency) => (
              <>
                <Typography sx={{
                  fontWeight: 800
                }}>{agency.users}</Typography>
                <Typography variant="caption" sx={{
                  color: "text.secondary"
                }}>
                  {agency.openAssignments} open assignments
                </Typography>
              </>
            ),
          },
          {
            key: "mfa",
            label: "MFA",
            render: (agency) => (
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
            ),
          },
          {
            key: "lastAudit",
            label: "Last audit",
            render: (agency) => formatDateTime(agency.lastAuditAt),
          },
        ]}
        emptyState={
          <EmptyState
            title="No agencies"
            detail="No agencies match the current search and filters. Adjust or clear them to see more."
          />
        }
      />
    </Paper>
  );
}
