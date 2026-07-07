import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

const appDir = "apps/dispatcher-mobile";
const requiredFiles = [
  "App.tsx",
  "app.json",
  "package.json",
  "assets/nadaa-logo.png",
  "src/app/DispatcherMobileApp.tsx",
  "src/app/config.ts",
  "src/app/navigation.ts",
  "src/app/theme.ts",
  "src/features/dispatcher-mobile/api.ts",
  "src/features/dispatcher-mobile/data.ts",
  "src/features/dispatcher-mobile/offline.ts",
  "src/features/dispatcher-mobile/permissions.ts",
  "src/features/dispatcher-mobile/useDispatcherMobileState.ts",
  "src/features/dispatcher-mobile/screens/QueueScreen.tsx",
  "src/features/dispatcher-mobile/screens/DetailScreen.tsx",
  "src/features/dispatcher-mobile/screens/ActionScreen.tsx",
  "src/features/dispatcher-mobile/screens/CapacityScreen.tsx",
  "src/features/dispatcher-mobile/screens/ProfileScreen.tsx",
];

for (const file of requiredFiles) {
  const path = join(appDir, file);
  if (!existsSync(path)) {
    throw new Error(`missing dispatcher mobile file: ${path}`);
  }
}

const packageJson = JSON.parse(
  readFileSync(join(appDir, "package.json"), "utf8"),
);
if (packageJson.name !== "@nadaa/dispatcher-mobile") {
  throw new Error("dispatcher mobile package name is incorrect");
}
for (const script of ["start", "typecheck", "build"]) {
  if (!packageJson.scripts?.[script]) {
    throw new Error(`dispatcher mobile missing script: ${script}`);
  }
}

const appJson = JSON.parse(readFileSync(join(appDir, "app.json"), "utf8"));
if (appJson.expo?.slug !== "nadaa-dispatcher") {
  throw new Error("dispatcher mobile Expo slug is incorrect");
}

const shell = readFileSync(
  join(appDir, "src/app/DispatcherMobileApp.tsx"),
  "utf8",
);
for (const expected of [
  "QueueScreen",
  "DetailScreen",
  "ActionScreen",
  "CapacityScreen",
  "ProfileScreen",
  "nadaa-logo.png",
]) {
  if (!shell.includes(expected)) {
    throw new Error(`dispatcher mobile shell missing ${expected}`);
  }
}

const offline = readFileSync(
  join(appDir, "src/features/dispatcher-mobile/offline.ts"),
  "utf8",
);
for (const expected of [
  "readIncidentCache",
  "writeIncidentCache",
  "readSession",
  "readCapacityCache",
  "writeCapacityCache",
]) {
  if (!offline.includes(expected)) {
    throw new Error(`dispatcher mobile offline primitive missing ${expected}`);
  }
}

const queue = readFileSync(
  join(appDir, "src/features/dispatcher-mobile/screens/QueueScreen.tsx"),
  "utf8",
);
for (const expected of [
  "Incident queue",
  "Dispatch triage",
  "refreshQueue",
  "selectIncident",
]) {
  if (!queue.includes(expected)) {
    throw new Error(`dispatcher mobile queue screen missing ${expected}`);
  }
}

const action = readFileSync(
  join(appDir, "src/features/dispatcher-mobile/screens/ActionScreen.tsx"),
  "utf8",
);
for (const expected of [
  "Update status",
  "Assign incident",
  "submitStatusUpdate",
  "submitAssignment",
  "submitTimelineNote",
]) {
  if (!action.includes(expected)) {
    throw new Error(`dispatcher mobile action screen missing ${expected}`);
  }
}

const api = readFileSync(
  join(appDir, "src/features/dispatcher-mobile/api.ts"),
  "utf8",
);
for (const expected of [
  "authorityHeaders",
  "fetchIncidentQueue",
  "updateIncidentStatus",
  "assignIncident",
  "fetchHospitalCapacity",
]) {
  if (!api.includes(expected)) {
    throw new Error(`dispatcher mobile API missing ${expected}`);
  }
}

console.log("dispatcher-mobile scaffold OK");
