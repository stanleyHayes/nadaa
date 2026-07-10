import { type ChangeEvent, type FormEvent, useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { HeartHandshake, Loader2, RefreshCw, Search } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CreateMissingPersonRequest,
  PublicMissingPersonListResponse,
  PublicMissingPersonRecord,
} from "@nadaa/shared-types";
import { MISSING_PERSON_API_BASE } from "@/app/config";
import { useCitizenSession } from "../session";

type LoadState = "loading" | "ready" | "fallback" | "error";

interface MissingPersonForm {
  personName: string;
  age: string;
  gender: string;
  description: string;
  photoUrl: string;
  lastSeenAt: string;
  locationLabel: string;
  region: string;
  district: string;
  latitude: string;
  longitude: string;
  relatedIncidentId: string;
  reporterName: string;
  reporterPhone: string;
  reporterEmail: string;
  reporterRelationship: string;
  consentToContact: boolean;
  consentToPublicShare: boolean;
}

const now = new Date();

const fallbackMissingPersons: PublicMissingPersonRecord[] = [
  {
    id: "missing_001",
    reference: "MP-20260707-001",
    personName: "Kojo Mensah",
    age: 12,
    gender: "male",
    description:
      "Last seen wearing a blue school shirt and black shorts near the shelter registration desk.",
    photoUrl: "https://example.test/photos/kojo-mensah.jpg",
    lastSeenAt: new Date(now.getTime() - 6 * 60 * 60 * 1000).toISOString(),
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
      "Child separated during shelter registration. Please contact authorities through 112 with credible sightings.",
    updatedAt: new Date(now.getTime() - 3 * 60 * 60 * 1000).toISOString(),
  },
];

