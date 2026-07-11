import { BrowserRouter, Route, Routes } from "react-router-dom";
import { SiteLayout } from "./components/SiteLayout";
import { ContactPage } from "./pages/ContactPage";
import { HomePage } from "./pages/HomePage";
import { HowItWorksPage } from "./pages/HowItWorksPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { PlatformsPage } from "./pages/PlatformsPage";
import { SignupPage } from "./pages/SignupPage";
import { TrustPage } from "./pages/TrustPage";

export default function MarketingApp() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<SiteLayout />}>
          <Route index element={<HomePage />} />
          <Route element={<PlatformsPage />} path="platforms" />
          <Route element={<HowItWorksPage />} path="how-it-works" />
          <Route element={<TrustPage />} path="trust" />
          <Route element={<SignupPage />} path="signup" />
          <Route element={<ContactPage />} path="contact" />
          <Route element={<NotFoundPage />} path="*" />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
