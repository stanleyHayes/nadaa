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
  TextField,
  Typography,
} from "@mui/material";
import {
  HandHeart,
  HeartHandshake,
  Loader2,
  Package,
  PackageOpen,
  RefreshCw,
} from "lucide-react";
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
import { DONATION_API_BASE } from "@/app/config";
import { useCitizenSession } from "../session";
import { DataTable, type DataTableColumn, type DataTableFilter } from "./DataTable";
import { DetailDialog, type DetailField } from "./DetailDialog";
import { EmptyState } from "./EmptyState";
import { FormDialogButton } from "./FormDialogButton";

type LoadState = "loading" | "ready" | "error";

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
  const { session, requestSignIn } = useCitizenSession();
  const [catalog, setCatalog] = useState<AidCatalogRecord[]>([]);
  const [aidRequests, setAidRequests] = useState<DonationAidRequestRecord[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [feedback, setFeedback] = useState("Loading aid opportunities");
  const [donorForm, setDonorForm] = useState<DonorRegistrationForm>(
    buildDefaultDonorForm(),
  );
  const [pledgeForms, setPledgeForms] = useState<Record<string, PledgeForm>>(
    {},
  );
  const [selectedRequestId, setSelectedRequestId] = useState("");
  const [busy, setBusy] = useState(false);
  const [registeredDonor, setRegisteredDonor] = useState<DonorRecord | null>(
    null,
  );
  // Dialog-scoped errors: validation/network failures inside a modal must
  // surface within that modal, not on the panel Alert hidden behind its backdrop.
  const [donorError, setDonorError] = useState("");
  const [pledgeError, setPledgeError] = useState("");
  const [detailRequest, setDetailRequest] =
    useState<DonationAidRequestRecord | null>(null);

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
        setCatalog(payload.items);
        catalogOk = true;
      }
    } catch {
      // Network failure surfaces as the error state decided below.
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
        setAidRequests(payload.requests);
        requestsOk = true;
      }
    } catch {
      // Network failure surfaces as the error state decided below.
    }

    if (signal?.aborted) {
      return;
    }

    if (catalogOk && requestsOk) {
      setLoadState("ready");
      setFeedback("Aid opportunities updated.");
    } else {
      setLoadState("error");
      setFeedback("Couldn't reach the donation service. Try refreshing.");
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

  const updatePledgeField = (
    requestId: string,
    key: keyof PledgeForm,
    value: string,
  ) => {
    setPledgeForms((current) => ({
      ...current,
      [requestId]: {
        ...(current[requestId] ?? buildDefaultPledgeForm()),
        [key]: value,
      },
    }));
  };

  const registerDonor = async (): Promise<boolean> => {
    if (!session) {
      requestSignIn();
      return false;
    }
    if (!donorForm.name.trim()) {
      setDonorError("Please enter your name or organization to register.");
      return false;
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
    setDonorError("");
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
      return true;
    } catch {
      setDonorError(
        "Donor registration could not reach the donation service. Try again later.",
      );
      return false;
    } finally {
      setBusy(false);
    }
  };

  const submitPledge = async (
    aidRequest: DonationAidRequestRecord,
  ): Promise<boolean> => {
    if (!session) {
      requestSignIn();
      return false;
    }
    const form = pledgeForms[aidRequest.id] ?? buildDefaultPledgeForm();
    const quantityPledged = Number(form.quantityPledged);
    if (
      !form.donorName.trim() ||
      !Number.isFinite(quantityPledged) ||
      quantityPledged <= 0
    ) {
      setPledgeError("Enter your name and a positive quantity to pledge.");
      return false;
    }

    const payload: CreatePledgeRequest = {
      donorName: form.donorName.trim(),
      quantityPledged,
      contactEmail: form.contactEmail.trim() || undefined,
      contactPhone: form.contactPhone.trim() || undefined,
      donorId: registeredDonor?.id,
    };

    setBusy(true);
    setPledgeError("");
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
      return true;
    } catch {
      setPledgeError(
        "Pledge could not be submitted. The donation service may be offline.",
      );
      return false;
    } finally {
      setBusy(false);
    }
  };

  const openRequests = aidRequests.filter(
    (request) => request.status !== "closed" && request.status !== "fulfilled",
  );

  // Keep a valid pledge target even after a request is fulfilled/removed.
  const activeRequestId =
    openRequests.find((request) => request.id === selectedRequestId)?.id ??
    openRequests[0]?.id ??
    "";
  const activeRequest = aidRequests.find(
    (request) => request.id === activeRequestId,
  );

  const priorityOptions = useMemo(
    () => Array.from(new Set(aidRequests.map((request) => request.priority))),
    [aidRequests],
  );
  const regionOptions = useMemo(
    () =>
      Array.from(
        new Set(
          aidRequests
            .map((request) => request.region)
            .filter((region): region is string => Boolean(region)),
        ),
      ).sort(),
    [aidRequests],
  );
  const statusOptions = useMemo(
    () =>
      Array.from(
        new Set(aidRequests.map((request) => statusLabel(request.status))),
      ),
    [aidRequests],
  );
  const catalogCategoryOptions = useMemo(
    () => Array.from(new Set(catalog.map((item) => item.category))).sort(),
    [catalog],
  );

  const requestColumns: DataTableColumn<DonationAidRequestRecord>[] = [
    {
      key: "title",
      label: "Item / need",
      render: (request) => (
        <Box>
          <Typography variant="body2" sx={{ fontWeight: 700 }}>
            {request.title}
          </Typography>
          <Typography variant="caption" sx={{
            color: "text.secondary"
          }}>
            {request.category}
          </Typography>
        </Box>
      ),
    },
    {
      key: "region",
      label: "Region / area",
      render: (request) => (
        <Box>
          <Typography variant="body2">{request.region ?? "—"}</Typography>
          {request.district ? (
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              {request.district}
            </Typography>
          ) : null}
        </Box>
      ),
    },
    {
      key: "quantityNeeded",
      label: "Quantity needed",
      align: "right",
      render: (request) =>
        `${request.quantityFulfilled}/${request.quantityNeeded} ${request.unit}`,
    },
    {
      key: "priority",
      label: "Urgency",
      render: (request) => (
        <Chip
          size="small"
          label={request.priority}
          color={priorityColor(request.priority)}
        />
      ),
    },
    {
      key: "status",
      label: "Status",
      render: (request) => (
        <Chip
          size="small"
          variant="outlined"
          label={statusLabel(request.status)}
        />
      ),
    },
  ];

  const requestFilters: DataTableFilter<DonationAidRequestRecord>[] = [
    {
      key: "priority",
      label: "Urgency",
      options: priorityOptions,
      valueOf: (request) => request.priority,
    },
    {
      key: "region",
      label: "Region",
      options: regionOptions,
      valueOf: (request) => request.region ?? "",
    },
    {
      key: "status",
      label: "Status",
      options: statusOptions,
      valueOf: (request) => statusLabel(request.status),
    },
  ];

  const catalogColumns: DataTableColumn<AidCatalogRecord>[] = [
    {
      key: "name",
      label: "Item",
      render: (item) => (
        <Box>
          <Typography variant="body2" sx={{ fontWeight: 700 }}>
            {item.name}
          </Typography>
          <Typography variant="caption" sx={{
            color: "text.secondary"
          }}>
            {item.code}
          </Typography>
        </Box>
      ),
    },
    { key: "category", label: "Category" },
    { key: "defaultUnit", label: "Unit" },
    { key: "priorityScore", label: "Priority", align: "right" },
  ];

  const catalogFilters: DataTableFilter<AidCatalogRecord>[] = [
    {
      key: "category",
      label: "Category",
      options: catalogCategoryOptions,
      valueOf: (item) => item.category,
    },
  ];

  const donorFormFields = (close: () => void) => (
    <Stack spacing={1.5} sx={{ pt: 1 }}>
      {donorError ? <Alert severity="error">{donorError}</Alert> : null}
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
        variant="contained"
        disabled={busy}
        onClick={async () => {
          const ok = await registerDonor();
          if (ok) {
            close();
          }
        }}
      >
        Register as donor
      </Button>
    </Stack>
  );

  const pledgeFormFields = (close: () => void) => (
    <Stack spacing={1.5} sx={{ pt: 1 }}>
      {pledgeError ? <Alert severity="error">{pledgeError}</Alert> : null}
      <TextField
        select
        label="Aid request"
        size="small"
        fullWidth
        value={activeRequestId}
        onChange={(event) => setSelectedRequestId(event.target.value)}
        helperText={
          openRequests.length
            ? "Choose the aid request you want to support."
            : "No open aid requests are available right now."
        }
      >
        {openRequests.map((request) => (
          <MenuItem key={request.id} value={request.id}>
            {request.title}
          </MenuItem>
        ))}
      </TextField>
      <Grid container spacing={1}>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label="Your name"
            size="small"
            fullWidth
            value={pledgeForms[activeRequestId]?.donorName ?? ""}
            onChange={(event) =>
              updatePledgeField(activeRequestId, "donorName", event.target.value)
            }
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label={`Quantity (${activeRequest?.unit ?? "units"})`}
            size="small"
            fullWidth
            value={pledgeForms[activeRequestId]?.quantityPledged ?? ""}
            onChange={(event) =>
              updatePledgeField(
                activeRequestId,
                "quantityPledged",
                event.target.value,
              )
            }
            slotProps={{
              htmlInput: { inputMode: "numeric" }
            }}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label="Email"
            size="small"
            fullWidth
            value={pledgeForms[activeRequestId]?.contactEmail ?? ""}
            onChange={(event) =>
              updatePledgeField(
                activeRequestId,
                "contactEmail",
                event.target.value,
              )
            }
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <TextField
            label="Phone"
            size="small"
            fullWidth
            value={pledgeForms[activeRequestId]?.contactPhone ?? ""}
            onChange={(event) =>
              updatePledgeField(
                activeRequestId,
                "contactPhone",
                event.target.value,
              )
            }
          />
        </Grid>
      </Grid>
      <Button
        type="button"
        variant="contained"
        disabled={busy || !activeRequest}
        onClick={async () => {
          if (!activeRequest) {
            return;
          }
          const ok = await submitPledge(activeRequest);
          if (ok) {
            close();
          }
        }}
      >
        Pledge now
      </Button>
    </Stack>
  );

  const detailFields: DetailField[] = detailRequest
    ? [
        { label: "Item / need", value: detailRequest.title, full: true },
        { label: "Item code", value: detailRequest.itemCode },
        {
          label: "Region / area",
          value:
            [detailRequest.region, detailRequest.district]
              .filter(Boolean)
              .join(" · ") || "—",
        },
        {
          label: "Quantity needed",
          value: `${detailRequest.quantityFulfilled}/${detailRequest.quantityNeeded} ${detailRequest.unit}`,
        },
        {
          label: "Urgency",
          value: (
            <Chip
              size="small"
              label={detailRequest.priority}
              color={priorityColor(detailRequest.priority)}
            />
          ),
        },
        {
          label: "Status",
          value: (
            <Chip
              size="small"
              variant="outlined"
              label={statusLabel(detailRequest.status)}
            />
          ),
        },
        {
          label: "Beneficiaries",
          value: detailRequest.beneficiaryCount
            ? detailRequest.beneficiaryCount.toLocaleString()
            : "—",
        },
        {
          label: "Requested",
          value: new Date(detailRequest.createdAt).toLocaleDateString(),
        },
        { label: "Reference", value: detailRequest.reference },
        {
          label: "Description / notes",
          value: detailRequest.description ?? "—",
          full: true,
        },
      ]
    : [];

  return (
    <Paper className="surface donor-portal">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          alignItems: { xs: "stretch", sm: "center" }
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <HeartHandshake size={21} color={nadaaBrand.colors.green} />
          <Box>
            <Typography variant="h6">Donor portal</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Browse aid requests, pledge support, and register as a donor
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
                : "error"
          }
          className="warning-alert"
        >
          {feedback}
        </Alert>
      ) : null}
      <Stack spacing={2}>
        <Box>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1.5}
            sx={{
              justifyContent: "space-between",
              alignItems: { xs: "stretch", sm: "center" },
              mb: 1.5
            }}>
            <Typography variant="subtitle2">Aid requests</Typography>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
              <FormDialogButton
                label="Register as donor"
                dialogTitle="Register as a donor"
                icon={HeartHandshake}
                color="secondary"
              >
                {(close) => donorFormFields(close)}
              </FormDialogButton>
              <FormDialogButton
                label="Pledge support"
                dialogTitle="Pledge support"
                icon={HandHeart}
                color="primary"
              >
                {(close) => pledgeFormFields(close)}
              </FormDialogButton>
            </Stack>
          </Stack>
          <DataTable
            rows={aidRequests}
            columns={requestColumns}
            getRowKey={(request) => request.id}
            searchOf={(request) =>
              `${request.title} ${request.itemCode} ${request.region ?? ""}`
            }
            searchPlaceholder="Search aid requests"
            filters={requestFilters}
            emptyMessage="No aid requests match your filters."
            emptyState={
              <EmptyState
                icon={PackageOpen}
                tone="green"
                title="No aid requests"
                description="No open aid requests match your search right now."
              />
            }
            onRowClick={setDetailRequest}
          />
        </Box>

        <Box>
          <Typography variant="subtitle2" gutterBottom>
            Aid catalog
          </Typography>
          <DataTable
            rows={catalog}
            columns={catalogColumns}
            getRowKey={(item) => item.id}
            searchOf={(item) => `${item.name} ${item.category}`}
            searchPlaceholder="Search catalog"
            filters={catalogFilters}
            pageSize={5}
            emptyMessage="No catalog items match your filters."
            emptyState={
              <EmptyState
                icon={Package}
                tone="green"
                title="No catalog items"
                description="Nothing matches your search."
              />
            }
          />
        </Box>
      </Stack>
      <DetailDialog
        open={Boolean(detailRequest)}
        onClose={() => setDetailRequest(null)}
        title={detailRequest?.title ?? ""}
        subtitle={detailRequest?.category}
        fields={detailFields}
      />
    </Paper>
  );
}

export default DonorPortal;
