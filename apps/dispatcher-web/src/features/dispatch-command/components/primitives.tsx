import type { ReactNode } from "react";
import { Box, Paper, Stack, Typography } from "@mui/material";
import type { LucideIcon } from "lucide-react";
import type { HazardType, RiskLevel } from "@nadaa/shared-types";
import { severityColors } from "../data";
import type { CommandIncident } from "../types";
import { hazardLabel, severityLabel } from "../utils";

export type AccentTone = "navy" | "gold" | "green" | "red" | "info";

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
    <Paper className={`cc-metric cc-metric--${accent}`} elevation={0}>
      <Box className="cc-metric__body">
        <Eyebrow>{label}</Eyebrow>
        <Typography className="cc-metric__value">{value}</Typography>
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
      className={`cc-section cc-section--${accent}${className ? ` ${className}` : ""}`}
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

const TRIAGE_ORDER: RiskLevel[] = [
  "emergency",
  "severe",
  "high",
  "moderate",
  "low",
];

function topHazards(incidents: CommandIncident[]): string {
  const counts = new Map<HazardType, number>();
  for (const incident of incidents) {
    counts.set(incident.type, (counts.get(incident.type) ?? 0) + 1);
  }
  const ranked = [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 2);
  if (!ranked.length) {
    return "No active reports";
  }
  return ranked
    .map(([hazard, count]) => `${hazardLabel(hazard)} ${count}`)
    .join(" · ");
}

/**
 * Signature element: a live severity-triage board. Each rung bands active
 * incidents from Emergency down to Low, with a proportional fill and hazard
 * breakdown, so a dispatcher reads the operational load at a glance.
 */
export function TriageLadder({ incidents }: { incidents: CommandIncident[] }) {
  const active = incidents.filter(
    (incident) =>
      incident.status !== "closed" && incident.status !== "false_report",
  );
  const total = active.length;

  return (
    <div className="cc-triage" role="list" aria-label="Severity triage board">
      {TRIAGE_ORDER.map((severity) => {
        const items = active.filter(
          (incident) => incident.severity === severity,
        );
        const count = items.length;
        const share = total ? Math.round((count / total) * 100) : 0;
        const fill = count === 0 ? 0 : Math.max(6, share);
        const color = severityColors[severity];
        return (
          <div
            className={`cc-triage__rung${count === 0 ? " cc-triage__rung--empty" : ""}`}
            role="listitem"
            key={severity}
          >
            <div className="cc-triage__label">
              <span
                className="cc-triage__dot"
                style={{ background: color }}
                aria-hidden
              />
              <span className="cc-triage__name">{severityLabel(severity)}</span>
            </div>
            <div className="cc-triage__meter">
              <div className="cc-triage__track">
                <div
                  className="cc-triage__fill"
                  style={{ width: `${fill}%`, background: color }}
                />
              </div>
              <span className="cc-triage__hazards">{topHazards(items)}</span>
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
  action,
}: {
  title: string;
  description: string;
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
      <Box>
        <Typography variant="h4" className="cc-view-intro__title">
          {title}
        </Typography>
        <Typography color="text.secondary" className="cc-view-intro__desc">
          {description}
        </Typography>
      </Box>
      {action ? <Box className="cc-view-intro__action">{action}</Box> : null}
    </Stack>
  );
}
