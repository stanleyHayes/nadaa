/**
 * Reactive appearance store for the command center.
 *
 * Wraps the shared `@nadaa/brand` theme-preference helpers in a tiny external
 * store so the MUI theme can rebuild when the operator flips light/dark. The
 * saved mode + tint are applied to the document at module load (before React
 * renders) to prevent a flash of the wrong theme.
 */
import { useSyncExternalStore } from "react";
import {
  initThemePreferences,
  readSavedTint,
  saveMode,
  saveTint,
  type DarkTint,
  type ThemeMode,
} from "@nadaa/brand";

// Apply saved prefs immediately on import — this runs before the first paint.
const initial = initThemePreferences();

let currentMode: ThemeMode = initial.mode;
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot(): ThemeMode {
  return currentMode;
}

/** Persist + apply the mode and notify subscribers so MUI rebuilds. */
export function setThemeMode(mode: ThemeMode) {
  if (mode === currentMode) return;
  currentMode = mode;
  saveMode(mode);
  emit();
}

/** Flip between light and dark. */
export function toggleThemeMode() {
  setThemeMode(currentMode === "dark" ? "light" : "dark");
}

/** Reactive current mode. `getServerSnapshot` returns light for SSR safety. */
export function useThemeMode(): ThemeMode {
  return useSyncExternalStore(subscribe, getSnapshot, () => "light");
}

/** Persist + apply a dark screen tint. Only affects CSS, not the MUI palette. */
export function setDarkTint(tint: DarkTint) {
  saveTint(tint);
}

/** The tint applied at boot (Ink unless the operator saved another). */
export function getInitialTint(): DarkTint {
  return readSavedTint();
}
