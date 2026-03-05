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
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		} else if parsed > maxRuntimeRequestLogsLimit {
			limit = strconv.Itoa(maxRuntimeRequestLogsLimit)
		}

		query := url.Values{}
		query.Set("limit", limit)
		if traceID := r.URL.Query().Get("traceId"); traceID != "" {
			query.Set("traceId", traceID)
		}

		req, err := http.NewRequest(http.MethodGet, baseURL+"/request-logs?"+query.Encode(), nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to query request logs", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			http.Error(w, "failed to query request logs", http.StatusServiceUnavailable)
			return
		}

		var payload RuntimeRequestLogsResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			http.Error(w, "failed to decode request logs", http.StatusServiceUnavailable)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payload)
	}
}
