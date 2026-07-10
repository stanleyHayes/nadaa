import { FormEvent, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  FormControl,
  FormControlLabel,
  FormHelperText,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
} from "@mui/material";
import { BookmarkCheck, Languages, ShieldCheck, UserPlus, X } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import { guideLanguageOptions } from "../data";
import { signInRegions, type CitizenSession } from "../session";

type SignInDialogProps = {
  open: boolean;
  onClose: () => void;
  onSignIn: (details: Omit<CitizenSession, "since">) => void;
};

type SignInErrors = Partial<
  Record<"name" | "phone" | "region" | "consent", string>
>;

/**
 * Optional light sign-in. Mirrors the marketing signup shape (name / +233 phone
 * / region / language / consent) in a split-screen layout: a solid navy brand
 * panel on the left, the form on the right. Purely local — it unlocks saved
 * reports and claims and never blocks any life-safety action.
 */
export function SignInDialog({ open, onClose, onSignIn }: SignInDialogProps) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("+233 ");
  const [region, setRegion] = useState<string>(signInRegions[0]);
  const [language, setLanguage] = useState("en");
  const [consent, setConsent] = useState(false);
  const [errors, setErrors] = useState<SignInErrors>({});

  const reset = () => {
    setName("");
    setPhone("+233 ");
    setRegion(signInRegions[0]);
    setLanguage("en");
    setConsent(false);
    setErrors({});
  };

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors: SignInErrors = {};
    if (name.trim().length < 2) {
      nextErrors.name = "Enter your name so responders can greet you.";
    }
    const digits = phone.replace(/[^\d]/g, "");
    if (!/^233\d{9}$/.test(digits)) {
      nextErrors.phone = "Enter a Ghana number, e.g. +233 20 000 0000.";
    }
    if (!region) {
      nextErrors.region = "Choose your region.";
    }
    if (!consent) {
      nextErrors.consent = "Please agree so we can save your reports.";
    }
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }
    onSignIn({
      name: name.trim(),
      phone: `+${digits}`,
      region,
      language,
    });
    reset();
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="md"
      aria-labelledby="citizen-signin-title"
      PaperProps={{ sx: { borderRadius: "16px", overflow: "hidden" } }}
    >
      <IconButton
        aria-label="Close sign in"
        onClick={onClose}
        className="signin-close"
      >
        <X size={20} />
      </IconButton>

      <div className="signin-split">
        <aside className="signin-brand">
          <span className="signin-brand__mark">
            <img alt="" src="/brand/nadaa-logo.png" />
            <strong>{nadaaBrand.name}</strong>
          </span>
          <h2 id="citizen-signin-title">Save your reports</h2>
          <p>
            Optional. NADAA stays fully usable without signing in — this only
            keeps a copy of your reports and claims on this device.
          </p>
          <ul className="signin-points">
            <li>
              <ShieldCheck size={18} aria-hidden="true" />
              Life-safety features never require sign-in.
            </li>
            <li>
              <BookmarkCheck size={18} aria-hidden="true" />
              Follow up on the reports and claims you send.
            </li>
            <li>
              <Languages size={18} aria-hidden="true" />
              Guidance shown in your language.
            </li>
          </ul>
        </aside>

        <div className="signin-form">
          <Stack component="form" spacing={2} onSubmit={submit} noValidate>
            <TextField
              id="signin-name"
              label="Full name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              error={Boolean(errors.name)}
              helperText={errors.name}
              fullWidth
              autoComplete="name"
            />
            <TextField
              id="signin-phone"
              label="Mobile number"
              value={phone}
              onChange={(event) => setPhone(event.target.value)}
              error={Boolean(errors.phone)}
              helperText={errors.phone ?? "Used only for report follow-up."}
              fullWidth
              inputMode="tel"
              autoComplete="tel"
            />
            <FormControl fullWidth error={Boolean(errors.region)}>
              <InputLabel id="signin-region-label">Region</InputLabel>
              <Select
                labelId="signin-region-label"
                id="signin-region"
                label="Region"
                value={region}
                onChange={(event) => setRegion(event.target.value)}
              >
                {signInRegions.map((item) => (
                  <MenuItem key={item} value={item}>
                    {item}
                  </MenuItem>
                ))}
              </Select>
              {errors.region ? (
                <FormHelperText>{errors.region}</FormHelperText>
              ) : null}
            </FormControl>
            <FormControl fullWidth>
              <InputLabel id="signin-language-label">Language</InputLabel>
              <Select
                labelId="signin-language-label"
                id="signin-language"
                label="Language"
                value={language}
                onChange={(event) => setLanguage(event.target.value)}
              >
                {guideLanguageOptions.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl error={Boolean(errors.consent)}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={consent}
                    onChange={(event) => setConsent(event.target.checked)}
                  />
                }
                label="I agree that NADAA can save my reports and claims on this device and contact me about them."
              />
              {errors.consent ? (
                <FormHelperText>{errors.consent}</FormHelperText>
              ) : null}
            </FormControl>
            <Alert
              severity="info"
              className="warning-alert"
              icon={<ShieldCheck />}
            >
              Sign-in is optional and stored on this device only. Life-safety
              features never require it.
            </Alert>
            <Button
              type="submit"
              variant="contained"
              color="warning"
              startIcon={<UserPlus size={18} />}
              className="signin-submit"
            >
              Save and continue
            </Button>
          </Stack>
        </div>
      </div>
    </Dialog>
  );
}

export default SignInDialog;
