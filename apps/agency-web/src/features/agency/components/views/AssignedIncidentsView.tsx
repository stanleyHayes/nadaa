import {
  Alert,
  Box,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  Stack,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { Activity, ClipboardList, ShieldAlert, Truck, X } from "lucide-react";
import type { AgencyData } from "../../useAgencyData";
import { MetricTile, ViewIntro } from "../primitives";
import {
  EmptyState,
  ErrorState,
  IncidentDetail,
  IncidentFilters,
  IncidentListItem,
  LoadingState,
  StatusUpdateForm,
} from "..";

export function AssignedIncidentsView({ data }: { data: AgencyData }) {
  const {
    incidents,
    filteredIncidents,
    incidentLoadState,
    incidentError,
    filters,
    setFilters,
    selectedIncident,
    selectedIncidentId,
    selectIncident,
    deselectIncident,
    loadIncidents,
    statusForm,
    setStatusForm,
    statusUpdateState,
    statusUpdateError,
    handleStatusUpdate,
    metrics,
  } = data;

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        title="Assigned incidents"
        description="Triage the incidents on your desk, then open one to review detail and update its status."
        icon={ClipboardList}
      />

      <Grid container spacing={2}>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricTile
            label="Assigned"
            value={metrics.assigned}
            icon={ClipboardList}
            accent="navy"
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricTile
            label="En route"
            value={metrics.enRoute}
            icon={Truck}
            accent="gold"
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricTile
            label="On scene"
            value={metrics.onScene}
            icon={Activity}
            accent="green"
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <MetricTile
            label="Priority"
            value={metrics.priority}
            icon={ShieldAlert}
            accent="red"
          />
        </Grid>
      </Grid>

      <IncidentFilters filters={filters} onChange={setFilters} />

      {incidentError && incidents.length > 0 ? (
        <Alert severity="warning">{incidentError} Showing fallback data.</Alert>
      ) : null}

      {incidentLoadState === "loading" ? (
        <LoadingState message="Loading assigned incidents" />
      ) : incidentLoadState === "error" && !incidents.length ? (
        <ErrorState
          message={incidentError ?? "Could not load incidents"}
          onRetry={loadIncidents}
        />
      ) : filteredIncidents.length === 0 ? (
        <EmptyState message="No incidents match the current filters." />
      ) : (
        <Stack spacing={2}>
          {filteredIncidents.map((incident) => (
            <IncidentListItem
              key={incident.id}
              incident={incident}
              onClick={() => selectIncident(incident.id)}
              selected={selectedIncidentId === incident.id}
            />
          ))}
        </Stack>
      )}

      <Dialog
        open={Boolean(selectedIncident)}
        onClose={deselectIncident}
        maxWidth="md"
        fullWidth
        scroll="paper"
        fullScreen={fullScreen}
        aria-labelledby="incident-detail-title"
      >
        <DialogTitle
          id="incident-detail-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>{selectedIncident?.reference ?? "Incident detail"}</span>
          <IconButton
            aria-label="Close incident detail"
            onClick={deselectIncident}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          {selectedIncident ? (
            <Stack spacing={2.5}>
              <IncidentDetail incident={selectedIncident} />
              <Box>
                <Box mb={2}>
                  <Typography fontWeight={800} variant="h6">
                    Update status
                  </Typography>
                  <Typography color="text.secondary" variant="body2">
                    Advance {selectedIncident.reference} through its response
                    lifecycle.
                  </Typography>
                </Box>
                <StatusUpdateForm
                  currentStatus={selectedIncident.status}
                  form={statusForm}
                  onChange={setStatusForm}
                  onSubmit={handleStatusUpdate}
                  submitLabel={
                    statusUpdateState === "loading"
                      ? "Updating..."
                      : "Update status"
                  }
                />
                {statusUpdateState === "success" ? (
                  <Alert severity="success" sx={{ mt: 2 }}>
                    Status updated successfully.
                  </Alert>
                ) : null}
                {statusUpdateState === "error" && statusUpdateError ? (
                  <Alert severity="error" sx={{ mt: 2 }}>
                    {statusUpdateError}
                  </Alert>
                ) : null}
              </Box>
            </Stack>
          ) : null}
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
