import type {
  FibonacciResponse,
  HealthResponse,
  RuntimeConfigResponse,
  RuntimeIncidentEventsResponse,
  RuntimeMetricsResponse,
  RuntimeReportResponse,
  RuntimeRequestLogsResponse,
  TokenResponse,
  UsersResponse,
} from "./types";

export function isHealthResponse(payload: unknown): payload is HealthResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  const checks = candidate.checks as Record<string, unknown> | undefined;

  return (
    (candidate.status === "ok" || candidate.status === "degraded") &&
    !!checks &&
    (checks.postgres === "up" || checks.postgres === "down") &&
    (checks.redis === "up" || checks.redis === "down") &&
    (checks.worker === "up" || checks.worker === "down")
  );
}

export function isUsersResponse(payload: unknown): payload is UsersResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  if (!Array.isArray(candidate.data)) {
    return false;
  }

  return candidate.data.every((item) => {
    if (typeof item !== "object" || item === null) {
      return false;
    }
    const user = item as Record<string, unknown>;
    return (
      typeof user.id === "number" &&
      typeof user.name === "string" &&
      typeof user.email === "string" &&
      typeof user.createdAt === "string"
    );
  });
}

export function isTokenResponse(payload: unknown): payload is TokenResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }
  const candidate = payload as Record<string, unknown>;
  return (
    typeof candidate.accessToken === "string" &&
    candidate.tokenType === "Bearer" &&
    typeof candidate.expiresIn === "number" &&
    (candidate.role === "admin" || candidate.role === "user")
  );
}

export function isFibonacciResponse(payload: unknown): payload is FibonacciResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }
  const candidate = payload as Record<string, unknown>;
  return typeof candidate.n === "number" && typeof candidate.value === "number";
}

export function isRuntimeConfigResponse(payload: unknown): payload is RuntimeConfigResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  const http = candidate.http as Record<string, unknown> | undefined;
  const security = candidate.security as Record<string, unknown> | undefined;
  const operations = candidate.operations as Record<string, unknown> | undefined;

  return (
    typeof candidate.profile === "string" &&
    !!http &&
    typeof http.legacyApiEnabled === "boolean" &&
    typeof http.legacyDeprecationDate === "string" &&
    typeof http.legacySunsetDate === "string" &&
    typeof http.authRateLimitPerMinute === "number" &&
    typeof http.apiRateLimitPerMinute === "number" &&
    typeof http.maxInFlightRequests === "number" &&
    Array.isArray(http.loadShedExemptPrefixes) &&
    typeof http.readyCacheTtlMs === "number" &&
    typeof http.readyStaleIfErrorTtlMs === "number" &&
    !!security &&
    typeof security.strictDependencies === "boolean" &&
    typeof security.loggerSignedIngestEnabled === "boolean" &&
    typeof security.loggerSharedSecretSet === "boolean" &&
    typeof security.localAuthMode === "string" &&
    typeof security.workerTlsMode === "string" &&
    typeof security.workerServerName === "string" &&
    typeof security.defaultCredentialsInUse === "boolean" &&
    !!operations &&
    typeof operations.autoMigrate === "boolean" &&
    typeof operations.rateLimitStore === "string" &&
    typeof operations.redisFailureMode === "string" &&
    typeof operations.runtimeDiagnosticsCacheTtlMs === "number" &&
    typeof operations.incidentEventSink === "string" &&
    typeof operations.incidentEventIntervalMs === "number" &&
    typeof operations.incidentEventDedupeWindowMs === "number" &&
    typeof operations.incidentEventWebhookConfigured === "boolean" &&
    typeof operations.incidentEventRetentionDays === "number" &&
    typeof operations.loggerEndpointConfigured === "boolean" &&
    Array.isArray(candidate.warnings)
  );
}

