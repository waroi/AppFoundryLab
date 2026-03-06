package main

import (
	"os"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/runtimeknobs"
)

func TestResolveLegacyAPIEnabled(t *testing.T) {
	t.Setenv("API_LEGACY_ENABLED", "false")
	t.Setenv("RUNTIME_PROFILE", "standard")
	if resolveLegacyAPIEnabled() {
		t.Fatal("expected false when API_LEGACY_ENABLED is explicitly false")
	}

	t.Setenv("API_LEGACY_ENABLED", "true")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if !resolveLegacyAPIEnabled() {
		t.Fatal("expected explicit API_LEGACY_ENABLED=true to override profile default")
	}

	_ = os.Unsetenv("API_LEGACY_ENABLED")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if resolveLegacyAPIEnabled() {
		t.Fatal("expected secure profile default to disable legacy API")
	}

	t.Setenv("RUNTIME_PROFILE", "minimal")
	if !resolveLegacyAPIEnabled() {
		t.Fatal("expected minimal profile default to enable legacy API")
	}
}

func TestResolveRedisLimiterFailureMode(t *testing.T) {
	t.Setenv("RATE_LIMIT_REDIS_FAILURE_MODE", "closed")
	t.Setenv("RUNTIME_PROFILE", "standard")
	if mode := resolveRedisLimiterFailureMode(); mode != "closed" {
		t.Fatalf("expected explicit closed mode, got %q", mode)
	}

	t.Setenv("RATE_LIMIT_REDIS_FAILURE_MODE", "open")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if mode := resolveRedisLimiterFailureMode(); mode != "open" {
		t.Fatalf("expected explicit open mode override, got %q", mode)
	}

	_ = os.Unsetenv("RATE_LIMIT_REDIS_FAILURE_MODE")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if mode := resolveRedisLimiterFailureMode(); mode != "closed" {
		t.Fatalf("expected secure profile default to closed mode, got %q", mode)
	}

	t.Setenv("RUNTIME_PROFILE", "minimal")
	if mode := resolveRedisLimiterFailureMode(); mode != "open" {
		t.Fatalf("expected minimal profile default to open mode, got %q", mode)
	}
}

func TestValidateLoggerConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("WORKER_GRPC_TLS_MODE", "insecure")
	t.Setenv("LOGGER_ENDPOINT", "")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")

	if err := validateLoggerConfig(); err != nil {
		t.Fatalf("expected nil error when LOGGER_ENDPOINT is empty, got %v", err)
	}

	t.Setenv("LOGGER_ENDPOINT", "http://logger:8090/ingest")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")
	if err := validateLoggerConfig(); err == nil {
		t.Fatal("expected error when secret missing and unsigned ingest disabled")
	}

	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "true")
	if err := validateLoggerConfig(); err != nil {
		t.Fatalf("expected nil error when unsigned ingest enabled, got %v", err)
	}

	t.Setenv("LOGGER_SHARED_SECRET", "test-secret")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")
	if err := validateLoggerConfig(); err != nil {
		t.Fatalf("expected nil error when secret is configured, got %v", err)
	}

	_ = os.Unsetenv("LOGGER_ENDPOINT")
	_ = os.Unsetenv("LOGGER_SHARED_SECRET")
	_ = os.Unsetenv("LOGGER_ALLOW_UNSIGNED_INGEST")
}

func TestResolveMaxInFlightRequests(t *testing.T) {
	t.Setenv("MAX_INFLIGHT_REQUESTS", "64")
	if got := resolveMaxInFlightRequests(); got != 64 {
		t.Fatalf("expected max in-flight=64, got %d", got)
	}

	t.Setenv("MAX_INFLIGHT_REQUESTS", "-5")
	if got := resolveMaxInFlightRequests(); got != 0 {
		t.Fatalf("negative value should disable load shedding, got %d", got)
	}
}

func TestResolveLoadShedExemptPrefixes(t *testing.T) {
	t.Setenv("LOAD_SHED_EXEMPT_PREFIXES", "/health,/metrics,/internal")
	got := resolveLoadShedExemptPrefixes()
	if len(got) != 3 {
		t.Fatalf("expected 3 prefixes, got %d (%v)", len(got), got)
	}
	if got[0] != "/health" || got[1] != "/metrics" || got[2] != "/internal" {
		t.Fatalf("unexpected prefixes: %v", got)
	}

	_ = os.Unsetenv("LOAD_SHED_EXEMPT_PREFIXES")
	got = resolveLoadShedExemptPrefixes()
	if len(got) < 2 || got[0] != "/health" || got[1] != "/metrics" {
		t.Fatalf("expected default prefixes [/health /metrics], got %v", got)
	}
}

func TestDiagnosticsSummaryIncludesRuntimeKnobs(t *testing.T) {
	t.Setenv(runtimeknobs.RequestLogTrustedProxyCIDRsEnv, "127.0.0.1,10.0.0.0/8,invalid")
	t.Setenv("LOGGER_HEALTH_TIMEOUT_MS", "1750")
	t.Setenv("LOGGER_INGEST_TIMESTAMP_MAX_AGE_SECONDS", "420")
	t.Setenv("LOGGER_INGEST_TIMESTAMP_MAX_FUTURE_SKEW_SECONDS", "9")

	summary := diagnosticsSummary(runtimeConfig{
		RuntimeProfile:             "secure",
		LegacyAPIEnabled:           true,
		LegacyDeprecationDate:      "Fri, 27 Feb 2026 00:00:00 GMT",
		LegacySunsetDate:           "Tue, 30 Jun 2026 23:59:59 GMT",
		AuthRateLimitPerMinute:     30,
		APIRateLimitPerMinute:      120,
		MaxInFlightRequests:        64,
		LoadShedExemptPrefixes:     []string{"/health"},
		ReadyCacheTTL:              1500 * time.Millisecond,
		ReadyStaleIfErrorWindow:    5 * time.Second,
		StrictDependencies:         true,
		LoggerAllowUnsignedIngest:  false,
		LoggerSharedSecretSet:      true,
		LocalAuthMode:              "generated",
		WorkerTLSMode:              "mtls",
		WorkerServerName:           "calculator.internal",
		AutoMigrate:                false,
		RateLimitStore:             "redis",
		RedisFailureMode:           "closed",
		RuntimeDiagnosticsCacheTTL: 2500 * time.Millisecond,
		IncidentEventSink:          "logger",
		IncidentEventInterval:      10 * time.Second,
		IncidentEventDedupeWindow:  5 * time.Minute,
		IncidentEventRetentionDays: 30,
	})

	if got := summary.Operations.RequestLogging.TrustedProxyCIDRs; len(got) != 2 || got[0] != "127.0.0.1/32" || got[1] != "10.0.0.0/8" {
		t.Fatalf("unexpected trusted proxy cidrs: %v", got)
	}
	if got := summary.Operations.LoggerTiming.HealthTimeoutMS; got != 1750 {
		t.Fatalf("expected logger health timeout 1750ms, got %d", got)
	}
	if got := summary.Operations.LoggerTiming.IngestTimestampMaxAgeSeconds; got != 420 {
		t.Fatalf("expected logger ingest max age 420s, got %d", got)
	}
	if got := summary.Operations.LoggerTiming.IngestTimestampMaxFutureSkewSeconds; got != 9 {
		t.Fatalf("expected logger ingest future skew 9s, got %d", got)
	}
}
