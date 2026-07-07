import { Chip, Grid, Paper, Stack, Typography } from "@mui/material";
import type { AlertRuleSummary } from "../../data/types";
import { formatDateTime, roleLabel, statusColor } from "../../lib/utils";

type ChipColor =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

export function AlertRulePanel({ rules }: { rules: AlertRuleSummary[] }) {
  return (
    <Grid container spacing={2}>
      {rules.map((rule) => (
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
                <Chip size="small" color="warning" label={rule.severity} />
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
      ))}
    </Grid>
  );
}
