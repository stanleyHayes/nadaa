import {
  Box,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import { X } from "lucide-react";
import { type ReactNode } from "react";

export type DetailField = {
  label: string;
  value: ReactNode;
  /** Span the full row width — use for long text (descriptions, addresses). */
  full?: boolean;
};

type DetailDialogProps = {
  open: boolean;
  onClose: () => void;
  title: ReactNode;
  subtitle?: ReactNode;
  fields: DetailField[];
  /** Optional footer actions, e.g. a link to a full detail page. */
  actions?: ReactNode;
};

/**
 * Read-only record detail shown in a dialog — the light-detail half of the
 * list/detail split (a row opens this instead of expanding inline). Renders the
 * record as a responsive definition list; use a dedicated page when the detail
 * is heavy rather than growing this.
 */
export function DetailDialog({
  open,
  onClose,
  title,
  subtitle,
  fields,
  actions,
}: DetailDialogProps) {
  return (
    <Dialog fullWidth maxWidth="sm" onClose={onClose} open={open} scroll="paper">
      <DialogTitle
        sx={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: "space-between",
          gap: 2,
        }}
      >
        <Box>
          <Typography
            component="span"
            sx={{ display: "block", fontWeight: 800, fontSize: "1.1rem" }}
          >
            {title}
          </Typography>
          {subtitle ? (
            <Typography
              component="span"
              sx={{
                display: "block",
                color: "text.secondary",
                fontSize: "0.9rem",
                fontWeight: 600,
              }}
            >
              {subtitle}
            </Typography>
          ) : null}
        </Box>
        <IconButton aria-label="Close" onClick={onClose} size="small">
          <X size={18} />
        </IconButton>
      </DialogTitle>
      <DialogContent dividers>
        <Box
          component="dl"
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr" },
            gap: 2,
            m: 0,
          }}
        >
          {fields.map((field) => (
            <Box
              key={field.label}
              sx={{ gridColumn: field.full ? "1 / -1" : undefined }}
            >
              <Typography
                component="dt"
                sx={{
                  color: "text.secondary",
                  fontSize: "0.72rem",
                  fontWeight: 800,
                  textTransform: "uppercase",
                  letterSpacing: "0.04em",
                  mb: 0.5,
                }}
              >
                {field.label}
              </Typography>
              <Typography
                component="dd"
                sx={{
                  m: 0,
                  fontSize: "0.95rem",
                  fontWeight: 600,
                  color: "text.primary",
                  wordBreak: "break-word",
                }}
              >
                {field.value ?? "—"}
              </Typography>
            </Box>
          ))}
        </Box>
        {actions ? (
          <Stack
            direction="row"
            spacing={1}
            sx={{ mt: 3, justifyContent: "flex-end" }}
          >
            {actions}
          </Stack>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
