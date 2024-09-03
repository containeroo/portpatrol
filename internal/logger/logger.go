package logger

import (
	"io"
	"log/slog"

	"github.com/containeroo/portpatrol/internal/config"
)

// SetupLogger configures the logger based on the configuration
func SetupLogger(cfg config.Config, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	if cfg.LogExtraFields {
		// Return a logger with the additional fields
		return slog.New(slog.NewTextHandler(output, handlerOpts)).With(
			slog.String("target_address", cfg.TargetAddress),
			slog.String("interval", cfg.CheckInterval.String()),
			slog.String("dial_timeout", cfg.DialTimeout.String()),
			slog.String("checker_type", cfg.TargetCheckType),
			slog.String("version", cfg.Version),
		)
	}

	// If logExtraFields is false, remove the error attribute from the handler
	// The error attribute is unwanted when no additional fields is set to true
	handlerOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "error" {
			return slog.Attr{}
		}
		return a
	}

	// Return a logger without the additional fields and with a function to remove the error attribute
	return slog.New(slog.NewTextHandler(output, handlerOpts))
}
