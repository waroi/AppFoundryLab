import { describe, expect, it } from "vitest";
import {
  isFibonacciResponse,
  isHealthResponse,
  isRuntimeConfigResponse,
  isRuntimeIncidentEventsResponse,
  isRuntimeMetricsResponse,
  isRuntimeReportResponse,
  isRuntimeRequestLogsResponse,
  isTokenResponse,
  isUsersResponse,
} from "../validators";

describe("validators", () => {
  describe("isHealthResponse", () => {
    it('should return true for valid "ok" payload', () => {
      const payload = {
        status: "ok",
        checks: {
          postgres: "up",
          redis: "up",
          worker: "up",
        },
      };
      expect(isHealthResponse(payload)).toBe(true);
    });

    it('should return true for valid "degraded" payload with components down', () => {
      const payload = {
        status: "degraded",
        checks: {
          postgres: "down",
          redis: "up",
          worker: "down",
        },
      };
      expect(isHealthResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object payloads", () => {
      expect(isHealthResponse(null)).toBe(false);
      expect(isHealthResponse("ok")).toBe(false);
      expect(isHealthResponse(123)).toBe(false);
    });

    it("should return false if missing fields", () => {
      const payload = {
        status: "ok",
        // missing checks
      };
      expect(isHealthResponse(payload)).toBe(false);
    });

    it("should return false for invalid nested fields", () => {
      const payload = {
        status: "ok",
        checks: {
          postgres: "invalid", // invalid value
          redis: "up",
          worker: "up",
        },
      };
      expect(isHealthResponse(payload)).toBe(false);
    });
  });

  describe("isUsersResponse", () => {
    it("should return true for a valid users list payload", () => {
      const payload = {
        data: [
          {
            id: 1,
            name: "Alice",
            email: "alice@example.com",
            createdAt: "2023-01-01T00:00:00Z",
          },
          {
            id: 2,
            name: "Bob",
            email: "bob@example.com",
            createdAt: "2023-01-02T00:00:00Z",
          },
        ],
      };
      expect(isUsersResponse(payload)).toBe(true);
    });

    it("should return true for an empty valid users list", () => {
      const payload = {
        data: [],
      };
      expect(isUsersResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isUsersResponse(null)).toBe(false);
      expect(isUsersResponse([])).toBe(false); // Validating object containing 'data' array
    });

    it("should return false if missing data array", () => {
      const payload = {};
      expect(isUsersResponse(payload)).toBe(false);
    });

    it("should return false if a user object has invalid types", () => {
      const payload = {
        data: [
          {
            id: "1", // Should be a number
            name: "Alice",
            email: "alice@example.com",
            createdAt: "2023-01-01T00:00:00Z",
          },
        ],
      };
      expect(isUsersResponse(payload)).toBe(false);
    });

    it("should return false if data contains non-objects", () => {
      const payload = {
        data: ["user1"],
      };
      expect(isUsersResponse(payload)).toBe(false);
    });
  });

  describe("isTokenResponse", () => {
    it("should return true for a valid user token payload", () => {
      const payload = {
        accessToken: "some-jwt-token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "user",
      };
      expect(isTokenResponse(payload)).toBe(true);
    });

    it("should return true for a valid admin token payload", () => {
      const payload = {
        accessToken: "admin-jwt-token",
        tokenType: "Bearer",
        expiresIn: 7200,
        role: "admin",
      };
      expect(isTokenResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isTokenResponse(null)).toBe(false);
      expect(isTokenResponse("token")).toBe(false);
    });

    it("should return false if missing required fields", () => {
      const payload = {
        accessToken: "token",
        tokenType: "Bearer",
        // missing expiresIn and role
      };
      expect(isTokenResponse(payload)).toBe(false);
    });

    it("should return false if wrong tokenType", () => {
      const payload = {
        accessToken: "token",
        tokenType: "Basic", // should be 'Bearer'
        expiresIn: 3600,
        role: "user",
      };
      expect(isTokenResponse(payload)).toBe(false);
    });

    it("should return false if wrong role", () => {
      const payload = {
        accessToken: "token",
        tokenType: "Bearer",
        expiresIn: 3600,
        role: "superadmin", // should be 'user' or 'admin'
      };
      expect(isTokenResponse(payload)).toBe(false);
    });
  });

  describe("isFibonacciResponse", () => {
    it("should return true for a valid fibonacci payload", () => {
      const payload = {
        n: 10,
        value: 55,
      };
      expect(isFibonacciResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isFibonacciResponse(null)).toBe(false);
      expect(isFibonacciResponse(55)).toBe(false);
    });

    it("should return false if missing fields", () => {
      const payload = {
        n: 10,
      };
      expect(isFibonacciResponse(payload)).toBe(false);
    });

    it("should return false if types are incorrect", () => {
      const payload = {
        n: "10", // should be number
        value: 55,
      };
      expect(isFibonacciResponse(payload)).toBe(false);
    });
  });

  describe("isRuntimeConfigResponse", () => {
    const validConfig = {
      profile: "test",
      http: {
        legacyApiEnabled: false,
        legacyDeprecationDate: "2024-01-01",
        legacySunsetDate: "2024-06-01",
        authRateLimitPerMinute: 10,
        apiRateLimitPerMinute: 100,
        maxInFlightRequests: 50,
        loadShedExemptPrefixes: ["/admin"],
        readyCacheTtlMs: 5000,
        readyStaleIfErrorTtlMs: 10000,
      },
      security: {
        strictDependencies: true,
        loggerSignedIngestEnabled: true,
        loggerSharedSecretSet: true,
        localAuthMode: "strict",
        workerTlsMode: "enforced",
        workerServerName: "worker",
        defaultCredentialsInUse: false,
      },
      operations: {
        autoMigrate: true,
        rateLimitStore: "redis",
        redisFailureMode: "reject",
        runtimeDiagnosticsCacheTtlMs: 60000,
        incidentEventSink: "webhook",
        incidentEventIntervalMs: 5000,
        incidentEventDedupeWindowMs: 60000,
        incidentEventWebhookConfigured: true,
        incidentEventRetentionDays: 30,
        loggerEndpointConfigured: true,
        requestLogging: {
          trustedProxyCidrs: ["127.0.0.1/32"],
        },
        loggerTiming: {
          healthTimeoutMs: 1500,
          ingestTimestampMaxAgeSeconds: 300,
          ingestTimestampMaxFutureSkewSeconds: 5,
        },
        dependencyPolicies: [
          {
            route: "GET /api/v1/users",
            dependency: "postgres",
            strictMode: "startup fails",
            nonStrictMode: "startup continues",
            runtimeBehavior: "returns 503 users_unavailable",
          },
        ],
      },
      warnings: ["warning1"],
    };

    it("should return true for a valid runtime config payload", () => {
      expect(isRuntimeConfigResponse(validConfig)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isRuntimeConfigResponse(null)).toBe(false);
      expect(isRuntimeConfigResponse("config")).toBe(false);
    });

    it("should return false if missing top-level fields", () => {
      const { http, ...missingHttp } = validConfig;
      expect(isRuntimeConfigResponse(missingHttp)).toBe(false);
    });

    it("should return false if nested object is missing fields", () => {
      const invalidHttp = { ...validConfig, http: { legacyApiEnabled: true } };
      expect(isRuntimeConfigResponse(invalidHttp)).toBe(false);
    });

    it("should return false if warnings is not an array", () => {
      const invalidConfig = { ...validConfig, warnings: "warning1" };
      expect(isRuntimeConfigResponse(invalidConfig)).toBe(false);
    });

    it("should return false if load shed prefixes contain non-string values", () => {
      const invalidConfig = {
        ...validConfig,
        http: {
          ...validConfig.http,
          loadShedExemptPrefixes: ["/health", 42],
        },
      };

      expect(isRuntimeConfigResponse(invalidConfig)).toBe(false);
    });

    it("should return false if dependency policies are missing required fields", () => {
      const invalidConfig = {
        ...validConfig,
        operations: {
          ...validConfig.operations,
          dependencyPolicies: [{ route: "GET /api/v1/users" }],
        },
      };

      expect(isRuntimeConfigResponse(invalidConfig)).toBe(false);
    });

    it("should return false if logger timing fields are missing", () => {
      const invalidConfig = {
        ...validConfig,
        operations: {
          ...validConfig.operations,
          loggerTiming: { healthTimeoutMs: 1500 },
        },
      };

      expect(isRuntimeConfigResponse(invalidConfig)).toBe(false);
    });
  });

  describe("isRuntimeMetricsResponse", () => {
    const validMetrics = {
      requestsTotal: 100,
      requestErrors: 5,
      errorRate: 0.05,
      latencyCount: 100,
      latencyAverageMs: 50,
      loadShedTotal: 0,
      inflightCurrent: 10,
      inflightPeak: 20,
      recentHistory: [
        {
          recordedAt: "2023-10-10T00:00:00Z",
          requestsTotal: 10,
          requestErrors: 0,
          errorRate: 0,
          latencyAverageMs: 40,
          loadShedTotal: 0,
          inflightCurrent: 5,
          inflightPeak: 10,
        },
      ],
      alerts: {
        activeCount: 1,
        highestSeverity: "high",
        recentlyBreached: true,
        items: [
          {
            code: "ERR1",
            severity: "high",
            status: "active",
            source: "gateway",
            message: "Error",
            recommendedAction: "Fix it",
            breachCount: 5,
            lastTriggeredAt: "2023-10-10T00:00:00Z",
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
        cacheAgeMs: 100,
        cacheTtlMs: 60000,
        staleIfErrorTtlMs: 120000,
        lastCheckedAt: "2023-10-10T00:00:00Z",
      },
      trace: {
        enabled: true,
        responseHeader: "x-trace-id",
        forwardedToLogger: true,
        storedOnLoggerAs: "trace_id",
        storageField: "trace",
      },
      gatewayLogger: {
        enabled: true,
        endpoint: "http://logger",
        queueDepth: 0,
        queueCapacity: 1000,
        workers: 5,
        retryMax: 3,
        droppedTotal: 0,
      },
      loggerService: {
        configured: true,
        reachable: true,
        endpointBase: "http://logger",
        healthStatus: "ok",
        queueDepth: 0,
        queueCapacity: 10000,
        workers: 10,
        enqueuedTotal: 100,
        droppedTotal: 0,
        processedTotal: 100,
        failedTotal: 0,
        retriedTotal: 0,
        inflightWorkers: 2,
        dropRatio: 0,
        dropAlertThresholdPct: 0.1,
        dropAlertThresholdHit: false,
        lastError: "",
      },
      incidentJournal: {
        enabled: true,
        sink: "webhook",
        configured: true,
        reachable: true,
        totalEvents: 10,
        activeEvents: 1,
        latestEventAt: "2023-10-10T00:00:00Z",
        lastEventStatus: "dispatched",
        dispatchFailures: 0,
        lastDispatchAt: "2023-10-10T00:00:00Z",
        lastDispatchError: "",
      },
      warnings: [],
    };

    it("should return true for a valid runtime metrics payload", () => {
      expect(isRuntimeMetricsResponse(validMetrics)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isRuntimeMetricsResponse(null)).toBe(false);
      expect(isRuntimeMetricsResponse("metrics")).toBe(false);
    });

    it("should return false if missing top-level metrics", () => {
      const { requestsTotal, ...missingFields } = validMetrics;
      expect(isRuntimeMetricsResponse(missingFields)).toBe(false);
    });

    it("should return false if recentHistory item is invalid", () => {
      const invalidHistory = {
        ...validMetrics,
        recentHistory: [{ invalid: "point" }],
      };
      expect(isRuntimeMetricsResponse(invalidHistory)).toBe(false);
    });

    it("should return false if alert item is invalid", () => {
      const invalidAlerts = {
        ...validMetrics,
        alerts: { ...validMetrics.alerts, items: [{ missingFields: true }] },
      };
      expect(isRuntimeMetricsResponse(invalidAlerts)).toBe(false);
    });

    it("should return false if missing a major nested block (e.g. loggerService)", () => {
      const { loggerService, ...missingLogger } = validMetrics;
      expect(isRuntimeMetricsResponse(missingLogger)).toBe(false);
    });
  });

  describe("isRuntimeReportResponse", () => {
    const validConfig = {
      profile: "test",
      http: {
        legacyApiEnabled: false,
        legacyDeprecationDate: "2024-01-01",
        legacySunsetDate: "2024-06-01",
        authRateLimitPerMinute: 10,
        apiRateLimitPerMinute: 100,
        maxInFlightRequests: 50,
        loadShedExemptPrefixes: ["/admin"],
        readyCacheTtlMs: 5000,
        readyStaleIfErrorTtlMs: 10000,
      },
      security: {
        strictDependencies: true,
        tlsEnabled: true,
        tlsMinVersion: "1.2",
        authJwtIssuer: "issuer",
        authJwtAudience: "audience",
        loggerSignedIngestEnabled: true,
        loggerSharedSecretSet: true,
        localAuthMode: "strict",
        workerTlsMode: "enforced",
        workerServerName: "worker",
        defaultCredentialsInUse: false,
      },
      operations: {
        autoMigrate: true,
        rateLimitStore: "redis",
        redisFailureMode: "reject",
        runtimeDiagnosticsCacheTtlMs: 60000,
        incidentEventSink: "webhook",
        incidentEventIntervalMs: 5000,
        incidentEventDedupeWindowMs: 60000,
        incidentEventWebhookConfigured: true,
        incidentEventRetentionDays: 30,
        loggerEndpointConfigured: true,
        requestLogging: {
          trustedProxyCidrs: ["127.0.0.1/32"],
        },
        loggerTiming: {
          healthTimeoutMs: 1500,
          ingestTimestampMaxAgeSeconds: 300,
          ingestTimestampMaxFutureSkewSeconds: 5,
        },
        dependencyPolicies: [
          {
            route: "GET /api/v1/users",
            dependency: "postgres",
            strictMode: "startup fails",
            nonStrictMode: "startup continues",
            runtimeBehavior: "returns 503 users_unavailable",
          },
        ],
      },
      warnings: ["warning1"],
    };

    const validMetrics = {
      requestsTotal: 100,
      requestErrors: 5,
      errorRate: 0.05,
      latencyCount: 100,
      latencyAverageMs: 50,
      loadShedTotal: 0,
      inflightCurrent: 10,
      inflightPeak: 20,
      recentHistory: [],
      alerts: {
        activeCount: 1,
        highestSeverity: "high",
        recentlyBreached: true,
        items: [],
      },
      health: {
        status: "ok",
        httpStatus: 200,
        postgres: "up",
        redis: "up",
        worker: "up",
        cacheState: "fresh",
        cacheAgeMs: 100,
        cacheTtlMs: 60000,
        staleIfErrorTtlMs: 120000,
        lastCheckedAt: "2023-10-10T00:00:00Z",
      },
      trace: {
        enabled: true,
        responseHeader: "x-trace-id",
        forwardedToLogger: true,
        storedOnLoggerAs: "trace_id",
        storageField: "trace",
      },
      gatewayLogger: {
        enabled: true,
        endpoint: "http://logger",
        queueDepth: 0,
        queueCapacity: 1000,
        workers: 5,
        retryMax: 3,
        droppedTotal: 0,
      },
      loggerService: {
        configured: true,
        reachable: true,
        endpointBase: "http://logger",
        healthStatus: "ok",
        queueDepth: 0,
        queueCapacity: 10000,
        workers: 10,
        enqueuedTotal: 100,
        droppedTotal: 0,
        processedTotal: 100,
        failedTotal: 0,
        retriedTotal: 0,
        inflightWorkers: 2,
        dropRatio: 0,
        dropAlertThresholdPct: 0.1,
        dropAlertThresholdHit: false,
        lastError: "",
      },
      incidentJournal: {
        enabled: true,
        sink: "webhook",
        configured: true,
        reachable: true,
        totalEvents: 10,
        activeEvents: 1,
        latestEventAt: "2023-10-10T00:00:00Z",
        lastEventStatus: "dispatched",
        dispatchFailures: 0,
        lastDispatchAt: "2023-10-10T00:00:00Z",
        lastDispatchError: "",
      },
      warnings: [],
    };

    const validReport = {
      generatedAt: "2023-10-10T00:00:00Z",
      reportVersion: "1.0",
      config: validConfig,
      metrics: validMetrics,
      runbooks: [
        {
          id: "rb1",
          title: "Runbook 1",
          path: "/runbook",
          reason: "High errors",
          priority: "high",
        },
      ],
      incident: {
        recommendedSeverity: "warning",
        category: "latency",
        title: "Gateway latency regression",
        summary: "Logger requests are slower than expected.",
        suspectedSystems: ["gateway"],
        triggeredAlerts: ["ERR1"],
        nextActions: ["Investigate"],
        evidence: [
          {
            kind: "metric",
            label: "latencyAverageMs",
            value: "50",
            source: "runtime-metrics",
          },
        ],
      },
    };

    it("should return true for a valid runtime report payload", () => {
      expect(isRuntimeReportResponse(validReport)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isRuntimeReportResponse(null)).toBe(false);
      expect(isRuntimeReportResponse("report")).toBe(false);
    });

    it("should return false if missing generatedAt", () => {
      const { generatedAt, ...missingFields } = validReport;
      expect(isRuntimeReportResponse(missingFields)).toBe(false);
    });

    it("should return false if runbook item is invalid", () => {
      const invalidRunbooks = {
        ...validReport,
        runbooks: [{ id: "rb1" }], // missing other required runbook fields
      };
      expect(isRuntimeReportResponse(invalidRunbooks)).toBe(false);
    });

    it("should return false if incident is invalid", () => {
      const invalidIncident = {
        ...validReport,
        incident: { suspectedSystems: "gateway" }, // suspectedSystems should be array
      };
      expect(isRuntimeReportResponse(invalidIncident)).toBe(false);
    });

    it("should return false if incident evidence entries are malformed", () => {
      const invalidReport = {
        ...validReport,
        incident: {
          ...validReport.incident,
          evidence: [{ label: "latencyAverageMs" }],
        },
      };

      expect(isRuntimeReportResponse(invalidReport)).toBe(false);
    });
  });

  describe("isRuntimeIncidentEventsResponse", () => {
    it("should return true for a valid incident events payload", () => {
      const payload = {
        items: [
          {
            id: "ev1",
            eventType: "alert",
            alertCode: "ERR1",
            severity: "high",
            status: "active",
            source: "gateway",
            title: "Error Title",
            summary: "Error Summary",
            message: "Error Message",
            recommendedAction: "Fix",
            recommendedSeverity: "high",
            triggeredAt: "2023-10-10T00:00:00Z",
            firstSeenAt: "2023-10-10T00:00:00Z",
            lastSeenAt: "2023-10-10T00:00:00Z",
            breachCount: 5,
            traceId: "trace1",
            reportGeneratedAt: "2023-10-10T00:00:00Z",
            reportVersion: "1.0",
            runbooks: [
              {
                id: "rb1",
                title: "Runtime Incident Response",
                path: "docs/runtime-incident-response.md",
                reason: "Primary runtime flow",
                priority: "high",
              },
            ],
          },
        ],
      };
      expect(isRuntimeIncidentEventsResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isRuntimeIncidentEventsResponse(null)).toBe(false);
      expect(isRuntimeIncidentEventsResponse("events")).toBe(false);
    });

    it("should return false if items is missing", () => {
      const payload = {};
      expect(isRuntimeIncidentEventsResponse(payload)).toBe(false);
    });

    it("should return false if an item is invalid", () => {
      const payload = {
        items: [
          {
            id: "ev1",
            // missing other fields
          },
        ],
      };
      expect(isRuntimeIncidentEventsResponse(payload)).toBe(false);
    });

    it("should return false if incident runbooks are malformed", () => {
      const payload = {
        items: [
          {
            id: "ev1",
            eventType: "alert",
            alertCode: "ERR1",
            severity: "high",
            status: "active",
            source: "gateway",
            title: "Error Title",
            summary: "Error Summary",
            message: "Error Message",
            recommendedAction: "Fix",
            recommendedSeverity: "high",
            triggeredAt: "2023-10-10T00:00:00Z",
            firstSeenAt: "2023-10-10T00:00:00Z",
            lastSeenAt: "2023-10-10T00:00:00Z",
            breachCount: 5,
            traceId: "trace1",
            reportGeneratedAt: "2023-10-10T00:00:00Z",
            reportVersion: "1.0",
            runbooks: [{ id: "rb1" }],
          },
        ],
      };

      expect(isRuntimeIncidentEventsResponse(payload)).toBe(false);
    });
  });

  describe("isRuntimeRequestLogsResponse", () => {
    it("should return true for a valid request logs payload", () => {
      const payload = {
        items: [
          {
            path: "/api",
            method: "GET",
            ip: "127.0.0.1",
            traceId: "trace1",
            durationMs: 10,
            statusCode: 200,
            occurredAt: "2023-10-10T00:00:00Z",
          },
          {
            path: "/api2",
            method: "POST",
            ip: "127.0.0.1",
            // missing traceId is valid
            durationMs: 20,
            statusCode: 201,
            occurredAt: "2023-10-10T00:00:00Z",
          },
        ],
      };
      expect(isRuntimeRequestLogsResponse(payload)).toBe(true);
    });

    it("should return false for null or non-object", () => {
      expect(isRuntimeRequestLogsResponse(null)).toBe(false);
      expect(isRuntimeRequestLogsResponse("logs")).toBe(false);
    });

    it("should return false if items is missing", () => {
      const payload = {};
      expect(isRuntimeRequestLogsResponse(payload)).toBe(false);
    });

    it("should return false if an item is invalid", () => {
      const payload = {
        items: [
          {
            path: "/api",
            // missing other fields
          },
        ],
      };
      expect(isRuntimeRequestLogsResponse(payload)).toBe(false);
    });
  });
});