export function isRuntimeMetricsResponse(payload: unknown): payload is RuntimeMetricsResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  const health = candidate.health as Record<string, unknown> | undefined;
  const trace = candidate.trace as Record<string, unknown> | undefined;
  const gatewayLogger = candidate.gatewayLogger as Record<string, unknown> | undefined;
  const loggerService = candidate.loggerService as Record<string, unknown> | undefined;
  const alerts = candidate.alerts as Record<string, unknown> | undefined;
  const incidentJournal = candidate.incidentJournal as Record<string, unknown> | undefined;
  return (
    typeof candidate.requestsTotal === "number" &&
    typeof candidate.requestErrors === "number" &&
    typeof candidate.errorRate === "number" &&
    typeof candidate.latencyCount === "number" &&
    typeof candidate.latencyAverageMs === "number" &&
    typeof candidate.loadShedTotal === "number" &&
    typeof candidate.inflightCurrent === "number" &&
    typeof candidate.inflightPeak === "number" &&
    Array.isArray(candidate.recentHistory) &&
    candidate.recentHistory.every((item) => {
      if (typeof item !== "object" || item === null) {
        return false;
      }
      const point = item as Record<string, unknown>;
      return (
        typeof point.recordedAt === "string" &&
        typeof point.requestsTotal === "number" &&
        typeof point.requestErrors === "number" &&
        typeof point.errorRate === "number" &&
        typeof point.latencyAverageMs === "number" &&
        typeof point.loadShedTotal === "number" &&
        typeof point.inflightCurrent === "number" &&
        typeof point.inflightPeak === "number"
      );
    }) &&
    !!alerts &&
    typeof alerts.activeCount === "number" &&
    typeof alerts.highestSeverity === "string" &&
    typeof alerts.recentlyBreached === "boolean" &&
    Array.isArray(alerts.items) &&
    alerts.items.every((item) => {
      if (typeof item !== "object" || item === null) {
        return false;
      }
      const alert = item as Record<string, unknown>;
      return (
        typeof alert.code === "string" &&
        typeof alert.severity === "string" &&
        typeof alert.status === "string" &&
        typeof alert.source === "string" &&
        typeof alert.message === "string" &&
        typeof alert.recommendedAction === "string" &&
        typeof alert.breachCount === "number" &&
        typeof alert.lastTriggeredAt === "string"
      );
    }) &&
    !!health &&
    typeof health.status === "string" &&
    typeof health.httpStatus === "number" &&
    typeof health.postgres === "string" &&
    typeof health.redis === "string" &&
    typeof health.worker === "string" &&
    typeof health.cacheState === "string" &&
    typeof health.cacheAgeMs === "number" &&
    typeof health.cacheTtlMs === "number" &&
    typeof health.staleIfErrorTtlMs === "number" &&
    typeof health.lastCheckedAt === "string" &&
    !!trace &&
    typeof trace.enabled === "boolean" &&
    typeof trace.responseHeader === "string" &&
    typeof trace.forwardedToLogger === "boolean" &&
    typeof trace.storedOnLoggerAs === "string" &&
    typeof trace.storageField === "string" &&
    !!gatewayLogger &&
    typeof gatewayLogger.enabled === "boolean" &&
    typeof gatewayLogger.endpoint === "string" &&
    typeof gatewayLogger.queueDepth === "number" &&
    typeof gatewayLogger.queueCapacity === "number" &&
    typeof gatewayLogger.workers === "number" &&
    typeof gatewayLogger.retryMax === "number" &&
    typeof gatewayLogger.droppedTotal === "number" &&
    !!loggerService &&
    typeof loggerService.configured === "boolean" &&
    typeof loggerService.reachable === "boolean" &&
    typeof loggerService.endpointBase === "string" &&
    typeof loggerService.healthStatus === "string" &&
    typeof loggerService.queueDepth === "number" &&
    typeof loggerService.queueCapacity === "number" &&
    typeof loggerService.workers === "number" &&
    typeof loggerService.enqueuedTotal === "number" &&
    typeof loggerService.droppedTotal === "number" &&
    typeof loggerService.processedTotal === "number" &&
    typeof loggerService.failedTotal === "number" &&
    typeof loggerService.retriedTotal === "number" &&
    typeof loggerService.inflightWorkers === "number" &&
    typeof loggerService.dropRatio === "number" &&
    typeof loggerService.dropAlertThresholdPct === "number" &&
    typeof loggerService.dropAlertThresholdHit === "boolean" &&
    typeof loggerService.lastError === "string" &&
    !!incidentJournal &&
    typeof incidentJournal.enabled === "boolean" &&
    typeof incidentJournal.sink === "string" &&
    typeof incidentJournal.configured === "boolean" &&
    typeof incidentJournal.reachable === "boolean" &&
    typeof incidentJournal.totalEvents === "number" &&
    typeof incidentJournal.activeEvents === "number" &&
    typeof incidentJournal.latestEventAt === "string" &&
    typeof incidentJournal.lastEventStatus === "string" &&
    typeof incidentJournal.dispatchFailures === "number" &&
    typeof incidentJournal.lastDispatchAt === "string" &&
    typeof incidentJournal.lastDispatchError === "string" &&
    Array.isArray(candidate.warnings)
  );
}

