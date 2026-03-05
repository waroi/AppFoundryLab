import { readFile } from "node:fs/promises";

async function assertFileContains(path, expected) {
  const body = await readFile(path, "utf8");
  if (!body.includes(expected)) {
    throw new Error(`${path} does not include expected text: ${expected}`);
  }
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function isObject(value) {
  return typeof value === "object" && value !== null;
}

async function fetchJson(url, init) {
  const res = await fetch(url, init);
  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = null;
  }
  return { res, data, text };
}

async function runApiContractSmoke() {
  const base = process.env.SMOKE_API_BASE_URL;
  if (!base) {
    return;
  }

  const username = process.env.SMOKE_API_USERNAME ?? "developer";
  const password = process.env.SMOKE_API_PASSWORD ?? "developer_dev_password";

  const health = await fetchJson(`${base}/health`);
  assert(health.res.ok, `health request failed: ${health.res.status}`);
  assert(isObject(health.data), "health response must be an object");
  assert(
    health.data.status === "ok" || health.data.status === "degraded",
    "health.status must be ok|degraded",
  );

  const token = await fetchJson(`${base}/api/v1/auth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  assert(token.res.ok, `token request failed: ${token.res.status}`);
  assert(isObject(token.data), "token response must be an object");
  assert(
    typeof token.data.accessToken === "string" && token.data.accessToken.length > 0,
    "token.accessToken must be non-empty",
  );
  assert(token.data.tokenType === "Bearer", "token.tokenType must be Bearer");

  const users = await fetchJson(`${base}/api/v1/users`, {
    headers: { Authorization: `Bearer ${token.data.accessToken}` },
  });
  assert(users.res.ok, `users request failed: ${users.res.status}`);
  assert(isObject(users.data), "users response must be an object");
  assert(Array.isArray(users.data.data), "users.data must be an array");

  console.info(`frontend api contract smoke passed (${base})`);
}

await assertFileContains("dist/index.html", 'data-testid="home-hero"');
await assertFileContains("dist/index.html", 'data-testid="preference-toolbar"');
await assertFileContains("dist/test/index.html", 'data-testid="test-page-intro"');
await assertFileContains("dist/test/index.html", 'data-testid="restore-drill-preview"');
await assertFileContains("dist/tr/index.html", 'data-testid="home-hero"');
await assertFileContains("dist/tr/index.html", 'data-testid="preference-toolbar"');
await assertFileContains("dist/tr/test/index.html", 'data-testid="test-page-intro"');
await assertFileContains("dist/tr/test/index.html", 'data-testid="restore-drill-preview"');
await runApiContractSmoke();

console.info("frontend smoke passed");
