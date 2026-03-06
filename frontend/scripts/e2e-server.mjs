import { readFile } from "node:fs/promises";
import { createServer } from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const distRoot = path.resolve(__dirname, "..", "dist");
const port = Number(process.env.E2E_PORT ?? 4173);
const apiBaseUrl = `http://127.0.0.1:${port}/mock-api`;
const adminUser = process.env.E2E_ADMIN_USER ?? "admin";
const adminPassword = process.env.E2E_ADMIN_PASSWORD ?? "mock-admin-password";
const developerUser = process.env.E2E_DEVELOPER_USER ?? "developer";
const developerPassword = process.env.E2E_DEVELOPER_PASSWORD ?? "mock-developer-password";
const degradedAdminUser = process.env.E2E_DEGRADED_ADMIN_USER ?? "degraded-admin";
const runtimeErrorUser = process.env.E2E_RUNTIME_ERROR_USER ?? "runtime-error";
const invalidUser = process.env.E2E_INVALID_USER ?? "wrong-user";
const invalidPassword = process.env.E2E_INVALID_PASSWORD ?? "wrong-password";

const runtimeConfigBody = `window.__APP_CONFIG__ = ${JSON.stringify({ apiBaseUrl })};\n`;

const runtimeConfig = {
  profile: "secure",
  http: {
    legacyApiEnabled: false,
    legacyDeprecationDate: "2026-06-01",
    legacySunsetDate: "2026-12-01",
    authRateLimitPerMinute: 30,
    apiRateLimitPerMinute: 120,
    maxInFlightRequests: 64,
    loadShedExemptPrefixes: ["/health"],
    readyCacheTtlMs: 1500,
    readyStaleIfErrorTtlMs: 5000,
  },
  security: {
    strictDependencies: true,
    loggerSignedIngestEnabled: true,
    loggerSharedSecretSet: true,
    localAuthMode: "generated",
    workerTlsMode: "mtls",
    workerServerName: "calculator.internal",
    defaultCredentialsInUse: false,
  },
  operations: {
    autoMigrate: false,
    rateLimitStore: "redis",
    redisFailureMode: "closed",
    loggerEndpointConfigured: true,
    runtimeDiagnosticsCacheTtlMs: 1500,
    incidentEventSink: "logger",
    incidentEventIntervalMs: 10000,
    incidentEventDedupeWindowMs: 120000,
    incidentEventWebhookConfigured: false,
    incidentEventRetentionDays: 30,
    requestLogging: {
      trustedProxyCidrs: ["127.0.0.1/32", "10.0.0.0/8"],
    },
    loggerTiming: {
      healthTimeoutMs: 1500,
      ingestTimestampMaxAgeSeconds: 300,
      ingestTimestampMaxFutureSkewSeconds: 5,
    },
    dependencyPolicies: [
      {
        route: "GET /health/ready",
        dependency: "postgres, redis, worker",
        strictMode: "gateway startup fails if a required dependency cannot initialize",
        nonStrictMode:
          "gateway startup continues and readiness stays degraded until the dependency recovers",
        runtimeBehavior:
          "returns 503 with per-dependency checks while any required dependency is down",
      },
      {
        route: "GET /api/v1/users",
        dependency: "postgres",
        strictMode: "gateway startup fails when postgres init, ping, or migration cannot complete",
        nonStrictMode: "gateway startup continues even when postgres is unavailable",
        runtimeBehavior:
          "returns 200 demo fallback users only when DEMO_FALLBACK_USERS=true; otherwise returns 503 users_unavailable",
      },
      {
        route: "POST /api/v1/compute/fibonacci and /api/v1/compute/hash",
        dependency: "worker",
        strictMode:
          "gateway startup fails when the worker gRPC client cannot initialize or pass health checks",
        nonStrictMode: "gateway startup continues without a worker client",
        runtimeBehavior:
          "returns 503 worker_unavailable until the worker becomes reachable; in-flight RPC failures return 502 worker_call_failed",
      },
      {
        route: "POST /api/v1/auth/token and authenticated /api/v1/* rate limiting",
        dependency: "redis",
        strictMode: "gateway startup fails when redis init or ping cannot complete",
        nonStrictMode: "gateway startup continues without a healthy redis client",
        runtimeBehavior:
          "distributed rate limiting follows RATE_LIMIT_REDIS_FAILURE_MODE=open|closed; open keeps serving traffic, closed returns 503 rate_limiter_unavailable",
      },
      {
        route: "GET /api/v1/admin/request-logs",
        dependency: "logger",
        strictMode: "unchanged; logger is optional at gateway startup",
        nonStrictMode: "unchanged; logger is optional at gateway startup",
        runtimeBehavior:
          "returns 200 with an empty list when LOGGER_ENDPOINT is unset; returns 503 logger_unavailable when logger cannot answer",
      },
      {
        route: "GET /api/v1/admin/runtime-metrics and /api/v1/admin/runtime-report",
        dependency: "logger",
        strictMode: "unchanged; logger is optional at gateway startup",
        nonStrictMode: "unchanged; logger is optional at gateway startup",
        runtimeBehavior:
          "returns 200 while surfacing logger reachability and degraded health warnings inside runtime diagnostics",
      },
    ],
  },
  warnings: [],
};

