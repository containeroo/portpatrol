package config

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	t.Run("Valid config with defaults", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:    "example.com", // Extracted from TargetAddress
			TargetAddress: "http://example.com",
			Interval:      2 * time.Second,
			DialTimeout:   1 * time.Second,
			CheckType:     "http",
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with www as scheme", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "www.example.com:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "http",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:    "www.example.com", // Extracted from TargetAddress
			TargetAddress: "www.example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "http",
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with kubernetes service", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://postgres.postgres.svc.cluster.local:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "http",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:    "postgres.postgres.svc.cluster.local", // Extracted from TargetAddress
			TargetAddress: "http://postgres.postgres.svc.cluster.local:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "http",
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with custom values", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "tcp://example.com:80",
				envInterval:      "5s",
				envDialTimeout:   "10s",
				envCheckType:     "tcp",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:    "example.com", // Extracted from TargetAddress
			TargetAddress: "tcp://example.com:80",
			Interval:      5 * time.Second,
			DialTimeout:   10 * time.Second,
			CheckType:     "tcp",
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Invalid interval (invalid)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envInterval:      "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envInterval)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid interval (zero)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envInterval:      "0s",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: 0s", envInterval)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid dial timeout (invalid)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envDialTimeout:   "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envDialTimeout)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid dial timeout (zero)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envDialTimeout:   "0s",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Invalid address (invalid address)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://exam ple.com",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "could not parse target address: parse \"http://exam ple.com\": invalid character \" \" in host name"
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid hostname (missing address)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://:8080",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "could not extract hostname from target address: http://:8080"
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Valid LOG_ADDITIONAL_FIELDS", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:       "http://example.com",
				envLogAdditionalFields: "true",
			}
			return env[key]
		}

		result, err := ParseConfig(mockEnv)
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

	t.Run("Invalid LOG_ADDITIONAL_FIELDS (not boolean)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:       "http://example.com",
				envLogAdditionalFields: "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envLogAdditionalFields)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid check type (invalid)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envCheckType:     "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unsupported check type: invalid"
		if err.Error() != expected {
			t.Errorf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid check type (infer invalid)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "htp://example.com",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "could not infer check type for address htp://example.com: unsupported scheme: htp"
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Missing target address", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return ""
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("%s environment variable is required", envTargetAddress)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})
}
