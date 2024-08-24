package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	const (
		envTargetName          = "TARGET_NAME"
		envTargetAddress       = "TARGET_ADDRESS"
		envInterval            = "INTERVAL"
		envDialTimeout         = "DIAL_TIMEOUT"
		envCheckType           = "CHECK_TYPE"
		envLogAdditionalFields = "LOG_ADDITIONAL_FIELDS"
		envHeaders             = "HEADERS"
	)

	t.Parallel()

	t.Run("HTTP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetAddress: "http://localhost:8081",
			envInterval:      "1s",
			envDialTimeout:   "1s",
			envCheckType:     "http",
		}

		getenv := func(key string) string {
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

		err := run(ctx, getenv, &output)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := "localhost is ready ✓"
		if !strings.Contains(output.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, output.String())
		}
	})

	t.Run("TCP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetAddress: "localhost:8082",
			envInterval:      "1s",
			envDialTimeout:   "1s",
		}

		getenv := func(key string) string {
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

		err = run(ctx, getenv, &output)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := "localhost is ready ✓"
		if !strings.Contains(output.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, output.String())
		}
	})

	t.Run("Config error: variable is required", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{}

		getenv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, getenv, &output)
		if err == nil {
			t.Fatalf("Expected configuration error, got none")
		}

		if !strings.Contains(err.Error(), "configuration error: TARGET_ADDRESS environment variable is required") {
			t.Errorf("Expected configuration error, got %q", err)
		}
	})

	t.Run("Config error: unsupported check type", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetName:    "TestService",
			envTargetAddress: "localhost:8080",
			envInterval:      "1s",
			envDialTimeout:   "1s",
			envCheckType:     "invalid",
		}

		getenv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, getenv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}

		expected := "configuration error: unsupported check type: invalid"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})

	t.Run("Inizalize error: invalid headers", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			envTargetName:    "TestService",
			envTargetAddress: "http://localhost:8080",
			envInterval:      "1s",
			envDialTimeout:   "1s",
			envHeaders:       "Authorization Bearer token",
		}

		getenv := func(key string) string {
			return env[key]
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var output bytes.Buffer

		err := run(ctx, getenv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}

		expected := "failed to initialize checker: invalid HEADERS value: invalid header format: Authorization Bearer token"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})
}
