package httputils

import (
	"reflect"
	"testing"
)

func TestParseHTTPHeaders(t *testing.T) {
	t.Parallel()

	t.Run("Valid headers", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Auportpatrolization=Bearer token"
		result, err := ParseHeaders(headers, true)
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
		result, err := ParseHeaders(headers, true)
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
		result, err := ParseHeaders(headers, true)
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
		_, err := ParseHeaders(headers, true)
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
		result, err := ParseHeaders(headers, true)
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
		_, err := ParseHeaders(headers, true)
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
		result, err := ParseHeaders(headers, true)
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
		result, err := ParseHeaders(headers, true)
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
		h, err := ParseHeaders(headers, true)
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
		_, err := ParseHeaders(headers, false)
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

		statuses, err := ParseStatusCodes("200")
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

		statuses, err := ParseStatusCodes("200,404,500")
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

		statuses, err := ParseStatusCodes("200-202")
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

		statuses, err := ParseStatusCodes("200-202,300-301,500")
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

		_, err := ParseStatusCodes("abc")
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

		_, err := ParseStatusCodes("200--202")
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

		_, err := ParseStatusCodes("202-200")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "invalid status range: 202-200"
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})
}
