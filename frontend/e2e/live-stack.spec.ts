import { readFileSync } from "node:fs";
import path from "node:path";
import { expect, test } from "@playwright/test";

function parseEnvFile(filePath: string): Record<string, string> {
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

const envPath =
  process.env.E2E_LIVE_ENV_FILE ?? path.resolve(process.cwd(), "..", ".env.docker.local");
const envFile = parseEnvFile(envPath);
const adminUser = process.env.E2E_LIVE_ADMIN_USER ?? envFile.BOOTSTRAP_ADMIN_USER ?? "admin";
const adminPassword = process.env.E2E_LIVE_ADMIN_PASSWORD ?? envFile.BOOTSTRAP_ADMIN_PASSWORD;

if (!adminPassword) {
  throw new Error("E2E_LIVE_ADMIN_PASSWORD or BOOTSTRAP_ADMIN_PASSWORD must be set");
}

test("real stack admin flow is usable", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByTestId("home-hero")).toBeVisible();
  await expect(page.getByTestId("system-status-root")).toBeVisible();

  await page.getByTestId("login-username").fill(adminUser);
  await page.getByTestId("login-password").fill(adminPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("runtime-knobs-panel")).toBeVisible();
  await expect(page.getByTestId("runtime-knobs-panel")).toContainText(
    "REQUEST_LOG_TRUSTED_PROXY_CIDRS",
  );
  await expect(page.getByTestId("dependency-policies-panel")).toBeVisible();
  await expect(page.getByTestId("runtime-metrics-summary")).toBeVisible();
  await expect(page.getByTestId("trace-lookup-panel")).toBeVisible();
  await expect(page.getByTestId("request-log-row").first()).toBeVisible();
});
