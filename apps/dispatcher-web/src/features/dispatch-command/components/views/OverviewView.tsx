import { Alert, Button, Grid, Stack, Typography } from "@mui/material";
import {
  ArrowRight,
  CloudRain,
  HeartPulse,
  RadioTower,
  ShieldAlert,
  Siren,
  Truck,
} from "lucide-react";
import type { DispatcherSession } from "@/app/session";
import type { DispatchData } from "../../useDispatchData";
import type { ViewId } from "../../navigation";
import {
  CapacityMeter,
  Eyebrow,
  MetricTile,
  SectionCard,
  TriageLadder,
} from "../primitives";
import { StatusLine } from "../shared";
import { DonutChart, ProgressRing } from "../charts";

function utilizationTone(pct: number): "green" | "gold" | "red" {
  if (pct >= 90) {
    return "red";
  }
  if (pct >= 70) {
    return "gold";
  }
  return "green";
}

export function OverviewView({
  data,
  session,
  onNavigate,
}: {
  data: DispatchData;
  session: DispatcherSession;
  onNavigate: (view: ViewId) => void;
}) {
  const {
    incidents,
    alerts,
    hospitalFacilities,
    hospitalLoadState,
    mlPredictions,
    reliefPoints,
    loadState,
    loadMessage,
  } = data;

  const active = incidents.filter(
    (incident) =>
      incident.status !== "closed" && incident.status !== "false_report",
  );
  const newReports = incidents.filter(
    (incident) =>
      incident.status === "reported" || incident.status === "under_review",
  ).length;
  const enRoute = incidents.filter(
    (incident) =>
      incident.status === "response_en_route" ||
      incident.status === "on_scene",
  ).length;
  const priorityReview = incidents.filter(
    (incident) => incident.priorityReview,
  ).length;
  const resolvedClosed = incidents.filter(
    (incident) =>
      incident.status === "contained" ||
      incident.status === "recovery_ongoing" ||
      incident.status === "closed",
  ).length;

  const pendingAlerts = alerts.filter(
    (alert) => alert.status === "draft" || alert.status === "submitted",
  ).length;
  const approvedAlerts = alerts.filter(
    (alert) => alert.status === "approved",
  ).length;
  const liveAlerts = alerts.filter(
    (alert) => alert.status === "published",
  ).length;

  const availableBeds = hospitalFacilities.reduce(
    (sum, facility) => sum + facility.availableBeds,
    0,
  );
  const totalBeds = hospitalFacilities.reduce(
    (sum, facility) => sum + facility.totalBeds,
    0,
  );
  const occupiedBeds = totalBeds - availableBeds;
  const bedUtilization =
    totalBeds > 0 ? Math.round((occupiedBeds / totalBeds) * 100) : 0;

  const responseStatusData = [
    { label: "New / review", value: newReports, color: "var(--nadaa-gold)" },
    { label: "En route / on scene", value: enRoute, color: "var(--nadaa-navy)" },
    {
      label: "Resolved / closed",
      value: resolvedClosed,
      color: "var(--nadaa-green)",
    },
  ];

  const mlNeedsReview = mlPredictions.filter(
    (prediction) => prediction.reviewStatus === "needs_review",
  ).length;
  const mlReviewed = mlPredictions.length - mlNeedsReview;

  const topFacilities = [...hospitalFacilities]
    .map((facility) => ({
      facility,
      pct:
        facility.totalBeds > 0
          ? Math.round(
              ((facility.totalBeds - facility.availableBeds) /
                facility.totalBeds) *
                100,
            )
          : 0,
    }))
    .sort((a, b) => b.pct - a.pct)
    .slice(0, 4);

  const pipeline = [
    { label: "Draft", value: alerts.filter((a) => a.status === "draft").length },
    {
      label: "Submitted",
      value: alerts.filter((a) => a.status === "submitted").length,
    },
    { label: "Approved", value: approvedAlerts },
    { label: "Live", value: liveAlerts },
  ];

  const feedLabel =
    loadState === "ready"
      ? "Live"
      : loadState === "empty"
        ? "Idle"
        : loadState === "loading"
          ? "Loading"
          : "Offline";

  return (
    <Stack spacing={2.5} className="cc-overview">
      {loadState === "fallback" ||
      loadState === "error" ||
      loadState === "empty" ? (
        <Alert
          severity={loadState === "empty" ? "info" : "warning"}
          className="feed-alert"
        >
          {loadMessage}
        </Alert>
      ) : null}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Active incidents"
            value={active.length}
            caption={`${newReports} new · ${enRoute} en route`}
            icon={Siren}
            accent="red"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Pending alerts"
            value={pendingAlerts}
            caption={`${approvedAlerts} approved · ${liveAlerts} live`}
            icon={RadioTower}
            accent="gold"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Hospital beds"
            value={availableBeds.toLocaleString()}
            caption={`available of ${totalBeds.toLocaleString()} across ${hospitalFacilities.length} sites`}
            icon={HeartPulse}
            accent="green"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="ML flood signals"
            value={mlNeedsReview}
            caption={`${mlReviewed} reviewed · ${priorityReview} priority incidents`}
            icon={CloudRain}
            accent="info"
          />
        </Grid>
      </Grid>
      <SectionCard
        title="Live triage"
        eyebrow="Active incidents by severity"
        icon={ShieldAlert}
        accent="navy"
        className="cc-triage-card"
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
              onClick={() => onNavigate("incidents")}
            >
              Open queue
            </Button>
          </Stack>
        }
      >
        <TriageLadder incidents={incidents} />
        <p className="cc-triage__foot">
          {active.length} active of {incidents.length} tracked incidents across
          Greater Accra.
        </p>
      </SectionCard>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Alert pipeline"
            eyebrow="Broadcast approval flow"
            icon={RadioTower}
            accent="gold"
            action={
              <Button
                size="small"
                variant="outlined"
                endIcon={<ArrowRight size={15} />}
                onClick={() => onNavigate("alerts")}
              >
                Review alerts
              </Button>
            }
          >
            <div className="cc-pipeline">
              {pipeline.map((stage, index) => (
                <div className="cc-pipeline__stage" key={stage.label}>
                  <span className="cc-pipeline__value">{stage.value}</span>
                  <span className="cc-pipeline__label">{stage.label}</span>
                  {index < pipeline.length - 1 ? (
                    <ArrowRight
                      size={16}
                      className="cc-pipeline__arrow"
                      aria-hidden
                    />
                  ) : null}
                </div>
              ))}
            </div>
            <p className="cc-muted-note">
              {pendingAlerts > 0
                ? `${pendingAlerts} alert${pendingAlerts === 1 ? "" : "s"} waiting on human approval.`
                : "No alerts are waiting on approval."}
            </p>
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Hospital capacity"
            eyebrow="Highest occupancy first"
            icon={HeartPulse}
            accent="green"
            action={
              <Button
                size="small"
                variant="outlined"
                endIcon={<ArrowRight size={15} />}
                onClick={() => onNavigate("capacity")}
              >
                View capacity
              </Button>
            }
          >
            <Stack spacing={1.5}>
              {topFacilities.map(({ facility, pct }) => (
                <div className="cc-shelter-row" key={facility.id}>
                  <div className="cc-shelter-row__head">
                    <span className="cc-shelter-row__name">{facility.name}</span>
                    <span className="cc-shelter-row__figure">
                      {facility.availableBeds}/{facility.totalBeds} free · {pct}%
                      full
                    </span>
                  </div>
                  <CapacityMeter
                    value={facility.totalBeds - facility.availableBeds}
                    max={facility.totalBeds}
                    tone={utilizationTone(pct)}
                  />
                </div>
              ))}
              {topFacilities.length === 0 ? (
                <p className="cc-muted-note">No hospital capacity is loaded yet.</p>
              ) : null}
            </Stack>
          </SectionCard>
        </Grid>
      </Grid>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Response status"
            eyebrow="Incidents by stage"
            icon={Truck}
            accent="red"
          >
            <DonutChart
              data={responseStatusData}
              centerValue={active.length}
              centerLabel="active"
            />
          </SectionCard>
        </Grid>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Hospital occupancy"
            eyebrow="Bed utilisation"
            icon={HeartPulse}
            accent="green"
          >
            <Stack
              direction="row"
              spacing={2.5}
              sx={{ alignItems: "center", flexWrap: "wrap" }}
            >
              <ProgressRing
                value={bedUtilization}
                color={`var(--nadaa-${utilizationTone(bedUtilization)})`}
                label="occupied"
              />
              <Stack spacing={0.5}>
                <Typography variant="h5" sx={{ fontWeight: 800 }}>
                  {occupiedBeds.toLocaleString()} / {totalBeds.toLocaleString()}
                </Typography>
                <Typography variant="body2" sx={{ color: "text.secondary" }}>
                  beds occupied across {hospitalFacilities.length} sites
                </Typography>
                <Typography variant="body2" sx={{ color: "text.secondary" }}>
                  {availableBeds.toLocaleString()} beds still available
                </Typography>
              </Stack>
            </Stack>
          </SectionCard>
        </Grid>
      </Grid>
      <SectionCard title="Operating posture" eyebrow="Desk status" accent="navy">
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, md: 6 }}>
            <Stack spacing={1.25}>
              <StatusLine
                label="Incident feed"
                value={loadState === "ready" ? "Live" : "Offline"}
                color={loadState === "ready" ? "success" : "warning"}
              />
              <StatusLine
                label="Dispatcher session"
                value={session.agency}
                color="success"
              />
            </Stack>
          </Grid>
          <Grid size={{ xs: 12, md: 6 }}>
            <Stack spacing={1.25}>
              <StatusLine
                label="Hospital capacity"
                value={hospitalLoadState === "ready" ? "Live" : "Offline"}
                color={hospitalLoadState === "ready" ? "success" : "warning"}
              />
              <StatusLine
                label="Relief points"
                value={`${reliefPoints.length} mapped`}
                color={reliefPoints.length ? "success" : "warning"}
              />
            </Stack>
          </Grid>
        </Grid>
        <Stack direction="row" spacing={1} className="cc-overview__jump">
          <Eyebrow>Jump to</Eyebrow>
          <Button
            size="small"
            variant="text"
            onClick={() => onNavigate("incidents")}
          >
            Incidents
          </Button>
          <Button size="small" variant="text" onClick={() => onNavigate("triage")}>
            AI triage
          </Button>
          <Button
            size="small"
            variant="text"
            onClick={() => onNavigate("ml-review")}
          >
            ML review
          </Button>
        </Stack>
      </SectionCard>
    </Stack>
  );
}
