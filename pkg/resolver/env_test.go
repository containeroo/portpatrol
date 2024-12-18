package resolver

import (
	"os"
	"testing"
)

func TestEnvResolver_Resolve(t *testing.T) {
	resolver := &EnvResolver{}

	os.Setenv("TEST_ENV_VAR", "test_value")
	os.Setenv("EMPTY_ENV_VAR", "")

	t.Run("Resolve existing environment variable", func(t *testing.T) {
		val, err := resolver.Resolve("TEST_ENV_VAR")
		if err != nil {
			t.Errorf("unexpected error resolving 'TEST_ENV_VAR': %v", err)
		}
		if val != "test_value" {
			t.Errorf("expected 'test_value' but got '%s'", val)
		}
	})

	t.Run("Resolve empty environment variable", func(t *testing.T) {
		val, err := resolver.Resolve("EMPTY_ENV_VAR")
		if err != nil {
			t.Errorf("unexpected error resolving 'EMPTY_ENV_VAR': %v", err)
		}
		if val != "" {
			t.Errorf("expected '' (empty string) but got '%s'", val)
		}
	})

	t.Run("Resolve missing environment variable", func(t *testing.T) {
		_, err := resolver.Resolve("MISSING_ENV_VAR")
		if err == nil {
			t.Error("expected an error resolving 'MISSING_ENV_VAR', but got none")
		}
	})
}
