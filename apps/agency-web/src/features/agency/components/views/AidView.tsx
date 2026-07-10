import { Alert, Button, Grid, Paper, Stack, Typography } from "@mui/material";
import { Download, HandHeart } from "lucide-react";
import type { AgencyData } from "../../useAgencyData";
import { ViewIntro } from "../primitives";
import {
  AidPledgeList,
  AidRequestCard,
  AidRequestForm,
  EmptyState,
  ErrorState,
  LoadingState,
} from "..";

export function AidView({ data }: { data: AgencyData }) {
  const {
    aidRequests,
    selectedAidRequest,
    selectedAidRequestId,
    selectAidRequest,
    aidForm,
    setAidForm,
    aidLoadState,
    aidUpdateState,
    aidError,
    loadAidRequests,
    handleCreateAidRequest,
    handleReviewAidRequest,
    handleNewAidRequest,
    handleAidExport,
  } = data;

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        title="Aid & donations"
        description="Create verified aid needs, review partner pledges, and export coordination reports without changing incident status."
        icon={HandHeart}
        action={
          <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
            <Button
              onClick={handleAidExport}
              startIcon={<Download size={18} />}
              variant="outlined"
            >
              Export CSV
            </Button>
            <Button
              onClick={handleNewAidRequest}
              startIcon={<HandHeart size={18} />}
              variant="contained"
            >
              New aid need
            </Button>
          </Stack>
        }
      />

      {aidError && aidLoadState === "fallback" ? (
        <Alert severity="warning">{aidError} Showing fixture aid requests.</Alert>
      ) : null}

      {aidLoadState === "loading" ? (
        <LoadingState message="Loading donation and aid requests" />
      ) : aidLoadState === "error" && !aidRequests.length ? (
        <ErrorState
          message={aidError ?? "Could not load aid requests"}
          onRetry={loadAidRequests}
        />
      ) : (
        <Grid container spacing={2.5} alignItems="flex-start">
          <Grid size={{ xs: 12, lg: 5 }}>
            <Stack spacing={2}>
              {aidRequests.length === 0 ? (
                <EmptyState message="No donation or aid needs have been created yet." />
              ) : (
                aidRequests.map((request) => (
                  <AidRequestCard
                    key={request.id}
                    onSelect={() => selectAidRequest(request.id)}
                    request={request}
                    selected={selectedAidRequestId === request.id}
                  />
                ))
              )}
            </Stack>
          </Grid>
          <Grid size={{ xs: 12, lg: 7 }}>
            <Stack spacing={2.5}>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  {selectedAidRequest ? "Review aid need" : "Create aid need"}
                </Typography>
                <AidRequestForm
                  form={aidForm}
                  onChange={setAidForm}
                  onSubmit={handleCreateAidRequest}
                  submitLabel={
                    aidUpdateState === "loading"
                      ? "Saving..."
                      : "Create for review"
                  }
                />
                {selectedAidRequest ? (
                  <Stack direction="row" flexWrap="wrap" gap={1} mt={2}>
                    <Button
                      disabled={aidUpdateState === "loading"}
                      onClick={() => void handleReviewAidRequest("approved")}
                      size="small"
                      variant="contained"
                    >
                      Approve listing
                    </Button>
                    <Button
                      disabled={aidUpdateState === "loading"}
                      onClick={() => void handleReviewAidRequest("paused")}
                      size="small"
                      variant="outlined"
                    >
                      Pause
                    </Button>
                    <Button
                      disabled={aidUpdateState === "loading"}
                      onClick={() => void handleReviewAidRequest("closed")}
                      size="small"
                      variant="outlined"
                    >
                      Close
                    </Button>
                  </Stack>
                ) : null}
                {aidUpdateState === "success" ? (
                  <Alert severity="success" sx={{ mt: 2 }}>
                    Aid coordination record saved.
                  </Alert>
                ) : null}
                {aidUpdateState === "error" && aidError ? (
                  <Alert severity="error" sx={{ mt: 2 }}>
                    {aidError}
                  </Alert>
                ) : null}
              </Paper>
              <Paper sx={{ p: 3 }}>
                <Typography fontWeight={800} gutterBottom variant="h6">
                  Partner pledges
                </Typography>
                <AidPledgeList pledges={selectedAidRequest?.pledges ?? []} />
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      )}
    </Stack>
  );
}