export function isRuntimeReportResponse(payload: unknown): payload is RuntimeReportResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  return (
    typeof candidate.generatedAt === "string" &&
    typeof candidate.reportVersion === "string" &&
    isRuntimeConfigResponse(candidate.config) &&
    isRuntimeMetricsResponse(candidate.metrics) &&
    Array.isArray(candidate.runbooks) &&
    candidate.runbooks.every((item) => {
      if (typeof item !== "object" || item === null) {
        return false;
      }
      const runbook = item as Record<string, unknown>;
      return (
        typeof runbook.id === "string" &&
        typeof runbook.title === "string" &&
        typeof runbook.path === "string" &&
        typeof runbook.reason === "string" &&
        typeof runbook.priority === "string"
      );
    }) &&
    typeof candidate.incident === "object" &&
    candidate.incident !== null &&
    Array.isArray((candidate.incident as Record<string, unknown>).suspectedSystems) &&
    Array.isArray((candidate.incident as Record<string, unknown>).triggeredAlerts) &&
    Array.isArray((candidate.incident as Record<string, unknown>).nextActions) &&
    Array.isArray((candidate.incident as Record<string, unknown>).evidence)
  );
}

export function isRuntimeIncidentEventsResponse(
  payload: unknown,
): payload is RuntimeIncidentEventsResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  return (
    Array.isArray(candidate.items) &&
    candidate.items.every((item) => {
      if (typeof item !== "object" || item === null) {
        return false;
      }
      const event = item as Record<string, unknown>;
      return (
        typeof event.id === "string" &&
        typeof event.eventType === "string" &&
        typeof event.alertCode === "string" &&
        typeof event.severity === "string" &&
        typeof event.status === "string" &&
        typeof event.source === "string" &&
        typeof event.title === "string" &&
        typeof event.summary === "string" &&
        typeof event.message === "string" &&
        typeof event.recommendedAction === "string" &&
        typeof event.recommendedSeverity === "string" &&
        typeof event.triggeredAt === "string" &&
        typeof event.firstSeenAt === "string" &&
        typeof event.lastSeenAt === "string" &&
        typeof event.breachCount === "number" &&
        typeof event.traceId === "string" &&
        typeof event.reportGeneratedAt === "string" &&
        typeof event.reportVersion === "string" &&
        Array.isArray(event.runbooks)
      );
    })
  );
}

export function isRuntimeRequestLogsResponse(
  payload: unknown,
): payload is RuntimeRequestLogsResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as Record<string, unknown>;
  return (
    Array.isArray(candidate.items) &&
    candidate.items.every((item) => {
      if (typeof item !== "object" || item === null) {
        return false;
      }
      const record = item as Record<string, unknown>;
      return (
        typeof record.path === "string" &&
        typeof record.method === "string" &&
        typeof record.ip === "string" &&
        (typeof record.traceId === "string" || typeof record.traceId === "undefined") &&
        typeof record.durationMs === "number" &&
        typeof record.statusCode === "number" &&
        typeof record.occurredAt === "string"
      );
    })
  );
}
