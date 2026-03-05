package runtimecfg

import (
	"os"
	"testing"
	"time"
)

func TestResolveLegacyAPIEnabled(t *testing.T) {
	t.Setenv("API_LEGACY_ENABLED", "false")
	t.Setenv("RUNTIME_PROFILE", "standard")
	if ResolveLegacyAPIEnabled() {
		t.Fatal("expected false when API_LEGACY_ENABLED is explicitly false")
	}

	t.Setenv("API_LEGACY_ENABLED", "true")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if !ResolveLegacyAPIEnabled() {
		t.Fatal("expected explicit API_LEGACY_ENABLED=true to override profile default")
	}

	_ = os.Unsetenv("API_LEGACY_ENABLED")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if ResolveLegacyAPIEnabled() {
		t.Fatal("expected secure profile default to disable legacy API")
	}

	t.Setenv("RUNTIME_PROFILE", "minimal")
	if !ResolveLegacyAPIEnabled() {
		t.Fatal("expected minimal profile default to enable legacy API")
	}
}

func TestResolveRedisLimiterFailureMode(t *testing.T) {
	t.Setenv("RATE_LIMIT_REDIS_FAILURE_MODE", "closed")
	t.Setenv("RUNTIME_PROFILE", "standard")
	if mode := ResolveRedisLimiterFailureMode(); mode != "closed" {
		t.Fatalf("expected explicit closed mode, got %q", mode)
	}

	t.Setenv("RATE_LIMIT_REDIS_FAILURE_MODE", "open")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if mode := ResolveRedisLimiterFailureMode(); mode != "open" {
		t.Fatalf("expected explicit open mode override, got %q", mode)
	}

	_ = os.Unsetenv("RATE_LIMIT_REDIS_FAILURE_MODE")
	t.Setenv("RUNTIME_PROFILE", "secure")
	if mode := ResolveRedisLimiterFailureMode(); mode != "closed" {
		t.Fatalf("expected secure profile default to closed mode, got %q", mode)
	}

	t.Setenv("RUNTIME_PROFILE", "minimal")
	if mode := ResolveRedisLimiterFailureMode(); mode != "open" {
		t.Fatalf("expected minimal profile default to open mode, got %q", mode)
	}
}

func TestValidateLoggerConfig(t *testing.T) {
	t.Setenv("LOGGER_ENDPOINT", "")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")

	if err := Validate(Load()); err != nil {
		t.Fatalf("expected nil error when LOGGER_ENDPOINT is empty, got %v", err)
	}

	t.Setenv("LOGGER_ENDPOINT", "http://logger:8090/ingest")
	t.Setenv("LOGGER_SHARED_SECRET", "")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")
	if err := Validate(Load()); err == nil {
		t.Fatal("expected error when secret missing and unsigned ingest disabled")
	}

	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "true")
	if err := Validate(Load()); err != nil {
		t.Fatalf("expected nil error when unsigned ingest enabled, got %v", err)
	}

	t.Setenv("LOGGER_SHARED_SECRET", "test-secret")
	t.Setenv("LOGGER_ALLOW_UNSIGNED_INGEST", "false")
	if err := Validate(Load()); err != nil {
		t.Fatalf("expected nil error when secret is configured, got %v", err)
	}
}

func TestResolveMaxInFlightRequests(t *testing.T) {
	t.Setenv("MAX_INFLIGHT_REQUESTS", "64")
	if got := ResolveMaxInFlightRequests(); got != 64 {
		t.Fatalf("expected max in-flight=64, got %d", got)
	}

	t.Setenv("MAX_INFLIGHT_REQUESTS", "-5")
	if got := ResolveMaxInFlightRequests(); got != 0 {
		t.Fatalf("negative value should disable load shedding, got %d", got)
	}
}

