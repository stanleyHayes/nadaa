import { type FormEvent, useState } from "react";
import {
  Alert,
  Button,
  Divider,
  FormControl,
  FormControlLabel,
  FormGroup,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { Save } from "lucide-react";
import { guideLanguageOptions } from "../../../data";
import {
  signInRegions,
  useCitizenSession,
  type CitizenPreferences,
} from "../../../session";

/** Edit alert language, delivery channels, region of interest and quiet hours. */
export function PreferencesForm({
  preferences,
}: {
  preferences: CitizenPreferences;
}) {
  const { updatePreferences } = useCitizenSession();
  const [language, setLanguage] = useState(preferences.language);
  const [regionOfInterest, setRegionOfInterest] = useState(
    preferences.regionOfInterest,
  );
  const [channels, setChannels] = useState(preferences.alertChannels);
  const [quietHours, setQuietHours] = useState(preferences.quietHours);
  const [soundAlerts, setSoundAlerts] = useState(
    preferences.soundAlerts ?? true,
  );
  const [saved, setSaved] = useState(false);

  const dirty = () => setSaved(false);

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    updatePreferences({
      language,
      regionOfInterest,
      alertChannels: channels,
      quietHours,
      soundAlerts,
    });
    setSaved(true);
  };

  return (
    <Stack
      component="form"
      spacing={2.5}
      onSubmit={submit}
      noValidate
      sx={{ maxWidth: 560 }}
    >
      {saved ? (
        <Alert
          severity="success"
          className="warning-alert"
          onClose={() => setSaved(false)}
        >
          Preferences saved. We will use these for how and when we reach you.
        </Alert>
      ) : null}
      <FormControl fullWidth>
        <InputLabel id="pref-language-label">Alert language</InputLabel>
        <Select
          labelId="pref-language-label"
          label="Alert language"
          value={language}
          onChange={(event) => {
            setLanguage(event.target.value);
            dirty();
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
        <InputLabel id="pref-region-label">Region of interest</InputLabel>
        <Select
          labelId="pref-region-label"
          label="Region of interest"
          value={regionOfInterest}
          onChange={(event) => {
            setRegionOfInterest(event.target.value);
            dirty();
          }}
        >
          {signInRegions.map((item) => (
            <MenuItem key={item} value={item}>
              {item}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <Divider />
      <div>
        <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
          Alert channels
        </Typography>
        <Typography variant="body2" sx={{
          color: "text.secondary"
        }}>
          Choose how official warnings reach you.
        </Typography>
        <FormGroup sx={{ mt: 1 }}>
          <FormControlLabel
            control={
              <Switch
                checked={channels.sms}
                onChange={(event) => {
                  setChannels((prev) => ({ ...prev, sms: event.target.checked }));
                  dirty();
                }}
              />
            }
            label="SMS text messages"
          />
          <FormControlLabel
            control={
              <Switch
                checked={channels.email}
                onChange={(event) => {
                  setChannels((prev) => ({
                    ...prev,
                    email: event.target.checked,
                  }));
                  dirty();
                }}
              />
            }
            label="Email"
          />
          <FormControlLabel
            control={
              <Switch
                checked={channels.push}
                onChange={(event) => {
                  setChannels((prev) => ({
                    ...prev,
                    push: event.target.checked,
                  }));
                  dirty();
                }}
              />
            }
            label="Push notifications"
          />
        </FormGroup>
      </div>
      <Divider />
      <div>
        <FormControlLabel
          control={
            <Switch
              checked={soundAlerts}
              onChange={(event) => {
                setSoundAlerts(event.target.checked);
                dirty();
              }}
            />
          }
          label="Sound alerts (audible warning tone)"
        />
        <Typography
          variant="body2"
          sx={{ color: "text.secondary", mt: 0.25 }}
        >
          Play a warning tone when a new alert arrives. Emergency (level 5)
          alerts always sound, even during quiet hours.
        </Typography>
      </div>
      <div>
        <FormControlLabel
          control={
            <Switch
              checked={quietHours.enabled}
              onChange={(event) => {
                setQuietHours((prev) => ({
                  ...prev,
                  enabled: event.target.checked,
                }));
                dirty();
              }}
            />
          }
          label="Quiet hours (mute non-emergency alerts)"
        />
        <Typography
          variant="body2"
          sx={{
            color: "text.secondary",
            mb: 1
          }}>
          Life-threatening emergency alerts always come through, day or night.
        </Typography>
        <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
          <TextField
            label="From"
            type="time"
            value={quietHours.start}
            onChange={(event) => {
              setQuietHours((prev) => ({ ...prev, start: event.target.value }));
              dirty();
            }}
            disabled={!quietHours.enabled}
            fullWidth
            slotProps={{
              inputLabel: { shrink: true }
            }}
          />
          <TextField
            label="To"
            type="time"
            value={quietHours.end}
            onChange={(event) => {
              setQuietHours((prev) => ({ ...prev, end: event.target.value }));
              dirty();
            }}
            disabled={!quietHours.enabled}
            fullWidth
            slotProps={{
              inputLabel: { shrink: true }
            }}
          />
        </Stack>
      </div>
      <div>
        <Button
          type="submit"
          variant="contained"
          color="warning"
          startIcon={<Save size={18} />}
          className="signin-submit"
        >
          Save preferences
        </Button>
      </div>
    </Stack>
  );
}

export default PreferencesForm;
