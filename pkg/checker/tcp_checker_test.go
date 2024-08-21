package checker

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPChecker(t *testing.T) {
	t.Parallel()

	t.Run("valid TCP check", func(t *testing.T) {
		t.Parallel()

		// Start a test TCP server
		ln, err := net.Listen("tcp", "127.0.0.1:7080")
		if err != nil {
			t.Fatalf("failed to start TCP server: %v", err)
		}
		defer ln.Close()

		// Mock environment variables (if any needed in the future)
		mockEnv := func(s string) string {
			return ""
		}

		// Create the TCP checker using the mock environment variables
		checker, err := NewTCPChecker(context.Background(), "example", ln.Addr().String(), 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create TCPChecker: %v", err)
		}

		// Perform the check
		err = checker.Check(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("failed TCP check", func(t *testing.T) {
		t.Parallel()

		// Mock environment variables (if any needed in the future)
		mockEnv := func(s string) string {
			return ""
		}

		// Create the TCP checker using the mock environment variables
		checker, err := NewTCPChecker(context.Background(), "example", "192.0.2.0:7090", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create TCPChecker: %v", err)
		}

		// Perform the check, expecting an error due to a non-existent server
		err = checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
