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

func TestRunHTTPReady(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.httpcheck.name=HTTPServer",
		"--target.httpcheck.address=http://localhost:8081",
		"--target.httpcheck.interval=1s",
		"--target.httpcheck.timeout=1s",
		"--target.httpcheck.type=http",
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
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}

	outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
	last := len(outputEntries) - 1

	expected := "HTTPServer is ready ✓"
	if !strings.Contains(outputEntries[last], expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, output.String())
	}
}

func TestRunTCPReady(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.tcptest.name=TCPServer",
		"--target.tcptest.address=localhost:8082",
		"--target.tcptest.interval=1s",
		"--target.tcptest.timeout=1s",
		"--target.tcptest.type=tcp",
	}

	listener, err := net.Listen("tcp", "localhost:8082")
	if err != nil {
		t.Fatalf("Failed to start TCP server: %q", err)
	}
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output strings.Builder
	err = run(ctx, args, &output)
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}

	outputEntries := strings.Split(strings.TrimSpace(output.String()), "\n")
	last := len(outputEntries) - 1

	expected := "TCPServer is ready ✓"
	if !strings.Contains(outputEntries[last], expected) {
		t.Errorf("Expected output to contain %q but got %q", expected, output.String())
	}
}

func TestRunConfigErrorMissingTarget(t *testing.T) {
	t.Parallel()

	args := []string{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)
	if err == nil {
		t.Fatalf("Expected configuration error, got none")
	}

	expected := "configuration error: no checkers configured"
	if err.Error() != expected {
		t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestRunConfigErrorUnsupportedCheckType(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.unsupported.name=TestService",
		"--target.unsupported.address=localhost:8080",
		"--target.unsupported.interval=1s",
		"--target.unsupported.timeout=1s",
		"--target.unsupported.type=unsupported",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	expected := "failed to initialize target checkers: unsupported check type \"unsupported\" for target \"unsupported\""
	if err.Error() != expected {
		t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestRunConfigErrorInvalidHeaders(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.invalidheaders.name=TestService",
		"--target.invalidheaders.address=http://localhost:8080",
		"--target.invalidheaders.interval=1s",
		"--target.invalidheaders.timeout=1s",
		"--target.invalidheaders.headers=InvalidHeader",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := run(ctx, args, &output)
	if err == nil {
		t.Fatal("Expected error, got none")
	}

	expected := "failed to initialize target checkers: invalid \"headers\": invalid header format: InvalidHeader"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error to contain %q, got %q", expected, err.Error())
	}
}
