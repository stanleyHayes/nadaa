import { Alert, Box, Grid, Paper, Stack, Typography } from "@mui/material";
import { Activity, ClipboardList, ShieldAlert, Truck } from "lucide-react";
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
    loadIncidents,
    statusForm,
    setStatusForm,
    statusUpdateState,
    statusUpdateError,
    handleStatusUpdate,
    metrics,
  } = data;

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

      <Grid container spacing={2.5} alignItems="flex-start">
        <Grid size={{ xs: 12, lg: 5 }}>
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
        </Grid>

        <Grid size={{ xs: 12, lg: 7 }}>
          {!selectedIncident ? (
            <EmptyState message="Select an incident from the list to view details and update status." />
          ) : (
            <Stack spacing={2.5}>
              <Paper sx={{ p: 3 }}>
                <IncidentDetail incident={selectedIncident} />
              </Paper>
              <Paper sx={{ p: 3 }}>
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
              </Paper>
            </Stack>
          )}
        </Grid>
      </Grid>
    </Stack>
  );
}
