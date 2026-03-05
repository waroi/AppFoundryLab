import { defineConfig } from "@playwright/test";

const port = Number(process.env.E2E_PORT ?? 4173);

export default defineConfig({
  testDir: "./e2e",
  timeout: 30000,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    headless: true,
  },
  webServer: {
    command: "./node_modules/.bin/astro build && node ./scripts/e2e-server.mjs",
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: false,
    timeout: 180000,
    stdout: "pipe",
    stderr: "pipe",
  },
});
