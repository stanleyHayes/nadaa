import { Alert, Button, Grid, Stack } from "@mui/material";
import {
  Activity,
  ArrowRight,
  Building2,
  ClipboardList,
  HandHeart,
  PackageCheck,
  ShieldAlert,
  Truck,
} from "lucide-react";
import type { AgencyData } from "../../useAgencyData";
import type { ViewId } from "../../navigation";
import {
  CapacityMeter,
  Eyebrow,
  MetricTile,
  ResponseLadder,
  SectionCard,
  StatusLine,
} from "../primitives";

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
  onNavigate,
}: {
  data: AgencyData;
  onNavigate: (view: ViewId) => void;
}) {
  const {
    session,
    incidents,
    incidentLoadState,
    metrics,
    shelters,
    hospitals,
    reliefPoints,
    aidRequests,
  } = data;

  const totalCapacity = shelters.reduce(
    (sum, shelter) => sum + shelter.capacity,
    0,
  );
  const totalOccupancy = shelters.reduce(
    (sum, shelter) => sum + shelter.currentOccupancy,
    0,
  );
  const utilization =
    totalCapacity > 0 ? Math.round((totalOccupancy / totalCapacity) * 100) : 0;

  const topShelters = [...shelters]
    .map((shelter) => ({
      shelter,
      pct:
        shelter.capacity > 0
          ? Math.round((shelter.currentOccupancy / shelter.capacity) * 100)
          : 0,
    }))
    .sort((a, b) => b.pct - a.pct)
    .slice(0, 4);

  const hospitalBeds = hospitals.reduce(
    (sum, facility) => sum + facility.availableBeds,
    0,
  );

  const feedLabel =
    incidentLoadState === "ready"
      ? "Live"
      : incidentLoadState === "loading"
        ? "Loading"
        : incidentLoadState === "error"
          ? "Offline"
          : "Idle";

  return (
    <Stack spacing={2.5} className="cc-overview">
      {incidentLoadState === "error" ? (
        <Alert severity="error" className="feed-alert">
          {data.incidentError ?? "Incident API unavailable."}
        </Alert>
      ) : null}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Assigned to desk"
            value={metrics.assigned}
            caption={`${metrics.open} open on your queue`}
            icon={ClipboardList}
            accent="navy"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="En route"
            value={metrics.enRoute}
            caption={`${metrics.onScene} on scene`}
            icon={Truck}
            accent="gold"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="On scene"
            value={metrics.onScene}
            caption={`${metrics.contained} contained · ${metrics.recovery} recovery`}
            icon={Activity}
            accent="green"
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricTile
            label="Priority review"
            value={metrics.priority}
            caption="Flagged incidents needing eyes"
            icon={ShieldAlert}
            accent="red"
          />
        </Grid>
      </Grid>
      <SectionCard
        title="Response posture"
        eyebrow="Active incidents by stage"
        icon={ShieldAlert}
        accent="navy"
        className="cc-triage-card"
        action={
          <Stack direction="row" spacing={1} sx={{
            alignItems: "center"
          }}>
            <span className={`cc-feed-chip cc-feed-chip--${incidentLoadState}`}>
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
        <ResponseLadder incidents={incidents} />
        <p className="cc-triage__foot">
          {metrics.open} open of {incidents.length} incidents assigned to{" "}
          {session.agency}.
        </p>
      </SectionCard>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Capacity snapshot"
            eyebrow="Highest occupancy first"
            icon={Building2}
            accent="green"
            action={
              <Button
                size="small"
                variant="outlined"
                endIcon={<ArrowRight size={15} />}
                onClick={() => onNavigate("capacity")}
              >
                Nearby capacity
              </Button>
            }
          >
            <Stack spacing={1.5}>
              {topShelters.map(({ shelter, pct }) => (
                <div className="cc-shelter-row" key={shelter.id}>
                  <div className="cc-shelter-row__head">
                    <span className="cc-shelter-row__name">{shelter.name}</span>
                    <span className="cc-shelter-row__figure">
                      {shelter.currentOccupancy}/{shelter.capacity} · {pct}%
                    </span>
                  </div>
                  <CapacityMeter
                    value={shelter.currentOccupancy}
                    max={shelter.capacity}
                    tone={utilizationTone(pct)}
                  />
                </div>
              ))}
              {topShelters.length === 0 ? (
                <p className="cc-muted-note">No shelters are loaded yet.</p>
              ) : null}
            </Stack>
            <p className="cc-muted-note">
              {utilization}% overall shelter occupancy · {hospitalBeds}{" "}
              hospital beds available nearby.
            </p>
          </SectionCard>
        </Grid>

        <Grid size={{ xs: 12, lg: 6 }}>
          <SectionCard
            title="Relief & aid"
            eyebrow="Distribution and donations"
            icon={PackageCheck}
            accent="gold"
            action={
              <Button
                size="small"
                variant="outlined"
                endIcon={<ArrowRight size={15} />}
                onClick={() => onNavigate("relief")}
              >
                Manage relief
              </Button>
            }
          >
            <Grid container spacing={1.5}>
              <Grid size={6}>
                <div className="cc-pipeline__stage">
                  <span className="cc-pipeline__value">{metrics.reliefOpen}</span>
                  <span className="cc-pipeline__label">Relief open</span>
                </div>
              </Grid>
              <Grid size={6}>
                <div className="cc-pipeline__stage">
                  <span className="cc-pipeline__value">{reliefPoints.length}</span>
                  <span className="cc-pipeline__label">Published</span>
                </div>
              </Grid>
              <Grid size={6}>
                <div className="cc-pipeline__stage">
                  <span className="cc-pipeline__value">{metrics.aidOpen}</span>
                  <span className="cc-pipeline__label">Aid open</span>
                </div>
              </Grid>
              <Grid size={6}>
                <div className="cc-pipeline__stage">
                  <span className="cc-pipeline__value">{metrics.aidPending}</span>
                  <span className="cc-pipeline__label">Aid pending</span>
                </div>
              </Grid>
            </Grid>
            <Stack direction="row" spacing={1} className="cc-overview__jump">
              <Eyebrow>Jump to</Eyebrow>
              <Button
                size="small"
                variant="text"
                startIcon={<HandHeart size={15} />}
                onClick={() => onNavigate("aid")}
              >
                Aid & donations
              </Button>
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
                value={feedLabel}
                color={incidentLoadState === "ready" ? "success" : "warning"}
              />
              <StatusLine
                label="Agency session"
                value={session.agency}
                color="success"
              />
            </Stack>
          </Grid>
          <Grid size={{ xs: 12, md: 6 }}>
            <Stack spacing={1.25}>
              <StatusLine
                label="Relief points"
                value={`${reliefPoints.length} published`}
                color="success"
              />
              <StatusLine
                label="Aid coordination"
                value={`${aidRequests.length} tracked`}
                color={metrics.aidPending > 0 ? "warning" : "success"}
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
          <Button
            size="small"
            variant="text"
            onClick={() => onNavigate("capacity")}
          >
            Capacity
          </Button>
          <Button
            size="small"
            variant="text"
            onClick={() => onNavigate("relief")}
          >
            Relief
          </Button>
        </Stack>
      </SectionCard>
    </Stack>
  );
}
