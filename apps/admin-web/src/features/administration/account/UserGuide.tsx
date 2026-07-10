import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { BookOpen, ChevronRight } from "lucide-react";
import { groupedPageGuides, type GuideKey } from "../pageGuides";

function slug(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "");
}

/**
 * The full user-guide view. Lists every registry entry as a titled card with
 * its numbered steps, grouped by section. Opened from the account menu; each
 * card can jump straight to the workspace it documents.
 */
export function UserGuide({ onOpen }: { onOpen: (key: GuideKey) => void }) {
  const groups = groupedPageGuides();

  return (
    <Box sx={{ maxWidth: 1120, mx: "auto" }}>
      <Paper
        elevation={0}
        sx={{
          display: "flex",
          gap: 2,
          alignItems: "center",
          p: { xs: 2.5, md: 3.5 },
          borderRadius: "16px",
          color: "var(--nadaa-white, #ffffff)",
          background:
            "linear-gradient(150deg, var(--nadaa-navy, #0d1b3d) 0%, #0a1531 100%)",
          boxShadow: "var(--nadaa-shadow-md)",
        }}
      >
        <Box
          aria-hidden
          sx={{
            flex: "0 0 auto",
            display: "grid",
            placeItems: "center",
            width: 52,
            height: 52,
            borderRadius: "13px",
            color: "var(--nadaa-gold, #f4c20d)",
            backgroundColor: "rgba(255, 255, 255, 0.08)",
          }}
        >
          <BookOpen size={24} />
        </Box>
        <Box sx={{ minWidth: 0 }}>
          <Typography
            sx={{
              fontSize: "0.68rem",
              fontWeight: 700,
              letterSpacing: "0.18em",
              textTransform: "uppercase",
              color: "var(--nadaa-gold, #f4c20d)",
            }}
          >
            Help
          </Typography>
          <Typography
            component="h1"
            sx={{
              mt: 0.25,
              fontSize: { xs: "1.6rem", md: "1.9rem" },
              fontWeight: 800,
              lineHeight: 1.1,
            }}
          >
            User guide
          </Typography>
          <Typography
            sx={{
              mt: 0.5,
              fontSize: "0.9rem",
              color: "rgba(255, 255, 255, 0.72)",
            }}
          >
            Step-by-step help for every admin-console workspace. Open any page's
            help button for the same steps, read aloud on demand.
          </Typography>
        </Box>
      </Paper>

      <Stack spacing={4} sx={{ mt: 3 }}>
        {groups.map(({ section, guides }) => (
          <section key={section} aria-labelledby={`guide-${slug(section)}`}>
            <Typography
              id={`guide-${slug(section)}`}
              component="h2"
              className="cc-guide__section"
            >
              {section}
            </Typography>
            <Box className="cc-guide__grid">
              {guides.map((guide) => (
                <Paper key={guide.key} elevation={0} className="cc-guide__card">
                  <div>
                    <Typography component="h3" className="cc-guide__card-title">
                      {guide.title}
                    </Typography>
                    <Typography className="cc-guide__card-desc">
                      {guide.description}
                    </Typography>
                  </div>
                  <ol className="cc-guide__steps">
                    {guide.steps.map((step, index) => (
                      <li key={step}>
                        <span className="cc-guide__num" aria-hidden>
                          {index + 1}
                        </span>
                        <span>{step}</span>
                      </li>
                    ))}
                  </ol>
                  {guide.key !== "guide" ? (
                    <Button
                      size="small"
                      variant="outlined"
                      className="cc-guide__open"
                      endIcon={<ChevronRight size={15} />}
                      onClick={() => onOpen(guide.key)}
                    >
                      Open page
                    </Button>
                  ) : null}
                </Paper>
              ))}
            </Box>
          </section>
        ))}
      </Stack>
    </Box>
  );
}
