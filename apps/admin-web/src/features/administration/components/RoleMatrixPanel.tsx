import {
  Box,
  Chip,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { LockKeyhole } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import { roleLabel, roleOptions } from "../utils";
import { SectionHeader } from "./shared";

export function RoleMatrixPanel() {
  const adminRoles = new Set(["system_admin", "agency_admin"]);
  const alertApprovalRoles = new Set([
    "system_admin",
    "agency_admin",
    "nadmo_officer",
  ]);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Role policy"
        title="Admin, alert, and operational permissions"
        icon={<LockKeyhole size={22} color={nadaaBrand.colors.navy} />}
      />
      <Box
        className="admin-table"
        tabIndex={0}
        aria-label="Role matrix table, scroll horizontally on small screens"
      >
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Role</TableCell>
              <TableCell>Admin console</TableCell>
              <TableCell>Alert approval</TableCell>
              <TableCell>Dispatcher operations</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {roleOptions.map((role) => (
              <TableRow key={role}>
                <TableCell>{roleLabel(role)}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={adminRoles.has(role) ? "success" : "default"}
                    label={adminRoles.has(role) ? "Permitted" : "Denied"}
                  />
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={alertApprovalRoles.has(role) ? "warning" : "default"}
                    label={alertApprovalRoles.has(role) ? "Reviewer" : "No"}
                  />
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    color={
                      role === "dispatcher" || role === "responder"
                        ? "primary"
                        : "default"
                    }
                    label={
                      role === "dispatcher" || role === "responder"
                        ? "Operational"
                        : "Scoped"
                    }
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Paper>
  );
}
