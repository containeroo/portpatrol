package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP check with default configuration", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("HTTP check with custom headers", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPHeaders(map[string]string{"Authorization": "Bearer token"}))
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("HTTP check with unexpected status code", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unexpected status code: got 404, expected one of [200]"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Invalid URL for HTTP check", func(t *testing.T) {
		t.Parallel()

		checker, err := newHTTPChecker("example", "://invalid-url")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = checker.Check(context.Background()) // Run the check to trigger the error.
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to create request: parse \"://invalid-url\": missing protocol scheme"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Timeout during HTTP check", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second) // Simulate delay
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPTimeout(1*time.Second))
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "context deadline exceeded"
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("expected error containing %q, got %q", expected, err.Error())
		}
	})

	t.Run("Custom expected status codes", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithExpectedStatusCodes([]int{202}))
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Custom HTTP method", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPMethod(http.MethodPost))
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Skip TLS verification", func(t *testing.T) {
		t.Parallel()

		// Create a test server with a self-signed certificate
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPSkipTLSVerify(true))
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})
}
