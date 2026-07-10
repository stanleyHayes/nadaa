import { CssBaseline, ThemeProvider } from "@mui/material";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { citizenTheme } from "@/app/theme";
import { CitizenLayout } from "./components/CitizenLayout";
import { AlertsPage } from "./pages/AlertsPage";
import { CommunityPage } from "./pages/CommunityPage";
import { GuidesPage } from "./pages/GuidesPage";
import { HomePage } from "./pages/HomePage";
import { ReportPage } from "./pages/ReportPage";
import { SheltersPage } from "./pages/SheltersPage";

/**
 * Router root for the citizen PWA. Every route renders a self-contained page
 * inside the shared `CitizenLayout` chrome (glass header with mobile drawer,
 * persistent emergency 112 band, scroll-to-top).
 */
export default function CitizenApp() {
  return (
    <ThemeProvider theme={citizenTheme}>
      <CssBaseline />
      <BrowserRouter>
        <Routes>
          <Route element={<CitizenLayout />}>
            <Route index element={<HomePage />} />
            <Route element={<AlertsPage />} path="alerts" />
            <Route element={<ReportPage />} path="report" />
            <Route element={<SheltersPage />} path="shelters" />
            <Route element={<GuidesPage />} path="guides" />
            <Route element={<CommunityPage />} path="community" />
            <Route element={<HomePage />} path="*" />
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}
