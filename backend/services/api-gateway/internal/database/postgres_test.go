package database

import (
	"testing"
)

func TestPostgresDSN(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		want        string
		expectPanic bool
	}{
		{
			name: "All variables provided",
			envVars: map[string]string{
				"POSTGRES_USER":     "myuser",
				"POSTGRES_PASSWORD": "mypassword",
				"POSTGRES_HOST":     "localhost",
				"POSTGRES_PORT":     "5433",
				"POSTGRES_DB":       "mydb",
			},
			want:        "postgres://myuser:mypassword@localhost:5433/mydb?sslmode=disable",
			expectPanic: false,
		},
		{
			name: "Default port used when POSTGRES_PORT is missing",
			envVars: map[string]string{
				"POSTGRES_USER":     "myuser",
				"POSTGRES_PASSWORD": "mypassword",
				"POSTGRES_HOST":     "localhost",
				"POSTGRES_DB":       "mydb",
			},
			want:        "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable",
			expectPanic: false,
		},
		{
			name: "Missing POSTGRES_USER panics",
			envVars: map[string]string{
				"POSTGRES_PASSWORD": "mypassword",
				"POSTGRES_HOST":     "localhost",
				"POSTGRES_DB":       "mydb",
			},
			expectPanic: true,
		},
		{
			name: "Missing POSTGRES_PASSWORD panics",
			envVars: map[string]string{
				"POSTGRES_USER": "myuser",
				"POSTGRES_HOST": "localhost",
				"POSTGRES_DB":   "mydb",
			},
			expectPanic: true,
		},
		{
			name: "Missing POSTGRES_HOST panics",
			envVars: map[string]string{
				"POSTGRES_USER":     "myuser",
				"POSTGRES_PASSWORD": "mypassword",
				"POSTGRES_DB":       "mydb",
			},
			expectPanic: true,
		},
		{
			name: "Missing POSTGRES_DB panics",
			envVars: map[string]string{
				"POSTGRES_USER":     "myuser",
				"POSTGRES_PASSWORD": "mypassword",
				"POSTGRES_HOST":     "localhost",
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars first to avoid cross-test contamination
			// in case the test runner environment has them set.
			t.Setenv("POSTGRES_USER", "")
			t.Setenv("POSTGRES_PASSWORD", "")
			t.Setenv("POSTGRES_HOST", "")
			t.Setenv("POSTGRES_PORT", "")
			t.Setenv("POSTGRES_DB", "")

			// Set test-specific variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Handle expected panic
			defer func() {
				r := recover()
				if tt.expectPanic && r == nil {
					t.Errorf("expected panic, but did not panic")
				} else if !tt.expectPanic && r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			got := PostgresDSN()
			if !tt.expectPanic && got != tt.want {
				t.Errorf("PostgresDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
