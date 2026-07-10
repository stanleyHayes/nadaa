import { Alert, Button, Grid, Paper, Stack, Typography } from "@mui/material";
import { PackageCheck } from "lucide-react";
import type { AgencyData } from "../../useAgencyData";
import { ViewIntro } from "../primitives";
import {
  EmptyState,
  ErrorState,
  LoadingState,
  ReliefPointCard,
  ReliefPointForm,
  ReliefStockHistoryList,
} from "..";

export function ReliefView({ data }: { data: AgencyData }) {
  const {
    reliefPoints,
    selectedReliefPoint,
    selectedReliefPointId,
    selectReliefPoint,
    reliefForm,
    setReliefForm,
    reliefHistory,
    reliefLoadState,
    reliefUpdateState,
    reliefError,
    loadReliefPoints,
    handleSaveReliefPoint,
    handleNewReliefPoint,
  } = data;

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        title="Relief distribution"
        description="Publish distribution points, update stock, and keep eligibility notes current."
        icon={PackageCheck}
        action={
          <Button
            onClick={handleNewReliefPoint}
            startIcon={<PackageCheck size={18} />}
            variant="contained"
          >
            New point
          </Button>
        }
      />

      {reliefError && reliefLoadState === "fallback" ? (
        <Alert severity="warning">
          {reliefError} Showing fixture relief distribution points.
        </Alert>
      ) : null}

      {reliefLoadState === "loading" ? (
        <LoadingState message="Loading relief distribution points" />
      ) : reliefLoadState === "error" && !reliefPoints.length ? (
        <ErrorState
          message={reliefError ?? "Could not load relief points"}
          onRetry={loadReliefPoints}
        />
      ) : (
        <Grid container spacing={2.5} alignItems="flex-start">
          <Grid size={{ xs: 12, lg: 5 }}>
            <Stack spacing={2}>
              {reliefPoints.length === 0 ? (
                <EmptyState message="No relief distribution points have been published yet." />
              ) : (
                reliefPoints.map((point) => (
                  <ReliefPointCard
                    key={point.id}
                    onSelect={() => selectReliefPoint(point.id)}
                    point={point}
                    selected={selectedReliefPointId === point.id}
                  />
                ))
              )}
            </Stack>
          </Grid>
          <Grid size={{ xs: 12, lg: 7 }}>
            <Stack spacing={2.5}>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  {selectedReliefPoint
                    ? "Manage distribution point"
                    : "Create distribution point"}
                </Typography>
                <ReliefPointForm
                  form={reliefForm}
                  onChange={setReliefForm}
                  onSubmit={handleSaveReliefPoint}
                  submitLabel={
                    reliefUpdateState === "loading"
                      ? "Saving..."
                      : selectedReliefPoint
                        ? "Update point"
                        : "Create point"
                  }
                />
                {reliefUpdateState === "success" ? (
                  <Alert severity="success" sx={{ mt: 2 }}>
                    Relief distribution point saved.
                  </Alert>
                ) : null}
                {reliefUpdateState === "error" && reliefError ? (
                  <Alert severity="error" sx={{ mt: 2 }}>
                    {reliefError}
                  </Alert>
                ) : null}
              </Paper>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  Stock history
                </Typography>
                <ReliefStockHistoryList history={reliefHistory} />
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      )}
    </Stack>
  );
}
