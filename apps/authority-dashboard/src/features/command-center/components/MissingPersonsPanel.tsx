import { type ChangeEvent, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { HeartHandshake, Loader2, RefreshCw, ShieldCheck } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CloseMissingPersonRequest,
  MissingPersonAuditEntry,
  MissingPersonAuditResponse,
  MissingPersonClosureType,
  MissingPersonListResponse,
  MissingPersonRecord,
  MissingPersonReviewDecision,
  ReviewMissingPersonRequest,
} from "@nadaa/shared-types";
import { MISSING_PERSON_API_BASE } from "@/app/config";
import { authorityHeaders } from "@/app/session";

type LoadState = "loading" | "ready" | "fallback" | "error";

interface ReviewFormState {
  decision: MissingPersonReviewDecision;
  publicSummary: string;
  reviewNotes: string;
}

interface CloseFormState {
  closureType: MissingPersonClosureType;
  closureNotes: string;
  reunitedWithFamily: boolean;
}

const fixtureTime = new Date().toISOString();

const fallbackRecords: MissingPersonRecord[] = [
  {
    id: "missing_001",
    reference: "MP-20260707-001",
    personName: "Kojo Mensah",
    age: 12,
    gender: "male",
    description:
      "Last seen wearing a blue school shirt and black shorts near the shelter registration desk.",
    photoUrl: "https://example.test/photos/kojo-mensah.jpg",
    lastSeenAt: fixtureTime,
    lastSeenLocation: {
      label: "Accra Metro Assembly Shelter",
      region: "Greater Accra",
      district: "Accra Metropolitan",
      lat: 5.56,
      lng: -0.2,
    },
    relatedIncidentId: "inc_accra_flood_0241",
    status: "active",
    publicSummary:
      "Child separated during shelter registration. Contact 112 with credible sightings.",
    updatedAt: fixtureTime,
    reporter: {
      name: "Ama Mensah",
      phone: "+233200000111",
      relationship: "mother",
      consentToContact: true,
      consentToPublicShare: true,
    },
    reviewStatus: "approved",
    publicVisibility: "public",
    reviewNotes: "Guardian consent verified.",
    createdBy: "public",
    createdAt: fixtureTime,
    reviewedBy: "usr_seed_reviewer",
    reviewedAt: fixtureTime,
  },
  {
    id: "missing_002",
    reference: "MP-20260707-002",
    personName: "Efua Boateng",
    age: 68,
    gender: "female",
    description: "Older adult reported missing after evacuation.",
    lastSeenAt: fixtureTime,
    lastSeenLocation: {
      label: "Osu Community Hall",
      region: "Greater Accra",
      district: "Korle Klottey",
    },
    status: "pending_review",
    updatedAt: fixtureTime,
    reporter: {
      name: "Kweku Boateng",
      phone: "+233200000222",
      relationship: "son",
      consentToContact: true,
      consentToPublicShare: false,
    },
    reviewStatus: "pending",
    publicVisibility: "private",
    createdBy: "public",
    createdAt: fixtureTime,
  },
];

const reviewDecisions: MissingPersonReviewDecision[] = [
  "approve_public",
  "approve_private",
  "reject",
];

const closureTypes: MissingPersonClosureType[] = [
  "reunited",
  "located_safe",
  "duplicate",
  "withdrawn",
  "deceased",
  "other",
];

