package middleware

import (
	"net/http"
	"strings"

	"github.com/example/appfoundrylab/backend/pkg/env"
)

func CORS(next http.Handler) http.Handler {
	allowedOrigins := strings.Split(
		env.GetWithDefault("ALLOWED_ORIGINS", "http://localhost:4321,http://127.0.0.1:4321"),
		",",
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if strings.TrimSpace(allowedOrigin) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
