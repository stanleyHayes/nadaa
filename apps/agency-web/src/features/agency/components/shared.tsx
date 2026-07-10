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
      alignItems="center"
      display="flex"
      justifyContent="center"
      minHeight={200}
    >
      <Stack alignItems="center" spacing={2}>
        <CircularProgress />
        <Typography color="text.secondary">{message ?? "Loading"}</Typography>
      </Stack>
    </Box>
  );
}

export function EmptyState({ message }: { message: string }) {
  return (
    <Alert severity="info" sx={{ mt: 2 }}>
      {message}
    </Alert>
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
        <Stack alignItems="center" direction="row" spacing={2}>
          <Box color="primary.main">
            <Icon size={24} />
          </Box>
          <Box>
            <Typography color="text.secondary" variant="caption">
              {label}
            </Typography>
            <Typography fontWeight={800} variant="h5">
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