func TestResolveLoadShedExemptPrefixes(t *testing.T) {
	t.Setenv("LOAD_SHED_EXEMPT_PREFIXES", "/health,/metrics,/internal")
	got := ResolveLoadShedExemptPrefixes()
	if len(got) != 3 {
		t.Fatalf("expected 3 prefixes, got %d (%v)", len(got), got)
	}
	if got[0] != "/health" || got[1] != "/metrics" || got[2] != "/internal" {
		t.Fatalf("unexpected prefixes: %v", got)
	}

	_ = os.Unsetenv("LOAD_SHED_EXEMPT_PREFIXES")
	got = ResolveLoadShedExemptPrefixes()
	if len(got) < 2 || got[0] != "/health" || got[1] != "/metrics" {
		t.Fatalf("expected default prefixes [/health /metrics], got %v", got)
	}
}

func TestLoadRuntimeDiagnosticsCacheTTL(t *testing.T) {
	t.Setenv("RUNTIME_DIAGNOSTICS_CACHE_TTL_MS", "2500")
	if got := Load().RuntimeDiagnosticsCacheTTL; got != 2500*time.Millisecond {
		t.Fatalf("expected diagnostics cache ttl 2500ms, got %s", got)
	}
}

func TestResolveLocalAuthMode(t *testing.T) {
	t.Setenv("LOCAL_AUTH_MODE", "generated")
	if mode := ResolveLocalAuthMode(); mode != "generated" {
		t.Fatalf("expected generated mode, got %q", mode)
	}

	t.Setenv("LOCAL_AUTH_MODE", "disabled")
	if mode := ResolveLocalAuthMode(); mode != "disabled" {
		t.Fatalf("expected disabled mode, got %q", mode)
	}

	t.Setenv("LOCAL_AUTH_MODE", "unexpected")
	if mode := ResolveLocalAuthMode(); mode != "demo" {
		t.Fatalf("expected fallback demo mode, got %q", mode)
	}
}

func TestResolveIncidentEventSink(t *testing.T) {
	t.Setenv("INCIDENT_EVENT_SINK", "webhook,logger")
	if sink := ResolveIncidentEventSink(); sink != "logger+webhook" {
		t.Fatalf("expected canonical sink logger+webhook, got %q", sink)
	}

	t.Setenv("INCIDENT_EVENT_SINK", "stdout+webhook+logger")
	if sink := ResolveIncidentEventSink(); sink != "logger+stdout+webhook" {
		t.Fatalf("expected canonical sink logger+stdout+webhook, got %q", sink)
	}

	t.Setenv("INCIDENT_EVENT_SINK", "disabled")
	if sink := ResolveIncidentEventSink(); sink != "disabled" {
		t.Fatalf("expected disabled sink, got %q", sink)
	}

	t.Setenv("INCIDENT_EVENT_SINK", "unexpected")
	if sink := ResolveIncidentEventSink(); sink != "logger" {
		t.Fatalf("expected logger fallback, got %q", sink)
	}
}

func TestResolveWebhookAllowedHosts(t *testing.T) {
	t.Setenv("INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS", "alerts.example.com, hooks.example.net ")
	got := ResolveWebhookAllowedHosts()
	if len(got) != 2 || got[0] != "alerts.example.com" || got[1] != "hooks.example.net" {
		t.Fatalf("unexpected hosts: %v", got)
	}
}

func TestValidateWebhookConfig(t *testing.T) {
	t.Setenv("INCIDENT_EVENT_SINK", "logger+webhook")
	t.Setenv("INCIDENT_EVENT_WEBHOOK_URL", "https://alerts.example.com/hooks/runtime")
	t.Setenv("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET", "signed")
	t.Setenv("INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS", "alerts.example.com")
	if err := Validate(Load()); err != nil {
		t.Fatalf("expected valid webhook config, got %v", err)
	}

	t.Setenv("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET", "")
	if err := Validate(Load()); err == nil {
		t.Fatal("expected error when webhook sink is enabled without HMAC secret")
	}

	t.Setenv("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET", "signed")
	t.Setenv("INCIDENT_EVENT_WEBHOOK_URL", "http://alerts.example.com/hooks/runtime")
	if err := Validate(Load()); err == nil {
		t.Fatal("expected error when remote webhook is not https")
	}
}
