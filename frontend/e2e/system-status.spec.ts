import { expect, test } from "@playwright/test";

test.use({ viewport: { width: 1280, height: 1400 } });

const adminUser = process.env.E2E_ADMIN_USER ?? "admin";
const adminPassword = process.env.E2E_ADMIN_PASSWORD ?? "mock-admin-password";
const degradedAdminUser = process.env.E2E_DEGRADED_ADMIN_USER ?? "degraded-admin";
const runtimeErrorUser = process.env.E2E_RUNTIME_ERROR_USER ?? "runtime-error";
const invalidUser = process.env.E2E_INVALID_USER ?? "wrong-user";
const invalidPassword = process.env.E2E_INVALID_PASSWORD ?? "wrong-password";

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

  await page.getByTestId("login-username").fill(adminUser);
  await page.getByTestId("login-password").fill(adminPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("trace-lookup-panel")).toBeVisible();
  await expect(page.getByTestId("trace-lookup-state")).toHaveAttribute("data-mode", "latest");
  await expect(page.getByTestId("runtime-metrics-summary")).toBeVisible();
  await expect(page.getByTestId("runtime-knobs-panel")).toBeVisible();
  await expect(page.getByTestId("runtime-knobs-panel")).toContainText(
    "REQUEST_LOG_TRUSTED_PROXY_CIDRS",
  );
  await expect(page.getByTestId("runtime-knobs-panel")).toContainText("127.0.0.1/32");
  await expect(page.getByTestId("dependency-policies-panel")).toBeVisible();
  await expect(page.getByTestId("dependency-policies-panel")).toContainText(
    "Rotalara gore bagimlilik davranisi",
  );
  await expect(page.getByTestId("dependency-policy-row")).toHaveCount(6);

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

test("keyboard flow reaches auth and trace controls with accessible labels", async ({ page }) => {
  await page.goto("/");

  await page.keyboard.press("Tab");
  await expect(page.getByTestId("locale-en")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("locale-tr")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("theme-light")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("theme-dark")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("login-username")).toBeFocused();

  await page.getByTestId("login-username").fill(adminUser);
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("login-password")).toBeFocused();
  await page.getByTestId("login-password").fill(adminPassword);
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("login-submit")).toBeFocused();
  await page.keyboard.press("Enter");

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByRole("textbox", { name: "username" })).toBeVisible();
  await expect(page.getByRole("table", { name: "Dependency behavior by route" })).toBeVisible();

  await page.getByTestId("trace-query-input").focus();
  await expect(page.getByTestId("trace-query-input")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("trace-search-button")).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByTestId("trace-latest-button")).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(page.getByTestId("trace-search-button")).toBeFocused();
});

test("trace lookup renders an empty state for unknown traces", async ({ page }) => {
  await page.goto("/");

  await page.getByTestId("login-username").fill(adminUser);
  await page.getByTestId("login-password").fill(adminPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("dependency-policies-panel")).toContainText(
    "Dependency behavior by route",
  );
  await expect(page.getByTestId("dependency-policies-panel")).toContainText("GET /api/v1/users");
  await page.getByTestId("trace-query-input").fill("trace-missing");
  await page.getByTestId("trace-search-button").click();

  await expect(page.getByTestId("trace-empty-state")).toBeVisible();
  await expect(page.getByTestId("request-log-row")).toHaveCount(0);
});

test("degraded admin diagnostics keep the dependency matrix and runtime knob fallbacks visible", async ({
  page,
}) => {
  await page.goto("/");

  await page.getByTestId("login-username").fill(degradedAdminUser);
  await page.getByTestId("login-password").fill(adminPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("runtime-knobs-panel")).toContainText(
    "No trusted proxies configured; request logs use the direct remote address.",
  );
  await expect(page.getByTestId("runtime-knob-ingest-max-age")).toContainText("600 seconds");
  await expect(page.getByRole("table", { name: "Dependency behavior by route" })).toBeVisible();
  await expect(page.getByTestId("dependency-policies-panel")).toContainText(
    "gateway startup continues and readiness stays degraded until the dependency recovers",
  );
  await expect(page.getByTestId("dependency-policies-panel")).toContainText(
    "returns 503 with per-dependency checks while any required dependency is down",
  );
});

test("invalid credentials surface an auth error without breaking the shell", async ({ page }) => {
  await page.goto("/");
  await page.getByTestId("login-username").fill(invalidUser);
  await page.getByTestId("login-password").fill(invalidPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-error")).toContainText("Invalid username or password.");
  await expect(page.getByTestId("auth-role")).toHaveCount(0);
});

test("admin runtime-report failure leaves the authenticated panel in a degraded state", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByTestId("login-username").fill(runtimeErrorUser);
  await page.getByTestId("login-password").fill(adminPassword);
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("auth-role")).toHaveAttribute("data-role", "admin");
  await expect(page.getByTestId("runtime-error")).toContainText(
    "Runtime diagnostics are temporarily unavailable.",
  );
  await expect(page.getByTestId("runtime-metrics-summary")).toHaveCount(0);
});

test("trace lookup shows the no-match state for unknown trace ids", async ({ page }) => {
  await page.goto("/");

  await page.getByTestId("login-username").fill(adminUser);
  await page.getByTestId("login-password").fill(adminPassword);
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
