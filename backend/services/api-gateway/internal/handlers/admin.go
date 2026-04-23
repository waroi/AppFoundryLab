package handlers

import (
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

func AdminPing(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, http.StatusForbidden, "missing_claims", "missing claims", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"scope":  "admin",
		"role":   claims.Role,
	})
}
