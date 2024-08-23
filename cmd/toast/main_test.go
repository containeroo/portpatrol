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
	t.Parallel()

	t.Run("HTTP Target is ready", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			"TARGET_ADDRESS": "http://localhost:8081",
			"INTERVAL":       "1s",
			"DIAL_TIMEOUT":   "1s",
			"CHECK_TYPE":     "http",
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
			"TARGET_ADDRESS": "localhost:8082",
			"INTERVAL":       "1s",
			"DIAL_TIMEOUT":   "1s",
			"CHECK_TYPE":     "tcp",
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

		var output bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := run(ctx, getenv, &output)
		if err == nil {
			t.Fatalf("Expected configuration error, got none")
		}

		if !strings.Contains(err.Error(), "configuration error: TARGET_ADDRESS environment variable is required") {
			t.Errorf("Expected configuration error, got %q", err)
		}
	})

	t.Run("Invalid check type", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			"TARGET_NAME":    "TestService",
			"TARGET_ADDRESS": "localhost:8080",
			"INTERVAL":       "1s",
			"DIAL_TIMEOUT":   "1s",
			"CHECK_TYPE":     "invalid",
		}

		getenv := func(key string) string {
			return env[key]
		}

		var output bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := run(ctx, getenv, &output)
		if err == nil {
			t.Error("Expected error, got none")
		}
		expected := "unsupported check type: invalid"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
		}
	})
}
