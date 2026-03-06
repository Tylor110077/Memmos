import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "@/app/App";
import "@/styles/index.css";
import "@xyflow/react/dist/style.css";

async function bootstrap() {
  if (import.meta.env.DEV && import.meta.env.VITE_API_MODE !== "live") {
    const { startMockWorker } = await import("@/mocks/browser");
    await startMockWorker();
  }

  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
}

void bootstrap();
