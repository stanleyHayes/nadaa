import { type ChangeEvent } from "react";
import {
  Alert,
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
import { DataTable } from "./DataTable";

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
  const nameInvalid = !form.name.trim();
  const emailInvalid = !form.email.includes("@");
  const phoneInvalid = !form.phone.startsWith("+233") || form.phone.length < 8;
  const agencyInvalid = !form.agencyId;

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
      {users.length === 0 ? (
        <EmptyState
          title="No users yet"
          detail="Provision authority access with Create user. New users appear here once created."
        />
      ) : (
        <DataTable
          rows={users}
          getRowKey={(user) => user.id}
          searchOf={(user) => `${user.name} ${user.email}`}
          searchPlaceholder="Search name or email"
          filters={[
            {
              key: "role",
              label: "Role",
              options: Array.from(
                new Set(users.map((user) => roleLabel(user.role))),
              ),
              valueOf: (user) => roleLabel(user.role),
            },
            {
              key: "agency",
              label: "Agency",
              options: Array.from(
                new Set(users.map((user) => user.agency.name)),
              ),
              valueOf: (user) => user.agency.name,
            },
            {
              key: "mfa",
              label: "MFA",
              options: Array.from(
                new Set(
                  users.map((user) =>
                    user.mfaEnabled ? "Enabled" : "Setup pending",
                  ),
                ),
              ),
              valueOf: (user) =>
                user.mfaEnabled ? "Enabled" : "Setup pending",
            },
          ]}
          columns={[
            {
              key: "user",
              label: "User",
              render: (user) => (
                <>
                  <Typography sx={{ fontWeight: 800 }}>{user.name}</Typography>
                  <Typography
                    variant="caption"
                    sx={{ color: "text.secondary" }}
                  >
                    {user.email}
                  </Typography>
                </>
              ),
            },
            {
              key: "role",
              label: "Role",
              render: (user) => (
                <Chip
                  className="role-chip"
                  size="small"
                  color={user.role === "system_admin" ? "primary" : "default"}
                  label={roleLabel(user.role)}
                />
              ),
            },
            {
              key: "agency",
              label: "Agency",
              render: (user) => (
                <>
                  <Typography variant="body2">{user.agency.name}</Typography>
                  <Typography
                    variant="caption"
                    sx={{ color: "text.secondary" }}
                  >
                    {user.accessScope}
                  </Typography>
                </>
              ),
            },
            {
              key: "mfa",
              label: "MFA",
              render: (user) => (
                <Chip
                  className="status-chip"
                  size="small"
                  color={user.mfaEnabled ? "success" : "warning"}
                  label={user.mfaEnabled ? "Enabled" : "Setup pending"}
                />
              ),
            },
            {
              key: "lastLogin",
              label: "Last login",
              render: (user) => formatDateTime(user.lastLoginAt),
            },
          ]}
          emptyState={
            <EmptyState
              title="No matching users"
              detail="No users match the current search and filters. Adjust or clear them to see more."
            />
          }
        />
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
