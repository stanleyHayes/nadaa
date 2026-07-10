import { type FormEvent, useState } from "react";
import {
  Alert,
  Button,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
} from "@mui/material";
import { Save } from "lucide-react";
import { guideLanguageOptions } from "../../../data";
import {
  signInRegions,
  useCitizenSession,
  type CitizenSession,
  type ContactChannel,
} from "../../../session";
import { contactChannelOptions } from "../data";

type ProfileErrors = Partial<Record<"name" | "phone" | "email", string>>;

/** Edit the citizen's contact profile and save it back to the session store. */
export function ProfileForm({ session }: { session: CitizenSession }) {
  const { updateProfile } = useCitizenSession();
  const [name, setName] = useState(session.name);
  const [phone, setPhone] = useState(session.phone);
  const [email, setEmail] = useState(session.email ?? "");
  const [region, setRegion] = useState(session.region);
  const [language, setLanguage] = useState(session.language);
  const [contactChannel, setContactChannel] = useState<ContactChannel>(
    session.contactChannel ?? "sms",
  );
  const [errors, setErrors] = useState<ProfileErrors>({});
  const [saved, setSaved] = useState(false);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors: ProfileErrors = {};
    if (name.trim().length < 2) {
      nextErrors.name = "Enter your name so responders can greet you.";
    }
    const digits = phone.replace(/[^\d]/g, "");
    if (!/^233\d{9}$/.test(digits)) {
      nextErrors.phone = "Enter a Ghana number, e.g. +233 20 000 0000.";
    }
    if (email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      nextErrors.email = "Enter a valid email address, or leave it blank.";
    }
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      setSaved(false);
      return;
    }
    setErrors({});
    updateProfile({
      name: name.trim(),
      phone: `+${digits}`,
      email: email.trim() || undefined,
      region,
      language,
      contactChannel,
    });
    setSaved(true);
  };

  return (
    <Stack
      component="form"
      spacing={2.25}
      onSubmit={submit}
      noValidate
      sx={{ maxWidth: 560 }}
    >
      {saved ? (
        <Alert severity="success" className="warning-alert" onClose={() => setSaved(false)}>
          Profile updated. Responders will use these details to reach you.
        </Alert>
      ) : null}

      <TextField
        label="Full name"
        value={name}
        onChange={(event) => {
          setName(event.target.value);
          setSaved(false);
        }}
        error={Boolean(errors.name)}
        helperText={errors.name}
        fullWidth
        autoComplete="name"
      />
      <TextField
        label="Mobile number"
        value={phone}
        onChange={(event) => {
          setPhone(event.target.value);
          setSaved(false);
        }}
        error={Boolean(errors.phone)}
        helperText={errors.phone ?? "Used for report follow-up and SMS alerts."}
        fullWidth
        inputMode="tel"
        autoComplete="tel"
      />
      <TextField
        label="Email (optional)"
        value={email}
        onChange={(event) => {
          setEmail(event.target.value);
          setSaved(false);
        }}
        error={Boolean(errors.email)}
        helperText={errors.email ?? "Add an email to receive alerts by inbox too."}
        fullWidth
        inputMode="email"
        autoComplete="email"
      />
      <FormControl fullWidth>
        <InputLabel id="profile-region-label">Region</InputLabel>
        <Select
          labelId="profile-region-label"
          label="Region"
          value={region}
          onChange={(event) => {
            setRegion(event.target.value);
            setSaved(false);
          }}
        >
          {signInRegions.map((item) => (
            <MenuItem key={item} value={item}>
              {item}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl fullWidth>
        <InputLabel id="profile-language-label">Language</InputLabel>
        <Select
          labelId="profile-language-label"
          label="Language"
          value={language}
          onChange={(event) => {
            setLanguage(event.target.value);
            setSaved(false);
          }}
        >
          {guideLanguageOptions.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl fullWidth>
        <InputLabel id="profile-channel-label">Preferred contact</InputLabel>
        <Select
          labelId="profile-channel-label"
          label="Preferred contact"
          value={contactChannel}
          onChange={(event) => {
            setContactChannel(event.target.value as ContactChannel);
            setSaved(false);
          }}
        >
          {contactChannelOptions.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      <div>
        <Button
          type="submit"
          variant="contained"
          color="warning"
          startIcon={<Save size={18} />}
          className="signin-submit"
        >
          Save profile
        </Button>
      </div>
    </Stack>
  );
}

export default ProfileForm;
