import { useSyncExternalStore } from "react";

/**
 * Command-center alarm. When an approved alert is published to citizens, the
 * console sounds an urgent klaxon so on-duty staff notice a warning just went
 * out — the "beeping" cue you'd expect in an operations room. Synthesised with
 * the Web Audio API (no asset); playback follows the operator's click that
 * triggered the publish, so the browser autoplay gate is satisfied.
 */

const STORAGE_KEY = "nadaa.authority.alarm.enabled";

let audioContext: AudioContext | null = null;

function getContext(): AudioContext | null {
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

export function isAlarmEnabled(): boolean {
  if (typeof localStorage === "undefined") {
    return true;
  }
  return localStorage.getItem(STORAGE_KEY) !== "false";
}

const listeners = new Set<() => void>();

export function setAlarmEnabled(enabled: boolean): void {
  try {
    localStorage.setItem(STORAGE_KEY, enabled ? "true" : "false");
  } catch {
    // Ignore storage failures (private mode); the in-memory listeners still fire.
  }
  listeners.forEach((listener) => listener());
}

/** Reactive `[enabled, toggle]` for a UI switch. */
export function useAlarmEnabled(): [boolean, () => void] {
  const enabled = useSyncExternalStore(
    (listener) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    isAlarmEnabled,
    () => true,
  );
  return [enabled, () => setAlarmEnabled(!enabled)];
}

/** Play the urgent two-tone klaxon (no-op when muted or Web Audio is absent). */
export function playCommandAlarm(): void {
  if (!isAlarmEnabled()) {
    return;
  }
  const ctx = getContext();
  if (!ctx) {
    return;
  }
  if (ctx.state === "suspended") {
    void ctx.resume();
  }

  // Alternating two-tone klaxon, ending on a longer high note.
  const pattern: Array<[number, number]> = [
    [740, 0.2],
    [560, 0.2],
    [740, 0.2],
    [560, 0.2],
    [880, 0.42],
  ];
  let at = ctx.currentTime + 0.01;
  for (const [freq, dur] of pattern) {
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = "square";
    osc.frequency.setValueAtTime(freq, at);
    gain.gain.setValueAtTime(0, at);
    gain.gain.linearRampToValueAtTime(0.18, at + 0.02);
    gain.gain.linearRampToValueAtTime(0, at + dur);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(at);
    osc.stop(at + dur + 0.03);
    at += dur + 0.05;
  }
}
