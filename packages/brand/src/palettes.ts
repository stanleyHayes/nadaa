/**
 * Brand palettes, keyed by the same slot names in both brands.
 *
 * The slots are historical: `navy` is the brand-ink slot every semantic role
 * derives from, and it has held a navy, then a near-black, and now SUBVENIO's
 * deep navy. The NAME is load-bearing — roughly 5,000 `--nadaa-navy` references
 * and 78 importers resolve through it — so brands re-point the slot rather than
 * rename it. See identity.ts for what else must not be renamed.
 *
 * SUBVENIO values come from the Brand, Look & Feel and Motion guide §3. Two
 * rules from that document are encoded here rather than left to judgement:
 * red is semantic-only and never the master brand colour, and orange
 * punctuates rather than fills.
 */

import { ACTIVE_BRAND, type BrandId } from "./identity.js";

export type BrandPalette = {
  /** Brand ink. The anchor every semantic role derives from. */
  readonly navy: string;
  readonly charcoal: string;
  readonly silver: string;
  readonly lightSilver: string;
  /** Safety / confirmed-safe. */
  readonly green: string;
  /** Danger. Semantic only — never the master brand colour. */
  readonly red: string;
  /** Accent / warning. */
  readonly gold: string;
  readonly slate: string;
  /** Elevated surface in light; re-pointed dark by dark.css. */
  readonly white: string;
  /** Recessed surface. */
  readonly mist: string;
  /** Primary text. */
  readonly ink: string;
  /** Informational. */
  readonly info: string;
  readonly border: string;
  readonly divider: string;
};

export const PALETTES: Record<BrandId, BrandPalette> = {
  // Onyx Command — black + silver.
  nadaa: {
    navy: "#1A1A1A",
    charcoal: "#2B2B2B",
    silver: "#8E8E8E",
    lightSilver: "#E6E6E6",
    green: "#118D4E",
    red: "#E53935",
    gold: "#F4C20D",
    slate: "#5A5A5A",
    white: "#FFFFFF",
    mist: "#F4F4F5",
    ink: "#0D0D0D",
    info: "#0B6FB8",
    border: "#E6E6E6",
    divider: "#EAEAEA",
  },
  // SUBVENIO — safe-blue / deep-navy / hope-orange / soft-teal.
  subvenio: {
    // deep-navy: foundation, navigation, hero surfaces.
    navy: "#0A1F3D",
    charcoal: "#132E52",
    // soft-teal: Halo, responder availability, supportive states.
    silver: "#22D3CE",
    lightSilver: "#CBD5E1",
    // success: confirmed-safe only.
    green: "#16A36A",
    // danger: critical/destructive only. The emergency brand accent is orange.
    red: "#E5484D",
    // hope-orange: urgency and human action. Punctuates; never floods.
    gold: "#FF8A3D",
    // muted — secondary text and metadata. Darkened from the guide's #64748B,
    // which measures 4.41:1 on the emergency tint and 4.51:1 on the soft card,
    // i.e. under or barely at AA on the surfaces it actually renders on. This
    // value clears 4.5:1 on white, warm-light, the neu card and every status
    // tint. The guide's own §14 targets WCAG 2.2 AA, so the hex loses.
    slate: "#606F85",
    white: "#FFFFFF",
    // warm-light: primary light background.
    mist: "#F8FAFC",
    ink: "#081522",
    // safe-blue: primary action and trust.
    info: "#0B6BFF",
    border: "#E2E8F0",
    divider: "#EEF2F6",
  },
};

/** Palette for the active brand. */
export const palette: BrandPalette = PALETTES[ACTIVE_BRAND];

/**
 * SUBVENIO's named roles from the guide, for code that wants to say what it
 * means rather than which historical slot it is borrowing.
 */
export const roles = {
  safeBlue: PALETTES.subvenio.info,
  deepNavy: PALETTES.subvenio.navy,
  hopeOrange: PALETTES.subvenio.gold,
  softTeal: PALETTES.subvenio.silver,
  warmLight: PALETTES.subvenio.mist,
} as const;
