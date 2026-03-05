import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const distRoot = path.resolve(__dirname, "..", "dist");
const port = Number(process.env.E2E_PORT ?? 4173);
const apiBaseUrl = `http://127.0.0.1:${port}/mock-api`;

const runtimeConfigBody = `window.__APP_CONFIG__ = ${JSON.stringify({ apiBaseUrl })};\n`;

const runtimeConfig = {
  profile: "single-host",
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
    redisFailureMode: "fail-open",
    loggerEndpointConfigured: true,
    runtimeDiagnosticsCacheTtlMs: 1500,
    incidentEventSink: "mongo",
    incidentEventIntervalMs: 10000,
    incidentEventDedupeWindowMs: 120000,
    incidentEventWebhookConfigured: false,
    incidentEventRetentionDays: 30,
  },
  warnings: [],
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

  return [
    path.join(distRoot, trimmed),
    path.join(distRoot, trimmed, "index.html"),
  ];
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
    sendJson(res, {
      accessToken: "mock-admin-token",
      tokenType: "Bearer",
      expiresIn: 3600,
      role: "admin",
    });
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
    sendJson(res, runtimeConfig);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/runtime-metrics") {
    sendJson(res, runtimeMetrics);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/runtime-report") {
    sendJson(res, runtimeReport);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/incident-events") {
    sendJson(res, incidentEvents);
    return;
  }

  if (pathname === "/mock-api/api/v1/admin/request-logs") {
    const traceId = searchParams.get("traceId")?.trim() ?? "";
    const items = traceId
      ? requestLogs.filter((entry) => entry.traceId === traceId)
      : requestLogs;
    sendJson(res, { items });
    return;
  }

  await handleStatic(res, pathname);
});

server.listen(port, "127.0.0.1", () => {
  console.log(`frontend e2e server listening on http://127.0.0.1:${port}`);
});
