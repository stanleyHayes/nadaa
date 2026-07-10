import { Grid, Stack } from "@mui/material";
import DamageClaimsPanel from "../DamageClaimsPanel";
import { DonationPanel } from "../DonationPanel";
import MissingPersonsPanel from "../MissingPersonsPanel";

export function RecoveryView() {
  return (
    <Stack spacing={2.5}>
      <Grid container spacing={2.5} alignItems="flex-start">
        <Grid size={{ xs: 12, xl: 6 }}>
          <DamageClaimsPanel />
        </Grid>
        <Grid size={{ xs: 12, xl: 6 }}>
          <MissingPersonsPanel />
        </Grid>
      </Grid>
      <DonationPanel />
    </Stack>
  );
}
