/**
 * Light/dark appearance for the marketing site.
 *
 * Marketing is plain HTML/CSS (no MUI), so "dark mode" is just flipping the
 * shared `--nadaa-*` tokens via `data-theme="dark"` on `<html>`. This module
 * wraps the `@nadaa/brand` theme-preference helpers in a tiny external store so
 * the header toggle icon can react, and adds the circular-reveal View
 * Transition used across the dashboards. The saved mode is applied at module
 * load (before first paint) to prevent a flash of the wrong theme.
 */
import { useSyncExternalStore } from "react";
import {
  applyThemeMode,
  initThemePreferences,
  saveMode,
  type ThemeMode,
} from "@nadaa/brand";

// Apply the saved (or OS) preference immediately on import — before first paint.
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

/** Persist + apply the mode and notify subscribers so the toggle icon updates.
 *  Applies the document attribute synchronously so a View Transition snapshot
 *  captures the new theme. */
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
 * revealed with an expanding circle from that point. Falls back to an instant
 * swap when unsupported or when the user prefers reduced motion.
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
    // No React tree state to flush here — just flip the DOM attribute inside
    // the transition callback so it is captured by the reveal snapshot.
    setThemeMode(next);
  });
  transition.finished.finally(() => {
    delete root.dataset.themeTransition;
  });
}

/** Reactive current mode. `getServerSnapshot` returns light for SSR safety. */
export function useThemeMode(): ThemeMode {
  return useSyncExternalStore(subscribe, getSnapshot, () => "light");
}
