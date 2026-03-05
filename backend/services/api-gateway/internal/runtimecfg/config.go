package runtimecfg

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
)

const (
	DefaultAdminUser     = "admin"
	DefaultAdminPassword = "admin_dev_password"
	DefaultRegularUser   = "developer"
	DefaultRegularPass   = "developer_dev_password"
)

type Config struct {
	StrictDependencies                bool
	AutoMigrate                       bool
	RuntimeProfile                    string
	APIGatewayPort                    string
	LoggerEndpoint                    string
	LoggerAllowUnsignedIngest         bool
	LoggerSharedSecretSet             bool
	LocalAuthMode                     string
	IncidentEventSink                 string
	IncidentEventInterval             time.Duration
	IncidentEventDedupeWindow         time.Duration
	IncidentEventWebhookURL           string
	IncidentEventWebhookHMACSecretSet bool
	IncidentEventWebhookAllowedHosts  []string
	IncidentEventRetentionDays        int
	AuthRateLimitPerMinute            int
	APIRateLimitPerMinute             int
	RateLimitStore                    string
	RedisFailureMode                  string
	LegacyAPIEnabled                  bool
	LegacyDeprecationDate             string
	LegacySunsetDate                  string
	MaxInFlightRequests               int
	LoadShedExemptPrefixes            []string
	ReadyCacheTTL                     time.Duration
	ReadyStaleIfErrorWindow           time.Duration
	RuntimeDiagnosticsCacheTTL        time.Duration
	WorkerTLSMode                     string
	WorkerServerName                  string
	BootstrapAdminUser                string
	BootstrapUser                     string
	DefaultCredentialsInUse           bool
}

func Load() Config {
	return Config{
		StrictDependencies:                env.GetWithDefault("STRICT_DEPENDENCIES", "true") == "true",
		AutoMigrate:                       env.GetWithDefault("POSTGRES_AUTO_MIGRATE", "true") == "true",
		RuntimeProfile:                    strings.ToLower(strings.TrimSpace(env.GetWithDefault("RUNTIME_PROFILE", "standard"))),
		APIGatewayPort:                    env.GetWithDefault("API_GATEWAY_PORT", "8080"),
		LoggerEndpoint:                    os.Getenv("LOGGER_ENDPOINT"),
		LoggerAllowUnsignedIngest:         env.GetWithDefault("LOGGER_ALLOW_UNSIGNED_INGEST", "false") == "true",
		LoggerSharedSecretSet:             strings.TrimSpace(os.Getenv("LOGGER_SHARED_SECRET")) != "",
		LocalAuthMode:                     ResolveLocalAuthMode(),
		IncidentEventSink:                 ResolveIncidentEventSink(),
		IncidentEventInterval:             time.Duration(env.GetIntWithDefault("INCIDENT_EVENT_INTERVAL_MS", 10000)) * time.Millisecond,
		IncidentEventDedupeWindow:         time.Duration(env.GetIntWithDefault("INCIDENT_EVENT_DEDUPE_WINDOW_SECONDS", 300)) * time.Second,
		IncidentEventWebhookURL:           strings.TrimSpace(os.Getenv("INCIDENT_EVENT_WEBHOOK_URL")),
		IncidentEventWebhookHMACSecretSet: strings.TrimSpace(os.Getenv("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET")) != "",
		IncidentEventWebhookAllowedHosts:  ResolveWebhookAllowedHosts(),
		IncidentEventRetentionDays:        env.GetIntWithDefault("INCIDENT_EVENT_RETENTION_DAYS", 30),
		AuthRateLimitPerMinute:            env.GetIntWithDefault("AUTH_RATE_LIMIT_PER_MINUTE", 30),
		APIRateLimitPerMinute:             env.GetIntWithDefault("API_RATE_LIMIT_PER_MINUTE", 120),
		RateLimitStore:                    env.GetWithDefault("RATE_LIMIT_STORE", "memory"),
		RedisFailureMode:                  ResolveRedisLimiterFailureMode(),
		LegacyAPIEnabled:                  ResolveLegacyAPIEnabled(),
		LegacyDeprecationDate:             env.GetWithDefault("API_LEGACY_DEPRECATION_DATE", "Fri, 27 Feb 2026 00:00:00 GMT"),
		LegacySunsetDate:                  env.GetWithDefault("API_LEGACY_SUNSET_DATE", "Tue, 30 Jun 2026 23:59:59 GMT"),
		MaxInFlightRequests:               ResolveMaxInFlightRequests(),
		LoadShedExemptPrefixes:            ResolveLoadShedExemptPrefixes(),
		ReadyCacheTTL:                     time.Duration(env.GetIntWithDefault("HEALTH_READY_CACHE_TTL_MS", 1000)) * time.Millisecond,
		ReadyStaleIfErrorWindow:           time.Duration(env.GetIntWithDefault("HEALTH_READY_STALE_IF_ERROR_MS", 10000)) * time.Millisecond,
		RuntimeDiagnosticsCacheTTL:        time.Duration(env.GetIntWithDefault("RUNTIME_DIAGNOSTICS_CACHE_TTL_MS", 1000)) * time.Millisecond,
		WorkerTLSMode:                     env.GetWithDefault("WORKER_GRPC_TLS_MODE", "mtls"),
		WorkerServerName:                  env.GetWithDefault("WORKER_GRPC_SERVER_NAME", "calculator"),
		BootstrapAdminUser:                env.GetWithDefault("BOOTSTRAP_ADMIN_USER", DefaultAdminUser),
		BootstrapUser:                     env.GetWithDefault("BOOTSTRAP_USER", DefaultRegularUser),
		DefaultCredentialsInUse:           ResolveDefaultCredentialsInUse(),
	}
}

