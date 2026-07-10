import { createTheme, type ThemeOptions } from "@mui/material/styles";
import {
  colors,
  semantic,
  spacing,
  typography,
  shadows,
  breakpoints,
  radii,
  appAccent,
} from "./tokens.js";

export type ThemeMode = "light" | "dark";

export interface NadaaThemeOptions {
  /** When true, disable non-essential motion. */
  reducedMotion?: boolean;
  /** Accent color used for topbar/brand accents. */
  accent?: "public" | "operational";
  /** Light (default) or dark palette. Must mirror the `data-theme` attribute
   *  applied by the theme-preferences helper so MUI surfaces match the CSS
   *  token flip. */
  mode?: ThemeMode;
}

/**
 * Dark-mode neutrals. These mirror the `--nadaa-*` dark token values in
 * `dark.css` (Ink default) so MUI-rendered surfaces — Dialog, Menu, TextField,
 * Table, Card — land on the same warm-navy palette as the `cc-*` surfaces.
 */
const darkNeutrals = {
  /** Screen base. */
  background: "#0b1120",
  /** Elevated card / paper (Menu, Dialog, Select popover). */
  paper: "#1a2440",
  textPrimary: "#eaf0fb",
  textSecondary: "#9aa8c0",
  divider: "#29324e",
  /** Navy lightened to an indigo so it reads as an accent on dark. */
  primary: "#8ea6dc",
  green: "#2eba71",
  red: "#ff5f5a",
  gold: "#f6ca3d",
  info: "#3a9fe0",
} as const;

/**
 * Creates the canonical NADAA MUI theme.
 *
 * All MUI-based web apps should consume this factory and avoid local theme
 * duplication. App-specific visual overrides should be rare and documented.
 */
export function createNadaaTheme(options: NadaaThemeOptions = {}) {
  const { reducedMotion = false, accent = "operational", mode = "light" } =
    options;
  const isDark = mode === "dark";
  const accentColor =
    accent === "public" ? appAccent.public : appAccent.operational;

  // Neutral surfaces/text resolve per mode; brand accents stay vivid (brightened
  // in dark for contrast). Kept in sync with `dark.css`.
  const bg = isDark
    ? { default: darkNeutrals.background, paper: darkNeutrals.paper }
    : { default: semantic.surface, paper: semantic.surfaceElevated };
  const text = isDark
    ? { primary: darkNeutrals.textPrimary, secondary: darkNeutrals.textSecondary }
    : { primary: semantic.textPrimary, secondary: semantic.textSecondary };
  const dividerColor = isDark ? darkNeutrals.divider : semantic.divider;

  const base: ThemeOptions = {
    breakpoints: {
      values: breakpoints.values,
    },
    palette: {
      mode,
      primary: {
        main: isDark ? darkNeutrals.primary : colors.navy,
        // On dark the indigo primary carries dark text; on light, navy carries white.
        contrastText: isDark ? "#0b1120" : colors.white,
      },
      secondary: {
        main: isDark ? darkNeutrals.green : colors.green,
        contrastText: isDark ? "#0b1120" : colors.white,
      },
      error: {
        main: isDark ? darkNeutrals.red : colors.red,
        contrastText: isDark ? "#0b1120" : colors.white,
      },
      warning: {
        main: isDark ? darkNeutrals.gold : colors.gold,
        contrastText: colors.ink,
      },
      info: {
        main: isDark ? darkNeutrals.info : semantic.info,
        contrastText: isDark ? "#0b1120" : colors.white,
      },
      success: {
        main: isDark ? darkNeutrals.green : colors.green,
        contrastText: isDark ? "#0b1120" : colors.white,
      },
      background: {
        default: bg.default,
        paper: bg.paper,
      },
      text: {
        primary: text.primary,
        secondary: text.secondary,
      },
      divider: dividerColor,
    },
    typography: {
      fontFamily: typography.fontFamily,
      h1: {
        fontSize: typography.sizes["4xl"].fontSize,
        lineHeight: typography.sizes["4xl"].lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      h2: {
        fontSize: typography.sizes["3xl"].fontSize,
        lineHeight: typography.sizes["3xl"].lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      h3: {
        fontSize: typography.sizes["2xl"].fontSize,
        lineHeight: typography.sizes["2xl"].lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      h4: {
        fontSize: typography.sizes.xl.fontSize,
        lineHeight: typography.sizes.xl.lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      h5: {
        fontSize: typography.sizes.lg.fontSize,
        lineHeight: typography.sizes.lg.lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      h6: {
        fontSize: typography.sizes.base.fontSize,
        lineHeight: typography.sizes.base.lineHeight,
        fontWeight: typography.weights.extrabold,
      },
      body1: {
        fontSize: typography.sizes.base.fontSize,
        lineHeight: typography.sizes.base.lineHeight,
      },
      body2: {
        fontSize: typography.sizes.sm.fontSize,
        lineHeight: typography.sizes.sm.lineHeight,
      },
      button: {
        fontWeight: typography.weights.semibold,
        textTransform: "none",
      },
      caption: {
        fontSize: typography.sizes.xs.fontSize,
        lineHeight: typography.sizes.xs.lineHeight,
      },
    },
    shape: {
      borderRadius: parseInt(radii.md, 10),
    },
    shadows: [
      shadows.none,
      shadows.sm,
      shadows.sm,
      shadows.md,
      shadows.md,
      shadows.md,
      shadows.lg,
      shadows.lg,
      shadows.lg,
      shadows.lg,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
      shadows.xl,
    ],
    spacing: (factor: number) => `${factor * 8}px`,
    transitions: {
      duration: {
        shortest: reducedMotion ? 0 : 150,
        shorter: reducedMotion ? 0 : 200,
        short: reducedMotion ? 0 : 250,
        standard: reducedMotion ? 0 : 300,
        complex: reducedMotion ? 0 : 375,
        enteringScreen: reducedMotion ? 0 : 225,
        leavingScreen: reducedMotion ? 0 : 195,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          html: {
            scrollBehavior: reducedMotion ? "auto" : "smooth",
          },
          body: {
            backgroundColor: bg.default,
            color: text.primary,
          },
          ":focus-visible": {
            outline: `2px solid ${accentColor}`,
            outlineOffset: "2px",
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          root: {
            minHeight: "42px",
            borderRadius: radii.md,
          },
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            borderRadius: radii.md,
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            borderRadius: radii.md,
            boxShadow: shadows.md,
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundColor: colors.navy,
            color: colors.white,
            boxShadow: shadows.sm,
          },
        },
      },
      MuiLink: {
        defaultProps: {
          underline: "hover",
        },
      },
    },
  };

  return createTheme(base);
}
