package middleware

import "net/http"

func DeprecatedAPIVersion(successorPath, deprecationDate, sunsetDate string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Deprecation", deprecationDate)
			if sunsetDate != "" {
				w.Header().Set("Sunset", sunsetDate)
			}
			if successorPath != "" {
				w.Header().Set("Link", "<"+successorPath+">; rel=\"successor-version\"")
			}
			w.Header().Set("Warning", `299 - "Deprecated API version; migrate to /api/v1"`)
			next.ServeHTTP(w, r)
		})
	}
}
