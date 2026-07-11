import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";

import {
  AlertOctagon,
  AlertTriangle,
  CheckCircle2,
  Inbox,
  Info,
  MapPin,
  Users,
} from "lucide-react";
import { hazardRoles, severityRoles } from "@nadaa/brand";
import type { IncidentRecord } from "@nadaa/shared-types";
import { hazardLabel, severityLabel } from "../data";
import { mapHazardRole, mapSeverityRole } from "../utils";

export function LoadingState({ message }: { message?: string }) {
  return (
    <Box
      sx={{
        alignItems: "center",
        display: "flex",
        justifyContent: "center",
        minHeight: 200
      }}>
      <Stack spacing={2} sx={{
        alignItems: "center"
      }}>
        <CircularProgress />
        <Typography sx={{
          color: "text.secondary"
        }}>{message ?? "Loading"}</Typography>
      </Stack>
    </Box>
  );
}

export function EmptyState({
  message,
  title = "Nothing to show yet",
}: {
  message: string;
  title?: string;
}) {
  return (
    <Stack className="empty-state" spacing={1} sx={{ alignItems: "center" }}>
      <span aria-hidden="true" className="empty-state__icon">
        <Inbox size={28} strokeWidth={1.75} />
      </span>
      <Typography variant="subtitle2" sx={{ fontWeight: 800 }}>
        {title}
      </Typography>
      <Typography variant="body2" sx={{ color: "text.secondary" }}>
        {message}
      </Typography>
    </Stack>
  );
}

export function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <Alert
      action={
        onRetry ? (
          <Button color="inherit" onClick={onRetry} size="small">
            Retry
          </Button>
        ) : null
      }
      severity="error"
      sx={{ mt: 2 }}
    >
      {message}
    </Alert>
  );
}

export function MetricCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Users;
  label: string;
  value: number | string;
}) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack direction="row" spacing={2} sx={{
          alignItems: "center"
        }}>
          <Box sx={{
            color: "primary.main"
          }}>
            <Icon size={24} />
          </Box>
          <Box>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              {label}
            </Typography>
            <Typography variant="h5" sx={{
              fontWeight: 800
            }}>
              {value}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

export const severityIcons = {
  low: CheckCircle2,
  medium: AlertTriangle,
  high: AlertTriangle,
  severe: AlertOctagon,
  info: Info,
} as const;

export function SeverityChip({
  severity,
  size = "small",
}: {
  severity: IncidentRecord["severity"];
  size?: "small" | "medium";
}) {
  const roleKey = mapSeverityRole(severity);
  const role = severityRoles[roleKey];
  const Icon = severityIcons[roleKey];
  return (
    <Chip
      icon={<Icon size={size === "small" ? 14 : 16} />}
      label={severityLabel(severity)}
      size={size}
      sx={{
        backgroundColor: role.background,
        border: `1px solid ${role.border}`,
        color: role.foreground,
        fontWeight: 700,
        minWidth: 78,
        ".MuiChip-icon": {
          color: role.foreground,
        },
      }}
    />
  );
}

export function HazardChip({
  hazard,
  size = "small",
}: {
  hazard: IncidentRecord["type"];
  size?: "small" | "medium";
}) {
  const roleKey = mapHazardRole(hazard);
  const role = hazardRoles[roleKey];
  return (
    <Chip
      icon={<MapPin size={size === "small" ? 14 : 16} />}
      label={hazardLabel(hazard)}
      size={size}
      sx={{
        backgroundColor: role.background,
        border: `1px solid ${role.border}`,
        color: role.foreground,
        fontWeight: 600,
        ".MuiChip-icon": {
          color: role.foreground,
        },
      }}
      variant="outlined"
    />
  );
}
