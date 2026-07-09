import { type ChangeEvent, useEffect, useState } from "react";
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
import type { SelectChangeEvent } from "@mui/material/Select";
import { HandCoins, HeartHandshake, Loader2, RefreshCw } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AidCatalogRecord,
  AidCatalogListResponse,
  DonationAidRequestListResponse,
  DonationAidRequestRecord,
  DonationAidRequestStatus,
  CreateDonationAidRequestRequest,
  CreateDonorRequest,
  DonorListResponse,
  DonorRecord,
  DonorType,
  PledgeListResponse,
  PledgeRecord,
  PledgeStatus,
  UpdateDonationAidRequestRequest,
  UpdatePledgeRequest,
} from "@nadaa/shared-types";
import { DONATION_API_BASE } from "../../app/config";
import { authorityHeaders } from "../../app/session";
import { CommandSelect } from "./components";

const aidRequestStatuses: DonationAidRequestStatus[] = [
  "open",
  "partially_fulfilled",
  "fulfilled",
  "closed",
];

const donorTypes: DonorType[] = [
  "individual",
  "organization",
  "ngo",
  "government",
  "other",
];

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

const fallbackDonors: DonorRecord[] = [
  {
    id: "donor_fixture_001",
    reference: "DON-20260707-001",
    name: "Ghana Red Cross Society",
    type: "ngo",
    contactName: "Relief Coordinator",
    contactEmail: "relief@redcross.gh",
    contactPhone: "0302123456",
    region: "Greater Accra",
    district: "Accra Metropolitan",
    itemsOffered: ["food_parcel", "water_liter", "medical_kit"],
    status: "active",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

const fallbackPledges: PledgeRecord[] = [];

export type DonationLoadState = "loading" | "ready" | "fallback" | "error";

export interface DonationPanelProps {
  loadState?: DonationLoadState;
  feedback?: string;
  onLoadStateChange?: (state: DonationLoadState) => void;
  onFeedbackChange?: (message: string) => void;
}

interface AidRequestFormState {
  title: string;
  description: string;
  category: string;
  itemCode: string;
  quantityNeeded: string;
  unit: string;
  priority: string;
  locationLabel: string;
  region: string;
  district: string;
  beneficiaryCount: string;
}

interface DonorFormState {
  name: string;
  type: DonorType;
  contactName: string;
  contactEmail: string;
  contactPhone: string;
  region: string;
  district: string;
  itemsOffered: string;
  monetaryPledgeGhs: string;
  notes: string;
}

interface PledgeFormState {
  aidRequestId: string;
  donorId: string;
  quantityPledged: string;
  deliveryNote: string;
}

interface AllocateFormState {
  pledgeId: string;
  quantity: string;
}

const buildDefaultAidRequestForm = (): AidRequestFormState => ({
  title: "",
  description: "",
  category: "food",
  itemCode: "",
  quantityNeeded: "",
  unit: "",
  priority: "medium",
  locationLabel: "",
  region: "Greater Accra",
  district: "",
  beneficiaryCount: "",
});

const buildDefaultDonorForm = (): DonorFormState => ({
  name: "",
  type: "individual",
  contactName: "",
  contactEmail: "",
  contactPhone: "",
  region: "Greater Accra",
  district: "",
  itemsOffered: "",
  monetaryPledgeGhs: "",
  notes: "",
});

const buildDefaultPledgeForm = (): PledgeFormState => ({
  aidRequestId: "",
  donorId: "",
  quantityPledged: "",
  deliveryNote: "",
});

const buildDefaultAllocateForm = (): AllocateFormState => ({
  pledgeId: "",
  quantity: "",
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

function statusColor(status: string) {
  switch (status) {
    case "open":
      return "warning";
    case "partially_fulfilled":
      return "info";
    case "fulfilled":
      return "success";
    case "closed":
      return "default";
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

export function DonationPanel({
  loadState = "fallback",
  feedback = "",
  onLoadStateChange,
  onFeedbackChange,
}: DonationPanelProps) {
  const [catalog, setCatalog] = useState<AidCatalogRecord[]>(fallbackCatalog);
  const [aidRequests, setAidRequests] =
    useState<DonationAidRequestRecord[]>(fallbackAidRequests);
  const [donors, setDonors] = useState<DonorRecord[]>(fallbackDonors);
  const [pledges, setPledges] = useState<PledgeRecord[]>(fallbackPledges);
  const [busy, setBusy] = useState(false);
  const [localLoadState, setLocalLoadState] =
    useState<DonationLoadState>(loadState);
  const [localFeedback, setLocalFeedback] = useState(feedback);

  const [aidForm, setAidForm] = useState<AidRequestFormState>(
    buildDefaultAidRequestForm(),
  );
  const [donorForm, setDonorForm] = useState<DonorFormState>(
    buildDefaultDonorForm(),
  );
  const [pledgeForms, setPledgeForms] = useState<
    Record<string, PledgeFormState>
  >({});
  const [allocateForms, setAllocateForms] = useState<
    Record<string, AllocateFormState>
  >({});

  useEffect(() => {
    setLocalLoadState(loadState);
  }, [loadState]);

  useEffect(() => {
    setLocalFeedback(feedback);
  }, [feedback]);

  const setLoadState = (state: DonationLoadState) => {
    setLocalLoadState(state);
    onLoadStateChange?.(state);
  };

  const setFeedback = (message: string) => {
    setLocalFeedback(message);
    onFeedbackChange?.(message);
  };

  const refresh = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading donation and aid coordination data");

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

    try {
      const donorsResponse = await fetch(`${DONATION_API_BASE}/donors`, {
        headers: authorityHeaders(),
        signal,
      });
      if (donorsResponse.ok) {
        const payload = (await donorsResponse.json()) as DonorListResponse;
        setDonors(payload.donors.length ? payload.donors : fallbackDonors);
      } else if (donorsResponse.status === 401) {
        setDonors(fallbackDonors);
      }
    } catch {
      setDonors(fallbackDonors);
    }

    try {
      const pledgesResponse = await fetch(`${DONATION_API_BASE}/pledges`, {
        headers: authorityHeaders(),
        signal,
      });
      if (pledgesResponse.ok) {
        const payload = (await pledgesResponse.json()) as PledgeListResponse;
        setPledges(payload.pledges);
      } else if (pledgesResponse.status === 401) {
        setPledges(fallbackPledges);
      }
    } catch {
      setPledges(fallbackPledges);
    }

    if (catalogOk && requestsOk) {
      setLoadState("ready");
      setFeedback("Donation coordination API connected.");
    } else {
      setLoadState("fallback");
      setFeedback(
        "Donation API unavailable. Showing seeded catalog and request fixtures.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refresh(controller.signal);
    return () => controller.abort();
  }, []);

  const selectedCatalogItem = catalog.find(
    (item) => item.code === aidForm.itemCode,
  );

  const updateAidForm =
    (key: keyof AidRequestFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      const value = event.target.value;
      setAidForm((current) => {
        if (key === "itemCode") {
          const item = catalog.find(
            (catalogItem) => catalogItem.code === value,
          );
          return {
            ...current,
            itemCode: value,
            unit: item?.defaultUnit ?? current.unit,
            category: item?.category ?? current.category,
          };
        }
        return { ...current, [key]: value };
      });
    };

  const updateDonorForm =
    (key: keyof DonorFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      setDonorForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const createAidRequest = async () => {
    const quantityNeeded = Number(aidForm.quantityNeeded);
    const beneficiaryCount = aidForm.beneficiaryCount
      ? Number(aidForm.beneficiaryCount)
      : 0;
    if (
      !aidForm.title.trim() ||
      !aidForm.itemCode ||
      !Number.isFinite(quantityNeeded) ||
      quantityNeeded <= 0
    ) {
      setFeedback("Title, item, and a positive quantity are required.");
      return;
    }

    const payload: CreateDonationAidRequestRequest = {
      title: aidForm.title.trim(),
      description: aidForm.description.trim(),
      category: aidForm.category,
      itemCode: aidForm.itemCode,
      quantityNeeded,
      unit:
        aidForm.unit.trim() || (selectedCatalogItem?.defaultUnit ?? "units"),
      priority:
        (aidForm.priority as DonationAidRequestRecord["priority"]) || "medium",
      locationLabel: aidForm.locationLabel.trim(),
      region: aidForm.region.trim(),
      district: aidForm.district.trim(),
      beneficiaryCount,
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(`${DONATION_API_BASE}/aid-requests`, {
        method: "POST",
        headers: authorityHeaders(),
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const request = (await response.json()) as DonationAidRequestRecord;
      setAidRequests((current) => [request, ...current]);
      setAidForm(buildDefaultAidRequestForm());
      setLoadState("ready");
      setFeedback(`Aid request ${request.reference} created.`);
    } catch {
      setFeedback(
        "Aid request creation needs the donation-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const createDonor = async () => {
    if (!donorForm.name.trim() || !donorForm.type) {
      setFeedback("Donor name and type are required.");
      return;
    }

    const monetaryPledgeGhs = donorForm.monetaryPledgeGhs
      ? Number(donorForm.monetaryPledgeGhs)
      : undefined;

    const payload: CreateDonorRequest = {
      name: donorForm.name.trim(),
      type: donorForm.type,
      contactName: donorForm.contactName.trim() || undefined,
      contactEmail: donorForm.contactEmail.trim() || undefined,
      contactPhone: donorForm.contactPhone.trim() || undefined,
      region: donorForm.region.trim() || undefined,
      district: donorForm.district.trim() || undefined,
      itemsOffered: donorForm.itemsOffered
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      monetaryPledgeGhs:
        monetaryPledgeGhs && Number.isFinite(monetaryPledgeGhs)
          ? monetaryPledgeGhs
          : undefined,
      notes: donorForm.notes.trim() || undefined,
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(`${DONATION_API_BASE}/donors`, {
        method: "POST",
        headers: authorityHeaders(),
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const donor = (await response.json()) as DonorRecord;
      setDonors((current) => [donor, ...current]);
      setDonorForm(buildDefaultDonorForm());
      setLoadState("ready");
      setFeedback(`Donor ${donor.reference} registered.`);
    } catch {
      setFeedback(
        "Donor registration needs the donation-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const updateDonationAidRequestStatus = async (
    request: DonationAidRequestRecord,
    status: DonationAidRequestStatus,
  ) => {
    setBusy(true);
    setFeedback("");
    try {
      const payload: UpdateDonationAidRequestRequest = { status };
      const response = await fetch(
        `${DONATION_API_BASE}/aid-requests/${request.id}`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const updated = (await response.json()) as DonationAidRequestRecord;
      setAidRequests((current) =>
        current.map((item) => (item.id === updated.id ? updated : item)),
      );
      setFeedback(
        `${updated.reference} marked ${statusLabel(updated.status)}.`,
      );
    } catch {
      setFeedback(
        "Aid request update needs the donation-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const deleteAidRequest = async (request: DonationAidRequestRecord) => {
    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DONATION_API_BASE}/aid-requests/${request.id}`,
        {
          method: "DELETE",
          headers: authorityHeaders(),
        },
      );
      if (!response.ok && response.status !== 404) {
        throw new Error(`donation API returned ${response.status}`);
      }
      setAidRequests((current) =>
        current.filter((item) => item.id !== request.id),
      );
      setFeedback(`${request.reference} removed from the active list.`);
    } catch {
      setAidRequests((current) =>
        current.filter((item) => item.id !== request.id),
      );
      setFeedback(
        "Removed locally. The donation-service does not yet persist deletions.",
      );
    } finally {
      setBusy(false);
    }
  };

  const submitPledge = async (aidRequestId: string) => {
    const form = pledgeForms[aidRequestId] ?? buildDefaultPledgeForm();
    const quantityPledged = Number(form.quantityPledged);
    if (
      !form.donorId ||
      !Number.isFinite(quantityPledged) ||
      quantityPledged <= 0
    ) {
      setFeedback("Choose a donor and enter a positive pledged quantity.");
      return;
    }

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DONATION_API_BASE}/aid-requests/${aidRequestId}/pledges`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify({
            donorId: form.donorId,
            quantityPledged,
            deliveryNote: form.deliveryNote.trim(),
          }),
        },
      );
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const pledge = (await response.json()) as PledgeRecord;
      setPledges((current) => [pledge, ...current]);
      setPledgeForms((current) => ({
        ...current,
        [aidRequestId]: buildDefaultPledgeForm(),
      }));
      setAidRequests((current) =>
        current.map((item) =>
          item.id === aidRequestId
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
      setFeedback(`Pledge ${pledge.reference} recorded.`);
    } catch {
      setFeedback(
        "Pledge creation needs the donation-service API, a registered donor, and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const updatePledge = async (pledge: PledgeRecord, status: PledgeStatus) => {
    setBusy(true);
    setFeedback("");
    try {
      const payload: UpdatePledgeRequest = { status };
      const response = await fetch(
        `${DONATION_API_BASE}/pledges/${pledge.id}`,
        {
          method: "PATCH",
          headers: authorityHeaders(),
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const updated = (await response.json()) as PledgeRecord;
      setPledges((current) =>
        current.map((item) => (item.id === updated.id ? updated : item)),
      );
      setFeedback(`${updated.reference} marked ${status}.`);
    } catch {
      setFeedback(
        "Pledge update needs the donation-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const allocatePledge = async (pledge: PledgeRecord) => {
    const form = allocateForms[pledge.id] ?? buildDefaultAllocateForm();
    const quantity = Number(form.quantity);
    if (
      !Number.isFinite(quantity) ||
      quantity <= 0 ||
      quantity > pledge.quantityPledged
    ) {
      setFeedback(
        `Allocate a positive quantity up to ${pledge.quantityPledged}.`,
      );
      return;
    }

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${DONATION_API_BASE}/aid-requests/${pledge.aidRequestId}/allocate`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify({ pledgeId: pledge.id, quantity }),
        },
      );
      if (!response.ok) {
        throw new Error(`donation API returned ${response.status}`);
      }
      const updated = (await response.json()) as PledgeRecord;
      setPledges((current) =>
        current.map((item) => (item.id === updated.id ? updated : item)),
      );
      setAllocateForms((current) => ({
        ...current,
        [pledge.id]: buildDefaultAllocateForm(),
      }));
      setFeedback(`${updated.reference} marked delivered (${quantity}).`);
    } catch {
      setFeedback(
        "Pledge allocation needs the donation-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const activeAidRequests = aidRequests.filter(
    (request) => request.status !== "closed",
  );

  return (
    <Paper className="surface donation-panel">
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
            <Typography variant="h6">Donation and aid</Typography>
            <Typography variant="caption" color="text.secondary">
              Catalog, requests, donors, and pledge allocation
            </Typography>
          </Box>
        </Stack>
        <Button
          type="button"
          variant="outlined"
          size="small"
          startIcon={
            localLoadState === "loading" ? (
              <Loader2 size={16} className="spin-icon" />
            ) : (
              <RefreshCw size={16} />
            )
          }
          onClick={() => void refresh()}
          disabled={localLoadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>

      {localFeedback ? (
        <Alert
          severity={
            localLoadState === "ready"
              ? "success"
              : localLoadState === "loading"
                ? "info"
                : "warning"
          }
          className="feed-alert"
        >
          {localFeedback}
        </Alert>
      ) : null}

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Create aid request</Typography>
        <Grid container spacing={1}>
          <Grid size={{ xs: 12 }}>
            <TextField
              label="Title"
              size="small"
              fullWidth
              value={aidForm.title}
              onChange={updateAidForm("title")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <CommandSelect
              label="Item"
              value={aidForm.itemCode}
              onChange={updateAidForm("itemCode")}
            >
              <MenuItem value="">Select catalog item</MenuItem>
              {catalog.map((item) => (
                <MenuItem value={item.code} key={item.code}>
                  {item.name} ({item.defaultUnit})
                </MenuItem>
              ))}
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Unit"
              size="small"
              fullWidth
              value={aidForm.unit}
              onChange={updateAidForm("unit")}
              disabled={Boolean(selectedCatalogItem)}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <CommandSelect
              label="Priority"
              value={aidForm.priority}
              onChange={updateAidForm("priority")}
            >
              <MenuItem value="low">Low</MenuItem>
              <MenuItem value="medium">Medium</MenuItem>
              <MenuItem value="high">High</MenuItem>
              <MenuItem value="critical">Critical</MenuItem>
            </CommandSelect>
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="Quantity needed"
              size="small"
              fullWidth
              value={aidForm.quantityNeeded}
              onChange={updateAidForm("quantityNeeded")}
              inputProps={{ inputMode: "numeric" }}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <TextField
              label="Beneficiaries"
              size="small"
              fullWidth
              value={aidForm.beneficiaryCount}
              onChange={updateAidForm("beneficiaryCount")}
              inputProps={{ inputMode: "numeric" }}
            />
          </Grid>
          <Grid size={{ xs: 12 }}>
            <TextField
              label="Location label"
              size="small"
              fullWidth
              value={aidForm.locationLabel}
              onChange={updateAidForm("locationLabel")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Region"
              size="small"
              fullWidth
              value={aidForm.region}
              onChange={updateAidForm("region")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="District"
              size="small"
              fullWidth
              value={aidForm.district}
              onChange={updateAidForm("district")}
            />
          </Grid>
        </Grid>
        <Button
          type="button"
          variant="contained"
          startIcon={<HandCoins size={17} />}
          onClick={() => void createAidRequest()}
          disabled={busy}
        >
          Publish aid request
        </Button>
      </Stack>

      <Stack spacing={1.5}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          gap={1}
        >
          <Typography variant="subtitle2">Aid requests</Typography>
          <Chip
            size="small"
            label={localLoadState === "ready" ? "Live" : "Fixture"}
            color={localLoadState === "ready" ? "success" : "warning"}
          />
        </Stack>
        {activeAidRequests.length ? (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Reference</TableCell>
                <TableCell>Item</TableCell>
                <TableCell>Priority</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Fulfilled</TableCell>
                <TableCell>Beneficiaries</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {activeAidRequests.map((request) => (
                <TableRow key={request.id}>
                  <TableCell>
                    <Typography variant="subtitle2">
                      {request.reference}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {request.district}
                    </Typography>
                  </TableCell>
                  <TableCell>{request.title}</TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      label={request.priority}
                      color={priorityColor(request.priority)}
                    />
                  </TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      label={statusLabel(request.status)}
                      color={statusColor(request.status)}
                    />
                  </TableCell>
                  <TableCell>
                    {request.quantityFulfilled}/{request.quantityNeeded}{" "}
                    {request.unit}
                  </TableCell>
                  <TableCell>{request.beneficiaryCount ?? 0}</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5} flexWrap="wrap">
                      {aidRequestStatuses.map((status) => (
                        <Button
                          key={status}
                          size="small"
                          variant="outlined"
                          disabled={busy || request.status === status}
                          onClick={() =>
                            void updateDonationAidRequestStatus(request, status)
                          }
                        >
                          {status === "partially_fulfilled"
                            ? "Partial"
                            : statusLabel(status)}
                        </Button>
                      ))}
                      <Button
                        size="small"
                        color="error"
                        disabled={busy}
                        onClick={() => void deleteAidRequest(request)}
                      >
                        Delete
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Alert severity="info">No active aid requests.</Alert>
        )}
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Register donor</Typography>
        <Grid container spacing={1}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Donor name"
              size="small"
              fullWidth
              value={donorForm.name}
              onChange={updateDonorForm("name")}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <CommandSelect
              label="Type"
              value={donorForm.type}
              onChange={updateDonorForm("type")}
            >
              {donorTypes.map((type) => (
                <MenuItem value={type} key={type}>
                  {type.charAt(0).toUpperCase() + type.slice(1)}
                </MenuItem>
              ))}
            </CommandSelect>
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
              label="Contact phone"
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
          startIcon={<HeartHandshake size={17} />}
          onClick={() => void createDonor()}
          disabled={busy}
        >
          Register donor
        </Button>
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Pledges</Typography>
        {pledges.length ? (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Reference</TableCell>
                <TableCell>Donor</TableCell>
                <TableCell>Quantity</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Allocate</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {pledges.map((pledge) => (
                <TableRow key={pledge.id}>
                  <TableCell>{pledge.reference}</TableCell>
                  <TableCell>{pledge.donorName}</TableCell>
                  <TableCell>
                    {pledge.quantityDelivered}/{pledge.quantityPledged}
                  </TableCell>
                  <TableCell>
                    <Chip
                      size="small"
                      label={pledge.status}
                      color={
                        pledge.status === "delivered" ? "success" : "default"
                      }
                    />
                  </TableCell>
                  <TableCell>
                    {pledge.status !== "delivered" ? (
                      <Stack direction="row" spacing={0.5}>
                        <TextField
                          size="small"
                          label="Qty"
                          value={allocateForms[pledge.id]?.quantity ?? ""}
                          onChange={(event) =>
                            setAllocateForms((current) => ({
                              ...current,
                              [pledge.id]: {
                                ...(current[pledge.id] ??
                                  buildDefaultAllocateForm()),
                                quantity: event.target.value,
                              },
                            }))
                          }
                          inputProps={{ inputMode: "numeric" }}
                          sx={{ width: 80 }}
                        />
                        <Button
                          size="small"
                          variant="outlined"
                          disabled={busy}
                          onClick={() => void allocatePledge(pledge)}
                        >
                          Deliver
                        </Button>
                      </Stack>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5} flexWrap="wrap">
                      {pledge.status === "pledged" ? (
                        <Button
                          size="small"
                          variant="outlined"
                          disabled={busy}
                          onClick={() => void updatePledge(pledge, "cancelled")}
                        >
                          Cancel
                        </Button>
                      ) : null}
                      {pledge.status === "cancelled" ? (
                        <Button
                          size="small"
                          variant="outlined"
                          disabled={busy}
                          onClick={() => void updatePledge(pledge, "pledged")}
                        >
                          Reopen
                        </Button>
                      ) : null}
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Alert severity="info">No pledges recorded yet.</Alert>
        )}
      </Stack>

      <Stack spacing={1.5}>
        <Typography variant="subtitle2">Record pledge</Typography>
        {activeAidRequests.length ? (
          activeAidRequests.slice(0, 3).map((request) => (
            <Paper
              variant="outlined"
              className="pledge-row"
              key={request.id}
              sx={{ p: 1.25 }}
            >
              <Stack spacing={1}>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                >
                  <Typography variant="subtitle2">{request.title}</Typography>
                  <Chip
                    size="small"
                    label={`${request.quantityFulfilled}/${request.quantityNeeded} ${request.unit}`}
                  />
                </Stack>
                <Grid container spacing={1}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <CommandSelect
                      label="Donor"
                      value={pledgeForms[request.id]?.donorId ?? ""}
                      onChange={(event) =>
                        setPledgeForms((current) => ({
                          ...current,
                          [request.id]: {
                            ...(current[request.id] ??
                              buildDefaultPledgeForm()),
                            donorId: event.target.value,
                          },
                        }))
                      }
                    >
                      <MenuItem value="">Select donor</MenuItem>
                      {donors.map((donor) => (
                        <MenuItem value={donor.id} key={donor.id}>
                          {donor.name}
                        </MenuItem>
                      ))}
                    </CommandSelect>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      label="Quantity pledged"
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
                </Grid>
                <Button
                  size="small"
                  variant="contained"
                  disabled={busy}
                  onClick={() => void submitPledge(request.id)}
                >
                  Record pledge
                </Button>
              </Stack>
            </Paper>
          ))
        ) : (
          <Alert severity="info">
            Create an aid request before recording pledges.
          </Alert>
        )}
      </Stack>
    </Paper>
  );
}

export default DonationPanel;
