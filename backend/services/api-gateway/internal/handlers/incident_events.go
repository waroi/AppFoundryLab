package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

func RuntimeIncidentReportHandler(config RuntimeConfigSummary, store MetricsSnapshotProvider, options RuntimeMetricsOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, BuildRuntimeReportSummary(config, store, options))
	}
}

func RuntimeIncidentEventsHandler(loggerEndpoint string, client *http.Client) http.HandlerFunc {
	if client == nil {
		client = &http.Client{Timeout: 800 * time.Millisecond}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := deriveLoggerBaseURL(loggerEndpoint)
		if baseURL == "" {
			httpx.WriteJSON(w, http.StatusOK, RuntimeIncidentEventsResponse{Items: []RuntimeIncidentEventRecord{}})
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, baseURL+"/incident-events?limit=20", nil)
		if err != nil {
			httpx.WriteError(w, r, http.StatusInternalServerError, "incident_events_request_failed", "failed to create request", nil)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "logger_unavailable", "failed to query incident journal", nil)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "logger_unavailable", "failed to query incident journal", nil)
			return
		}

		var payload RuntimeIncidentEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "incident_events_invalid_response", "failed to decode incident journal", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payload)
	}
}
