import {
  Alert,
  Box,
  Button,
  Chip,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  Typography,
} from "@mui/material";

import "leaflet/dist/leaflet.css";
import { RefreshCw, Truck } from "lucide-react";
import { nadaaBrand } from "@nadaa/brand";
import type { ReliefPointRecord } from "@nadaa/shared-types";

import type { CapacityLoadState } from "../types";

import { EmptyState } from "./shared";

export function ReliefPointPanel({
  loadMessage,
  loadState,
  onRefresh,
  reliefPoints,
}: {
  loadMessage: string;
  loadState: CapacityLoadState;
  onRefresh: () => void;
  reliefPoints: ReliefPointRecord[];
}) {
  return (
    <Paper className="surface capacity-panel">
      <Stack
        direction={{ xs: "column", md: "row" }}
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          gap: 1.5
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <Truck size={22} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h5">Relief distribution points</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Food, water, medical, hygiene, and blanket distribution locations
              for affected communities.
            </Typography>
          </Box>
        </Stack>
        <Stack
          direction="row"
          spacing={1}
          sx={{
            alignItems: "center",
            flexWrap: "wrap"
          }}>
          <Chip
            size="small"
            label={
              loadState === "ready"
                ? "Live relief points"
                : loadState === "loading"
                  ? "Loading"
                  : loadState === "empty"
                    ? "No matches"
                    : "Unavailable"
            }
            color={
              loadState === "ready"
                ? "success"
                : loadState === "empty"
                  ? "default"
                  : "warning"
            }
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshCw size={16} />}
            onClick={onRefresh}
          >
            Refresh
          </Button>
        </Stack>
      </Stack>
      {loadState === "loading" ? (
        <LinearProgress />
      ) : loadState === "fallback" || loadState === "empty" ? (
        <Alert severity={loadState === "empty" ? "info" : "warning"}>
          {loadMessage}
        </Alert>
      ) : null}
      {reliefPoints.length > 0 ? (
        <Grid container spacing={2}>
          {reliefPoints.map((point) => (
            <Grid size={{ xs: 12, md: 6, lg: 4 }} key={point.id}>
              <Box className="capacity-card">
                <Stack spacing={1}>
                  <Stack
                    direction="row"
                    sx={{
                      justifyContent: "space-between",
                      alignItems: "center"
                    }}>
                    <Typography variant="subtitle1" sx={{
                      fontWeight: 700
                    }}>
                      {point.name}
                    </Typography>
                    <Chip label={point.status} size="small" />
                  </Stack>
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    {point.type} · {point.district}
                  </Typography>
                  <Typography variant="body2">{point.address}</Typography>
                  <Stack
                    direction="row"
                    sx={{
                      flexWrap: "wrap",
                      gap: 0.5
                    }}>
                    {point.stockCategories.map((stock) => (
                      <Chip
                        key={stock.category}
                        label={`${stock.category}: ${stock.quantity} ${stock.unit}`}
                        size="small"
                        variant="outlined"
                      />
                    ))}
                  </Stack>
                  {point.operatingHours ? (
                    <Typography variant="caption" sx={{
                      color: "text.secondary"
                    }}>
                      Hours: {point.operatingHours}
                    </Typography>
                  ) : null}
                  {point.eligibility ? (
                    <Alert severity="info" sx={{ py: 0.5 }}>
                      {point.eligibility}
                    </Alert>
                  ) : null}
                </Stack>
              </Box>
            </Grid>
          ))}
        </Grid>
      ) : (
        <EmptyState
          title="No relief points"
          detail="No relief distribution points are currently available."
        />
      )}
    </Paper>
  );
}
