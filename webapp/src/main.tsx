import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "@uwpokerclub/components/tokens.css";
import "@uwpokerclub/components/components.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
