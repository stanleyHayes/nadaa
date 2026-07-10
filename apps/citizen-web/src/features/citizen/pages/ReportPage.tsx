import { ChangeEvent, FormEvent, useState } from "react";
import {
  Alert,
  Button,
  Chip,
  FormControl,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { ImagePlus, Loader2, LocateFixed, Siren } from "lucide-react";
import type {
  CreateIncidentRequest,
  CreateIncidentResponse,
  HazardType,
  IncidentMediaContentType,
  IncidentUrgency,
} from "@nadaa/shared-types";
import { INCIDENT_API_BASE } from "@/app/config";
import {
  hazardOptions,
  initialReportForm,
  mediaSizeLimits,
  supportedMediaTypes,
  urgencyOptions,
} from "../data";
import type { ReportForm, ReportState } from "../types";
import { extractAPIError, formatFileSize, initiateMediaUploads } from "../utils";
import { useCitizenSession } from "../session";
import { PageHeader, Reveal, SavedReports } from "../components";
import { PageBanner } from "../components/PageBanner";

/**
 * Incident reporting (route `/report`). Self-contained migration of the legacy
 * `#report` form together with the signed-in `SavedReports` list.
 */
export function ReportPage() {
  const { session, savedReports, saveReport } = useCitizenSession();
  const [reportForm, setReportForm] = useState<ReportForm>(initialReportForm);
  const [reportState, setReportState] = useState<ReportState>({
    status: "idle",
  });
  const [reportErrors, setReportErrors] = useState<
    Partial<Record<keyof ReportForm, string>>
  >({});

  const clearReportError = (key: keyof ReportForm) => {
    setReportErrors((current) => {
      if (!current[key]) return current;
      const next = { ...current };
      delete next[key];
      return next;
    });
  };

  const updateReportForm = <Key extends keyof ReportForm>(
    key: Key,
    value: ReportForm[Key],
  ) => {
    setReportForm((current) => ({ ...current, [key]: value }));
  };

  const useCurrentLocation = () => {
    if (!navigator.geolocation) {
      setReportState({
        status: "error",
        message: "Location is not available on this device.",
      });
      return;
    }

    setReportState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setReportForm((current) => ({
          ...current,
          lat: position.coords.latitude.toFixed(6),
          lng: position.coords.longitude.toFixed(6),
        }));
        setReportState({ status: "idle" });
      },
      () => {
        setReportState({
          status: "error",
          message: "Location permission was not granted.",
        });
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  };

  const handleFileSelection = (event: ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(event.target.files ?? []);
    event.currentTarget.value = "";

    if (selectedFiles.length > 10) {
      setReportState({
        status: "error",
        message: "Attach at most 10 media files to one report.",
      });
      return;
    }

    const invalidFile = selectedFiles.find((file) => {
      if (
        !supportedMediaTypes.includes(file.type as IncidentMediaContentType)
      ) {
        return true;
      }

      return (
        file.size <= 0 ||
        file.size > mediaSizeLimits[file.type as IncidentMediaContentType]
      );
    });

    if (invalidFile) {
      setReportState({
        status: "error",
        message: `${invalidFile.name} is not supported or is too large for this report.`,
      });
      return;
    }

    updateReportForm("files", selectedFiles);
    setReportState({ status: "idle" });
  };

  const submitReport = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const lat = Number(reportForm.lat);
    const lng = Number(reportForm.lng);
    const peopleAffected = Number(reportForm.peopleAffected || 0);
    const nextErrors: Partial<Record<keyof ReportForm, string>> = {};

    if (
      !Number.isFinite(lat) ||
      lat < -90 ||
      lat > 90 ||
      !Number.isFinite(lng) ||
      lng < -180 ||
      lng > 180
    ) {
      nextErrors.lat = "Enter a valid latitude.";
      nextErrors.lng = "Enter a valid longitude.";
    }

    if (reportForm.description.trim().length < 5) {
      nextErrors.description = "Add a short description of what happened.";
    }

    if (!Number.isInteger(peopleAffected) || peopleAffected < 0) {
      nextErrors.peopleAffected = "People affected must be zero or more.";
    }

    if (Object.keys(nextErrors).length > 0) {
      setReportErrors(nextErrors);
      setReportState({
        status: "error",
        message: "Please correct the highlighted fields before sending.",
      });
      return;
    }

    if (!navigator.onLine) {
      setReportState({
        status: "error",
        message:
          "You appear to be offline. Keep this report open and try again when the connection returns.",
      });
      return;
    }

    setReportState({ status: "loading", message: "Sending report" });

    try {
      const mediaIds = await initiateMediaUploads(reportForm.files);
      const payload: CreateIncidentRequest = {
        type: reportForm.hazard,
        description: reportForm.description.trim(),
        location: { lat, lng },
        peopleAffected,
        injuriesReported: reportForm.injuriesReported,
        urgency: reportForm.urgency,
        anonymous: reportForm.anonymous,
        contactPermission: reportForm.anonymous
          ? false
          : reportForm.contactPermission,
        accessibilityNeeds: reportForm.accessibilityNeeds.trim() || undefined,
        media: mediaIds,
        reporter: reportForm.anonymous
          ? undefined
          : {
              userId: "usr_demo_citizen",
              phone: reportForm.contactPermission ? "+233200000000" : undefined,
            },
      };

      const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(await extractAPIError(response));
      }

      const incident = (await response.json()) as CreateIncidentResponse;
      if (session) {
        saveReport({
          reference: incident.reference,
          hazard: reportForm.hazard,
          urgency: reportForm.urgency,
          priorityReview: incident.priorityReview,
          at: new Date().toISOString(),
        });
      }
      setReportState({
        status: "success",
        reference: incident.reference,
        priorityReview: incident.priorityReview,
      });
      setReportErrors({});
      setReportForm(initialReportForm);
    } catch (error) {
      setReportState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not send report.",
      });
    }
  };

  return (
    <>
      <PageBanner
        eyebrow="Report an incident"
        subtitle="Tell NADMO what you're seeing — with photos and location, online or offline."
        title="Report a flood or hazard"
      />
      <div className="citizen-shell">
        <div className="citizen-section">
          <Reveal delay={80}>
            <Paper
              className="surface report-surface"
              id="report"
              component="section"
            >
              <PageHeader
                icon={Siren}
                title="Report an incident"
                subtitle="Flag a flood, fire, road crash or hazard for responders."
                tone="gold"
              />
              <Stack
                component="form"
                spacing={1.5}
                onSubmit={submitReport}
                noValidate
              >
                <FormControl fullWidth error={Boolean(reportErrors.hazard)}>
                  <InputLabel id="report-hazard-label">Hazard type</InputLabel>
                  <Select
                    id="report-hazard"
                    labelId="report-hazard-label"
                    value={reportForm.hazard}
                    label="Hazard type"
                    onChange={(event) => {
                      clearReportError("hazard");
                      updateReportForm(
                        "hazard",
                        event.target.value as HazardType,
                      );
                    }}
                    aria-describedby="report-hazard-error"
                  >
                    {hazardOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                  {reportErrors.hazard ? (
                    <FormHelperText id="report-hazard-error">
                      {reportErrors.hazard}
                    </FormHelperText>
                  ) : null}
                </FormControl>
                <Grid container spacing={1.25}>
                  <Grid size={{ xs: 6 }}>
                    <TextField
                      id="report-lat"
                      label="Latitude"
                      value={reportForm.lat}
                      onChange={(event) => {
                        clearReportError("lat");
                        updateReportForm("lat", event.target.value);
                      }}
                      fullWidth
                      inputMode="decimal"
                      error={Boolean(reportErrors.lat)}
                      helperText={reportErrors.lat}
                      FormHelperTextProps={{ id: "report-lat-error" }}
                      inputProps={{
                        "aria-describedby": "report-lat-error",
                      }}
                    />
                  </Grid>
                  <Grid size={{ xs: 6 }}>
                    <TextField
                      id="report-lng"
                      label="Longitude"
                      value={reportForm.lng}
                      onChange={(event) => {
                        clearReportError("lng");
                        updateReportForm("lng", event.target.value);
                      }}
                      fullWidth
                      inputMode="decimal"
                      error={Boolean(reportErrors.lng)}
                      helperText={reportErrors.lng}
                      FormHelperTextProps={{ id: "report-lng-error" }}
                      inputProps={{
                        "aria-describedby": "report-lng-error",
                      }}
                    />
                  </Grid>
                </Grid>
                <Button
                  type="button"
                  variant="outlined"
                  startIcon={<LocateFixed size={18} />}
                  onClick={useCurrentLocation}
                  disabled={reportState.status === "loading"}
                >
                  Use GPS
                </Button>
                <TextField
                  id="report-description"
                  label="What happened?"
                  value={reportForm.description}
                  onChange={(event) => {
                    clearReportError("description");
                    updateReportForm("description", event.target.value);
                  }}
                  multiline
                  minRows={3}
                  error={Boolean(reportErrors.description)}
                  helperText={reportErrors.description}
                  FormHelperTextProps={{ id: "report-description-error" }}
                  inputProps={{
                    maxLength: 2000,
                    "aria-describedby": "report-description-error",
                  }}
                />
                <Grid container spacing={1.25}>
                  <Grid size={{ xs: 6 }}>
                    <TextField
                      id="report-people-affected"
                      label="People affected"
                      value={reportForm.peopleAffected}
                      onChange={(event) => {
                        clearReportError("peopleAffected");
                        updateReportForm("peopleAffected", event.target.value);
                      }}
                      fullWidth
                      inputMode="numeric"
                      error={Boolean(reportErrors.peopleAffected)}
                      helperText={reportErrors.peopleAffected}
                      FormHelperTextProps={{
                        id: "report-people-affected-error",
                      }}
                      inputProps={{
                        "aria-describedby": "report-people-affected-error",
                      }}
                    />
                  </Grid>
                  <Grid size={{ xs: 6 }}>
                    <FormControl fullWidth error={Boolean(reportErrors.urgency)}>
                      <InputLabel id="report-urgency-label">Urgency</InputLabel>
                      <Select
                        id="report-urgency"
                        labelId="report-urgency-label"
                        value={reportForm.urgency}
                        label="Urgency"
                        onChange={(event) => {
                          clearReportError("urgency");
                          updateReportForm(
                            "urgency",
                            event.target.value as IncidentUrgency,
                          );
                        }}
                        aria-describedby="report-urgency-error"
                      >
                        {urgencyOptions.map((option) => (
                          <MenuItem key={option.value} value={option.value}>
                            {option.label}
                          </MenuItem>
                        ))}
                      </Select>
                      {reportErrors.urgency ? (
                        <FormHelperText id="report-urgency-error">
                          {reportErrors.urgency}
                        </FormHelperText>
                      ) : null}
                    </FormControl>
                  </Grid>
                </Grid>
                {reportForm.urgency === "life_threatening" ? (
                  <Alert severity="error" className="warning-alert">
                    <Typography variant="body2">
                      Call 112 immediately after sending this report.
                    </Typography>
                  </Alert>
                ) : null}
                <TextField
                  id="report-accessibility-needs"
                  label="Accessibility needs"
                  value={reportForm.accessibilityNeeds}
                  onChange={(event) =>
                    updateReportForm("accessibilityNeeds", event.target.value)
                  }
                  inputProps={{ maxLength: 500 }}
                />
                <Button
                  component="label"
                  variant="outlined"
                  startIcon={<ImagePlus size={18} />}
                >
                  Add media
                  <input
                    type="file"
                    hidden
                    multiple
                    accept={supportedMediaTypes.join(",")}
                    onChange={handleFileSelection}
                  />
                </Button>
                {reportForm.files.length > 0 ? (
                  <Stack spacing={0.75}>
                    {reportForm.files.map((file) => (
                      <Chip
                        key={`${file.name}-${file.size}`}
                        label={`${file.name} · ${formatFileSize(file.size)}`}
                        className="media-chip"
                      />
                    ))}
                  </Stack>
                ) : null}
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                >
                  <Typography>Injuries reported</Typography>
                  <Switch
                    checked={reportForm.injuriesReported}
                    onChange={(event) =>
                      updateReportForm("injuriesReported", event.target.checked)
                    }
                  />
                </Stack>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                >
                  <Typography>Report anonymously</Typography>
                  <Switch
                    checked={reportForm.anonymous}
                    onChange={(event) =>
                      setReportForm((current) => ({
                        ...current,
                        anonymous: event.target.checked,
                        contactPermission: event.target.checked
                          ? false
                          : current.contactPermission,
                      }))
                    }
                  />
                </Stack>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                >
                  <Typography>Allow contact</Typography>
                  <Switch
                    checked={
                      !reportForm.anonymous && reportForm.contactPermission
                    }
                    onChange={(event) =>
                      updateReportForm("contactPermission", event.target.checked)
                    }
                    disabled={reportForm.anonymous}
                  />
                </Stack>
                <Alert severity="info" className="warning-alert">
                  NADAA uses report location to route emergency response, detect
                  duplicates, and coordinate verified authority actions.
                  Anonymous reports hide your identity; disabling contact means
                  responders cannot call you back through this report.
                </Alert>
                {reportState.status === "error" ? (
                  <Alert severity="error" className="warning-alert">
                    {reportState.message}
                  </Alert>
                ) : null}
                {reportState.status === "success" ? (
                  <Alert
                    severity={
                      reportState.priorityReview ? "warning" : "success"
                    }
                    className="warning-alert"
                  >
                    <Typography variant="subtitle2">
                      Report {reportState.reference} received
                    </Typography>
                    <Typography variant="body2">
                      Call 112 if anyone is in immediate danger.
                      {session
                        ? " Saved to your reports below."
                        : " Sign in to keep a copy of your reports."}
                    </Typography>
                  </Alert>
                ) : null}
                <Button
                  type="submit"
                  variant="contained"
                  color="error"
                  disabled={reportState.status === "loading"}
                  startIcon={
                    reportState.status === "loading" ? (
                      <Loader2 size={18} className="spin-icon" />
                    ) : (
                      <Siren size={18} />
                    )
                  }
                >
                  {reportState.status === "loading"
                    ? reportState.message
                    : "Send report"}
                </Button>
              </Stack>
            </Paper>
          </Reveal>

          {session ? (
            <Reveal delay={80} className="citizen-subsection">
              <SavedReports reports={savedReports} />
            </Reveal>
          ) : null}
        </div>
      </div>
    </>
  );
}

export default ReportPage;
