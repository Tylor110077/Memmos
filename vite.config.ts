import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

const liveApiTarget = process.env.VITE_LIVE_API_TARGET ?? "http://127.0.0.1:8080";
const useMockApi = process.env.VITE_API_MODE === "mock";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    port: 5173,
    proxy:
      !useMockApi
        ? {
            "/api": {
              target: liveApiTarget,
              changeOrigin: true,
            },
            "/healthz": {
              target: liveApiTarget,
              changeOrigin: true,
            },
            "/readyz": {
              target: liveApiTarget,
              changeOrigin: true,
            },
            "/metrics": {
              target: liveApiTarget,
              changeOrigin: true,
            },
          }
        : undefined,
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    globals: true,
  },
});
