import { Alert, Button, Grid, Stack } from "@mui/material";
import {
  ArrowRight,
  Building2,
  DatabaseZap,
  RefreshCw,
  ScrollText,
  ShieldCheck,
  UsersRound,
} from "lucide-react";
import type { AdminSession } from "@/app/session";
import type { AdminData } from "../../useAdminData";
import type { ViewId } from "../../navigation";
import {
  auditTargetSummary,
  formatDateTime,
  formatPercent,
} from "../../utils";
import {
  CoverageMeter,
  Eyebrow,
  MetricTile,
  PostureRow,
  SectionCard,
} from "../primitives";
import { DonutChart, ProgressRing } from "../charts";

function coverageTone(pct: number): "green" | "gold" | "red" {
  if (pct >= 90) {
    return "green";
  }
  if (pct >= 70) {
    return "gold";
  }
  return "red";
}

export function OverviewView({
  data,
  session,
  onNavigate,
}: {
  data: AdminData;
  session: AdminSession;
  onNavigate: (view: ViewId) => void;
}) {
  const { agencies, users, auditLogs, dataSources, loadState, loadMessage } =
    data;

  const mfaReady = users.filter((user) => user.mfaEnabled).length;
  const awaitingMfa = users.length - mfaReady;
  const mfaCoverage =
    users.length > 0 ? Math.round((mfaReady / users.length) * 100) : 0;
  const activeSources = dataSources.filter(
    (source) => source.status === "pilot" || source.status === "production",
  ).length;

  const agencyActive = agencies.filter((a) => a.status === "active").length;
  const agencyPilot = agencies.filter((a) => a.status === "pilot").length;
  const agencyOther = agencies.length - agencyActive - agencyPilot;
  const agencyMix = [
    { label: "Active", value: agencyActive, color: "var(--nadaa-green)" },
    { label: "Pilot", value: agencyPilot, color: "var(--nadaa-gold)" },
    { label: "Other", value: agencyOther, color: "var(--nadaa-slate)" },
  ];

  const coverageBoard = [...agencies].sort(
    (a, b) => (a.mfaCoverage ?? -1) - (b.mfaCoverage ?? -1),
  );
  const recentAudit = auditLogs.slice(0, 5);

  const feedLabel =
    loadState === "ready" ? "Live" : loadState === "loading" ? "Loading" : "Offline";

  return (
    <Stack spacing={2.5} className="cc-overview">
      {loadState === "error" ? (
        <Alert
          severity="error"
          className="feed-alert"
          action={
            <Button
              color="inherit"
              size="small"
              startIcon={<RefreshCw size={16} />}
              onClick={data.refresh}
            >
              Refresh
            </Button>
          }
        >
          {loadMessage}
        </Alert>
      ) : null}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Agencies"
            value={agencies.length}
            caption={`${agencies.filter((a) => a.status === "active").length} active · ${agencies.filter((a) => a.status === "pilot").length} pilot`}
            icon={Building2}
            accent="navy"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="MFA coverage"
            value={mfaCoverage}
            suffix="%"
            caption={`${mfaReady} of ${users.length} authority users`}
            icon={ShieldCheck}
            accent="green"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Awaiting MFA"
            value={awaitingMfa}
            caption={
              awaitingMfa > 0 ? "Setup pending before sign-in" : "All users verified"
            }
            icon={UsersRound}
            accent={awaitingMfa > 0 ? "gold" : "green"}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Data sources"
            value={dataSources.length}
            caption={`${activeSources} pilot or production`}
            icon={DatabaseZap}
            accent={activeSources > 0 ? "green" : "info"}
          />
        </Grid>
      </Grid>
      <SectionCard
        title="MFA readiness by agency"
        eyebrow="Lowest coverage first"
        icon={ShieldCheck}
        accent="navy"
        action={
          <Stack direction="row" spacing={1} sx={{
            alignItems: "center"
          }}>
            <span className={`cc-feed-chip cc-feed-chip--${loadState}`}>
              {feedLabel}
            </span>
            <Button
              size="small"
              variant="outlined"
              endIcon={<ArrowRight size={15} />}
              onClick={() => onNavigate("mfa")}
            >
              MFA readiness
            </Button>
          </Stack>
        }
      >
        <Stack spacing={1.5}>
          {coverageBoard.map((agency) => (
            <div className="cc-shelter-row" key={agency.id}>
              <div className="cc-shelter-row__head">
                <span className="cc-shelter-row__name">{agency.name}</span>
                <span className="cc-shelter-row__figure">
                  {agency.mfaCoverage === null
                    ? "users unavailable"
                    : `${agency.users} users · ${formatPercent(agency.mfaCoverage)}`}
                </span>
              </div>
              <CoverageMeter
                value={agency.mfaCoverage ?? 0}
                tone={
                  agency.mfaCoverage === null
                    ? "red"
                    : coverageTone(agency.mfaCoverage)
                }
              />
            </div>
          ))}
          {coverageBoard.length === 0 ? (
            <p className="cc-muted-note">No agencies are registered yet.</p>
          ) : null}
        </Stack>
      </SectionCard>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Agency mix"
            eyebrow="By registration status"
            icon={Building2}
            accent="navy"
          >
            <DonutChart
              data={agencyMix}
              centerValue={agencies.length}
              centerLabel="agencies"
            />
          </SectionCard>
        </Grid>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="MFA coverage"
            eyebrow="Authority sign-in security"
            icon={ShieldCheck}
            accent="green"
          >
            <Stack
              direction="row"
              spacing={2.5}
              sx={{ alignItems: "center", flexWrap: "wrap" }}
            >
              <ProgressRing
                value={mfaCoverage}
                color="var(--nadaa-green)"
                label="enrolled"
              />
              <span className="cc-shelter-row__figure">
                {mfaReady} of {users.length} users
              </span>
            </Stack>
          </SectionCard>
        </Grid>
      </Grid>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <SectionCard
            title="Recent audit activity"
            eyebrow="Sensitive-action trace"
            icon={ScrollText}
            accent="gold"
            action={
              <Button
                size="small"
                variant="outlined"
                endIcon={<ArrowRight size={15} />}
                onClick={() => onNavigate("audit")}
              >
                Open audit trail
              </Button>
            }
          >
            <Stack spacing={1.25}>
              {recentAudit.map((log) => (
                <div className="cc-audit-row" key={log.id}>
                  <div className="cc-audit-row__main">
                    <span className="cc-audit-row__action">{log.action}</span>
                    <span className="cc-audit-row__meta">
                      {(log.actorRole ?? "system")} · {auditTargetSummary(log)}
                    </span>
                  </div>
                  <span className="cc-audit-row__time">
                    {formatDateTime(log.createdAt)}
                  </span>
                </div>
              ))}
              {recentAudit.length === 0 ? (
                <p className="cc-muted-note">No audit events are loaded yet.</p>
              ) : null}
            </Stack>
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12, lg: 5 }}>
          <SectionCard title="Governance posture" eyebrow="Desk status" accent="green">
            <Stack spacing={1.25}>
              <PostureRow
                label="Governance APIs"
                value={
                  loadState === "ready"
                    ? "Connected"
                    : loadState === "loading"
                      ? "Connecting"
                      : "Unavailable"
                }
                tone={
                  loadState === "ready"
                    ? "green"
                    : loadState === "loading"
                      ? "gold"
                      : "red"
                }
              />
              <PostureRow
                label="Admin session"
                value={session.agency}
                tone="green"
              />
              <PostureRow
                label="MFA coverage"
                value={`${mfaCoverage}%`}
                tone={coverageTone(mfaCoverage)}
              />
              <PostureRow
                label="Alerts"
                value="Human approval"
                tone="gold"
              />
            </Stack>
            <Stack direction="row" spacing={1} className="cc-overview__jump">
              <Eyebrow>Jump to</Eyebrow>
              <Button size="small" variant="text" onClick={() => onNavigate("users")}>
                Users
              </Button>
              <Button size="small" variant="text" onClick={() => onNavigate("roles")}>
                Roles
              </Button>
              <Button
                size="small"
                variant="text"
                onClick={() => onNavigate("integrations")}
              >
                Data sources
              </Button>
            </Stack>
          </SectionCard>
        </Grid>
      </Grid>
    </Stack>
  );
}
