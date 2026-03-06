package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCORS(t *testing.T) {
	// Set environment variable for allowed origins
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	tests := []struct {
		name               string
		method             string
		origin             string
		expectedStatus     int
		expectedCORSOrigin string
		handlerCalled      bool
	}{
		{
			name:               "allowed origin, GET request",
			method:             http.MethodGet,
			origin:             "http://localhost:3000",
			expectedStatus:     http.StatusOK,
			expectedCORSOrigin: "http://localhost:3000",
			handlerCalled:      true,
		},
		{
			name:               "allowed origin 2, POST request",
			method:             http.MethodPost,
			origin:             "https://example.com",
			expectedStatus:     http.StatusOK,
			expectedCORSOrigin: "https://example.com",
			handlerCalled:      true,
		},
		{
			name:               "disallowed origin, GET request",
			method:             http.MethodGet,
			origin:             "http://malicious.com",
			expectedStatus:     http.StatusOK,
			expectedCORSOrigin: "",
			handlerCalled:      true,
		},
		{
			name:               "no origin, GET request",
			method:             http.MethodGet,
			origin:             "",
			expectedStatus:     http.StatusOK,
			expectedCORSOrigin: "",
			handlerCalled:      true,
		},
		{
			name:               "allowed origin, OPTIONS request (preflight)",
			method:             http.MethodOptions,
			origin:             "http://localhost:3000",
			expectedStatus:     http.StatusNoContent,
			expectedCORSOrigin: "http://localhost:3000",
			handlerCalled:      false,
		},
		{
			name:               "disallowed origin, OPTIONS request (preflight)",
			method:             http.MethodOptions,
			origin:             "http://malicious.com",
			expectedStatus:     http.StatusNoContent,
			expectedCORSOrigin: "",
			handlerCalled:      false,
		},
		{
			name:               "no origin, OPTIONS request",
			method:             http.MethodOptions,
			origin:             "",
			expectedStatus:     http.StatusNoContent,
			expectedCORSOrigin: "",
			handlerCalled:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware instance
			// Note: since env.GetWithDefault is called when CORS() is invoked,
			// the os.Setenv affects it properly here.
			handler := CORS(nextHandler)

			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check Access-Control-Allow-Origin header
			if corsOrigin := rr.Header().Get("Access-Control-Allow-Origin"); corsOrigin != tt.expectedCORSOrigin {
				t.Errorf("handler returned wrong Access-Control-Allow-Origin: got %v want %v", corsOrigin, tt.expectedCORSOrigin)
			}

			// Check standard headers that should always be present
			if vary := rr.Header().Get("Vary"); vary != "Origin" {
				t.Errorf("handler returned wrong Vary header: got %v want %v", vary, "Origin")
			}
			if allowHeaders := rr.Header().Get("Access-Control-Allow-Headers"); allowHeaders != "Authorization, Content-Type" {
				t.Errorf("handler returned wrong Access-Control-Allow-Headers: got %v want %v", allowHeaders, "Authorization, Content-Type")
			}
			if allowMethods := rr.Header().Get("Access-Control-Allow-Methods"); allowMethods != "GET, POST, OPTIONS" {
				t.Errorf("handler returned wrong Access-Control-Allow-Methods: got %v want %v", allowMethods, "GET, POST, OPTIONS")
			}

			// Check if next handler was called
			if handlerCalled != tt.handlerCalled {
				t.Errorf("next handler called: got %v want %v", handlerCalled, tt.handlerCalled)
			}
		})
	}
}

func TestCORS_DefaultOrigins(t *testing.T) {
	// Ensure ALLOWED_ORIGINS is unset to test default behavior
	os.Unsetenv("ALLOWED_ORIGINS")

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:4321")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if corsOrigin := rr.Header().Get("Access-Control-Allow-Origin"); corsOrigin != "http://localhost:4321" {
		t.Errorf("handler returned wrong Access-Control-Allow-Origin: got %v want %v", corsOrigin, "http://localhost:4321")
	}

	if !handlerCalled {
		t.Errorf("next handler was not called")
	}
}
