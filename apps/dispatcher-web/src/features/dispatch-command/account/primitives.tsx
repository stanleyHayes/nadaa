import { useId, type ReactNode } from "react";
import { Box, Chip, Paper, Stack, Switch, Typography } from "@mui/material";
import type { LucideIcon } from "lucide-react";

/** Tones borrow the command-center palette: navy accent, gold/green/red status. */
export type AccountTone = "navy" | "gold" | "green" | "red";

const TONE_STYLES: Record<
  AccountTone,
  { color: string; background: string; border: string }
> = {
  navy: {
    color: "var(--nadaa-navy, #0d1b3d)",
    background: "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 8%, transparent)",
    border: "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 22%, transparent)",
  },
  gold: {
    color: "var(--cc-gold-ink, #a97e00)",
    background: "#fef9e7",
    border: "#f7e28d",
  },
  green: {
    color: "var(--nadaa-green, #118d4e)",
    background: "#e8f6ee",
    border: "#b8e4cc",
  },
  red: {
    color: "var(--nadaa-red, #e53935)",
    background: "#fdecec",
    border: "#f6c6c6",
  },
};

/** Two-letter avatar/initials, shared with the top-bar user control. */
export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).slice(0, 2);
  return parts.map((part) => part[0]?.toUpperCase() ?? "").join("") || "ND";
}

/** Long date + time in the Ghana locale, e.g. "9 Jul 2026, 07:42". */
export function formatDateTime(value?: string | null): string {
  if (!value) {
    return "Not recorded";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Not recorded";
  }
  return new Intl.DateTimeFormat("en-GH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

/** Small tinted status pill (Active, Enabled, Not enabled, …). */
export function StatusChip({
  label,
  tone,
}: {
  label: string;
  tone: AccountTone;
}) {
  const styles = TONE_STYLES[tone];
  return (
    <Chip
      label={label}
      size="small"
      sx={{
        height: 24,
        fontWeight: 700,
        fontSize: "0.74rem",
        letterSpacing: "0.01em",
        color: styles.color,
        backgroundColor: styles.background,
        border: `1px solid ${styles.border}`,
        "& .MuiChip-label": { px: 1.1 },
      }}
    />
  );
}

/**
 * Surface card with a tinted icon chip, title, muted description, and body —
 * the building block for every settings section.
 */
export function SettingCard({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
  children: ReactNode;
}) {
  return (
    <Paper
      elevation={0}
      sx={{
        p: { xs: 2.5, md: 3 },
        border: "1px solid var(--nadaa-border, #dfeaf2)",
        borderRadius: "14px",
        backgroundColor: "var(--nadaa-white, #ffffff)",
        boxShadow: "var(--nadaa-shadow-md)",
      }}
    >
      <Stack direction="row" spacing={1.5} sx={{
        alignItems: "flex-start"
      }}>
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 40,
            height: 40,
            borderRadius: "10px",
            color: "var(--nadaa-navy, #0d1b3d)",
            backgroundColor:
              "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 8%, transparent)",
          }}
        >
          <Icon size={20} />
        </Box>
        <Box sx={{ minWidth: 0 }}>
          <Typography
            component="h3"
            sx={{
              fontSize: "1.02rem",
              fontWeight: 800,
              lineHeight: 1.2,
              color: "var(--nadaa-ink, #101828)",
            }}
          >
            {title}
          </Typography>
          <Typography
            sx={{
              mt: 0.25,
              fontSize: "0.85rem",
              color: "var(--nadaa-text-secondary, #555b66)",
            }}
          >
            {description}
          </Typography>
        </Box>
      </Stack>
      <Box sx={{ mt: 2.5 }}>{children}</Box>
    </Paper>
  );
}

/** Bordered read-only row: icon chip + label + value, for the profile summary. */
export function InfoRow({
  icon: Icon,
  label,
  children,
}: {
  icon: LucideIcon;
  label: string;
  children: ReactNode;
}) {
  return (
    <Box
      sx={{
        display: "flex",
        gap: 1.5,
        alignItems: "flex-start",
        p: 1.5,
        border: "1px solid var(--nadaa-border, #dfeaf2)",
        borderRadius: "12px",
        backgroundColor: "var(--nadaa-mist, #f5f8fc)",
      }}
    >
      <Box
        aria-hidden
        sx={{
          flex: "0 0 auto",
          display: "grid",
          placeItems: "center",
          width: 36,
          height: 36,
          borderRadius: "9px",
          color: "var(--nadaa-navy, #0d1b3d)",
          backgroundColor: "var(--nadaa-white, #ffffff)",
          border: "1px solid var(--nadaa-border, #dfeaf2)",
        }}
      >
        <Icon size={17} />
      </Box>
      <Box sx={{ minWidth: 0, flex: 1 }}>
        <Typography
          sx={{
            fontSize: "0.68rem",
            fontWeight: 700,
            letterSpacing: "0.08em",
            textTransform: "uppercase",
            color: "var(--nadaa-text-secondary, #555b66)",
          }}
        >
          {label}
        </Typography>
        <Box
          sx={{
            mt: 0.5,
            fontSize: "0.9rem",
            fontWeight: 600,
            color: "var(--nadaa-ink, #101828)",
          }}
        >
          {children}
        </Box>
      </Box>
    </Box>
  );
}

/** Bordered toggle row with a switch, used across notifications and preferences. */
export function PreferenceRow({
  icon: Icon,
  label,
  description,
  checked,
  disabled = false,
  onChange,
}: {
  icon?: LucideIcon;
  label: string;
  description: string;
  checked: boolean;
  /** Renders the switch inert — for channels that are not connected yet. */
  disabled?: boolean;
  onChange: (checked: boolean) => void;
}) {
  const labelId = useId();
  const descriptionId = useId();
  return (
    <Box
      sx={{
        display: "flex",
        gap: 1.5,
        alignItems: "flex-start",
        p: 1.75,
        border: "1px solid var(--nadaa-border, #dfeaf2)",
        borderRadius: "12px",
        backgroundColor: "var(--nadaa-mist, #f5f8fc)",
        opacity: disabled ? 0.62 : 1,
        transition: "border-color 150ms ease",
        "&:focus-within": {
          borderColor: "var(--nadaa-gold, #f4c20d)",
          boxShadow:
            "0 0 0 3px color-mix(in srgb, var(--nadaa-gold, #f4c20d) 32%, transparent)",
        },
        "@media (prefers-reduced-motion: reduce)": { transition: "none" },
      }}
    >
      {Icon ? (
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 34,
            height: 34,
            mt: 0.25,
            borderRadius: "9px",
            color: "var(--nadaa-navy, #0d1b3d)",
            backgroundColor:
              "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 8%, transparent)",
          }}
        >
          <Icon size={17} />
        </Box>
      ) : null}
      <Box sx={{ minWidth: 0, flex: 1 }}>
        <Typography
          component="label"
          htmlFor={labelId}
          sx={{
            display: "block",
            fontSize: "0.9rem",
            fontWeight: 700,
            color: "var(--nadaa-ink, #101828)",
            cursor: "pointer",
          }}
        >
          {label}
        </Typography>
        <Typography
          id={descriptionId}
          sx={{
            mt: 0.25,
            fontSize: "0.82rem",
            lineHeight: 1.45,
            color: "var(--nadaa-text-secondary, #555b66)",
          }}
        >
          {description}
        </Typography>
      </Box>
      <Switch
        id={labelId}
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
        color="primary"
        disabled={disabled}
        slotProps={{ input: { "aria-describedby": descriptionId } }}
        sx={{ flex: "0 0 auto", mt: -0.25 }}
      />
    </Box>
  );
}
