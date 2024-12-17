package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunHTTPReady(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://localhost:8081",
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",
	}

	server := &http.Server{Addr: ":8081"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go func() { _ = server.ListenAndServe() }()
	defer server.Close()

	var output strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := run(ctx, args, &output)
	assert.NoError(t, err)

	outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
	last := len(outputEntries) - 1

	assert.Contains(t, outputEntries[last], "HTTPServer is ready ✓")
}

func TestRunTCPReady(t *testing.T) {
	t.Parallel()

	args := []string{
		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=localhost:8082",
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	listener, err := net.Listen("tcp", "localhost:8082")
	assert.NoError(t, err)
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output strings.Builder
	err = run(ctx, args, &output)
	assert.NoError(t, err)

	outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
	last := len(outputEntries) - 1

	assert.Contains(t, outputEntries[last], "TCPServer is ready ✓")
}

func TestRunConfigErrorMissingTarget(t *testing.T) {
	t.Parallel()

	args := []string{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)

	assert.Error(t, err)
	assert.EqualError(t, err, "configuration error: no checkers configured")
}

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

	var output bytes.Buffer
	err := run(ctx, args, &output)

	assert.Error(t, err)
	assert.EqualError(t, err, "configuration error: no checkers configured")
}

func TestRunConfigErrorInvalidHeaders(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--http.invalidheaders.address=http://localhost:8080",
		"--http.invalidheaders.interval=1s",
		"--http.invalidheaders.timeout=1s",
		"--http.invalidheaders.headers=InvalidHeader",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)

	assert.Error(t, err)
	assert.EqualError(t, err, "failed to initialize target checkers: invalid \"--http.invalidheaders.headers\": invalid header format: InvalidHeader")
}

func TestRunParseError(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--invalid",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)

	assert.Error(t, err)
	assert.EqualError(t, err, "configuration error: Flag parsing error: unknown flag: --invalid")
}

func TestRunShowVersion(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--http.invalidheaders.address=http://localhost:8080",
		"--version",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)

	assert.NoError(t, err)
}
