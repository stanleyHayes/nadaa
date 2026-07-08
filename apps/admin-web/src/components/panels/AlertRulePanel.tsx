import { Chip, Grid, Paper, Stack, Typography } from "@mui/material";
import { AlertOctagon, AlertTriangle, CheckCircle2, Info } from "lucide-react";
import type { AlertSeverity } from "@nadaa/shared-types";
import type { AlertRuleSummary } from "../../data/types";
import {
  alertSeverityRole,
  formatDateTime,
  roleLabel,
  statusColor,
} from "../../lib/utils";

type ChipColor =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

const severityIconMap: Record<AlertSeverity, typeof CheckCircle2> = {
  advisory: Info,
  watch: CheckCircle2,
  warning: AlertTriangle,
  severe_warning: AlertTriangle,
  emergency: AlertOctagon,
};

export function AlertRulePanel({ rules }: { rules: AlertRuleSummary[] }) {
  return (
    <Grid container spacing={2}>
      {rules.map((rule) => {
        const severityRole = alertSeverityRole(rule.severity);
        const SeverityIcon = severityIconMap[rule.severity];
        return (
          <Grid key={rule.id} size={{ xs: 12, md: 6, xl: 4 }}>
            <Paper className="rule-card">
              <Stack spacing={1.25}>
                <Stack direction="row" justifyContent="space-between" gap={1}>
                  <Typography variant="h6">{rule.name}</Typography>
                  <Chip
                    size="small"
                    color={statusColor(rule.status) as ChipColor}
                    label={rule.status}
                  />
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {rule.scope}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Chip size="small" label={rule.targetType} />
                  <Chip
                    size="small"
                    icon={<SeverityIcon size={16} aria-hidden="true" />}
                    label={rule.severity}
                    sx={{
                      backgroundColor: severityRole.background,
                      color: severityRole.foreground,
                      border: `1px solid ${severityRole.border}`,
                      fontWeight: 800,
                      "& .MuiChip-icon": { color: severityRole.foreground },
                    }}
                  />
                  <Chip
                    size="small"
                    color={rule.mfaRequired ? "success" : "error"}
                    label={rule.mfaRequired ? "MFA required" : "MFA missing"}
                  />
                </Stack>
                <Typography variant="caption" color="text.secondary">
                  Approvers: {rule.approverRoles.map(roleLabel).join(", ")}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Override:{" "}
                  {rule.emergencyOverrideRoles.map(roleLabel).join(", ")}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Audit: {rule.auditAction}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Reviewed {formatDateTime(rule.lastReviewedAt)}
                </Typography>
              </Stack>
            </Paper>
          </Grid>
        );
      })}
    </Grid>
  );
}
