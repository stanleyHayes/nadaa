import { useEffect, useState, type ChangeEvent } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
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
import { KeyRound, UserPlus, X } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type {
  AdminActionResult,
  AdminUserFormState,
  ManagedAgency,
  ManagedAgencyUser,
} from "../types";
import { formatDateTime, roleLabel, roleOptions } from "../utils";
import { EmptyState, SectionHeader } from "./shared";

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
  const [open, setOpen] = useState(false);

  const nameInvalid = !form.name.trim();
  const emailInvalid = !form.email.includes("@");
  const phoneInvalid = !form.phone.startsWith("+233") || form.phone.length < 8;
  const agencyInvalid = !form.agencyId;

  // Close the dialog once a create succeeds; keep it open on validation or
  // network failure so the operator does not lose their input.
  useEffect(() => {
    if (actionResult?.severity === "success") {
      setOpen(false);
    }
  }, [actionResult]);

  const closeDialog = () => setOpen(false);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Authority access"
        title="Users, roles, and MFA state"
        icon={<KeyRound size={22} color={nadaaBrand.colors.navy} />}
        action={
          <Button
            variant="contained"
            startIcon={<UserPlus size={18} />}
            onClick={() => setOpen(true)}
          >
            Create user
          </Button>
        }
      />
      {actionResult?.severity === "success" ? (
        <Alert severity="success" sx={{ mb: 2 }}>
          {actionResult.message}
        </Alert>
      ) : null}
      {users.length ? (
        <Box
          className="admin-table"
          tabIndex={0}
          aria-label="User management table, scroll horizontally on small screens"
        >
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
                    <Typography variant="body2">{user.agency.name}</Typography>
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
      ) : (
        <EmptyState
          title="No users yet"
          detail="Provision authority access with Create user. New users appear here once created."
        />
      )}

      <Dialog open={open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          Provision authority access
          <IconButton aria-label="Close" size="small" onClick={closeDialog}>
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            {actionResult && actionResult.severity !== "success" ? (
              <Alert severity={actionResult.severity}>
                {actionResult.message}
              </Alert>
            ) : null}
            <TextField
              id="user-name"
              name="name"
              label="Full name"
              size="small"
              required
              value={form.name}
              onChange={onFormChange}
              error={nameInvalid}
              inputProps={{ "aria-invalid": nameInvalid }}
            />
            <TextField
              id="user-email"
              name="email"
              label="Email"
              type="email"
              size="small"
              required
              value={form.email}
              onChange={onFormChange}
              error={emailInvalid}
              inputProps={{ "aria-invalid": emailInvalid }}
            />
            <TextField
              id="user-phone"
              name="phone"
              label="Phone"
              type="tel"
              size="small"
              required
              value={form.phone}
              onChange={onFormChange}
              error={phoneInvalid}
              inputProps={{ "aria-invalid": phoneInvalid }}
            />
            <TextField
              select
              id="user-agency"
              name="agencyId"
              label="Agency"
              size="small"
              required
              value={form.agencyId}
              onChange={onSelectChange}
              error={agencyInvalid}
              inputProps={{ "aria-invalid": agencyInvalid }}
            >
              {agencies.map((agency) => (
                <MenuItem key={agency.id} value={agency.id}>
                  {agency.name}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              select
              id="user-role"
              name="role"
              label="Role"
              size="small"
              required
              value={form.role}
              onChange={onSelectChange}
            >
              {roleOptions.map((role) => (
                <MenuItem key={role} value={role}>
                  {roleLabel(role)}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={closeDialog} disabled={busy}>
            Cancel
          </Button>
          <Button
            variant="contained"
            startIcon={<UserPlus size={18} />}
            disabled={busy}
            onClick={onSubmit}
          >
            {busy ? "Creating" : "Create user"}
          </Button>
        </DialogActions>
      </Dialog>
    </Paper>
  );
}
