import { Box, Chip, Paper, Stack, Typography } from "@mui/material";
import { CheckCircle2 } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { ManagedAgencyUser } from "../types";
import { roleLabel } from "../utils";
import { EmptyState, SectionHeader } from "./shared";

export function MfaSupportPanel({ users }: { users: ManagedAgencyUser[] }) {
  const pendingUsers = users.filter((user) => !user.mfaEnabled);

  return (
    <Paper className="surface">
      <SectionHeader
        eyebrow="MFA support"
        title="Authority users waiting on setup"
        icon={<CheckCircle2 size={22} color={nadaaBrand.colors.green} />}
      />
      {pendingUsers.length ? (
        <Stack spacing={1.5}>
          {pendingUsers.map((user) => (
            <Stack
              key={user.id}
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              gap={1}
            >
              <Box>
                <Typography fontWeight={800}>{user.name}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {user.agency.name} / {roleLabel(user.role)}
                </Typography>
              </Box>
              <Chip color="warning" label="Setup pending" />
            </Stack>
          ))}
        </Stack>
      ) : (
        <EmptyState
          title="MFA coverage complete"
          detail="Every authority user in the current admin view has completed MFA setup."
        />
      )}
    </Paper>
  );
}