const buildDefaultForm = (): MissingPersonForm => ({
  personName: "",
  age: "",
  gender: "unknown",
  description: "",
  photoUrl: "",
  lastSeenAt: new Date(Date.now() - 60 * 60 * 1000).toISOString().slice(0, 16),
  locationLabel: "",
  region: "Greater Accra",
  district: "",
  latitude: "",
  longitude: "",
  relatedIncidentId: "",
  reporterName: "",
  reporterPhone: "",
  reporterEmail: "",
  reporterRelationship: "",
  consentToContact: true,
  consentToPublicShare: false,
});

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export default function MissingPersonsPanel() {
  const { session, requestSignIn } = useCitizenSession();
  const [records, setRecords] = useState<PublicMissingPersonRecord[]>(
    fallbackMissingPersons,
  );
  const [query, setQuery] = useState("");
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("Loading approved public cases");
  const [form, setForm] = useState<MissingPersonForm>(buildDefaultForm());
  const [busy, setBusy] = useState(false);

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading approved public cases");
    try {
      const search = query.trim()
        ? `?q=${encodeURIComponent(query.trim())}`
        : "";
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/missing-persons${search}`,
        { signal },
      );
      if (!response.ok) {
        throw new Error(`missing-person API returned ${response.status}`);
      }
      const payload =
        (await response.json()) as PublicMissingPersonListResponse;
      setRecords(payload.records.length ? payload.records : []);
      setLoadState("ready");
      setFeedback(
        payload.records.length
          ? "Showing authority-approved public cases."
          : "No approved public cases match this search.",
      );
    } catch {
      setRecords(fallbackMissingPersons);
      setLoadState("fallback");
      setFeedback("Using public fixture cases until the service is available.");
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    return () => controller.abort();
  }, []);

  const updateForm =
    (key: keyof MissingPersonForm) =>
    (event: ChangeEvent<HTMLInputElement>) => {
      const value =
        event.target.type === "checkbox"
          ? event.target.checked
          : event.target.value;
      setForm((current) => ({ ...current, [key]: value }));
    };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!session) {
      requestSignIn();
      return;
    }
    if (
      !form.personName.trim() ||
      !form.description.trim() ||
      !form.locationLabel.trim() ||
      !form.district.trim() ||
      !form.reporterName.trim() ||
      !form.reporterPhone.trim() ||
      !form.reporterRelationship.trim()
    ) {
      setFeedback("Complete person, location, and reporter details first.");
      setLoadState("error");
      return;
    }
    if (!form.consentToContact) {
      setFeedback("Contact consent is required so authorities can follow up.");
      setLoadState("error");
      return;
    }

    const payload: CreateMissingPersonRequest = {
      personName: form.personName.trim(),
      age: form.age ? Number(form.age) : undefined,
      gender: form.gender,
      description: form.description.trim(),
      photoUrl: form.photoUrl.trim() || undefined,
      lastSeenAt: new Date(form.lastSeenAt).toISOString(),
      lastSeenLocation: {
        label: form.locationLabel.trim(),
        region: form.region.trim(),
        district: form.district.trim(),
        lat: form.latitude ? Number(form.latitude) : undefined,
        lng: form.longitude ? Number(form.longitude) : undefined,
      },
      relatedIncidentId: form.relatedIncidentId.trim() || undefined,
      reporter: {
        name: form.reporterName.trim(),
        phone: form.reporterPhone.trim(),
        email: form.reporterEmail.trim() || undefined,
        relationship: form.reporterRelationship.trim(),
        consentToContact: form.consentToContact,
        consentToPublicShare: form.consentToPublicShare,
      },
    };

    setBusy(true);
    try {
      const response = await fetch(
        `${MISSING_PERSON_API_BASE}/missing-persons`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`missing-person API returned ${response.status}`);
      }
      setForm(buildDefaultForm());
      setLoadState("ready");
      setFeedback(
        "Report received privately. Authorities will review before any public visibility.",
      );
    } catch {
      setLoadState("error");
      setFeedback(
        "Could not submit the report. Call 112 immediately if someone is at risk.",
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <Paper className="surface report-surface">
      <Stack spacing={2}>
        <Stack
          direction="row"
          spacing={1}
          alignItems="center"
          className="section-heading"
        >
          <HeartHandshake size={21} color={nadaaBrand.colors.gold} />
          <Typography variant="h6">Family reunification</Typography>
        </Stack>

        <Alert
          severity={loadState === "error" ? "error" : "info"}
          className="warning-alert"
        >
          {feedback}
        </Alert>

        <Box component="form" onSubmit={(event) => void submit(event)}>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 12, md: 6 }}>
              <TextField
                label="Missing person name"
                value={form.personName}
                onChange={updateForm("personName")}
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3 }}>
              <TextField
                label="Age"
                type="number"
                value={form.age}
                onChange={updateForm("age")}
                fullWidth
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3 }}>
              <TextField
                label="Gender"
                value={form.gender}
                onChange={updateForm("gender")}
                select
                fullWidth
              >
                <option value="unknown">Unknown</option>
                <option value="female">Female</option>
                <option value="male">Male</option>
                <option value="non_binary">Non-binary</option>
              </TextField>
            </Grid>
            <Grid size={{ xs: 12 }}>
              <TextField
                label="Description"
                value={form.description}
                onChange={updateForm("description")}
                minRows={3}
                multiline
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <TextField
                label="Last seen place"
                value={form.locationLabel}
                onChange={updateForm("locationLabel")}
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 12, md: 3 }}>
              <TextField
                label="District"
                value={form.district}
                onChange={updateForm("district")}
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 12, md: 3 }}>
              <TextField
                label="Last seen time"
                type="datetime-local"
                value={form.lastSeenAt}
                onChange={updateForm("lastSeenAt")}
                fullWidth
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3 }}>
              <TextField
                label="Latitude"
                value={form.latitude}
                onChange={updateForm("latitude")}
                fullWidth
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3 }}>
              <TextField
                label="Longitude"
                value={form.longitude}
                onChange={updateForm("longitude")}
                fullWidth
              />
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <TextField
                label="Photo URL"
                value={form.photoUrl}
                onChange={updateForm("photoUrl")}
                fullWidth
              />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <TextField
                label="Reporter name"
                value={form.reporterName}
                onChange={updateForm("reporterName")}
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <TextField
                label="Reporter phone"
                value={form.reporterPhone}
                onChange={updateForm("reporterPhone")}
                fullWidth
                required
              />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <TextField
                label="Relationship"
                value={form.reporterRelationship}
                onChange={updateForm("reporterRelationship")}
                fullWidth
                required
              />
            </Grid>
          </Grid>
          <Stack spacing={1.25} sx={{ mt: 1.5 }}>
            <Stack direction="row" spacing={1} alignItems="center">
              <Switch
                checked={form.consentToContact}
                onChange={updateForm("consentToContact")}
              />
              <Typography variant="body2">
                I consent to authority follow-up using reporter contact details.
              </Typography>
            </Stack>
            <Stack direction="row" spacing={1} alignItems="center">
              <Switch
                checked={form.consentToPublicShare}
                onChange={updateForm("consentToPublicShare")}
              />
              <Typography variant="body2">
                I consent to safe public sharing after authority review.
              </Typography>
            </Stack>
            <Button
              type="submit"
              variant="contained"
              startIcon={busy ? <Loader2 className="spin-icon" /> : <Search />}
              disabled={busy}
            >
              {busy ? "Submitting" : "Submit private report"}
            </Button>
          </Stack>
        </Box>

        <Stack direction="row" spacing={1}>
          <TextField
            label="Search approved cases"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            fullWidth
          />
          <Button
            variant="outlined"
            startIcon={<RefreshCw size={17} />}
            onClick={() => void refresh()}
          >
            Search
          </Button>
        </Stack>

        <Stack spacing={1.25}>
          {records.map((record) => (
            <Box className="support-card" key={record.id}>
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                justifyContent="space-between"
              >
                <Typography variant="subtitle1">{record.personName}</Typography>
                <Chip label={record.reference} size="small" />
              </Stack>
              <Typography variant="body2" color="text.secondary">
                Last seen {formatDate(record.lastSeenAt)} at{" "}
                {record.lastSeenLocation.label},{" "}
                {record.lastSeenLocation.district}
              </Typography>
              <Typography variant="body2">{record.publicSummary}</Typography>
            </Box>
          ))}
        </Stack>
      </Stack>
    </Paper>
  );
}
