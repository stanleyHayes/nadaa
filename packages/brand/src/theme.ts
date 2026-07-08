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

export interface NadaaThemeOptions {
  /** When true, disable non-essential motion. */
  reducedMotion?: boolean;
  /** Accent color used for topbar/brand accents. */
  accent?: "public" | "operational";
}

/**
 * Creates the canonical NADAA MUI theme.
 *
 * All MUI-based web apps should consume this factory and avoid local theme
 * duplication. App-specific visual overrides should be rare and documented.
 */
export function createNadaaTheme(options: NadaaThemeOptions = {}) {
  const { reducedMotion = false, accent = "operational" } = options;
  const accentColor =
    accent === "public" ? appAccent.public : appAccent.operational;

  const base: ThemeOptions = {
    breakpoints: {
      values: breakpoints.values,
    },
    palette: {
      mode: "light",
      primary: {
        main: colors.navy,
        contrastText: colors.white,
      },
      secondary: {
        main: colors.green,
        contrastText: colors.white,
      },
      error: {
        main: colors.red,
        contrastText: colors.white,
      },
      warning: {
        main: colors.gold,
        contrastText: colors.ink,
      },
      info: {
        main: semantic.info,
        contrastText: colors.white,
      },
      success: {
        main: colors.green,
        contrastText: colors.white,
      },
      background: {
        default: semantic.surface,
        paper: semantic.surfaceElevated,
      },
      text: {
        primary: semantic.textPrimary,
        secondary: semantic.textSecondary,
      },
      divider: semantic.divider,
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
            backgroundColor: semantic.surface,
            color: semantic.textPrimary,
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
