import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  server: {
    proxy: {
      "/mcp": {
        target: process.env.NANOBOT_URL || "http://localhost:9999",
      },
    },
  },
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],
});
