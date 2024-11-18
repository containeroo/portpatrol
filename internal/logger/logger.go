package logger

import (
	"io"
	"log/slog"
)

// SetupLogger configures the application logger.
func SetupLogger(version string, output io.Writer) *slog.Logger {
	logger := slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{}))
	return logger.With(slog.String("version", version))
}
