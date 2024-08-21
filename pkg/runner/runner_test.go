package runner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/toast/pkg/checker"
	"github.com/containeroo/toast/pkg/config"
)

func TestRunLoop_HTTPChecker_Success(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulHTTPCheckerWithPath", func(t *testing.T) {
		t.Parallel()

		server := &http.Server{Addr: ":9081"}
		http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		go func() {
			_ = server.ListenAndServe()
		}()

		defer server.Close()

		cfg := config.Config{
			TargetName:    "HTTPServer",
			TargetAddress: "http://localhost:9081/ping",
			Interval:      50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
			CheckType:     "http",
		}

		// Mock environment variables for HTTPChecker
		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"HEADERS":           "",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %v", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = RunLoop(ctx, cfg, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := "HTTPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("SuccessfulHTTPChecker", func(t *testing.T) {
		t.Parallel()

		server := &http.Server{Addr: ":9082"}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		go func() {
			_ = server.ListenAndServe()
		}()

		defer server.Close()

		cfg := config.Config{
			TargetName:    "HTTPServer",
			TargetAddress: "http://localhost:9082/",
			Interval:      50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
			CheckType:     "http",
		}

		// Mock environment variables for HTTPChecker
		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"HEADERS":           "",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %v", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = RunLoop(ctx, cfg, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := "HTTPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})
}

// Test with TCPChecker and a real TCP server
func TestRunLoop_TCPChecker_Success(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "localhost:7082")
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer listener.Close()

	cfg := config.Config{
		TargetName:    "TCPServer",
		TargetAddress: "localhost:7082",
		Interval:      50 * time.Millisecond,
		DialTimeout:   50 * time.Millisecond,
		CheckType:     "tcp",
	}

	// Mock environment variables for TCPChecker
	mockEnv := func(key string) string {
		return ""
	}

	checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var stdOut strings.Builder
	logger := slog.New(slog.NewTextHandler(&stdOut, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = RunLoop(ctx, cfg, checker, logger)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "TCPServer is ready ✓"
	if !strings.Contains(stdOut.String(), expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
	}
}

// Test HTTPChecker with context cancellation
func TestRunLoop_HTTPChecker_ContextCancel(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		TargetName:    "HTTPServer",
		TargetAddress: "http://localhost:7083/fail",
		Interval:      50 * time.Millisecond,
		DialTimeout:   50 * time.Millisecond,
		CheckType:     "http",
	}

	// Mock environment variables for HTTPChecker
	mockEnv := func(key string) string {
		env := map[string]string{
			"METHOD":            "GET",
			"HEADERS":           "",
			"EXPECTED_STATUSES": "200",
		}
		return env[key]
	}

	checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var stdOut strings.Builder
	logger := slog.New(slog.NewTextHandler(&stdOut, nil))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = RunLoop(ctx, cfg, checker, logger)
	if err != nil && err != context.Canceled {
		t.Errorf("Expected context canceled error, got %v", err)
	}

	expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
	if !strings.Contains(stdOut.String(), expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
	}

	expected = fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
	if !strings.Contains(stdOut.String(), expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
	}
}

// Test TCPChecker with context cancellation
func TestRunLoop_TCPChecker_ContextCancel(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		TargetName:    "TCPServer",
		TargetAddress: "localhost:7084",
		Interval:      50 * time.Millisecond,
		DialTimeout:   50 * time.Millisecond,
		CheckType:     "tcp",
	}

	// Mock environment variables for TCPChecker
	mockEnv := func(key string) string {
		return ""
	}

	checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var stdOut strings.Builder
	logger := slog.New(slog.NewTextHandler(&stdOut, nil))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = RunLoop(ctx, cfg, checker, logger)
	if err != nil && err != context.Canceled {
		t.Errorf("Expected context canceled error, got %v", err)
	}

	expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
	if !strings.Contains(stdOut.String(), expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
	}

	expected = fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
	if !strings.Contains(stdOut.String(), expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
	}
}
