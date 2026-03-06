package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFibonacciRejectsNilWorker(t *testing.T) {
	h := NewComputeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/compute/fibonacci", strings.NewReader(`{"n":10}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Fibonacci(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHashRejectsNilWorker(t *testing.T) {
	h := NewComputeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/compute/hash", strings.NewReader(`{"input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Hash(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestValidateFibonacciRequest(t *testing.T) {
	cases := []struct {
		name    string
		n       uint32
		wantOK  bool
		wantErr string
	}{
		{"n=0 is valid", 0, true, ""},
		{"n=1 is valid", 1, true, ""},
		{"n=93 is valid (boundary)", 93, true, ""},
		{"n=94 is invalid (above max)", 94, false, "n_out_of_range"},
		{"n=max uint32 is invalid", ^uint32(0), false, "n_out_of_range"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, ok := validateFibonacciRequest(tc.n)
			if ok != tc.wantOK {
				t.Fatalf("n=%d: expected ok=%v, got %v", tc.n, tc.wantOK, ok)
			}
			if !ok && code != tc.wantErr {
				t.Fatalf("n=%d: expected error code %q, got %q", tc.n, tc.wantErr, code)
			}
		})
	}
}

func TestValidateHashRequest(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantOK  bool
		wantErr string
	}{
		{"valid input", "hello world", true, ""},
		{"empty string is invalid", "", false, "input_required"},
		{"whitespace only is invalid", "   ", false, "input_required"},
		{"exactly at limit is valid", strings.Repeat("a", hashMaxInputLen), true, ""},
		{"over limit is invalid", strings.Repeat("a", hashMaxInputLen+1), false, "input_too_long"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, ok := validateHashRequest(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if !ok && code != tc.wantErr {
				t.Fatalf("expected error code %q, got %q", tc.wantErr, code)
			}
		})
	}
}

