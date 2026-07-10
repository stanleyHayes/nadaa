import type { ReactNode } from "react";
import {
  Alert,
  Box,
  Button,
  LinearProgress,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { RefreshCw, ShieldCheck } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { AdminLoadState, AdminMetric } from "../types";
import { toneColors } from "../utils";

export function StatusLine({
  loadState,
  message,
  onRefresh,
}: {
  loadState: AdminLoadState;
  message: string;
  onRefresh: () => void;
}) {
  const severity =
    loadState === "ready"
      ? "success"
      : loadState === "loading"
        ? "info"
        : "warning";

  return (
    <>
      <Alert
        className="feed-alert"
        severity={severity}
        action={
          <Button
            color="inherit"
            size="small"
            startIcon={<RefreshCw size={16} />}
            onClick={onRefresh}
          >
            Refresh
          </Button>
        }
      >
        {message}
      </Alert>
      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}
    </>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <Alert severity="error" className="feed-alert">
      {message}
    </Alert>
  );
}

export function EmptyState({
  detail,
  title,
}: {
  title: string;
  detail: string;
}) {
  return (
    <Box className="empty-state">
      <Stack alignItems="center" spacing={1}>
        <ShieldCheck size={30} color="var(--nadaa-slate)" />
        <Typography variant="subtitle1" fontWeight={800}>
          {title}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {detail}
        </Typography>
      </Stack>
    </Box>
  );
}

export function SectionHeader({
  action,
  eyebrow,
  icon,
  title,
}: {
  title: string;
  eyebrow: string;
  icon: ReactNode;
  action?: ReactNode;
}) {
  return (
    <Stack
      className="section-heading"
      direction={{ xs: "column", md: "row" }}
      justifyContent="space-between"
      gap={1}
    >
      <Stack direction="row" spacing={1.25} alignItems="center">
        {icon}
        <Box>
          <Typography variant="caption" color="text.secondary" fontWeight={800}>
            {eyebrow}
          </Typography>
          <Typography variant="h6">{title}</Typography>
        </Box>
      </Stack>
      {action}
    </Stack>
  );
}

export function MetricCard({
  icon,
  metric,
}: {
  metric: AdminMetric;
  icon: ReactNode;
}) {
  return (
    <Paper className="metric-card">
      <Stack direction="row" spacing={2} alignItems="center">
        <Box className="metric-icon" sx={{ color: toneColors[metric.tone] }}>
          {icon}
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary" fontWeight={800}>
            {metric.label}
          </Typography>
          <Typography variant="h4">{metric.value}</Typography>
          <Typography variant="body2" color="text.secondary">
            {metric.detail}
          </Typography>
        </Box>
      </Stack>
    </Paper>
  );
}
