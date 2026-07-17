import {
  Alert,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { Bed, Building2, HeartPulse, X } from "lucide-react";
import { useState } from "react";
import { canManageShelterResources } from "@/app/session";
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
    session,
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
    selectedShelterId,
    selectShelterTarget,
    selectedHospitalId,
    selectHospitalTarget,
    capacityUpdateState,
    capacityUpdateError,
    handleShelterOccupancyUpdate,
    handleHospitalCapacityUpdate,
  } = data;

  const canWrite = canManageShelterResources(session);

  // Data-entry forms live behind buttons in dialogs, not always-visible panels.
  const [shelterOpen, setShelterOpen] = useState(false);
  const [hospitalOpen, setHospitalOpen] = useState(false);

  const contextLabel = selectedIncident
    ? `near ${selectedIncident.reference}`
    : "around Accra Metro";

  const openShelterDialog = () => {
    setShelterForm(initialShelterOccupancyForm);
    selectShelterTarget(shelters[0]?.id ?? "");
    setShelterOpen(true);
  };

  const openHospitalDialog = () => {
    setHospitalForm(initialHospitalCapacityForm);
    selectHospitalTarget(hospitals[0]?.id ?? "");
    setHospitalOpen(true);
  };

  const submitShelter = async () => {
    if (await handleShelterOccupancyUpdate()) {
      setShelterForm(initialShelterOccupancyForm);
      setShelterOpen(false);
    }
  };

  const submitHospital = async () => {
    if (await handleHospitalCapacityUpdate()) {
      setHospitalForm(initialHospitalCapacityForm);
      setHospitalOpen(false);
    }
  };

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
                direction="row"
                spacing={1}
                sx={{
                  alignItems: "center",
                  justifyContent: "space-between",
                  mb: 1.5
                }}>
                <Typography variant="h6" sx={{
                  fontWeight: 800
                }}>
                  Nearby shelters
                </Typography>
                {canWrite ? (
                  <Button
                    disabled={shelters.length === 0}
                    onClick={openShelterDialog}
                    size="small"
                    startIcon={<Bed size={16} />}
                    variant="outlined"
                  >
                    Update occupancy
                  </Button>
                ) : null}
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
                direction="row"
                spacing={1}
                sx={{
                  alignItems: "center",
                  justifyContent: "space-between",
                  mb: 1.5
                }}>
                <Typography variant="h6" sx={{
                  fontWeight: 800
                }}>
                  Hospital capacity
                </Typography>
                {canWrite ? (
                  <Button
                    disabled={hospitals.length === 0}
                    onClick={openHospitalDialog}
                    size="small"
                    startIcon={<HeartPulse size={16} />}
                    variant="outlined"
                  >
                    Update capacity
                  </Button>
                ) : null}
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
              <Typography variant="h6" sx={{
                fontWeight: 800
              }}>
                Nearby road closures
              </Typography>
              <Stack spacing={2}>
                {roadClosures.map((closure) => (
                  <Paper key={closure.id} sx={{ p: 2 }}>
                    <Stack spacing={0.5}>
                      <Typography sx={{
                        fontWeight: 700
                      }}>
                        {closure.roadName}
                      </Typography>
                      <Typography variant="body2" sx={{
                        color: "text.secondary"
                      }}>
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
              <Typography variant="h6" sx={{
                fontWeight: 800
              }}>
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
          <Stack spacing={2}>
            <TextField
              label="Shelter"
              onChange={(event) => selectShelterTarget(event.target.value)}
              select
              size="small"
              value={selectedShelterId ?? ""}
            >
              {shelters.map((shelter) => (
                <MenuItem key={shelter.id} value={shelter.id}>
                  {shelter.name}
                </MenuItem>
              ))}
            </TextField>
            <ShelterOccupancyForm
              form={shelterForm}
              onChange={setShelterForm}
              onSubmit={() => void submitShelter()}
              submitting={capacityUpdateState === "loading"}
            />
            {capacityUpdateState === "error" && capacityUpdateError ? (
              <Alert severity="error">{capacityUpdateError}</Alert>
            ) : null}
          </Stack>
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
          <Stack spacing={2}>
            <TextField
              label="Hospital"
              onChange={(event) => selectHospitalTarget(event.target.value)}
              select
              size="small"
              value={selectedHospitalId ?? ""}
            >
              {hospitals.map((facility) => (
                <MenuItem key={facility.id} value={facility.id}>
                  {facility.name}
                </MenuItem>
              ))}
            </TextField>
            <HospitalCapacityUpdateForm
              form={hospitalForm}
              onChange={setHospitalForm}
              onSubmit={() => void submitHospital()}
              submitting={capacityUpdateState === "loading"}
            />
            {capacityUpdateState === "error" && capacityUpdateError ? (
              <Alert severity="error">{capacityUpdateError}</Alert>
            ) : null}
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
