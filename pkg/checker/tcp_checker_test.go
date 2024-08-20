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

		ln, err := net.Listen("tcp", "127.0.0.1:7080")
		if err != nil {
			t.Fatalf("failed to start TCP server: %v", err)
		}
		defer ln.Close()

		checker, _ := NewTCPChecker("example", ln.Addr().String(), 1*time.Second, func(s string) string {
			return ""
		})
		err = checker.Check(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("failed TCP check", func(t *testing.T) {
		t.Parallel()

		checker, _ := NewTCPChecker("example", "192.0.2.0:7090", 1*time.Second, func(s string) string {
			return ""
		})
		err := checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
