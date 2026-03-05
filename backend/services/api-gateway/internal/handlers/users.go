package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/models"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/repository"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type UsersHandler struct {
	repo repository.UserRepository
}

func NewUsersHandler(repo repository.UserRepository) *UsersHandler {
	return &UsersHandler{repo: repo}
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	users, err := h.repo.List(ctx, 20)
	if err != nil {
		if os.Getenv("DEMO_FALLBACK_USERS") == "true" {
			fallbackUsers := []models.User{
				{ID: 1, Name: "Demo User", Email: "demo@appfoundrylab.local", CreatedAt: time.Now().UTC()},
			}
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": fallbackUsers})
			return
		}
		httpx.WriteError(w, r, http.StatusInternalServerError, "failed_to_fetch_users", "failed to fetch users", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": users})
}
