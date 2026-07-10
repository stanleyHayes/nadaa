import { useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { Download, HandHeart, X } from "lucide-react";
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
    deselectAidRequest,
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

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));
  const [detailOpen, setDetailOpen] = useState(false);

  const openReview = (aidRequestId: string) => {
    selectAidRequest(aidRequestId);
    setDetailOpen(true);
  };

  const openCreate = () => {
    handleNewAidRequest();
    setDetailOpen(true);
  };

  const closeDetail = () => {
    setDetailOpen(false);
    deselectAidRequest();
  };

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
              onClick={openCreate}
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
        <Stack spacing={2}>
          {aidRequests.length === 0 ? (
            <EmptyState message="No donation or aid needs have been created yet." />
          ) : (
            aidRequests.map((request) => (
              <AidRequestCard
                key={request.id}
                onSelect={() => openReview(request.id)}
                request={request}
                selected={selectedAidRequestId === request.id}
              />
            ))
          )}
        </Stack>
      )}

      <Dialog
        open={detailOpen}
        onClose={closeDetail}
        maxWidth="md"
        fullWidth
        scroll="paper"
        fullScreen={fullScreen}
        aria-labelledby="aid-detail-title"
      >
        <DialogTitle
          id="aid-detail-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>
            {selectedAidRequest ? selectedAidRequest.title : "Create aid need"}
          </span>
          <IconButton
            aria-label="Close aid need"
            onClick={closeDetail}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5}>
            <Stack spacing={2}>
              <AidRequestForm
                form={aidForm}
                onChange={setAidForm}
                onSubmit={handleCreateAidRequest}
                submitLabel={
                  aidUpdateState === "loading" ? "Saving..." : "Create for review"
                }
              />
              {selectedAidRequest ? (
                <Stack direction="row" flexWrap="wrap" gap={1}>
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
                <Alert severity="success">Aid coordination record saved.</Alert>
              ) : null}
              {aidUpdateState === "error" && aidError ? (
                <Alert severity="error">{aidError}</Alert>
              ) : null}
            </Stack>
            <Stack spacing={1}>
              <Typography fontWeight={800} variant="h6">
                Partner pledges
              </Typography>
              <AidPledgeList pledges={selectedAidRequest?.pledges ?? []} />
            </Stack>
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
