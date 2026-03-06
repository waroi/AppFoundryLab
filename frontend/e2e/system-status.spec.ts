import { expect, test } from "@playwright/test";

test.use({ viewport: { width: 1280, height: 1400 } });

test("locale and theme shell remains visually stable", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByTestId("system-status-root")).toBeVisible();
  await expect(page.getByTestId("system-status-root")).toHaveScreenshot(
    "system-status-en-light.png",
  );

  await page.getByTestId("locale-tr").click();
  await page.getByTestId("theme-dark").click();

  await expect(page.locator("html")).toHaveAttribute("lang", "tr");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await expect(page.getByTestId("system-status-root")).toHaveScreenshot(
    "system-status-tr-dark.png",
  );
});

test("admin trace lookup shows correlated request logs", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByTestId("system-status-root")).toBeVisible();
  await page.getByTestId("locale-tr").click();
  await expect(page).toHaveURL(/\/tr\/?$/);
  await page.getByTestId("theme-dark").click();
  await expect(page.locator("html")).toHaveAttribute("lang", "tr");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await expect(page.getByTestId("system-status-root")).toHaveScreenshot(
    "system-status-home-tr-dark.png",
  );

  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("admin_dev_password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("trace-lookup-panel")).toBeVisible();
  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-mode", "latest");
  await expect(page.getByTestId("runtime-metrics-summary")).toBeVisible();

  await expect(page.getByTestId("request-log-row")).toHaveCount(2);
  await page.getByTestId("incident-trace-trace-admin-a").click();

  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-mode", "filtered");
  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute(
    "data-trace-id",
    "trace-admin-a",
  );
  await expect(page.getByTestId("request-log-row")).toHaveCount(1);
  await expect(page.getByTestId("request-log-row").first()).toContainText(
    "/api/v1/admin/request-logs",
  );

  await page.getByTestId("trace-query-input").fill("trace-login-a");
  await page.getByTestId("trace-search-button").click();
  await expect(page.getByTestId("request-log-row")).toHaveCount(1);
  await expect(page.getByTestId("request-log-row").first()).toContainText("/api/v1/auth/token");

  await page.getByTestId("trace-latest-button").click();
  await expect(page.getByTestId("request-log-row")).toHaveCount(2);
  await expect(page.getByTestId("system-status-root")).toHaveScreenshot(
    "system-status-admin-runtime.png",
  );
});

test("trace lookup renders an empty state for unknown traces", async ({ page }) => {
  await page.goto("/");

  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("admin_dev_password");
  await page.getByTestId("login-submit").click();

  await page.getByTestId("trace-query-input").fill("trace-missing");
  await page.getByTestId("trace-search-button").click();

  await expect(page.getByTestId("trace-empty-state")).toBeVisible();
  await expect(page.getByTestId("request-log-row")).toHaveCount(0);
});

test("invalid credentials surface an auth error without breaking the shell", async ({ page }) => {
  await page.goto("/");
  await page.getByTestId("login-username").fill("wrong-user");
  await page.getByTestId("login-password").fill("wrong-password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-error")).toContainText("Invalid username or password.");
  await expect(page.getByTestId("auth-role")).toHaveCount(0);
});

test("admin runtime-report failure leaves the authenticated panel in a degraded state", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByTestId("login-username").fill("runtime-error");
  await page.getByTestId("login-password").fill("admin_dev_password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("runtime-error")).toContainText(
    "Runtime diagnostics are temporarily unavailable.",
  );
  await expect(page.getByTestId("runtime-metrics-summary")).toHaveCount(0);
});

test("trace lookup shows the no-match state for unknown trace ids", async ({ page }) => {
  await page.goto("/");

  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("admin_dev_password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await page.getByTestId("trace-query-input").fill("trace-does-not-exist");
  await page.getByTestId("trace-search-button").click();

  await expect(page.getByTestId("trace-lookup-panel")).toContainText(
    "No request logs matched this trace.",
  );
  await expect(page.getByTestId("request-log-row")).toHaveCount(0);
});

test("fibonacci action requires login before execution", async ({ page }) => {
  await page.goto("/");

  await page.getByText("Compute").click();
  await expect(page.getByTestId("fibonacci-error")).toContainText(
    "Sign in before running this action.",
  );
});

test("restore drill preview renders sample verification artifacts", async ({ page }) => {
  await page.goto("/test/");

  await expect(page.getByTestId("test-page-intro")).toBeVisible();
  await page.getByTestId("locale-tr").click();
  await expect(page).toHaveURL(/\/tr\/test\/?$/);
  await expect(page.locator("html")).toHaveAttribute("lang", "tr");
  await expect(page.getByTestId("restore-drill-preview")).toBeVisible();
  await expect(page.getByTestId("restore-drill-marker")).toHaveText("restore-drill-sample");
  await expect(page.getByTestId("restore-drill-status")).toHaveAttribute("data-status", "ok");
  await expect(page.getByTestId("restore-drill-manifest-line")).toContainText([
    "fixture_marker=restore-drill-sample",
    "verification_file=fixture-verification-restore-drill-sample.json",
  ]);
});
