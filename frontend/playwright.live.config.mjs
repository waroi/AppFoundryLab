import { readFileSync } from "node:fs";
import path from "node:path";
import { defineConfig } from "@playwright/test";

function parseEnvFile(filePath) {
  try {
    const body = readFileSync(filePath, "utf8");
    return Object.fromEntries(
      body
        .split(/\r?\n/)
        .filter((line) => line && !line.startsWith("#") && line.includes("="))
        .map((line) => {
          const [key, ...rest] = line.split("=");
          return [key, rest.join("=")];
        }),
    );
  } catch {
    return {};
  }
}

function browserHost(bindAddress) {
  if (!bindAddress || bindAddress === "0.0.0.0" || bindAddress === "::" || bindAddress === "[::]") {
    return "127.0.0.1";
  }
  return bindAddress;
}

const rootEnvPath =
  process.env.E2E_LIVE_ENV_FILE ?? path.resolve(process.cwd(), "..", ".env.docker.local");
const envFile = parseEnvFile(rootEnvPath);
const frontendPort = process.env.E2E_LIVE_FRONTEND_PORT ?? envFile.FRONTEND_HOST_PORT ?? "4321";
const bindAddress =
  process.env.E2E_LIVE_BIND_ADDRESS ?? envFile.DOCKER_HOST_BIND_ADDRESS ?? "127.0.0.1";
const baseURL =
  process.env.E2E_LIVE_BASE_URL ?? `http://${browserHost(bindAddress)}:${frontendPort}`;

export default defineConfig({
  testDir: "./e2e",
  testMatch: /live-stack\.spec\.ts/,
  timeout: 45000,
  retries: 0,
  use: {
    baseURL,
    headless: true,
  },
});