const degradedRuntimeConfig = {
  ...runtimeConfig,
  security: {
    ...runtimeConfig.security,
    strictDependencies: false,
  },
  operations: {
    ...runtimeConfig.operations,
    redisFailureMode: "open",
    requestLogging: {
      trustedProxyCidrs: [],
    },
    loggerTiming: {
      healthTimeoutMs: 2500,
      ingestTimestampMaxAgeSeconds: 600,
      ingestTimestampMaxFutureSkewSeconds: 15,
    },
  },
  warnings: ["strict dependencies are disabled; dependency-backed routes degrade per endpoint"],
};

const runtimeMetrics = {
  requestsTotal: 128,
  requestErrors: 3,
  errorRate: 0.023,
  latencyCount: 128,
  latencyAverageMs: 42.5,
  loadShedTotal: 0,
  inflightCurrent: 2,
  inflightPeak: 8,
  recentHistory: [
    {
      recordedAt: "2026-03-01T00:00:00Z",
      requestsTotal: 90,
      requestErrors: 1,
      errorRate: 0.011,
      latencyAverageMs: 35.5,
      loadShedTotal: 0,
      inflightCurrent: 1,
      inflightPeak: 4,
    },
    {
      recordedAt: "2026-03-01T00:05:00Z",
      requestsTotal: 128,
      requestErrors: 3,
      errorRate: 0.023,
      latencyAverageMs: 42.5,
      loadShedTotal: 0,
      inflightCurrent: 2,
      inflightPeak: 8,
    },
  ],
  alerts: {
    activeCount: 1,
    highestSeverity: "warning",
    recentlyBreached: false,
    items: [
      {
        code: "gateway-latency",
        severity: "warning",
        status: "open",
        source: "api-gateway",
        message: "Latency trending upward",
        recommendedAction: "Inspect logger queue and worker saturation",
        breachCount: 2,
        lastTriggeredAt: "2026-03-01T00:04:00Z",
      },
    ],
  },
  health: {
    status: "ok",
    httpStatus: 200,
    postgres: "up",
    redis: "up",
    worker: "up",
    cacheState: "fresh",
    cacheAgeMs: 200,
    cacheTtlMs: 1500,
    staleIfErrorTtlMs: 5000,
    lastCheckedAt: "2026-03-01T00:05:00Z",
  },
  trace: {
    enabled: true,
    responseHeader: "X-Trace-Id",
    forwardedToLogger: true,
    storedOnLoggerAs: "X-Trace-Id",
    storageField: "traceId",
  },
  gatewayLogger: {
    enabled: true,
    endpoint: "http://logger:8090",
    queueDepth: 1,
    queueCapacity: 128,
    workers: 2,
    retryMax: 3,
    droppedTotal: 0,
  },
  loggerService: {
    configured: true,
    reachable: true,
    endpointBase: "http://logger:8090",
    healthStatus: "ok",
    queueDepth: 1,
    queueCapacity: 128,
    workers: 2,
    enqueuedTotal: 128,
    droppedTotal: 0,
    processedTotal: 128,
    failedTotal: 0,
    retriedTotal: 1,
    inflightWorkers: 0,
    dropRatio: 0,
    dropAlertThresholdPct: 5,
    dropAlertThresholdHit: false,
    lastError: "",
  },
  incidentJournal: {
    enabled: true,
    sink: "mongo",
    configured: true,
    reachable: true,
    totalEvents: 2,
    activeEvents: 1,
    latestEventAt: "2026-03-01T00:05:00Z",
    lastEventStatus: "stored",
    dispatchFailures: 0,
    lastDispatchAt: "2026-03-01T00:05:00Z",
    lastDispatchError: "",
  },
  warnings: [],
};

