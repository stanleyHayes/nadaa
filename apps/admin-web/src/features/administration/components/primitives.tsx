import { useEffect, useRef, useState, type ReactNode } from "react";
import { Box, Paper, Stack, Typography } from "@mui/material";
import type { LucideIcon } from "lucide-react";

export type AccentTone = "navy" | "gold" | "green" | "red" | "info";

function prefersReducedMotion() {
  return (
    typeof window !== "undefined" &&
    window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
  );
}

/**
 * Count a stat value up from zero on mount. Honours reduced-motion: when the
 * user asks for less motion, the final value is shown immediately.
 */
function useCountUp(target: number, duration = 650): number {
  const [value, setValue] = useState(() =>
    prefersReducedMotion() ? target : 0,
  );
  const frame = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (prefersReducedMotion()) {
      setValue(target);
      return;
    }
    const start = performance.now();
    const from = 0;
    const step = (now: number) => {
      const progress = Math.min(1, (now - start) / duration);
      // easeOutCubic for a settled finish
      const eased = 1 - Math.pow(1 - progress, 3);
      setValue(Math.round(from + (target - from) * eased));
      if (progress < 1) {
        frame.current = requestAnimationFrame(step);
      }
    };
    frame.current = requestAnimationFrame(step);
    return () => {
      if (frame.current) {
        cancelAnimationFrame(frame.current);
      }
    };
  }, [target, duration]);

  return value;
}

export function Eyebrow({
  children,
  tone = "muted",
}: {
  children: ReactNode;
  tone?: "muted" | "gold" | "green" | "inverse";
}) {
  return <span className={`cc-eyebrow cc-eyebrow--${tone}`}>{children}</span>;
}

function CountValue({ target, suffix }: { target: number; suffix?: string }) {
  const value = useCountUp(target);
  return (
    <>
      {value.toLocaleString()}
      {suffix ?? ""}
    </>
  );
}

export function MetricTile({
  label,
  value,
  suffix,
  caption,
  icon: Icon,
  accent = "navy",
}: {
  label: string;
  value: ReactNode;
  suffix?: string;
  caption?: string;
  icon: LucideIcon;
  accent?: AccentTone;
}) {
  return (
    <Paper className={`cc-metric cc-metric--${accent} cc-rise`} elevation={0}>
      <span className="cc-metric__watermark" aria-hidden>
        <Icon size={120} strokeWidth={1.5} />
      </span>
      <Box className="cc-metric__body">
        <Eyebrow>{label}</Eyebrow>
        <Typography className="cc-metric__value">
          {typeof value === "number" ? (
            <CountValue target={value} suffix={suffix} />
          ) : (
            value
          )}
        </Typography>
        {caption ? (
          <Typography className="cc-metric__caption">{caption}</Typography>
        ) : null}
      </Box>
      <span className="cc-metric__icon" aria-hidden>
        <Icon size={20} />
      </span>
    </Paper>
  );
}

export function SectionCard({
  title,
  eyebrow,
  icon: Icon,
  action,
  children,
  className,
  accent = "navy",
}: {
  title: string;
  eyebrow?: string;
  icon?: LucideIcon;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
  accent?: AccentTone;
}) {
  return (
    <Paper
      className={`cc-section cc-section--${accent} cc-rise${className ? ` ${className}` : ""}`}
      elevation={0}
    >
      <Stack
        direction="row"
        justifyContent="space-between"
        alignItems="flex-start"
        gap={1.5}
        className="cc-section__head"
      >
        <Stack direction="row" spacing={1.25} alignItems="center">
          {Icon ? (
            <span className="cc-section__icon" aria-hidden>
              <Icon size={18} />
            </span>
          ) : null}
          <Box>
            {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
            <Typography variant="h6" className="cc-section__title">
              {title}
            </Typography>
          </Box>
        </Stack>
        {action ? <Box className="cc-section__action">{action}</Box> : null}
      </Stack>
      <Box className="cc-section__body">{children}</Box>
    </Paper>
  );
}

/**
 * Horizontal coverage bar. Used for MFA-coverage rows where a higher share is
 * healthier: full coverage reads green, partial gold, and thin coverage red.
 */
export function CoverageMeter({
  value,
  tone,
}: {
  value: number;
  tone?: "green" | "gold" | "red";
}) {
  const pct = Math.max(0, Math.min(100, Math.round(value)));
  const resolvedTone = tone ?? (pct >= 90 ? "green" : pct >= 70 ? "gold" : "red");
  return (
    <div
      className={`cc-capacity cc-capacity--${resolvedTone}`}
      role="progressbar"
      aria-valuenow={pct}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <div className="cc-capacity__fill" style={{ width: `${pct}%` }} />
    </div>
  );
}

export function PostureRow({
  label,
  value,
  tone = "green",
}: {
  label: string;
  value: string;
  tone?: "green" | "gold" | "red" | "navy";
}) {
  return (
    <div className="cc-posture">
      <span className="cc-posture__label">{label}</span>
      <span className={`cc-posture__value cc-posture__value--${tone}`}>
        <span className="cc-posture__dot" aria-hidden />
        {value}
      </span>
    </div>
  );
}

export function ViewIntro({
  title,
  description,
  icon: Icon,
  action,
}: {
  title: string;
  description: string;
  icon?: LucideIcon;
  action?: ReactNode;
}) {
  return (
    <Stack
      direction={{ xs: "column", md: "row" }}
      justifyContent="space-between"
      alignItems={{ xs: "flex-start", md: "center" }}
      gap={2}
      className="cc-view-intro"
    >
      <Stack direction="row" spacing={1.5} alignItems="center">
        {Icon ? (
          <span className="cc-view-intro__chip" aria-hidden>
            <Icon size={20} />
          </span>
        ) : null}
        <Box>
          <Typography variant="h4" className="cc-view-intro__title">
            {title}
          </Typography>
          <Typography color="text.secondary" className="cc-view-intro__desc">
            {description}
          </Typography>
        </Box>
      </Stack>
      {action ? <Box className="cc-view-intro__action">{action}</Box> : null}
    </Stack>
  );
}

/**
 * Shimmer skeleton rows shown while a data surface is loading.
 */
export function SkeletonRows({ rows = 4 }: { rows?: number }) {
  return (
    <Stack spacing={1.25} aria-hidden>
      {Array.from({ length: rows }).map((_, index) => (
        <div className="cc-skeleton" key={index} />
      ))}
    </Stack>
  );
}
