package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

const maxRuntimeRequestLogsLimit = 100

type RuntimeRequestLogRecord struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	IP         string `json:"ip"`
	TraceID    string `json:"traceId,omitempty"`
	DurationMS int64  `json:"durationMs"`
	StatusCode int    `json:"statusCode"`
	OccurredAt string `json:"occurredAt"`
}

type RuntimeRequestLogsResponse struct {
	Items []RuntimeRequestLogRecord `json:"items"`
}

func RuntimeRequestLogsHandler(loggerEndpoint string, client *http.Client) http.HandlerFunc {
	if client == nil {
		client = &http.Client{Timeout: 800 * time.Millisecond}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := deriveLoggerBaseURL(loggerEndpoint)
		if baseURL == "" {
			httpx.WriteJSON(w, http.StatusOK, RuntimeRequestLogsResponse{Items: []RuntimeRequestLogRecord{}})
			return
		}

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "20"
		} else if parsed, err := strconv.Atoi(limit); err != nil {
			httpx.WriteError(w, r, http.StatusBadRequest, "invalid_query_limit", "invalid limit", nil)
			return
		} else if parsed > maxRuntimeRequestLogsLimit {
			limit = strconv.Itoa(maxRuntimeRequestLogsLimit)
		}

		query := url.Values{}
		query.Set("limit", limit)
		if traceID := r.URL.Query().Get("traceId"); traceID != "" {
			query.Set("traceId", traceID)
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, baseURL+"/request-logs?"+query.Encode(), nil)
		if err != nil {
			httpx.WriteError(w, r, http.StatusInternalServerError, "request_logs_request_failed", "failed to create request", nil)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "logger_unavailable", "failed to query request logs", nil)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "logger_unavailable", "failed to query request logs", nil)
			return
		}

		var payload RuntimeRequestLogsResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			httpx.WriteError(w, r, http.StatusServiceUnavailable, "request_logs_invalid_response", "failed to decode request logs", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payload)
	}
}
