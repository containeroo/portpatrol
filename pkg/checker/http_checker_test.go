package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("valid HTTP check", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		headers := map[string]string{"Authorization": "Bearer token"}
		expectedStatusCodes := []int{200}

		checker, _ := NewHTTPChecker("example", server.URL, "GET", headers, expectedStatusCodes, 1*time.Second)
		err := checker.Check(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unexpected status code", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		headers := map[string]string{"Authorization": "Bearer token"}
		expectedStatusCodes := []int{200}

		checker, _ := NewHTTPChecker("example", server.URL, "GET", headers, expectedStatusCodes, 1*time.Second)
		err := checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}
