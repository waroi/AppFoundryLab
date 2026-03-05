package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/runtimecfg"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
	"github.com/golang-jwt/jwt/v5"
)

type tokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func IssueToken(w http.ResponseWriter, r *http.Request) {
	var payload tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	role, ok := resolveRole(payload.Username, payload.Password)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid username or password", nil)
		return
	}

	now := time.Now()
	ttlSeconds := env.GetIntWithDefault("JWT_TTL_SECONDS", 3600)
	audience := env.GetWithDefault("JWT_AUDIENCE", "appfoundrylab-clients")
	claims := middleware.Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    env.GetWithDefault("JWT_ISSUER", "appfoundrylab"),
			Subject:   payload.Username,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlSeconds) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(env.MustGet("JWT_SECRET")))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "token_sign_failed", "failed to issue access token", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"accessToken": signed,
		"tokenType":   "Bearer",
		"expiresIn":   ttlSeconds,
		"role":        role,
	})
}

func resolveRole(username, password string) (string, bool) {
	mode := runtimecfg.ResolveLocalAuthMode()
	if mode == "disabled" {
		return "", false
	}

	adminUser := env.GetWithDefault("BOOTSTRAP_ADMIN_USER", runtimecfg.DefaultAdminUser)
	adminPass := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	regularUser := env.GetWithDefault("BOOTSTRAP_USER", runtimecfg.DefaultRegularUser)
	regularPass := os.Getenv("BOOTSTRAP_USER_PASSWORD")

	if mode == "generated" && runtimecfg.BootstrapDefaultsStillActive(adminUser, adminPass, regularUser, regularPass) {
		return "", false
	}

	switch {
	case adminPass != "" && username == adminUser && password == adminPass:
		return "admin", true
	case regularPass != "" && username == regularUser && password == regularPass:
		return "user", true
	default:
		return "", false
	}
}
