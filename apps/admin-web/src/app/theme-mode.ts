/**
 * Reactive appearance store for the admin console.
 *
 * Wraps the shared `@nadaa/brand` theme-preference helpers in a tiny external
 * store so the MUI theme can rebuild when the admin flips light/dark. The saved
 * mode + tint are applied to the document at module load (before React renders)
 * to prevent a flash of the wrong theme.
 */
import { useSyncExternalStore } from "react";
import { flushSync } from "react-dom";
import {
  applyThemeMode,
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

/** Persist + apply the mode and notify subscribers so MUI rebuilds. Applies the
 * document attribute synchronously so a View Transition snapshot captures it. */
export function setThemeMode(mode: ThemeMode) {
  if (mode === currentMode) return;
  currentMode = mode;
  applyThemeMode(mode);
  saveMode(mode);
  emit();
}

function prefersReducedMotion(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia?.("(prefers-reduced-motion: reduce)").matches === true
  );
}

/**
 * Flip between light and dark. When an `origin` (the toggle's screen position)
 * is given and the browser supports the View Transitions API, the new theme is
 * revealed with an expanding circle from that point.
 */
export function toggleThemeMode(origin?: { x: number; y: number }) {
  const next: ThemeMode = currentMode === "dark" ? "light" : "dark";
  const doc = typeof document === "undefined" ? undefined : document;

  if (
    !doc ||
    !("startViewTransition" in doc) ||
    !origin ||
    prefersReducedMotion()
  ) {
    setThemeMode(next);
    return;
  }

  const root = doc.documentElement;
  root.style.setProperty("--nadaa-reveal-x", `${origin.x}px`);
  root.style.setProperty("--nadaa-reveal-y", `${origin.y}px`);
  root.dataset.themeTransition = "reveal";

  const transition = (
    doc as Document & {
      startViewTransition: (cb: () => void) => { finished: Promise<void> };
    }
  ).startViewTransition(() => {
    // flushSync so the DOM/React change happens inside the transition capture.
    flushSync(() => setThemeMode(next));
  });
  transition.finished.finally(() => {
    delete root.dataset.themeTransition;
  });
}

/** Reactive current mode. `getServerSnapshot` returns light for SSR safety. */
export function useThemeMode(): ThemeMode {
  return useSyncExternalStore(subscribe, getSnapshot, () => "light");
}

/** Persist + apply a dark screen tint. Only affects CSS, not the MUI palette. */
export function setDarkTint(tint: DarkTint) {
  saveTint(tint);
}

/** The tint applied at boot (Ink unless the admin saved another). */
export function getInitialTint(): DarkTint {
  return readSavedTint();
}
