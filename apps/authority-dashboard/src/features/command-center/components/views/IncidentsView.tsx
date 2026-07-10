import { useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  LinearProgress,
  MenuItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import {
  Eye,
  Filter,
  LocateFixed,
  MapPinned,
  RefreshCw,
  X,
} from "lucide-react";
import type { CommandData } from "../../useCommandData";
import {
  CommandSelect,
  EmptyState,
  HazardChip,
  IncidentMap,
  PrivacyChip,
  ScrollableTable,
  SeverityChip,
} from "../shared";
import { IncidentDetailPanel } from "../IncidentDetailPanel";
import { RoutePlannerPanel } from "../RoutePlannerPanel";
import { SectionCard } from "../primitives";
import {
  formatIncidentAge,
  hazardLabel,
  severityLabel,
  statusLabel,
} from "../../utils";

export function IncidentsView({ data }: { data: CommandData }) {
  const {
    incidents,
    filteredIncidents,
    filters,
    filterOptions,
    updateFilter,
    loadState,
    refreshIncidents,
    selectedIncident,
    setSelectedIncidentId,
    imageryFeatures,
    statusBusy,
    statusFeedback,
    statusForm,
    updateStatusForm,
    verifySelectedIncident,
    updateIncidentStatus,
    abuseBusy,
    abuseFeedback,
    abuseForm,
    updateAbuseForm,
    reviewSelectedIncidentAbuse,
    assignmentBusy,
    assignmentFeedback,
    assignmentForm,
    updateAssignmentForm,
    assignSelectedIncident,
    duplicateReviewCandidates,
    selectedDuplicateIds,
    toggleDuplicateSelection,
    mergeBusy,
    mergeFeedback,
    mergeSelectedDuplicates,
  } = data;

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));
  const [detailOpen, setDetailOpen] = useState(false);

  const openIncident = (incidentId: string) => {
    setSelectedIncidentId(incidentId);
    setDetailOpen(true);
  };

  const closeDetail = () => setDetailOpen(false);

  return (
    <Stack spacing={2.5}>
      <Stack
        direction={{ xs: "column", sm: "row" }}
        justifyContent="space-between"
        alignItems={{ xs: "flex-start", sm: "center" }}
        gap={1.5}
      >
        <Typography color="text.secondary">
          Monitor emergencies by place, severity, hazard, time, and response
          status.
        </Typography>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
          <span className={`cc-feed-chip cc-feed-chip--${loadState}`}>
            <Eye size={13} />
            {loadState === "ready"
              ? "Live API"
              : loadState === "empty"
                ? "No active incidents"
                : loadState === "loading"
                  ? "Connecting"
                  : "Offline"}
          </span>
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshCw size={16} />}
            onClick={() => void refreshIncidents()}
            disabled={loadState === "loading"}
          >
            Refresh
          </Button>
        </Stack>
      </Stack>

      {loadState === "loading" ? (
        <LinearProgress className="feed-progress" />
      ) : null}

      <SectionCard title="Filters" eyebrow="Narrow the queue" icon={Filter}>
        <Grid container spacing={1.5}>
          <Grid size={{ xs: 12, sm: 6, md: 2.4 }}>
            <CommandSelect
              label="Hazard"
              value={filters.hazard}
              onChange={updateFilter("hazard")}
            >
              <MenuItem value="all">All hazards</MenuItem>
              {filterOptions.hazards.map((hazard) => (
                <MenuItem value={hazard} key={hazard}>
                  {hazardLabel(hazard)}
                </MenuItem>
              ))}
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 2.4 }}>
            <CommandSelect
              label="Region / district"
              value={filters.regionDistrict}
              onChange={updateFilter("regionDistrict")}
            >
              <MenuItem value="all">All districts</MenuItem>
              {filterOptions.regionDistricts.map((district) => (
                <MenuItem value={district} key={district}>
                  {district}
                </MenuItem>
              ))}
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 2.4 }}>
            <CommandSelect
              label="Severity"
              value={filters.severity}
              onChange={updateFilter("severity")}
            >
              <MenuItem value="all">All severities</MenuItem>
              {filterOptions.severities.map((severity) => (
                <MenuItem value={severity} key={severity}>
                  {severityLabel(severity)}
                </MenuItem>
              ))}
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 2.4 }}>
            <CommandSelect
              label="Status"
              value={filters.status}
              onChange={updateFilter("status")}
            >
              <MenuItem value="all">All statuses</MenuItem>
              {filterOptions.statuses.map((status) => (
                <MenuItem value={status} key={status}>
                  {statusLabel(status)}
                </MenuItem>
              ))}
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 2.4 }}>
            <CommandSelect
              label="Time"
              value={filters.time}
              onChange={updateFilter("time")}
            >
              <MenuItem value="all">Any time</MenuItem>
              <MenuItem value="1h">Last hour</MenuItem>
              <MenuItem value="6h">Last 6 hours</MenuItem>
              <MenuItem value="24h">Last 24 hours</MenuItem>
            </CommandSelect>
          </Grid>
        </Grid>
      </SectionCard>

      <SectionCard
        title="Incident map"
        eyebrow={`${filteredIncidents.length} visible of ${incidents.length}`}
        icon={MapPinned}
        action={
          <Stack direction="row" spacing={0.75} flexWrap="wrap">
            {filterOptions.hazards.slice(0, 4).map((hazard) => (
              <HazardChip key={hazard} hazard={hazard} />
            ))}
          </Stack>
        }
      >
        <IncidentMap
          incidents={filteredIncidents}
          imageryFeatures={imageryFeatures}
          selectedIncidentId={selectedIncident?.id}
          onSelect={openIncident}
        />
      </SectionCard>

      <SectionCard
        title="Incident queue"
        eyebrow="Click a row to open details"
        icon={LocateFixed}
      >
        {filteredIncidents.length ? (
          <ScrollableTable label="Incident queue table">
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Reference</TableCell>
                  <TableCell>Hazard</TableCell>
                  <TableCell>District</TableCell>
                  <TableCell>Severity</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Privacy</TableCell>
                  <TableCell>Assigned</TableCell>
                  <TableCell>Age</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredIncidents.map((incident) => (
                  <TableRow
                    key={incident.id}
                    hover
                    selected={incident.id === selectedIncident?.id}
                    onClick={() => openIncident(incident.id)}
                    className="incident-row"
                  >
                    <TableCell>
                      <Typography variant="subtitle2">
                        {incident.reference}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {incident.locality}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <HazardChip hazard={incident.type} />
                    </TableCell>
                    <TableCell>{incident.district}</TableCell>
                    <TableCell>
                      <SeverityChip severity={incident.severity} />
                    </TableCell>
                    <TableCell>{statusLabel(incident.status)}</TableCell>
                    <TableCell>
                      <PrivacyChip incident={incident} />
                    </TableCell>
                    <TableCell>{incident.assignedAgency}</TableCell>
                    <TableCell>
                      {formatIncidentAge(incident.createdAt)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollableTable>
        ) : (
          <EmptyState
            title="No incidents match these filters"
            detail="Adjust filters or refresh the feed."
          />
        )}
      </SectionCard>

      {loadState === "empty" ? (
        <Alert severity="info" className="feed-alert">
          No incidents are currently in the command queue.
        </Alert>
      ) : null}

      <Dialog
        open={detailOpen}
        onClose={closeDetail}
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
            onClick={closeDetail}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5}>
            <IncidentDetailPanel
              abuseBusy={abuseBusy}
              abuseFeedback={abuseFeedback}
              abuseForm={abuseForm}
              assignmentBusy={assignmentBusy}
              assignmentFeedback={assignmentFeedback}
              assignmentForm={assignmentForm}
              busy={statusBusy}
              duplicateCandidates={duplicateReviewCandidates}
              feedback={statusFeedback}
              form={statusForm}
              incident={selectedIncident}
              mergeBusy={mergeBusy}
              mergeFeedback={mergeFeedback}
              onAssign={assignSelectedIncident}
              onMergeDuplicates={mergeSelectedDuplicates}
              onReviewAbuse={reviewSelectedIncidentAbuse}
              onToggleDuplicate={toggleDuplicateSelection}
              onUpdateAbuseForm={updateAbuseForm}
              onUpdateAssignmentForm={updateAssignmentForm}
              onUpdateForm={updateStatusForm}
              onUpdateStatus={updateIncidentStatus}
              onVerify={verifySelectedIncident}
              selectedDuplicateIds={selectedDuplicateIds}
            />

            <RoutePlannerPanel selectedIncident={selectedIncident} />
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
