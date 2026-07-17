import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath, URL } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv, type Plugin } from "vite";

// Canonical/OG/sitemap/robots origin: set VITE_SITE_ORIGIN per deployment
// (e.g. a staging alias) and the built index.html, robots.txt, and
// sitemap.xml follow; unset keeps the production origin.
const defaultSiteOrigin = "https://nadaa.gov.gh";

function siteOriginPlugin(siteOrigin: string): Plugin {
  let outDir = "dist";
  return {
    name: "nadaa-site-origin",
    configResolved: (config) => {
      outDir = config.build.outDir;
    },
    transformIndexHtml: (html) =>
      html.replaceAll("%NADAA_SITE_ORIGIN%", siteOrigin),
    // public/ files are copied verbatim, so rewrite their origins on disk.
    closeBundle: async () => {
      for (const fileName of ["robots.txt", "sitemap.xml"]) {
        const filePath = join(outDir, fileName);
        try {
          const content = await readFile(filePath, "utf8");
          await writeFile(
            filePath,
            content.replaceAll(defaultSiteOrigin, siteOrigin),
          );
        } catch {
          // Not emitted in this build — nothing to rewrite.
        }
      }
    },
  };
}

export default defineConfig(({ mode }) => {
  const siteOrigin = (
    loadEnv(mode, process.cwd(), "").VITE_SITE_ORIGIN ?? defaultSiteOrigin
  ).replace(/\/+$/, "");
  return {
    plugins: [react(), siteOriginPlugin(siteOrigin)],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
  };
});
