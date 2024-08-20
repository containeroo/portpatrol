package config_test

import (
	"testing"
	"time"

	"github.com/containeroo/toast/pkg/config"
)

func TestParseConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config with defaults", func(t *testing.T) {
		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
			}
			return env[key]
		}

		expectedCfg := config.Config{
			TargetAddress: "http://example.com",
			Interval:      2 * time.Second,
			DialTimeout:   1 * time.Second,
			CheckType:     "http",
		}

		cfg, err := config.ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("valid config with www as scheme", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "www.example.com:80",
				"INTERVAL":       "5s",
				"DIAL_TIMEOUT":   "10s",
				"CHECK_TYPE":     "http",
			}
			return env[key]
		}

		expectedCfg := config.Config{
			TargetAddress: "www.example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "http",
		}

		cfg, err := config.ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("valid config with custom values", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "tcp://example.com:80",
				"INTERVAL":       "5s",
				"DIAL_TIMEOUT":   "10s",
				"CHECK_TYPE":     "tcp",
			}
			return env[key]
		}

		expectedCfg := config.Config{
			TargetAddress: "tcp://example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "tcp",
		}

		cfg, err := config.ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})
}

func TestParseConfig_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("invalid interval", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
				"INTERVAL":       "invalid",
			}
			return env[key]
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("zero interval", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
				"INTERVAL":       "0s",
			}
			return env[key]
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
	t.Run("invalid dial timeout", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
				"DIAL_TIMEOUT":   "invalid",
			}
			return env[key]
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("zero dial timeout", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
				"DIAL_TIMEOUT":   "0s",
			}
			return env[key]
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("invalid check type", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				"TARGET_ADDRESS": "http://example.com",
				"CHECK_TYPE":     "invalid",
			}
			return env[key]
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("missing target address", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			return ""
		}

		_, err := config.ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
