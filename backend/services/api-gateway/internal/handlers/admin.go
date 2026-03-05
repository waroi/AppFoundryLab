package handlers

import (
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

func AdminPing(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"scope":  "admin",
		"role":   claims.Role,
	})
}