func Validate(cfg Config) error {
	if cfg.LoggerEndpoint == "" {
		if err := validateIncidentWebhook(cfg); err != nil {
			return err
		}
		return nil
	}
	if !cfg.LoggerAllowUnsignedIngest && !cfg.LoggerSharedSecretSet {
		return errors.New("LOGGER_SHARED_SECRET is required when LOGGER_ENDPOINT is set and LOGGER_ALLOW_UNSIGNED_INGEST=false")
	}
	if err := validateIncidentWebhook(cfg); err != nil {
		return err
	}
	return nil
}

func validateIncidentWebhook(cfg Config) error {
	if !strings.Contains(cfg.IncidentEventSink, "webhook") {
		return nil
	}
	if cfg.IncidentEventWebhookURL == "" {
		return errors.New("INCIDENT_EVENT_WEBHOOK_URL is required when INCIDENT_EVENT_SINK includes webhook")
	}
	if !cfg.IncidentEventWebhookHMACSecretSet {
		return errors.New("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET is required when INCIDENT_EVENT_SINK includes webhook")
	}

	parsed, err := url.Parse(cfg.IncidentEventWebhookURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("INCIDENT_EVENT_WEBHOOK_URL must be a valid absolute URL")
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme != "https" && host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return errors.New("INCIDENT_EVENT_WEBHOOK_URL must use https outside local development")
	}
	if len(cfg.IncidentEventWebhookAllowedHosts) == 0 {
		return nil
	}
	for _, allowed := range cfg.IncidentEventWebhookAllowedHosts {
		if strings.EqualFold(host, allowed) {
			return nil
		}
	}
	return errors.New("INCIDENT_EVENT_WEBHOOK_URL host is not present in INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS")
}

func ResolveDefaultCredentialsInUse() bool {
	adminUser := env.GetWithDefault("BOOTSTRAP_ADMIN_USER", DefaultAdminUser)
	adminPass := env.GetWithDefault("BOOTSTRAP_ADMIN_PASSWORD", DefaultAdminPassword)
	regularUser := env.GetWithDefault("BOOTSTRAP_USER", DefaultRegularUser)
	regularPass := env.GetWithDefault("BOOTSTRAP_USER_PASSWORD", DefaultRegularPass)

	return BootstrapDefaultsStillActive(adminUser, adminPass, regularUser, regularPass)
}

func BootstrapDefaultsStillActive(adminUser, adminPass, regularUser, regularPass string) bool {
	return adminUser == DefaultAdminUser &&
		adminPass == DefaultAdminPassword &&
		regularUser == DefaultRegularUser &&
		regularPass == DefaultRegularPass
}

func ResolveLegacyAPIEnabled() bool {
	if raw, exists := os.LookupEnv("API_LEGACY_ENABLED"); exists && raw != "" {
		return strings.EqualFold(raw, "true")
	}

	profile := strings.ToLower(strings.TrimSpace(env.GetWithDefault("RUNTIME_PROFILE", "standard")))
	return profile != "secure"
}

func ResolveRedisLimiterFailureMode() string {
	if mode, exists := os.LookupEnv("RATE_LIMIT_REDIS_FAILURE_MODE"); exists && mode != "" {
		mode = strings.ToLower(strings.TrimSpace(mode))
		if mode == "closed" {
			return "closed"
		}
		return "open"
	}

	profile := strings.ToLower(strings.TrimSpace(env.GetWithDefault("RUNTIME_PROFILE", "standard")))
	if profile == "secure" {
		return "closed"
	}
	return "open"
}

func ResolveMaxInFlightRequests() int {
	maxInFlight := env.GetIntWithDefault("MAX_INFLIGHT_REQUESTS", 0)
	if maxInFlight < 0 {
		return 0
	}
	return maxInFlight
}

func ResolveLoadShedExemptPrefixes() []string {
	raw := env.GetWithDefault("LOAD_SHED_EXEMPT_PREFIXES", "/health,/metrics")
	parts := strings.Split(raw, ",")
	prefixes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		prefixes = append(prefixes, part)
	}
	return prefixes
}

func ResolveLocalAuthMode() string {
	mode := strings.ToLower(strings.TrimSpace(env.GetWithDefault("LOCAL_AUTH_MODE", "demo")))
	switch mode {
	case "generated", "disabled":
		return mode
	case "demo":
		return "demo"
	default:
		return "generated"
	}
}

func ResolveIncidentEventSink() string {
	raw := strings.ToLower(strings.TrimSpace(env.GetWithDefault("INCIDENT_EVENT_SINK", "logger")))
	if raw == "" {
		return "logger"
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '+' || r == ','
	})

	enabled := map[string]bool{
		"logger":  false,
		"stdout":  false,
		"webhook": false,
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if _, ok := enabled[part]; ok {
			enabled[part] = true
		}
	}

	ordered := make([]string, 0, len(enabled))
	for _, key := range []string{"logger", "stdout", "webhook"} {
		if enabled[key] {
			ordered = append(ordered, key)
		}
	}

	if len(ordered) == 0 {
		if raw == "disabled" {
			return "disabled"
		}
		return "logger"
	}

	return strings.Join(ordered, "+")
}

func ResolveWebhookAllowedHosts() []string {
	raw := strings.TrimSpace(os.Getenv("INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS"))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	hosts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		hosts = append(hosts, part)
	}
	return hosts
}
