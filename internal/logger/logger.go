package logger

import (
	"io"
	"log/slog"

	"github.com/containeroo/portpatrol/internal/flags"
)

// SetupLogger configures the logger based on the configuration.
func SetupLogger(f *flags.ParsedFlags, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	if f.LogExtraFields {
		// Return a logger with the additional fields
		return slog.New(slog.NewTextHandler(output, handlerOpts)).With(
			slog.String("interval", f.DefaultCheckInterval.String()),
			slog.String("dial_timeout", f.DefaultDialTimeout.String()),
			slog.String("version", f.Version),
		)
	}

	// If logExtraFields is false, remove the error attribute from the handler.
	// The error attribute is unwanted when no additional fields is set to true.
	handlerOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "error" {
			return slog.Attr{}
		}
		return a
	}

	// Return a logger without the additional fields and with a function to remove the error attribute
	return slog.New(slog.NewTextHandler(output, handlerOpts))
}
