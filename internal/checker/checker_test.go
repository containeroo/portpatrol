package checker

import (
	"testing"
)

func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(HTTP, "example", "http://example.com")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expectedName := "example"
		if check.GetName() != expectedName {
			t.Fatalf("expected name to be %q, got %q", expectedName, check.GetName())
		}

		expectedType := "HTTP"
		if check.GetType() != expectedType {
			t.Fatalf("expected type to be %q, got %q", expectedType, check.GetType())
		}
	})

	t.Run("Valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(TCP, "example", "example.com:80")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expectedName := "example"
		if check.GetName() != expectedName {
			t.Fatalf("expected name to be %q, got %q", expectedName, check.GetName())
		}

		expectedType := "TCP"
		if check.GetType() != expectedType {
			t.Fatalf("expected type to be %q, got %q", expectedType, check.GetType())
		}
	})

	t.Run("Valid ICMP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(ICMP, "example", "example.com")
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		expectedName := "example"
		if check.GetName() != expectedName {
			t.Fatalf("expected name to be %q, got %q", expectedName, check.GetName())
		}

		expectedType := "ICMP"
		if check.GetType() != expectedType {
			t.Fatalf("expected type to be %q, got %q", expectedType, check.GetType())
		}
	})

	t.Run("Invalid checker type", func(t *testing.T) {
		t.Parallel()

		_, err := NewChecker(99, "example", "example.com")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expectedErr := "unsupported check type: 99"
		if err.Error() != expectedErr {
			t.Fatalf("expected error to be %q, got %q", expectedErr, err.Error())
		}
	})
}

func TestGetCheckTypeFromString(t *testing.T) {
	t.Parallel()

	t.Run("Valid check types", func(t *testing.T) {
		tests := []struct {
			input    string
			expected CheckType
		}{
			{"http", HTTP},
			{"tcp", TCP},
			{"icmp", ICMP},
			{"HTTP", HTTP},
			{"TCP", TCP},
			{"ICMP", ICMP},
		}

		for _, tc := range tests {
			t.Run(tc.input, func(t *testing.T) {
				t.Parallel()

				result, err := GetCheckTypeFromString(tc.input)
				if err != nil {
					t.Fatalf("expected no error, got %q", err)
				}

				if result != tc.expected {
					t.Fatalf("expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("Invalid check type", func(t *testing.T) {
		t.Parallel()

		_, err := GetCheckTypeFromString("invalid")
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expectedErr := "unsupported check type: invalid"
		if err.Error() != expectedErr {
			t.Fatalf("expected error to be %q, got %q", expectedErr, err.Error())
		}
	})
}

func TestCheckTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		checkType CheckType
		expected  string
	}{
		{HTTP, "HTTP"},
		{TCP, "TCP"},
		{ICMP, "ICMP"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()

			if result := tc.checkType.String(); result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
