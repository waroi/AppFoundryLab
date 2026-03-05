package main

import (
	"strings"

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
		},
		Warnings: warnings,
	}
}
