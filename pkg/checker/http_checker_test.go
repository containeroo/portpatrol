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

	t.Run("valid HTTP check", func(t *testing.T) {
		t.Parallel()

		// Set up a test HTTP server
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		// Mock environment variables
		mockEnv := func(key string) string {
			env := map[string]string{
				envMethod:           "GET",
				envHeaders:          "Authorization=Bearer token",
				envExpectedStatuses: "200",
			}
			return env[key]
		}

		// Create the HTTP checker using the mock environment variables
		checker, err := NewHTTPChecker("example", server.URL, 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %v", err)
		}

		// Perform the check
		err = checker.Check(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unexpected status code", func(t *testing.T) {
		t.Parallel()

		// Set up a test HTTP server with a different status code
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		// Mock environment variables
		mockEnv := func(key string) string {
			env := map[string]string{
				envMethod:           "GET",
				envHeaders:          "Authorization=Bearer token",
				envExpectedStatuses: "200",
			}
			return env[key]
		}

		// Create the HTTP checker using the mock environment variables
		checker, err := NewHTTPChecker("example", server.URL, 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %v", err)
		}

		// Perform the check, expecting an error due to the unexpected status code
		err = checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("Test cancel HTTP check", func(t *testing.T) {
		t.Parallel()

		// Set up a test HTTP server that deliberately delays the response
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second) // Delay to ensure the context has time to be canceled
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		// Mock environment variables
		mockEnv := func(key string) string {
			env := map[string]string{
				envMethod:           "GET",
				envHeaders:          "Authorization=Bearer token",
				envExpectedStatuses: "200",
			}
			return env[key]
		}

		// Create the HTTP checker using the mock environment variables
		checker, err := NewHTTPChecker("example", server.URL, 5*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %v", err)
		}

		// Cancel the context after a very short time
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Perform the check, expecting a context canceled error
		err = checker.Check(ctx)
		if err == nil || !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
			t.Errorf("expected context canceled or deadline exceeded error, got %v", err)
		}
	})
}
