package wait

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	c, err := checker.DefaultHTTPConfig().NewChecker(httpServerName, server.URL)
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, WaitUntilReady(ctx, 100*time.Millisecond, -1, c, logger))
	assert.Contains(t, output.String(), "HTTPServer is ready ✓")
}

// TestWaitUntilReady_HTTPFailsInitially tests HTTP target readiness after initial failures.
func TestWaitUntilReady_HTTPFailsInitially(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := checker.DefaultHTTPConfig().NewChecker(httpServerName, server.URL)
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, WaitUntilReady(ctx, 100*time.Millisecond, -1, c, logger))
	assert.Contains(t, output.String(), "HTTPServer is ready ✓")
}

// TestWaitUntilReady_HTTPContextCanceled tests behavior when the context is canceled.
func TestWaitUntilReady_HTTPContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := checker.DefaultHTTPConfig().NewChecker(httpServerName, server.URL)
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, c, logger)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, output.String(), "Waiting for HTTPServer to become ready...")
}

// TestWaitUntilReady_ReadyTCP ensures WaitUntilReady succeeds for a ready TCP target.
func TestWaitUntilReady_ReadyTCP(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	c, err := checker.DefaultTCPConfig().NewChecker(tcpServerName, listener.Addr().String())
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, WaitUntilReady(ctx, 100*time.Millisecond, -1, c, logger))
	assert.Contains(t, output.String(), "TCPServer is ready ✓")
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
		started <- listenResult{listener: listener, err: err}
	}()
	t.Cleanup(func() {
		result := <-started
		if assert.NoError(t, result.err) && assert.NotNil(t, result.listener) {
			assert.NoError(t, result.listener.Close())
		}
	})

	c, err := checker.DefaultTCPConfig().NewChecker(tcpServerName, addr)
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, WaitUntilReady(ctx, 100*time.Millisecond, -1, c, logger))
	assert.Contains(t, output.String(), "TCPServer is ready ✓")
}

// TestWaitUntilReady_TCPContextCanceled tests behavior when the TCP target's context is canceled.
func TestWaitUntilReady_TCPContextCanceled(t *testing.T) {
	t.Parallel()

	c, err := checker.DefaultTCPConfig().NewChecker(tcpServerName, testutils.LocalTCPAddr(t))
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, c, logger)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, output.String(), "Waiting for TCPServer to become ready...")
}

// TestNewStoppedTimer verifies the expected behavior.
func TestNewStoppedTimer(t *testing.T) {
	t.Parallel()

	timer := newStoppedTimer(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		assert.Fail(t, "timer fired while stopped")
	default:
	}

	timer.Reset(5 * time.Millisecond)
	select {
	case <-timer.C:
	case <-time.After(50 * time.Millisecond):
		require.FailNow(t, "timer did not fire after reset")
	}
}

// TestWaitUntilReady_MaxAttempts verifies the expected behavior.
func TestWaitUntilReady_MaxAttempts(t *testing.T) {
	t.Parallel()

	c, err := checker.DefaultTCPConfig().NewChecker(tcpServerName, testutils.LocalTCPAddr(t))
	require.NoError(t, err)

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 10*time.Millisecond, 2, c, logger)
	require.ErrorIs(t, err, ErrMaxAttemptsExceeded)
}

// TestWaitUntilReady_ContextCanceledDuringCheckStopsGracefully verifies the expected behavior.
func TestWaitUntilReady_ContextCanceledDuringCheckStopsGracefully(t *testing.T) {
	t.Parallel()

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	err := WaitUntilReady(context.Background(), 10*time.Millisecond, -1, staticErrorChecker{err: context.Canceled}, logger)
	require.ErrorIs(t, err, context.Canceled)
	assert.NotContains(t, output.String(), "is not ready")
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
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()

	protocolConfig := checker.DefaultHTTPConfig()
	protocolConfig.Timeout = 20 * time.Millisecond
	c, err := protocolConfig.NewChecker("slow", server.URL)
	require.NoError(t, err)

	var output strings.Builder
	err = WaitUntilReady(context.Background(), time.Millisecond, 3, c, slog.New(slog.NewTextHandler(&output, nil)))
	require.ErrorIs(t, err, ErrMaxAttemptsExceeded)
	assert.Contains(t, output.String(), "attempt=3")
}
