import { Stack } from "@mui/material";
import { LockKeyhole } from "lucide-react";
import { RoleMatrixPanel } from "../RoleMatrixPanel";
import { ViewIntro } from "../primitives";

export function RolesView() {
  return (
    <Stack spacing={2.5}>
      <ViewIntro
        icon={LockKeyhole}
        title="Roles & access"
        description="The permission matrix for admin console, alert approval, and operational duties across every platform role."
      />
      <RoleMatrixPanel />
    </Stack>
  );
}
