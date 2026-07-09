import { type ChangeEvent, useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  Grid,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { HeartHandshake, Loader2, RefreshCw } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AidCatalogListResponse,
  AidCatalogRecord,
  DonationAidRequestListResponse,
  DonationAidRequestRecord,
  CreateDonorRequest,
  CreatePledgeRequest,
  DonorRecord,
} from "@nadaa/shared-types";
import { DONATION_API_BASE } from "../../app/config";

const fallbackCatalog: AidCatalogRecord[] = [
  {
    id: "catalog_001",
    code: "food_parcel",
    name: "Ready-to-eat food parcels",
    category: "food",
    defaultUnit: "parcels",
    priorityScore: 90,
  },
  {
    id: "catalog_002",
    code: "water_liter",
    name: "Clean drinking water",
    category: "water",
    defaultUnit: "liters",
    priorityScore: 95,
  },
  {
    id: "catalog_003",
    code: "medical_kit",
    name: "Emergency medical kit",
    category: "medical",
    defaultUnit: "kits",
    priorityScore: 100,
  },
  {
    id: "catalog_004",
    code: "shelter_kit",
    name: "Family shelter kit",
    category: "shelter",
    defaultUnit: "kits",
    priorityScore: 85,
  },
  {
    id: "catalog_005",
    code: "hygiene_kit",
    name: "Hygiene and sanitation kit",
    category: "sanitation",
    defaultUnit: "kits",
    priorityScore: 80,
  },
];

