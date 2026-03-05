package main

import (
	"os"
	"testing"
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
