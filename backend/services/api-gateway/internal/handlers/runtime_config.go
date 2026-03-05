package handlers

import (
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type RuntimeHTTPSummary struct {
	LegacyAPIEnabled       bool     `json:"legacyApiEnabled"`
	LegacyDeprecationDate  string   `json:"legacyDeprecationDate"`
	LegacySunsetDate       string   `json:"legacySunsetDate"`
	AuthRateLimitPerMinute int      `json:"authRateLimitPerMinute"`
	APIRateLimitPerMinute  int      `json:"apiRateLimitPerMinute"`
	MaxInFlightRequests    int      `json:"maxInFlightRequests"`
	LoadShedExemptPrefixes []string `json:"loadShedExemptPrefixes"`
	ReadyCacheTTLMS        int64    `json:"readyCacheTtlMs"`
	ReadyStaleIfErrorTTLMS int64    `json:"readyStaleIfErrorTtlMs"`
}

type RuntimeSecuritySummary struct {
	StrictDependencies        bool   `json:"strictDependencies"`
	LoggerSignedIngestEnabled bool   `json:"loggerSignedIngestEnabled"`
	LoggerSharedSecretSet     bool   `json:"loggerSharedSecretSet"`
	LocalAuthMode             string `json:"localAuthMode"`
	WorkerTLSMode             string `json:"workerTlsMode"`
	WorkerServerName          string `json:"workerServerName"`
	DefaultCredentialsInUse   bool   `json:"defaultCredentialsInUse"`
}

type RuntimeOperationsSummary struct {
	AutoMigrate                    bool   `json:"autoMigrate"`
	RateLimitStore                 string `json:"rateLimitStore"`
	RedisFailureMode               string `json:"redisFailureMode"`
	LoggerEndpointConfigured       bool   `json:"loggerEndpointConfigured"`
	RuntimeDiagnosticsCacheTTLMS   int64  `json:"runtimeDiagnosticsCacheTtlMs"`
	IncidentEventSink              string `json:"incidentEventSink"`
	IncidentEventIntervalMS        int64  `json:"incidentEventIntervalMs"`
	IncidentEventDedupeWindow      int64  `json:"incidentEventDedupeWindowMs"`
	IncidentEventWebhookConfigured bool   `json:"incidentEventWebhookConfigured"`
	IncidentEventRetentionDays     int    `json:"incidentEventRetentionDays"`
}

type RuntimeConfigSummary struct {
	Profile    string                   `json:"profile"`
	HTTP       RuntimeHTTPSummary       `json:"http"`
	Security   RuntimeSecuritySummary   `json:"security"`
	Operations RuntimeOperationsSummary `json:"operations"`
	Warnings   []string                 `json:"warnings"`
}

func RuntimeConfigHandler(summary RuntimeConfigSummary) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, summary)
	}
}
