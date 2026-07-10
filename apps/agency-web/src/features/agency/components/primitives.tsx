import { useEffect, useState, type ReactNode } from "react";
import { Box, Chip, Paper, Stack, Typography } from "@mui/material";
import type { LucideIcon } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { IncidentRecord, IncidentStatus } from "@nadaa/shared-types";
import { statusLabel } from "../data";

export type AccentTone = "navy" | "gold" | "green" | "red" | "info";

function prefersReducedMotion(): boolean {
  return (
    typeof window !== "undefined" &&
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

/**
 * Animates a number from 0 up to `target` once on mount. Respects the user's
 * reduced-motion preference by jumping straight to the final value.
 */
function useCountUp(target: number): number {
  const [value, setValue] = useState<number>(() =>
    prefersReducedMotion() ? target : 0,
  );

  useEffect(() => {
    if (prefersReducedMotion()) {
      setValue(target);
      return;
    }
    let raf = 0;
    const duration = 620;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      const eased = 1 - Math.pow(1 - t, 3);
      setValue(Math.round(target * eased));
      if (t < 1) {
        raf = requestAnimationFrame(tick);
      }
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target]);

  return value;
}

function MetricValue({ value }: { value: ReactNode }) {
  const numeric = typeof value === "number";
  const counted = useCountUp(numeric ? value : 0);
  return (
    <Typography className="cc-metric__value">
      {numeric ? counted : value}
    </Typography>
  );
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

export function MetricTile({
  label,
  value,
  caption,
  icon: Icon,
  accent = "navy",
}: {
  label: string;
  value: ReactNode;
  caption?: string;
  icon: LucideIcon;
  accent?: AccentTone;
}) {
  return (
    <Paper className={`cc-metric cc-metric--${accent} cc-reveal`} elevation={0}>
      <span className="cc-metric__watermark" aria-hidden>
        <Icon size={120} strokeWidth={1.5} />
      </span>
      <Box className="cc-metric__body">
        <Eyebrow>{label}</Eyebrow>
        <MetricValue value={value} />
        {caption ? (
          <Typography className="cc-metric__caption">{caption}</Typography>
        ) : null}
      </Box>
      <span className="cc-chip cc-metric__icon" aria-hidden>
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
      className={`cc-section cc-section--${accent} cc-reveal${className ? ` ${className}` : ""}`}
      elevation={0}
    >
      <Stack
        direction="row"
        className="cc-section__head"
        sx={{
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: 1.5
        }}>
        <Stack direction="row" spacing={1.25} sx={{
          alignItems: "center"
        }}>
          {Icon ? (
            <span className="cc-chip cc-section__icon" aria-hidden>
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

export function CapacityMeter({
  value,
  max,
  tone = "green",
}: {
  value: number;
  max: number;
  tone?: "green" | "gold" | "red";
}) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;
  return (
    <div
      className={`cc-capacity cc-capacity--${tone}`}
      role="progressbar"
      aria-valuenow={pct}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <div className="cc-capacity__fill" style={{ width: `${pct}%` }} />
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
      className="cc-view-intro"
      sx={{
        justifyContent: "space-between",
        alignItems: { xs: "flex-start", md: "center" },
        gap: 2
      }}>
      <Stack
        direction="row"
        spacing={1.5}
        sx={{
          alignItems: "center",
          minWidth: 0
        }}>
        {Icon ? (
          <span className="cc-chip cc-view-intro__chip" aria-hidden>
            <Icon size={20} />
          </span>
        ) : null}
        <Box sx={{
          minWidth: 0
        }}>
          <Typography variant="h4" className="cc-view-intro__title">
            {title}
          </Typography>
          <Typography className="cc-view-intro__desc" sx={{
            color: "text.secondary"
          }}>
            {description}
          </Typography>
        </Box>
      </Stack>
      {action ? <Box className="cc-view-intro__action">{action}</Box> : null}
    </Stack>
  );
}

export function StatusLine({
  color,
  label,
  value,
}: {
  color: "success" | "warning" | "default";
  label: string;
  value: string;
}) {
  return (
    <Stack
      direction="row"
      sx={{
        justifyContent: "space-between",
        alignItems: "center",
        gap: 1
      }}>
      <Typography variant="body2">{label}</Typography>
      <Chip size="small" label={value} color={color} />
    </Stack>
  );
}

/**
 * Signature element: a live response-stage board. Each rung bands the active
 * incidents on the desk from Assigned through Recovery, with a proportional
 * fill, so a duty officer reads the operational load at a glance.
 */
const RESPONSE_STAGES: Array<{ status: IncidentStatus; color: string }> = [
  { status: "assigned", color: "var(--nadaa-navy)" },
  { status: "response_en_route", color: nadaaBrand.colors.gold },
  { status: "on_scene", color: nadaaBrand.colors.red },
  { status: "contained", color: nadaaBrand.colors.green },
  { status: "recovery_ongoing", color: "var(--nadaa-slate)" },
];

export function ResponseLadder({
  incidents,
}: {
  incidents: IncidentRecord[];
}) {
  const active = incidents.filter(
    (incident) =>
      incident.status !== "closed" && incident.status !== "false_report",
  );
  const total = active.length;

  return (
    <div className="cc-triage" role="list" aria-label="Response stage board">
      {RESPONSE_STAGES.map(({ status, color }) => {
        const items = active.filter((incident) => incident.status === status);
        const count = items.length;
        const share = total ? Math.round((count / total) * 100) : 0;
        const fill = count === 0 ? 0 : Math.max(6, share);
        const priority = items.filter(
          (incident) => incident.priorityReview,
        ).length;
        return (
          <div
            className={`cc-triage__rung${count === 0 ? " cc-triage__rung--empty" : ""}`}
            role="listitem"
            key={status}
          >
            <div className="cc-triage__label">
              <span
                className="cc-triage__dot"
                style={{ background: color }}
                aria-hidden
              />
              <span className="cc-triage__name">{statusLabel(status)}</span>
            </div>
            <div className="cc-triage__meter">
              <div className="cc-triage__track">
                <div
                  className="cc-triage__fill"
                  style={{ width: `${fill}%`, background: color }}
                />
              </div>
              <span className="cc-triage__hazards">
                {priority > 0
                  ? `${priority} flagged for priority review`
                  : "No priority flags"}
              </span>
            </div>
            <div className="cc-triage__figures">
              <span className="cc-triage__count" style={{ color }}>
                {count}
              </span>
              <span className="cc-triage__share">{share}%</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
