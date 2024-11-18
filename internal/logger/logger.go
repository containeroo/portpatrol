package logger

import (
	"io"
	"log/slog"
)

// SetupLogger configures the logger based on the configuration.
func SetupLogger(version string, output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{})).With(
		slog.String("version", version),
	)
}
