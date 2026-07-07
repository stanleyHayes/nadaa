import type { ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { KeyRound, UserPlus } from "lucide-react";
import type {
  AdminActionResult,
  AdminUserFormState,
  ManagedAgency,
  ManagedAgencyUser,
} from "../../data/types";
import { formatDateTime, roleLabel, roleOptions } from "../../lib/utils";
import { SectionHeader } from "../shared";

export function UserManagementPanel({
  actionResult,
  agencies,
  busy,
  form,
  onFormChange,
  onSelectChange,
  onSubmit,
  users,
}: {
  agencies: ManagedAgency[];
  users: ManagedAgencyUser[];
  form: AdminUserFormState;
  busy: boolean;
  actionResult?: AdminActionResult;
  onFormChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSelectChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSubmit: () => void;
}) {
  return (
    <Grid container spacing={2}>
      <Grid size={{ xs: 12, lg: 8 }}>
        <Paper className="surface">
          <SectionHeader
            eyebrow="Authority access"
            title="Users, roles, and MFA state"
            icon={<KeyRound size={22} color="#0D1B3D" />}
          />
          <Box className="admin-table">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>User</TableCell>
                  <TableCell>Role</TableCell>
                  <TableCell>Agency</TableCell>
                  <TableCell>MFA</TableCell>
                  <TableCell>Last login</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <Typography fontWeight={800}>{user.name}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {user.email}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        className="role-chip"
                        size="small"
                        color={
                          user.role === "system_admin" ? "primary" : "default"
                        }
                        label={roleLabel(user.role)}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {user.agency.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {user.accessScope}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={user.mfaEnabled ? "success" : "warning"}
                        label={user.mfaEnabled ? "Enabled" : "Setup pending"}
                      />
                    </TableCell>
                    <TableCell>{formatDateTime(user.lastLoginAt)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        </Paper>
      </Grid>
      <Grid size={{ xs: 12, lg: 4 }}>
        <Paper className="user-form">
          <SectionHeader
            eyebrow="Create user"
            title="Provision authority access"
            icon={<UserPlus size={22} color="#118D4E" />}
          />
          <Stack spacing={1.5}>
            {actionResult ? (
              <Alert severity={actionResult.severity}>
                {actionResult.message}
              </Alert>
            ) : null}
            <TextField
              name="name"
              label="Full name"
              size="small"
              value={form.name}
              onChange={onFormChange}
            />
            <TextField
              name="email"
              label="Email"
              size="small"
              value={form.email}
              onChange={onFormChange}
            />
            <TextField
              name="phone"
              label="Phone"
              size="small"
              value={form.phone}
              onChange={onFormChange}
            />
            <TextField
              select
              name="agencyId"
              label="Agency"
              size="small"
              value={form.agencyId}
              onChange={onSelectChange}
            >
              {agencies.map((agency) => (
                <MenuItem key={agency.id} value={agency.id}>
                  {agency.name}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              select
              name="role"
              label="Role"
              size="small"
              value={form.role}
              onChange={onSelectChange}
            >
              {roleOptions.map((role) => (
                <MenuItem key={role} value={role}>
                  {roleLabel(role)}
                </MenuItem>
              ))}
            </TextField>
            <Button
              variant="contained"
              startIcon={<UserPlus size={18} />}
              disabled={busy}
              onClick={onSubmit}
            >
              {busy ? "Creating" : "Create user"}
            </Button>
          </Stack>
        </Paper>
      </Grid>
    </Grid>
  );
}
