import { Grid, Stack } from "@mui/material";
import { ShieldCheck } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { formatPercent } from "../../utils";
import { MfaSupportPanel } from "../MfaSupportPanel";
import { CoverageMeter, SectionCard, ViewIntro } from "../primitives";

function coverageTone(pct: number): "green" | "gold" | "red" {
  if (pct >= 90) {
    return "green";
  }
  if (pct >= 70) {
    return "gold";
  }
  return "red";
}

export function MfaView({ data }: { data: AdminData }) {
  const { agencies, users } = data;
  const mfaReady = users.filter((user) => user.mfaEnabled).length;
  const mfaCoverage =
    users.length > 0 ? Math.round((mfaReady / users.length) * 100) : 0;
  const board = [...agencies].sort(
    (a, b) => (a.mfaCoverage ?? -1) - (b.mfaCoverage ?? -1),
  );

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={ShieldCheck}
        title="MFA readiness"
        description="Two-step verification coverage across agencies and the authority users still awaiting setup."
      />
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <SectionCard
            title="Coverage by agency"
            eyebrow={`${mfaCoverage}% of users verified`}
            icon={ShieldCheck}
            accent="navy"
          >
            <Stack spacing={1.5}>
              {board.map((agency) => (
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
              {board.length === 0 ? (
                <p className="cc-muted-note">No agencies are registered yet.</p>
              ) : null}
            </Stack>
          </SectionCard>
        </Grid>
        <Grid size={{ xs: 12, lg: 5 }}>
          <MfaSupportPanel users={users} />
        </Grid>
      </Grid>
    </Stack>
  );
}
