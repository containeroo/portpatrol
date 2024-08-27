package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP check", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Auportpatrolization=Bearer token",
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

		checker, err := NewHTTPChecker("example", server.URL, 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Unexpected status code", func(t *testing.T) {
		t.Parallel()

		// Set up a test HTTP server with a unexpected status code
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Auportpatrolization=Bearer token",
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

		checker, err := NewHTTPChecker("example", server.URL, 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

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

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second) // Delay to ensure the context has time to be canceled
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Auportpatrolization=Bearer token",
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

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

	t.Run("Invalid HTTP check (malformed URL)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"HEADERS":           "Auportpatrolization=Bearer token",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		// Use a malformed URL to trigger an error in creating the request
		checker, err := NewHTTPChecker("example", "://invalid-url", 5*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := "failed to create request: parse \"://invalid-url\": missing protocol scheme"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Valid HTTP check (duplicate headers)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPAllowDuplicateHeaders: "true",
				envHTTPMethod:                "GET",
				envHTTPHeaders:               "Content-Type=application/json,Content-Type=application/json",
				envHTTPExpectedStatusCodes:   "200",
			}
			return env[key]
		}

		checker, err := NewHTTPChecker("example", "localhost:8080", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		c := checker.(*HTTPChecker) // Cast the checker to HTTPChecker

		expectedHeaders := map[string]string{
			"Content-Type": "application/json",
		}

		if !reflect.DeepEqual(c.Headers, expectedHeaders) {
			t.Fatalf("expected headers %v, got %v", expectedHeaders, c.Headers)
		}
	})

	t.Run("Invalid HTTP check (duplicate headers)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Content-Type=application/json,Content-Type=application/json",
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

		_, err := NewHTTPChecker("example", "localhost:8080", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := "invalid HTTP_HEADERS value: duplicate header key found: Content-Type"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid HTTP check (malformed status range)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Auportpatrolization=Bearer token",
				envHTTPExpectedStatusCodes: "202-200",
			}
			return env[key]
		}

		_, err := NewHTTPChecker("example", "localhost:7654", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid status range: 202-200", envHTTPExpectedStatusCodes)
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid HTTP check (malformed header)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPHeaders:             "Auportpatrolization Bearer token", // Missing '=' in the header
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

		_, err := NewHTTPChecker("example", "http://example.com", 1*time.Second, mockEnv)
		if err == nil {
			t.Errorf("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid header format: Auportpatrolization Bearer token", envHTTPHeaders)
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid HTTP check (malformed HTTP_ALLOW_DUPLICATE_HEADERS)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:                "GET",
				envHTTPHeaders:               "Content-Type=application/json,Content-Type=application/json",
				envHTTPAllowDuplicateHeaders: "invalid",
				envHTTPExpectedStatusCodes:   "200",
			}
			return env[key]
		}

		_, err := NewHTTPChecker("example", "localhost:8080", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: strconv.ParseBool: parsing \"invalid\": invalid syntax", envHTTPAllowDuplicateHeaders)
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}

func TestParseHTTPHeaders(t *testing.T) {
	t.Parallel()

	t.Run("Valid headers", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Auportpatrolization=Bearer token"
		result, err := parseHTTPHeaders(headers, true)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"Content-Type": "application/json", "Auportpatrolization": "Bearer token"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Single header", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json"
		result, err := parseHTTPHeaders(headers, true)
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
		result, err := parseHTTPHeaders(headers, true)
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

		headers := "Content-Type=application/json,AuportpatrolizationBearer token"
		_, err := parseHTTPHeaders(headers, true)
		if err == nil {
			t.Error("Expected error, got nil")
		}

		expected := "invalid header format: AuportpatrolizationBearer token"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Header with spaces", func(t *testing.T) {
		t.Parallel()

		headers := "  Content-Type = application/json  , Auportpatrolization = Bearer token  "
		result, err := parseHTTPHeaders(headers, true)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"Content-Type": "application/json", "Auportpatrolization": "Bearer token"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Header with empty key", func(t *testing.T) {
		t.Parallel()

		headers := "=value"
		_, err := parseHTTPHeaders(headers, true)
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
		result, err := parseHTTPHeaders(headers, true)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := map[string]string{"key": ""}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected result: %q, got: %q", expected, result)
		}
	})

	t.Run("Trailing comma", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,"
		result, err := parseHTTPHeaders(headers, true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := map[string]string{"Content-Type": "application/json"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("Valid header with duplicate headers (allowDuplicates=true)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Content-Type=application/json"
		h, err := parseHTTPHeaders(headers, true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := map[string]string{"Content-Type": "application/json"}

		if !reflect.DeepEqual(h, expected) {
			t.Fatalf("expected %v, got %v", expected, h)
		}
	})

	t.Run("Invalid header with duplicate headers (allowDuplicates=false)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Content-Type=application/json"
		_, err := parseHTTPHeaders(headers, false)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := "duplicate header key found: Content-Type"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}

func TestParseHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	t.Run("Valid status code", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseHTTPStatusCodes("200")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Valid multiple status codes", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseHTTPStatusCodes("200,404,500")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200, 404, 500}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Valid status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseHTTPStatusCodes("200-202")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expected := []int{200, 201, 202}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %q, got %q", expected, statuses)
		}
	})

	t.Run("Valid multiple status code ranges", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseHTTPStatusCodes("200-202,300-301,500")
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

		_, err := parseHTTPStatusCodes("abc")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status code: abc"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid status range double dash", func(t *testing.T) {
		t.Parallel()

		_, err := parseHTTPStatusCodes("200--202")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status range: 200--202"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Invalid status range (start > end)", func(t *testing.T) {
		t.Parallel()

		_, err := parseHTTPStatusCodes("202-200")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status range: 202-200"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}

func TestIsValidCheckTypeWithProxy(t *testing.T) {
	t.Run("Invalid HTTP check (invalid proxy)", func(t *testing.T) {
		// Do not use t.Parallel here since we're modifying global state (environment variables)
		// t.Parallel()

		// Set the HTTP_PROXY environment variable to an invalid proxy
		err := os.Setenv("HTTP_PROXY", "http://invalid-proxy:8080")
		if err != nil {
			t.Fatalf("Failed to set HTTP_PROXY environment variable: %v", err)
		}
		defer os.Unsetenv("HTTP_PROXY") // Clean up after the test

		// Create the HTTPChecker instance
		checker, err := NewHTTPChecker("example", "http://example.com", 1*time.Second, os.Getenv)
		if err != nil {
			t.Errorf("expected no error, got %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		// Github throws a different error message than on my local machine, so check with contains
		expected := "Get \"http://example.com\": proxyconnect tcp: dial tcp: lookup invalid-proxy"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected error containing %q, got %q", expected, err.Error())
		}
	})
}
