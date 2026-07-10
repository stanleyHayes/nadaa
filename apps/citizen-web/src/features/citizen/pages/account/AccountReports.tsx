import { Button, Chip, Paper, Stack, Typography } from "@mui/material";
import { FileText, Siren } from "lucide-react";
import { Link } from "react-router-dom";
import type { HazardType } from "@nadaa/shared-types";
import {
  DataTable,
  PageHeader,
  type DataTableColumn,
} from "../../components";
import { urgencyOptions } from "../../data";
import { useCitizenSession, type SavedReport } from "../../session";
import { formatDateTime, hazardLabel } from "../../utils";

const urgencyLabel = (value: string) =>
  urgencyOptions.find((option) => option.value === value)?.label ?? value;

/** Columns for the signed-in citizen's own report history. */
const reportColumns: DataTableColumn<SavedReport>[] = [
  {
    key: "reference",
    label: "Reference",
    render: (report) => (
      <Typography variant="body2" sx={{ fontWeight: 700 }}>
        {report.reference}
      </Typography>
    ),
  },
  {
    key: "hazard",
    label: "Hazard",
    render: (report) => hazardLabel(report.hazard as HazardType),
  },
  {
    key: "urgency",
    label: "Urgency",
    render: (report) => urgencyLabel(report.urgency),
  },
  {
    key: "at",
    label: "Submitted",
    render: (report) => formatDateTime(report.at),
  },
  {
    key: "status",
    label: "Status",
    align: "right",
    render: (report) => (
      <Chip
        size="small"
        label={report.priorityReview ? "Priority review" : "Submitted"}
        color={report.priorityReview ? "warning" : "success"}
      />
    ),
  },
];

/** My reports (route `/account/reports`). Read-only history via the shared DataTable. */
export function AccountReports() {
  const { session, savedReports } = useCitizenSession();

  if (!session) {
    return null;
  }

  return (
    <Stack spacing={2.5} className="account-section">
      <Paper className="surface" component="section">
        <PageHeader
          icon={FileText}
          title="My reports"
          subtitle="Every incident you've reported from this device, newest first."
          tone="green"
          as="h2"
          action={
            <Button
              component={Link}
              to="/report"
              variant="outlined"
              size="small"
              startIcon={<Siren size={16} />}
            >
              File a report
            </Button>
          }
        />
      </Paper>

      <DataTable
        rows={savedReports}
        columns={reportColumns}
        getRowKey={(report) => report.reference}
        searchOf={(report) =>
          `${report.reference} ${hazardLabel(report.hazard as HazardType)} ${urgencyLabel(report.urgency)}`
        }
        searchPlaceholder="Search by reference or hazard"
        emptyMessage="You haven't filed any reports yet — filing one keeps a copy here."
        pageSize={8}
      />
    </Stack>
  );
}

export default AccountReports;