const runtimeReport = {
  generatedAt: "2026-03-01T00:05:00Z",
  reportVersion: "2026.03.01",
  config: runtimeConfig,
  metrics: runtimeMetrics,
  runbooks: [
    {
      id: "rb-1",
      title: "Runtime Incident Response",
      path: "docs/runtime-incident-response.md",
      reason: "Primary runtime diagnostics flow",
      priority: "high",
    },
  ],
  incident: {
    recommendedSeverity: "warning",
    category: "latency",
    title: "Gateway latency regression",
    summary: "Trace lookup should show correlated logger records.",
    suspectedSystems: ["api-gateway", "logger"],
    triggeredAlerts: ["gateway-latency"],
    nextActions: ["Check trace correlation", "Inspect latest request logs"],
    evidence: [
      {
        kind: "metric",
        label: "latencyAverageMs",
        value: "42.5",
        source: "runtime-metrics",
      },
    ],
  },
};

const degradedRuntimeReport = {
  ...runtimeReport,
  config: degradedRuntimeConfig,
};

const incidentEvents = {
  items: [
    {
      id: "evt-1",
      eventType: "alert",
      alertCode: "gateway-latency",
      severity: "warning",
      status: "open",
      source: "api-gateway",
      title: "Gateway latency regression",
      summary: "Correlate with trace lookup",
      message: "Latency threshold crossed",
      recommendedAction: "Open trace lookup with the incident trace id",
      recommendedSeverity: "warning",
      triggeredAt: "2026-03-01T00:04:00Z",
      firstSeenAt: "2026-03-01T00:04:00Z",
      lastSeenAt: "2026-03-01T00:05:00Z",
      breachCount: 2,
      traceId: "trace-admin-a",
      reportGeneratedAt: "2026-03-01T00:05:00Z",
      reportVersion: "2026.03.01",
      runbooks: runtimeReport.runbooks,
    },
  ],
};

const requestLogs = [
  {
    path: "/api/v1/admin/request-logs",
    method: "GET",
    ip: "127.0.0.1",
    traceId: "trace-admin-a",
    durationMs: 24,
    statusCode: 200,
    occurredAt: "2026-03-01T00:05:00Z",
  },
  {
    path: "/api/v1/auth/token",
    method: "POST",
    ip: "127.0.0.1",
    traceId: "trace-login-a",
    durationMs: 18,
    statusCode: 200,
    occurredAt: "2026-03-01T00:03:00Z",
  },
];

const mimeTypes = new Map([
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".txt", "text/plain; charset=utf-8"],
  [".css", "text/css; charset=utf-8"],
]);

function sendJson(res, payload, status = 200) {
  res.writeHead(status, { "Content-Type": "application/json; charset=utf-8" });
  res.end(JSON.stringify(payload));
}

function sendText(res, body, contentType = "text/plain; charset=utf-8", status = 200) {
  res.writeHead(status, { "Content-Type": contentType });
  res.end(body);
}

async function readJsonBody(req) {
  const chunks = [];
  for await (const chunk of req) {
    chunks.push(chunk);
  }
  if (chunks.length === 0) {
    return {};
  }

  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    return {};
  }
}

function bearerToken(req) {
  const header = req.headers.authorization ?? "";
  if (!header.startsWith("Bearer ")) {
    return "";
  }
  return header.slice("Bearer ".length);
}

function resolveStaticCandidates(urlPathname) {
  if (urlPathname === "/") {
    return [path.join(distRoot, "index.html")];
  }

  const trimmed = urlPathname.replace(/^\/+/, "");
  if (urlPathname.endsWith("/")) {
    return [path.join(distRoot, trimmed, "index.html")];
  }

  const ext = path.extname(trimmed);
  if (ext) {
    return [path.join(distRoot, trimmed)];
  }

  return [path.join(distRoot, trimmed), path.join(distRoot, trimmed, "index.html")];
}

