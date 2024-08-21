package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/toast/pkg/config"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("WithAdditionalFields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		cfg := config.Config{
			Version:             "0.0.1",
			TargetAddress:       "localhost:8080",
			Interval:            1 * time.Second,
			DialTimeout:         2 * time.Second,
			CheckType:           "http",
			LogAdditionalFields: true,
		}

		logger := SetupLogger(cfg, &buf)
		logger.Info("Test log")

		logOutput := buf.String()

		if !strings.Contains(logOutput, "target_address=localhost:8080") ||
			!strings.Contains(logOutput, "interval=1s") ||
			!strings.Contains(logOutput, "dial_timeout=2s") ||
			!strings.Contains(logOutput, "checker_type=http") ||
			!strings.Contains(logOutput, "version=0.0.1") {
			t.Errorf("Logger output does not contain expected fields: %s", logOutput)
		}
	})

	t.Run("WithoutAdditionalFields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		cfg := config.Config{
			LogAdditionalFields: false,
		}

		logger := SetupLogger(cfg, &buf)
		logger.Error("Test error", slog.String("error", "some error"))

		logOutput := buf.String()

		expected := "error=some error"
		if strings.Contains(logOutput, expected) {
			t.Errorf("Expected error to contain %q, got %q", expected, logOutput)
		}
	})
}
