/**
 * Swappable brand identity.
 *
 * The product is mid-rebrand from NADAA to SUBVENIO. Per the product blueprint,
 * SUBVENIO is the working product identity while trademark, domain and
 * app-store clearance is completed — it is NOT yet a cleared legal mark, and
 * there is an existing software company plus a US filing on the name. So the
 * identity has to be reversible on one line rather than baked across the tree.
 *
 * `ACTIVE_BRAND` is that line. Everything user-facing resolves through it:
 * copy via `brand`, colour via `palettes`, CSS via the `[data-brand]` blocks in
 * brand.css. Flipping it back to "nadaa" restores the previous identity whole.
 *
 * What does NOT live here, and must NOT be renamed with the brand:
 *   - wire identifiers: the `X-NADAA-*` request headers and `NADAA_*` env vars
 *     that services authenticate each other with. Renaming them breaks
 *     service-to-service auth and needs a coordinated redeploy.
 *   - the `@nadaa/*` npm scope and `--nadaa-*` CSS custom properties, which are
 *     internal identifiers. The CSS slots are re-pointed per brand below; their
 *     names stay put so the ~5,000 existing references keep resolving.
 */

export type BrandId = "nadaa" | "subvenio";

/** The identity the product currently ships under. One line to revert. */
export const ACTIVE_BRAND: BrandId = "subvenio";

export type BrandIdentity = {
  /** Wordmark, as displayed. */
  readonly name: string;
  /** Expanded descriptor for titles, manifests and legal surfaces. */
  readonly fullName: string;
  /** Primary market. */
  readonly country: string;
  /** Brand line. Not an operational status — see `voice` in the motion guide. */
  readonly tagline: string;
  /** Emergency number shown to citizens. Ghana: 112. */
  readonly supportLine: string;
  /** Logo asset basenames served from each web app's /brand/ directory. */
  readonly assets: {
    readonly icon: string;
    readonly horizontalLight: string;
    readonly horizontalDark: string;
    readonly verticalLight: string;
    readonly verticalDark: string;
  };
};

const IDENTITIES: Record<BrandId, BrandIdentity> = {
  nadaa: {
    name: "NADAA",
    fullName: "National Disaster Alert & Response Platform",
    country: "Ghana",
    tagline: "Be Aware. Be Prepared. Be Safe.",
    supportLine: "112",
    // These files are NOT on disk — they were removed in 7257e76 when the
    // SUBVENIO system replaced them. Reverting the brand therefore needs the
    // artwork back as well as the flag:
    //     git checkout 7257e76^ -- 'apps/*/public/brand/nadaa-logo.png'
    // Everything else about the revert is this one constant; the logos are the
    // exception because they are binaries, not values.
    assets: {
      icon: "/brand/nadaa-logo.png",
      horizontalLight: "/brand/nadaa-logo.png",
      horizontalDark: "/brand/nadaa-logo.png",
      verticalLight: "/brand/nadaa-logo.png",
      verticalDark: "/brand/nadaa-logo.png",
    },
  },
  subvenio: {
    name: "SUBVENIO",
    fullName: "Emergency Alert & Response Platform",
    country: "Ghana",
    tagline: "Someone is coming.",
    supportLine: "112",
    assets: {
      icon: "/brand/subvenio-icon.png",
      horizontalLight: "/brand/subvenio-horizontal-light.png",
      horizontalDark: "/brand/subvenio-horizontal-dark.png",
      verticalLight: "/brand/subvenio-vertical-light.png",
      verticalDark: "/brand/subvenio-vertical-dark.png",
    },
  },
};

/** The active identity. Prefer this over the legacy `nadaaBrand` alias. */
export const brand: BrandIdentity = IDENTITIES[ACTIVE_BRAND];

/** Look up a non-active identity, e.g. to preview a switch. */
export function identityFor(id: BrandId): BrandIdentity {
  return IDENTITIES[id];
}
