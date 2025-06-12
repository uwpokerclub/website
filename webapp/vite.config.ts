import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    proxy: {
      "/api": process.env.VITE_API_HOST || "http://localhost:5000",
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
      "@components": path.resolve(__dirname, "src/components"),
      "@hooks": path.resolve(__dirname, "src/hooks"),
      "@contexts": path.resolve(__dirname, "src/contexts"),
      "@interfaces": path.resolve(__dirname, "src/interfaces"),
    },
  },
});
