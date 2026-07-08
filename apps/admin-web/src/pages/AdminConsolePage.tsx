import {
  type ChangeEvent,
  type ReactElement,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  AppBar,
  Box,
  Button,
  Chip,
  Container,
  CssBaseline,
  Grid,
  Paper,
  Stack,
  Tab,
  Tabs,
  ThemeProvider,
  Toolbar,
  Typography,
} from "@mui/material";
import {
  BellRing,
  Building2,
  DatabaseZap,
  KeyRound,
  LockKeyhole,
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
  UsersRound,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AuditLogRecord,
  CreateAgencyUserResponse,
} from "@nadaa/shared-types";
import {
  fetchAlertRules,
  fetchAuditLogs,
  fetchDataSources,
} from "../api/admin";
import { AUTH_API_BASE } from "../api/config";
import { adminHeaders, adminSession } from "../auth/session";
import {
  AgencyGovernancePanel,
  AlertRulePanel,
  AuditLogPanel,
  DataSourcePanel,
  EmptyState,
  MetricCard,
  MfaSupportPanel,
  RoleMatrixPanel,
  SectionHeader,
  StatusLine,
  UserManagementPanel,
} from "../components";
import {
  defaultUserForm,
  fallbackAgencies,
  fallbackAlertRules,
  fallbackAuditLogs,
  fallbackDataSources,
  fallbackUsers,
} from "../data/fixtures";
import type {
  AdminActionResult,
  AdminLoadState,
  AdminUserFormState,
  AdminView,
  AlertRuleSummary,
  DataSourceSummary,
  ManagedAgency,
  ManagedAgencyUser,
} from "../data/types";
import { adminTheme } from "../lib/theme";
import {
  buildAdminMetrics,
  managedUserFromCreateResponse,
  roleLabel,
  validateUserForm,
} from "../lib/utils";
import { viewTabs } from "../nav";
import { hasAdminAccess } from "../rbac";

