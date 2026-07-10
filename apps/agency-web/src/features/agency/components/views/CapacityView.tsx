import {
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { Bed, Building2, HeartPulse, X } from "lucide-react";
import { useState } from "react";
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
import {
  initialHospitalCapacityForm,
  initialShelterOccupancyForm,
} from "../../data";

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

  // Data-entry forms live behind buttons in dialogs, not always-visible panels.
  const [shelterOpen, setShelterOpen] = useState(false);
  const [hospitalOpen, setHospitalOpen] = useState(false);

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
              <Stack
                alignItems="center"
                direction="row"
                justifyContent="space-between"
                spacing={1}
                sx={{ mb: 1.5 }}
              >
                <Typography fontWeight={800} variant="h6">
                  Nearby shelters
                </Typography>
                <Button
                  onClick={() => setShelterOpen(true)}
                  size="small"
                  startIcon={<Bed size={16} />}
                  variant="outlined"
                >
                  Update occupancy
                </Button>
              </Stack>
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
              <Stack
                alignItems="center"
                direction="row"
                justifyContent="space-between"
                spacing={1}
                sx={{ mb: 1.5 }}
              >
                <Typography fontWeight={800} variant="h6">
                  Hospital capacity
                </Typography>
                <Button
                  onClick={() => setHospitalOpen(true)}
                  size="small"
                  startIcon={<HeartPulse size={16} />}
                  variant="outlined"
                >
                  Update capacity
                </Button>
              </Stack>
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
        </>
      )}

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setShelterOpen(false)}
        open={shelterOpen}
      >
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
            fontWeight: 800,
          }}
        >
          Update shelter occupancy
          <IconButton
            aria-label="Close"
            onClick={() => setShelterOpen(false)}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <ShelterOccupancyForm
            form={shelterForm}
            onChange={setShelterForm}
            onSubmit={() => {
              setShelterForm(initialShelterOccupancyForm);
              setShelterOpen(false);
            }}
          />
        </DialogContent>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setHospitalOpen(false)}
        open={hospitalOpen}
      >
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
            fontWeight: 800,
          }}
        >
          Update hospital capacity
          <IconButton
            aria-label="Close"
            onClick={() => setHospitalOpen(false)}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <HospitalCapacityUpdateForm
            form={hospitalForm}
            onChange={setHospitalForm}
            onSubmit={() => {
              setHospitalForm(initialHospitalCapacityForm);
              setHospitalOpen(false);
            }}
          />
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
