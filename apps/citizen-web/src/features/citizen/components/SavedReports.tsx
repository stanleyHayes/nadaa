import { Chip, Paper, Stack, Typography } from "@mui/material";
import { BookmarkCheck } from "lucide-react";
import type { SavedReport } from "../session";
import { formatDateTime, hazardLabel } from "../utils";
import type { HazardType } from "@nadaa/shared-types";
import { PageHeader } from "./PageHeader";

type SavedReportsProps = {
  reports: SavedReport[];
};

/** Convenience list of reports saved while signed in. Never gates anything. */
export function SavedReports({ reports }: SavedReportsProps) {
  if (reports.length === 0) {
    return null;
  }

  return (
    <Paper className="surface saved-reports">
      <PageHeader
        icon={BookmarkCheck}
        title="Your saved reports"
        subtitle="Kept on this device so you can follow them up."
        tone="green"
        as="h3"
      />
      <Stack spacing={1.25}>
        {reports.map((report) => (
          <Paper
            variant="outlined"
            className="shelter-row"
            key={report.reference}
          >
            <Stack
              direction="row"
              justifyContent="space-between"
              spacing={1}
              alignItems="center"
            >
              <div>
                <Typography variant="subtitle2">{report.reference}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {hazardLabel(report.hazard as HazardType)} ·{" "}
                  {formatDateTime(report.at)}
                </Typography>
              </div>
              <Chip
                size="small"
                label={report.priorityReview ? "Priority review" : "Submitted"}
                color={report.priorityReview ? "warning" : "success"}
              />
            </Stack>
          </Paper>
        ))}
      </Stack>
    </Paper>
  );
}

export default SavedReports;