function label(value: string) {
  return value
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export default function MissingPersonsPanel() {
  const [records, setRecords] =
    useState<MissingPersonRecord[]>(fallbackRecords);
  const [selectedId, setSelectedId] = useState(fallbackRecords[0]?.id ?? "");
  const [auditEntries, setAuditEntries] = useState<MissingPersonAuditEntry[]>(
    [],
  );
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("Loading missing-person queue");
  const [busy, setBusy] = useState(false);
  const [reviewForm, setReviewForm] = useState<ReviewFormState>({
    decision: "approve_private",
    publicSummary: "",
    reviewNotes: "",
  });
  const [closeForm, setCloseForm] = useState<CloseFormState>({
    closureType: "reunited",
    closureNotes: "",
    reunitedWithFamily: true,
  });

  const selectedRecord = useMemo(
    () => records.find((record) => record.id === selectedId) ?? records[0],
    [records, selectedId],
  );

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading missing-person queue");
    try {
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/authority/missing-persons`,
        {
          headers: authorityHeaders(),
          signal,
        },
      );
      if (!response.ok) {
        throw new Error(`missing-person API returned ${response.status}`);
      }
      const payload = (await response.json()) as MissingPersonListResponse;
      setRecords(payload.records.length ? payload.records : []);
      setSelectedId((current) => current || payload.records[0]?.id || "");
      setLoadState("ready");
      setFeedback("Missing-person queue is live.");
    } catch {
      setRecords(fallbackRecords);
      setSelectedId(fallbackRecords[0]?.id ?? "");
      setLoadState("fallback");
      setFeedback("Using fixture missing-person cases until the service runs.");
    }
  };

  const refreshAudit = async (recordId: string) => {
    try {
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/authority/missing-persons/${recordId}/audit`,
        { headers: authorityHeaders() },
      );
      if (!response.ok) {
        throw new Error(`audit API returned ${response.status}`);
      }
      const payload = (await response.json()) as MissingPersonAuditResponse;
      setAuditEntries(payload.entries);
    } catch {
      setAuditEntries([]);
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (selectedId && loadState === "ready") {
      void refreshAudit(selectedId);
    }
  }, [selectedId, loadState]);

  const updateReview =
    (key: keyof ReviewFormState) => (event: ChangeEvent<HTMLInputElement>) => {
      setReviewForm((current) => ({
        ...current,
        [key]: event.target.value,
      }));
    };

  const updateClose =
    (key: keyof CloseFormState) => (event: ChangeEvent<HTMLInputElement>) => {
      setCloseForm((current) => ({
        ...current,
        [key]:
          event.target.type === "checkbox"
            ? event.target.checked
            : event.target.value,
      }));
    };

  const applyRecord = (record: MissingPersonRecord) => {
    setRecords((current) =>
      current.map((item) => (item.id === record.id ? record : item)),
    );
    setSelectedId(record.id);
  };

  const reviewSelected = async () => {
    if (!selectedRecord) return;
    const request: ReviewMissingPersonRequest = {
      decision: reviewForm.decision,
      publicSummary: reviewForm.publicSummary.trim(),
      reviewNotes: reviewForm.reviewNotes.trim(),
    };
    setBusy(true);
    try {
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/authority/missing-persons/${selectedRecord.id}/review`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(`review API returned ${response.status}`);
      }
      const record = (await response.json()) as MissingPersonRecord;
      applyRecord(record);
      setFeedback(`${record.reference} review saved.`);
      setLoadState("ready");
      await refreshAudit(record.id);
    } catch {
      setFeedback("Review needs the missing-person-service running.");
      setLoadState("error");
    } finally {
      setBusy(false);
    }
  };

  const closeSelected = async () => {
    if (!selectedRecord) return;
    const request: CloseMissingPersonRequest = {
      closureType: closeForm.closureType,
      closureNotes: closeForm.closureNotes.trim(),
      reunitedWithFamily: closeForm.reunitedWithFamily,
    };
    setBusy(true);
    try {
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/authority/missing-persons/${selectedRecord.id}/close`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(request),
        },
      );
      if (!response.ok) {
        throw new Error(`close API returned ${response.status}`);
      }
      const record = (await response.json()) as MissingPersonRecord;
      applyRecord(record);
      setFeedback(`${record.reference} closure saved.`);
      setLoadState("ready");
      await refreshAudit(record.id);
    } catch {
      setFeedback("Closure needs the missing-person-service running.");
      setLoadState("error");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Paper className="surface">
      <Stack spacing={2}>
        <Stack
          direction="row"
          spacing={1}
          alignItems="center"
          className="section-heading"
        >
          <HeartHandshake size={21} color={nadaaBrand.colors.gold} />
          <Typography variant="h6">Missing persons</Typography>
        </Stack>

        <Alert severity={loadState === "error" ? "error" : "info"}>
          {feedback}
        </Alert>

        <Box className="table-wrap">
          <Table size="small" aria-label="Missing persons queue">
            <TableHead>
              <TableRow>
                <TableCell>Reference</TableCell>
                <TableCell>Person</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Visibility</TableCell>
                <TableCell>Last seen</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {records.map((record) => (
                <TableRow
                  key={record.id}
                  hover
                  selected={record.id === selectedRecord?.id}
                  onClick={() => setSelectedId(record.id)}
                >
                  <TableCell>{record.reference}</TableCell>
                  <TableCell>{record.personName}</TableCell>
                  <TableCell>
                    <Chip size="small" label={label(record.status)} />
                  </TableCell>
                  <TableCell>{label(record.publicVisibility)}</TableCell>
                  <TableCell>{record.lastSeenLocation.district}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Box>

        {selectedRecord ? (
          <Grid container spacing={2}>
            <Grid size={{ xs: 12, md: 5 }}>
              <Box className="support-card">
                <Typography variant="subtitle1">
                  {selectedRecord.personName}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Last seen {formatDate(selectedRecord.lastSeenAt)} at{" "}
                  {selectedRecord.lastSeenLocation.label},{" "}
                  {selectedRecord.lastSeenLocation.district}
                </Typography>
                <Typography variant="body2">
                  {selectedRecord.description}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Chip
                    size="small"
                    label={`Reporter: ${selectedRecord.reporter.relationship}`}
                  />
                  <Chip
                    size="small"
                    label={
                      selectedRecord.reporter.consentToPublicShare
                        ? "Public consent"
                        : "Private only"
                    }
                  />
                </Stack>
              </Box>
            </Grid>
            <Grid size={{ xs: 12, md: 7 }}>
              <Grid container spacing={1.5}>
                <Grid size={{ xs: 12, md: 4 }}>
                  <TextField
                    label="Review decision"
                    select
                    fullWidth
                    value={reviewForm.decision}
                    onChange={updateReview("decision")}
                  >
                    {reviewDecisions.map((decision) => (
                      <MenuItem key={decision} value={decision}>
                        {label(decision)}
                      </MenuItem>
                    ))}
                  </TextField>
                </Grid>
                <Grid size={{ xs: 12, md: 8 }}>
                  <TextField
                    label="Public summary"
                    value={reviewForm.publicSummary}
                    onChange={updateReview("publicSummary")}
                    fullWidth
                  />
                </Grid>
                <Grid size={{ xs: 12 }}>
                  <TextField
                    label="Review notes"
                    value={reviewForm.reviewNotes}
                    onChange={updateReview("reviewNotes")}
                    minRows={2}
                    multiline
                    fullWidth
                  />
                </Grid>
                <Grid size={{ xs: 12 }}>
                  <Button
                    variant="contained"
                    startIcon={
                      busy ? (
                        <Loader2 className="spin-icon" />
                      ) : (
                        <ShieldCheck size={17} />
                      )
                    }
                    onClick={() => void reviewSelected()}
                    disabled={busy}
                  >
                    Save review
                  </Button>
                </Grid>
                <Grid size={{ xs: 12, md: 4 }}>
                  <TextField
                    label="Closure type"
                    select
                    fullWidth
                    value={closeForm.closureType}
                    onChange={updateClose("closureType")}
                  >
                    {closureTypes.map((closureType) => (
                      <MenuItem key={closureType} value={closureType}>
                        {label(closureType)}
                      </MenuItem>
                    ))}
                  </TextField>
                </Grid>
                <Grid size={{ xs: 12, md: 8 }}>
                  <TextField
                    label="Closure notes"
                    value={closeForm.closureNotes}
                    onChange={updateClose("closureNotes")}
                    fullWidth
                  />
                </Grid>
                <Grid size={{ xs: 12 }}>
                  <Button
                    variant="outlined"
                    startIcon={<RefreshCw size={17} />}
                    onClick={() => void closeSelected()}
                    disabled={busy}
                  >
                    Save closure
                  </Button>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        ) : null}

        {auditEntries.length ? (
          <Stack spacing={0.75}>
            <Typography variant="caption" color="text.secondary">
              Audit trail
            </Typography>
            {auditEntries.slice(-3).map((entry) => (
              <Typography
                variant="caption"
                color="text.secondary"
                key={entry.id}
              >
                {formatDate(entry.createdAt)} · {entry.action} ·{" "}
                {entry.actorUserId || "system"}
              </Typography>
            ))}
          </Stack>
        ) : null}
      </Stack>
    </Paper>
  );
}
