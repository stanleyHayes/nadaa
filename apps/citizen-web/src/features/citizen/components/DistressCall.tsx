import {
  type KeyboardEvent,
  type PointerEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { LocateFixed, PhoneCall, ShieldAlert, Siren, X } from "lucide-react";
import type {
  CreateIncidentRequest,
  CreateIncidentResponse,
  HazardType,
} from "@nadaa/shared-types";
import { INCIDENT_API_BASE } from "@/app/config";
import { useCitizenSession } from "../session";
import { extractAPIError } from "../utils";

const HOLD_DURATION_MS = 2200;
const QUEUE_KEY = "nadaa.citizen.distressQueue.v1";

type DistressStatus =
  | { state: "idle"; message: string }
  | { state: "locating"; message: string }
  | { state: "sending"; message: string }
  | { state: "queued"; message: string }
  | { state: "success"; message: string; reference: string }
  | { state: "error"; message: string };

type QueuedDistress = {
  id: string;
  queuedAt: string;
  payload: CreateIncidentRequest;
};

const distressHazards: { label: string; value: HazardType }[] = [
  { label: "Medical emergency", value: "medical_emergency" },
  { label: "Crime or security danger", value: "security_incident" },
  { label: "Flood or trapped by water", value: "flood" },
  { label: "Fire", value: "fire" },
  { label: "Road crash", value: "road_crash" },
  { label: "Building collapse", value: "building_collapse" },
  { label: "Other immediate danger", value: "other" },
];

function readQueue(): QueuedDistress[] {
  try {
    const value = JSON.parse(
      window.localStorage.getItem(QUEUE_KEY) ?? "[]",
    ) as unknown;
    return Array.isArray(value) ? (value as QueuedDistress[]) : [];
  } catch {
    return [];
  }
}

function writeQueue(queue: QueuedDistress[]) {
  try {
    window.localStorage.setItem(QUEUE_KEY, JSON.stringify(queue));
  } catch {
    // The dialog still directs the citizen to 112 if storage is unavailable.
  }
}

async function postDistress(payload: CreateIncidentRequest) {
  const response = await fetch(`${INCIDENT_API_BASE}/incidents`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response.ok) throw new Error(await extractAPIError(response));
  return (await response.json()) as CreateIncidentResponse;
}

/** Always-reachable, guarded SOS intake with GPS and offline replay. */
export function DistressCall() {
  const { session, saveReport } = useCitizenSession();
  const [open, setOpen] = useState(false);
  const [hazard, setHazard] = useState<HazardType>("medical_emergency");
  const [description, setDescription] = useState(
    "I am in danger and need rescue now.",
  );
  const [lat, setLat] = useState("");
  const [lng, setLng] = useState("");
  const [allowContact, setAllowContact] = useState(true);
  const [holding, setHolding] = useState(false);
  const [queuedCount, setQueuedCount] = useState(() => readQueue().length);
  const [status, setStatus] = useState<DistressStatus>({
    state: "idle",
    message: "Share your location so dispatch can coordinate rescue.",
  });
  const holdTimer = useRef<number | null>(null);

  const locate = useCallback(() => {
    if (!navigator.geolocation) {
      setStatus({
        state: "error",
        message: "GPS is unavailable. Call 112 now.",
      });
      return;
    }
    setStatus({ state: "locating", message: "Finding your precise location…" });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLat(position.coords.latitude.toFixed(6));
        setLng(position.coords.longitude.toFixed(6));
        setStatus({
          state: "idle",
          message: `Location ready (±${Math.round(position.coords.accuracy)} m).`,
        });
      },
      () =>
        setStatus({
          state: "error",
          message:
            "We could not access your location. Enable GPS or call 112 now.",
        }),
      { enableHighAccuracy: true, timeout: 12000, maximumAge: 15000 },
    );
  }, []);

  useEffect(() => {
    if (open && !lat && !lng) locate();
  }, [lat, lng, locate, open]);

  useEffect(() => {
    const replay = async () => {
      if (!navigator.onLine) return;
      const remaining: QueuedDistress[] = [];
      for (const item of readQueue()) {
        try {
          await postDistress(item.payload);
        } catch {
          remaining.push(item);
        }
      }
      writeQueue(remaining);
      setQueuedCount(remaining.length);
    };
    window.addEventListener("online", replay);
    void replay();
    return () => window.removeEventListener("online", replay);
  }, []);

  const stopHolding = () => {
    if (holdTimer.current !== null) window.clearTimeout(holdTimer.current);
    holdTimer.current = null;
    setHolding(false);
  };

  const buildPayload = (): CreateIncidentRequest | null => {
    const latitude = Number(lat);
    const longitude = Number(lng);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      setStatus({
        state: "error",
        message:
          "A valid rescue location is required. Try GPS again or call 112.",
      });
      return null;
    }
    if (description.trim().length < 5) {
      setStatus({
        state: "error",
        message: "Briefly tell rescuers what danger you are facing.",
      });
      return null;
    }
    return {
      requestKind: "distress_request",
      type: hazard,
      description: description.trim(),
      location: { lat: latitude, lng: longitude },
      peopleAffected: 1,
      injuriesReported: hazard === "medical_emergency",
      urgency: "life_threatening",
      anonymous: !session,
      contactPermission: Boolean(session && allowContact),
      media: [],
      reporter:
        session && allowContact
          ? {
              userId: `citizen_${new Date(session.since).getTime()}`,
              phone: session.phone,
            }
          : undefined,
    };
  };

  const queuePayload = (payload: CreateIncidentRequest, message: string) => {
    const queue = [
      ...readQueue(),
      { id: crypto.randomUUID(), queuedAt: new Date().toISOString(), payload },
    ];
    writeQueue(queue);
    setQueuedCount(queue.length);
    setStatus({ state: "queued", message });
  };

  const send = async () => {
    stopHolding();
    const payload = buildPayload();
    if (!payload) return;
    if (!navigator.onLine) {
      queuePayload(
        payload,
        "No connection. Your SOS is queued and will retry automatically. Call 112 if any signal is available.",
      );
      return;
    }
    setStatus({ state: "sending", message: "Sending your rescue request…" });
    try {
      const response = await postDistress(payload);
      saveReport({
        reference: response.reference,
        hazard,
        urgency: "life_threatening",
        priorityReview: true,
        at: new Date().toISOString(),
      });
      setStatus({
        state: "success",
        reference: response.reference,
        message:
          "Dispatch received your location and rescue request. Keep your phone nearby and move only if it is safe.",
      });
    } catch (error) {
      queuePayload(
        payload,
        `${error instanceof Error ? error.message : "The network request failed."} Your SOS is queued for automatic retry. Call 112 now if possible.`,
      );
    }
  };

  const startHolding = () => {
    if (
      status.state === "sending" ||
      !lat ||
      !lng ||
      holdTimer.current !== null
    )
      return;
    setHolding(true);
    holdTimer.current = window.setTimeout(() => void send(), HOLD_DURATION_MS);
  };

  const onPointerDown = (event: PointerEvent<HTMLButtonElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId);
    startHolding();
  };
  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    if ((event.key === " " || event.key === "Enter") && !event.repeat) {
      event.preventDefault();
      startHolding();
    }
  };
  const close = () => {
    stopHolding();
    setOpen(false);
  };

  return (
    <>
      {!open ? (
        <button
          aria-label={`SOS: request rescue${queuedCount ? `, ${queuedCount} queued` : ""}`}
          className="distress-fab"
          onClick={() => setOpen(true)}
          type="button"
        >
          <span className="distress-fab__icon">
            <Siren aria-hidden="true" size={22} />
          </span>
          <span>
            <strong>SOS</strong>
            <small>Request rescue</small>
          </span>
          {queuedCount ? (
            <span className="distress-fab__count">{queuedCount}</span>
          ) : null}
        </button>
      ) : null}

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={close}
        open={open}
        className="distress-dialog"
      >
        <DialogTitle className="distress-dialog__title">
          <span>
            <ShieldAlert aria-hidden="true" size={24} /> Request rescue
          </span>
          <Button
            aria-label="Close rescue request"
            color="inherit"
            onClick={close}
          >
            <X size={19} />
          </Button>
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="error" icon={<PhoneCall size={20} />}>
              If you can call, dial <strong>112</strong> now. This SOS also
              sends your GPS to the NADAA dispatch queue; it does not replace a
              voice call.
            </Alert>
            <FormControl fullWidth>
              <InputLabel id="distress-hazard-label">
                What danger are you facing?
              </InputLabel>
              <Select
                label="What danger are you facing?"
                labelId="distress-hazard-label"
                onChange={(event) =>
                  setHazard(event.target.value as HazardType)
                }
                value={hazard}
              >
                {distressHazards.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              label="Tell rescuers what is happening"
              maxRows={5}
              minRows={3}
              multiline
              onChange={(event) => setDescription(event.target.value)}
              slotProps={{ htmlInput: { maxLength: 500 } }}
              value={description}
            />
            <Stack
              className="distress-location"
              direction={{ xs: "column", sm: "row" }}
              spacing={1.25}
            >
              <div>
                <Typography variant="subtitle2">Rescue location</Typography>
                <Typography color="text.secondary" variant="body2">
                  {lat && lng ? `${lat}, ${lng}` : "Waiting for GPS location"}
                </Typography>
              </div>
              <Button
                disabled={status.state === "locating"}
                onClick={locate}
                startIcon={<LocateFixed size={17} />}
                variant="outlined"
              >
                {lat && lng ? "Refresh GPS" : "Use GPS"}
              </Button>
            </Stack>
            {session ? (
              <FormControlLabel
                control={
                  <Switch
                    checked={allowContact}
                    onChange={(event) => setAllowContact(event.target.checked)}
                  />
                }
                label={`Let responders call ${session.phone}`}
              />
            ) : (
              <Alert severity="info">
                You can send this SOS without signing in. Dispatch will receive
                your location, but cannot call you back.
              </Alert>
            )}
            <Alert
              severity={
                status.state === "success"
                  ? "success"
                  : status.state === "error" || status.state === "queued"
                    ? "warning"
                    : "info"
              }
            >
              {status.state === "success" ? (
                <strong>{status.reference}: </strong>
              ) : null}
              {status.message}
            </Alert>
          </Stack>
        </DialogContent>
        <DialogActions className="distress-dialog__actions">
          <Button
            component="a"
            href="tel:112"
            startIcon={<PhoneCall size={18} />}
            variant="outlined"
          >
            Call 112
          </Button>
          <button
            aria-busy={status.state === "sending"}
            className={`distress-hold${holding ? " is-holding" : ""}`}
            disabled={status.state === "sending" || !lat || !lng}
            onKeyDown={onKeyDown}
            onKeyUp={stopHolding}
            onPointerCancel={stopHolding}
            onPointerDown={onPointerDown}
            onPointerLeave={stopHolding}
            onPointerUp={stopHolding}
            type="button"
          >
            <span className="distress-hold__progress" />
            <Siren aria-hidden="true" size={19} />
            <span>
              {status.state === "sending"
                ? "Sending SOS…"
                : holding
                  ? "Keep holding…"
                  : "Hold to send SOS"}
            </span>
          </button>
        </DialogActions>
      </Dialog>
    </>
  );
}
