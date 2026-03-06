export type HealthResponse = {
  status: "ok" | "degraded";
  checks: {
    postgres: "up" | "down";
    redis: "up" | "down";
    worker: "up" | "down";
  };
};

export type TokenResponse = {
  accessToken: string;
  tokenType: "Bearer";
  expiresIn: number;
  role: "admin" | "user";
};

export type UsersResponse = {
  data: Array<{
    id: number;
    name: string;
    email: string;
    createdAt: string;
  }>;
};

export type FibonacciResponse = {
  n: number;
  value: number;
};

export type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
  };
};

export type RuntimeDependencyPolicy = {
  route: string;
  dependency: string;
  strictMode: string;
  nonStrictMode: string;
  runtimeBehavior: string;
};

export type RuntimeConfigResponse = {
  profile: string;
  http: {
    legacyApiEnabled: boolean;
    legacyDeprecationDate: string;
    legacySunsetDate: string;
    authRateLimitPerMinute: number;
    apiRateLimitPerMinute: number;
    maxInFlightRequests: number;
    loadShedExemptPrefixes: string[];
    readyCacheTtlMs: number;
    readyStaleIfErrorTtlMs: number;
  };
  security: {
    strictDependencies: boolean;
    loggerSignedIngestEnabled: boolean;
    loggerSharedSecretSet: boolean;
    localAuthMode: string;
    workerTlsMode: string;
    workerServerName: string;
    defaultCredentialsInUse: boolean;
  };
  operations: {
    autoMigrate: boolean;
    rateLimitStore: string;
    redisFailureMode: string;
    loggerEndpointConfigured: boolean;
    runtimeDiagnosticsCacheTtlMs: number;
    incidentEventSink: string;
    incidentEventIntervalMs: number;
    incidentEventDedupeWindowMs: number;
    incidentEventWebhookConfigured: boolean;
    incidentEventRetentionDays: number;
    requestLogging: {
      trustedProxyCidrs: string[];
    };
    loggerTiming: {
      healthTimeoutMs: number;
      ingestTimestampMaxAgeSeconds: number;
      ingestTimestampMaxFutureSkewSeconds: number;
    };
    dependencyPolicies: RuntimeDependencyPolicy[];
  };
  warnings: string[];
};

export type RuntimeMetricsResponse = {
  requestsTotal: number;
  requestErrors: number;
  errorRate: number;
  latencyCount: number;
  latencyAverageMs: number;
  loadShedTotal: number;
  inflightCurrent: number;
  inflightPeak: number;
  recentHistory: Array<{
    recordedAt: string;
    requestsTotal: number;
    requestErrors: number;
    errorRate: number;
    latencyAverageMs: number;
    loadShedTotal: number;
    inflightCurrent: number;
    inflightPeak: number;
  }>;
  alerts: {
    activeCount: number;
    highestSeverity: string;
    recentlyBreached: boolean;
    items: Array<{
      code: string;
      severity: string;
      status: string;
      source: string;
      message: string;
      recommendedAction: string;
      breachCount: number;
      lastTriggeredAt: string;
    }>;
  };
  health: {
    status: string;
    httpStatus: number;
    postgres: string;
    redis: string;
    worker: string;
    cacheState: string;
    cacheAgeMs: number;
    cacheTtlMs: number;
    staleIfErrorTtlMs: number;
    lastCheckedAt: string;
  };
  trace: {
    enabled: boolean;
    responseHeader: string;
    forwardedToLogger: boolean;
    storedOnLoggerAs: string;
    storageField: string;
  };
  gatewayLogger: {
    enabled: boolean;
    endpoint: string;
    queueDepth: number;
    queueCapacity: number;
    workers: number;
    retryMax: number;
    droppedTotal: number;
  };
  loggerService: {
    configured: boolean;
    reachable: boolean;
    endpointBase: string;
    healthStatus: string;
    queueDepth: number;
    queueCapacity: number;
    workers: number;
    enqueuedTotal: number;
    droppedTotal: number;
    processedTotal: number;
    failedTotal: number;
    retriedTotal: number;
    inflightWorkers: number;
    dropRatio: number;
    dropAlertThresholdPct: number;
    dropAlertThresholdHit: boolean;
    lastError: string;
  };
  incidentJournal: {
    enabled: boolean;
    sink: string;
    configured: boolean;
    reachable: boolean;
    totalEvents: number;
    activeEvents: number;
    latestEventAt: string;
    lastEventStatus: string;
    dispatchFailures: number;
    lastDispatchAt: string;
    lastDispatchError: string;
  };
  warnings: string[];
};

export type RuntimeReportResponse = {
  generatedAt: string;
  reportVersion: string;
  config: RuntimeConfigResponse;
  metrics: RuntimeMetricsResponse;
  runbooks: Array<{
    id: string;
    title: string;
    path: string;
    reason: string;
    priority: string;
  }>;
  incident: {
    recommendedSeverity: string;
    category: string;
    title: string;
    summary: string;
    suspectedSystems: string[];
    triggeredAlerts: string[];
    nextActions: string[];
    evidence: Array<{
      kind: string;
      label: string;
      value: string;
      source: string;
    }>;
  };
};

export type RuntimeIncidentEventsResponse = {
  items: Array<{
    id: string;
    eventType: string;
    alertCode: string;
    severity: string;
    status: string;
    source: string;
    title: string;
    summary: string;
    message: string;
    recommendedAction: string;
    recommendedSeverity: string;
    triggeredAt: string;
    firstSeenAt: string;
    lastSeenAt: string;
    breachCount: number;
    traceId: string;
    reportGeneratedAt: string;
    reportVersion: string;
    runbooks: Array<{
      id: string;
      title: string;
      path: string;
      reason: string;
      priority: string;
    }>;
  }>;
};

export type RuntimeRequestLogRecord = {
  path: string;
  method: string;
  ip: string;
  traceId?: string;
  durationMs: number;
  statusCode: number;
  occurredAt: string;
};

export type RuntimeRequestLogsResponse = {
  items: RuntimeRequestLogRecord[];
};
