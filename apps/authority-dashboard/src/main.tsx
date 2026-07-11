import "@fontsource/outfit/400.css";
import "@fontsource/outfit/500.css";
import "@fontsource/outfit/600.css";
import "@fontsource/outfit/700.css";
import "@fontsource/outfit/800.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { installUnauthorizedRedirect } from "./app/authGuard";
import { signOutAuthority } from "./app/session";
import "./styles/global.css";

installUnauthorizedRedirect(signOutAuthority);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
