package middleware

import (
	"net/http"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
)

type metricsRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *metricsRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func HTTPMetrics(store *metrics.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			store.IncInflight()
			defer store.DecInflight()

			start := time.Now()
			recorder := &metricsRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(recorder, r)

			// Skip self-scrape endpoint to avoid inflating metrics from polling.
			if r.URL.Path == "/metrics" {
				return
			}

			store.Observe(recorder.statusCode, time.Since(start))
		})
	}
}
