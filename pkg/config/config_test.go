package config

import (
	"reflect"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	t.Run("Valid config with defaults", func(t *testing.T) {
		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
			}
			return env[key]
		}

		expectedCfg := Config{
			TargetName:    "example.com", // Extracted from TargetAddress
			TargetAddress: "http://example.com",
			Interval:      2 * time.Second,
			DialTimeout:   1 * time.Second,
			CheckType:     "http",
		}

		cfg, err := ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("Valid config with www as scheme", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "www.example.com:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "http",
			}
			return env[key]
		}

		expectedCfg := Config{
			TargetName:    "www.example.com", // Extracted from TargetAddress
			TargetAddress: "www.example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "http",
		}

		cfg, err := ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("Valid config with kubernetes service", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://postgres.postgres.svc.cluster.local:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "http",
			}
			return env[key]
		}

		expectedCfg := Config{
			TargetName:    "postgres.postgres.svc.cluster.local", // Extracted from TargetAddress
			TargetAddress: "http://postgres.postgres.svc.cluster.local:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "http",
		}

		cfg, err := ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("Valid config with custom values", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "tcp://example.com:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "tcp",
			}
			return env[key]
		}

		expectedCfg := Config{
			TargetName:    "example.com", // Extracted from TargetAddress
			TargetAddress: "tcp://example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "tcp",
		}

		cfg, err := ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if cfg != expectedCfg {
			t.Fatalf("expected config %+v, got %+v", expectedCfg, cfg)
		}
	})

	t.Run("Invalid interval", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envInterval:      "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Zero interval", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envInterval:      "0s",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Invalid dial timeout", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envDialTimeout:   "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Zero dial timeout", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envDialTimeout:   "0s",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Invalid address", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://exam ple.com",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "could not parse target address: parse \"http://exam ple.com\": invalid character \" \" in host name"
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid hostname", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://:8080",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "could not extract hostname from target address: http://:8080"
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid LOG_ADDITIONAL_FIELDS", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress:       "http://example.com",
				envLogAdditionalFields: "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Valid LOG_ADDITIONAL_FIELDS", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress:       "http://example.com",
				envLogAdditionalFields: "true",
			}
			return env[key]
		}

		result, err := ParseConfig(getenv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:          "example.com",
			TargetAddress:       "http://example.com",
			CheckType:           "http",
			Interval:            2 * time.Second,
			DialTimeout:         1 * time.Second,
			LogAdditionalFields: true,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Invalid check type", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envCheckType:     "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Infer invalid check type", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "htp://example.com",
			}
			return env[key]
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Missing target address", func(t *testing.T) {
		t.Parallel()

		getenv := func(key string) string {
			return ""
		}

		_, err := ParseConfig(getenv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
