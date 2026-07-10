import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { AgencyThemeProvider } from "./app/AgencyThemeProvider";
import "./styles/global.css";

createRoot(document.getElementById("root") as HTMLElement).render(
  <StrictMode>
    <AgencyThemeProvider>
      <App />
    </AgencyThemeProvider>
  </StrictMode>,
);
