package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		payload        any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success with map",
			status:         http.StatusOK,
			payload:        map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}` + "\n",
		},
		{
			name:           "success with struct",
			status:         http.StatusCreated,
			payload:        struct{ ID int `json:"id"` }{ID: 123},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":123}` + "\n",
		},
		{
			name:           "nil payload",
			status:         http.StatusNoContent,
			payload:        nil,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "null\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tt.status, tt.payload)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if ct := res.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", ct)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
