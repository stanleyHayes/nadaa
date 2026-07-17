import { useState } from "react";
import { Button, Stack } from "@mui/material";
import { UserPlus, UsersRound } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { UserManagementPanel } from "../UserManagementPanel";
import { ViewIntro } from "../primitives";

export function UsersView({ data }: { data: AdminData }) {
  const [createOpen, setCreateOpen] = useState(false);

  // Closing the dialog also discards the one-time credentials, so the
  // temporary password is only ever visible inside this dialog.
  const closeCreate = () => {
    setCreateOpen(false);
    data.dismissCreatedCredentials();
  };

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={UsersRound}
        title="Users"
        description="Authority users, their roles and MFA state, and provisioning of new access."
        action={
          <Button
            variant="contained"
            startIcon={<UserPlus size={18} />}
            onClick={() => setCreateOpen(true)}
          >
            Create user
          </Button>
        }
      />
      <UserManagementPanel
        actionResult={data.actionResult}
        agencies={data.agencies}
        busy={data.createBusy}
        createdCredentials={data.createdCredentials}
        form={data.userForm}
        onClose={closeCreate}
        onFormChange={data.onFieldChange}
        onSelectChange={data.onFieldChange}
        onSubmit={data.createUser}
        open={createOpen}
        users={data.users}
      />
    </Stack>
  );
}
