import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { ThemeProvider } from "@mui/material";
import { CssBaseline } from "@mui/material";
import App from "./App";
import { agencyTheme } from "./app/theme";
import "./styles/global.css";

createRoot(document.getElementById("root") as HTMLElement).render(
  <StrictMode>
    <ThemeProvider theme={agencyTheme}>
      <CssBaseline />
      <App />
    </ThemeProvider>
  </StrictMode>,
);
