package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/portpatrol/internal/config"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("Log with additional fields", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Version:         "0.0.1",
			TargetAddress:   "localhost:8080",
			CheckInterval:   1 * time.Second,
			DialTimeout:     2 * time.Second,
			TargetCheckType: "http",
			LogExtraFields:  true,
		}
		var buf bytes.Buffer

		logger := SetupLogger(cfg, &buf)
		logger.Info("Test log")

		logOutput := buf.String()

		expected := "target_address=localhost:8080"
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log output to contain %q, got %q", expected, logOutput)
		}

		expected = "interval=1s"
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log output to contain %q, got %q", expected, logOutput)
		}

		expected = "dial_timeout=2s"
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log output to contain %q, got %q", expected, logOutput)
		}

		expected = "checker_type=http"
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log output to contain %q, got %q", expected, logOutput)
		}

		expected = "version=0.0.1"
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log output to contain %q, got %q", expected, logOutput)
		}
	})

	t.Run("Log without additional fields", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			LogExtraFields: false,
		}
		var buf bytes.Buffer

		logger := SetupLogger(cfg, &buf)
		logger.Error("Test error", slog.String("error", "some error"))

		logOutput := buf.String()

		expected := "error=some error"
		if strings.Contains(logOutput, expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, logOutput)
		}
	})
}
