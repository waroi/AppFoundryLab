package handlers

import (
	"fmt"
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

func MetricsHandler(store *metrics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		snapshot := store.Snapshot()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_requests_total Total number of HTTP requests handled by API gateway.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_requests_total counter\n")
		_, _ = fmt.Fprintf(w, "api_gateway_requests_total %d\n", snapshot.RequestsTotal)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_request_errors_total Total number of HTTP requests that returned 5xx.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_request_errors_total counter\n")
		_, _ = fmt.Fprintf(w, "api_gateway_request_errors_total %d\n", snapshot.RequestErrors)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_request_error_rate Share of requests returning 5xx.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_request_error_rate gauge\n")
		_, _ = fmt.Fprintf(w, "api_gateway_request_error_rate %.6f\n", snapshot.ErrorRate)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_request_duration_ms Request latency histogram in milliseconds.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_request_duration_ms histogram\n")

		cumulative := uint64(0)
		for _, bucket := range snapshot.LatencyBucketsMS {
			cumulative += bucket.Count
			_, _ = fmt.Fprintf(w, "api_gateway_request_duration_ms_bucket{le=\"%.0f\"} %d\n", bucket.UpperBoundMS, cumulative)
		}
		cumulative += snapshot.LatencyOverflowMS
		_, _ = fmt.Fprintf(w, "api_gateway_request_duration_ms_bucket{le=\"+Inf\"} %d\n", cumulative)
		_, _ = fmt.Fprintf(w, "api_gateway_request_duration_ms_sum %.0f\n", snapshot.LatencySumMS)
		_, _ = fmt.Fprintf(w, "api_gateway_request_duration_ms_count %d\n", snapshot.LatencyCount)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_load_shed_total Total number of requests rejected by load shedding.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_load_shed_total counter\n")
		_, _ = fmt.Fprintf(w, "api_gateway_load_shed_total %d\n", snapshot.LoadShedTotal)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_inflight_requests Current in-flight request count observed by middleware.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_inflight_requests gauge\n")
		_, _ = fmt.Fprintf(w, "api_gateway_inflight_requests %d\n", snapshot.InflightCurrent)

		_, _ = fmt.Fprintf(w, "# HELP api_gateway_inflight_requests_peak Peak in-flight request count since process start.\n")
		_, _ = fmt.Fprintf(w, "# TYPE api_gateway_inflight_requests_peak gauge\n")
		_, _ = fmt.Fprintf(w, "api_gateway_inflight_requests_peak %d\n", snapshot.InflightPeak)
	}
}
