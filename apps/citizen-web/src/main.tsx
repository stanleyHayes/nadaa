import "@fontsource/outfit/400.css";
import "@fontsource/outfit/500.css";
import "@fontsource/outfit/600.css";
import "@fontsource/outfit/700.css";
import "@fontsource/outfit/800.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { registerCitizenServiceWorker } from "./features/citizen/utils";
import "./styles/global.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);

// Registers /sw.js in production builds only (no-op in dev or without SW
// support), which backs the "available offline" guide copy.
registerCitizenServiceWorker();
