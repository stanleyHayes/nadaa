import { Link } from "react-router-dom";
import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { Compass, House } from "lucide-react";

/**
 * Friendly citizen 404, rendered inside CitizenLayout so the header and the
 * standing emergency band remain. Reuses the animated empty-state icon badge.
 */
export function NotFoundPage() {
  return (
    <Box component="section" sx={{ py: { xs: 4, md: 6 } }}>
      <Paper
        className="surface"
        sx={{ maxWidth: 560, mx: "auto", textAlign: "center", p: { xs: 3, md: 5 } }}
      >
        <Stack spacing={1.5} sx={{ alignItems: "center" }}>
          <span aria-hidden="true" className="empty-state__icon">
            <Compass size={30} strokeWidth={1.75} />
          </span>
          <Typography sx={{ fontWeight: 800, letterSpacing: "0.08em" }} variant="h4">
            404
          </Typography>
          <Typography sx={{ fontWeight: 700 }} variant="h6">
            We couldn't find that page
          </Typography>
          <Typography sx={{ color: "text.secondary", maxWidth: 420 }} variant="body2">
            The page may have moved or the link may be incomplete. You can head
            back home to check your risk, report an incident, or find a shelter.
          </Typography>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1.25}
            sx={{ pt: 1, width: { xs: "100%", sm: "auto" } }}
          >
            <Button
              component={Link}
              startIcon={<House size={18} />}
              to="/"
              variant="contained"
            >
              Back to home
            </Button>
            <Button component={Link} to="/shelters" variant="outlined">
              Find a shelter
            </Button>
          </Stack>
          <Typography sx={{ color: "text.secondary", pt: 1 }} variant="caption">
            In an emergency, call <a href="tel:112">112</a>.
          </Typography>
        </Stack>
      </Paper>
    </Box>
  );
}
