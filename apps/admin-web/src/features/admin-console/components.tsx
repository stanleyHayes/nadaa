import { type ChangeEvent, type ReactNode } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  LinearProgress,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import {
  CheckCircle2,
  DatabaseZap,
  KeyRound,
  LockKeyhole,
  RefreshCw,
  ShieldCheck,
  UserPlus,
  UsersRound,
} from "lucide-react";
import type { AuditLogRecord } from "@nadaa/shared-types";
import type {
  AdminActionResult,
  AdminLoadState,
  AdminMetric,
  AdminUserFormState,
  AlertRuleSummary,
  DataSourceSummary,
  ManagedAgency,
  ManagedAgencyUser,
} from "./types";
import {
  agencyTypeLabel,
  auditSnapshotSummary,
  auditTargetSummary,
  formatDateTime,
  formatPercent,
  roleLabel,
  roleOptions,
  statusColor,
  toneColors,
} from "./utils";

type ChipColor =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

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
        <ShieldCheck size={30} color="#555B66" />
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
        icon={<UsersRound size={22} color="#0D1B3D" />}
      />
      <Box className="admin-table">
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

export function UserManagementPanel({
  actionResult,
  agencies,
  busy,
  form,
  onFormChange,
  onSelectChange,
  onSubmit,
  users,
}: {
  agencies: ManagedAgency[];
  users: ManagedAgencyUser[];
  form: AdminUserFormState;
  busy: boolean;
  actionResult?: AdminActionResult;
  onFormChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSelectChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSubmit: () => void;
}) {
  return (
    <Grid container spacing={2}>
      <Grid size={{ xs: 12, lg: 8 }}>
        <Paper className="surface">
          <SectionHeader
            eyebrow="Authority access"
            title="Users, roles, and MFA state"
            icon={<KeyRound size={22} color="#0D1B3D" />}
          />
          <Box className="admin-table">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>User</TableCell>
                  <TableCell>Role</TableCell>
                  <TableCell>Agency</TableCell>
                  <TableCell>MFA</TableCell>
                  <TableCell>Last login</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <Typography fontWeight={800}>{user.name}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {user.email}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        className="role-chip"
                        size="small"
                        color={
                          user.role === "system_admin" ? "primary" : "default"
                        }
                        label={roleLabel(user.role)}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {user.agency.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {user.accessScope}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={user.mfaEnabled ? "success" : "warning"}
                        label={user.mfaEnabled ? "Enabled" : "Setup pending"}
                      />
                    </TableCell>
                    <TableCell>{formatDateTime(user.lastLoginAt)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        </Paper>
      </Grid>
      <Grid size={{ xs: 12, lg: 4 }}>
        <Paper className="user-form">
          <SectionHeader
            eyebrow="Create user"
            title="Provision authority access"
            icon={<UserPlus size={22} color="#118D4E" />}
          />
          <Stack spacing={1.5}>
            {actionResult ? (
              <Alert severity={actionResult.severity}>
                {actionResult.message}
              </Alert>
            ) : null}
            <TextField
              name="name"
              label="Full name"
              size="small"
              value={form.name}
              onChange={onFormChange}
            />
            <TextField
              name="email"
              label="Email"
              size="small"
              value={form.email}
              onChange={onFormChange}
            />
            <TextField
              name="phone"
              label="Phone"
              size="small"
              value={form.phone}
              onChange={onFormChange}
            />
            <TextField
              select
              name="agencyId"
              label="Agency"
              size="small"
              value={form.agencyId}
              onChange={onSelectChange}
            >
              {agencies.map((agency) => (
                <MenuItem key={agency.id} value={agency.id}>
                  {agency.name}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              select
              name="role"
              label="Role"
              size="small"
              value={form.role}
              onChange={onSelectChange}
            >
              {roleOptions.map((role) => (
                <MenuItem key={role} value={role}>
                  {roleLabel(role)}
                </MenuItem>
              ))}
            </TextField>
            <Button
              variant="contained"
              startIcon={<UserPlus size={18} />}
              disabled={busy}
              onClick={onSubmit}
            >
              {busy ? "Creating" : "Create user"}
            </Button>
          </Stack>
        </Paper>
      </Grid>
    </Grid>
  );
}

export function RoleMatrixPanel() {
  const adminRoles = new Set(["system_admin", "agency_admin"]);
  const alertApprovalRoles = new Set([
    "system_admin",
    "agency_admin",
    "nadmo_officer",
  ]);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Role policy"
        title="Admin, alert, and operational permissions"
        icon={<LockKeyhole size={22} color="#0D1B3D" />}
      />
      <Box className="admin-table">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Role</TableCell>
              <TableCell>Admin console</TableCell>
              <TableCell>Alert approval</TableCell>
              <TableCell>Dispatcher operations</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {roleOptions.map((role) => (
              <TableRow key={role}>
                <TableCell>{roleLabel(role)}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={adminRoles.has(role) ? "success" : "default"}
                    label={adminRoles.has(role) ? "Permitted" : "Denied"}
                  />
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={alertApprovalRoles.has(role) ? "warning" : "default"}
                    label={alertApprovalRoles.has(role) ? "Reviewer" : "No"}
                  />
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={
                      role === "dispatcher" || role === "responder"
                        ? "primary"
                        : "default"
                    }
                    label={
                      role === "dispatcher" || role === "responder"
                        ? "Operational"
                        : "Scoped"
                    }
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Paper>
  );
}

export function MfaSupportPanel({ users }: { users: ManagedAgencyUser[] }) {
  const pendingUsers = users.filter((user) => !user.mfaEnabled);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="MFA support"
        title="Authority users waiting on setup"
        icon={<CheckCircle2 size={22} color="#118D4E" />}
      />
      {pendingUsers.length ? (
        <Stack spacing={1.5}>
          {pendingUsers.map((user) => (
            <Stack
              key={user.id}
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              gap={1}
            >
              <Box>
                <Typography fontWeight={800}>{user.name}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {user.agency.name} / {roleLabel(user.role)}
                </Typography>
              </Box>
              <Chip color="warning" label="Setup pending" />
            </Stack>
          ))}
        </Stack>
      ) : (
        <EmptyState
          title="MFA coverage complete"
          detail="Every authority user in the current admin view has completed MFA setup."
        />
      )}
    </Paper>
  );
}

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

export function DataSourcePanel({
  dataSources,
}: {
  dataSources: DataSourceSummary[];
}) {
  return (
    <Grid container spacing={2}>
      {dataSources.map((source) => (
        <Grid key={source.id} size={{ xs: 12, md: 6, xl: 4 }}>
          <Paper className="data-source-card">
            <Stack spacing={1.2}>
              <Stack direction="row" justifyContent="space-between" gap={1}>
                <DatabaseZap size={24} color="#118D4E" />
                <Chip
                  size="small"
                  color={statusColor(source.status) as ChipColor}
                  label={source.status}
                />
              </Stack>
              <Box>
                <Typography variant="h6">{source.partner}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {source.domain} / {source.direction}
                </Typography>
              </Box>
              <Typography variant="body2">{source.cadence}</Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                <Chip size="small" label={`PII: ${source.pii}`} />
                <Chip
                  size="small"
                  label={`Fresh ${source.freshnessWindowMinutes}m`}
                />
                <Chip
                  size="small"
                  label={`Auth: ${source.authenticationMode}`}
                />
              </Stack>
              <Typography variant="caption" color="text.secondary">
                Secret scope: {source.secretScope ?? "none configured"}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Fallback: {source.manualFallback}
              </Typography>
            </Stack>
          </Paper>
        </Grid>
      ))}
    </Grid>
  );
}

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
