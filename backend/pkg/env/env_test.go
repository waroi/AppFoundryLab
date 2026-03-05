package env

import (
	"os"
	"testing"
)

func TestGetIntWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     *string // nil means don't set
		defaultValue int
		expected     int
	}{
		{
			name:         "env var is set to valid int",
			key:          "TEST_ENV_INT_VALID",
			envValue:     stringPtr("42"),
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "env var is not set",
			key:          "TEST_ENV_INT_UNSET",
			envValue:     nil,
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "env var is empty string",
			key:          "TEST_ENV_INT_EMPTY",
			envValue:     stringPtr(""),
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "env var is invalid int",
			key:          "TEST_ENV_INT_INVALID",
			envValue:     stringPtr("abc"),
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != nil {
				t.Setenv(tt.key, *tt.envValue)
			} else {
				os.Unsetenv(tt.key) // ensure it's not set
			}

			result := GetIntWithDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetIntWithDefault(%q, %d) = %d; want %d", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
