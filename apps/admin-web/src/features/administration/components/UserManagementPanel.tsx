import { useMemo, useState, type ChangeEvent } from "react";
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
  InputAdornment,
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
import { KeyRound, Search, UserPlus, X } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { AgencyUserRole } from "@nadaa/shared-types";
import type {
  AdminActionResult,
  AdminUserFormState,
  ManagedAgency,
  ManagedAgencyUser,
} from "../types";
import { formatDateTime, roleLabel, roleOptions } from "../utils";
import { EmptyState, SectionHeader } from "./shared";

type RoleFilter = AgencyUserRole | "all";
type MfaFilter = "all" | "enabled" | "pending";

export function UserManagementPanel({
  actionResult,
  agencies,
  busy,
  form,
  onClose,
  onFormChange,
  onSelectChange,
  onSubmit,
  open,
  users,
}: {
  agencies: ManagedAgency[];
  users: ManagedAgencyUser[];
  form: AdminUserFormState;
  busy: boolean;
  open: boolean;
  actionResult?: AdminActionResult;
  onClose: () => void;
  onFormChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSelectChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onSubmit: () => void;
}) {
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("all");
  const [agencyFilter, setAgencyFilter] = useState<string>("all");
  const [mfaFilter, setMfaFilter] = useState<MfaFilter>("all");

  const nameInvalid = !form.name.trim();
  const emailInvalid = !form.email.includes("@");
  const phoneInvalid = !form.phone.startsWith("+233") || form.phone.length < 8;
  const agencyInvalid = !form.agencyId;

  // Agency options are derived from the loaded users so the filter always
  // matches what is actually present in the table.
  const agencyOptions = useMemo(() => {
    const seen = new Map<string, string>();
    for (const user of users) {
      if (!seen.has(user.agency.id)) {
        seen.set(user.agency.id, user.agency.name);
      }
    }
    return Array.from(seen, ([id, name]) => ({ id, name }));
  }, [users]);

  const filteredUsers = useMemo(() => {
    const query = search.trim().toLowerCase();
    return users.filter((user) => {
      if (roleFilter !== "all" && user.role !== roleFilter) {
        return false;
      }
      if (agencyFilter !== "all" && user.agency.id !== agencyFilter) {
        return false;
      }
      if (mfaFilter === "enabled" && !user.mfaEnabled) {
        return false;
      }
      if (mfaFilter === "pending" && user.mfaEnabled) {
        return false;
      }
      if (
        query &&
        !`${user.name} ${user.email}`.toLowerCase().includes(query)
      ) {
        return false;
      }
      return true;
    });
  }, [users, search, roleFilter, agencyFilter, mfaFilter]);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="Authority access"
        title="Users, roles, and MFA state"
        icon={<KeyRound size={22} color="var(--nadaa-navy)" />}
      />
      {actionResult?.severity === "success" ? (
        <Alert severity="success" sx={{ mb: 2 }}>
          {actionResult.message}
        </Alert>
      ) : null}
      <Box className="cc-table-toolbar">
        <TextField
          className="cc-table-toolbar__search"
          size="small"
          placeholder="Search name or email"
          aria-label="Search users by name or email"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <Search size={16} color="var(--nadaa-slate)" />
                </InputAdornment>
              ),
            }
          }}
        />
        <TextField
          className="cc-table-toolbar__filter"
          select
          size="small"
          label="Role"
          value={roleFilter}
          onChange={(event) => setRoleFilter(event.target.value as RoleFilter)}
        >
          <MenuItem value="all">All roles</MenuItem>
          {roleOptions.map((role) => (
            <MenuItem key={role} value={role}>
              {roleLabel(role)}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          className="cc-table-toolbar__filter"
          select
          size="small"
          label="Agency"
          value={agencyFilter}
          onChange={(event) => setAgencyFilter(event.target.value)}
        >
          <MenuItem value="all">All agencies</MenuItem>
          {agencyOptions.map((agency) => (
            <MenuItem key={agency.id} value={agency.id}>
              {agency.name}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          className="cc-table-toolbar__filter"
          select
          size="small"
          label="MFA"
          value={mfaFilter}
          onChange={(event) => setMfaFilter(event.target.value as MfaFilter)}
        >
          <MenuItem value="all">All MFA states</MenuItem>
          <MenuItem value="enabled">Enabled</MenuItem>
          <MenuItem value="pending">Setup pending</MenuItem>
        </TextField>
      </Box>
      {users.length === 0 ? (
        <EmptyState
          title="No users yet"
          detail="Provision authority access with Create user. New users appear here once created."
        />
      ) : filteredUsers.length === 0 ? (
        <EmptyState
          title="No matching users"
          detail="No users match the current search and filters. Adjust or clear them to see more."
        />
      ) : (
        <Box
          className="admin-table cc-datatable"
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
              {filteredUsers.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <Typography sx={{
                      fontWeight: 800
                    }}>{user.name}</Typography>
                    <Typography variant="caption" sx={{
                      color: "text.secondary"
                    }}>
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
                    <Typography variant="caption" sx={{
                      color: "text.secondary"
                    }}>
                      {user.accessScope}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      className="status-chip"
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
      )}
      <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          Provision authority access
          <IconButton aria-label="Close" size="small" onClick={onClose}>
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
              slotProps={{
                htmlInput: { "aria-invalid": nameInvalid }
              }}
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
              slotProps={{
                htmlInput: { "aria-invalid": emailInvalid }
              }}
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
              slotProps={{
                htmlInput: { "aria-invalid": phoneInvalid }
              }}
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
              slotProps={{
                htmlInput: { "aria-invalid": agencyInvalid }
              }}
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
          <Button onClick={onClose} disabled={busy}>
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
