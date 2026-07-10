import { Stack } from "@mui/material";
import { UsersRound } from "lucide-react";
import type { AdminData } from "../../useAdminData";
import { UserManagementPanel } from "../UserManagementPanel";
import { ViewIntro } from "../primitives";

export function UsersView({ data }: { data: AdminData }) {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={UsersRound}
        title="Users"
        description="Authority users, their roles and MFA state, and provisioning of new access."
      />
      <UserManagementPanel
        actionResult={data.actionResult}
        agencies={data.agencies}
        busy={data.createBusy}
        form={data.userForm}
        onFormChange={data.onFieldChange}
        onSelectChange={data.onFieldChange}
        onSubmit={data.createUser}
        users={data.users}
      />
    </Stack>
  );
}
