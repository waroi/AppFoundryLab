package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := randomTraceID()
		w.Header().Set(httpx.TraceIDHeader, traceID)
		ctx := httpx.ContextWithTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func randomTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "trace-id-unavailable"
	}
	return hex.EncodeToString(buf)
}
