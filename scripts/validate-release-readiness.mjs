import { readFileSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();

const requiredDocuments = [
  {
    path: "docs/uat.md",
    tokens: [
      "# UAT Plan",
      "## Feedback Capture",
      "## UAT Scripts",
      "## Exit Criteria",
      "UAT-001",
      "UAT-012",
    ],
  },
  {
    path: "docs/release-readiness.md",
    tokens: [
      "# Release Readiness",
      "## Release Candidate Gate",
      "## Acceptance Checklist",
      "## Release Notes Template",
      "## Rollback Plan",
      "## No-Go Criteria",
    ],
  },
  {
    path: "docs/user-guide.md",
    tokens: [
      "# User Guide And Training",
      "## Citizen Quick Guide",
      "## Dispatcher Quick Guide",
      "## Officer And Approver Quick Guide",
      "## Admin Quick Guide",
      "## Training Completion Checklist",
    ],
  },
  {
    path: "docs/beta-monitoring.md",
    tokens: [
      "# Beta Monitoring",
      "## Monitoring Cadence",
      "## Metrics Dashboard Definition",
      "## Daily Beta Review",
      "## Escalation Triggers",
    ],
  },
  {
    path: "docs/hypercare.md",
    tokens: [
      "# Hypercare",
      "## Hypercare Roles",
      "## Severity Response",
      "## Launch Room Checklist",
      "## Daily Hypercare Checklist",
      "## Support Handoff",
    ],
  },
];

for (const document of requiredDocuments) {
  const content = readFileSync(join(root, document.path), "utf8");

  for (const token of document.tokens) {
    if (!content.includes(token)) {
      throw new Error(
        `${document.path} is missing release-readiness token: ${token}`,
      );
    }
  }
}

const dashboardRecords = JSON.parse(
  readFileSync(
    join(root, "docs/project-dashboard/sample-records.json"),
    "utf8",
  ),
);
const readinessRecord = dashboardRecords.find(
  (record) => record.jiraKey === "NADAA-102",
);

if (!readinessRecord) {
  throw new Error(
    "docs/project-dashboard/sample-records.json is missing NADAA-102",
  );
}

for (const expected of [
  "docs/uat.md",
  "docs/release-readiness.md",
  "docs/user-guide.md",
  "docs/beta-monitoring.md",
  "docs/hypercare.md",
]) {
  if (!readinessRecord.notes.includes(expected)) {
    throw new Error(`NADAA-102 dashboard notes must reference ${expected}`);
  }
}

console.log(
  `Validated ${requiredDocuments.length} release-readiness documents.`,
);
