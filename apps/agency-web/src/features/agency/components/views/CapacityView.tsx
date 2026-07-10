import { Grid, Paper, Stack, Typography } from "@mui/material";
import { Building2 } from "lucide-react";
import type { AgencyData } from "../../useAgencyData";
import type { ViewId } from "../../navigation";
import { ViewIntro } from "../primitives";
import {
  EmptyState,
  HospitalCapacityCard,
  HospitalCapacityUpdateForm,
  LoadingState,
  ReliefPointCard,
  ShelterCard,
  ShelterOccupancyForm,
} from "..";
import { initialHospitalCapacityForm, initialShelterOccupancyForm } from "../../data";

export function CapacityView({
  data,
  onNavigate,
}: {
  data: AgencyData;
  onNavigate: (view: ViewId) => void;
}) {
  const {
    selectedIncident,
    capacityLoadState,
    shelters,
    hospitals,
    roadClosures,
    nearbyReliefPoints,
    selectedReliefPointId,
    selectReliefPoint,
    shelterForm,
    setShelterForm,
    hospitalForm,
    setHospitalForm,
  } = data;

  const contextLabel = selectedIncident
    ? `near ${selectedIncident.reference}`
    : "around Accra Metro";

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        title="Nearby capacity"
        description={`Shelter occupancy, hospital beds, and road closures ${contextLabel}. Open an incident to recentre on its scene.`}
        icon={Building2}
      />

      {capacityLoadState === "loading" ? (
        <LoadingState message="Loading nearby capacity" />
      ) : (
        <>
          <Grid container spacing={2.5}>
            <Grid size={{ xs: 12, md: 6 }}>
              <Typography fontWeight={800} gutterBottom variant="h6">
                Nearby shelters
              </Typography>
              {shelters.length === 0 ? (
                <EmptyState message="No shelters found nearby." />
              ) : (
                <Stack spacing={2}>
                  {shelters.map((shelter) => (
                    <ShelterCard key={shelter.id} shelter={shelter} />
                  ))}
                </Stack>
              )}
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <Typography fontWeight={800} gutterBottom variant="h6">
                Hospital capacity
              </Typography>
              {hospitals.length === 0 ? (
                <EmptyState message="No hospitals found nearby." />
              ) : (
                <Stack spacing={2}>
                  {hospitals.map((facility) => (
                    <HospitalCapacityCard key={facility.id} facility={facility} />
                  ))}
                </Stack>
              )}
            </Grid>
          </Grid>

          {roadClosures.length > 0 ? (
            <Stack spacing={2}>
              <Typography fontWeight={800} variant="h6">
                Nearby road closures
              </Typography>
              <Stack spacing={2}>
                {roadClosures.map((closure) => (
                  <Paper key={closure.id} sx={{ p: 2 }}>
                    <Stack spacing={0.5}>
                      <Typography fontWeight={700}>
                        {closure.roadName}
                      </Typography>
                      <Typography color="text.secondary" variant="body2">
                        {closure.reason ?? "Road closure"} · {closure.severity}{" "}
                        · {closure.status}
                      </Typography>
                      {closure.detourNote ? (
                        <Typography variant="body2">
                          Detour: {closure.detourNote}
                        </Typography>
                      ) : null}
                    </Stack>
                  </Paper>
                ))}
              </Stack>
            </Stack>
          ) : null}

          {nearbyReliefPoints.length > 0 ? (
            <Stack spacing={2}>
              <Typography fontWeight={800} variant="h6">
                Nearby relief distribution points
              </Typography>
              <Stack spacing={2}>
                {nearbyReliefPoints.map((point) => (
                  <ReliefPointCard
                    key={point.id}
                    onSelect={() => {
                      selectReliefPoint(point.id);
                      onNavigate("relief");
                    }}
                    point={point}
                    selected={selectedReliefPointId === point.id}
                  />
                ))}
              </Stack>
            </Stack>
          ) : null}

          <Grid container spacing={2.5}>
            <Grid size={{ xs: 12, md: 6 }}>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  Update shelter occupancy
                </Typography>
                <ShelterOccupancyForm
                  form={shelterForm}
                  onChange={setShelterForm}
                  onSubmit={() => {
                    setShelterForm(initialShelterOccupancyForm);
                  }}
                />
              </Paper>
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  Update hospital capacity
                </Typography>
                <HospitalCapacityUpdateForm
                  form={hospitalForm}
                  onChange={setHospitalForm}
                  onSubmit={() => {
                    setHospitalForm(initialHospitalCapacityForm);
                  }}
                />
              </Paper>
            </Grid>
          </Grid>
        </>
      )}
    </Stack>
  );
}