async function handleStatic(res, pathname) {
  for (const filePath of resolveStaticCandidates(pathname)) {
    try {
      const body = await readFile(filePath);
      const ext = path.extname(filePath);
      sendText(res, body, mimeTypes.get(ext) ?? "application/octet-stream");
      return;
    } catch {
      // Try the next static candidate.
    }
  }

  sendText(res, "not_found", "text/plain; charset=utf-8", 404);
}

const server = createServer(async (req, res) => {
  const url = new URL(req.url ?? "/", `http://127.0.0.1:${port}`);
  const { pathname, searchParams } = url;

  if (pathname === "/runtime-config.js") {
    sendText(res, runtimeConfigBody, "text/javascript; charset=utf-8");
    return;
  }

  if (pathname === "/mock-api/health") {
    sendJson(res, {
      status: "ok",
      checks: { postgres: "up", redis: "up", worker: "up" },
    });
    return;
  }

  if (pathname === "/mock-api/api/v1/auth/token" && req.method === "POST") {
    const body = await readJsonBody(req);
    if (body.password === invalidPassword || body.username === invalidUser) {
      sendJson(
        res,
        {
          error: {
            code: "invalid_credentials",
            message: "invalid username or password",
          },
        },
        401,
      );
      return;
    }

    if (body.username === runtimeErrorUser && body.password === adminPassword) {
      sendJson(res, {
        accessToken: "mock-runtime-error-token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "admin",
      });
      return;
    }

    if (body.username === degradedAdminUser && body.password === adminPassword) {
      sendJson(res, {
        accessToken: "mock-degraded-admin-token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "admin",
      });
      return;
    }

    if (body.username === adminUser && body.password === adminPassword) {
      sendJson(res, {
        accessToken: "mock-admin-token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "admin",
      });
      return;
    }

    if (body.username === developerUser && body.password === developerPassword) {
      sendJson(res, {
        accessToken: "mock-developer-token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "user",
      });
      return;
    }

    sendJson(
      res,
      {
        error: {
          code: "invalid_credentials",
          message: "invalid username or password",
        },
      },
      401,
    );
    return;
  }

  if (pathname === "/mock-api/api/v1/users") {
    sendJson(res, {
      data: [
        {
          id: 1,
          name: "Admin User",
          email: "admin@example.com",
          createdAt: "2026-03-01T00:00:00Z",
        },
      ],
    });
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/runtime-config") {
    if (bearerToken(req) === "mock-runtime-error-token") {
      sendJson(
        res,
        {
          error: { code: "runtime_report_unavailable", message: "runtime diagnostics unavailable" },
        },
        503,
      );
      return;
    }
    if (bearerToken(req) === "mock-degraded-admin-token") {
      sendJson(res, degradedRuntimeConfig);
      return;
    }
    sendJson(res, runtimeConfig);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/runtime-metrics") {
    if (bearerToken(req) === "mock-runtime-error-token") {
      sendJson(
        res,
        {
          error: { code: "runtime_report_unavailable", message: "runtime diagnostics unavailable" },
        },
        503,
      );
      return;
    }
    sendJson(res, runtimeMetrics);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/runtime-report") {
    if (bearerToken(req) === "mock-runtime-error-token") {
      sendJson(
        res,
        {
          error: { code: "runtime_report_unavailable", message: "runtime diagnostics unavailable" },
        },
        503,
      );
      return;
    }
    if (bearerToken(req) === "mock-degraded-admin-token") {
      sendJson(res, degradedRuntimeReport);
      return;
    }
    sendJson(res, runtimeReport);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/incident-events") {
    if (bearerToken(req) === "mock-runtime-error-token") {
      sendJson(
        res,
        {
          error: { code: "runtime_report_unavailable", message: "runtime diagnostics unavailable" },
        },
        503,
      );
      return;
    }
    sendJson(res, incidentEvents);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/request-logs") {
    const traceId = searchParams.get("traceId")?.trim() ?? "";
    if (traceId === "trace-error") {
      sendJson(
        res,
        { error: { code: "request_logs_unavailable", message: "request log lookup failed" } },
        503,
      );
      return;
    }
    const items = traceId ? requestLogs.filter((entry) => entry.traceId === traceId) : requestLogs;
    sendJson(res, { items });
    return;
  }

  await handleStatic(res, pathname);
});

server.listen(port, "127.0.0.1", () => {
  process.stdout.write(`frontend e2e server listening on http://127.0.0.1:${port}\n`);
});
