package wait

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/testutils"
)

const (
	httpServerName = "HTTPServer"
	tcpServerName  = "TCPServer"
)

// TestWaitUntilReady_ReadyHTTP ensures WaitUntilReady returns success when the HTTP target is ready.
func TestWaitUntilReady_ReadyHTTP(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker, err := checker.NewHTTPChecker(httpServerName, server.URL, checker.DefaultHTTPConfig())
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "HTTPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_HTTPFailsInitially tests HTTP target readiness after initial failures.
func TestWaitUntilReady_HTTPFailsInitially(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // Simulate a delayed start
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker, err := checker.NewHTTPChecker(httpServerName, server.URL, checker.DefaultHTTPConfig())
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "HTTPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_HTTPContextCanceled tests behavior when the context is canceled.
func TestWaitUntilReady_HTTPContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker, err := checker.NewHTTPChecker(httpServerName, server.URL, checker.DefaultHTTPConfig())
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, checker, logger)
	if err == nil {
		t.Fatalf("Expected context cancellation error, got nil")
	}

	expectedLog := "Waiting for HTTPServer to become ready..."
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_ReadyTCP ensures WaitUntilReady succeeds for a ready TCP target.
func TestWaitUntilReady_ReadyTCP(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	checker, err := checker.NewTCPChecker(tcpServerName, listener.Addr().String(), checker.DefaultTCPConfig())
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "TCPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_TCPFailsInitially tests TCP readiness after initial failures.
func TestWaitUntilReady_TCPFailsInitially(t *testing.T) {
	t.Parallel()

	addr := testutils.LocalTCPAddr(t)

	type listenResult struct {
		listener net.Listener
		err      error
	}
	started := make(chan listenResult, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		listener, err := net.Listen("tcp", addr)
		started <- listenResult{listener, err}
	}()
	defer func() {
		result := <-started
		if result.err != nil {
			t.Errorf("failed to start TCP server: %v", result.err)
			return
		}
		_ = result.listener.Close()
	}()

	checker, err := checker.NewTCPChecker(tcpServerName, addr, checker.DefaultTCPConfig())
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "TCPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_TCPContextCanceled tests behavior when the TCP target's context is canceled.
func TestWaitUntilReady_TCPContextCanceled(t *testing.T) {
	t.Parallel()

	checker, err := checker.NewTCPChecker(tcpServerName, testutils.LocalTCPAddr(t), checker.DefaultTCPConfig())
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, checker, logger)
	if err == nil {
		t.Fatalf("Expected context cancellation error, got nil")
	}

	expectedLog := "Waiting for TCPServer to become ready..."
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestNewStoppedTimer verifies the expected behavior.
func TestNewStoppedTimer(t *testing.T) {
	t.Parallel()

	timer := newStoppedTimer(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("expected timer to be stopped")
	default:
	}

	timer.Reset(5 * time.Millisecond)
	select {
	case <-timer.C:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected timer to fire after reset")
	}
}

// TestWaitUntilReady_MaxAttempts verifies the expected behavior.
func TestWaitUntilReady_MaxAttempts(t *testing.T) {
	t.Parallel()

	checker, err := checker.NewTCPChecker(tcpServerName, testutils.LocalTCPAddr(t), checker.DefaultTCPConfig())
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 10*time.Millisecond, 2, checker, logger)
	if err == nil {
		t.Fatal("Expected max attempts error, got nil")
	}
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Fatalf("Expected ErrMaxAttemptsExceeded, got %v", err)
	}
}

// TestWaitUntilReady_ContextCanceledDuringCheckStopsGracefully verifies the expected behavior.
func TestWaitUntilReady_ContextCanceledDuringCheckStopsGracefully(t *testing.T) {
	t.Parallel()

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	err := WaitUntilReady(context.Background(), 10*time.Millisecond, -1, staticErrorChecker{err: context.Canceled}, logger)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected cancellation, got %v", err)
	}
	if strings.Contains(output.String(), "is not ready") {
		t.Fatalf("Expected cancellation to avoid not-ready log, got %q", output.String())
	}
}

type staticErrorChecker struct {
	err error
}

// Check performs the checker operation.
func (c staticErrorChecker) Check(context.Context) error { return c.err }

// Name returns the checker name.
func (c staticErrorChecker) Name() string { return "CanceledServer" }

// Type returns the checker type.
func (c staticErrorChecker) Type() string { return "TCP" }

// Address returns the checker address.
func (c staticErrorChecker) Address() string { return testutils.LocalhostAddr("1") }

func TestRequestTimeoutRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	protocolConfig := checker.DefaultHTTPConfig()
	protocolConfig.Timeout = 20 * time.Millisecond
	c, err := checker.NewHTTPChecker("slow", server.URL, protocolConfig)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	err = WaitUntilReady(context.Background(), time.Millisecond, 3, c, slog.New(slog.NewTextHandler(&output, nil)))
	if !errors.Is(err, ErrMaxAttemptsExceeded) || !strings.Contains(output.String(), "attempt=3") {
		t.Fatalf("expected three attempts: %v %s", err, output.String())
	}
}
