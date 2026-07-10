/**
 * Theme-preference helpers, shared by every NADAA web app.
 *
 * These apply the appearance choice to the document and persist it, matching
 * the AURA model but in NADAA's namespace:
 *
 * - `applyThemeMode(mode)` sets `documentElement.dataset.theme` (drives the
 *   `:root[data-theme="dark"]` token overrides in `dark.css`).
 * - `applyDarkTint(tint)` sets `documentElement.dataset.darkTint` (drives the
 *   `[data-dark-tint="…"]` screen-cast overrides).
 * - read/save helpers persist to `localStorage` (`nadaa.theme` / `nadaa.dark-tint`).
 *
 * Every function guards `typeof window`/`document` so it is safe to call from
 * module scope during app boot (used to apply saved prefs before first paint,
 * preventing a light-mode flash). No clock is read at module load.
 */

import { DARK_TINTS, DEFAULT_TINT, type DarkTint } from "./dark-tints.js";

export { DARK_TINTS, DEFAULT_TINT };
export type { DarkTint };

export type ThemeMode = "light" | "dark";

export const THEME_STORAGE_KEY = "nadaa.theme";
export const DARK_TINT_STORAGE_KEY = "nadaa.dark-tint";

/** Event dispatched on `window` whenever the mode or tint changes, so any
 *  listener (e.g. the MUI theme provider) can react without prop threading. */
export const THEME_CHANGE_EVENT = "nadaa-theme-change";

const DEFAULT_MODE: ThemeMode = "light";

const tintValues = new Set<string>(DARK_TINTS.map((t) => t.value));

function prefersDarkTheme(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) return false;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

/** Coerce arbitrary input to a known mode. */
export function normaliseMode(value: unknown): ThemeMode {
  return value === "dark" || value === "light" ? value : DEFAULT_MODE;
}

/** Coerce arbitrary input to a known tint, falling back to Ink. */
export function normaliseTint(value: unknown): DarkTint {
  return typeof value === "string" && tintValues.has(value)
    ? (value as DarkTint)
    : DEFAULT_TINT;
}

/** Saved mode, or the OS preference on first visit. */
export function readSavedMode(): ThemeMode {
  if (typeof window === "undefined") return DEFAULT_MODE;
  const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
  if (stored === "light" || stored === "dark") return stored;
  return prefersDarkTheme() ? "dark" : "light";
}

/** Saved tint, or the Ink default. */
export function readSavedTint(): DarkTint {
  if (typeof window === "undefined") return DEFAULT_TINT;
  return normaliseTint(window.localStorage.getItem(DARK_TINT_STORAGE_KEY));
}

/** Reflect the mode onto `<html data-theme>` and `color-scheme`. */
export function applyThemeMode(mode: ThemeMode): void {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  root.dataset.theme = mode;
  root.style.colorScheme = mode;
}

/** Reflect the tint onto `<html data-dark-tint>`. */
export function applyDarkTint(tint: DarkTint): void {
  if (typeof document === "undefined") return;
  document.documentElement.dataset.darkTint = tint;
}

/** Persist + apply the mode, then broadcast the change. */
export function saveMode(mode: ThemeMode): void {
  applyThemeMode(mode);
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, mode);
  } catch {
    // Storage can be unavailable (private mode); the applied attribute stands.
  }
  window.dispatchEvent(new Event(THEME_CHANGE_EVENT));
}

/** Persist + apply the tint, then broadcast the change. */
export function saveTint(tint: DarkTint): void {
  applyDarkTint(tint);
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(DARK_TINT_STORAGE_KEY, tint);
  } catch {
    // Storage can be unavailable; the applied attribute stands.
  }
  window.dispatchEvent(new Event(THEME_CHANGE_EVENT));
}

/**
 * Apply the saved mode + tint to the document. Call once during app boot,
 * before first render, to avoid a flash of the wrong theme.
 */
export function initThemePreferences(): { mode: ThemeMode; tint: DarkTint } {
  const mode = readSavedMode();
  const tint = readSavedTint();
  applyThemeMode(mode);
  applyDarkTint(tint);
  return { mode, tint };
}
