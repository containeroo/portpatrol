package checker

import (
	"reflect"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker("http", "example", "http://example.com", 5*time.Second, func(s string) string {
			return ""
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if check.String() != "example" {
			t.Fatalf("expected name to be 'example', got %v", check.String())
		}
	})

	t.Run("valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker("tcp", "example", "example.com:80", 5*time.Second, func(s string) string {
			return ""
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if check.String() != "example" {
			t.Fatalf("expected name to be 'example', got %v", check.String())
		}
	})

	t.Run("invalid checker type", func(t *testing.T) {
		t.Parallel()

		_, err := NewChecker("invalid", "example", "example.com", 5*time.Second, func(s string) string {
			return ""
		})
		if err == nil {
			t.Fatal("expected an error, got none")
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

	t.Run("http scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("http://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if checkType != "http" {
			t.Fatalf("expected 'http', got %v", checkType)
		}
	})

	t.Run("tcp scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("tcp://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if checkType != "tcp" {
			t.Fatalf("expected 'tcp', got %v", checkType)
		}
	})

	t.Run("no scheme", func(t *testing.T) {
		t.Parallel()

		checkType, err := InferCheckType("example.com:80")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if checkType != "tcp" {
			t.Fatalf("expected 'tcp', got %v", checkType)
		}
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		t.Parallel()

		_, err := InferCheckType("ftp://example.com")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}

func TestParseExpectedStatuses(t *testing.T) {
	t.Parallel()

	t.Run("single status code", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{200}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %v, got %v", expected, statuses)
		}
	})

	t.Run("multiple status codes", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200,404,500")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{200, 404, 500}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %v, got %v", expected, statuses)
		}
	})

	t.Run("status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200-202")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{200, 201, 202}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %v, got %v", expected, statuses)
		}
	})

	t.Run("multipl status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := parseExpectedStatuses("200-202,300-301,500")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{200, 201, 202, 300, 301, 500}
		if !reflect.DeepEqual(statuses, expected) {
			t.Fatalf("expected %v, got %v", expected, statuses)
		}
	})

	t.Run("invalid status code", func(t *testing.T) {
		t.Parallel()

		_, err := parseExpectedStatuses("abc")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})

	t.Run("invalid status range", func(t *testing.T) {
		t.Parallel()

		_, err := parseExpectedStatuses("202-200")
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}

func TestParseHeaders(t *testing.T) {
	t.Parallel()

	t.Run("single header", func(t *testing.T) {
		t.Parallel()

		headers := parseHeaders("Content-Type=application/json")
		expected := map[string]string{"Content-Type": "application/json"}
		if !reflect.DeepEqual(headers, expected) {
			t.Fatalf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("multiple headers", func(t *testing.T) {
		t.Parallel()

		headers := parseHeaders("Content-Type=application/json, Authorization=Bearer token")
		expected := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		}
		if !reflect.DeepEqual(headers, expected) {
			t.Fatalf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("headers with spaces", func(t *testing.T) {
		t.Parallel()

		headers := parseHeaders("Content-Type = application/json, Authorization = Bearer token")
		expected := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		}
		if !reflect.DeepEqual(headers, expected) {
			t.Fatalf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("empty headers", func(t *testing.T) {
		t.Parallel()

		headers := parseHeaders("")
		expected := map[string]string{}
		if !reflect.DeepEqual(headers, expected) {
			t.Fatalf("expected %v, got %v", expected, headers)
		}
	})
}
