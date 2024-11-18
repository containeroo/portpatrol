package logging

import (
	"strings"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	// Test that the logger includes the version and outputs correctly.
	t.Run("Logger includes version and writes to output", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		version := "1.0.0"

		logger := SetupLogger(version, &output)
		if logger == nil {
			t.Fatalf("Expected a logger instance, got nil")
		}

		logger.Info("Test log message")

		logOutput := output.String()
		if !strings.Contains(logOutput, "Test log message") {
			t.Errorf("Expected log output to contain 'Test log message', got %q", logOutput)
		}

		if !strings.Contains(logOutput, "version=1.0.0") {
			t.Errorf("Expected log output to contain 'version=1.0.0', got %q", logOutput)
		}
	})

	// Test that the logger writes output to the correct writer.
	t.Run("Logger writes to specified output", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		version := "2.0.0"

		logger := SetupLogger(version, &output)
		if logger == nil {
			t.Fatalf("Expected a logger instance, got nil")
		}

		logger.Warn("This is a warning")

		logOutput := output.String()
		if !strings.Contains(logOutput, "This is a warning") {
			t.Errorf("Expected log output to contain 'This is a warning', got %q", logOutput)
		}
	})
}
