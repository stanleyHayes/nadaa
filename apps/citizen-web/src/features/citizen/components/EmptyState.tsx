import { Box, Stack, Typography } from "@mui/material";
import { type LucideIcon } from "lucide-react";
import { type ReactNode } from "react";

type EmptyStateProps = {
  icon: LucideIcon;
  title: string;
  description?: string;
  /** Optional call-to-action(s) rendered beneath the copy. */
  action?: ReactNode;
  /** Accent tone of the animated icon badge. */
  tone?: "navy" | "gold" | "green";
};

/**
 * Friendly empty state: a gently animated icon badge (float + halo, disabled
 * under reduced motion) over a title, description, and optional action. Used as
 * the DataTable empty slot so "no rows" reads as a designed moment, not blank text.
 */
export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  tone = "navy",
}: EmptyStateProps) {
  return (
    <Stack
      alignItems="center"
      className="empty-state"
      data-tone={tone}
      spacing={1.25}
      sx={{ py: { xs: 5, sm: 7 }, px: 3, textAlign: "center" }}
    >
      <span aria-hidden="true" className="empty-state__icon">
        <Icon size={30} strokeWidth={1.75} />
      </span>
      <Typography className="empty-state__title" variant="h6">
        {title}
      </Typography>
      {description ? (
        <Typography
          color="text.secondary"
          sx={{ maxWidth: 400 }}
          variant="body2"
        >
          {description}
        </Typography>
      ) : null}
      {action ? <Box sx={{ pt: 0.75 }}>{action}</Box> : null}
    </Stack>
  );
}
