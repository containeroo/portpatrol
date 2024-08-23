package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP check", func(t *testing.T) {
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
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		// Perform the check
		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Unexpected status code", func(t *testing.T) {
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
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		// Perform the check
		err = checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "unexpected status code: got 404, expected one of [200]"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Cancel HTTP check", func(t *testing.T) {
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
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		// Cancel the context after a very short time
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Perform the check, expecting a context canceled error
		err = checker.Check(ctx)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := "context deadline exceeded"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Header parsing error", func(t *testing.T) {
		t.Parallel()

		// Mock environment variables with an invalid header
		mockEnv := func(key string) string {
			env := map[string]string{
				envMethod:           "GET",
				envHeaders:          "Authorization Bearer token", // Missing '=' in the header
				envExpectedStatuses: "200",
			}
			return env[key]
		}

		// Attempt to create the HTTP checker using the mock environment variables
		_, err := NewHTTPChecker("example", "http://example.com", 1*time.Second, mockEnv)
		if err == nil {
			t.Errorf("expected an error, got none")
		}

		expected := "invalid HEADERS value: invalid header format: Authorization Bearer token"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}

func TestParseHeaders(t *testing.T) {
	t.Parallel()

	t.Run("Valid headers", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Authorization=Bearer token"
		result, err := parseHeaders(headers)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Single header", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json"
		result, err := parseHeaders(headers)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"Content-Type": "application/json"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Empty headers string", func(t *testing.T) {
		t.Parallel()

		headers := ""
		result, err := parseHeaders(headers)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Malformed header (missing =)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,AuthorizationBearer token"
		_, err := parseHeaders(headers)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		expected := "invalid header format: AuthorizationBearer token"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Header with spaces", func(t *testing.T) {
		t.Parallel()

		headers := "  Content-Type = application/json  , Authorization = Bearer token  "
		result, err := parseHeaders(headers)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Header with empty key", func(t *testing.T) {
		t.Parallel()

		headers := "=value"
		_, err := parseHeaders(headers)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		expected := "header key cannot be empty: =value"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Header with empty value", func(t *testing.T) {
		t.Parallel()

		headers := "key="
		result, err := parseHeaders(headers)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"key": ""}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Malformed header (empty pair)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,"
		_, err := parseHeaders(headers)
		if err != nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestParseExpectedStatuses(t *testing.T) {
	t.Parallel()

	t.Run("Single status code", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Multiple status codes", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200,404,500")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200, 404, 500}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200-202")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200, 201, 202}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Multipl status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200-202,300-301,500")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200, 201, 202, 300, 301, 500}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Invalid status code", func(t *testing.T) {
		t.Parallel()

		_, err := parseExpectedStatuses("abc")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status code: abc"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid status range", func(t *testing.T) {
		t.Parallel()

		_, err := parseExpectedStatuses("202-200")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status range: 202-200"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}
