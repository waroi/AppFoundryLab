package main

import (
	"strings"

	"github.com/example/appfoundrylab/backend/pkg/runtimeknobs"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/runtimecfg"
)

type runtimeConfig = runtimecfg.Config

func loadRuntimeConfig() runtimeConfig {
	return runtimecfg.Load()
}

func validateRuntimeConfig(cfg runtimeConfig) error {
	return runtimecfg.Validate(cfg)
}

func validateLoggerConfig() error {
	return runtimecfg.Validate(runtimecfg.Load())
}

func resolveLegacyAPIEnabled() bool {
	return runtimecfg.ResolveLegacyAPIEnabled()
}

func resolveRedisLimiterFailureMode() string {
	return runtimecfg.ResolveRedisLimiterFailureMode()
}

func resolveMaxInFlightRequests() int {
	return runtimecfg.ResolveMaxInFlightRequests()
}

func resolveLoadShedExemptPrefixes() []string {
	return runtimecfg.ResolveLoadShedExemptPrefixes()
}

func diagnosticsSummary(cfg runtimeConfig) handlers.RuntimeConfigSummary {
	warnings := make([]string, 0, 6)
	if cfg.DefaultCredentialsInUse {
		warnings = append(warnings, "default bootstrap credentials are still active")
	}
	if cfg.LocalAuthMode == "demo" {
		warnings = append(warnings, "local auth demo mode is enabled")
	}
	if cfg.LoggerAllowUnsignedIngest {
		warnings = append(warnings, "logger unsigned ingest is enabled")
	}
	if cfg.IncidentEventSink == "disabled" {
		warnings = append(warnings, "incident event sink is disabled")
	}
	if strings.Contains(cfg.IncidentEventSink, "webhook") && cfg.IncidentEventWebhookURL == "" {
		warnings = append(warnings, "incident webhook sink is selected but webhook url is empty")
	}
	if strings.Contains(cfg.IncidentEventSink, "webhook") && !cfg.IncidentEventWebhookHMACSecretSet {
		warnings = append(warnings, "incident webhook sink is selected but webhook hmac signing is disabled")
	}
	if cfg.RuntimeProfile != "secure" {
		warnings = append(warnings, "runtime profile is not secure")
	}
	if cfg.MaxInFlightRequests == 0 {
		warnings = append(warnings, "load shedding is disabled")
	}
	if !cfg.StrictDependencies {
		warnings = append(warnings, "strict dependencies are disabled; dependency-backed routes degrade per endpoint")
	}

	return handlers.RuntimeConfigSummary{
		Profile: cfg.RuntimeProfile,
		HTTP: handlers.RuntimeHTTPSummary{
			LegacyAPIEnabled:       cfg.LegacyAPIEnabled,
			LegacyDeprecationDate:  cfg.LegacyDeprecationDate,
			LegacySunsetDate:       cfg.LegacySunsetDate,
			AuthRateLimitPerMinute: cfg.AuthRateLimitPerMinute,
			APIRateLimitPerMinute:  cfg.APIRateLimitPerMinute,
			MaxInFlightRequests:    cfg.MaxInFlightRequests,
			LoadShedExemptPrefixes: cfg.LoadShedExemptPrefixes,
			ReadyCacheTTLMS:        cfg.ReadyCacheTTL.Milliseconds(),
			ReadyStaleIfErrorTTLMS: cfg.ReadyStaleIfErrorWindow.Milliseconds(),
		},
		Security: handlers.RuntimeSecuritySummary{
			StrictDependencies:        cfg.StrictDependencies,
			LoggerSignedIngestEnabled: !cfg.LoggerAllowUnsignedIngest,
			LoggerSharedSecretSet:     cfg.LoggerSharedSecretSet,
			LocalAuthMode:             cfg.LocalAuthMode,
			WorkerTLSMode:             cfg.WorkerTLSMode,
			WorkerServerName:          cfg.WorkerServerName,
			DefaultCredentialsInUse:   cfg.DefaultCredentialsInUse,
		},
		Operations: handlers.RuntimeOperationsSummary{
			AutoMigrate:                    cfg.AutoMigrate,
			RateLimitStore:                 cfg.RateLimitStore,
			RedisFailureMode:               cfg.RedisFailureMode,
			LoggerEndpointConfigured:       cfg.LoggerEndpoint != "",
			RuntimeDiagnosticsCacheTTLMS:   cfg.RuntimeDiagnosticsCacheTTL.Milliseconds(),
			IncidentEventSink:              cfg.IncidentEventSink,
			IncidentEventIntervalMS:        cfg.IncidentEventInterval.Milliseconds(),
			IncidentEventDedupeWindow:      cfg.IncidentEventDedupeWindow.Milliseconds(),
			IncidentEventWebhookConfigured: cfg.IncidentEventWebhookURL != "",
			IncidentEventRetentionDays:     cfg.IncidentEventRetentionDays,
			RequestLogging: handlers.RuntimeRequestLoggingSummary{
				TrustedProxyCIDRs: runtimeknobs.RequestLogTrustedProxyCIDRs(),
			},
			LoggerTiming: handlers.RuntimeLoggerTimingSummary{
				HealthTimeoutMS:                     runtimeknobs.LoggerHealthTimeout().Milliseconds(),
				IngestTimestampMaxAgeSeconds:        int64(runtimeknobs.LoggerIngestTimestampMaxAge().Seconds()),
				IngestTimestampMaxFutureSkewSeconds: int64(runtimeknobs.LoggerIngestTimestampMaxFutureSkew().Seconds()),
			},
			DependencyPolicies: dependencyPoliciesSummary(cfg),
		},
		Warnings: warnings,
	}
}

