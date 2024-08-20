package checker

import (
	"testing"
)

func TestExtractScheme(t *testing.T) {
	t.Parallel()

	t.Run("valid address with scheme", func(t *testing.T) {
		t.Parallel()
		scheme, err := extractScheme("http://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if scheme != "http" {
			t.Fatalf("expected scheme 'http', got %v", scheme)
		}
	})

	t.Run("valid address with another scheme", func(t *testing.T) {
		t.Parallel()
		scheme, err := extractScheme("https://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if scheme != "https" {
			t.Fatalf("expected scheme 'https', got %v", scheme)
		}
	})

	t.Run("invalid address without scheme", func(t *testing.T) {
		t.Parallel()
		_, err := extractScheme("example.com")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("address with scheme only", func(t *testing.T) {
		t.Parallel()
		scheme, err := extractScheme("ftp://")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if scheme != "ftp" {
			t.Fatalf("expected scheme 'ftp', got %v", scheme)
		}
	})

	t.Run("invalid address with missing colon", func(t *testing.T) {
		t.Parallel()
		_, err := extractScheme("http//example.com")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
