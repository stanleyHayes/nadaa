import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

const appDir = "apps/citizen-mobile";
const requiredFiles = [
  "App.tsx",
  "app.json",
  "package.json",
  "assets/nadaa-logo.png",
  "src/app/CitizenMobileApp.tsx",
  "src/app/config.ts",
  "src/app/navigation.ts",
  "src/app/theme.ts",
  "src/features/citizen-mobile/api.ts",
  "src/features/citizen-mobile/data.ts",
  "src/features/citizen-mobile/offline.ts",
  "src/features/citizen-mobile/permissions.ts",
  "src/features/citizen-mobile/useCitizenMobileState.ts",
  "src/features/citizen-mobile/screens/HomeScreen.tsx",
  "src/features/citizen-mobile/screens/AlertsScreen.tsx",
  "src/features/citizen-mobile/screens/ReportScreen.tsx",
  "src/features/citizen-mobile/screens/CommunityScreen.tsx",
  "src/features/citizen-mobile/screens/GuidesScreen.tsx",
  "src/features/citizen-mobile/screens/SupportScreen.tsx",
];

for (const file of requiredFiles) {
  const path = join(appDir, file);
  if (!existsSync(path)) {
    throw new Error(`missing citizen mobile file: ${path}`);
  }
}

const packageJson = JSON.parse(
  readFileSync(join(appDir, "package.json"), "utf8"),
);
if (packageJson.name !== "@nadaa/citizen-mobile") {
  throw new Error("citizen mobile package name is incorrect");
}
for (const script of ["start", "typecheck", "build"]) {
  if (!packageJson.scripts?.[script]) {
    throw new Error(`citizen mobile missing script: ${script}`);
  }
}

const appJson = JSON.parse(readFileSync(join(appDir, "app.json"), "utf8"));
if (appJson.expo?.slug !== "nadaa-citizen") {
  throw new Error("citizen mobile Expo slug is incorrect");
}

const shell = readFileSync(
  join(appDir, "src/app/CitizenMobileApp.tsx"),
  "utf8",
);
for (const expected of [
  "HomeScreen",
  "AlertsScreen",
  "ReportScreen",
  "CommunityScreen",
  "GuidesScreen",
  "SupportScreen",
  "nadaa-logo.png",
]) {
  if (!shell.includes(expected)) {
    throw new Error(`citizen mobile shell missing ${expected}`);
  }
}

const offline = readFileSync(
  join(appDir, "src/features/citizen-mobile/offline.ts"),
  "utf8",
);
for (const expected of [
  "readGuideCache",
  "writeReportDraft",
  "readSession",
  "readVolunteerProfile",
  "writeVolunteerTasks",
]) {
  if (!offline.includes(expected)) {
    throw new Error(`citizen mobile offline primitive missing ${expected}`);
  }
}

const community = readFileSync(
  join(appDir, "src/features/citizen-mobile/screens/CommunityScreen.tsx"),
  "utf8",
);
for (const expected of [
  "Volunteer assignments",
  "Safety rules",
  "Submit observation",
  "Request authority escalation",
]) {
  if (!community.includes(expected)) {
    throw new Error(`citizen mobile community screen missing ${expected}`);
  }
}

const api = readFileSync(
  join(appDir, "src/features/citizen-mobile/api.ts"),
  "utf8",
);
for (const expected of [
  "registerVolunteerProfile",
  "fetchVolunteerTasks",
  "updateVolunteerTaskStatus",
  "submitVolunteerObservation",
]) {
  if (!api.includes(expected)) {
    throw new Error(`citizen mobile volunteer API missing ${expected}`);
  }
}

console.log("citizen-mobile scaffold OK");
