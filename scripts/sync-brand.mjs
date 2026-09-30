#!/usr/bin/env node
/**
 * Keep the static brand surfaces in step with packages/brand.
 *
 * index.html, site.webmanifest and the Expo app.json files are not rendered
 * through React, so they cannot read ACTIVE_BRAND at runtime the way every
 * component does. Without this they drift: flipping the identity would change
 * every rendered string and leave the document title, PWA name and app name
 * announcing the previous brand.
 *
 *   node scripts/sync-brand.mjs           rewrite the static surfaces
 *   node scripts/sync-brand.mjs --check   fail if any are out of step (CI)
 *
 * Only brand-dependent VALUES are touched. The `%NADAA_SITE_ORIGIN%` build
 * placeholder, `--nadaa-*` properties, `NADAA_*` variables and the `@nadaa/*`
 * scope are identifiers and are never rewritten — see identity.ts.
 */
import { readFileSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const check = process.argv.includes("--check");

/** Read the active identity without needing a TypeScript loader. */
function activeIdentity() {
  const src = readFileSync(join(root, "packages/brand/src/identity.ts"), "utf8");
  const active = src.match(/ACTIVE_BRAND: BrandId = "([a-z]+)"/)?.[1];
  if (!active) throw new Error("could not read ACTIVE_BRAND from identity.ts");
  const block = src.split(`  ${active}: {`)[1];
  if (!block) throw new Error(`no identity block for "${active}"`);
  const field = (k) => block.match(new RegExp(`${k}: "([^"]*)"`))?.[1];
  const id = {
    id: active,
    name: field("name"),
    fullName: field("fullName"),
    tagline: field("tagline"),
  };
  for (const [k, v] of Object.entries(id)) {
    if (!v) throw new Error(`identity "${active}" is missing ${k}`);
  }
  const palette = readFileSync(join(root, "packages/brand/src/palettes.ts"), "utf8");
  const pblock = palette.split(`  ${active}: {`)[1];
  id.navy = pblock?.match(/navy: "(#[0-9A-Fa-f]{6})"/)?.[1];
  if (!id.navy) throw new Error(`palette "${active}" is missing navy`);
  return id;
}

const APPS = ["marketing-web", "citizen-web", "authority-dashboard",
               "dispatcher-web", "admin-web", "agency-web"];
const MOBILE = [["citizen-mobile", "Citizen"], ["dispatcher-mobile", "Dispatcher"]];

const brand = activeIdentity();
const drift = [];

function apply(path, next) {
  const current = readFileSync(join(root, path), "utf8");
  if (current === next) return;
  if (check) drift.push(path);
  else writeFileSync(join(root, path), next);
}

for (const app of APPS) {
  // site.webmanifest — name/short_name/theme are brand-dependent.
  const mp = `apps/${app}/public/site.webmanifest`;
  const m = JSON.parse(readFileSync(join(root, mp), "utf8"));
  m.name = brand.name;
  m.short_name = brand.name;
  m.theme_color = brand.navy.toUpperCase();
  m.background_color = brand.navy.toUpperCase();
  apply(mp, JSON.stringify(m, null, 2) + "\n");

  // index.html — the splash wordmark, tagline and theme-color.
  const hp = `apps/${app}/index.html`;
  let h = readFileSync(join(root, hp), "utf8");
  h = h.replace(/(<p class="nadaa-splash__word">)[^<]*(<\/p>)/, `$1${brand.name}$2`);
  h = h.replace(/(<p class="nadaa-splash__slogan">)[^<]*(<\/p>)/, `$1${brand.tagline}$2`);
  h = h.replace(/(<meta name="theme-color" content=")[^"]*(")/, `$1${brand.navy.toLowerCase()}$2`);
  h = h.replace(/(aria-label="Loading )[^"]*(")/, `$1${brand.name}$2`);
  apply(hp, h);
}

for (const [app, label] of MOBILE) {
  const p = `apps/${app}/app.json`;
  const c = JSON.parse(readFileSync(join(root, p), "utf8"));
  const e = c.expo;
  e.name = `${brand.name} ${label}`;
  e.backgroundColor = brand.navy.toUpperCase();
  e.splash.backgroundColor = brand.navy.toUpperCase();
  e.android.adaptiveIcon.backgroundColor = brand.navy.toUpperCase();
  apply(p, JSON.stringify(c, null, 2) + "\n");
}

if (check && drift.length) {
  console.error(
    `Static brand surfaces are out of step with ACTIVE_BRAND="${brand.id}":\n` +
    drift.map((d) => `  ${d}`).join("\n") +
    `\n\nRun: node scripts/sync-brand.mjs`,
  );
  process.exit(1);
}
console.log(
  check
    ? `brand surfaces in step with "${brand.id}" (${brand.name})`
    : `synced ${APPS.length} web + ${MOBILE.length} mobile surfaces to "${brand.id}" (${brand.name})`,
);
