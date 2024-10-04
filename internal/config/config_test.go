package config

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	t.Run("Valid config with defaults", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "example.com:80",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "example.com", // Extracted from TargetAddress
			TargetAddress:   "example.com:80",
			TargetCheckType: checker.TCP,
			CheckInterval:   2 * time.Second,
			DialTimeout:     1 * time.Second,
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with www as scheme", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:   "www.example.com:80",
				envTargetCheckType: "http",
				envCheckInterval:   "5s",
				envDialTimeout:     "10s",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "www.example.com", // Extracted from TargetAddress
			TargetAddress:   "www.example.com:80",
			TargetCheckType: checker.HTTP,
			CheckInterval:   5 * time.Second,
			DialTimeout:     10 * time.Second,
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with kubernetes service", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:   "http://postgres.postgres.svc.cluster.local:80",
				envTargetCheckType: "http",
				envCheckInterval:   "5s",
				envDialTimeout:     "10s",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "postgres.postgres.svc.cluster.local", // Extracted from TargetAddress
			TargetAddress:   "http://postgres.postgres.svc.cluster.local:80",
			TargetCheckType: checker.HTTP,
			CheckInterval:   5 * time.Second,
			DialTimeout:     10 * time.Second,
		}
		if !reflect.DeepEqual(cfg, expected) {
			t.Fatalf("expected config %+v, got %+v", expected, cfg)
		}
	})

	t.Run("Valid config with custom values", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:   "tcp://example.com:80",
				envTargetCheckType: "tcp",
				envCheckInterval:   "5s",
				envDialTimeout:     "10s",
			}
			return env[key]
		}

		cfg, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "example.com", // Extracted from TargetAddress
			TargetAddress:   "tcp://example.com:80",
			TargetCheckType: checker.TCP,
			CheckInterval:   5 * time.Second,
			DialTimeout:     10 * time.Second,
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
				envCheckInterval: "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envCheckInterval)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Invalid interval (zero)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "http://example.com",
				envCheckInterval: "0s",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: 0s", envCheckInterval)
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

	t.Run("Valid LOG_EXTRA_FIELDS", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:  "http://example.com",
				envLogExtraFields: "true",
			}
			return env[key]
		}

		result, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "example.com",
			TargetAddress:   "http://example.com",
			TargetCheckType: checker.HTTP,
			CheckInterval:   2 * time.Second,
			DialTimeout:     1 * time.Second,
			LogExtraFields:  true,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Invalid LOG_EXTRA_FIELDS (not boolean)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:  "http://example.com",
				envLogExtraFields: "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envLogExtraFields)
		if err.Error() != expected {
			t.Fatalf("expected error to contain %q, got %q", expected, err)
		}
	})

	t.Run("Valid check type (defaults to tcp)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress: "example.com:80",
			}
			return env[key]
		}

		result, err := ParseConfig(mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := Config{
			TargetName:      "example.com",
			TargetAddress:   "example.com:80",
			TargetCheckType: checker.TCP,
			CheckInterval:   2 * time.Second,
			DialTimeout:     1 * time.Second,
			LogExtraFields:  false,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Invalid check type (invalid)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envTargetAddress:   "http://example.com",
				envTargetCheckType: "invalid",
			}
			return env[key]
		}

		_, err := ParseConfig(mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid check type from environment: unsupported check type: invalid"
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

		expected := "could not infer check type from address htp://example.com: unsupported check type: htp"
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
