package checker

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid TCP check", func(t *testing.T) {
		t.Parallel()

		// Start a test TCP server
		ln, err := net.Listen("tcp", "127.0.0.1:7080")
		if err != nil {
			t.Fatalf("failed to start TCP server: %q", err)
		}
		defer ln.Close()

		mockEnv := func(s string) string {
			return ""
		}

		// Create the TCP checker using the mock environment variables
		checker, err := NewTCPChecker("example", ln.Addr().String(), 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create TCPChecker: %q", err)
		}

		// Perform the check
		err = checker.Check(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Failed TCP check", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(s string) string {
			return ""
		}

		// Create the TCP checker using the mock environment variables
		checker, err := NewTCPChecker("example", "localhost:7090", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create TCPChecker: %q", err)
		}

		// Perform the check, expecting an error due to a non-existent server
		err = checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "dial tcp [::1]:7090: connect: connection refused"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})
}
