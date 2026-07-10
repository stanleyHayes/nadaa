import { Box, Chip, Grid, Paper, Stack, Typography } from "@mui/material";
import { DatabaseZap } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { DataSourceSummary } from "../types";
import { statusColor } from "../utils";

type ChipColor =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

export function DataSourcePanel({
  dataSources,
}: {
  dataSources: DataSourceSummary[];
}) {
  return (
    <Grid container spacing={2}>
      {dataSources.map((source) => (
        <Grid key={source.id} size={{ xs: 12, md: 6, xl: 4 }}>
          <Paper className="data-source-card">
            <Stack spacing={1.2}>
              <Stack
                direction="row"
                sx={{
                  justifyContent: "space-between",
                  gap: 1
                }}>
                <DatabaseZap size={24} color={nadaaBrand.colors.green} />
                <Chip
                  size="small"
                  color={statusColor(source.status) as ChipColor}
                  label={source.status}
                />
              </Stack>
              <Box>
                <Typography variant="h6">{source.partner}</Typography>
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  {source.domain} / {source.direction}
                </Typography>
              </Box>
              <Typography variant="body2">{source.cadence}</Typography>
              <Stack direction="row" spacing={1} sx={{
                flexWrap: "wrap"
              }}>
                <Chip size="small" label={`PII: ${source.pii}`} />
                <Chip
                  size="small"
                  label={`Fresh ${source.freshnessWindowMinutes}m`}
                />
                <Chip
                  size="small"
                  label={`Auth: ${source.authenticationMode}`}
                />
              </Stack>
              <Typography variant="caption" sx={{
                color: "text.secondary"
              }}>
                Secret scope: {source.secretScope ?? "none configured"}
              </Typography>
              <Typography variant="caption" sx={{
                color: "text.secondary"
              }}>
                Fallback: {source.manualFallback}
              </Typography>
            </Stack>
          </Paper>
        </Grid>
      ))}
    </Grid>
  );
}
