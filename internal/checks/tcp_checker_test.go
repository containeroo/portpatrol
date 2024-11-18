package checks

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewTCPChecker_Valid(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7080")
	if err != nil {
		t.Fatalf("failed to start TCP server: %q", err)
	}
	defer ln.Close()

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("failed to create TCPChecker: %q", err)
	}

	if checker.GetName() != "example" {
		t.Errorf("expected name to be 'example', got %q", checker.GetName())
	}

	if checker.GetAddress() != ln.Addr().String() {
		t.Errorf("expected address to be %q, got %q", ln.Addr().String(), checker.GetAddress())
	}

	if checker.GetType() != TCP.String() {
		t.Errorf("expected type to be %q, got %q", TCP.String(), checker.GetType())
	}
}

func TestTCPChecker_ValidConnection(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7081")
	if err != nil {
		t.Fatalf("failed to start TCP server: %q", err)
	}
	defer ln.Close()

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("failed to create TCPChecker: %q", err)
	}

	ctx := context.Background()
	err = checker.Check(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}
}

func TestTCPChecker_FailedConnection(t *testing.T) {
	t.Parallel()

	checker, err := newTCPChecker("example", "127.0.0.1:7090", WithTCPTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("failed to create TCPChecker: %q", err)
	}

	ctx := context.Background()
	err = checker.Check(ctx)
	if err == nil {
		t.Fatal("expected an error, got none")
	}

	expected := "dial tcp 127.0.0.1:7090: connect: connection refused"
	if err.Error() != expected {
		t.Errorf("expected error to be %q, got %q", expected, err.Error())
	}
}

func TestTCPChecker_InvalidAddress(t *testing.T) {
	t.Parallel()

	checker, err := newTCPChecker("example", "invalid-address", WithTCPTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("failed to create TCPChecker: %q", err)
	}

	ctx := context.Background()
	err = checker.Check(ctx)
	if err == nil {
		t.Fatal("expected an error, got none")
	}

	expected := "dial tcp: address invalid-address: missing port in address"
	if err.Error() != expected {
		t.Errorf("expected error to be %q, got %q", expected, err.Error())
	}
}

func TestTCPChecker_Timeout(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7082")
	if err != nil {
		t.Fatalf("failed to start TCP server: %q", err)
	}
	defer ln.Close()

	// Simulate a timeout by setting an impossibly short timeout
	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Nanosecond))
	if err != nil {
		t.Fatalf("failed to create TCPChecker: %q", err)
	}

	ctx := context.Background()
	err = checker.Check(ctx)
	if err == nil {
		t.Fatal("expected an error, got none")
	}

	expected := "dial tcp 127.0.0.1:7082: i/o timeout"
	if err.Error() != expected {
		t.Errorf("expected error to be %q, got %q", expected, err.Error())
	}
}