function AdminConsolePage() {
  const hasAccess = hasAdminAccess(adminSession.role, adminSession.mfaEnabled);
  const [view, setView] = useState<AdminView>("overview");
  const [loadState, setLoadState] = useState<AdminLoadState>("loading");
  const [loadMessage, setLoadMessage] = useState("Loading governance data");
  const [agencies] = useState<ManagedAgency[]>(fallbackAgencies);
  const [users, setUsers] = useState<ManagedAgencyUser[]>(fallbackUsers);
  const [auditLogs, setAuditLogs] =
    useState<AuditLogRecord[]>(fallbackAuditLogs);
  const [dataSources, setDataSources] =
    useState<DataSourceSummary[]>(fallbackDataSources);
  const [alertRules, setAlertRules] =
    useState<AlertRuleSummary[]>(fallbackAlertRules);
  const [userForm, setUserForm] = useState<AdminUserFormState>(defaultUserForm);
  const [createBusy, setCreateBusy] = useState(false);
  const [actionResult, setActionResult] = useState<AdminActionResult>();

  const refreshAdminData = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setLoadMessage("Loading governance data");

    const [auditResult, sourceResult, alertResult] = await Promise.allSettled([
      fetchAuditLogs(signal),
      fetchDataSources(signal),
      fetchAlertRules(signal),
    ]);

    if (signal?.aborted) {
      return;
    }

    let fallbackCount = 0;
    if (auditResult.status === "fulfilled") {
      setAuditLogs(auditResult.value);
    } else {
      fallbackCount += 1;
      setAuditLogs(fallbackAuditLogs);
    }

    if (sourceResult.status === "fulfilled") {
      setDataSources(sourceResult.value);
    } else {
      fallbackCount += 1;
      setDataSources(fallbackDataSources);
    }

    if (alertResult.status === "fulfilled") {
      setAlertRules(alertResult.value);
    } else {
      fallbackCount += 1;
      setAlertRules(fallbackAlertRules);
    }

    if (fallbackCount === 0) {
      setLoadState("ready");
      setLoadMessage("Governance APIs connected.");
      return;
    }

    setLoadState("fallback");
    setLoadMessage(
      `${fallbackCount} governance API surface${fallbackCount === 1 ? "" : "s"} unavailable. Showing safe admin fixture data.`,
    );
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshAdminData(controller.signal);
    return () => controller.abort();
  }, []);

  const metrics = useMemo(
    () => buildAdminMetrics(agencies, users, auditLogs, dataSources),
    [agencies, auditLogs, dataSources, users],
  );

  const handleFormChange = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = event.target;
    setUserForm((current) => ({ ...current, [name]: value }));
  };

  const handleCreateUser = async () => {
    const validationMessage = validateUserForm(userForm);
    if (validationMessage) {
      setActionResult({ severity: "warning", message: validationMessage });
      return;
    }

    setCreateBusy(true);
    setActionResult(undefined);
    try {
      const response = await fetch(`${AUTH_API_BASE}/auth/agency-users`, {
        method: "POST",
        headers: adminHeaders(),
        body: JSON.stringify(userForm),
      });
      if (!response.ok) {
        throw new Error(`auth API returned ${response.status}`);
      }

      const payload = (await response.json()) as CreateAgencyUserResponse;
      setUsers((current) => [
        managedUserFromCreateResponse(payload),
        ...current,
      ]);
      setUserForm(defaultUserForm);
      setActionResult({
        severity: "success",
        message:
          "Authority user created. MFA setup is required before the user can sign in.",
      });
    } catch {
      setActionResult({
        severity: "error",
        message:
          "User was not created. The auth API is unavailable or rejected the current admin session.",
      });
    } finally {
      setCreateBusy(false);
    }
  };

  const renderActiveView = () => {
    if (view === "overview") {
      return (
        <Grid className="main-grid" container spacing={2}>
          <Grid size={{ xs: 12, xl: 8 }}>
            <AgencyGovernancePanel agencies={agencies} />
          </Grid>
          <Grid size={{ xs: 12, xl: 4 }}>
            <Stack spacing={2}>
              <MfaSupportPanel users={users} />
              <RoleMatrixPanel />
            </Stack>
          </Grid>
        </Grid>
      );
    }

    if (view === "access") {
      return (
        <Stack className="section-stack" spacing={2}>
          <UserManagementPanel
            actionResult={actionResult}
            agencies={agencies}
            busy={createBusy}
            form={userForm}
            onFormChange={handleFormChange}
            onSelectChange={handleFormChange}
            onSubmit={handleCreateUser}
            users={users}
          />
          <RoleMatrixPanel />
        </Stack>
      );
    }

    if (view === "audit") {
      return (
        <Stack className="section-stack" spacing={2}>
          <AuditLogPanel logs={auditLogs} />
        </Stack>
      );
    }

    if (view === "integrations") {
      return (
        <Stack className="section-stack" spacing={2}>
          <Paper className="surface">
            <SectionHeader
              eyebrow="Integration governance"
              title="Data-source contracts and safe secret scopes"
              icon={<DatabaseZap size={22} color={nadaaBrand.colors.green} />}
            />
            {dataSources.length ? (
              <DataSourcePanel dataSources={dataSources} />
            ) : (
              <EmptyState
                title="No data sources"
                detail="No integration contracts are currently visible to the admin console."
              />
            )}
          </Paper>
        </Stack>
      );
    }

    return (
      <Stack className="section-stack" spacing={2}>
        <Paper className="surface">
          <SectionHeader
            eyebrow="Alert governance"
            title="Approval, override, targeting, and audit rules"
            icon={<BellRing size={22} color={nadaaBrand.colors.red} />}
          />
          <AlertRulePanel rules={alertRules} />
        </Paper>
      </Stack>
    );
  };

  if (!hasAccess) {
    return (
      <ThemeProvider theme={adminTheme}>
        <CssBaseline />
        <a href="#main-content" className="skip-link">
          Skip to main content
        </a>
        <Box component="main" id="main-content" className="access-shell">
          <Paper className="access-panel">
            <ShieldAlert size={38} color={nadaaBrand.colors.red} />
            <Typography variant="h5">Admin access denied</Typography>
            <Typography color="text.secondary">
              The governance console requires a permitted admin role and
              completed MFA before any platform settings are visible.
            </Typography>
          </Paper>
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={adminTheme}>
      <CssBaseline />
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <AppBar className="topbar" position="static" elevation={0}>
        <Container maxWidth="xl">
          <Toolbar className="toolbar" disableGutters>
            <Stack
              className="brand-lockup"
              direction="row"
              spacing={1.5}
              alignItems="center"
            >
              <img
                src="/brand/nadaa-logo.png"
                alt="NADAA shield"
                className="brand-logo"
              />
              <Box>
                <Typography variant="h6">NADAA Admin Console</Typography>
                <Typography variant="caption">{nadaaBrand.slogan}</Typography>
              </Box>
            </Stack>
            <Stack className="topbar-actions" direction="row" spacing={1}>
              <Chip
                className="session-chip"
                label={`${roleLabel(adminSession.role)} / MFA complete`}
              />
              <Chip className="session-chip" label={adminSession.agency} />
            </Stack>
          </Toolbar>
        </Container>
      </AppBar>

      <Box component="main" id="main-content">
        <Container className="dashboard-shell" maxWidth="xl">
          <Stack
            className="page-heading"
            direction={{ xs: "column", lg: "row" }}
            justifyContent="space-between"
            gap={2}
          >
            <Box>
              <Typography variant="h4">Governance workspace</Typography>
              <Typography color="text.secondary">
                Agencies, authority users, MFA readiness, audit trace,
                integration contracts, and alert-rule controls.
              </Typography>
            </Box>
            <Button
              variant="outlined"
              startIcon={<RefreshCw size={18} />}
              onClick={() => void refreshAdminData()}
            >
              Refresh data
            </Button>
          </Stack>

          <StatusLine
            loadState={loadState}
            message={loadMessage}
            onRefresh={() => void refreshAdminData()}
          />

          <Grid container spacing={2}>
            {metrics.map((metric, index) => (
              <Grid key={metric.label} size={{ xs: 12, sm: 6, xl: 3 }}>
                <MetricCard
                  metric={metric}
                  icon={
                    index === 0 ? (
                      <Building2 size={25} />
                    ) : index === 1 ? (
                      <UsersRound size={25} />
                    ) : index === 2 ? (
                      <ShieldCheck size={25} />
                    ) : (
                      <DatabaseZap size={25} />
                    )
                  }
                />
              </Grid>
            ))}
          </Grid>

          <Paper className="surface view-tabs">
            <Tabs
              value={view}
              onChange={(_, nextView) => setView(nextView as AdminView)}
              variant="scrollable"
              scrollButtons="auto"
              aria-label="Admin console views"
            >
              {viewTabs.map((tab) => (
                <Tab
                  key={tab.id}
                  value={tab.id}
                  icon={tab.icon}
                  iconPosition="start"
                  label={tab.label}
                />
              ))}
            </Tabs>
          </Paper>

          {renderActiveView()}
        </Container>
      </Box>
    </ThemeProvider>
  );
}

export default AdminConsolePage;
