import { useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { createCitizenTheme } from "@/app/theme";
import { useThemeMode } from "@/app/theme-mode";
import { CitizenLayout } from "./components/CitizenLayout";
import {
  AccountLayout,
  AccountNotifications,
  AccountOverview,
  AccountReports,
  AccountSettings,
} from "./pages/account";
import { AlertsPage } from "./pages/AlertsPage";
import { CommunityPage } from "./pages/CommunityPage";
import { GuidesPage } from "./pages/GuidesPage";
import { HomePage } from "./pages/HomePage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { ReportPage } from "./pages/ReportPage";
import { SheltersPage } from "./pages/SheltersPage";

/**
 * Router root for the citizen PWA. Every route renders a self-contained page
 * inside the shared `CitizenLayout` chrome (glass header with mobile drawer,
 * persistent emergency 112 band, scroll-to-top).
 */
export default function CitizenApp() {
  const mode = useThemeMode();
  const theme = useMemo(() => createCitizenTheme({ mode }), [mode]);

  return (
    <ThemeProvider theme={theme}>
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
            <Route element={<AccountLayout />} path="account">
              <Route index element={<AccountOverview />} />
              <Route element={<AccountReports />} path="reports" />
              <Route element={<AccountNotifications />} path="notifications" />
              <Route element={<AccountSettings />} path="settings" />
            </Route>
            <Route element={<NotFoundPage />} path="*" />
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}
