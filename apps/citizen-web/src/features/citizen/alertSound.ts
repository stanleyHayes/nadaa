import type { AlertSeverity } from "@nadaa/shared-types";
import type { QuietHours } from "./session";

/**
 * Audible warning tones for incoming citizen alerts.
 *
 * The web has no access to the OS Do-Not-Disturb state, so "respect DND" is
 * modelled on the citizen's Quiet Hours preference. A level-5 alert
 * (`emergency`) is treated like a real critical alert: it overrides Quiet Hours
 * and always sounds (as long as the master sound toggle is on), matching how
 * national emergency alerts behave. Lower levels stay silent during Quiet Hours.
 *
 * Tones are synthesised with the Web Audio API (no asset to load). Browsers gate
 * audio behind a user gesture; alerts arrive while the citizen is using the app,
 * so the context resumes on the interaction that triggered the fetch.
 */

const LEVEL_5: AlertSeverity = "emergency";

const SEVERITY_ORDER: AlertSeverity[] = [
  "advisory",
  "watch",
  "warning",
  "severe_warning",
  "emergency",
];

/** 0..4 rank so the loudest new alert in a batch can be picked. */
export function severityRank(severity: AlertSeverity): number {
  const index = SEVERITY_ORDER.indexOf(severity);
  return index === -1 ? 0 : index;
}

let audioContext: AudioContext | null = null;

function getAudioContext(): AudioContext | null {
  if (typeof window === "undefined") {
    return null;
  }
  const Ctor =
    window.AudioContext ??
    (window as unknown as { webkitAudioContext?: typeof AudioContext })
      .webkitAudioContext;
  if (!Ctor) {
    return null;
  }
  if (!audioContext) {
    audioContext = new Ctor();
  }
  return audioContext;
}

/** True when Quiet Hours are enabled and "now" falls inside the window. */
export function quietHoursActive(quietHours: QuietHours, now = new Date()): boolean {
  if (!quietHours.enabled) {
    return false;
  }
  const toMinutes = (value: string): number | null => {
    const match = /^(\d{1,2}):(\d{2})$/.exec(value.trim());
    if (!match) {
      return null;
    }
    return Number(match[1]) * 60 + Number(match[2]);
  };
  const start = toMinutes(quietHours.start);
  const end = toMinutes(quietHours.end);
  if (start === null || end === null) {
    return false;
  }
  const current = now.getHours() * 60 + now.getMinutes();
  // Same-day window (e.g. 09:00–17:00) vs. overnight window (e.g. 22:00–06:00).
  return start <= end
    ? current >= start && current < end
    : current >= start || current < end;
}

/**
 * Decide whether an alert of `severity` should sound. Level-5 (`emergency`)
 * overrides Quiet Hours; everything else respects it. The master `soundEnabled`
 * toggle silences everything, including emergencies, so the citizen keeps final
 * control on their own device.
 */
export function shouldPlayAlertSound(
  severity: AlertSeverity,
  opts: { soundEnabled: boolean; quietHoursActive: boolean },
): boolean {
  if (!opts.soundEnabled) {
    return false;
  }
  if (severity === LEVEL_5) {
    return true;
  }
  return !opts.quietHoursActive;
}

/**
 * Play an urgent two-tone warning. Emergencies get a longer, more insistent
 * pattern. Safe to call anywhere — it no-ops when Web Audio is unavailable.
 */
export function playAlertTone(severity: AlertSeverity): void {
  const ctx = getAudioContext();
  if (!ctx) {
    return;
  }
  if (ctx.state === "suspended") {
    void ctx.resume();
  }

  const emergency = severity === LEVEL_5;
  const beeps = emergency ? 5 : 2;
  const high = emergency ? 988 : 784;
  const low = emergency ? 784 : 622;
  const beepDur = 0.18;
  const gap = emergency ? 0.1 : 0.14;
  const peak = emergency ? 0.34 : 0.2;
  const start = ctx.currentTime + 0.01;

  for (let index = 0; index < beeps; index += 1) {
    const at = start + index * (beepDur + gap);
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = "sine";
    osc.frequency.setValueAtTime(index % 2 === 0 ? high : low, at);
    gain.gain.setValueAtTime(0, at);
    gain.gain.linearRampToValueAtTime(peak, at + 0.02);
    gain.gain.linearRampToValueAtTime(0, at + beepDur);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(at);
    osc.stop(at + beepDur + 0.03);
  }
}
