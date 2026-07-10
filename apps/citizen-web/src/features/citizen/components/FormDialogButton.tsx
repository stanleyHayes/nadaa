import {
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
} from "@mui/material";
import { type LucideIcon, Plus, X } from "lucide-react";
import { type ReactNode, useState } from "react";
import { useCitizenSession } from "../session";

type FormDialogButtonProps = {
  /** Button label, e.g. "Report a missing person". */
  label: string;
  dialogTitle: string;
  icon?: LucideIcon;
  color?: "primary" | "secondary" | "error";
  /** The form; receives a `close` callback to dismiss the dialog on success. */
  children: (close: () => void) => ReactNode;
};

/**
 * A "Report…" button that keeps the form hidden until pressed. Viewing the data
 * table is public, but opening the form requires a signed-in citizen — a
 * signed-out press opens the sign-in dialog instead.
 */
export function FormDialogButton({
  label,
  dialogTitle,
  icon: Icon = Plus,
  color = "primary",
  children,
}: FormDialogButtonProps) {
  const { session, requestSignIn } = useCitizenSession();
  const [open, setOpen] = useState(false);

  const handleClick = () => {
    if (!session) {
      requestSignIn();
      return;
    }
    setOpen(true);
  };

  const close = () => setOpen(false);

  return (
    <>
      <Button
        color={color}
        onClick={handleClick}
        startIcon={<Icon size={18} />}
        sx={{ fontWeight: 800 }}
        variant="contained"
      >
        {label}
      </Button>
      <Dialog fullWidth maxWidth="md" onClose={close} open={open} scroll="paper">
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
            fontWeight: 800,
          }}
        >
          {dialogTitle}
          <IconButton aria-label="Close" onClick={close} size="small">
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>{open ? children(close) : null}</DialogContent>
      </Dialog>
    </>
  );
}
