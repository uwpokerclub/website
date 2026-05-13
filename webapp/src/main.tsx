import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import { QueryProvider } from "./lib/QueryProvider";
import { ErrorBoundary } from "./components";
import "@uwpokerclub/components/tokens.css";
import "@uwpokerclub/components/components.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <QueryProvider>
        <App />
      </QueryProvider>
    </ErrorBoundary>
  </React.StrictMode>,
);
