import path from "node:path";
import { defineConfig } from "@playwright/test";

const port = Number(process.env.E2E_PORT ?? 4173);
const bunBin =
  process.env.BUN_BIN ?? path.resolve(process.cwd(), "..", ".toolchain", "bun", "bin", "bun");

export default defineConfig({
  testDir: "./e2e",
  testIgnore: /live-stack\.spec\.ts/,
  timeout: 30000,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    headless: true,
  },
  webServer: {
    command: `${bunBin} run build && node ./scripts/e2e-server.mjs`,
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: false,
    timeout: 180000,
    stdout: "pipe",
    stderr: "pipe",
  },
});
