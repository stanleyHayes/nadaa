import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { installUnauthorizedRedirect } from "./app/authGuard";
import { signOutAgency } from "./app/session";
import { AgencyThemeProvider } from "./app/AgencyThemeProvider";
import "./styles/global.css";

installUnauthorizedRedirect(signOutAgency);

createRoot(document.getElementById("root") as HTMLElement).render(
  <StrictMode>
    <AgencyThemeProvider>
      <App />
    </AgencyThemeProvider>
  </StrictMode>,
);
