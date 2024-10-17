package logger

import (
	"io"
	"log/slog"

	"github.com/containeroo/portpatrol/internal/flags"
)

// SetupLogger configures the logger based on the configuration.
func SetupLogger(f *flags.ParsedFlags, output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{})).With(
		slog.String("version", f.Version),
	)
}
