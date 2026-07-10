import { FormEvent, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  ClipboardList,
  FileCheck,
  Loader2,
  MapPin,
  ShieldCheck,
} from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  CreateDamageClaimRequest,
  DamageClaimRecord,
  DamageType,
} from "@nadaa/shared-types";
import { DAMAGE_CLAIM_API_BASE } from "@/app/config";
import { useCitizenSession } from "../session";
import { extractAPIError } from "../utils";
import { FormDialogButton } from "./FormDialogButton";

const damageTypeOptions: { label: string; value: DamageType }[] = [
  { label: "Structural", value: "structural" },
  { label: "Flood", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Vehicle", value: "vehicle" },
  { label: "Other", value: "other" },
];

const claimChecklist = [
  "The incident reference, if your damage is linked to a reported event.",
  "Reporter contact details so the verification team can follow up.",
  "A short description of the damage and an estimated loss amount.",
  "The property location or GPS coordinates, and links to any damage photos.",
];

const initialForm: DamageClaimForm = {
  incidentId: "",
  reporterName: "",
  reporterPhone: "",
  reporterEmail: "",
  damageType: "flood",
  damageDescription: "",
  estimatedLossAmount: "",
  lat: "",
  lng: "",
  address: "",
  photoUrls: "",
  privacyConsent: false,
};

type DamageClaimForm = {
  incidentId: string;
  reporterName: string;
  reporterPhone: string;
  reporterEmail: string;
  damageType: DamageType;
  damageDescription: string;
  estimatedLossAmount: string;
  lat: string;
  lng: string;
  address: string;
  photoUrls: string;
  privacyConsent: boolean;
};

type DamageClaimState =
  | { status: "idle" }
  | { status: "loading"; message: string }
  | {
      status: "success";
      reference: string;
      verificationStatus: DamageClaimRecord["verificationStatus"];
    }
  | { status: "error"; message: string };

function DamageClaim() {
  const { session, requestSignIn } = useCitizenSession();
  const [form, setForm] = useState<DamageClaimForm>(initialForm);
  const [state, setState] = useState<DamageClaimState>({ status: "idle" });

  const updateForm = <Key extends keyof DamageClaimForm>(
    key: Key,
    value: DamageClaimForm[Key],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const useCurrentLocation = () => {
    if (!navigator.geolocation) {
      setState({
        status: "error",
        message: "Location is not available on this device.",
      });
      return;
    }

    setState({ status: "loading", message: "Getting location" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setForm((current) => ({
          ...current,
          lat: position.coords.latitude.toFixed(6),
          lng: position.coords.longitude.toFixed(6),
        }));
        setState({ status: "idle" });
      },
      () => {
        setState({
          status: "error",
          message: "Location permission was not granted.",
        });
      },
      { enableHighAccuracy: true, timeout: 10000 },
    );
  };

  const submitClaim =
    (close: () => void) => async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();

      if (!session) {
        requestSignIn();
        return;
      }

      const lat = Number(form.lat);
      const lng = Number(form.lng);

      if (
        !Number.isFinite(lat) ||
        lat < -90 ||
        lat > 90 ||
        !Number.isFinite(lng) ||
        lng < -180 ||
        lng > 180
      ) {
        setState({
          status: "error",
          message: "Enter valid coordinates for the damaged property.",
        });
        return;
      }

      if (form.reporterName.trim().length === 0) {
        setState({
          status: "error",
          message: "Enter the reporter's full name.",
        });
        return;
      }

      if (form.reporterPhone.trim().length === 0) {
        setState({
          status: "error",
          message:
            "Enter a phone number so the verification team can follow up.",
        });
        return;
      }

      if (form.damageDescription.trim().length < 5) {
        setState({
          status: "error",
          message: "Add a short description of the damage.",
        });
        return;
      }

      if (
        form.estimatedLossAmount.trim().length === 0 ||
        Number.isNaN(Number(form.estimatedLossAmount))
      ) {
        setState({
          status: "error",
          message: "Enter a valid estimated loss amount.",
        });
        return;
      }

      if (!form.privacyConsent) {
        setState({
          status: "error",
          message: "Agree to the privacy statement to submit the claim.",
        });
        return;
      }

      if (!navigator.onLine) {
        setState({
          status: "error",
          message:
            "You appear to be offline. Keep this claim open and try again when the connection returns.",
        });
        return;
      }

      setState({ status: "loading", message: "Submitting damage claim" });

      const photoUrls = form.photoUrls
        .split(",")
        .map((url) => url.trim())
        .filter((url) => url.length > 0);

      const payload: CreateDamageClaimRequest = {
        incidentId: form.incidentId.trim() || undefined,
        reporter: {
          name: form.reporterName.trim(),
          phone: form.reporterPhone.trim(),
          email: form.reporterEmail.trim() || undefined,
        },
        damageType: form.damageType,
        damageDescription: form.damageDescription.trim(),
        estimatedLossAmount: form.estimatedLossAmount.trim(),
        damagePhotos: photoUrls.length > 0 ? photoUrls : undefined,
        location: {
          lat,
          lng,
          address: form.address.trim() || undefined,
        },
        privacyConsent: true,
      };

      try {
        const response = await fetch(`${DAMAGE_CLAIM_API_BASE}/claims`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });

        if (!response.ok) {
          throw new Error(await extractAPIError(response));
        }

        const claim = (await response.json()) as DamageClaimRecord;
        setState({
          status: "success",
          reference: claim.reference,
          verificationStatus: claim.verificationStatus,
        });
        setForm(initialForm);
        close();
      } catch (error) {
        setState({
          status: "error",
          message:
            error instanceof Error
              ? error.message
              : "Could not submit damage claim.",
        });
      }
    };

  return (
    <Paper className="surface damage-claim-surface">
      <Stack spacing={2}>
        <Stack
          direction="row"
          spacing={1}
          alignItems="center"
          className="section-heading"
        >
          <FileCheck size={21} color={nadaaBrand.colors.green} />
          <Box>
            <Typography variant="h6">Damage claim</Typography>
            <Typography variant="caption" color="text.secondary">
              Insurance and relief support
            </Typography>
          </Box>
        </Stack>

        <Typography variant="body2" color="text.secondary">
          A damage claim is a private record of property or asset loss after a
          disaster. Residents and businesses affected by a flood, fire, or other
          reported incident can file one. Verified claims can be exported for
          insurance and NADMO relief processes, and each claim stays private to
          you and the verification team.
        </Typography>

        <Alert severity="info" className="warning-alert">
          <Typography variant="subtitle2" gutterBottom>
            Have this ready before you file
          </Typography>
          <Box component="ul" sx={{ m: 0, pl: 2.5 }}>
            {claimChecklist.map((item) => (
              <li key={item}>
                <Typography variant="body2">{item}</Typography>
              </li>
            ))}
          </Box>
        </Alert>

        {state.status === "success" ? (
          <Alert
            severity="success"
            className="warning-alert"
            icon={<ShieldCheck size={20} />}
          >
            <Typography variant="subtitle2">
              Claim {state.reference} received
            </Typography>
            <Typography variant="body2">
              Verification status: {state.verificationStatus}. Keep this
              reference for follow-up.
            </Typography>
          </Alert>
        ) : null}

        <Box>
          <FormDialogButton
            label="File a damage claim"
            dialogTitle="File a damage claim"
            icon={ClipboardList}
            color="primary"
          >
            {(close) => (
              <Stack
                component="form"
                spacing={1.5}
                onSubmit={submitClaim(close)}
                noValidate
              >
                <TextField
                  label="Incident ID (optional)"
                  value={form.incidentId}
                  onChange={(event) =>
                    updateForm("incidentId", event.target.value)
                  }
                  helperText="Link this claim to a reported incident if you have the reference."
                  inputProps={{ maxLength: 100 }}
                />

                <TextField
                  label="Reporter full name"
                  value={form.reporterName}
                  onChange={(event) =>
                    updateForm("reporterName", event.target.value)
                  }
                  required
                  inputProps={{ maxLength: 200 }}
                />

                <Grid container spacing={1.25}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      label="Phone number"
                      value={form.reporterPhone}
                      onChange={(event) =>
                        updateForm("reporterPhone", event.target.value)
                      }
                      fullWidth
                      required
                      inputProps={{ maxLength: 50 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      label="Email address"
                      value={form.reporterEmail}
                      onChange={(event) =>
                        updateForm("reporterEmail", event.target.value)
                      }
                      fullWidth
                      type="email"
                      inputProps={{ maxLength: 200 }}
                    />
                  </Grid>
                </Grid>

                <FormControl fullWidth>
                  <InputLabel id="damage-type-label">Damage type</InputLabel>
                  <Select
                    labelId="damage-type-label"
                    value={form.damageType}
                    label="Damage type"
                    onChange={(event) =>
                      updateForm("damageType", event.target.value as DamageType)
                    }
                  >
                    {damageTypeOptions.map((option) => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>

                <TextField
                  label="Describe the damage"
                  value={form.damageDescription}
                  onChange={(event) =>
                    updateForm("damageDescription", event.target.value)
                  }
                  multiline
                  minRows={3}
                  required
                  inputProps={{ maxLength: 2000 }}
                />

                <TextField
                  label="Estimated loss amount"
                  value={form.estimatedLossAmount}
                  onChange={(event) =>
                    updateForm("estimatedLossAmount", event.target.value)
                  }
                  required
                  inputMode="decimal"
                  helperText="Enter a number, for example 12500.00"
                  inputProps={{ maxLength: 50 }}
                />

                <Grid container spacing={1.25}>
                  <Grid size={{ xs: 6 }}>
                    <TextField
                      label="Latitude"
                      value={form.lat}
                      onChange={(event) => updateForm("lat", event.target.value)}
                      fullWidth
                      inputMode="decimal"
                      required
                    />
                  </Grid>
                  <Grid size={{ xs: 6 }}>
                    <TextField
                      label="Longitude"
                      value={form.lng}
                      onChange={(event) => updateForm("lng", event.target.value)}
                      fullWidth
                      inputMode="decimal"
                      required
                    />
                  </Grid>
                </Grid>

                <Button
                  type="button"
                  variant="outlined"
                  startIcon={<MapPin size={18} />}
                  onClick={useCurrentLocation}
                  disabled={state.status === "loading"}
                >
                  Use GPS
                </Button>

                <TextField
                  label="Property address"
                  value={form.address}
                  onChange={(event) => updateForm("address", event.target.value)}
                  inputProps={{ maxLength: 300 }}
                />

                <TextField
                  label="Photo URLs (optional, comma-separated)"
                  value={form.photoUrls}
                  onChange={(event) =>
                    updateForm("photoUrls", event.target.value)
                  }
                  helperText="Paste links to photos hosted elsewhere; each link must be safe and 500 characters or fewer."
                  inputProps={{ maxLength: 2000 }}
                />

                <FormControlLabel
                  control={
                    <Checkbox
                      checked={form.privacyConsent}
                      onChange={(event) =>
                        updateForm("privacyConsent", event.target.checked)
                      }
                    />
                  }
                  label="I agree that NADAA may share this claim with insurers and relief agencies for verification and support."
                />

                {state.status === "error" ? (
                  <Alert severity="error" className="warning-alert">
                    {state.message}
                  </Alert>
                ) : null}

                <Button
                  type="submit"
                  variant="contained"
                  disabled={state.status === "loading"}
                  startIcon={
                    state.status === "loading" ? (
                      <Loader2 size={18} className="spin-icon" />
                    ) : (
                      <FileCheck size={18} />
                    )
                  }
                >
                  {state.status === "loading"
                    ? state.message
                    : "Submit damage claim"}
                </Button>
              </Stack>
            )}
          </FormDialogButton>
        </Box>
      </Stack>
    </Paper>
  );
}

export default DamageClaim;
