package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/models"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/repository"
)

type stubUserRepository struct {
	users []models.User
	err   error
}

func (s stubUserRepository) List(context.Context, int) ([]models.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.users, nil
}

func TestUsersHandlerReturnsFallbackOnlyForUnavailableRepository(t *testing.T) {
	t.Setenv("DEMO_FALLBACK_USERS", "true")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	res := httptest.NewRecorder()

	NewUsersHandler(stubUserRepository{err: errors.New("query timeout")}).List(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for generic repository failures, got %d", res.Code)
	}
}

func TestUsersHandlerReturnsFallbackUsersForUnavailableRepository(t *testing.T) {
	t.Setenv("DEMO_FALLBACK_USERS", "true")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	res := httptest.NewRecorder()

	NewUsersHandler(repository.NewUserRepository(nil)).List(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 when fallback users are enabled, got %d", res.Code)
	}
}

func TestUsersHandlerReturnsServiceUnavailableForUnavailableRepository(t *testing.T) {
	t.Setenv("DEMO_FALLBACK_USERS", "false")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	res := httptest.NewRecorder()

	NewUsersHandler(repository.NewUserRepository(nil)).List(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for unavailable repository, got %d", res.Code)
	}
}
