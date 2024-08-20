package main

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/containeroo/toast/pkg/config"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("WithAdditionalFields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		cfg := config.Config{
			TargetAddress:       "localhost:8080",
			Interval:            1 * time.Second,
			DialTimeout:         2 * time.Second,
			CheckType:           "http",
			LogAdditionalFields: true,
		}

		logger := setupLogger(cfg, &buf)
		logger.Info("Test log")

		logOutput := buf.String()

		if !strings.Contains(logOutput, "target_address=localhost:8080") ||
			!strings.Contains(logOutput, "interval=1s") ||
			!strings.Contains(logOutput, "dial_timeout=2s") ||
			!strings.Contains(logOutput, "checker_type=http") ||
			!strings.Contains(logOutput, "version=0.0.1") {
			t.Errorf("Logger output does not contain expected fields: %s", logOutput)
		}
	})

	t.Run("WithoutAdditionalFields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		cfg := config.Config{
			LogAdditionalFields: false,
		}

		logger := setupLogger(cfg, &buf)
		logger.Error("Test error", slog.String("error", "some error"))

		logOutput := buf.String()

		expected := "error=some error"
		if strings.Contains(logOutput, expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, logOutput)
		}
	})
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulHTTPChecker", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			"TARGET_NAME":    "TestHTTP",
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

		var output bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := run(ctx, getenv, &output)
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Expected no error or context deadline exceeded, got %v", err)
		}

		if !strings.Contains(output.String(), "is ready ✓") && ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected success message in log output, got: %s", output.String())
		}

		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("Failed to start HTTP server: %v", err)
		}
	})

	t.Run("SuccessfulTCPChecker", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{
			"TARGET_NAME":    "TestTCP",
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
			t.Fatalf("Failed to start TCP server: %v", err)
		}
		defer listener.Close()

		var output bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err = run(ctx, getenv, &output)
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Expected no error or context deadline exceeded, got %v", err)
		}

		if !strings.Contains(output.String(), "is ready ✓") && ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected success message in log output, got: %s", output.String())
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		t.Parallel()

		env := map[string]string{}

		getenv := func(key string) string {
			return env[key]
		}

		var output bytes.Buffer
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := run(ctx, getenv, &output)
		if err == nil || !strings.Contains(err.Error(), "configuration error") {
			t.Errorf("Expected configuration error, got %v", err)
		}
	})

	t.Run("CheckerInitializationError", func(t *testing.T) {
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

// Test signal handling in main function (indirect test)
func TestMainFunction_SignalHandling(t *testing.T) {
	t.Parallel()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(done)
	}()

	// Simulate sending a SIGTERM signal
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM) // Make linter happy

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Errorf("Expected context to be canceled, but it was not")
	}
}