func dependencyPoliciesSummary(cfg runtimeConfig) []handlers.RuntimeDependencyPolicy {
	return []handlers.RuntimeDependencyPolicy{
		{
			Route:           "GET /health/ready",
			Dependency:      "postgres, redis, worker",
			StrictMode:      "gateway startup fails if a required dependency cannot initialize",
			NonStrictMode:   "gateway startup continues and readiness stays degraded until the dependency recovers",
			RuntimeBehavior: "returns 503 with per-dependency checks while any required dependency is down",
		},
		{
			Route:           "GET /api/v1/users",
			Dependency:      "postgres",
			StrictMode:      "gateway startup fails when postgres init, ping, or migration cannot complete",
			NonStrictMode:   "gateway startup continues even when postgres is unavailable",
			RuntimeBehavior: "returns 200 demo fallback users only when DEMO_FALLBACK_USERS=true; otherwise returns 503 users_unavailable",
		},
		{
			Route:           "POST /api/v1/compute/fibonacci and /api/v1/compute/hash",
			Dependency:      "worker",
			StrictMode:      "gateway startup fails when the worker gRPC client cannot initialize or pass health checks",
			NonStrictMode:   "gateway startup continues without a worker client",
			RuntimeBehavior: "returns 503 worker_unavailable until the worker becomes reachable; in-flight RPC failures return 502 worker_call_failed",
		},
		{
			Route:           "POST /api/v1/auth/token and authenticated /api/v1/* rate limiting",
			Dependency:      "redis",
			StrictMode:      "gateway startup fails when redis init or ping cannot complete",
			NonStrictMode:   "gateway startup continues without a healthy redis client",
			RuntimeBehavior: "distributed rate limiting follows RATE_LIMIT_REDIS_FAILURE_MODE=open|closed; open keeps serving traffic, closed returns 503 rate_limiter_unavailable",
		},
		{
			Route:           "GET /api/v1/admin/request-logs",
			Dependency:      "logger",
			StrictMode:      "unchanged; logger is optional at gateway startup",
			NonStrictMode:   "unchanged; logger is optional at gateway startup",
			RuntimeBehavior: "returns 200 with an empty list when LOGGER_ENDPOINT is unset; returns 503 logger_unavailable when logger cannot answer",
		},
		{
			Route:           "GET /api/v1/admin/runtime-metrics and /api/v1/admin/runtime-report",
			Dependency:      "logger",
			StrictMode:      "unchanged; logger is optional at gateway startup",
			NonStrictMode:   "unchanged; logger is optional at gateway startup",
			RuntimeBehavior: "returns 200 while surfacing logger reachability and degraded health warnings inside runtime diagnostics",
		},
	}
}
