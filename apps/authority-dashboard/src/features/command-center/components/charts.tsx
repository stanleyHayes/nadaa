import { Box, Stack, Typography } from "@mui/material";

/**
 * Dependency-free SVG chart primitives for the overview dashboards. Kept small
 * and brand-styled (navy/gold/green/red + --nadaa-mist track) so they match the
 * existing custom SVG (CapacityMeter) rather than pulling in a chart library.
 */

export type ChartDatum = { label: string; value: number; color: string };

/** Ring chart of category proportions, with a center figure and a legend. */
export function DonutChart({
  data,
  size = 148,
  thickness = 18,
  centerValue,
  centerLabel,
}: {
  data: ChartDatum[];
  size?: number;
  thickness?: number;
  centerValue?: string | number;
  centerLabel?: string;
}) {
  const total = data.reduce((sum, datum) => sum + datum.value, 0);
  const radius = (size - thickness) / 2;
  const circumference = 2 * Math.PI * radius;
  let offset = 0;

  return (
    <Stack
      direction="row"
      spacing={2.5}
      sx={{ alignItems: "center", flexWrap: "wrap" }}
    >
      <Box sx={{ position: "relative", width: size, height: size, flexShrink: 0 }}>
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          role="img"
          aria-label={centerLabel ?? "Breakdown"}
        >
          <g transform={`rotate(-90 ${size / 2} ${size / 2})`}>
            <circle
              cx={size / 2}
              cy={size / 2}
              r={radius}
              fill="none"
              stroke="var(--nadaa-mist)"
              strokeWidth={thickness}
            />
            {total > 0
              ? data.map((datum) => {
                  const length = (datum.value / total) * circumference;
                  const node = (
                    <circle
                      key={datum.label}
                      cx={size / 2}
                      cy={size / 2}
                      r={radius}
                      fill="none"
                      stroke={datum.color}
                      strokeWidth={thickness}
                      strokeDasharray={`${length} ${circumference - length}`}
                      strokeDashoffset={-offset}
                      style={{ transition: "stroke-dasharray 600ms ease" }}
                    />
                  );
                  offset += length;
                  return node;
                })
              : null}
          </g>
        </svg>
        {centerValue !== undefined || centerLabel ? (
          <Box
            sx={{
              position: "absolute",
              inset: 0,
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            {centerValue !== undefined ? (
              <Typography sx={{ fontWeight: 800, fontSize: "1.5rem", lineHeight: 1 }}>
                {centerValue}
              </Typography>
            ) : null}
            {centerLabel ? (
              <Typography variant="caption" sx={{ color: "text.secondary" }}>
                {centerLabel}
              </Typography>
            ) : null}
          </Box>
        ) : null}
      </Box>
      <Stack spacing={0.75} sx={{ minWidth: 0, flex: 1 }}>
        {data.map((datum) => (
          <Stack
            key={datum.label}
            direction="row"
            spacing={1}
            sx={{ alignItems: "center" }}
          >
            <Box
              sx={{
                width: 10,
                height: 10,
                borderRadius: "3px",
                bgcolor: datum.color,
                flexShrink: 0,
              }}
            />
            <Typography variant="body2" sx={{ color: "text.secondary" }} noWrap>
              {datum.label}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 800, ml: "auto" }}>
              {datum.value}
            </Typography>
          </Stack>
        ))}
      </Stack>
    </Stack>
  );
}

/** Horizontal proportional bars — a compact ranked breakdown. */
export function BarRows({
  data,
  max,
}: {
  data: ChartDatum[];
  max?: number;
}) {
  const peak = max ?? Math.max(1, ...data.map((datum) => datum.value));
  return (
    <Stack spacing={1.25}>
      {data.map((datum) => (
        <Box key={datum.label}>
          <Stack
            direction="row"
            sx={{ justifyContent: "space-between", mb: 0.25 }}
          >
            <Typography variant="caption" sx={{ color: "text.secondary" }}>
              {datum.label}
            </Typography>
            <Typography variant="caption" sx={{ fontWeight: 800 }}>
              {datum.value}
            </Typography>
          </Stack>
          <Box
            sx={{
              height: 9,
              borderRadius: 999,
              bgcolor: "var(--nadaa-mist)",
              overflow: "hidden",
            }}
          >
            <Box
              sx={{
                height: "100%",
                borderRadius: 999,
                width: `${(datum.value / peak) * 100}%`,
                bgcolor: datum.color,
                transition: "width 600ms ease",
              }}
            />
          </Box>
        </Box>
      ))}
    </Stack>
  );
}

/** Single-value progress ring (e.g. occupancy or coverage %). */
export function ProgressRing({
  value,
  size = 128,
  thickness = 13,
  color = "var(--nadaa-green)",
  label,
}: {
  value: number;
  size?: number;
  thickness?: number;
  color?: string;
  label?: string;
}) {
  const radius = (size - thickness) / 2;
  const circumference = 2 * Math.PI * radius;
  const pct = Math.max(0, Math.min(100, Math.round(value)));
  const dash = (pct / 100) * circumference;
  return (
    <Box sx={{ position: "relative", width: size, height: size }}>
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${size} ${size}`}
        role="img"
        aria-label={label ? `${label}: ${pct}%` : `${pct}%`}
      >
        <g transform={`rotate(-90 ${size / 2} ${size / 2})`}>
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="var(--nadaa-mist)"
            strokeWidth={thickness}
          />
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth={thickness}
            strokeLinecap="round"
            strokeDasharray={`${dash} ${circumference - dash}`}
            style={{ transition: "stroke-dasharray 700ms ease" }}
          />
        </g>
      </svg>
      <Box
        sx={{
          position: "absolute",
          inset: 0,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Typography sx={{ fontWeight: 800, fontSize: "1.5rem", lineHeight: 1 }}>
          {pct}%
        </Typography>
        {label ? (
          <Typography variant="caption" sx={{ color: "text.secondary" }}>
            {label}
          </Typography>
        ) : null}
      </Box>
    </Box>
  );
}
