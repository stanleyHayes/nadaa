import { ChangeEvent, FormEvent, useCallback, useEffect, useState } from "react";
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
import {
  BookmarkCheck,
  FileText,
  ImagePlus,
  Loader2,
  LocateFixed,
  LogIn,
  Megaphone,
  ShieldCheck,
  Siren,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
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
import {
  extractAPIError,
  formatDateTime,
  formatFileSize,
  hazardLabel,
  initiateMediaUploads,
} from "../utils";
import { useCitizenSession, type SavedReport } from "../session";
import {
  DataTable,
  type DataTableColumn,
  EmptyState,
  FormDialogButton,
  DetailDialog,
  type DetailField,
  PageHeader,
  Reveal,
} from "../components";
import { PageBanner } from "../components/PageBanner";

/** Human-friendly urgency label for the saved-reports table. */
const urgencyLabel = (value: string) =>
  urgencyOptions.find((option) => option.value === value)?.label ?? value;

/** The three-step "what happens after you report" explainer, shown publicly. */
const reportSteps: { icon: LucideIcon; title: string; text: string }[] = [
  {
    icon: Siren,
    title: "1. You report",
    text: "Send what you're seeing with photos and a location. Reports go privately to NADMO — they are not published.",
  },
  {
    icon: ShieldCheck,
    title: "2. NADMO verifies",
    text: "Trained officers review, de-duplicate, and verify every report before any response is coordinated.",
  },
  {
    icon: Megaphone,
    title: "3. Approved public alert",
    text: "Only human-approved alerts go public. A verified report can trigger an official alert others will see.",
  },
];

/** Columns for a signed-in citizen's own saved reports. */
const reportColumns: DataTableColumn<SavedReport>[] = [
  {
    key: "reference",
    label: "Reference",
    render: (report) => (
      <Typography variant="body2" sx={{ fontWeight: 700 }}>
        {report.reference}
      </Typography>
    ),
  },
  {
    key: "hazard",
    label: "Hazard type",
    render: (report) => hazardLabel(report.hazard as HazardType),
  },
  {
    key: "urgency",
    label: "Urgency",
    render: (report) => urgencyLabel(report.urgency),
  },
  {
    key: "at",
    label: "Submitted",
    render: (report) => formatDateTime(report.at),
  },
  {
    key: "status",
    label: "Status",
    align: "right",
    render: (report) => (
      <Chip
        size="small"
        label={report.priorityReview ? "Priority review" : "Submitted"}
        color={report.priorityReview ? "warning" : "success"}
      />
    ),
  },
];

/**
 * Runs `onOpen` once each time it mounts. `FormDialogButton` mounts the form
 * only while its dialog is open, so rendering this inside the form resets the
 * report's transient state on every open — no stale success/error banner ever
 * lingers over a freshly opened form. Renders nothing.
 */
function ResetOnOpen({ onOpen }: { onOpen: () => void }) {
  useEffect(onOpen, [onOpen]);
  return null;
}

/**
 * Incident reporting (route `/report`). Info-first: everyone sees how reporting
 * works, the form lives behind an auth-gated `FormDialogButton`, and signed-in
 * citizens get their own saved reports as a table. Migrated from the legacy
 * `#report` form.
 */
export function ReportPage() {
  const { session, savedReports, saveReport, requestSignIn } =
    useCitizenSession();
  const [reportForm, setReportForm] = useState<ReportForm>(initialReportForm);
  const [reportState, setReportState] = useState<ReportState>({
    status: "idle",
  });
  const [reportErrors, setReportErrors] = useState<
    Partial<Record<keyof ReportForm, string>>
  >({});
  // The saved report opened in the read-only detail dialog (list/detail split).
  const [detailReport, setDetailReport] = useState<SavedReport | null>(null);

  // Clear any transient success/error banner (and stale field errors) so the
  // form starts clean each time the dialog reopens. Stable so `ResetOnOpen`
  // fires it once per open, not on every keystroke.
  const resetReportState = useCallback(() => {
    setReportState({ status: "idle" });
    setReportErrors({});
  }, []);

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

  const submitReport = async (
    event: FormEvent<HTMLFormElement>,
    close: () => void,
  ) => {
    event.preventDefault();

    // Submissions require a signed-in citizen — no anonymous reports.
    const currentSession = session;
    if (!currentSession) {
      requestSignIn();
      return;
    }

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
        anonymous: false,
        contactPermission: reportForm.contactPermission,
        accessibilityNeeds: reportForm.accessibilityNeeds.trim() || undefined,
        media: mediaIds,
        reporter: {
          userId: "usr_demo_citizen",
          phone: reportForm.contactPermission
            ? currentSession.phone
            : undefined,
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
      // Dismiss the dialog — the report now shows in "Your reports" below.
      close();
    } catch (error) {
      setReportState({
        status: "error",
        message:
          error instanceof Error ? error.message : "Could not send report.",
      });
    }
  };

  // The full report form, rendered inside the auth-gated dialog. `close`
  // dismisses the dialog once the report is sent.
  const renderReportForm = (close: () => void) => (
    <Stack
      component="form"
      spacing={1.5}
      onSubmit={(event) => void submitReport(event, close)}
      noValidate
    >
      <ResetOnOpen onOpen={resetReportState} />
      <FormControl fullWidth error={Boolean(reportErrors.hazard)}>
        <InputLabel id="report-hazard-label">Hazard type</InputLabel>
        <Select
          id="report-hazard"
          labelId="report-hazard-label"
          value={reportForm.hazard}
          label="Hazard type"
          onChange={(event) => {
            clearReportError("hazard");
            updateReportForm("hazard", event.target.value as HazardType);
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
        <Typography>Allow responders to contact me</Typography>
        <Switch
          checked={reportForm.contactPermission}
          onChange={(event) =>
            updateReportForm("contactPermission", event.target.checked)
          }
        />
      </Stack>
      <Alert severity="info" className="warning-alert">
        You're reporting as {session?.name}. NADAA uses your report and location
        to route emergency response, detect duplicates, and coordinate verified
        authority actions. Turning off contact means responders cannot call you
        back about this report.
      </Alert>
      {reportState.status === "error" ? (
        <Alert severity="error" className="warning-alert">
          {reportState.message}
        </Alert>
      ) : null}
      {/* The success confirmation is surfaced at the panel level (below "Your
          reports") after the dialog closes — see the panel-level Alert. */}
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
  );

  return (
    <>
      <PageBanner
        eyebrow="Report an incident"
        subtitle="Tell NADMO what you're seeing — with photos and location, online or offline."
        title="Report a flood or hazard"
      />
      <div className="citizen-shell">
        <div className="citizen-section">
          {/* Public explainer + primary action. Visible to everyone. */}
          <Reveal delay={80}>
            <Paper className="surface report-surface" component="section">
              <PageHeader
                icon={Megaphone}
                title="How reporting works"
                subtitle="Anyone can browse alerts, risk, shelters and guides. Reporting an incident takes a quick sign-in so responders can verify and follow up."
                tone="gold"
                action={
                  <FormDialogButton
                    label="Report an incident"
                    dialogTitle="Report an incident"
                    icon={Siren}
                    color="primary"
                  >
                    {(close) => renderReportForm(close)}
                  </FormDialogButton>
                }
              />
              <Stack spacing={1.5}>
                {reportSteps.map(({ icon: StepIcon, title, text }) => (
                  <Stack
                    key={title}
                    direction="row"
                    spacing={1.5}
                    alignItems="flex-start"
                  >
                    <StepIcon size={20} aria-hidden="true" />
                    <div>
                      <Typography variant="subtitle2">{title}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {text}
                      </Typography>
                    </div>
                  </Stack>
                ))}
                <Alert severity="error" className="warning-alert">
                  Call 112 now if anyone is in immediate danger — do not wait for
                  this report to be reviewed.
                </Alert>
              </Stack>
            </Paper>
          </Reveal>

          {/* Signed-in citizens see their own saved reports; others are invited
              to sign in. There is no public incidents list. */}
          {session ? (
            <Reveal delay={80} className="citizen-subsection">
              <Stack spacing={1.5}>
                <PageHeader
                  icon={BookmarkCheck}
                  title="Your reports"
                  subtitle="Reports you've submitted on this device, so you can follow them up."
                  tone="green"
                  as="h3"
                />
                {/* Panel-level success confirmation, shown after the dialog
                    closes — the in-dialog Alert could never render because the
                    dialog unmounts its children on close. Cleared on reopen. */}
                {reportState.status === "success" ? (
                  <Alert
                    severity={reportState.priorityReview ? "warning" : "success"}
                    className="warning-alert"
                  >
                    <Typography variant="subtitle2">
                      Report {reportState.reference} received
                    </Typography>
                    <Typography variant="body2">
                      Call 112 if anyone is in immediate danger. Saved to your
                      reports.
                    </Typography>
                  </Alert>
                ) : null}
                <DataTable
                  rows={savedReports}
                  columns={reportColumns}
                  getRowKey={(report) => report.reference}
                  onRowClick={setDetailReport}
                  searchOf={(report) =>
                    `${report.reference} ${hazardLabel(
                      report.hazard as HazardType,
                    )} ${urgencyLabel(report.urgency)}`
                  }
                  searchPlaceholder="Search your reports"
                  emptyMessage="You haven't submitted any reports yet."
                  emptyState={
                    <EmptyState
                      icon={FileText}
                      tone="navy"
                      title="No reports yet"
                      description="When you report an incident, a copy is saved here so you can track its status."
                    />
                  }
                />
              </Stack>
            </Reveal>
          ) : (
            <Reveal delay={80} className="citizen-subsection">
              <Paper className="surface" component="section">
                <PageHeader
                  icon={BookmarkCheck}
                  title="Track your reports"
                  subtitle="Sign in with your name and phone to submit reports and follow the ones you've sent."
                  tone="green"
                  as="h3"
                />
                <Button
                  onClick={requestSignIn}
                  variant="contained"
                  startIcon={<LogIn size={18} />}
                  sx={{ fontWeight: 800 }}
                >
                  Sign in to continue
                </Button>
              </Paper>
            </Reveal>
          )}
        </div>
      </div>

      {/* Light-detail half of the list/detail split: a report row opens this
          read-only dialog instead of expanding inline. */}
      <DetailDialog
        open={Boolean(detailReport)}
        onClose={() => setDetailReport(null)}
        title={detailReport ? `Report ${detailReport.reference}` : ""}
        subtitle={
          detailReport
            ? hazardLabel(detailReport.hazard as HazardType)
            : undefined
        }
        fields={
          detailReport
            ? ([
                { label: "Reference", value: detailReport.reference },
                {
                  label: "Hazard type",
                  value: hazardLabel(detailReport.hazard as HazardType),
                },
                {
                  label: "Urgency",
                  value: urgencyLabel(detailReport.urgency),
                },
                {
                  label: "Submitted",
                  value: formatDateTime(detailReport.at),
                },
                {
                  label: "Status",
                  value: (
                    <Chip
                      size="small"
                      label={
                        detailReport.priorityReview
                          ? "Priority review"
                          : "Submitted"
                      }
                      color={
                        detailReport.priorityReview ? "warning" : "success"
                      }
                    />
                  ),
                },
              ] satisfies DetailField[])
            : []
        }
      />
    </>
  );
}

export default ReportPage;
