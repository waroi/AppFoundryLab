package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type errorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestAuthenticateJWTValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ISSUER", "appfoundrylab")
	t.Setenv("JWT_AUDIENCE", "appfoundrylab-clients")
	t.Setenv("JWT_LEEWAY_SECONDS", "15")

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "appfoundrylab",
			Subject:   "developer",
			Audience:  jwt.ClaimStrings{"appfoundrylab-clients"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
		},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	h := AuthenticateJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAuthenticateJWTInvalidAudience(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ISSUER", "appfoundrylab")
	t.Setenv("JWT_AUDIENCE", "appfoundrylab-clients")
	t.Setenv("JWT_LEEWAY_SECONDS", "15")

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "appfoundrylab",
			Subject:   "developer",
			Audience:  jwt.ClaimStrings{"wrong-audience"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
		},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	h := AuthenticateJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}

	var body errorEnvelope
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != "invalid_token" {
		t.Fatalf("expected invalid_token, got %s", body.Error.Code)
	}
}
