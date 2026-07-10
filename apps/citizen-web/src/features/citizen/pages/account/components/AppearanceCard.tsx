import { useId, useState, type ReactNode } from "react";
import { Box, Paper, Stack, Typography } from "@mui/material";
import { Check, Moon, Sun, type LucideIcon } from "lucide-react";
import { DARK_TINTS, type DarkTint } from "@nadaa/brand";
import { getInitialTint, setDarkTint, useThemeMode } from "@/app/theme-mode";

/**
 * Surface card with a tinted icon plaque, title, muted description, and body.
 * Self-contained so the account area does not depend on the command-center
 * primitives; the `--nadaa-*` tokens flip it for dark automatically.
 */
function SettingCard({
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
        backgroundColor: "var(--nadaa-surface-elevated, #ffffff)",
        boxShadow: "var(--nadaa-shadow-md)",
      }}
    >
      <Stack direction="row" spacing={1.5} alignItems="flex-start">
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

/** One selectable tint: label, description, and a 3-swatch preview strip. */
function TintOption({
  option,
  selected,
  onSelect,
}: {
  option: (typeof DARK_TINTS)[number];
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <Box
      component="label"
      sx={{
        position: "relative",
        display: "block",
        cursor: "pointer",
        p: 1.5,
        borderRadius: "12px",
        border: "1px solid",
        borderColor: selected
          ? "var(--nadaa-navy, #0d1b3d)"
          : "var(--nadaa-border, #dfeaf2)",
        backgroundColor: "var(--nadaa-mist, #f5f8fc)",
        transition: "border-color 150ms ease, box-shadow 150ms ease",
        boxShadow: selected
          ? "0 0 0 2px color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 22%, transparent)"
          : "none",
        "&:hover": { borderColor: "var(--nadaa-navy, #0d1b3d)" },
        "&:focus-within": {
          borderColor: "var(--nadaa-gold, #f4c20d)",
          boxShadow:
            "0 0 0 3px color-mix(in srgb, var(--nadaa-gold, #f4c20d) 32%, transparent)",
        },
        "@media (prefers-reduced-motion: reduce)": { transition: "none" },
      }}
    >
      <Box
        component="input"
        type="radio"
        name="nadaa-dark-tint"
        value={option.value}
        checked={selected}
        onChange={onSelect}
        sx={{
          position: "absolute",
          width: 1,
          height: 1,
          padding: 0,
          margin: "-1px",
          overflow: "hidden",
          clip: "rect(0 0 0 0)",
          whiteSpace: "nowrap",
          border: 0,
        }}
      />
      <Stack
        direction="row"
        alignItems="flex-start"
        justifyContent="space-between"
        gap={1}
      >
        <Box sx={{ minWidth: 0 }}>
          <Typography
            sx={{
              fontSize: "0.9rem",
              fontWeight: 700,
              color: "var(--nadaa-ink, #101828)",
            }}
          >
            {option.label}
          </Typography>
          <Typography
            sx={{
              mt: 0.25,
              fontSize: "0.78rem",
              lineHeight: 1.4,
              color: "var(--nadaa-text-secondary, #555b66)",
            }}
          >
            {option.description}
          </Typography>
        </Box>
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 22,
            height: 22,
            borderRadius: "999px",
            border: "1px solid",
            borderColor: selected
              ? "var(--nadaa-navy, #0d1b3d)"
              : "var(--nadaa-border, #dfeaf2)",
            backgroundColor: selected
              ? "var(--nadaa-navy, #0d1b3d)"
              : "transparent",
            /* --nadaa-text-inverse is white in both modes, so the check reads on
               the navy/indigo fill regardless of light or dark. */
            color: selected ? "var(--nadaa-text-inverse, #fff)" : "transparent",
          }}
        >
          <Check size={13} />
        </Box>
      </Stack>
      <Box
        aria-hidden
        sx={{
          mt: 1.25,
          display: "flex",
          height: 34,
          borderRadius: "8px",
          overflow: "hidden",
          border: "1px solid var(--nadaa-border, #dfeaf2)",
        }}
      >
        {option.swatches.map((swatch) => (
          <Box key={swatch} sx={{ flex: 1, backgroundColor: swatch }} />
        ))}
      </Box>
    </Box>
  );
}

/**
 * Appearance settings — a light/dark note plus the dark screen-tint picker.
 * Mode is flipped from the header Sun/Moon button; the tints here recolour the
 * dark screen cast only, mirroring the command-center AppearanceCard.
 */
export function AppearanceCard() {
  const mode = useThemeMode();
  const [tint, setTint] = useState<DarkTint>(getInitialTint);
  const groupLabelId = useId();

  const choose = (value: DarkTint) => {
    setTint(value);
    setDarkTint(value);
  };

  return (
    <SettingCard
      icon={mode === "dark" ? Moon : Sun}
      title="Appearance"
      description="Choose the app's light or dark screen and its dark tint."
    >
      <Stack spacing={2}>
        <Box
          sx={{
            display: "flex",
            gap: 1.25,
            alignItems: "flex-start",
            p: 1.5,
            borderRadius: "12px",
            border: "1px solid var(--nadaa-border, #dfeaf2)",
            backgroundColor: "var(--nadaa-mist, #f5f8fc)",
          }}
        >
          <Box
            aria-hidden
            sx={{
              flex: "0 0 auto",
              display: "flex",
              gap: 0.5,
              color: "var(--nadaa-navy, #0d1b3d)",
              mt: 0.25,
            }}
          >
            <Sun size={16} />
            <Moon size={16} />
          </Box>
          <Typography
            sx={{
              fontSize: "0.82rem",
              lineHeight: 1.5,
              color: "var(--nadaa-text-secondary, #555b66)",
            }}
          >
            Currently in <strong>{mode === "dark" ? "dark" : "light"}</strong>{" "}
            mode. Use the theme button in the header to switch. These tints
            change the dark screen colour only.
          </Typography>
        </Box>

        <Box>
          <Typography
            id={groupLabelId}
            sx={{
              fontSize: "0.68rem",
              fontWeight: 700,
              letterSpacing: "0.08em",
              textTransform: "uppercase",
              color: "var(--nadaa-text-secondary, #555b66)",
              mb: 1,
            }}
          >
            Dark screen tint
          </Typography>
          <Box
            role="radiogroup"
            aria-labelledby={groupLabelId}
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr" },
            }}
          >
            {DARK_TINTS.map((option) => (
              <TintOption
                key={option.value}
                option={option}
                selected={option.value === tint}
                onSelect={() => choose(option.value)}
              />
            ))}
          </Box>
        </Box>
      </Stack>
    </SettingCard>
  );
}

export default AppearanceCard;
