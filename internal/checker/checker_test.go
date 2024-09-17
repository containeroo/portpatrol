package checker

import (
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(HTTP, "example", "http://example.com", 5*time.Second, func(s string) string {
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

		check, err := NewChecker(TCP, "example", "example.com:80", 5*time.Second, func(s string) string {
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

	t.Run("Valid ICMP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(ICMP, "example", "example.com", 5*time.Second, func(s string) string {
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

		_, err := NewChecker(8, "example", "example.com", 5*time.Second, func(s string) string {
			return ""
		})
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unsupported check type: 8"
		if err.Error() != expected {
			t.Errorf("expected error to be %q, got %q", expected, err.Error())
		}
	})
}

func TestGetCheckTypeString(t *testing.T) {
	t.Parallel()

	t.Run("Check type string (enum)", func(t *testing.T) {
		t.Parallel()

		if HTTP.String() != "HTTP" {
			t.Fatalf("expected 'HTTP', got %q", HTTP.String())
		}
		if TCP.String() != "TCP" {
			t.Fatalf("expected 'TCP', got %q", TCP.String())
		}
		if ICMP.String() != "ICMP" {
			t.Fatalf("expected 'ICMP', got %q", ICMP.String())
		}
	})

	t.Run("Check type string (func)", func(t *testing.T) {
		want := HTTP
		got, err := GetCheckTypeFromString("http")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		want = TCP
		got, err = GetCheckTypeFromString("tcp")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		want = ICMP
		got, err = GetCheckTypeFromString("icmp")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		want = -1
		got, err = GetCheckTypeFromString("invalid")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
