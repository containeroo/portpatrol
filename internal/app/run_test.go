package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/never/internal/testutils"
)

// fake version for testing
const version string = "0.0.0"

// TestRunHTTPReady verifies the expected behavior.
func TestRunHTTPReady(t *testing.T) {
	t.Parallel()

	userAgent := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent <- r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=" + server.URL,
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",
	}

	var stdOut, stdErr strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := Run(ctx, version, args, &stdOut, &stdErr)
	require.NoError(t, err)

	stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
	last := len(stdOutEntries) - 1

	assert.Contains(t, stdOutEntries[last], "HTTPServer is ready ✓")
	assert.Equal(t, "never/"+version, <-userAgent)
}

// TestRunTCPReady verifies the expected behavior.
func TestRunTCPReady(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	args := []string{
		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=" + listener.Addr().String(),
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr strings.Builder

	err := Run(ctx, version, args, &stdOut, &stdErr)
	require.NoError(t, err)

	stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
	last := len(stdOutEntries) - 1

	assert.Contains(t, stdOutEntries[last], "TCPServer is ready ✓")
}

// TestRunConfigErrorMissingTarget verifies the expected behavior.
func TestRunConfigErrorMissingTarget(t *testing.T) {
	t.Parallel()

	args := []string{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "no checkers to run")
}

// TestRunConfigErrorUnsupportedCheckType verifies the expected behavior.
func TestRunConfigErrorUnsupportedCheckType(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.unsupported.name=TestService",
		"--target.unsupported.address=localhost:8080",
		"--target.unsupported.interval=1s",
		"--target.unsupported.timeout=1s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "unknown dynamic group \"target\" in flag --target.unsupported.name=TestService\nunknown dynamic group \"target\" in flag --target.unsupported.address=localhost:8080\nunknown dynamic group \"target\" in flag --target.unsupported.interval=1s\nunknown dynamic group \"target\" in flag --target.unsupported.timeout=1s")
}

// TestRunConfigErrorInvalidHeaders verifies the expected behavior.
func TestRunConfigErrorInvalidHeaders(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--http.invalidheaders.address=http://localhost:8080",
		"--http.invalidheaders.interval=1s",
		"--http.invalidheaders.timeout=1s",
		"--http.invalidheaders.header=InvalidHeader",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, `HTTP target "invalidheaders": invalid HTTP header: invalid header format: InvalidHeader`)
}

// TestRunParseError verifies the expected behavior.
func TestRunParseError(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--invalid",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "unknown flag --invalid")
}

func TestRunCanceledIsNotReady(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, stderr bytes.Buffer
	err := Run(ctx, version, []string{"--tcp.test.address=127.0.0.1:1"}, &out, &stderr)
	require.ErrorIs(t, err, context.Canceled)
	assert.NotContains(t, out.String(), "is ready")
}

func TestURLPrivacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()

	address := strings.Replace(server.URL, "http://", "http://user:password@", 1) + "/private?token=secret#fragment"
	for _, format := range []string{"json", "text"} {
		for _, tt := range []struct {
			detail       string
			wantPath     bool
			wantQuery    bool
			wantUserInfo bool
			wantFragment bool
		}{
			{detail: "origin"},
			{detail: "path", wantPath: true},
			{detail: "query", wantPath: true, wantQuery: true},
			{detail: "full", wantPath: true, wantQuery: true, wantUserInfo: true, wantFragment: true},
		} {
			t.Run(fmt.Sprintf("%s/%s", format, tt.detail), func(t *testing.T) {
				args := []string{
					"--http.test.address=" + address,
					"--http.test.timeout=10ms",
					"--max-attempts=1",
					"--log-format=" + format,
					"--http-address-detail=" + tt.detail,
				}

				var out, stderr bytes.Buffer
				require.Error(t, Run(context.Background(), version, args, &out, &stderr))

				logOutput := out.String()
				assert.Equal(t, tt.wantPath, strings.Contains(logOutput, "/private"), logOutput)
				assert.Equal(t, tt.wantQuery, strings.Contains(logOutput, "token=secret"), logOutput)
				assert.Equal(t, tt.wantUserInfo, strings.Contains(logOutput, "user:password@"), logOutput)
				assert.Equal(t, tt.wantFragment, strings.Contains(logOutput, "#fragment"), logOutput)
				assert.Contains(t, logOutput, strings.TrimPrefix(server.URL, "http://"))
			})
		}
	}
}
