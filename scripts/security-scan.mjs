import { execFileSync } from "node:child_process";
import { readdir, readFile, stat } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const rootDir = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const failures = [];
let checks = 0;

const requiredServiceTokens = [
  "NADAA_ALLOWED_ORIGINS",
  "applySecurityHeaders",
  "X-Content-Type-Options",
  "Content-Security-Policy",
  "Strict-Transport-Security",
  "Cache-Control",
  "allowedOriginsFromEnv",
];

function record(condition, message) {
  checks += 1;
  if (!condition) {
    failures.push(message);
  }
}

async function readRepoFile(relativePath) {
  return readFile(path.join(rootDir, relativePath), "utf8");
}

async function fileExists(relativePath) {
  try {
    await stat(path.join(rootDir, relativePath));
    return true;
  } catch {
    return false;
  }
}

async function listChildDirs(relativePath) {
  const entries = await readdir(path.join(rootDir, relativePath), {
    withFileTypes: true,
  });

  return entries
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(relativePath, entry.name))
    .sort();
}

async function checkServiceHttpHardening() {
  const serviceDirs = await listChildDirs("services");

  for (const serviceDir of serviceDirs) {
    const mainPath = path.join(serviceDir, "main.go");
    if (!(await fileExists(mainPath))) {
      continue;
    }

    const source = await readRepoFile(mainPath);
    for (const token of requiredServiceTokens) {
      record(source.includes(token), `${mainPath} is missing ${token}`);
    }

    record(
      source.includes("Vary") && source.includes("Origin"),
      `${mainPath} must vary CORS responses by Origin when using an allowlist`,
    );
  }
}

async function checkDockerRuntimeUsers() {
  const appDirs = await listChildDirs("apps");
  const serviceDirs = await listChildDirs("services");
  const dockerfiles = [...appDirs, ...serviceDirs].map((directory) =>
    path.join(directory, "Dockerfile"),
  );

  for (const dockerfile of dockerfiles) {
    if (!(await fileExists(dockerfile))) {
      continue;
    }

    const source = await readRepoFile(dockerfile);
    record(
      /\nUSER\s+(?!root\b).+/u.test(`\n${source}`),
      `${dockerfile} must set a non-root runtime USER`,
    );
  }
}

async function checkEnvironmentGuardrails() {
  const gitignore = await readRepoFile(".gitignore");
  record(gitignore.includes(".env"), ".gitignore must ignore .env");
  record(gitignore.includes(".env.*"), ".gitignore must ignore .env.*");
  record(
    gitignore.includes("!.env.example"),
    ".gitignore must allow checked-in .env.example templates",
  );

  const trackedFiles = execFileSync("git", ["ls-files"], {
    cwd: rootDir,
    encoding: "utf8",
  })
    .split("\n")
    .filter(Boolean);

  const trackedEnvFiles = trackedFiles.filter((trackedFile) => {
    const basename = path.basename(trackedFile);
    return basename.startsWith(".env") && basename !== ".env.example";
  });

  record(
    trackedEnvFiles.length === 0,
    `tracked non-template env files are not allowed: ${trackedEnvFiles.join(", ")}`,
  );

  const stagingEnv = await readRepoFile("infra/staging/staging.env.example");
  record(
    stagingEnv.includes("NADAA_ALLOWED_ORIGINS"),
    "infra/staging/staging.env.example must document NADAA_ALLOWED_ORIGINS",
  );
}

async function checkSecurityDocsAndCi() {
  const packageJson = JSON.parse(await readRepoFile("package.json"));
  record(
    packageJson.scripts?.["security:scan"] === "node scripts/security-scan.mjs",
    "package.json must expose security:scan",
  );

  const ci = await readRepoFile(".github/workflows/ci.yml");
  record(ci.includes("pnpm security:scan"), "CI must run pnpm security:scan");

  const securityReportExists = await fileExists("docs/security-review.md");
  record(securityReportExists, "docs/security-review.md must exist");

  if (securityReportExists) {
    const securityReport = await readRepoFile("docs/security-review.md");
    record(
      securityReport.includes("NADAA-092"),
      "docs/security-review.md must identify the NADAA-092 review scope",
    );
    record(
      securityReport.includes("Residual Risks"),
      "docs/security-review.md must include residual risks",
    );
  }
}

await checkServiceHttpHardening();
await checkDockerRuntimeUsers();
await checkEnvironmentGuardrails();
await checkSecurityDocsAndCi();

if (failures.length > 0) {
  console.error("[security-scan] Failed security checks:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log(`[security-scan] Passed ${checks} repository security checks.`);
