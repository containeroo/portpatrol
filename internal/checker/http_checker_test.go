package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP check config", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:                "POST",
				envHTTPHeaders:               "Authorization=Bearer token",
				envHTTPExpectedStatusCodes:   "201",
				envHTTPSkipTLSVerify:         "false",
				envHTTPAllowDuplicateHeaders: "false",
			}
			return env[key]
		}

		checker, err := NewHTTPChecker("example", "http://localhost:8080", 10*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		checkerConfig := checker.(*HTTPChecker) // Type assertion to *HTTPChecker

		expected := "example"
		if checkerConfig.name != expected {
			t.Errorf("expected Name to be '%s', got %v", expected, checkerConfig.name)
		}

		expected = "http://localhost:8080"
		if checkerConfig.address != expected {
			t.Errorf("expected Address to be '%s', got %v", expected, checkerConfig.address)
		}

		expected = "POST"
		if checkerConfig.method != expected {
			t.Errorf("expected Method to be '%s', got %v", expected, checkerConfig.method)
		}

		expectedInsecureSkipVerify := false
		if checkerConfig.client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify != expectedInsecureSkipVerify {
			t.Errorf("expected client.Transport.TLSClientConfig.InsecureSkipVerify to be %v, got %v", expectedInsecureSkipVerify, checkerConfig.client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
		}

		expectedStatusCodes := []int{201}
		if len(checkerConfig.expectedStatusCodes) != len(expectedStatusCodes) || checkerConfig.expectedStatusCodes[0] != expectedStatusCodes[0] {
			t.Errorf("expected ExpectedStatusCodes to be %v, got %v", expectedStatusCodes, checkerConfig.expectedStatusCodes)
		}

		expectedHeaders := map[string]string{"Authorization": "Bearer token"}
		for key, value := range expectedHeaders {
			if checkerConfig.headers[key] != value {
				t.Errorf("expected Headers[%s] to be '%s', got '%s'", key, value, checkerConfig.headers[key])
			}
		}

		expectedTimeout := 10 * time.Second
		if checkerConfig.client.Timeout != expectedTimeout {
			t.Errorf("expected client Timeout to be '%v', got %v", expectedTimeout, checkerConfig.client.Timeout)
		}
	})

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
	t.Run("no scheme", func(t *testing.T) {
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

		url, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		checker, err := NewHTTPChecker("example", fmt.Sprintf("%s:%s", url.Hostname(), url.Port()), 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("failed to create request: parse \"%s:%s\": first path segment in URL cannot contain colon", url.Hostname(), url.Port())
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
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

		if !reflect.DeepEqual(c.headers, expectedHeaders) {
			t.Fatalf("expected headers %v, got %v", expectedHeaders, c.headers)
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

	t.Run("Invalid HTTP check (malformed HTTP_SKIP_TLS_VERIFY)", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envHTTPMethod:              "GET",
				envHTTPSkipTLSVerify:       "invalid",
				envHTTPExpectedStatusCodes: "200",
			}
			return env[key]
		}

		_, err := NewHTTPChecker("example", "localhost:8080", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: strconv.ParseBool: parsing \"invalid\": invalid syntax", envHTTPSkipTLSVerify)
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
