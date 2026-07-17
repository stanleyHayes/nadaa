import { type FormEvent, useState } from "react";
import { Alert, Button, Stack, TextField } from "@mui/material";
import { KeyRound } from "lucide-react";
import { useCitizenSession } from "../../../session";

type PasswordFeedback =
  | { kind: "idle" }
  | { kind: "success" }
  | { kind: "error"; message: string };

/**
 * Change-password form. The store's `changePassword` is a validation-only mock
 * (there is no credential backend yet), so this focuses on clear client-side
 * checks and success/error feedback.
 */
export function PasswordForm() {
  const { changePassword } = useCitizenSession();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [feedback, setFeedback] = useState<PasswordFeedback>({ kind: "idle" });

  const reset = () => {
    setCurrent("");
    setNext("");
    setConfirm("");
  };

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (next !== confirm) {
      setFeedback({ kind: "error", message: "New passwords do not match." });
      return;
    }
    const result = changePassword(current, next);
    if (result.ok) {
      setFeedback({ kind: "success" });
      reset();
    } else {
      setFeedback({ kind: "error", message: result.error });
    }
  };

  return (
    <Stack
      component="form"
      spacing={2.25}
      onSubmit={submit}
      noValidate
      sx={{ maxWidth: 560 }}
    >
      {feedback.kind === "success" ? (
        <Alert
          severity="success"
          className="warning-alert"
          onClose={() => setFeedback({ kind: "idle" })}
        >
          Password updated on this device (preview — no credential backend
          enforces it yet).
        </Alert>
      ) : null}
      {feedback.kind === "error" ? (
        <Alert
          severity="error"
          className="warning-alert"
          onClose={() => setFeedback({ kind: "idle" })}
        >
          {feedback.message}
        </Alert>
      ) : null}

      <TextField
        label="Current password"
        type="password"
        value={current}
        onChange={(event) => setCurrent(event.target.value)}
        fullWidth
        autoComplete="current-password"
      />
      <TextField
        label="New password"
        type="password"
        value={next}
        onChange={(event) => setNext(event.target.value)}
        helperText="Use at least 8 characters."
        fullWidth
        autoComplete="new-password"
      />
      <TextField
        label="Confirm new password"
        type="password"
        value={confirm}
        onChange={(event) => setConfirm(event.target.value)}
        fullWidth
        autoComplete="new-password"
      />

      <div>
        <Button
          type="submit"
          variant="contained"
          color="warning"
          startIcon={<KeyRound size={18} />}
          className="signin-submit"
        >
          Update password
        </Button>
      </div>
    </Stack>
  );
}

export default PasswordForm;
