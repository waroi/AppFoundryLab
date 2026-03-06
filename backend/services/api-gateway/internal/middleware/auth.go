package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	claimsContextKey contextKey = "jwt_claims"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func AuthenticateJWT(next http.Handler) http.Handler {
	secret := env.MustGet("JWT_SECRET")
	issuer := env.GetWithDefault("JWT_ISSUER", "appfoundrylab")
	audience := env.GetWithDefault("JWT_AUDIENCE", "appfoundrylab-clients")
	leeway := time.Duration(env.GetIntWithDefault("JWT_LEEWAY_SECONDS", 15)) * time.Second
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithLeeway(leeway),
		jwt.WithIssuedAt(),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.WriteError(w, r, http.StatusUnauthorized, "missing_bearer_token", "missing bearer token", nil)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			httpx.WriteError(w, r, http.StatusUnauthorized, "invalid_token", "invalid token", nil)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			httpx.WriteError(w, r, http.StatusUnauthorized, "invalid_claims", "invalid claims", nil)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, r, http.StatusForbidden, "missing_claims", "missing claims", nil)
				return
			}
			if _, exists := allowed[claims.Role]; !exists {
				httpx.WriteError(w, r, http.StatusForbidden, "insufficient_role", "insufficient role", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
