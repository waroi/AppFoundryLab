import { expect, test } from "@playwright/test";

test("admin trace lookup shows correlated request logs", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByTestId("system-status-root")).toBeVisible();
  await page.getByTestId("locale-tr").click();
  await expect(page).toHaveURL(/\/tr\/?$/);
  await page.getByTestId("theme-dark").click();
  await expect(page.locator("html")).toHaveAttribute("lang", "tr");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("admin_dev_password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("trace-lookup-panel")).toBeVisible();
  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-mode", "latest");

  await expect(page.getByTestId("request-log-row")).toHaveCount(2);
  await page.getByTestId("incident-trace-trace-admin-a").click();

  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-mode", "filtered");
  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-trace-id", "trace-admin-a");
  await expect(page.getByTestId("request-log-row")).toHaveCount(1);
  await expect(page.getByTestId("request-log-row").first()).toContainText("/api/v1/admin/request-logs");

  await page.getByTestId("trace-query-input").fill("trace-login-a");
  await page.getByTestId("trace-search-button").click();
  await expect(page.getByTestId("request-log-row")).toHaveCount(1);
  await expect(page.getByTestId("request-log-row").first()).toContainText("/api/v1/auth/token");

  await page.getByTestId("trace-latest-button").click();
  await expect(page.getByTestId("request-log-row")).toHaveCount(2);
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
