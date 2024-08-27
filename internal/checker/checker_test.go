package checker

import (
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker("http", "example", "http://example.com", 5*time.Second, func(s string) string {
			return ""
		})
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := "example"
		if check.String() != expected {
			t.Fatalf("expected name to be %q got %q", expected, check.String())
		}
	})

	t.Run("Valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker("tcp", "example", "example.com:80", 5*time.Second, func(s string) string {
			return ""
		})
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := "example"
		if check.String() != expected {
			t.Fatalf("expected name to be %q got %q", expected, check.String())
		}
	})

	t.Run("Invalid checker type", func(t *testing.T) {
		t.Parallel()

		_, err := NewChecker("invalid", "example", "example.com", 5*time.Second, func(s string) string {
			return ""
		})
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unknown check type: invalid"
		if err.Error() != expected {
			t.Errorf("expected error to be %q, got %q", expected, err.Error())
		}
	})
}

func TestIsValidCheckType(t *testing.T) {
	t.Parallel()

	t.Run("Valid TCP Check Type", func(t *testing.T) {
		t.Parallel()

		if isValid := IsValidCheckType("tcp"); !isValid {
			t.Errorf("expected true for check type 'tcp', got false")
		}
	})

	t.Run("Valid HTTP Check Type", func(t *testing.T) {
		t.Parallel()

		if isValid := IsValidCheckType("http"); !isValid {
			t.Errorf("expected true for check type 'http', got false")
		}
	})

	t.Run("Invalid Check Type", func(t *testing.T) {
		t.Parallel()

		if isValid := IsValidCheckType("invalid"); isValid {
			t.Errorf("expected false for check type 'invalid', got true")
		}
	})

	t.Run("Empty Check Type", func(t *testing.T) {
		t.Parallel()

		if isValid := IsValidCheckType(""); isValid {
			t.Errorf("expected false for empty check type, got true")
		}
	})

	t.Run("Random String Check Type", func(t *testing.T) {
		t.Parallel()

		if isValid := IsValidCheckType("random"); isValid {
			t.Errorf("expected false for check type 'random', got true")
		}
	})
}

func TestInferCheckType(t *testing.T) {
	t.Parallel()

	t.Run("HTTP scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("http://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if checkType != "http" {
			t.Fatalf("expected 'http', got %q", checkType)
		}
	})

	t.Run("TCP scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("tcp://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if checkType != "tcp" {
			t.Fatalf("expected 'tcp', got %q", checkType)
		}
	})

	t.Run("No scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("example.com:80")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if checkType != "" {
			t.Fatalf("expected 'tcp', got %q", checkType)
		}
	})

	t.Run("Unsupported scheme", func(t *testing.T) {
		t.Parallel()

		_, err := InferCheckType("ftp://example.com")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unsupported scheme: ftp"
		if err.Error() != expected {
			t.Errorf("expected error to be %q, got %q", expected, err.Error())
		}
	})
}
