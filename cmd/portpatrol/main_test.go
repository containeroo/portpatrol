package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Parallel()

	const (
		envTargetName          = "TARGET_NAME"
		envTargetAddress       = "TARGET_ADDRESS"
		envTargetCheckType     = "TARGET_CHECK_TYPE"
		envCheckInterval       = "CHECK_INTERVAL"
		envDialTimeout         = "DIAL_TIMEOUT"
		envLogAdditionalFields = "LOG_EXTRA_FIELDS"
		envHTTPHeaders         = "HTTP_HEADERS"
	)

	t.Run("HTTP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetAddress:   "http://localhost:8081",
			envCheckInterval:   "1s",
			envDialTimeout:     "1s",
			envTargetCheckType: "http",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		server := &http.Server{Addr: ":8081"}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		go func() { // make linter happy
			_ = server.ListenAndServe()
		}()
		defer server.Close()

		var output strings.Builder

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// cancel after 2 Seconds
		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		err := run(ctx, mockEnv, &output)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
		last := len(outputEntries) - 1

		expected := "localhost is ready ✓"
		if !strings.Contains(outputEntries[last], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, output.String())
		}
	})

	t.Run("TCP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetAddress: "localhost:8082",
			envCheckInterval: "1s",
			envDialTimeout:   "1s",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		listener, err := net.Listen("tcp", "localhost:8082")
		if err != nil {
			t.Fatalf("Failed to start TCP server: %q", err)
		}
		defer listener.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// cancel after 2 Seconds
		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		var output strings.Builder

		err = run(ctx, mockEnv, &output)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
		last := len(outputEntries) - 1

		expected := "localhost is ready ✓"
		if !strings.Contains(outputEntries[last], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, output.String())
		}
	})

	t.Run("ICMP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetAddress:   "icmp://127.0.0.1",
			envCheckInterval:   "1s",
			envDialTimeout:     "1s",
			envTargetCheckType: "icmp",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		var output strings.Builder

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// cancel after 2 Seconds
		go func() {
			time.Sleep(2 * time.Second)
			cancel()
		}()

		err := run(ctx, mockEnv, &output)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
		last := len(outputEntries) - 1

		expected := "127.0.0.1 is ready ✓"
		if !strings.Contains(outputEntries[last], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, output.String())
		}
	})

	t.Run("Config error: variable is required", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{}

		mockEnv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, mockEnv, &output)
		if err == nil {
			t.Fatalf("Expected configuration error, got none")
		}

		expected := fmt.Sprintf("configuration error: %s environment variable is required", envTargetAddress)
		if err.Error() != expected {
			t.Errorf("Expected configuration error, got %q", err)
		}
	})

	t.Run("Config error: unsupported check type", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetName:      "TestService",
			envTargetAddress:   "localhost:8080",
			envCheckInterval:   "1s",
			envDialTimeout:     "1s",
			envTargetCheckType: "invalid",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, mockEnv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}

		expected := "configuration error: unsupported check type: invalid"
		if err.Error() != expected {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})

	t.Run("Inizalize error: unknown check type", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetName:    "TestService",
			envTargetAddress: "htp://localhost:8080",
			envCheckInterval: "1s",
			envDialTimeout:   "1s",
			envHTTPHeaders:   "Auportpatrolization Bearer token",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, mockEnv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}

		expected := "configuration error: could not infer check type for address htp://localhost:8080: unsupported scheme: htp"
		if err.Error() != expected {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})

	t.Run("Inizalize error: invalid headers", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetName:    "TestService",
			envTargetAddress: "http://localhost:8080",
			envCheckInterval: "1s",
			envDialTimeout:   "1s",
			envHTTPHeaders:   "Auportpatrolization Bearer token",
		}

		mockEnv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, mockEnv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}

		expected := fmt.Sprintf("failed to initialize checker: invalid %s value: invalid header format: Auportpatrolization Bearer token", envHTTPHeaders)
		if err.Error() != expected {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})
}
