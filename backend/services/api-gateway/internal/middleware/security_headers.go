package middleware

import (
	"net/http"
	"os"
	"strings"
)

func SecurityHeaders(next http.Handler) http.Handler {
	connectSrc := buildConnectSrc(os.Getenv("CSP_CONNECT_SRC"))
	csp := "default-src 'self'; connect-src " + connectSrc

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}

// buildConnectSrc constructs the connect-src directive value.
// extra is an optional space-separated list of additional origins from the
// CSP_CONNECT_SRC environment variable. When empty only 'self' is allowed.
func buildConnectSrc(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return "'self'"
	}
	return "'self' " + extra
}
