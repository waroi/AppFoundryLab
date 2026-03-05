package env

import (
	"testing"
)

func TestMustGet(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		key := "TEST_MUST_GET_KEY"
		expected := "test_value"
		t.Setenv(key, expected)

		result := MustGet(key)
		if result != expected {
			t.Errorf("MustGet(%q) = %q, want %q", key, result, expected)
		}
	})

	t.Run("panics when env var is missing", func(t *testing.T) {
		key := "TEST_MUST_GET_MISSING_KEY"

		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("MustGet(%q) did not panic as expected", key)
				return
			}

			expectedMsg := "missing required env var: " + key
			if r != expectedMsg {
				t.Errorf("MustGet(%q) paniced with %v, want %v", key, r, expectedMsg)
			}
		}()

		MustGet(key)
	})
}

func TestGetWithDefault(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		key := "TEST_GET_WITH_DEFAULT_KEY"
		expected := "test_value"
		t.Setenv(key, expected)

		result := GetWithDefault(key, "default")
		if result != expected {
			t.Errorf("GetWithDefault(%q) = %q, want %q", key, result, expected)
		}
	})

	t.Run("returns default when env var is missing", func(t *testing.T) {
		key := "TEST_GET_WITH_DEFAULT_MISSING_KEY"
		expected := "default_value"

		result := GetWithDefault(key, expected)
		if result != expected {
			t.Errorf("GetWithDefault(%q) = %q, want %q", key, result, expected)
		}
	})
}

func TestGetIntWithDefault(t *testing.T) {
	t.Run("returns int value when env var is set to valid int", func(t *testing.T) {
		key := "TEST_GET_INT_WITH_DEFAULT_KEY"
		t.Setenv(key, "42")
		expected := 42

		result := GetIntWithDefault(key, 10)
		if result != expected {
			t.Errorf("GetIntWithDefault(%q) = %d, want %d", key, result, expected)
		}
	})

	t.Run("returns default when env var is missing", func(t *testing.T) {
		key := "TEST_GET_INT_WITH_DEFAULT_MISSING_KEY"
		expected := 10

		result := GetIntWithDefault(key, expected)
		if result != expected {
			t.Errorf("GetIntWithDefault(%q) = %d, want %d", key, result, expected)
		}
	})

	t.Run("returns default when env var is set to invalid int", func(t *testing.T) {
		key := "TEST_GET_INT_WITH_DEFAULT_INVALID_KEY"
		t.Setenv(key, "not_an_int")
		expected := 10

		result := GetIntWithDefault(key, expected)
		if result != expected {
			t.Errorf("GetIntWithDefault(%q) = %d, want %d", key, result, expected)
		}
	})
}
