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

	return func(w http.ResponseWriter, _ *http.Request) {
		baseURL := deriveLoggerBaseURL(loggerEndpoint)
		if baseURL == "" {
			httpx.WriteJSON(w, http.StatusOK, RuntimeIncidentEventsResponse{Items: []RuntimeIncidentEventRecord{}})
			return
		}

		req, err := http.NewRequest(http.MethodGet, baseURL+"/incident-events?limit=20", nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to query incident journal", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			http.Error(w, "failed to query incident journal", http.StatusServiceUnavailable)
			return
		}

		var payload RuntimeIncidentEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			http.Error(w, "failed to decode incident journal", http.StatusServiceUnavailable)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payload)
	}
}
