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
import { PackageCheck, X } from "lucide-react";
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
    deselectReliefPoint,
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

  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down("sm"));
  const [detailOpen, setDetailOpen] = useState(false);

  const openManage = (reliefPointId: string) => {
    selectReliefPoint(reliefPointId);
    setDetailOpen(true);
  };

  const openCreate = () => {
    handleNewReliefPoint();
    setDetailOpen(true);
  };

  const closeDetail = () => {
    setDetailOpen(false);
    deselectReliefPoint();
  };

  return (
    <Stack spacing={2.5}>
      <ViewIntro
        title="Relief distribution"
        description="Publish distribution points, update stock, and keep eligibility notes current."
        icon={PackageCheck}
        action={
          <Button
            onClick={openCreate}
            startIcon={<PackageCheck size={18} />}
            variant="contained"
          >
            New point
          </Button>
        }
      />
      {reliefLoadState === "loading" ? (
        <LoadingState message="Loading relief distribution points" />
      ) : reliefLoadState === "error" && !reliefPoints.length ? (
        <ErrorState
          message={reliefError ?? "Could not load relief points"}
          onRetry={loadReliefPoints}
        />
      ) : (
        <Stack spacing={2}>
          {reliefPoints.length === 0 ? (
            <EmptyState message="No relief distribution points have been published yet." />
          ) : (
            reliefPoints.map((point) => (
              <ReliefPointCard
                key={point.id}
                onSelect={() => openManage(point.id)}
                point={point}
                selected={selectedReliefPointId === point.id}
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
        aria-labelledby="relief-detail-title"
      >
        <DialogTitle
          id="relief-detail-title"
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 1,
          }}
        >
          <span>
            {selectedReliefPoint
              ? selectedReliefPoint.name
              : "Create distribution point"}
          </span>
          <IconButton
            aria-label="Close distribution point"
            onClick={closeDetail}
            size="small"
          >
            <X size={18} />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5}>
            <Stack spacing={2}>
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
                <Alert severity="success">
                  Relief distribution point saved.
                </Alert>
              ) : null}
              {reliefUpdateState === "error" && reliefError ? (
                <Alert severity="error">{reliefError}</Alert>
              ) : null}
            </Stack>
            <Stack spacing={1}>
              <Typography variant="h6" sx={{
                fontWeight: 800
              }}>
                Stock history
              </Typography>
              <ReliefStockHistoryList history={reliefHistory} />
            </Stack>
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
