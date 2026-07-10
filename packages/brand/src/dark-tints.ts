/**
 * NADAA dark-screen tint catalogue — the single source of truth for the eight
 * Navy Command dark tints.
 *
 * A tint changes the dark screen CAST only (the surface / inset / card hue);
 * text and brand accents stay put. Each entry carries display copy plus a
 * three-swatch preview `[deep base, raised card, accent hue]` rendered as sRGB
 * hex (NADAA styles in hex, not oklch).
 *
 * The `value`s map 1:1 to the `:root[data-theme="dark"][data-dark-tint="…"]`
 * blocks in `dark.css`. Order is load-bearing (Ink first / default).
 */

export const DARK_TINTS = [
  {
    value: "ink",
    label: "Ink",
    description: "The default warm navy-black command screen.",
    swatches: ["#0b1120", "#1a2440", "#8ea6dc"],
  },
  {
    value: "burgundy",
    label: "Burgundy",
    description: "A deeper maroon cast for night operations.",
    swatches: ["#150a12", "#2a1622", "#c65a6e"],
  },
  {
    value: "midnight",
    label: "Midnight",
    description: "Cooler charcoal with a quiet blue undertone.",
    swatches: ["#080d1a", "#17203a", "#5b86d6"],
  },
  {
    value: "canopy",
    label: "Canopy",
    description: "A calm green-black tint for low glare.",
    swatches: ["#08130e", "#16281d", "#2eba71"],
  },
  {
    value: "slate",
    label: "Slate",
    description: "A neutral cool-grey screen with very little colour.",
    swatches: ["#0d1017", "#20262f", "#8ea0b8"],
  },
  {
    value: "ocean",
    label: "Ocean",
    description: "A deep blue-teal cast for focused evenings.",
    swatches: ["#06121a", "#142a33", "#22b1c4"],
  },
  {
    value: "plum",
    label: "Plum",
    description: "A soft violet-black for a warmer night mood.",
    swatches: ["#120a18", "#241a30", "#a878d6"],
  },
  {
    value: "espresso",
    label: "Espresso",
    description: "A warm brown-black, easy on the eyes at night.",
    swatches: ["#130d08", "#281f16", "#c79a5c"],
  },
] as const satisfies readonly {
  value: string;
  label: string;
  description: string;
  swatches: readonly [string, string, string];
}[];

/** The value-union of the eight tints. `DARK_TINTS` is the ordering authority. */
export type DarkTint = (typeof DARK_TINTS)[number]["value"];

/** Ink is the neutral warm-navy default screen. */
export const DEFAULT_TINT: DarkTint = "ink";

/** All tint values, in catalogue order. */
export const DARK_TINT_VALUES = DARK_TINTS.map((t) => t.value) as DarkTint[];