const fallbackAidRequests: DonationAidRequestRecord[] = [
  {
    id: "request_001",
    reference: "AR-20260707-001",
    title: "Flood relief food parcels for Accra Metropolitan",
    description:
      "Ready-to-eat food parcels for households displaced by flooding in central Accra.",
    category: "food",
    itemCode: "food_parcel",
    quantityNeeded: 500,
    quantityFulfilled: 0,
    unit: "parcels",
    priority: "high",
    locationLabel: "Accra Metropolitan Assembly Hall",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    beneficiaryCount: 2500,
    status: "open",
    requestedBy: "seed",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: "request_002",
    reference: "AR-20260707-002",
    title: "Emergency medical supplies for Tema",
    description:
      "First-aid and emergency medical kits for flood-affected communities in Tema.",
    category: "medical",
    itemCode: "medical_kit",
    quantityNeeded: 200,
    quantityFulfilled: 0,
    unit: "kits",
    priority: "critical",
    locationLabel: "Tema General Hospital",
    region: "Greater Accra",
    district: "Tema Metropolitan",
    beneficiaryCount: 800,
    status: "open",
    requestedBy: "seed",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

type LoadState = "loading" | "ready" | "fallback" | "error";

interface DonorRegistrationForm {
  name: string;
  contactName: string;
  contactEmail: string;
  contactPhone: string;
  region: string;
  district: string;
}

interface PledgeForm {
  donorName: string;
  quantityPledged: string;
  contactEmail: string;
  contactPhone: string;
}

const buildDefaultDonorForm = (): DonorRegistrationForm => ({
  name: "",
  contactName: "",
  contactEmail: "",
  contactPhone: "",
  region: "Greater Accra",
  district: "",
});

const buildDefaultPledgeForm = (): PledgeForm => ({
  donorName: "",
  quantityPledged: "",
  contactEmail: "",
  contactPhone: "",
});

function priorityColor(priority: string) {
  switch (priority) {
    case "critical":
      return "error";
    case "high":
      return "warning";
    case "medium":
      return "info";
    default:
      return "default";
  }
}

function statusLabel(status: string) {
  return status
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function DonorPortal() {
  const [catalog, setCatalog] = useState<AidCatalogRecord[]>(fallbackCatalog);
  const [aidRequests, setAidRequests] =
    useState<DonationAidRequestRecord[]>(fallbackAidRequests);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("Loading aid opportunities");
  const [donorForm, setDonorForm] = useState<DonorRegistrationForm>(
    buildDefaultDonorForm(),
  );
  const [pledgeForms, setPledgeForms] = useState<Record<string, PledgeForm>>(
    {},
  );
  const [busy, setBusy] = useState(false);
  const [registeredDonor, setRegisteredDonor] = useState<DonorRecord | null>(
    null,
  );

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading aid opportunities");

    let catalogOk = false;
    let requestsOk = false;

    try {
      const catalogResponse = await fetch(`${DONATION_API_BASE}/aid-catalog`, {
        signal,
      });
      if (catalogResponse.ok) {
        const payload =
          (await catalogResponse.json()) as AidCatalogListResponse;
        setCatalog(payload.items.length ? payload.items : fallbackCatalog);
        catalogOk = true;
      }
    } catch {
      setCatalog(fallbackCatalog);
    }

    try {
      const requestsResponse = await fetch(
        `${DONATION_API_BASE}/aid-requests`,
        {
          signal,
        },
      );
      if (requestsResponse.ok) {
        const payload =
          (await requestsResponse.json()) as DonationAidRequestListResponse;
        setAidRequests(
          payload.requests.length ? payload.requests : fallbackAidRequests,
        );
        requestsOk = true;
      }
    } catch {
      setAidRequests(fallbackAidRequests);
    }

    if (catalogOk && requestsOk) {
      setLoadState("ready");
      setFeedback("Aid opportunities updated.");
    } else {
      setLoadState("fallback");
      setFeedback(
        "Donation service is offline. Showing seeded aid requests and catalog.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    return () => controller.abort();
  }, []);

  const updateDonorForm =
    (key: keyof DonorRegistrationForm) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setDonorForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const registerDonor = async () => {
    if (!donorForm.name.trim()) {
      setFeedback("Please enter your name or organization to register.");
      return;
    }

    const payload: CreateDonorRequest = {
      name: donorForm.name.trim(),
      type: "individual",
      contactName: donorForm.contactName.trim() || undefined,
      contactEmail: donorForm.contactEmail.trim() || undefined,
      contactPhone: donorForm.contactPhone.trim() || undefined,
      region: donorForm.region.trim() || undefined,
      district: donorForm.district.trim() || undefined,
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(`${DONATION_API_BASE}/donors`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const donor = (await response.json()) as DonorRecord;
      setRegisteredDonor(donor);
      setDonorForm(buildDefaultDonorForm());
      setFeedback(
        `Thank you, ${donor.name}. Your donor reference is ${donor.reference}.`,
      );
    } catch {
      setFeedback(
        "Donor registration could not reach the donation service. Try again later.",
      );
    } finally {
      setBusy(false);
    }
  };

  const submitPledge = async (aidRequest: DonationAidRequestRecord) => {
    const form = pledgeForms[aidRequest.id] ?? buildDefaultPledgeForm();
    const quantityPledged = Number(form.quantityPledged);
    if (
      !form.donorName.trim() ||
      !Number.isFinite(quantityPledged) ||
      quantityPledged <= 0
    ) {
      setFeedback("Enter your name and a positive quantity to pledge.");
      return;
    }

    const payload: CreatePledgeRequest = {
      donorName: form.donorName.trim(),
      quantityPledged,
      contactEmail: form.contactEmail.trim() || undefined,
      contactPhone: form.contactPhone.trim() || undefined,
      donorId: registeredDonor?.id,
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DONATION_API_BASE}/aid-requests/${aidRequest.id}/pledges`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const pledge = (await response.json()) as {
        id: string;
        reference: string;
        donorName: string;
        quantityPledged: number;
      };
      setPledgeForms((current) => ({
        ...current,
        [aidRequest.id]: buildDefaultPledgeForm(),
      }));
      setAidRequests((current) =>
        current.map((item) =>
          item.id === aidRequest.id
            ? {
                ...item,
                quantityFulfilled: item.quantityFulfilled + quantityPledged,
                status:
                  item.quantityFulfilled + quantityPledged >=
                  item.quantityNeeded
                    ? "fulfilled"
                    : "partially_fulfilled",
              }
            : item,
        ),
      );
      setFeedback(
        `Thank you ${pledge.donorName}. Pledge ${pledge.reference} recorded for ${quantityPledged} ${aidRequest.unit}.`,
      );
    } catch {
      setFeedback(
        "Pledge could not be submitted. The donation service may be offline.",
      );
    } finally {
      setBusy(false);
    }
  };

  const openRequests = aidRequests.filter(
    (request) => request.status !== "closed" && request.status !== "fulfilled",
  );

  return (
    <Paper className="surface donor-portal">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", sm: "center" }}
        className="section-heading"
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <HeartHandshake size={21} color={nadaaBrand.colors.green} />
          <Box>
            <Typography variant="h6">Donor portal</Typography>
            <Typography variant="caption" color="text.secondary">
              Pledge aid and register as a donor
            </Typography>
          </Box>
        </Stack>
        <Button
          type="button"
          variant="outlined"
          size="small"
          startIcon={
            loadState === "loading" ? (
              <Loader2 size={16} className="spin-icon" />
            ) : (
              <RefreshCw size={16} />
            )
          }
          onClick={() => void refresh()}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>

      {feedback ? (
        <Alert
          severity={
            loadState === "ready"
              ? "success"
              : loadState === "loading"
                ? "info"
                : "warning"
          }
          className="warning-alert"
        >
          {feedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Open aid requests</Typography>
        {openRequests.length ? (
          openRequests.map((request) => (
            <Paper
              variant="outlined"
              className="aid-request-card"
              key={request.id}
              sx={{ p: 1.5 }}
            >
              <Stack spacing={1.25}>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="flex-start"
                  gap={1}
                >
                  <Box>
                    <Typography variant="subtitle2">{request.title}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {request.district} · {request.locationLabel}
                    </Typography>
                  </Box>
                  <Chip
                    size="small"
                    label={request.priority}
                    color={priorityColor(request.priority)}
                  />
                </Stack>
                <Typography variant="body2">{request.description}</Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap">
                  <Chip
                    size="small"
                    variant="outlined"
                    label={`${request.quantityFulfilled}/${request.quantityNeeded} ${request.unit}`}
                  />
                  <Chip
                    size="small"
                    variant="outlined"
                    label={`${request.beneficiaryCount ?? 0} beneficiaries`}
                  />
                  <Chip
                    size="small"
                    variant="outlined"
                    label={statusLabel(request.status)}
                  />
                </Stack>

                <Divider />

                <Stack spacing={1}>
                  <Typography variant="subtitle2">Make a pledge</Typography>
                  <Grid container spacing={1}>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <TextField
                        label="Your name"
                        size="small"
                        fullWidth
                        value={pledgeForms[request.id]?.donorName ?? ""}
                        onChange={(event) =>
                          setPledgeForms((current) => ({
                            ...current,
                            [request.id]: {
                              ...(current[request.id] ??
                                buildDefaultPledgeForm()),
                              donorName: event.target.value,
                            },
                          }))
                        }
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <TextField
                        label={`Quantity (${request.unit})`}
                        size="small"
                        fullWidth
                        value={pledgeForms[request.id]?.quantityPledged ?? ""}
                        onChange={(event) =>
                          setPledgeForms((current) => ({
                            ...current,
                            [request.id]: {
                              ...(current[request.id] ??
                                buildDefaultPledgeForm()),
                              quantityPledged: event.target.value,
                            },
                          }))
                        }
                        inputProps={{ inputMode: "numeric" }}
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <TextField
                        label="Email"
                        size="small"
                        fullWidth
                        value={pledgeForms[request.id]?.contactEmail ?? ""}
                        onChange={(event) =>
                          setPledgeForms((current) => ({
                            ...current,
                            [request.id]: {
                              ...(current[request.id] ??
                                buildDefaultPledgeForm()),
                              contactEmail: event.target.value,
                            },
                          }))
                        }
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <TextField
                        label="Phone"
                        size="small"
                        fullWidth
                        value={pledgeForms[request.id]?.contactPhone ?? ""}
                        onChange={(event) =>
                          setPledgeForms((current) => ({
                            ...current,
                            [request.id]: {
                              ...(current[request.id] ??
                                buildDefaultPledgeForm()),
                              contactPhone: event.target.value,
                            },
                          }))
                        }
                      />
                    </Grid>
                  </Grid>
                  <Button
                    size="small"
                    variant="contained"
                    disabled={busy}
                    onClick={() => void submitPledge(request)}
                  >
                    Pledge now
                  </Button>
                </Stack>
              </Stack>
            </Paper>
          ))
        ) : (
          <Alert severity="info">No open aid requests at the moment.</Alert>
        )}
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Become a donor</Typography>
        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Name or organization"
              size="small"
              fullWidth
              value={donorForm.name}
              onChange={updateDonorForm("name")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Contact name"
              size="small"
              fullWidth
              value={donorForm.contactName}
              onChange={updateDonorForm("contactName")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Email"
              size="small"
              fullWidth
              value={donorForm.contactEmail}
              onChange={updateDonorForm("contactEmail")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Phone"
              size="small"
              fullWidth
              value={donorForm.contactPhone}
              onChange={updateDonorForm("contactPhone")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Region"
              size="small"
              fullWidth
              value={donorForm.region}
              onChange={updateDonorForm("region")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="District"
              size="small"
              fullWidth
              value={donorForm.district}
              onChange={updateDonorForm("district")}
            />
          </Grid>
        </Grid>
        <Button
          type="button"
          variant="outlined"
          disabled={busy}
          onClick={() => void registerDonor()}
        >
          Register as donor
        </Button>
      </Stack>
    </Paper>
  );
}

export default DonorPortal;
