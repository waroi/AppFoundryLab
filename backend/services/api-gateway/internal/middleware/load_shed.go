package middleware

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type loadShedder struct {
	maxInFlight    int64
	inflight       atomic.Int64
	exemptPrefixes []string
	store          *metrics.Store
}

func newLoadShedder(store *metrics.Store, maxInFlight int, exemptPrefixes []string) *loadShedder {
	if maxInFlight <= 0 {
		return nil
	}
	normalized := make([]string, 0, len(exemptPrefixes))
	for _, prefix := range exemptPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		normalized = append(normalized, prefix)
	}
	return &loadShedder{
		maxInFlight:    int64(maxInFlight),
		exemptPrefixes: normalized,
		store:          store,
	}
}

func (s *loadShedder) isExempt(path string) bool {
	for _, prefix := range s.exemptPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func LoadShedding(store *metrics.Store, maxInFlight int, exemptPrefixes []string) func(http.Handler) http.Handler {
	shedder := newLoadShedder(store, maxInFlight, exemptPrefixes)
	if shedder == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shedder.isExempt(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			current := shedder.inflight.Add(1)
			if current > shedder.maxInFlight {
				shedder.inflight.Add(-1)
				if shedder.store != nil {
					shedder.store.ObserveLoadShed()
				}
				w.Header().Set("Retry-After", "1")
				httpx.WriteError(w, r, http.StatusServiceUnavailable, "server_overloaded", "server overloaded, try again soon", nil)
				return
			}

			defer shedder.inflight.Add(-1)
			next.ServeHTTP(w, r)
		})
	}
}
